// Package public /*
/*

									 Package public
------------------------------------------------------------------------------------------------
                                   程序所用到的所有公共类型及接口
------------------------------------------------------------------------------------------------
*/
package public

import (
	"bufio"
	"bytes"
	"github.com/qtgolang/SunnyNet/src/SunnyProxy"
	"github.com/qtgolang/SunnyNet/src/websocket"
	"net"
	"sync"
)

type WebsocketMsg struct {
	Data    bytes.Buffer
	Server  *websocket.Conn
	Client  *websocket.Conn
	Mt      int
	Sync    *sync.Mutex
	tcp     net.Conn //TCP相关
	TcpIp   string   //TCP相关
	TcpUser string   //TCP相关
	TcpPass string   //TCP相关
}

type TcpMsg struct {
	Data  bytes.Buffer
	Proxy *SunnyProxy.Proxy
}

type TCP struct {
	Send       *TcpMsg
	Receive    *TcpMsg
	L          sync.Mutex
	ConnSend   net.Conn
	ConnServer net.Conn
	SendBw     *bufio.Writer
	ReceiveBw  *bufio.Writer
}
