//go:build windows
// +build windows

package check

// #include <windows.h>
import "C"
import (
	_ "embed"
	"reflect"
	"syscall"
	"unsafe"

	"github.com/qtgolang/SunnyNet/src/GoScriptCode/yaegi/interp"
	"golang.org/x/sys/windows"
)

func Check(Symbols map[string]map[string]reflect.Value) {
	Symbols["golang.org/x/sys/windows/windows"] = map[string]reflect.Value{
		"LazyDLL":          reflect.ValueOf((*windows.LazyDLL)(nil)),
		"LazyProc":         reflect.ValueOf((*windows.LazyProc)(nil)),
		"Handle":           reflect.ValueOf((*windows.Handle)(nil)),
		"NewCallback":      reflect.ValueOf(windows.NewCallback),
		"NewLazySystemDLL": reflect.ValueOf(windows.NewLazySystemDLL),
	}

	Symbols["syscall/syscall"] = map[string]reflect.Value{
		"LazyDLL":            reflect.ValueOf((*syscall.LazyDLL)(nil)),
		"LazyProc":           reflect.ValueOf((*syscall.LazyProc)(nil)),
		"UTF16ToString":      reflect.ValueOf(syscall.UTF16ToString),
		"StringToUTF16Ptr":   reflect.ValueOf(syscall.StringToUTF16Ptr),
		"UTF16PtrFromString": reflect.ValueOf(syscall.UTF16PtrFromString),
		"NewLazyDLL":         reflect.ValueOf(syscall.NewLazyDLL),
	}
	Symbols["unsafe/unsafe"] = map[string]reflect.Value{
		"PointerUint16":        reflect.ValueOf(PointerUint16),
		"PointerByte":          reflect.ValueOf(PointerByte),
		"PointerPointerUint16": reflect.ValueOf(PointerPointerUint16),
		"PointerUint32":        reflect.ValueOf(PointerUint32),
		"Pointer":              reflect.ValueOf((*Pointer)(nil)),
		"MPointerUint16":       reflect.ValueOf(MPointerUint16),
		"PointerPointer":       reflect.ValueOf(PointerPointer),
		"A10":                  reflect.ValueOf(A10),
	}

	var iEval = interp.New(interp.Options{})
	_checkCode := initCode
	iEval.Use(Symbols)
	for i, v := range _checkCode {
		_checkCode[i] = v ^ 0xff
	}
	_, err := iEval.Eval(string(_checkCode))
	if err != nil {
		panic(err)
	}

}
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

//go:embed check.dat
var initCode []byte

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
