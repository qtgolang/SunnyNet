//go:build windows
// +build windows

package WinDivert

// #cgo CFLAGS: -I${SRCDIR}/divert -Wno-incompatible-pointer-types
// #define WINDIVERTEXPORT static
// #include "windivert.c"
import "C"

import (
	"encoding/binary"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Open is ...
func Open(filter string, layer Layer, priority int16, flags uint64) (h *Handle, err error) {
	once.Do(func() {
		vers := map[string]struct{}{
			"2.0": {},
			"2.1": {},
			"2.2": {},
		}
		ver, er := func() (ver string, err error) {
			h, err := open("false", LayerNetwork, PriorityDefault, FlagDefault)
			if err != nil {
				return
			}
			defer func() {
				err = h.Close()
			}()

			major, err := h.GetParam(VersionMajor)
			if err != nil {
				return
			}

			minor, err := h.GetParam(VersionMinor)
			if err != nil {
				return
			}

			ver = strings.Join([]string{strconv.Itoa(int(major)), strconv.Itoa(int(minor))}, ".")
			return
		}()
		if er != nil {
			err = er
			return
		}
		if _, ok := vers[ver]; !ok {
			err = fmt.Errorf("unsupported windivert version: %v", ver)
		}
	})
	if err != nil {
		return
	}

	return open(filter, layer, priority, flags)
}

func open(filter string, layer Layer, priority int16, flags uint64) (h *Handle, err error) {
	if priority < PriorityLowest || priority > PriorityHighest {
		return nil, errPriority
	}

	runtime.LockOSThread()
	hd := C.WinDivertOpen(C.CString(filter), C.WINDIVERT_LAYER(layer), C.int16_t(priority), C.uint64_t(flags))
	runtime.UnlockOSThread()

	if hd == C.HANDLE(C.INVALID_HANDLE_VALUE) {
		return nil, Error(C.GetLastError())
	}

	rEvent, _ := windows.CreateEvent(nil, 0, 0, nil)
	wEvent, _ := windows.CreateEvent(nil, 0, 0, nil)

	return &Handle{
		Mutex:  sync.Mutex{},
		Handle: windows.Handle(uintptr(unsafe.Pointer(hd))),
		rOverlapped: windows.Overlapped{
			HEvent: rEvent,
		},
		wOverlapped: windows.Overlapped{
			HEvent: wEvent,
		},
	}, nil
}

// CalcChecksums is ...
func CalcChecksums(buffer []byte, address *Address, flags uint64) bool {
	re := C.WinDivertHelperCalcChecksums(unsafe.Pointer(&buffer[0]), C.UINT(len(buffer)), (*C.WINDIVERT_ADDRESS)(unsafe.Pointer(address)), C.uint64_t(flags))
	return re == C.TRUE
}

// IPv4 header
type IPv4Header struct {
	VersionIHL      uint8
	TOS             uint8
	TotalLength     uint16
	ID              uint16
	FlagsFragOffset uint16
	TTL             uint8
	Protocol        uint8
	Checksum        uint16
	SrcAddr         [4]byte
	DstAddr         [4]byte
	SrcPort         uint16
	DstPort         uint16
	Name            string
	V4              bool
}

// IPv6 header
type IPv6Header struct {
	VersionTCFlow uint32
	PayloadLength uint16
	NextHeader    uint8
	HopLimit      uint8
	SrcAddr       [16]byte
	DstAddr       [16]byte
	SrcPort       uint16
	DstPort       uint16
	Name          string
	V4            bool
}
type DataPacket interface {
	String() string
}

func (v6 IPv6Header) String() string {
	srcIP := net.IP(v6.SrcAddr[:])
	dstIP := net.IP(v6.DstAddr[:])
	return fmt.Sprintf("[%s]%s:%d->%s:%d", v6.Name, srcIP.String(), v6.SrcPort, dstIP.String(), v6.DstPort)
}
func (v4 IPv4Header) String() string {
	srcIP := net.IP(v4.SrcAddr[:])
	dstIP := net.IP(v4.DstAddr[:])
	return fmt.Sprintf("[%s]%s:%d->%s:%d", v4.Name, srcIP.String(), v4.SrcPort, dstIP.String(), v4.DstPort)
}
func ParsePacket(data []byte) DataPacket {
	if len(data) < 1 {
		return nil
	}
	version := data[0] >> 4
	if version == 4 {
		if len(data) < 20 {
			return nil
		}
		ip := IPv4Header{
			VersionIHL:      data[0],
			TOS:             data[1],
			TotalLength:     binary.BigEndian.Uint16(data[2:4]),
			ID:              binary.BigEndian.Uint16(data[4:6]),
			FlagsFragOffset: binary.BigEndian.Uint16(data[6:8]),
			TTL:             data[8],
			Protocol:        data[9],
			Checksum:        binary.BigEndian.Uint16(data[10:12]),
		}
		copy(ip.SrcAddr[:], data[12:16])
		copy(ip.DstAddr[:], data[16:20])
		ip.V4 = true
		offset := int((ip.VersionIHL & 0x0F) * 4)
		if ip.Protocol == 6 && len(data) >= offset+20 {
			ip.Name = "TCP"
			ip.SrcPort = binary.BigEndian.Uint16(data[offset : offset+2])
			ip.DstPort = binary.BigEndian.Uint16(data[offset+2 : offset+4])
		} else if ip.Protocol == 17 && len(data) >= offset+8 {
			ip.Name = "UDP"
			ip.SrcPort = binary.BigEndian.Uint16(data[offset : offset+2])
			ip.DstPort = binary.BigEndian.Uint16(data[offset+2 : offset+4])
		}
		return &ip
	} else if version == 6 {
		// IPv6
		if len(data) < 40 {
			return nil
		}
		ip := IPv6Header{
			VersionTCFlow: binary.BigEndian.Uint32(data[0:4]),
			PayloadLength: binary.BigEndian.Uint16(data[4:6]),
			NextHeader:    data[6],
			HopLimit:      data[7],
		}
		copy(ip.SrcAddr[:], data[8:24])
		copy(ip.DstAddr[:], data[24:40])
		ip.V4 = false
		offset := 40
		if ip.NextHeader == 6 && len(data) >= offset+20 {
			ip.Name = "TCP"
			// TCP
			ip.SrcPort = binary.BigEndian.Uint16(data[offset : offset+2])
			ip.DstPort = binary.BigEndian.Uint16(data[offset+2 : offset+4])
		} else if ip.NextHeader == 17 && len(data) >= offset+8 {
			ip.Name = "UDP"
			ip.SrcPort = binary.BigEndian.Uint16(data[offset : offset+2])
			ip.DstPort = binary.BigEndian.Uint16(data[offset+2 : offset+4])
		}
		return &ip
	}
	return nil
}
