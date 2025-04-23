package SunnyNet

import "C"
import (
	"github.com/qtgolang/SunnyNet/src/Call"
	"github.com/qtgolang/SunnyNet/src/dns"
	"github.com/qtgolang/SunnyNet/src/public"
	"strconv"
	"strings"
)

const debug = "Debug"

func GetSceneProxyRequest(MessageId int) (*proxyRequest, bool) {
	messageIdLock.Lock()
	defer messageIdLock.Unlock()
	k := httpStorage[MessageId]
	if k == nil {
		return nil, false
	}
	return k, true
}
func GetSceneWebSocketMsg(MessageId int) (*public.WebsocketMsg, bool) {
	messageIdLock.Lock()
	defer messageIdLock.Unlock()
	k := wsStorage[MessageId]
	if k == nil {
		return nil, false
	}
	return k, true
}
func GetSceneWebSocketClient(Theology int) (*public.WebsocketMsg, bool) {
	messageIdLock.Lock()
	defer messageIdLock.Unlock()
	k := wsClientStorage[Theology]
	if k == nil {
		return nil, false
	}
	return k, true
}

// CallbackTCPRequest TCP请求处理回调
func (s *proxyRequest) CallbackTCPRequest(callType int, _msg *public.TcpMsg, RemoteAddr string) {
	if RemoteAddr == dns.GetDnsServer() {
		return
	}
	if s.noCallback(RemoteAddr) {
		return
	}
	if s.Global.disableTCP {
		//由于用户可能在软件中途禁用TCP,所有这里允许触发关闭的回调
		if callType != public.SunnyNetMsgTypeTCPClose {
			//这里如果禁用了TCP,那么这里就不允许触发回调了，并且手动关闭连接
			TcpSceneLock.Lock()
			w := TcpStorage[s.Theology]
			TcpSceneLock.Unlock()
			if w == nil {
				return
			}
			w.L.Lock()
			_ = w.ConnSend.Close()
			_ = w.ConnServer.Close()
			w.L.Unlock()
			return
		}
	}
	LocalAddr := s.Conn.RemoteAddr().String()
	hostname := RemoteAddr
	pid, _ := strconv.Atoi(s.Pid)
	MessageId := NewMessageId()

	messageIdLock.Lock()
	httpStorage[MessageId] = s
	messageIdLock.Unlock()

	defer func() {
		messageIdLock.Lock()
		httpStorage[MessageId] = nil
		delete(httpStorage, MessageId)
		messageIdLock.Unlock()
	}()
	Ams := &tcpConn{
		c:                _msg,
		messageId:        MessageId,
		_type:            callType,
		theology:         s.Theology,
		localAddr:        LocalAddr,
		remoteAddr:       hostname,
		pid:              pid,
		sunnyContext:     s.Global.SunnyContext,
		_Display:         true,
		_OutRouterIPFunc: s.SetOutRouterIP,
	}
	s.Global.scriptTCPCall(Ams)
	if !Ams._Display {
		return
	}
	msg := Ams.c
	if callType == public.SunnyNetMsgTypeTCPAboutToConnect {
		if msg.Proxy != nil {
			_msg.Proxy = msg.Proxy
		}
	}
	if s.TcpCall < 10 {
		if s.TcpGoCall != nil {
			s.TcpGoCall(Ams)
			if callType == public.SunnyNetMsgTypeTCPAboutToConnect {
				if msg.Proxy != nil {
					_msg.Proxy = msg.Proxy
				}
			}
		}
		return
	}
	if callType == public.SunnyNetMsgTypeTCPConnectOK {
		Call.Call(s.TcpCall, s.Global.SunnyContext, LocalAddr, hostname, int(callType), MessageId, msg.Data.Bytes(), msg.Data.Len(), s.Theology, pid)
		return
	}
	if callType == public.SunnyNetMsgTypeTCPClose {
		Call.Call(s.TcpCall, s.Global.SunnyContext, LocalAddr, hostname, int(callType), MessageId, []byte{}, 0, s.Theology, pid)
		return
	}
	if callType == public.SunnyNetMsgTypeTCPClientSend || callType == public.SunnyNetMsgTypeTCPAboutToConnect {
		s.TCP.Send = msg
	} else {
		s.TCP.Receive = msg
	}
	Call.Call(s.TcpCall, s.Global.SunnyContext, LocalAddr, hostname, int(callType), MessageId, msg.Data.Bytes(), msg.Data.Len(), s.Theology, pid)
}

