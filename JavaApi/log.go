//go:build android
// +build android

package JavaJni

/*
#include <android/log.h>
#include <stdlib.h>
static void LogError_C(const char* tag, const char* msg) {
    __android_log_print(ANDROID_LOG_ERROR, tag, "%s", msg);
}
*/
import "C"
import (
	"unsafe"
)

func LogError(msg string) {
	tag := "SunnyNet"
	cTag := C.CString(tag)
	cMsg := C.CString(msg)
	C.LogError_C(cTag, cMsg)
	C.free(unsafe.Pointer(cTag))
	C.free(unsafe.Pointer(cMsg))
}
