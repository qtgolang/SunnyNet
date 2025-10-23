//go:build android || darwin || linux
// +build android darwin linux

package Tun

import (
	"bytes"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type DevConn struct {
	// 标识：客户端和伪服务端的四元组
	clientIP   net.IP
	clientPort uint16
	serverIP   net.IP
	serverPort uint16

	// TCP 序列号跟踪
	clientNext    uint32 // 客户端下一个期望的 seq
	serverISN     uint32 // 我们伪造的 server 初始序列号
	serverSeqNext uint32 // 我们发送给客户端时的 seq

	// 缓存和同步
	buff   bytes.Buffer
	mu     sync.Mutex
	closed bool
	// 通知 channel：避免每次等待都 spawn goroutine
	dataCh chan struct{}
	// deadline
	_outRead  time.Time
	_outWrite time.Time
	// 记录已发 ack
	highestClientAckSent uint32
	tun                  io.ReadWriteCloser
	v4                   bool
	pid                  uint32
}

func (d *DevConn) GetRemoteAddress() string {
	return net.JoinHostPort(d.serverIP.String(), strconv.Itoa(int(d.serverPort)))
}

func (d *DevConn) GetRemotePort() uint16 {
	return d.serverPort
}

func (d *DevConn) GetPid() string {
	return strconv.Itoa(int(d.pid))
}

func (d *DevConn) IsV6() bool {
	return !d.v4
}

func (d *DevConn) ID() uint64 {
	return uint64(d.clientPort)
}

// 构造函数
func NewDevConn(h io.ReadWriteCloser, clientIP net.IP, clientPort uint16, serverIP net.IP, serverPort uint16, ipv4 bool, seq, ack uint32) *DevConn {
	d := &DevConn{
		clientIP:      clientIP,
		clientPort:    clientPort,
		serverIP:      serverIP,
		serverPort:    serverPort,
		dataCh:        make(chan struct{}, 1),
		v4:            ipv4,
		serverSeqNext: seq,
		clientNext:    ack,
		tun:           h,
	}
	d.serverISN = rand.Uint32()
	d.serverSeqNext = d.serverISN + 1
	return d
}

// --- net.Conn 接口实现 ---

// Read 从缓冲区读取客户端发来的数据
func (d *DevConn) Read(b []byte) (int, error) {
	a1, a2 := d.read(b)
	return a1, a2
}

// Read 从缓冲区读取客户端发来的数据（使用 dataCh 通知，避免额外 goroutine）
func (d *DevConn) read(b []byte) (int, error) {
	for {
		d.mu.Lock()
		if d.buff.Len() > 0 {
			n, _ := d.buff.Read(b)
			d.mu.Unlock()
			return n, nil
		}
		if d.closed {
			d.mu.Unlock()
			return 0, io.EOF
		}
		// 拿出 deadline 本地变量，避免在 select 中访问共享状态
		deadline := d._outRead
		d.mu.Unlock()

		if deadline.IsZero() {
			// 阻塞等待通知
			<-d.dataCh
			// loop to check buffer
			continue
		}

		// 有 deadline，则等待 dataCh 或超时
		now := time.Now()
		if !deadline.After(now) {
			return 0, os.ErrDeadlineExceeded
		}
		timer := time.NewTimer(time.Until(deadline))
		select {
		case <-d.dataCh:
			if !timer.Stop() {
				<-timer.C
			}
			// loop to read
		case <-timer.C:
			return 0, os.ErrDeadlineExceeded
		}
	}
}

// Write 将数据发回客户端（通过 WinDivert 注入包）
func (d *DevConn) Write(b []byte) (int, error) {
	// 写超时检查
	if !d._outWrite.IsZero() && time.Now().After(d._outWrite) {
		return 0, os.ErrDeadlineExceeded
	}
	return SendDataToClient(d, b)
}

// Close 关闭连接：注入 FIN 并清理
func (d *DevConn) Close() error {
	d.mu.Lock()
	already := d.closed
	d.closed = true
	// signal readers (non-blocking send to channel)
	select {
	case d.dataCh <- struct{}{}:
	default:
	}
	d.mu.Unlock()
	if already {
		return nil
	}
	_, _ = d.tun.Write(SendFinToClient(d))
	return nil
}

// RemoteAddr 返回真实的 server 地址
func (d *DevConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{
		IP:   d.serverIP,
		Port: int(d.serverPort),
	}
}

// LocalAddr 返回伪造的 client 地址
func (d *DevConn) LocalAddr() net.Addr {
	return &net.TCPAddr{
		IP:   d.clientIP,
		Port: int(d.clientPort),
	}
}

// SetDeadline 同时设置读写超时
func (d *DevConn) SetDeadline(t time.Time) error {
	_ = d.SetReadDeadline(t)
	_ = d.SetWriteDeadline(t)
	return nil
}

// SetReadDeadline 设置读超时
func (d *DevConn) SetReadDeadline(t time.Time) error {
	d.mu.Lock()
	d._outRead = t
	d.mu.Unlock()
	select {
	case d.dataCh <- struct{}{}:
	default:
	}
	return nil
}

// SetWriteDeadline 设置写超时
func (d *DevConn) SetWriteDeadline(t time.Time) error {
	d.mu.Lock()
	d._outWrite = t
	d.mu.Unlock()
	return nil
}

func (d *DevConn) PushClientPayload(payload []byte, seq uint32) {
	if seq != d.clientNext {
		return
	}
	defer func() {
		d.mu.Lock()
		clientNext := d.clientNext
		d.mu.Unlock()
		if clientNext != 0 {
			_, _ = d.tun.Write(SendAckToKernel(d, clientNext))
		}
	}()
	if len(payload) == 0 {
		// 仍更新 seq
		d.mu.Lock()
		d.clientNext = seq + uint32(len(payload))
		d.mu.Unlock()
		return
	}
	d.mu.Lock()
	d.buff.Write(payload)
	d.clientNext = seq + uint32(len(payload))
	// 非阻塞通知
	select {
	case d.dataCh <- struct{}{}:
	default:
	}
	d.mu.Unlock()
}
