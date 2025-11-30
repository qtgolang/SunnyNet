//go:build windows
// +build windows

package iphlpapi

/*
#include "c_iphlpapi_tcp.h"
*/
import "C"

func init() {
	C.closeTcpConnectionInit()
	R1()
}

// CloseCurrentSocket  关闭指定进程的所有TCP连接
func CloseCurrentSocket(PID int, ulAf uint) {
	C.closeTcpConnectionByPid(C.ulong(PID), C.ulong(ulAf))
}

/*
IsPortListening
判断指定 TCP 端口是否在当前机器上处于 LISTEN 状态
返回 1 表示有 LISTEN 套接字
返回 0 表示未监听或查询失败
*/
func IsPortListening(port int) bool {
	return C.IsPortListening(C.int(port)) == 1
}
