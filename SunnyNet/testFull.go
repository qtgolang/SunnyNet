//go:build !mini
// +build !mini

package SunnyNet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/qtgolang/SunnyNet/src/GoScriptCode"
	"github.com/qtgolang/SunnyNet/src/Interface"
	"github.com/qtgolang/SunnyNet/src/Resource"
	"github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/public"
	"strings"
)

// SetScriptPage 设置脚本页面
func (s *Sunny) SetScriptPage(Page string) string {
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(Page) < 8 {
		return s.script.AdminPage
	}
	if strings.HasPrefix(Page, "/") {
		s.script.AdminPage = Page[1:]
	} else {
		s.script.AdminPage = Page
	}
	return s.script.AdminPage
}

// SetScriptCode 设置脚本代码
func (s *Sunny) SetScriptCode(code string) string {
	Code := []byte(code)
	if len(strings.TrimSpace(code)) < 1 {
		Code = GoScriptCode.DefaultCode
	}
	err, _ScriptFuncHTTP, _ScriptFuncWS, _ScriptFuncTCP, _ScriptFuncUDP := GoScriptCode.RunCode(s.SunnyContext, Code, s.script.LogCallback)
	if err == "" {
		s.lock.Lock()
		s.userScriptCode = Code
		s.script.http = _ScriptFuncHTTP
		s.script.websocket = _ScriptFuncWS
		s.script.tcp = _ScriptFuncTCP
		s.script.udp = _ScriptFuncUDP
		s.lock.Unlock()
		if s.script.SaveCallback != nil {
			s.script.SaveCallback(s.SunnyContext, Code)
		}
	}
	return err
}

// 是否是用户自定义脚本编辑请求
func (s *proxyRequest) isUserScriptCodeEditRequest(request *http.Request) bool {
	ScriptPage := "/" + s.Global.script.AdminPage
	if !strings.HasPrefix(request.URL.Path, ScriptPage) {
		return false
	}
	if request.URL.Path == ScriptPage {
		_, _ = s.RwObj.Write(public.LocalBuildBody("text/html", bytes.ReplaceAll(Resource.FrontendIndex, []byte(`/assets/index`), []byte(ScriptPage+`/assets/index`))))
		return true
	}
	if request.URL.Path == strings.ReplaceAll(ScriptPage+"/WebSocketServer", "//", "/") {
		s.scriptCodeEditServerHandleWebSocket(s.Response.rw, request)
		return true
	}
	if request.URL.Path == strings.ReplaceAll(ScriptPage+"/getEventFunc", "//", "/") {
		data, _ := json.Marshal(Interface.ExportEvent)
		_, _ = s.RwObj.WriteString("HTTP/1.1 200 OK\r\nCache-Control: no-cache, must-revalidate\r\nPragma: no-cache\r\nExpires: 0\r\nContent-Length: ")
		_, _ = s.RwObj.WriteString(fmt.Sprintf("%d\r\nContent-Type:  application/json\r\n\r\n", len(data)))
		_, _ = s.RwObj.Write(data)
		return true
	}

	_FileType := strings.ToLower(request.URL.Path)
	if !strings.HasSuffix(_FileType, ".css") && !strings.HasSuffix(_FileType, ".js") && !strings.HasSuffix(_FileType, ".ttf") {
		fmt.Println(request.URL.Path, "is not support")
		return false
	}
	data, err := Resource.ReadVueFile(strings.ReplaceAll(request.URL.Path, ScriptPage, ""))
	if err != nil {
		fmt.Println(strings.ReplaceAll(request.URL.Path, ScriptPage, ""), err)
		return false
	}
	_, _ = s.RwObj.WriteString("HTTP/1.1 200 OK\r\nCache-Control: no-cache, must-revalidate\r\nPragma: no-cache\r\nExpires: 0\r\nContent-Length: ")
	if strings.HasSuffix(_FileType, ".css") {
		mData := bytes.ReplaceAll(data, []byte("url(/assets/codicon"), []byte(strings.ReplaceAll("url("+ScriptPage+"/assets/codicon", "//", "/")))
		data = mData
		_, _ = s.RwObj.WriteString(fmt.Sprintf("%d\r\nContent-Type:  text/css\r\n\r\n", len(data)))
	}
	if strings.HasSuffix(_FileType, ".js") {
		_, _ = s.RwObj.WriteString(fmt.Sprintf("%d\r\nContent-Type:  application/x-javascript\r\n\r\n", len(data)))
	}
	if strings.HasSuffix(_FileType, ".ttf") {
		_, _ = s.RwObj.WriteString(fmt.Sprintf("%d\r\nContent-Type:  application/application/x-font-ttf\r\n\r\n", len(data)))
	}
	_, _ = s.RwObj.Write(data)
	return true
}
