//go:build windows
// +build windows

package WinDivert

import (
	"unsafe"
)

// Ethernet is ...
type Ethernet struct {
	InterfaceIndex    uint32
	SubInterfaceIndex uint32
	_                 [7]uint64
}

// Network is ...
// The WINDIVERT_LAYER_NETWORK and WINDIVERT_LAYER_NETWORK_FORWARD layers allow the user
// application to capture/block/inject network packets passing to/from (and through) the
// local machine. Due to technical limitations, process ID information is not available
// at these layers.
type Network struct {
	InterfaceIndex    uint32
	SubInterfaceIndex uint32
	_                 [7]uint64
}

// Socket is ...
// The WINDIVERT_LAYER_SOCKET layer can capture or block events corresponding to socket
// operations, such as bind(), connect(), listen(), etc., or the termination of socket
// operations, such as a TCP socket disconnection. Unlike the flow layer, most socket-related
// events can be blocked. However, it is not possible to inject new or modified socket events.
// Process ID information (of the process responsible for the socket operation) is available
// at this layer. Due to technical limitations, this layer cannot capture events that occurred
// before the handle was opened.
type Socket struct {
	EndpointID       uint64
	ParentEndpointID uint64
	ProcessID        uint32
	LocalAddress     [16]uint8
	RemoteAddress    [16]uint8
	LocalPort        uint16
	RemotePort       uint16
	Protocol         uint8
	_                [3]uint8
	_                uint32
}

// Flow is ...
// The WINDIVERT_LAYER_FLOW layer captures information about network flow establishment/deletion
// events. Here, a flow represents either (1) a TCP connection, or (2) an implicit "flow" created
// by the first sent/received packet for non-TCP traffic, e.g., UDP. Old flows are deleted when
// the corresponding connection is closed (for TCP), or based on an activity timeout (non-TCP).
// Flow-related events can be captured, but not blocked nor injected. Process ID information is
// also available at this layer. Due to technical limitations, the WINDIVERT_LAYER_FLOW layer
// cannot capture flow events that occurred before the handle was opened.
type Flow struct {
	EndpointID       uint64
	ParentEndpointID uint64
	ProcessID        uint32
	LocalAddress     [16]uint8
	RemoteAddress    [16]uint8
	LocalPort        uint16
	RemotePort       uint16
	Protocol         uint8
	_                [3]uint8
	_                uint32
}

// Reflect is ...
// Finally, the WINDIVERT_LAYER_REFLECT layer can capture events relating to WinDivert itself,
// such as when another process opens a new WinDivert handle, or closes an old WinDivert handle.
// WinDivert events can be captured but not injected nor blocked. Process ID information
// (of the process responsible for opening the WinDivert handle) is available at this layer.
// This layer also returns data in the form of an "object" representation of the filter string
// used to open the handle. The object representation can be converted back into a human-readable
// filter string using the WinDivertHelperFormatFilter() function. This layer can also capture
// events that occurred before the handle was opened. This layer cannot capture events related
// to other WINDIVERT_LAYER_REFLECT-layer handles.
type Reflect struct {
	TimeStamp int64
	ProcessID uint32
	layer     uint32
	Flags     uint64
	Priority  int16
	_         int16
	_         int32
	_         [4]uint64
}

// Layer is ...
func (r *Reflect) Layer() Layer {
	return Layer(r.layer)
}

// Address is ...
type Address struct {
	Timestamp int64
	Bitfield  uint32 // Layer/Event/Sniffed/Outbound/Loopback/Impostor/IPv6/Checksums/Reserved1
	Reserved2 uint32 // 4 字节
	union     [64]uint8
}

func (a *Address) Clone() *Address {
	b := new(Address)
	b.Timestamp = a.Timestamp
	b.Bitfield = a.Bitfield
	b.Reserved2 = a.Reserved2
	b.union = a.union
	return b
}

func (a *Address) Layer() uint8 {
	return uint8(a.Bitfield & 0xFF)
}

func (a *Address) Event() uint8 {
	return uint8((a.Bitfield >> 8) & 0xFF)
}

func (a *Address) Sniffed() bool {
	return (a.Bitfield>>16)&1 == 1
}

func (a *Address) Outbound() bool {
	return (a.Bitfield>>17)&1 == 1
}

func (a *Address) Loopback() bool {
	return (a.Bitfield>>18)&1 == 1
}

func (a *Address) Impostor() bool {
	return (a.Bitfield>>19)&1 == 1
}

func (a *Address) IPv6() bool {
	return (a.Bitfield>>20)&1 == 1
}

func (a *Address) IPChecksum() bool {
	return (a.Bitfield>>21)&1 == 1
}

func (a *Address) TCPChecksum() bool {
	return (a.Bitfield>>22)&1 == 1
}

func (a *Address) UDPChecksum() bool {
	return (a.Bitfield>>23)&1 == 1
}

func (a *Address) SetLayer(v uint8) {
	a.Bitfield = (a.Bitfield &^ 0xFF) | uint32(v)
}

func (a *Address) SetEvent(v uint8) {
	a.Bitfield = (a.Bitfield &^ (0xFF << 8)) | (uint32(v) << 8)
}

func (a *Address) SetSniffed(v bool) {
	if v {
		a.Bitfield |= 1 << 16
	} else {
		a.Bitfield &^= 1 << 16
	}
}

func (a *Address) SetOutbound(v bool) {
	if v {
		a.Bitfield |= 1 << 17
	} else {
		a.Bitfield &^= 1 << 17
	}
}

func (a *Address) SetLoopback(v bool) {
	if v {
		a.Bitfield |= 1 << 18
	} else {
		a.Bitfield &^= 1 << 18
	}
}

func (a *Address) SetImpostor(v bool) {
	if v {
		a.Bitfield |= 1 << 19
	} else {
		a.Bitfield &^= 1 << 19
	}
}

func (a *Address) SetIPv6(v bool) {
	if v {
		a.Bitfield |= 1 << 20
	} else {
		a.Bitfield &^= 1 << 20
	}
}

func (a *Address) SetIPChecksum(v bool) {
	if v {
		a.Bitfield |= 1 << 21
	} else {
		a.Bitfield &^= 1 << 21
	}
}

func (a *Address) SetTCPChecksum(v bool) {
	if v {
		a.Bitfield |= 1 << 22
	} else {
		a.Bitfield &^= 1 << 22
	}
}

func (a *Address) SetUDPChecksum(v bool) {
	if v {
		a.Bitfield |= 1 << 23
	} else {
		a.Bitfield &^= 1 << 23
	}
}

// Ethernet is ...
func (a *Address) Ethernet() *Ethernet {
	return (*Ethernet)(unsafe.Pointer(&a.union))
}

// Network is ...
func (a *Address) Network() *Network {
	return (*Network)(unsafe.Pointer(&a.union))
}

// Socket is ...
func (a *Address) Socket() *Socket {
	return (*Socket)(unsafe.Pointer(&a.union))
}

// Flow is ...
func (a *Address) Flow() *Flow {
	return (*Flow)(unsafe.Pointer(&a.union))
}

// Reflect is ...
func (a *Address) Reflect() *Reflect {
	return (*Reflect)(unsafe.Pointer(&a.union))
}
