//go:build windows
// +build windows

package WinDivert

import (
	"bytes"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/SunnyNetUDP"
	"github.com/qtgolang/SunnyNet/src/public"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
)

type expiryUDP struct {
	pid                int32
	name               string
	Theology           int64
	v4                 bool
	h                  *Handle
	addr               *Address
	SrcPort, DstPort   layers.UDPPort
	clientIP, serverIP net.IP
}

func (e expiryUDP) ToClient(i []byte) bool {
	return e.send(i, false)
}

func (e expiryUDP) ToServer(i []byte) bool {
	return e.send(i, true)
}

// toServer = true  表示发往 server (client -> server)
// toServer = false 表示发往 client (server -> client)
func (e expiryUDP) send(payload []byte, toServer bool) bool {
	var srcPort, dstPort layers.UDPPort
	var srcIP, dstIP net.IP
	if toServer {
		srcPort = e.SrcPort
		dstPort = e.DstPort
		srcIP = e.clientIP
		dstIP = e.serverIP
	} else {
		srcPort = e.DstPort
		dstPort = e.SrcPort
		srcIP = e.serverIP
		dstIP = e.clientIP
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	newUDP := &layers.UDP{SrcPort: srcPort, DstPort: dstPort}
	if e.v4 {
		ip := &layers.IPv4{
			Version:  4,
			IHL:      5,
			TTL:      64,
			Protocol: layers.IPProtocolUDP,
			SrcIP:    srcIP,
			DstIP:    dstIP,
		}
		_ = newUDP.SetNetworkLayerForChecksum(ip)
		if err := gopacket.SerializeLayers(buf, opts, ip, newUDP, gopacket.Payload(payload)); err != nil {
			return false
		}
	} else {
		ip6 := &layers.IPv6{
			Version:    6,
			HopLimit:   64,
			NextHeader: layers.IPProtocolUDP,
			SrcIP:      srcIP,
			DstIP:      dstIP,
		}
		_ = newUDP.SetNetworkLayerForChecksum(ip6)
		if err := gopacket.SerializeLayers(buf, opts, ip6, newUDP, gopacket.Payload(payload)); err != nil {
			return false
		}
	}
	a := e.addr.Clone()
	a.SetOutbound(toServer)
	_, err := e.h.Send(buf.Bytes(), a)
	if err != nil {
		return false
	}
	return true
}

var (
	udpCache   = make(map[uint16]*expiryUDP)
	udpCacheMu sync.RWMutex
)

func (d *Divert) handleCommandUDP(h *Handle, data []byte, addr *Address, udp *layers.UDP, clientIP, serverIP net.IP, clientPort, serverPort uint16, v4 bool, pkt gopacket.Packet) {
	payload := udp.LayerPayload()
	if len(payload) == 0 {
		// 没有数据直接转发
		_, _ = h.Send(data, addr)
		return
	}
	var obj *expiryUDP
	{
		var ok bool
		var port uint16
		udpCacheMu.Lock()
		if addr.Outbound() {
			port = clientPort
		} else {
			port = serverPort
		}
		obj, ok = udpCache[port]
		if !ok {
			pid, name := getPidByPort("udp", port)
			if d.pidFromCheck(pid, name) {
				udpCacheMu.Unlock()
				_, _ = h.Send(data, addr)
				return
			}
			obj = &expiryUDP{pid: pid, name: name, v4: v4, h: h, addr: addr, Theology: atomic.AddInt64(&public.Theology, 1)}
			if addr.Outbound() {
				obj.clientIP, obj.serverIP = clientIP, serverIP
				obj.SrcPort, obj.DstPort = udp.SrcPort, udp.DstPort
			} else {
				obj.clientIP, obj.serverIP = serverIP, clientIP
				obj.SrcPort, obj.DstPort = udp.DstPort, udp.SrcPort
			}
			udpCache[port] = obj
		} else {
			/*
				if clientPort == 53 || serverPort == 53 {
					if d.handleDNS53(h, data, addr, clientIP, serverIP, false, pkt) {
						return
					}
				}
			*/
		}
		udpCacheMu.Unlock()
		if obj == nil {
			_, _ = h.Send(data, addr)
			return
		}
		SunnyNetUDP.AddUDPItem(obj.Theology, obj)
	}
	sessionsMu.Lock()
	call := d.handleUDP
	sessionsMu.Unlock()
	if call != nil {
		LocalAddress := net.JoinHostPort(clientIP.String(), strconv.Itoa(int(clientPort)))
		RemoteAddress := net.JoinHostPort(serverIP.String(), strconv.Itoa(int(serverPort)))
		var bs []byte
		if addr.Outbound() {
			bs = call(public.SunnyNetUDPTypeSend, obj.Theology, uint32(obj.pid), LocalAddress, RemoteAddress, payload)
		} else {
			bs = call(public.SunnyNetUDPTypeReceive, obj.Theology, uint32(obj.pid), RemoteAddress, LocalAddress, payload)
		}
		if len(bs) > 0 {
			if bytes.Equal(bs, payload) {
				//未作修改
				_, _ = h.Send(data, addr)
				return
			}
			buf := gopacket.NewSerializeBuffer()
			opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
			newUDP := &layers.UDP{SrcPort: udp.SrcPort, DstPort: udp.DstPort}
			var layerIP gopacket.SerializableLayer
			if v4 {
				ip := &layers.IPv4{Version: 4, IHL: 5, TTL: 64, Protocol: layers.IPProtocolUDP}
				ip.SrcIP, ip.DstIP = clientIP, serverIP
				layerIP = ip
				_ = newUDP.SetNetworkLayerForChecksum(ip)
			} else {
				ip6 := &layers.IPv6{Version: 6, HopLimit: 64, NextHeader: layers.IPProtocolUDP}
				ip6.SrcIP, ip6.DstIP = clientIP, serverIP
				layerIP = ip6
				_ = newUDP.SetNetworkLayerForChecksum(ip6)
			}
			_ = gopacket.SerializeLayers(buf, opts, layerIP, newUDP, gopacket.Payload(bs))
			_, _ = h.Send(buf.Bytes(), addr.Clone())
		}
		return
	}
	_, _ = h.Send(data, addr)
	return
}
func (d *Divert) handleUDPv4(h *Handle, data []byte, addr *Address, ip4 *layers.IPv4, pkt gopacket.Packet) bool {
	udpLayer := pkt.Layer(layers.LayerTypeUDP)
	if udpLayer == nil {
		return false
	}
	udp := udpLayer.(*layers.UDP)
	clientIP := ip4.SrcIP
	clientPort := uint16(udp.SrcPort)
	serverIP := ip4.DstIP
	serverPort := uint16(udp.DstPort)
	d.handleCommandUDP(h, data, addr, udp, clientIP, serverIP, clientPort, serverPort, true, pkt)
	return true
}
func (d *Divert) handleUDPv6(h *Handle, data []byte, addr *Address, ip6 *layers.IPv6, pkt gopacket.Packet) bool {
	udpLayer := pkt.Layer(layers.LayerTypeUDP)
	if udpLayer == nil {
		return false
	}
	udp := udpLayer.(*layers.UDP)
	clientIP := ip6.SrcIP
	clientPort := uint16(udp.SrcPort)
	serverIP := ip6.DstIP
	serverPort := uint16(udp.DstPort)
	d.handleCommandUDP(h, data, addr, udp, clientIP, serverIP, clientPort, serverPort, false, pkt)
	return true
}

func (d *Divert) runFlow() bool {
	h, e := Open("true", LayerFlow, 0, FlagSniff|FlagRecvOnly)
	if e != nil {
		d.handleMutex.Unlock()
		return false
	}
	d.wg.Add(1)
	d.handle2, d.stopCh2 = h, make(chan struct{})
	go func() {
		defer d.wg.Done()
		packetBuf := make([]byte, 0xffff)
		for {
			select {
			case <-d.stopCh2:
				return
			default:
			}
			addr := &Address{}
			_, err := h.Recv(packetBuf, addr)
			if err != nil {
				continue
			}
			Protocol := addr.Flow().Protocol
			if Protocol == flowTcp {
				continue
			}
			if Protocol == flowudp {
				if addr.Event() == 1 {
					//fmt.Println("udp建立连接", addr.Flow().ProcessID, addr.IPv6(), LocalAddress, LocalAddressPort, RemoteAddress, RemoteAddressPort)
				} else {
					var LocalAddress, RemoteAddress string
					LocalAddressPort := addr.Flow().LocalPort
					RemoteAddressPort := addr.Flow().RemotePort
					if addr.IPv6() {
						LocalAddress = fmt.Sprintf("[%s]:%d", flowAddrToIP(addr.Flow().LocalAddress), LocalAddressPort)
						RemoteAddress = fmt.Sprintf("[%s]:%d", flowAddrToIP(addr.Flow().RemoteAddress), RemoteAddressPort)
					} else {
						LocalAddress = fmt.Sprintf("%s:%d", flowAddrToIP(addr.Flow().LocalAddress), LocalAddressPort)
						RemoteAddress = fmt.Sprintf("%s:%d", flowAddrToIP(addr.Flow().RemoteAddress), RemoteAddressPort)
					}
					udpCacheMu.Lock()
					obj, ok := udpCache[LocalAddressPort]
					if !ok {
						udpCacheMu.Unlock()
						continue
					}
					delete(udpCache, LocalAddressPort)
					udpCacheMu.Unlock()
					sessionsMu.Lock()
					call := d.handleUDP
					sessionsMu.Unlock()
					if call != nil {
						call(public.SunnyNetUDPTypeClosed, obj.Theology, uint32(obj.pid), LocalAddress, RemoteAddress, nil)
					}
					SunnyNetUDP.DelUDPItem(obj.Theology)
				}
			}
		}
	}()
	return true
}
func flowAddrToIP(addr [16]uint8) net.IP {
	out := make([]byte, 16)
	for i := 0; i < 16; i += 1 {
		out[i] = addr[15-i]
	}
	if addr[0] == 0 && addr[1] == 0 && addr[2] == 0 && addr[3] == 0 &&
		addr[4] == 0 && addr[5] == 0 && addr[6] == 0 && addr[7] == 0 &&
		addr[8] == 0 && addr[9] == 0 {
		return net.IPv4(addr[12], addr[13], addr[14], addr[15])
	}
	return out
}
