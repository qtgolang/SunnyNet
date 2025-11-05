//go:build windows
// +build windows

package WinDivert

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/ProcessCheck"
	CrossCompiled "github.com/qtgolang/SunnyNet/src/iphlpapi/net"
	"github.com/shirou/gopsutil/process"
	"net"
	"sync"
	"time"
)

var (
	sessionsMu sync.Mutex
	sessions   = make(map[uint16]*DevConn)
)
 
type expiry struct {
	pid    int32
	name   string
	expiry time.Time
}

var (
	pidCache   = make(map[uint16]expiry)
	pidCacheMu sync.RWMutex
	pidTTL     = 3 * time.Second
	pidExpiry  = make(map[uint16]time.Time)
)

func getPidByPort(kind string, port uint16) (int32, string) {
	pidCacheMu.Lock()
	defer pidCacheMu.Unlock()
	for k, _ := range pidCache {
		if !time.Now().Before(pidExpiry[k]) {
			delete(pidExpiry, k)
			delete(pidCache, k)
		}
	}
	if obj, ok := pidCache[port]; ok && time.Now().Before(pidExpiry[port]) {
		return obj.pid, obj.name
	}
	all, _ := CrossCompiled.Connections(kind)
	for _, conn := range all {
		if conn.Laddr.Port == uint32(port) {
			pid := conn.Pid
			p, _ := process.NewProcess(pid)
			ch := expiry{pid: pid}
			if p != nil {
				ch.name, _ = p.Name()
			}
			pidCache[port] = ch
			pidExpiry[port] = time.Now().Add(pidTTL)
			return ch.pid, ch.name
		}
	}
	return 0, ""
}

var loopbackV4 = net.IPv4(127, 0, 0, 1)
var loopbackV6 = net.IPv6loopback
var mm sync.Mutex

func (d *Divert) handleCommand(h *Handle, data []byte, addr *Address, tcp *layers.TCP, clientIP, serverIP net.IP, clientPort, serverPort uint16, v4 bool) {
	mm.Lock()
	defer mm.Unlock()
	if !addr.Outbound() || (clientIP.Equal(serverIP) && (serverIP.Equal(loopbackV4) || serverIP.Equal(loopbackV6))) {
		_, _ = h.Send(data, addr)
		return
	}

	// 只处理 SYN 或 TCP payload
	// 处理 SYN（client 发起连接）
	if tcp.SYN && !tcp.ACK {
		pid, name := getPidByPort("tcp", uint16(tcp.SrcPort))
		if d.pidFromCheck(pid, name) {
			_, _ = h.Send(data, addr)
			return
		}
		// 创建会话并伪造 SYN/ACK 返回客户端
		s := NewDevConn(h, clientIP, clientPort, serverIP, serverPort, v4, addr.Clone(), 0, 0)
		s.pid = uint32(pid)
		sessionsMu.Lock()
		sessions[clientPort] = s
		sessionsMu.Unlock()
		// send SYN/ACK to client
		if err := SendSynAckToClient(h, s, addr, tcp.Seq); err != nil {
			// 如果发送失败，删除会话
			sessionsMu.Lock()
			delete(sessions, clientPort)
			sessionsMu.Unlock()
			return
		}
		sessionsMu.Lock()
		call := d.handleTCP
		sessionsMu.Unlock()
		if call != nil {
			ProcessCheck.AddDevObj(clientPort, s)
			go call(s)
		}
		return
	}
	sessionsMu.Lock()
	sess, ok := sessions[clientPort]
	sessionsMu.Unlock()
	if !ok {
		pid, name := getPidByPort("tcp", uint16(tcp.SrcPort))
		if d.pidFromCheck(pid, name) {
			_, _ = h.Send(data, addr)
			return
		}
		h2 := NewDevConn(h, clientIP, clientPort, serverIP, serverPort, v4, addr.Clone(), tcp.Seq, tcp.Ack)
		_ = SendRstToClient(h, h2)
		return
	}
	// 如果收到 FIN 或 RST，则清理 session 并放行
	if tcp.FIN || tcp.RST {
		_, _ = h.Send(data, addr) // 放行原始包（可选）
		sessionsMu.Lock()
		delete(sessions, clientPort)
		sessionsMu.Unlock()
		ProcessCheck.DelDevObj(clientPort)
		_ = sess.Close()
		return
	}

	// 处理 payload：写入 devConn 并向内核注入 ACK（告知 we've consumed bytes）
	if len(tcp.Payload) > 0 {
		// 写入 session buffer 并更新 clientNext
		sess.PushClientPayload(tcp.Payload, tcp.Seq)
		return
	}
	sess.mu.Lock()
	sess.clientNext = tcp.Seq
	sess.mu.Unlock()
	// 其他情况原样放行
	//_, _ = h.Send(data, addr)
	return
}
func (d *Divert) handleIPv4(h *Handle, data []byte, addr *Address, ip4 *layers.IPv4, pkt gopacket.Packet) bool {
	tcpLayer := pkt.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return false
	}
	tcp := tcpLayer.(*layers.TCP)
	clientIP := ip4.SrcIP
	clientPort := uint16(tcp.SrcPort)
	serverIP := ip4.DstIP
	serverPort := uint16(tcp.DstPort)
	d.handleCommand(h, data, addr, tcp, clientIP, serverIP, clientPort, serverPort, true)
	return true
}
func (d *Divert) handleIPv6(h *Handle, data []byte, addr *Address, ip6 *layers.IPv6, pkt gopacket.Packet) bool {
	tcpLayer := pkt.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return false
	}
	tcp := tcpLayer.(*layers.TCP)
	clientIP := ip6.SrcIP
	clientPort := uint16(tcp.SrcPort)
	serverIP := ip6.DstIP
	serverPort := uint16(tcp.DstPort)
	d.handleCommand(h, data, addr, tcp, clientIP, serverIP, clientPort, serverPort, false)
	return true
}
