package Tun

import (
	n2 "net"
	"os"
)

type Interface interface {
	SetOutRouterIP(RouterIP string) bool
	Port() int
}
type UdpFunc func(Type int, Theoni int64, pid uint32, LocalAddress string, RemoteAddress string, data []byte) []byte
type TcpFunc func(conn n2.Conn)

var _myPid = int32(os.Getpid())
