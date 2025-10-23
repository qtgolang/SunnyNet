//go:build android || darwin || linux
// +build android darwin linux

package Tun

import (
	"io"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var sessionsMu sync.Mutex

var sessions = make(map[uint16]*DevConn)

func (n *NewTun) pidFromCheck(pid int32, name string) (ok bool) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	if n.CheckProcess == nil {
		return false
	}
	if _myPid == pid {
		return true
	}
	if n.CheckProcess == nil {
		return false
	}
	return n.CheckProcess(pid, name)
}

type NewTun struct {
	IsRunning         bool
	ProxyPort         uint16
	tun               io.ReadWriteCloser
	handleTCPCallback TcpFunc
	handleUDPCallback UdpFunc
	Sunny             Interface
	CheckProcess      func(int32, string) bool
}

func (n *NewTun) SetHandle(callbackTCP TcpFunc, udpSendReceiveFunc UdpFunc) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	n.handleTCPCallback = callbackTCP
	n.handleUDPCallback = udpSendReceiveFunc
}

func (n *NewTun) Send(packetData []byte) {
	_, _ = n.tun.Write(packetData)
}
func (n *NewTun) parsePacket(packetData []byte) {
	first := packetData[0] >> 4
	var packet gopacket.Packet
	if first == 4 {
		packet = gopacket.NewPacket(packetData, layers.LayerTypeIPv4, gopacket.Default)
	} else if first == 6 {
		packet = gopacket.NewPacket(packetData, layers.LayerTypeIPv6, gopacket.Default)
	} else {
		//	n.Send(packetData)
		return
	}
	if ipv4 := packet.Layer(layers.LayerTypeIPv4); ipv4 != nil {
		ip := ipv4.(*layers.IPv4)
		switch ip.Protocol {
		case layers.IPProtocolTCP:
			if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
				tcp := tcpLayer.(*layers.TCP)
				n.handleTCP4(ip, tcp)
			}
		case layers.IPProtocolUDP:
			if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
				udp := udpLayer.(*layers.UDP)
				n.handleUDP(ip.SrcIP, ip.DstIP, udp, true)
			}

		}
	} else if ipv6 := packet.Layer(layers.LayerTypeIPv6); ipv6 != nil {
		ip := ipv6.(*layers.IPv6)
		switch ip.NextHeader {
		case layers.IPProtocolTCP:
			if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
				tcp := tcpLayer.(*layers.TCP)
				n.handleTCP6(ip, tcp)
			}
		case layers.IPProtocolUDP:
			if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
				udp := udpLayer.(*layers.UDP)
				n.handleUDP(ip.SrcIP, ip.DstIP, udp, false)
			}

		}
	}
	return
}