// CallbackBeforeRequest HTTP发起请求处理回调
func (s *proxyRequest) CallbackBeforeRequest() {
	if s.noCallback() {
		return
	}
	if s.Response.Response != nil {
		if s.Response.Body != nil {
			_ = s.Response.Body.Close()
		}
	}
	pid, _ := strconv.Atoi(s.Pid)
	s.Response.Response = nil
	defer func() {
		if s.Response.Response != nil {
			if s.Response.Response.StatusCode == 0 && len(s.Response.Header) == 0 {
				if s.Response.ContentLength < 1 {
					if s.Response.Body != nil {
						_ = s.Response.Body.Close()
					}
					s.Response.Response = nil
				}
			}
		}
	}()
	MessageId := NewMessageId()
	messageIdLock.Lock()
	httpStorage[MessageId] = s
	messageIdLock.Unlock()
	defer func() {
		messageIdLock.Lock()
		httpStorage[MessageId] = nil
		delete(httpStorage, MessageId)
		messageIdLock.Unlock()
	}()

	m := &httpConn{
		_Theology:        s.Theology,
		_getRawBody:      s.RawRequestDataToFile,
		_MessageId:       MessageId,
		_PID:             pid,
		_Context:         s.Global.SunnyContext,
		_Type:            public.HttpSendRequest,
		_request:         s.Request,
		_response:        s.Response.Response,
		_err:             "",
		_proxy:           s.Proxy,
		_ClientIP:        s.Conn.RemoteAddr().String(),
		_Display:         true,
		_Break:           false,
		_tls:             s.TlsConfig,
		_serverIP:        s.Response.ServerIP,
		_localAddress:    s.Conn.LocalAddr().String(),
		_OutRouterIPFunc: s.SetOutRouterIP,
	}
	s.Global.scriptHTTPCall(m)
	s.TlsConfig = m._tls
	s.Response.Response = m._response
	s._Display = m._Display
	if m._proxy != nil {
		s.Proxy = m._proxy
	}
	s._isRandomCipherSuites = m._isRandomCipherSuites
	if s._Display == false {
		return
	}
	err := ""
	if m._Break {
		err = debug
	}

	if s.HttpCall < 10 {
		if s.HttpGoCall != nil {
			s.HttpGoCall(m)
			s.Response.Response = m._response
			if m._proxy != nil {
				s.Proxy = m._proxy
			}
		}
		return
	}
	Method := s.Request.Method
	Url := s.Request.URL.String()
	Call.Call(s.HttpCall, s.Global.SunnyContext, s.Theology, MessageId, int(public.HttpSendRequest), Method, Url, err, pid)
}

// CallbackBeforeResponse HTTP请求完成处理回调
func (s *proxyRequest) CallbackBeforeResponse() {
	if s.noCallback() {
		return
	}
	pid, _ := strconv.Atoi(s.Pid)

	MessageId := NewMessageId()

	messageIdLock.Lock()
	httpStorage[MessageId] = s
	messageIdLock.Unlock()
	defer func() {
		messageIdLock.Lock()
		httpStorage[MessageId] = nil
		delete(httpStorage, MessageId)
		messageIdLock.Unlock()
	}()

	m := &httpConn{
		_Theology:        s.Theology,
		_getRawBody:      s.RawRequestDataToFile,
		_MessageId:       MessageId,
		_PID:             pid,
		_Context:         s.Global.SunnyContext,
		_Type:            public.HttpResponseOK,
		_request:         s.Request,
		_response:        s.Response.Response,
		_err:             "",
		_ClientIP:        s.Conn.RemoteAddr().String(),
		_Display:         true,
		_Break:           false,
		_tls:             s.TlsConfig,
		_serverIP:        s.Response.ServerIP,
		_localAddress:    s.Conn.LocalAddr().String(),
		_OutRouterIPFunc: s.SetOutRouterIP,
	}
	s.Global.scriptHTTPCall(m)
	s.Response.Response = m._response
	if s._Display == false {
		return
	}
	err := ""
	if m._Break {
		err = debug
	}
	if s.HttpCall < 10 {
		if s.HttpGoCall != nil {
			s.HttpGoCall(m)
			s.Response.Response = m._response
		}
		return
	}
	Method := s.Request.Method
	Url := s.Request.URL.String()
	Call.Call(s.HttpCall, s.Global.SunnyContext, s.Theology, MessageId, public.HttpResponseOK, Method, Url, err, pid)
}

