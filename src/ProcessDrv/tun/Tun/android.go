//go:build android
// +build android

package Tun

import (
	"os"

	"github.com/qtgolang/SunnyNet/JavaApi"
)

func LogError(msg string) {
	JavaJni.LogError(msg)
}

var defaultGatewayIP, defaultGatewayIf = "", ""

func (n *NewTun) OnTunCreated(fd int) {
	tun := os.NewFile(uintptr(fd), "tun0")
	defer func() {
		_ = tun.Close()
	}()
	n.tun = tun
	buf := make([]byte, 65535)
	for {
		nBytes, err := tun.Read(buf)
		if err != nil || !n.IsRunning {
			_ = tun.Close()
			break
		}
		var packet []byte
		packet = append(packet, buf[:nBytes]...)
		// 异步调用解析函数处理数据包
		go n.parsePacket(packet)
	}
}
func getPidByPort(kind string, port uint16) (int32, string) {
	return 0, ""
}
