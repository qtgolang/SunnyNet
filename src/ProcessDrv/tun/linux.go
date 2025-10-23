//go:build linux && !android
// +build linux,!android

package tun

import (
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/tun/Tun"
)

var dev = Tun.NewTun{}

func IsRun() bool {
	return dev.IsRunning
}

func Install() bool {
	return true
}

func SetHandle(Handle Tun.TcpFunc, udpSendReceiveFunc Tun.UdpFunc, sunny Tun.Interface) bool {
	dev.ProxyPort = uint16(sunny.Port())
	dev.Sunny = sunny
	dev.SetHandle(Handle, udpSendReceiveFunc)
	return true
}
func Run() bool {
	return dev.OnTunCreated(0)
}
func Close() bool {
	dev.IsRunning = false
	return true
}
func Name() string {
	return "tun"
}

func UnInstall() bool {
	return true
}
func SetFd(fd int) bool {
	return true
}
