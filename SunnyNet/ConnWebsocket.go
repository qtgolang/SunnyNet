package SunnyNet

import (
	"github.com/qtgolang/SunnyNet/src/CrossCompiled"
	"github.com/qtgolang/SunnyNet/src/Interface"
	"github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/public"
)

type ConnWebSocket Interface.ConnWebSocketCall
type wsConn struct {
	c             *public.WebsocketMsg
	SunnyContext  int
	_MessageId    int           //仅标识消息ID,不能用于API函数
	Pid           int           //Pid
	_Type         int           //消息类型 	public.Websocket...
	Url           string        //连接请求地址
	_Method       string        //连接时的Method
	_Theology     int           //请求唯一ID
	_ClientIP     string        //来源IP地址,请求从哪里来
	Request       *http.Request //请求体
	_Display      bool
	_localAddress string
}

func (k *wsConn) LocalAddress() string {
	return k._localAddress
}

func (k *wsConn) SetDisplay(Display bool) {
	k._Display = Display
}

func (k *wsConn) Method() string {
	return k._Method
}

func (k *wsConn) GetSocket5User() string {
	return GetSocket5User(k._Theology)
}

func (k *wsConn) GetProcessName() string {
	if k.Pid == 0 {
		return "代理连接"
	}
	return CrossCompiled.GetPidName(int32(k.Pid))
}
func (k *wsConn) Context() int {
	return k.SunnyContext
}
func (k *wsConn) MessageId() int {
	return k._MessageId
}
func (k *wsConn) Theology() int {
	return k._Theology
}
func (k *wsConn) PID() int {
	return k.Pid
}
func (k *wsConn) URL() string {
	return k.Url
}

func (k *wsConn) Type() int {
	return k._Type
}

func (k *wsConn) ClientIP() string {
	return k._ClientIP
}
func (k *wsConn) Body() []byte {
	k.c.Sync.Lock()
	defer k.c.Sync.Unlock()
	return public.CopyBytes(k.c.Data.Bytes())
}

// MessageType 获取 消息类型
// Text=1 Binary=2 Close=8 Ping=9 Pong=10 Invalid=-1/255
func (k *wsConn) MessageType() int {
	k.c.Sync.Lock()
	defer k.c.Sync.Unlock()
	return k.c.Mt
}

// BodyLen 获取 消息长度
func (k *wsConn) BodyLen() int {
	k.c.Sync.Lock()
	defer k.c.Sync.Unlock()
	return k.c.Data.Len()
}

// SetBody 修改 消息
func (k *wsConn) SetBody(data []byte) bool {
	k.c.Sync.Lock()
	defer k.c.Sync.Unlock()
	k.c.Data.Reset()
	k.c.Data.Write(data)
	return true
}

// SendToServer 主动向Websocket服务器发送消息
func (k *wsConn) SendToServer(MessageType int, data []byte) bool {
	k.c.Sync.Lock()
	defer k.c.Sync.Unlock()
	if k.c.Server != nil {
		e := k.c.Server.WriteMessage(MessageType, data)
		if e != nil {
			return false
		}
	}
	return true
}

// SendToClient 主动向Websocket客户端发送消息
func (k *wsConn) SendToClient(MessageType int, data []byte) bool {
	k.c.Sync.Lock()
	defer k.c.Sync.Unlock()
	if k.c.Client != nil {
		e := k.c.Client.WriteMessage(MessageType, data)
		if e != nil {
			return false
		}
	}
	return true
}

// Close 关闭Websocket连接
func (k *wsConn) Close() bool {
	k.c.Sync.Lock()
	defer k.c.Sync.Unlock()
	if k.c.Server != nil {
		_ = k.c.Server.Close()
	}
	if k.c.Client != nil {
		_ = k.c.Client.Close()
	}
	return true
}
