package SunnyNet

import (
	"bytes"
	"github.com/qtgolang/SunnyNet/src/CrossCompiled"
	"github.com/qtgolang/SunnyNet/src/Interface"
	"github.com/qtgolang/SunnyNet/src/SunnyProxy"
	"github.com/qtgolang/SunnyNet/src/crypto/tls"
	"github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/public"
	"io"
	"net/url"
	"strconv"
)

type ConnHTTP Interface.ConnHTTPCall

type httpConn struct {
	_Context              int
	_Theology             int               //唯一ID
	_MessageId            int               //消息ID,仅标识消息ID,不能用于API函数
	_PID                  int               //请求进程PID， 为0 表示 非本机设备通过代理连接
	_Type                 int               //请求类型 例如 public.HttpSendRequest  public.Http....
	_ClientIP             string            //来源IP地址,请求从哪里来
	_request              *http.Request     //请求体
	_response             *http.Response    //响应体
	_err                  string            //错误信息
	_proxy                *SunnyProxy.Proxy //代理信息
	_getRawBody           func(path string) bool
	_Display              bool
	_Break                bool
	_tls                  *tls.Config
	_serverIP             string
	_isRandomCipherSuites bool
	_localAddress         string
	_OutRouterIPFunc      func(string) bool
	updateRawTarget       func(int uint32)
}

func (k *httpConn) SetOutRouterIP(way string) bool {
	if k._OutRouterIPFunc != nil {
		return k._OutRouterIPFunc(way)
	}
	return false
}
func (h *httpConn) LocalAddress() string {
	return h._localAddress
}

func (h *httpConn) GetSocket5User() string {
	return GetSocket5User(h._Theology)
}

func (h *httpConn) ServerAddress() string {
	if h._Type != public.HttpResponseOK {
		return ""
	}
	return h._serverIP
}

func (h *httpConn) RandomCipherSuites() {
	h._isRandomCipherSuites = true
}

func (h *httpConn) UpdateURL(NewUrl string) bool {
	if h == nil {
		return false
	}
	if h._Type != public.HttpSendRequest {
		return false
	}
	if h._request == nil {
		return false
	}
	if h._request.URL == nil {
		return false
	}
	a, _ := url.Parse(NewUrl)
	if a == nil {
		return false
	}
	h._request.URL = a
	h._request.Host = h._request.URL.Host
	h._request.RequestURI = ""
	if h.updateRawTarget != nil {
		h.updateRawTarget(0)
	}
	h._request.SetContext(public.Connect_Raw_Address, func() string { return a.Host })
	if h._request.Header.Get("host") != "" {
		h._request.Header.Set("host", h._request.URL.Host)
	}
	return true
}

func (h *httpConn) SetHTTP2Config(h2Config string) bool {
	if h == nil {
		return false
	}
	if h._Type != public.HttpSendRequest {
		return false
	}
	if h._request == nil {
		return false
	}
	if h._tls == nil {
		h._request.SetHTTP2Config(nil)
		return false
	}
	h._tls.NextProtos = public.HTTP2NextProtos
	if h2Config != "" {
		c, e := http.StringToH2Config(h2Config)
		if e != nil {
			h._request.SetHTTP2Config(nil)
			return false
		}
		h._request.SetHTTP2Config(c)
		return true
	}
	return false
}

func (h *httpConn) GetProcessName() string {
	if h._PID == 0 {
		return "代理连接"
	}
	return CrossCompiled.GetPidName(int32(h._PID))
}
func (h *httpConn) GetResponseProto() string {
	if h == nil {
		return ""
	}
	if h._response == nil {
		return ""
	}
	return h._response.Proto
}

/*
SetBreak
设置是否需要通知回调拦截该请求(默认为false)
如果设置为true 回调函数中的err为Debug字符串,表示请求需要拦截

仅在脚本代码中使用
*/
func (h *httpConn) SetBreak(Break bool) {
	h._Break = Break
}

/*
SetDisplay
是否显示请求信息
默认为true
如果设置为false 将不再得到Call回调消息
仅在脚本代码中使用，且 仅在发起请求时有效
*/
func (h *httpConn) SetDisplay(Display bool) {
	h._Display = Display
}
func (h *httpConn) Context() int {
	return h._Context
}
func (h *httpConn) MessageId() int {
	return h._MessageId
}

func (h *httpConn) PID() int {
	return h._PID
}

func (h *httpConn) Theology() int {
	return h._Theology
}

func (h *httpConn) Type() int {
	return h._Type
}

func (h *httpConn) ClientIP() string {
	return h._ClientIP
}

// StopRequest 阻止请求,仅支持在发起请求时使用
// StatusCode要响应的状态码
// Data=要响应的数据 可以是string 也可以是[]byte
// Header=要响应的Header 可以忽略
func (h *httpConn) StopRequest(StatusCode int, Data any, Header ...http.Header) {
	var ResponseData []byte
	switch v := Data.(type) {
	case string:
		ResponseData = []byte(v)
		break
	case []byte:
		ResponseData = v
		break
	default:
		return
	}
	h._response = new(http.Response)
	if StatusCode < 100 {
		h._response.StatusCode = 200
	} else {
		h._response.StatusCode = StatusCode
	}
	h._response.Body = io.NopCloser(bytes.NewBuffer(ResponseData))
	if len(Header) > 0 {
		h._response.Header = Header[0]
	}
	if h._response.Header == nil {
		h._response.Header = make(http.Header)
		h._response.Header.Set("Server", "Sunny")
		h._response.Header.Set("Accept-Ranges", "bytes")
		h._response.Header.Set("Connection", "Close")
	}
	h._response.Header.Set("Content-Length", strconv.Itoa(len(ResponseData)))
	h._response.ContentLength = int64(len(ResponseData))
}

