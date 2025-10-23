//go:build android || darwin || linux
// +build android darwin linux

package Tun

import (
	"io"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/SunnyNetUDP"
	"github.com/qtgolang/SunnyNet/src/public"
)

type connEntry struct {
	LastSeen     time.Time
	ClientIP     net.IP
	ClientPort   uint16
	ServerIP     net.IP
	ServerPort   uint16
	conn         *net.UDPConn
	Theology     int64
	fd           io.ReadWriteCloser
	v4           bool
	pid          int32
	pidFromCheck bool
	callback     func(Type int, Theoni int64, pid uint32, LocalAddress, RemoteAddress string, data []byte) []byte
	mu           *sync.Mutex
}

func (c *connEntry) ToClient(payload []byte) bool {
	if c.fd == nil {
		return false
	}
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	var err error

	if c.v4 {
		ip := &layers.IPv4{
			Version:  4,
			IHL:      5,
			TTL:      64,
			SrcIP:    c.ServerIP,
			DstIP:    c.ClientIP,
			Protocol: layers.IPProtocolUDP,
		}
		udp := &layers.UDP{
			SrcPort: layers.UDPPort(c.ServerPort),
			DstPort: layers.UDPPort(c.ClientPort),
		}
		_ = udp.SetNetworkLayerForChecksum(ip)

		err = gopacket.SerializeLayers(buf, opts, ip, udp, gopacket.Payload(payload))
	} else {
		ip6 := &layers.IPv6{
			Version:    6,
			HopLimit:   64,
			SrcIP:      c.ServerIP,
			DstIP:      c.ClientIP,
			NextHeader: layers.IPProtocolUDP,
		}
		udp := &layers.UDP{
			SrcPort: layers.UDPPort(c.ServerPort),
			DstPort: layers.UDPPort(c.ClientPort),
		}
		_ = udp.SetNetworkLayerForChecksum(ip6)

		err = gopacket.SerializeLayers(buf, opts, ip6, udp, gopacket.Payload(payload))
	}

	if err != nil {
		return false
	}

	packetData := buf.Bytes()
	_, err = c.fd.Write(packetData)
	return err == nil
}

func (c *connEntry) ToServer(bytes []byte) bool {
	if c.conn == nil {
		return false
	}
	_, e := c.conn.Write(bytes)
	return e == nil
}

// 全局表
var (
	connTable = make(map[uint16]*connEntry)
	connMu    sync.Mutex
)

func (n *NewTun) handleUDP(srcIP, dstIP net.IP, udp *layers.UDP, v4 bool) {
	Payload := udp.Payload
	if len(Payload) == 0 {
		return
	}
	clientIP := srcIP
	clientPort := uint16(udp.SrcPort)
	serverIP := dstIP
	serverPort := uint16(udp.DstPort)
	connMu.Lock()
	obj := connTable[clientPort]
	if obj == nil {
		var mu sync.Mutex
		obj = &connEntry{ClientIP: clientIP, ClientPort: clientPort, ServerIP: serverIP, ServerPort: serverPort, Theology: atomic.AddInt64(&public.Theology, 1), v4: v4, callback: n.handleUDPCallback, mu: &mu, fd: n.tun}
		pid, name := getPidByPort("udp", clientPort)
		obj.pid = pid
		obj.pidFromCheck = n.pidFromCheck(obj.pid, name)
		connTable[clientPort] = obj
		connMu.Unlock()
		mu.Lock()
		target := &net.UDPAddr{IP: serverIP, Port: int(serverPort)}
		var localAddr *net.UDPAddr
		if defaultGatewayIP != "" {
			localAddr = &net.UDPAddr{IP: net.ParseIP(defaultGatewayIP), Port: 0}
		}
		conn, er := net.DialUDP("udp", localAddr, target)
		if er != nil {
			mu.Unlock()
			connMu.Lock()
			delete(connTable, clientPort)
			connMu.Unlock()
			return
		}
		mu.Unlock()
		connMu.Lock()
		obj.conn = conn
		SunnyNetUDP.AddUDPItem(obj.Theology, obj)
		go obj.loop()
	}
	obj.mu.Lock()
	defer obj.mu.Unlock()
	obj.LastSeen = time.Now()
	connMu.Unlock()
	if obj.conn != nil {
		if obj.callback != nil {
			if obj.pidFromCheck {
				obj.ToServer(Payload)
				return
			}
			LocalAddress := net.JoinHostPort(obj.ClientIP.String(), strconv.Itoa(int(obj.ClientPort)))
			RemoteAddress := net.JoinHostPort(obj.ServerIP.String(), strconv.Itoa(int(obj.ServerPort)))
			bs := obj.callback(public.SunnyNetUDPTypeSend, obj.Theology, uint32(obj.pid), LocalAddress, RemoteAddress, Payload)
			if len(bs) < 1 {
				return
			}
			obj.ToServer(bs)
			return
		}
		obj.ToServer(Payload)
	}
	return
}
func (c *connEntry) loop() {
	LocalAddress := net.JoinHostPort(c.ClientIP.String(), strconv.Itoa(int(c.ClientPort)))
	RemoteAddress := net.JoinHostPort(c.ServerIP.String(), strconv.Itoa(int(c.ServerPort)))
	buff := make([]byte, 0xffff)
	for {
		_ = c.conn.SetReadDeadline(time.Now().Add(time.Duration(10) * time.Second))
		nt, _, _ := c.conn.ReadFromUDP(buff)
		if nt == 0 {
			connMu.Lock()
			if time.Now().After(c.LastSeen.Add(time.Duration(30) * time.Second)) {
				connMu.Unlock()
				break
			}
			connMu.Unlock()
			continue
		}
		connMu.Lock()
		c.LastSeen = time.Now()
		connMu.Unlock()
		if c.callback != nil {
			if c.pidFromCheck {
				c.ToClient(buff[:nt])
				return
			}
			bs := c.callback(public.SunnyNetUDPTypeReceive, c.Theology, uint32(c.pid), LocalAddress, RemoteAddress, buff[:nt])
			if len(bs) < 1 {
				continue
			}
			c.ToClient(bs)
		} else {
			c.ToClient(buff[:nt])
		}
	}
	connMu.Lock()
	if connTable[c.ClientPort] != nil {
		SunnyNetUDP.DelUDPItem(c.Theology)
		if c.pidFromCheck {
			return
		}
		if c.callback != nil {
			c.callback(public.SunnyNetUDPTypeClosed, c.Theology, uint32(c.pid), LocalAddress, RemoteAddress, nil)
		}
		delete(connTable, c.ClientPort)
	}
	connMu.Unlock()
}
