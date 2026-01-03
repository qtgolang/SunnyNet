//go:build windows
// +build windows

package WinDivert

import (
	"net"
	"os"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/tun/Tun"
)

var _myPid = int32(os.Getpid())

type Divert struct {
	handle, handle2 *Handle
	stopCh, stopCh2 chan struct{}
	handleMutex     sync.Mutex
	wg              sync.WaitGroup
	handleTCP       Tun.TcpFunc
	handleUDP       Tun.UdpFunc
	checkProcess    func(int32, string) bool
}

func NewWinDivert() *Divert {
	return &Divert{}
}
func (d *Divert) IsRunning() bool {
	return d.handle != nil
}
func (d *Divert) Close() {
	d.handleMutex.Lock()
	if d.handle != nil {
		_ = d.handle.Close()
	}
	if d.handle2 != nil {
		_ = d.handle2.Close()
	}
	d.handle = nil
	d.handle2 = nil
	close(d.stopCh)
	close(d.stopCh2)
	d.handleMutex.Unlock()
	d.wg.Wait()
}

const flowTcp = 6
const flowudp = 17

func (d *Divert) Run() bool {
	d.handleMutex.Lock()
	if d.handle != nil {
		d.handleMutex.Unlock()
		return true
	}
	if !d.runFlow() {
		return false
	}
	h, err := Open("true", LayerNetwork, 0, 0)
	if err != nil {
		d.handleMutex.Unlock()
		return false
	}
	d.wg.Add(1)
	d.handle, d.stopCh = h, make(chan struct{})
	d.handleMutex.Unlock()

	go func() {
		defer d.wg.Done()
		packetBuf := make([]byte, 0xffff)
		for {
			select {
			case <-d.stopCh:
				return
			default:
			}
			addr := &Address{}
			n, e := h.Recv(packetBuf, addr)
			if e != nil || n == 0 {
				continue
			}
			data := append([]byte(nil), packetBuf[:n]...)
			go func(bs []byte, a *Address) {
				if pkt := gopacket.NewPacket(bs, layers.LayerTypeIPv4, gopacket.Default); pkt.Layer(layers.LayerTypeIPv4) != nil {
					ip4 := pkt.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
					// TCP v4
					if tcp := pkt.Layer(layers.LayerTypeTCP); tcp != nil {
						if d.handleIPv4(h, bs, a, ip4, pkt) {
							return
						}
					}

					// UDP v4
					if udp := pkt.Layer(layers.LayerTypeUDP); udp != nil {
						if d.handleUDPv4(h, bs, a, ip4, pkt) {
							return
						}
					}
				}

				if pkt := gopacket.NewPacket(bs, layers.LayerTypeIPv6, gopacket.Default); pkt.Layer(layers.LayerTypeIPv6) != nil {
					ip6 := pkt.Layer(layers.LayerTypeIPv6).(*layers.IPv6)
					// TCP v6
					if tcp := pkt.Layer(layers.LayerTypeTCP); tcp != nil {
						if d.handleIPv6(h, bs, a, ip6, pkt) {
							return
						}
					}

					// UDP v6
					if udp := pkt.Layer(layers.LayerTypeUDP); udp != nil {
						if d.handleUDPv6(h, bs, a, ip6, pkt) {
							return
						}
					}
				}
				_, _ = h.Send(bs, a)
			}(data, addr.Clone())
		}
	}()
	return true
}
func (d *Divert) SetHandle(
	callbackTCP func(conn net.Conn),
	checkProcess func(int32, string) bool,
	udpSendReceiveFunc func(Type int, Theoni int64, pid uint32, LocalAddress, RemoteAddress string, data []byte) []byte) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	d.handleTCP = callbackTCP
	d.handleUDP = udpSendReceiveFunc
	d.checkProcess = checkProcess
}
