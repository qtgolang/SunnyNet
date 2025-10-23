package SunnyNet

import (
	"github.com/qtgolang/SunnyNet/src/CrossCompiled"
	"github.com/qtgolang/SunnyNet/src/Interface"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/SunnyNetUDP"
)

type ConnUDP Interface.ConnUDPCall

type udpConn struct {
	sunnyContext  int
	theology      int64 //唯一ID
	messageId     int   //消息ID
	_type         int   //请求类型 例如 public.SunnyNetUDPType...
	pid           int
	localAddress  string
	remoteAddress string
	data          []byte
	_Display      bool
	_note         string
}

func (u *udpConn) SetNote(s string) {
	u._note = s
}

func (u *udpConn) GetNote() string {
	return u._note
}

func (u *udpConn) SetDisplay(Display bool) {
	u._Display = Display
}
func (u *udpConn) GetSocket5User() string {
	return ""
}

func (u *udpConn) GetProcessName() string {
	if u.pid == 0 {
		return "代理连接"
	}
	return CrossCompiled.GetPidName(int32(u.pid))
}

// SetBody 修改消息
func (u *udpConn) SetBody(i []byte) bool {
	u.data = i
	return true
}
func (u *udpConn) BodyLen() int {
	return len(u.data)
}

func (u *udpConn) Context() int {
	return u.sunnyContext
}
func (u *udpConn) Type() int {
	return u._type
}
func (u *udpConn) MessageId() int {
	return u.messageId
}
func (u *udpConn) Theology() int {
	return int(u.theology)
}

func (u *udpConn) PID() int {
	return u.pid
}

func (u *udpConn) LocalAddress() string {
	return u.localAddress
}

func (u *udpConn) RemoteAddress() string {
	return u.remoteAddress
}

func (u *udpConn) Body() []byte {
	return u.data
}

// SendToServer 主动向服务器发送消息
func (u *udpConn) SendToServer(data []byte) bool {
	obj := SunnyNetUDP.GetUDPItem(u.theology)
	if obj != nil {
		return obj.ToServer(data)
	}
	return false
}

// SendToClient 主动向客户端发送消息
func (u *udpConn) SendToClient(data []byte) bool {
	obj := SunnyNetUDP.GetUDPItem(u.theology)
	if obj != nil {
		return obj.ToClient(data)
	}
	return false
}
