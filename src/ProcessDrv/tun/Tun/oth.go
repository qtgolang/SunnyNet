//go:build !android && !darwin && !linux
// +build !android,!darwin,!linux

package Tun

import (
	"io"
)

var defaultGatewayIP, defaultGatewayIf = "", ""

type NewTun struct {
	IsRunning         bool
	ProxyPort         uint16
	tun               io.ReadWriteCloser
	handleTCPCallback TcpFunc
	handleUDPCallback UdpFunc
	Sunny             Interface
	CheckProcess      func(int32, string) bool
}
