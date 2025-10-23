//go:build android || darwin || linux
// +build android darwin linux

package Tun

import (
	"io"
	"net"
	"strconv"
	"sync"

	"github.com/google/gopacket/layers"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/ProcessCheck"
)

// ------------------------------------------------
// IPv4 TCP 处理函数
// ------------------------------------------------
func (n *NewTun) handleTCPCommand(tcp *layers.TCP, srcIP, dstIP net.IP, clientPort, serverPort uint16, v4 bool) {
	if tcp.SYN && !tcp.ACK {
		pid, name := getPidByPort("tcp", uint16(tcp.SrcPort))
		s := NewDevConn(n.tun, srcIP, clientPort, dstIP, serverPort, v4, tcp.Seq, tcp.Ack)
		s.pid = uint32(pid)
		sessionsMu.Lock()
		sessions[clientPort] = s
		sessionsMu.Unlock()
		// send SYN/ACK to client
		if _, err := SendSynAckToClient(n.tun, s, tcp.Seq); err != nil {
			// 如果发送失败，删除会话
			sessionsMu.Lock()
			delete(sessions, clientPort)
			sessionsMu.Unlock()
			return
		}
		sessionsMu.Lock()
		call := n.handleTCPCallback
		sessionsMu.Unlock()
		if n.pidFromCheck(pid, name) {
			go func() {
				var loader *net.TCPAddr
				if defaultGatewayIP != "" {
					loader = &net.TCPAddr{
						IP:   net.ParseIP(defaultGatewayIP),
						Port: 0,
					}
				}
				dialer := &net.Dialer{LocalAddr: loader}

				c, e := dialer.Dial("tcp", net.JoinHostPort(dstIP.String(), strconv.Itoa(int(serverPort))))
				if e != nil {
					_ = SendRstToClient(s)
					_ = s.Close()
					return
				}
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					_, _ = io.Copy(s, c)
					_ = s.Close()
					_ = c.Close()
					wg.Done()
				}()
				go func() {
					_, _ = io.Copy(c, s)
					_ = s.Close()
					_ = c.Close()
					wg.Done()
				}()
				wg.Wait()
				sessionsMu.Lock()
				delete(sessions, clientPort)
				sessionsMu.Unlock()
			}()
			return
		}
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
		h2 := NewDevConn(n.tun, srcIP, clientPort, dstIP, serverPort, true, tcp.Seq, tcp.Ack)
		_ = SendRstToClient(h2)
		return
	}
	if tcp.FIN || tcp.RST {
		sessionsMu.Lock()
		delete(sessions, clientPort)
		sessionsMu.Unlock()
		ProcessCheck.DelDevObj(clientPort)
		if tcp.FIN {
			_, _ = buildTCPReply(sess, tcp, true, true, false)
		} else if tcp.RST {
			_, _ = buildTCPReply(sess, tcp, true, false, false)
		}
		return
	}
	if len(tcp.Payload) > 0 {
		sess.PushClientPayload(tcp.Payload, tcp.Seq)
		return
	}
	sess.mu.Lock()
	sess.clientNext = tcp.Seq
	sess.mu.Unlock()
	return
}
func (n *NewTun) handleTCP4(ip *layers.IPv4, tcp *layers.TCP) {
	clientIP := ip.SrcIP
	clientPort := uint16(tcp.SrcPort)
	serverIP := ip.DstIP
	serverPort := uint16(tcp.DstPort)
	n.handleTCPCommand(tcp, clientIP, serverIP, clientPort, serverPort, true)
}

// ------------------------------------------------
// IPv6 TCP 处理函数
// ------------------------------------------------
func (n *NewTun) handleTCP6(ip *layers.IPv6, tcp *layers.TCP) {
	clientIP := ip.SrcIP
	clientPort := uint16(tcp.SrcPort)
	serverIP := ip.DstIP
	serverPort := uint16(tcp.DstPort)
	n.handleTCPCommand(tcp, clientIP, serverIP, clientPort, serverPort, false)
}