// 不要进入回调的一些请求
func (s *proxyRequest) noCallback(n ...string) bool {
	if s == nil {
		return false
	}
	if len(n) > 0 {
		if n[0] == "127.0.0.1:9229" || n[0] == "[::1]:9229" {
			//疑似Chrome 开发人员工具正在使用专用的DevTools 即使所有选项卡都关闭，除了空白的新选项卡，它仍可能继续发送
			//https://superuser.com/questions/1419223/google-chrome-developer-tools-start-knocking-to-127-0-0-1-and-1-ip-on-9229-por
			return true
		}
	}
	request := s.Request
	Port := int(s.Target.Port)
	if (s.Target.Host == "localhost" || s.Target.Host == "127.0.0.1" || s.Target.Host == "::1") && Port == 9229 {
		//疑似Chrome 开发人员工具正在使用专用的DevTools 即使所有选项卡都关闭，除了空白的新选项卡，它仍可能继续发送
		//https://superuser.com/questions/1419223/google-chrome-developer-tools-start-knocking-to-127-0-0-1-and-1-ip-on-9229-por
		return true
	}

	//下面判断是否为证书安装页面 或 脚本编辑页面  如果是 则不触发回调页面
	if (s.Target.Host == "localhost" && Port == s.Global.port) || (s.Target.Host == "127.0.0.1" && Port == s.Global.port) || (s.Target.Host == "::1" && Port == s.Global.port) || (s.Target.Host == public.CertDownloadHost2) || (s.Target.Host == public.CertDownloadHost1) {
		if request != nil {
			if request.URL != nil {
				ScriptPage := "/" + s.Global.script.AdminPage
				if strings.HasPrefix(request.URL.Path, ScriptPage) {
					return true
				}
				if request.URL.Path == "/favicon.ico" {
					return true
				}
				if request.URL.Path == "/" || request.URL.Path == "/ssl" || request.URL.Path == public.NULL {
					return true
				}
				if strings.HasPrefix(request.URL.Path, "/SunnyRoot") {
					return true
				}
				if strings.HasPrefix(request.URL.Path, "/install.html") {
					return true
				}
				if strings.HasPrefix(request.URL.Path, "/install/") {
					return true
				}
			}
		}
	}
	return false
}

// CallbackError HTTP请求失败处理回调
func (s *proxyRequest) CallbackError(err string) {
	if s.noCallback() {
		return
	}
	pid, _ := strconv.Atoi(s.Pid)
	MessageId := NewMessageId()
	messageIdLock.Lock()
	httpStorage[MessageId] = s
	messageIdLock.Unlock()
	defer func() {
		messageIdLock.Lock()
		httpStorage[MessageId] = nil
		delete(httpStorage, MessageId)
		messageIdLock.Unlock()
	}()
	m := &httpConn{
		_Theology:        s.Theology,
		_getRawBody:      s.RawRequestDataToFile,
		_MessageId:       NewMessageId(),
		_PID:             pid,
		_Context:         s.Global.SunnyContext,
		_Type:            public.HttpRequestFail,
		_request:         s.Request,
		_response:        nil,
		_err:             err,
		_ClientIP:        s.Conn.RemoteAddr().String(),
		_tls:             nil,
		_serverIP:        s.Response.ServerIP,
		_localAddress:    s.Conn.LocalAddr().String(),
		_OutRouterIPFunc: s.SetOutRouterIP,
	}
	s.Global.scriptHTTPCall(m)
	if s._Display == false {
		return
	}
	if s.HttpCall < 10 {
		if s.HttpGoCall != nil {
			s.HttpGoCall(m)
		}
		return
	}
	//请求失败
	Method := s.Request.Method
	Url := "Unknown URL"
	if s.Request.URL != nil {
		Url = s.Request.URL.String()
	}
	Call.Call(s.HttpCall, s.Global.SunnyContext, s.Theology, MessageId, int(public.HttpRequestFail), Method, Url, err, pid)

}

// CallbackWssRequest HTTP->Websocket请求处理回调
func (s *proxyRequest) CallbackWssRequest(State int, Method, Url string, msg *public.WebsocketMsg, MessageId int) {
	if s._Display == false {
		return
	}
	pid, _ := strconv.Atoi(s.Pid)
	m := &wsConn{
		_Method:       Method,
		Pid:           pid,
		_Type:         State,
		SunnyContext:  s.Global.SunnyContext,
		Url:           Url,
		c:             msg,
		_MessageId:    MessageId,
		_Theology:     s.Theology,
		Request:       s.Request,
		_ClientIP:     s.Conn.RemoteAddr().String(),
		_localAddress: s.Conn.LocalAddr().String(),
		_Display:      true,
	}
	s.Global.scriptWebsocketCall(m)
	if !s._Display {
		return
	}
	if s.wsCall < 10 {
		if s.wsGoCall != nil {
			s.wsGoCall(m)
		}
		return
	}
	Call.Call(s.wsCall, s.Global.SunnyContext, s.Theology, MessageId, State, Method, Url, pid, msg.Mt)
}
