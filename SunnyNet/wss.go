package SunnyNet

import (
	"strings"
	"sync"
	"time"

	"github.com/qtgolang/SunnyNet/src/public"
	"github.com/qtgolang/SunnyNet/src/websocket"
)

func (s *proxyRequest) handleWss() bool {
	if s.Request == nil || s.Request.Header == nil {
		return true
	}
	if s.Request.ProtoMajor != 1 {
		return false
	}
	//判断是否是websocket的请求体 如果不是直接返回继续正常处理请求

	ok := strings.ToLower(s.Request.Header.Get("Upgrade")) == "websocket"
	if !ok {
		m := s.Request.Header["upgrade"]
		if len(m) > 0 {
			ok = strings.ToLower(m[0]) == "websocket"
		}
	}
	if ok {
		Method := "wss"
		Url := s.Request.URL.String()
		if strings.HasPrefix(Url, "net://") || strings.HasPrefix(Url, "http://") {
			Method = "ws"
		}
		var dialer *websocket.Dialer
		if s.Request.URL.Scheme == "https" {
			s.TlsConfig.NextProtos = []string{"http/1.1"}
			dialer = &websocket.Dialer{TLSClientConfig: s.TlsConfig}
		} else {
			dialer = &websocket.Dialer{}
		}
		//发送请求
		Server, r, er := dialer.ConnDialContext(s.Request, s.Proxy, s.outRouterIP)
		ip, _ := s.Request.Context().Value(public.SunnyNetServerIpTags).(string)
		if ip != "" {
			s.Response.ServerIP = ip
		} else {
			s.Response.ServerIP = "unknown"
		}
		s.Response.Response = r
		defer func() {
			if Server != nil {
				_ = Server.Close()
			}
		}()
		if er != nil {
			//如果发送错误
			s.Error(er, true)
			return true
		}
		s.Response.ServerIP = Server.RemoteAddr().String()
		_ = s.Conn.SetDeadline(time.Time{})
		//通知http请求完成回调
		s.CallbackBeforeResponse()
		//将当前客户端的连接升级为Websocket会话
		upgrade := &websocket.Upgrader{}
		Client, er := upgrade.UpgradeClient(s.Request, r, s.RwObj)
		if er != nil {
			return true
		}
		defer func() {
			if Client != nil {
				_ = Client.Close()
			}
		}()
		var sc sync.Mutex
		var wg sync.WaitGroup
		wg.Add(1)
		//开始转发消息
		receive := func() {
			as := &public.WebsocketMsg{Mt: 255, Server: Server, Client: Client, Sync: &sc}
			MessageId := 0
			Server.SetCloseHandler(func(code int, text string) error {
				message := websocket.FormatCloseMessage(code, text)
				as1 := &public.WebsocketMsg{Mt: websocket.CloseMessage, Server: Server, Client: Client, Sync: &sc}
				as1.Data.Write(message)
				//构造一个新的MessageId
				MessageId1 := NewMessageId()
				//储存对象
				messageIdLock.Lock()
				wsStorage[MessageId1] = as1
				httpStorage[MessageId1] = s
				messageIdLock.Unlock()
				defer func() {
					as1.Data.Reset()
					messageIdLock.Lock()
					wsStorage[MessageId1] = nil
					delete(wsStorage, MessageId1)
					httpStorage[MessageId1] = nil
					delete(httpStorage, MessageId1)
					messageIdLock.Unlock()
				}()
				s.CallbackWssRequest(public.WebsocketServerSend, Method, Url, as1, MessageId1)
				_ = Client.WriteControl(websocket.CloseMessage, as1.Data.Bytes(), time.Now().Add(time.Second*30))
				return nil
			})
			Server.SetPingHandler(func(appData []byte) error {
				as1 := &public.WebsocketMsg{Mt: websocket.PingMessage, Server: Server, Client: Client, Sync: &sc}
				as1.Data.Write(appData)
				//构造一个新的MessageId
				MessageId1 := NewMessageId()
				//储存对象
				messageIdLock.Lock()
				wsStorage[MessageId1] = as1
				httpStorage[MessageId1] = s
				messageIdLock.Unlock()
				defer func() {
					as1.Data.Reset()
					messageIdLock.Lock()
					wsStorage[MessageId1] = nil
					delete(wsStorage, MessageId1)
					httpStorage[MessageId1] = nil
					delete(httpStorage, MessageId1)
					messageIdLock.Unlock()
				}()
				s.CallbackWssRequest(public.WebsocketServerSend, Method, Url, as1, MessageId1)
				_ = Client.WriteMessage(websocket.PingMessage, as1.Data.Bytes())
				return nil
			})
			Server.SetPongHandler(func(appData []byte) error {
				as1 := &public.WebsocketMsg{Mt: websocket.PongMessage, Server: Server, Client: Client, Sync: &sc}
				as1.Data.Write(appData)
				//构造一个新的MessageId
				MessageId1 := NewMessageId()
				//储存对象
				messageIdLock.Lock()
				wsStorage[MessageId1] = as1
				httpStorage[MessageId] = s
				messageIdLock.Unlock()
				defer func() {
					as1.Data.Reset()
					messageIdLock.Lock()
					wsStorage[MessageId1] = nil
					delete(wsStorage, MessageId1)
					httpStorage[MessageId1] = nil
					delete(httpStorage, MessageId1)
					messageIdLock.Unlock()
				}()
				s.CallbackWssRequest(public.WebsocketServerSend, Method, Url, as1, MessageId1)
				_ = Client.WriteMessage(websocket.PongMessage, as1.Data.Bytes())
				return nil
			})
			for {
				{
					//清除上次的 MessageId
					messageIdLock.Lock()
					wsStorage[MessageId] = nil
					delete(wsStorage, MessageId)
					httpStorage[MessageId] = nil
					delete(httpStorage, MessageId)
					messageIdLock.Unlock()

					//构造一个新的MessageId
					MessageId = NewMessageId()

					//储存对象
					messageIdLock.Lock()
					httpStorage[MessageId] = s
					wsStorage[MessageId] = as
					messageIdLock.Unlock()
				}
				as.Data.Reset()
				mt, message, err := Server.ReadMessage()
				if message == nil && err == nil {
					as.Data.Reset()
					continue
				}
				if err != nil {
					as.Data.Reset()
					break
				}
				as.Data.Write(message)
				as.Mt = mt
				s.CallbackWssRequest(public.WebsocketServerSend, Method, Url, as, MessageId)
				sc.Lock()
				//发到客户端
				err = Client.WriteMessage(as.Mt, as.Data.Bytes())
				sc.Unlock()
				if err != nil {
					as.Data.Reset()
					break
				}
			}
			messageIdLock.Lock()
			wsStorage[MessageId] = nil
			delete(wsStorage, MessageId)
			httpStorage[MessageId] = nil
			delete(httpStorage, MessageId)
			messageIdLock.Unlock()
			_ = Client.Close()
			_ = Server.Close()
			wg.Done()
		}
		as := &public.WebsocketMsg{Mt: 255, Server: Server, Client: Client, Sync: &sc}
		MessageId := NewMessageId()
		messageIdLock.Lock()
		wsStorage[MessageId] = as
		httpStorage[MessageId] = s
		wsClientStorage[s.Theology] = as
		messageIdLock.Unlock()
		s.CallbackWssRequest(public.WebsocketConnectionOK, Method, Url, as, MessageId)
		go receive()

		// Client > Server
		Client.SetCloseHandler(func(code int, text string) error {
			message := websocket.FormatCloseMessage(code, text)
			as1 := &public.WebsocketMsg{Mt: websocket.CloseMessage, Server: Server, Client: Client, Sync: &sc}
			as1.Data.Write(message)
			//构造一个新的MessageId
			MessageId1 := NewMessageId()
			//储存对象
			messageIdLock.Lock()
			wsStorage[MessageId1] = as1
			httpStorage[MessageId1] = s
			messageIdLock.Unlock()
			defer func() {
				as1.Data.Reset()
				messageIdLock.Lock()
				wsStorage[MessageId1] = nil
				delete(wsStorage, MessageId1)
				httpStorage[MessageId1] = nil
				delete(httpStorage, MessageId1)
				messageIdLock.Unlock()
			}()
			s.CallbackWssRequest(public.WebsocketUserSend, Method, Url, as1, MessageId1)
			_ = Server.WriteControl(websocket.CloseMessage, as1.Data.Bytes(), time.Now().Add(time.Second*30))
			return nil
		})
		Client.SetPingHandler(func(appData []byte) error {
			as1 := &public.WebsocketMsg{Mt: websocket.PingMessage, Server: Server, Client: Client, Sync: &sc}
			as1.Data.Write(appData)
			//构造一个新的MessageId
			MessageId1 := NewMessageId()
			//储存对象
			messageIdLock.Lock()
			wsStorage[MessageId1] = as1
			httpStorage[MessageId1] = s
			messageIdLock.Unlock()
			defer func() {
				as1.Data.Reset()
				messageIdLock.Lock()
				wsStorage[MessageId1] = nil
				delete(wsStorage, MessageId1)
				httpStorage[MessageId1] = nil
				delete(httpStorage, MessageId1)
				messageIdLock.Unlock()
			}()
			s.CallbackWssRequest(public.WebsocketUserSend, Method, Url, as1, MessageId1)
			_ = Server.WriteMessage(websocket.PingMessage, as1.Data.Bytes())
			return nil
		})
		Client.SetPongHandler(func(appData []byte) error {
			as1 := &public.WebsocketMsg{Mt: websocket.PongMessage, Server: Server, Client: Client, Sync: &sc}
			as1.Data.Write(appData)
			//构造一个新的MessageId
			MessageId1 := NewMessageId()
			//储存对象
			messageIdLock.Lock()
			wsStorage[MessageId1] = as1
			httpStorage[MessageId1] = s
			messageIdLock.Unlock()
			defer func() {
				as1.Data.Reset()
				messageIdLock.Lock()
				wsStorage[MessageId1] = nil
				delete(wsStorage, MessageId1)
				httpStorage[MessageId1] = nil
				delete(httpStorage, MessageId1)
				messageIdLock.Unlock()
			}()
			s.CallbackWssRequest(public.WebsocketUserSend, Method, Url, as1, MessageId1)
			_ = Server.WriteMessage(websocket.PongMessage, as1.Data.Bytes())
			return nil
		})

		for {
			{
				//清除上次的 MessageId
				messageIdLock.Lock()
				wsStorage[MessageId] = nil
				delete(wsStorage, MessageId)
				httpStorage[MessageId] = nil
				delete(httpStorage, MessageId)
				messageIdLock.Unlock()

				//构造一个新的MessageId
				MessageId = NewMessageId()

				//储存对象
				messageIdLock.Lock()
				wsStorage[MessageId] = as
				httpStorage[MessageId] = s
				messageIdLock.Unlock()
			}
			as.Data.Reset()
			mt, message1, err := Client.ReadMessage()
			if message1 == nil && err == nil {
				as.Data.Reset()
				continue
			}
			as.Data.Write(message1)
			as.Mt = mt
			if err != nil {
				_ = Client.Close()
				_ = Server.Close()
				as.Data.Reset()
				s.CallbackWssRequest(public.WebsocketDisconnect, Method, Url, as, MessageId)
				break
			}
			s.CallbackWssRequest(public.WebsocketUserSend, Method, Url, as, MessageId)
			sc.Lock()
			err = Server.WriteMessage(as.Mt, as.Data.Bytes())
			sc.Unlock()
			if err != nil {
				_ = Client.Close()
				_ = Server.Close()
				as.Data.Reset()
				s.CallbackWssRequest(public.WebsocketDisconnect, Method, Url, as, MessageId)
				break
			}
		}
		wg.Wait()
		messageIdLock.Lock()

		wsStorage[MessageId] = nil
		delete(wsStorage, MessageId)

		httpStorage[MessageId] = nil
		delete(httpStorage, MessageId)

		wsClientStorage[s.Theology] = nil
		delete(wsClientStorage, s.Theology)

		messageIdLock.Unlock()
		return true
	}
	return false
}
