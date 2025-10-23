//go:build windows
// +build windows

package WinDivert

// #cgo CFLAGS: -I${SRCDIR}/divert -Wno-incompatible-pointer-types
// #include "windivert.h"
import "C"
import "github.com/qtgolang/SunnyNet/src/ProcessDrv/tun/WinDivert/divert"

func init() {
	divert.Init()
}

const (
	LayerNetwork        = Layer(C.WINDIVERT_LAYER_NETWORK)
	LayerNetworkForward = Layer(C.WINDIVERT_LAYER_NETWORK_FORWARD)
	LayerFlow           = Layer(C.WINDIVERT_LAYER_FLOW)
	LayerSocket         = Layer(C.WINDIVERT_LAYER_SOCKET)
	LayerReflect        = Layer(C.WINDIVERT_LAYER_REFLECT)
	//LayerEthernet       = Layer(C.WINDIVERT_LAYER_ETHERNET)
)

const (
	EventNetworkPacket   = Event(C.WINDIVERT_EVENT_NETWORK_PACKET)
	EventFlowEstablished = Event(C.WINDIVERT_EVENT_FLOW_ESTABLISHED)
	EventFlowDeleted     = Event(C.WINDIVERT_EVENT_FLOW_DELETED)
	EventSocketBind      = Event(C.WINDIVERT_EVENT_SOCKET_BIND)
	EventSocketConnect   = Event(C.WINDIVERT_EVENT_SOCKET_CONNECT)
	EventSocketListen    = Event(C.WINDIVERT_EVENT_SOCKET_LISTEN)
	EventSocketAccept    = Event(C.WINDIVERT_EVENT_SOCKET_ACCEPT)
	EventSocketClose     = Event(C.WINDIVERT_EVENT_SOCKET_CLOSE)
	EventReflectOpen     = Event(C.WINDIVERT_EVENT_REFLECT_OPEN)
	EventReflectClose    = Event(C.WINDIVERT_EVENT_REFLECT_CLOSE)
	//EventEthernetFrame   = Event(C.WINDIVERT_EVENT_ETHERNET_FRAME)
)

const (
	ShutdownRecv = Shutdown(C.WINDIVERT_SHUTDOWN_RECV)
	ShutdownSend = Shutdown(C.WINDIVERT_SHUTDOWN_SEND)
	ShutdownBoth = Shutdown(C.WINDIVERT_SHUTDOWN_BOTH)
)

const (
	QueueLength  = Param(C.WINDIVERT_PARAM_QUEUE_LENGTH)
	QueueTime    = Param(C.WINDIVERT_PARAM_QUEUE_TIME)
	QueueSize    = Param(C.WINDIVERT_PARAM_QUEUE_SIZE)
	VersionMajor = Param(C.WINDIVERT_PARAM_VERSION_MAJOR)
	VersionMinor = Param(C.WINDIVERT_PARAM_VERSION_MINOR)
)

const (
	FlagDefault   = uint64(0)
	FlagSniff     = uint64(C.WINDIVERT_FLAG_SNIFF)
	FlagDrop      = uint64(C.WINDIVERT_FLAG_DROP)
	FlagRecvOnly  = uint64(C.WINDIVERT_FLAG_RECV_ONLY)
	FlagSendOnly  = uint64(C.WINDIVERT_FLAG_SEND_ONLY)
	FlagNoInstall = uint64(C.WINDIVERT_FLAG_NO_INSTALL)
	FlagFragments = uint64(C.WINDIVERT_FLAG_FRAGMENTS)
)

const (
	PriorityDefault    = int16(0)
	PriorityHighest    = int16(C.WINDIVERT_PRIORITY_HIGHEST)
	PriorityLowest     = int16(C.WINDIVERT_PRIORITY_LOWEST)
	QueueLengthDefault = uint64(C.WINDIVERT_PARAM_QUEUE_LENGTH_DEFAULT)
	QueueLengthMin     = uint64(C.WINDIVERT_PARAM_QUEUE_LENGTH_MIN)
	QueueLengthMax     = uint64(C.WINDIVERT_PARAM_QUEUE_LENGTH_MAX)
	QueueTimeDefault   = uint64(C.WINDIVERT_PARAM_QUEUE_TIME_DEFAULT)
	QueueTimeMin       = uint64(C.WINDIVERT_PARAM_QUEUE_TIME_MIN)
	QueueTimeMax       = uint64(C.WINDIVERT_PARAM_QUEUE_TIME_MAX)
	QueueSizeDefault   = uint64(C.WINDIVERT_PARAM_QUEUE_SIZE_DEFAULT)
	QueueSizeMin       = uint64(C.WINDIVERT_PARAM_QUEUE_SIZE_MIN)
	QueueSizeMax       = uint64(C.WINDIVERT_PARAM_QUEUE_SIZE_MAX)
)

const (
	ChecksumDefault  = uint64(0)
	NoIPChecksum     = uint64(C.WINDIVERT_HELPER_NO_IP_CHECKSUM)
	NoICMPChecksum   = uint64(C.WINDIVERT_HELPER_NO_ICMP_CHECKSUM)
	NoICMPV6Checksum = uint64(C.WINDIVERT_HELPER_NO_ICMPV6_CHECKSUM)
	NoTCPChecksum    = uint64(C.WINDIVERT_HELPER_NO_TCP_CHECKSUM)
	NoUDPChecksum    = uint64(C.WINDIVERT_HELPER_NO_UDP_CHECKSUM)
)

const (
	BatchMax = int(C.WINDIVERT_BATCH_MAX)
	MTUMax   = int(C.WINDIVERT_MTU_MAX)
)
