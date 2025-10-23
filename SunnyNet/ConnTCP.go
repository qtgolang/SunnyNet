package SunnyNet

import (
	"github.com/qtgolang/SunnyNet/src/CrossCompiled"
	"github.com/qtgolang/SunnyNet/src/Interface"
	"github.com/qtgolang/SunnyNet/src/SunnyProxy"
	"github.com/qtgolang/SunnyNet/src/public"
)

type ConnTCP Interface.ConnTCPCall

type tcpConn struct {
	sunnyContext     int
	theology         int //唯一ID
	messageId        int
	c                *public.TcpMsg //事件消息
	_type            int            //事件类型_ 例如  public.SunnyNetMsgTypeTCP.....
	localAddr        string         //本地地址
	remoteAddr       string         //远程地址
	pid              int            //Pid
	_Display         bool
	_OutRouterIPFunc func(string) bool
	_note            string
}

func (t *tcpConn) SetNote(s string) {
	t._note = s
}

func (t *tcpConn) GetNote() string {
	return t._note
}

func (t *tcpConn) SetOutRouterIP(way string) bool {
	if t._type != public.SunnyNetMsgTypeTCPAboutToConnect {
		return false
	}
	if t._OutRouterIPFunc != nil {
		return t._OutRouterIPFunc(way)
	}
	return false
}

func (t *tcpConn) SetDisplay(Display bool) {
	t._Display = Display
}

func (t *tcpConn) GetSocket5User() string {
	return GetSocket5User(t.theology)
}

func (t *tcpConn) GetProcessName() string {
	if t.pid == 0 {
		return "代理连接"
	}
	return CrossCompiled.GetPidName(int32(t.pid))
}
func (t *tcpConn) Context() int {
	return t.sunnyContext
}

func (t *tcpConn) Theology() int {
	return t.theology
}

func (t *tcpConn) MessageId() int {
	return t.messageId
}

func (t *tcpConn) Type() int {
	return t._type
}

func (t *tcpConn) PID() int {
	return t.pid
}

func (t *tcpConn) LocalAddress() string {
	return t.localAddr
}

func (t *tcpConn) RemoteAddress() string {
	return t.remoteAddr
}

// SetAgent Set仅支持S5代理 例如 socket5://admin:123456@127.0.0.1:8888
func (t *tcpConn) SetAgent(ProxyUrl string, outTime ...int) bool {
	if t._type != public.SunnyNetMsgTypeTCPAboutToConnect {
		return false
	}
	if t.c == nil {
		return false
	}
	var er error
	t.c.Proxy, er = SunnyProxy.ParseProxy(ProxyUrl, outTime...)
	if er != nil {
		return false
	}
	return t.c.Proxy != nil
}

// SetBody 修改 TCP/发送接收数据
func (t *tcpConn) SetBody(data []byte) bool {
	if t._type != public.SunnyNetMsgTypeTCPClientReceive && t._type != public.SunnyNetMsgTypeTCPClientSend {
		return false
	}
	if t.c == nil {
		return false
	}
	t.c.Data.Reset()
	t.c.Data.Write(data)
	return true
}

// Close 关闭TCP连接
func (t *tcpConn) Close() bool {
	if t._type == public.SunnyNetMsgTypeTCPAboutToConnect {
		return false
	}
	TcpSceneLock.Lock()
	w := TcpStorage[t.theology]
	TcpSceneLock.Unlock()
	if w == nil {
		return false
	}
	w.L.Lock()
	if w.ConnSend != nil {
		_ = w.ConnSend.Close()
	}
	if w.ConnServer != nil {
		_ = w.ConnServer.Close()
	}
	w.L.Unlock()
	return true
}

// SetNewAddress 修改目标连接地址 目标地址必须带端口号 例如 baidu.com:443 [仅限即将连接时使用]
func (t *tcpConn) SetNewAddress(ip string) bool {
	if t.c == nil {
		return false
	}
	if t._type == public.SunnyNetMsgTypeTCPAboutToConnect {
		t.c.Data.Reset()
		t.c.Data.WriteString(ip)
		return true
	}
	return false
}

// SendToServer 模拟客户端向服务器端主动发送数据
func (t *tcpConn) SendToServer(data []byte) bool {
	TcpSceneLock.Lock()
	w := TcpStorage[t.theology]
	TcpSceneLock.Unlock()
	if w == nil {
		return false
	}
	if w.Send == nil {
		return false
	}
	w.L.Lock()
	defer w.L.Unlock()
	if len(data) > 0 {
		x, e := w.ReceiveBw.Write(data)
		if e == nil {
			_ = w.ReceiveBw.Flush()
		}
		return x > 0
	}
	return false
}

// SendToClient  模拟服务器端向客户端主动发送数据
func (t *tcpConn) SendToClient(data []byte) bool {
	TcpSceneLock.Lock()
	w := TcpStorage[t.theology]
	TcpSceneLock.Unlock()
	if w == nil {
		return false
	}
	if w.Receive == nil {
		return false
	}
	if len(data) > 0 {
		w.L.Lock()
		defer w.L.Unlock()
		x, e := w.SendBw.Write(data)
		if e == nil {
			_ = w.SendBw.Flush()
		}
		return x > 0
	}
	return false
}

// Body  获取发送、接收的数据
func (t *tcpConn) Body() []byte {
	if t == nil {
		return []byte{}
	}
	if t.c == nil {
		return []byte{}
	}
	return public.CopyBytes(t.c.Data.Bytes())
}

// BodyLen  获取发送、接收的数据长度
func (t *tcpConn) BodyLen() int {
	if t == nil {
		return 0
	}
	if t.c == nil {
		return 0
	}
	return t.c.Data.Len()
}
