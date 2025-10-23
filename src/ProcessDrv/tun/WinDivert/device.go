//go:build windows
// +build windows

package WinDivert

type CtlCode uint32

const (
	METHOD_IN_DIRECT  = 1
	METHOD_OUT_DIRECT = 2
)

const (
	FILE_READ_DATA  = 1
	FILE_WRITE_DATA = 2
)

const FILE_DEVICE_NETWORK = 0x00000012

const (
	ioCtlInitialize = CtlCode(((FILE_DEVICE_NETWORK) << 16) | ((FILE_READ_DATA | FILE_WRITE_DATA) << 14) | ((0x921) << 2) | (METHOD_OUT_DIRECT))
	ioCtlStartup    = CtlCode(((FILE_DEVICE_NETWORK) << 16) | ((FILE_READ_DATA | FILE_WRITE_DATA) << 14) | ((0x922) << 2) | (METHOD_IN_DIRECT))
	ioCtlRecv       = CtlCode(((FILE_DEVICE_NETWORK) << 16) | ((FILE_READ_DATA) << 14) | ((0x923) << 2) | (METHOD_OUT_DIRECT))
	ioCtlSend       = CtlCode(((FILE_DEVICE_NETWORK) << 16) | ((FILE_READ_DATA | FILE_WRITE_DATA) << 14) | ((0x924) << 2) | (METHOD_IN_DIRECT))
	ioCtlSetParam   = CtlCode(((FILE_DEVICE_NETWORK) << 16) | ((FILE_READ_DATA | FILE_WRITE_DATA) << 14) | ((0x925) << 2) | (METHOD_IN_DIRECT))
	ioCtlGetParam   = CtlCode(((FILE_DEVICE_NETWORK) << 16) | ((FILE_READ_DATA) << 14) | ((0x926) << 2) | (METHOD_OUT_DIRECT))
	ioCtlShutdown   = CtlCode(((FILE_DEVICE_NETWORK) << 16) | ((FILE_READ_DATA | FILE_WRITE_DATA) << 14) | ((0x927) << 2) | (METHOD_IN_DIRECT))
)

func (c CtlCode) String() string {
	switch c {
	case ioCtlInitialize:
		return "IOCTL_WINDIVERT_INITIALIZE"
	case ioCtlStartup:
		return "IOCTL_WINDIVERT_STARTUP"
	case ioCtlRecv:
		return "IOCTL_WINDIVERT_RECV"
	case ioCtlSend:
		return "IOCTL_WINDIVERT_SEND"
	case ioCtlSetParam:
		return "IOCTL_WINDIVERT_SET_PARAM"
	case ioCtlGetParam:
		return "IOCTL_WINDIVERT_GET_PARAM"
	case ioCtlShutdown:
		return "IOCTL_WINDIVERT_SHUTDOWN"
	default:
		return ""
	}
}

type ioCtl struct {
	b1, b2, b3, b4 uint32
}

type recv struct {
	Addr       uint64
	AddrLenPtr uint64
}

type send struct {
	Addr    uint64
	AddrLen uint64
}

type initialize struct {
	Layer    uint32
	Priority uint32
	Flags    uint64
}

type startup struct {
	Flags uint64
	_     uint64
}

type shutdown struct {
	How uint32
	_   uint32
	_   uint64
}

type getParam struct {
	Param uint32
	_     uint32
	Value uint64
}

type setParam struct {
	Value uint64
	Param uint32
	_     uint32
}
