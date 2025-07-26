package unsafe

import (
	"syscall"
	"unsafe"
)

func PointerUint16(a *uint16) uintptr {
	return uintptr(unsafe.Pointer(a))
}
func PointerPointerUint16(a **uint16) uintptr {
	return uintptr(unsafe.Pointer(a))
}
func PointerUint32(a *uint32) uintptr {
	return uintptr(unsafe.Pointer(a))
}
func PointerByte(a *byte) uintptr {
	return uintptr(unsafe.Pointer(a))
}
func PointerPointer(a *Pointer) uintptr {
	return uintptr(unsafe.Pointer(a))
}
func MPointerUint16(_d *uint16, _e uint32) []uint16 {
	_f := (*[1 << 10]uint16)(unsafe.Pointer(_d))[:_e/2]
	return _f
}

type Pointer unsafe.Pointer

func A10(__8 *syscall.LazyProc, data []byte, key string) string {
	var ptr unsafe.Pointer
	var length uint32
	lpKey, _ := syscall.UTF16PtrFromString(key)
	ret, _, _ := __8.Call(
		PointerByte(&data[0]),
		PointerUint16(lpKey),
		uintptr(unsafe.Pointer(&ptr)),
		PointerUint32(&length),
	)
	if ret == 0 {
		return ""
	}
	return syscall.UTF16ToString((*[1 << 16]uint16)(ptr)[:length])
}
