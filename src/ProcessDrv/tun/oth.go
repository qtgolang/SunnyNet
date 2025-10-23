//go:build !android && !darwin && !linux && !windows
// +build !android,!darwin,!linux,!windows

package tun

import (
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/tun/Tun"
)

func IsRun() bool {
	return false
}

func Install() bool {
	return false
}

func SetHandle(Handle Tun.TcpFunc, udpSendReceiveFunc Tun.UdpFunc, sunny Tun.Interface) bool {
	return true
}
func Run() bool {
	return false
}
func Close() bool {
	return true
}
func Name() string {
	return "null"
}

func UnInstall() bool {
	return true
}
func SetFd(fd int) bool {
	return false
}
