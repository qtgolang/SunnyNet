//go:build windows
// +build windows

package Call

/*
#include <stdlib.h>
*/
import "C"
import (
	"syscall"
	"unsafe"
)

func Call(address int, arg ...interface{}) int {
	if address == 0 {
		return 0
	}

	var args []uintptr
	var frees []unsafe.Pointer

	defer func() {
		for _, p := range frees {
			C.free(p)
		}
	}()

	for _, v := range arg {
		switch val := v.(type) {
		case uintptr:
			args = append(args, val)
		case unsafe.Pointer:
			args = append(args, uintptr(val))
		case nil:
			args = append(args, 0)
		case int:
			args = append(args, uintptr(val))
		case int8:
			args = append(args, uintptr(val))
		case int16:
			args = append(args, uintptr(val))
		case int32:
			args = append(args, uintptr(val))
		case int64:
			args = append(args, uintptr(val))
		case uint:
			args = append(args, uintptr(val))
		case uint8:
			args = append(args, uintptr(val))
		case uint16:
			args = append(args, uintptr(val))
		case uint32:
			args = append(args, uintptr(val))
		case uint64:
			args = append(args, uintptr(val))
		case bool:
			if val {
				args = append(args, 1)
			} else {
				args = append(args, 0)
			}
		case string:
			p := C.CString(val)
			frees = append(frees, unsafe.Pointer(p))
			args = append(args, uintptr(unsafe.Pointer(p)))
		case []byte:
			// 原始字节不要用 CString，避免被 0 截断
			if len(val) == 0 {
				args = append(args, 0)
				continue
			}
			p := C.CBytes(val)
			frees = append(frees, p)
			args = append(args, uintptr(p))
		default:
			panic("参数类型错误")
		}
	}

	ch <- true
	defer func() {
		<-ch
	}()

	ret, _, _ := syscall.SyscallN(uintptr(address), args...)
	return int(ret)
}
