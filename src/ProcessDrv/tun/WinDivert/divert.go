//go:build windows
// +build windows

package WinDivert

import (
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var once = sync.Once{}

// GerVersionInfo is ...
func GetVersionInfo() (ver string, err error) {
	h, err := Open("false", LayerNetwork, PriorityDefault, FlagDefault)
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
}

func ioControlEx(h windows.Handle, code CtlCode, ioctl unsafe.Pointer, buf *byte, bufLen uint32, overlapped *windows.Overlapped) (iolen uint32, err error) {
	err = windows.DeviceIoControl(h, uint32(code), (*byte)(ioctl), uint32(unsafe.Sizeof(ioCtl{})), buf, bufLen, &iolen, overlapped)
	if err != windows.ERROR_IO_PENDING {
		return
	}

	err = windows.GetOverlappedResult(h, overlapped, &iolen, true)

	return
}

func ioControl(h windows.Handle, code CtlCode, ioctl unsafe.Pointer, buf *byte, bufLen uint32) (iolen uint32, err error) {
	event, _ := windows.CreateEvent(nil, 0, 0, nil)

	overlapped := windows.Overlapped{
		HEvent: event,
	}

	iolen, err = ioControlEx(h, code, ioctl, buf, bufLen, &overlapped)

	windows.CloseHandle(event)
	return
}

// Handle is ...
type Handle struct {
	sync.Mutex
	windows.Handle
	rOverlapped windows.Overlapped
	wOverlapped windows.Overlapped
}

// Recv is ...
func (h *Handle) Recv(buffer []byte, address *Address) (uint, error) {
	addrLen := uint(unsafe.Sizeof(Address{}))
	recv := recv{
		Addr:       uint64(uintptr(unsafe.Pointer(address))),
		AddrLenPtr: uint64(uintptr(unsafe.Pointer(&addrLen))),
	}

	iolen, err := ioControlEx(h.Handle, ioCtlRecv, unsafe.Pointer(&recv), &buffer[0], uint32(len(buffer)), &h.rOverlapped)
	if err != nil {
		return uint(iolen), Error(err.(windows.Errno))
	}

	return uint(iolen), nil
}

// RecvEx is ...
func (h *Handle) RecvEx(buffer []byte, address []Address) (uint, uint, error) {
	addrLen := uint(len(address)) * uint(unsafe.Sizeof(Address{}))
	recv := recv{
		Addr:       uint64(uintptr(unsafe.Pointer(&address[0]))),
		AddrLenPtr: uint64(uintptr(unsafe.Pointer(&addrLen))),
	}

	iolen, err := ioControlEx(h.Handle, ioCtlRecv, unsafe.Pointer(&recv), &buffer[0], uint32(len(buffer)), &h.rOverlapped)
	if err != nil {
		return uint(iolen), addrLen / uint(unsafe.Sizeof(Address{})), Error(err.(windows.Errno))
	}

	return uint(iolen), addrLen / uint(unsafe.Sizeof(Address{})), nil
}

// Send is ...
func (h *Handle) Send(buffer []byte, address *Address) (uint, error) {
	send := send{
		Addr:    uint64(uintptr(unsafe.Pointer(address))),
		AddrLen: uint64(unsafe.Sizeof(Address{})),
	}

	iolen, err := ioControlEx(h.Handle, ioCtlSend, unsafe.Pointer(&send), &buffer[0], uint32(len(buffer)), &h.wOverlapped)
	if err != nil {
		return uint(iolen), Error(err.(windows.Errno))
	}

	return uint(iolen), nil
}

// SendEx is ...
func (h *Handle) SendEx(buffer []byte, address []Address) (uint, error) {
	send := send{
		Addr:    uint64(uintptr(unsafe.Pointer(&address[0]))),
		AddrLen: uint64(unsafe.Sizeof(Address{})) * uint64(len(address)),
	}

	iolen, err := ioControlEx(h.Handle, ioCtlSend, unsafe.Pointer(&send), &buffer[0], uint32(len(buffer)), &h.wOverlapped)
	if err != nil {
		return uint(iolen), Error(err.(windows.Errno))
	}

	return uint(iolen), nil
}

// Shutdown is ...
func (h *Handle) Shutdown(how Shutdown) error {
	shutdown := shutdown{
		How: uint32(how),
	}

	_, err := ioControl(h.Handle, ioCtlShutdown, unsafe.Pointer(&shutdown), nil, 0)
	if err != nil {
		return Error(err.(windows.Errno))
	}

	return nil
}

// Close is ...
func (h *Handle) Close() error {
	windows.CloseHandle(h.rOverlapped.HEvent)
	windows.CloseHandle(h.wOverlapped.HEvent)

	err := windows.CloseHandle(h.Handle)
	if err != nil {
		return Error(err.(windows.Errno))
	}

	return nil
}

// GetParam is ...
func (h *Handle) GetParam(p Param) (uint64, error) {
	getParam := getParam{
		Param: uint32(p),
		Value: 0,
	}

	_, err := ioControl(h.Handle, ioCtlGetParam, unsafe.Pointer(&getParam), (*byte)(unsafe.Pointer(&getParam.Value)), uint32(unsafe.Sizeof(getParam.Value)))
	if err != nil {
		return getParam.Value, Error(err.(windows.Errno))
	}

	return getParam.Value, nil
}

// SetParam is ...
func (h *Handle) SetParam(p Param, v uint64) error {
	switch p {
	case QueueLength:
		if v < QueueLengthMin || v > QueueLengthMax {
			return errQueueLength
		}
	case QueueTime:
		if v < QueueTimeMin || v > QueueTimeMax {
			return errQueueTime
		}
	case QueueSize:
		if v < QueueSizeMin || v > QueueSizeMax {
			return errQueueSize
		}
	default:
		return errQueueParam
	}

	setParam := setParam{
		Value: v,
		Param: uint32(p),
	}

	_, err := ioControl(h.Handle, ioCtlSetParam, unsafe.Pointer(&setParam), nil, 0)
	if err != nil {
		return Error(err.(windows.Errno))
	}

	return nil
}
