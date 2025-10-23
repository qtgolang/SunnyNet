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
	_note         string
}

func (w *wsConn) SetNote(s string) {
	w._note = s
}

func (w *wsConn) GetNote() string {
	return w._note
}

func (w *wsConn) LocalAddress() string {
	return w._localAddress
}

func (w *wsConn) SetDisplay(Display bool) {
	w._Display = Display
}

func (w *wsConn) Method() string {
	return w._Method
}

func (w *wsConn) GetSocket5User() string {
	return GetSocket5User(w._Theology)
}

func (w *wsConn) GetProcessName() string {
	if w.Pid == 0 {
		return "代理连接"
	}
	return CrossCompiled.GetPidName(int32(w.Pid))
}
func (w *wsConn) Context() int {
	return w.SunnyContext
}
func (w *wsConn) MessageId() int {
	return w._MessageId
}
func (w *wsConn) Theology() int {
	return w._Theology
}
func (w *wsConn) PID() int {
	return w.Pid
}
func (w *wsConn) URL() string {
	return w.Url
}

func (w *wsConn) Type() int {
	return w._Type
}

func (w *wsConn) ClientIP() string {
	return w._ClientIP
}
func (w *wsConn) Body() []byte {
	w.c.Sync.Lock()
	defer w.c.Sync.Unlock()
	return public.CopyBytes(w.c.Data.Bytes())
}

// MessageType 获取 消息类型
// Text=1 Binary=2 Close=8 Ping=9 Pong=10 Invalid=-1/255
func (w *wsConn) MessageType() int {
	w.c.Sync.Lock()
	defer w.c.Sync.Unlock()
	return w.c.Mt
}

// BodyLen 获取 消息长度
func (w *wsConn) BodyLen() int {
	w.c.Sync.Lock()
	defer w.c.Sync.Unlock()
	return w.c.Data.Len()
}

// SetBody 修改 消息
func (w *wsConn) SetBody(data []byte) bool {
	w.c.Sync.Lock()
	defer w.c.Sync.Unlock()
	w.c.Data.Reset()
	w.c.Data.Write(data)
	return true
}

// SendToServer 主动向Websocket服务器发送消息
func (w *wsConn) SendToServer(MessageType int, data []byte) bool {
	w.c.Sync.Lock()
	defer w.c.Sync.Unlock()
	if w.c.Server != nil {
		e := w.c.Server.WriteMessage(MessageType, data)
		if e != nil {
			return false
		}
	}
	return true
}

// SendToClient 主动向Websocket客户端发送消息
func (w *wsConn) SendToClient(MessageType int, data []byte) bool {
	w.c.Sync.Lock()
	defer w.c.Sync.Unlock()
	if w.c.Client != nil {
		e := w.c.Client.WriteMessage(MessageType, data)
		if e != nil {
			return false
		}
	}
	return true
}

// Close 关闭Websocket连接
func (w *wsConn) Close() bool {
	w.c.Sync.Lock()
	defer w.c.Sync.Unlock()
	if w.c.Server != nil {
		_ = w.c.Server.Close()
	}
	if w.c.Client != nil {
		_ = w.c.Client.Close()
	}
	return true
}