// GetError 获取错误信息
func (h *httpConn) Error() string {
	return h._err
}

// URL 获取请求地址
func (h *httpConn) URL() string {
	if h == nil {
		return ""
	}
	if h._request == nil {
		return ""
	}
	if h._request.URL == nil {
		return ""
	}
	return h._request.URL.String()
}

// Proto 获取请求协议
func (h *httpConn) Proto() string {
	if h == nil {
		return ""
	}
	if h._request == nil {
		return ""
	}
	return h._request.Proto
}

// Method 获取请求方法
func (h *httpConn) Method() string {
	if h == nil {
		return ""
	}
	if h._request == nil {
		return ""
	}
	return h._request.Method
}

// GetRequestHeader 获取请求头
func (h *httpConn) GetRequestHeader() http.Header {
	if h == nil {
		return make(http.Header)
	}
	if h._request == nil {
		return make(http.Header)
	}
	return h._request.Header
}

// GetRequestBody 获取请求提交内容,当请求提交数据超过一定大小时,使用GetRawBody
func (h *httpConn) GetRequestBody() []byte {
	if h == nil {
		return nil
	}
	if h._request == nil {
		return nil
	}
	return h._request.GetData()
}

// SetRequestBody 修改请求提交内容
func (h *httpConn) SetRequestBody(data []byte) bool {
	if h == nil {
		return false
	}
	if h._Type != public.HttpSendRequest {
		return false
	}
	if h._request == nil {
		return false
	}
	h._request.SetData(data)
	return true
}

// SetRequestBodyIO 修改请求提交内容
func (h *httpConn) SetRequestBodyIO(data io.ReadCloser) bool {
	if h == nil {
		return false
	}
	if h._Type != public.HttpSendRequest {
		return false
	}
	if h._request == nil {
		return false
	}
	if h._request.IsRawBody {
		return false
	}
	if h._request.Body != nil {
		_, _ = io.ReadAll(h._request.Body)
		_ = h._request.Body.Close()
	}
	h._request.Body = data
	return true
}

// SaveRawRequestData 获取请求提交的原始内容,当请求提交的原始数据超过一定大小时使用
func (h *httpConn) SaveRawRequestData(SaveFilePath string) bool {
	if h == nil {
		return false
	}
	if h._getRawBody != nil {
		return h._getRawBody(SaveFilePath)
	}
	return false
}

// IsRawRequestBody 当前请求是否转发Body模式
func (h *httpConn) IsRawRequestBody() bool {
	if h == nil {
		return false
	}
	if h._request != nil {
		return h._request.IsRawBody
	}
	return false
}

// SetAgent 设置HTTP/S请求代理，仅支持Socket5和http 例如 socket5://admin:123456@127.0.0.1:8888 或 http://admin:123456@127.0.0.1:8888
func (h *httpConn) SetAgent(ProxyUrl string, timeout ...int) bool {
	if h == nil {
		return false
	}
	if h._Type != public.HttpSendRequest {
		return false
	}
	h._proxy, _ = SunnyProxy.ParseProxy(ProxyUrl, timeout...)
	ok := h._proxy != nil
	return ok
}

// GetResponseHeader 获取响应协议头
func (h *httpConn) GetResponseHeader() http.Header {
	if h == nil {
		return make(http.Header)
	}
	if h._response == nil {
		h._response = new(http.Response)
	}
	if h._response.Header == nil {
		h._response.Header = make(http.Header)
	}
	return h._response.Header
}

// GetResponseBody 获取响应内容
func (h *httpConn) GetResponseBody() []byte {
	if h == nil {
		return nil
	}
	if h._response == nil {
		return nil
	}
	return h._response.GetData()
}

// SetResponseBody 修改响应内容
func (h *httpConn) SetResponseBody(data []byte) bool {
	if h == nil {
		return false
	}
	if h._response == nil {
		h._response = new(http.Response)
	}
	h._response.SetData(data)
	return true
}

// SetResponseBodyIO 修改响应内容
func (h *httpConn) SetResponseBodyIO(data io.ReadCloser) bool {
	if h == nil {
		return false
	}
	if h._response == nil {
		h._response = new(http.Response)
	}
	if h._response.Body != nil {
		_ = h._response.Body.Close()
	}
	h._response.Body = data
	return true
}

// GetResponseCode 获取响应状态码
func (h *httpConn) GetResponseCode() int {
	if h == nil {
		return 0
	}
	if h._response == nil {
		return 0
	}
	return h._response.StatusCode
}

// SetResponseCode 修改响应状态码
func (h *httpConn) SetResponseCode(code int) bool {
	if h == nil {
		return false
	}
	if h._response == nil {
		h._response = new(http.Response)
	}
	h._response.StatusCode = code
	return true
}
