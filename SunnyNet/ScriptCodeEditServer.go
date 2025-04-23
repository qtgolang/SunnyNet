package SunnyNet

import (
	"encoding/json"
	"github.com/qtgolang/SunnyNet/src/GoScriptCode"
	"github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/public"
	"github.com/qtgolang/SunnyNet/src/websocket"
	"go/format"
	"strings"
	"time"
)

var _userScriptCodeEditUpgrade = &websocket.Upgrader{
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	}}

type sunnyWebSocketServer struct {
	Cmd  string `json:"cmd"`
	Data string `json:"data"`
}

func (s sunnyWebSocketServer) Error(conn *websocket.Conn, err string) {
	s.Cmd = "Error"
	s.Data = err
	_ = conn.WriteJSON(s)
}
func (s sunnyWebSocketServer) Message(conn *websocket.Conn, Message string) {
	s.Cmd = "Message"
	s.Data = Message
	_ = conn.WriteJSON(s)
}
func (s sunnyWebSocketServer) LoadDefaultCode(conn *websocket.Conn, code []byte, Global *Sunny, okMsg string) {
	code, _err := format.Source(code)
	if _err != nil {
		s.Error(conn, public.ProcessError(_err))
		return
	}
	s.Cmd = "SetCode"
	s.Data = string(code)
	_ = conn.WriteJSON(s)
	str := Global.SetScriptCode(string(code))
	if str == "" {
		Global.userScriptCode = code
		s.Message(conn, okMsg)
	} else {
		s.Error(conn, str)
	}
}
func (s *proxyRequest) scriptCodeEditServerHandleWebSocket(w http.ResponseWriter, r *http.Request) {
	for k, v := range r.Header {
		sss := strings.ReplaceAll(k, "-WebSocket-", "-Websocket-")
		r.Header[sss] = v
	}
	// 将  连接升级为 WebSocket
	conn, err := _userScriptCodeEditUpgrade.UpgradeSunnyNetWebsocket(s.RwObj, r, nil, s.Conn, s.RwObj.ReadWriter)
	if err != nil {
		_, _ = s.RwObj.Write(public.LocalBuildBody("text/html", err.Error()))
		return
	}
	defer conn.Close()
	_ = s.RwObj.SetDeadline(time.Time{})
	var ServerMsg sunnyWebSocketServer
	ServerMsg.Cmd = "SetCodeInit"
	ServerMsg.Data = string(s.Global.userScriptCode)
	_ = conn.WriteJSON(ServerMsg)
	for {
		_, message, er := conn.ReadMessage()
		if er != nil {
			break
		}
		_ = json.Unmarshal(message, &ServerMsg)
		{
			if ServerMsg.Cmd == "CodeLoadSave" {
				ServerMsg.LoadDefaultCode(conn, []byte(ServerMsg.Data), s.Global, "格式化代码,并加载代码 成功")
				continue
			}
			if ServerMsg.Cmd == "LoadDefaultCode" {
				switch ServerMsg.Data {
				case "DefaultCode":
					ServerMsg.LoadDefaultCode(conn, GoScriptCode.DefaultCode, s.Global, "恢复到默认代码 成功")
					break
				case "httpDefaultCode":
					ServerMsg.LoadDefaultCode(conn, GoScriptCode.DefaultHTTPCode, s.Global, "恢复到默认HTTP示例代码 成功")
					break
				case "tcpDefaultCode":
					ServerMsg.LoadDefaultCode(conn, GoScriptCode.DefaultTCPCode, s.Global, "恢复到默认 TCP 示例代码 成功")
					break
				case "udpDefaultCode":
					ServerMsg.LoadDefaultCode(conn, GoScriptCode.DefaultUDPCode, s.Global, "恢复到默认 UDP 示例代码 成功")
					break
				case "WebsocketDefaultCode":
					ServerMsg.LoadDefaultCode(conn, GoScriptCode.DefaultWSCode, s.Global, "恢复到默认Websocket示例代码 成功")
					break
				}
				continue
			}
		}
	}
}
