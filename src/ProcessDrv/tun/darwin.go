//go:build darwin
// +build darwin

package tun

import (
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/ProcessCheck"
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
	dev.CheckProcess = ProcessCheck.CheckPidByName
	dev.SetHandle(Handle, udpSendReceiveFunc)
	return true
}
func Run() bool {
	if dev.IsRunning {
		return true
	}
	return dev.OnTunCreated(0)
}
func Close() bool {
	dev.IsRunning = false
	return true
}
func Name() string {
	return "utun"
}

func UnInstall() bool {
	return true
}
func SetFd(fd int) bool {
	return true
}
