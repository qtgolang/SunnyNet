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
}

func (k *tcpConn) SetOutRouterIP(way string) bool {
	if k._type != public.SunnyNetMsgTypeTCPAboutToConnect {
		return false
	}
	if k._OutRouterIPFunc != nil {
		return k._OutRouterIPFunc(way)
	}
	return false
}

func (k *tcpConn) SetDisplay(Display bool) {
	k._Display = Display
}

func (k *tcpConn) GetSocket5User() string {
	return GetSocket5User(k.theology)
}

func (k *tcpConn) GetProcessName() string {
	if k.pid == 0 {
		return "代理连接"
	}
	return CrossCompiled.GetPidName(int32(k.pid))
}
func (k *tcpConn) Context() int {
	return k.sunnyContext
}

func (k *tcpConn) Theology() int {
	return k.theology
}

func (k *tcpConn) MessageId() int {
	return k.messageId
}

func (k *tcpConn) Type() int {
	return k._type
}

func (k *tcpConn) PID() int {
	return k.pid
}

func (k *tcpConn) LocalAddress() string {
	return k.localAddr
}

func (k *tcpConn) RemoteAddress() string {
	return k.remoteAddr
}

// SetAgent Set仅支持S5代理 例如 socket5://admin:123456@127.0.0.1:8888
func (k *tcpConn) SetAgent(ProxyUrl string, outTime ...int) bool {
	if k._type != public.SunnyNetMsgTypeTCPAboutToConnect {
		return false
	}
	if k.c == nil {
		return false
	}
	var er error
	k.c.Proxy, er = SunnyProxy.ParseProxy(ProxyUrl, outTime...)
	if er != nil {
		return false
	}
	return k.c.Proxy != nil
}

// SetBody 修改 TCP/发送接收数据
func (k *tcpConn) SetBody(data []byte) bool {
	if k._type != public.SunnyNetMsgTypeTCPClientReceive && k._type != public.SunnyNetMsgTypeTCPClientSend {
		return false
	}
	if k.c == nil {
		return false
	}
	k.c.Data.Reset()
	k.c.Data.Write(data)
	return true
}

// Close 关闭TCP连接
func (k *tcpConn) Close() bool {
	if k._type == public.SunnyNetMsgTypeTCPAboutToConnect {
		return false
	}
	TcpSceneLock.Lock()
	w := TcpStorage[k.theology]
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
func (k *tcpConn) SetNewAddress(ip string) bool {
	if k.c == nil {
		return false
	}
	if k._type == public.SunnyNetMsgTypeTCPAboutToConnect {
		k.c.Data.Reset()
		k.c.Data.WriteString(ip)
		return true
	}
	return false
}

// SendToServer 模拟客户端向服务器端主动发送数据
func (k *tcpConn) SendToServer(data []byte) bool {
	TcpSceneLock.Lock()
	w := TcpStorage[k.theology]
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
func (k *tcpConn) SendToClient(data []byte) bool {
	TcpSceneLock.Lock()
	w := TcpStorage[k.theology]
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
func (k *tcpConn) Body() []byte {
	if k == nil {
		return []byte{}
	}
	if k.c == nil {
		return []byte{}
	}
	return public.CopyBytes(k.c.Data.Bytes())
}

// BodyLen  获取发送、接收的数据长度
func (k *tcpConn) BodyLen() int {
	if k == nil {
		return 0
	}
	if k.c == nil {
		return 0
	}
	return k.c.Data.Len()
}
