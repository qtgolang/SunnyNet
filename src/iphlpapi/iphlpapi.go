//go:build windows
// +build windows

package iphlpapi

/*
#include "c_iphlpapi_tcp.h"
*/
import "C"

func init() {
	C.closeTcpConnectionInit()
}

// CloseCurrentSocket  关闭指定进程的所有TCP连接
func CloseCurrentSocket(PID int, ulAf uint) {
	C.closeTcpConnectionByPid(C.ulong(PID), C.ulong(ulAf))
}
