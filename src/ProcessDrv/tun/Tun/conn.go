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

	"github.com/qtgolang/SunnyNet/src/ProcessDrv/ProcessCheck"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/tun/tunPublic"
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
	inChSeg              map[uint32][]byte
	ts                   time.Time
}

func (d *DevConn) GetRemoteAddress() string {
	if tunPublic.IsLocalIp(d.serverIP) {
		return net.JoinHostPort("127.0.0.1", strconv.Itoa(int(d.serverPort)))
	}
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
func NewDevConn(h io.ReadWriteCloser, clientIP net.IP, clientPort uint16, serverIP net.IP, serverPort uint16, ipv4 bool, clientSynSeq uint32) *DevConn {
	// 创建 DevConn 基本信息
	d := &DevConn{
		clientIP:   clientIP,               // 客户端 IP
		clientPort: clientPort,             // 客户端端口
		serverIP:   serverIP,               // 目标服务器 IP
		serverPort: serverPort,             // 目标服务器端口
		dataCh:     make(chan struct{}, 1), // 数据通知 channel（缓冲 1，避免阻塞）
		v4:         ipv4,                   // 是否 IPv4
		tun:        h,                      // TUN 句柄
	}
	d.ts = time.Now()
	// 客户端期望 seq：客户端 ISN + 1
	// 第一个数据包的 tcp.Seq 必须等于这个值才会被接收
	d.clientNext = clientSynSeq + 1

	// 伪造一个服务端 ISN
	d.serverISN = rand.Uint32()

	// 我们接下来发给客户端的 Seq 从 serverISN+1 开始
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
	select {
	case d.dataCh <- struct{}{}:
	default:
	}
	d.inChSeg = nil
	d.mu.Unlock()
	sessionsMu.Lock()
	delete(sessions, d.clientPort)
	sessionsMu.Unlock()
	ProcessCheck.DelDevObj(d.clientPort)
	if !already {
		_, _ = d.tun.Write(SendRstToClient(d))
	}
	return nil
}

// RemoteAddr 返回真实的 server 地址
func (d *DevConn) RemoteAddr() net.Addr {
	if tunPublic.IsLocalIp(d.serverIP) {
		return &net.TCPAddr{
			IP:   loopIp,
			Port: int(d.serverPort),
		}
	}
	return &net.TCPAddr{
		IP:   d.serverIP,
		Port: int(d.serverPort),
	}
}

var loopIp = net.ParseIP("127.0.0.1")

// LocalAddr 返回伪造的 client 地址
func (d *DevConn) LocalAddr() net.Addr {
	if tunPublic.IsLocalIp(d.clientIP) {
		return &net.TCPAddr{
			IP:   loopIp,
			Port: int(d.clientPort),
		}
	}
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
	// 递归式补齐乱序分片：当前分片处理完之后，再看缓存里有没有刚好接上的
	defer func() {
		d.mu.Lock()
		for s, seg := range d.inChSeg {
			// 过滤掉“比当前 clientNext 还早”的旧分片
			if seqBefore(s, d.clientNext) {
				delete(d.inChSeg, s)
				continue
			}
			// 如果刚好是下一个期望 seq，就拿出来递归处理
			if seqEqual(s, d.clientNext) {
				delete(d.inChSeg, s)
				// 递归调用前先解锁，避免死锁
				d.mu.Unlock()
				d.PushClientPayload(seg, s)
				return
			}
			// 剩下的是“未来分片”，继续保留在 inChSeg 里
		}
		d.mu.Unlock()
	}()
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return
	}
	// 严格按序接收：这里只处理“刚好等于期望 seq”的分片
	if !seqEqual(seq, d.clientNext) {
		// 如果是未来的 seq，先缓存起来，等缺失的分片到了再补
		if seqAfter(seq, d.clientNext) {
			if d.inChSeg == nil {
				d.inChSeg = make(map[uint32][]byte)
			}
			// 这里直接覆盖同一 seq 的旧缓存，一般没有问题
			d.inChSeg[seq] = payload
		}
		d.mu.Unlock()
		return
	}
	needNotify := len(payload) > 0
	if !needNotify {
		d.mu.Unlock()
		return
	}
	_, _ = d.buff.Write(payload)
	d.clientNext = seq + uint32(len(payload))
	clientNext := d.clientNext
	d.mu.Unlock()
	// 锁外发送 ACK
	if bs := SendAckToKernel(d, clientNext); len(bs) > 0 {
		_, _ = d.tun.Write(bs)
	}
	// 锁外通知
	if needNotify {
		select {
		case d.dataCh <- struct{}{}:
			return // 发送成功就退出
		default:
		}
	}
}

// 回绕安全：a 是否在 b 之后（a > b）,避免uint32溢出问题
func seqAfter(a, b uint32) bool {
	return int32(a-b) > 0
}

// 回绕安全：a 是否在 b 之前（a < b）,避免uint32溢出问题
func seqBefore(a, b uint32) bool {
	return int32(a-b) < 0
}

// 回绕安全：a == b,避免uint32溢出问题
func seqEqual(a, b uint32) bool {
	return a == b
}
