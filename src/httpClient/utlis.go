package httpClient

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/qtgolang/SunnyNet/src/SunnyProxy"
	"github.com/qtgolang/SunnyNet/src/crypto/tls"
	"github.com/qtgolang/SunnyNet/src/dns"
	"github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/loop"
	"github.com/qtgolang/SunnyNet/src/public"
)

type Event struct {
	Connection func(conn net.Conn)
	Read       func(conn net.Conn, bs []byte)
	Write      func(conn net.Conn, bs []byte)
	Close      func(conn net.Conn)
}
type hConn struct {
	net.Conn
	Event
}

func (h hConn) Read(b []byte) (n int, err error) {
	n, e := h.Conn.Read(b)
	if h.Event.Read != nil {
		h.Event.Read(h.Conn, b[:n])
	}
	return n, e
}

func (h hConn) Write(b []byte) (n int, err error) {
	if h.Event.Write != nil {
		h.Event.Write(h.Conn, b)
	}
	return h.Conn.Write(b)
}

func (h hConn) Close() error {
	if h.Event.Close != nil {
		h.Event.Close(h.Conn)
	}
	return h.Conn.Close()
}

func (h hConn) LocalAddr() net.Addr {
	return h.Conn.LocalAddr()
}

func (h hConn) RemoteAddr() net.Addr {
	return h.Conn.RemoteAddr()
}

func newHConn(event Event, conn net.Conn) net.Conn {
	if conn == nil {
		return nil
	}
	if event.Connection != nil {
		event.Connection(conn)
	}
	return &hConn{Event: event, Conn: conn}
}

// Options 收敛DoOptions的参数
type Options struct {
	RequestProxy  *SunnyProxy.Proxy //代理配置
	CheckRedirect bool              //是否允许重定向
	TLSConfig     *tls.Config       //TLS配置
	OutTime       time.Duration     //超时(注意: do里会强制覆盖成30s)
	GetTLSValues  func() []uint16   //动态CipherSuites
	MConn         net.Conn          //客户端连接(用于探测断开)
	Event         Event
}

// Result 收敛DoOptions的返回值
type Result struct {
	Response *http.Response //响应
	Conn     net.Conn       //底层连接
	Err      error          //错误
	Close    func()         //成功时归还连接池
}

// doArgs 收敛do的参数
type doArgs struct {
	req           *http.Request     //请求
	RequestProxy  *SunnyProxy.Proxy //代理配置
	CheckRedirect bool              //是否允许重定向
	config        *tls.Config       //TLS配置
	outTime       time.Duration     //超时(会在do里覆盖)
	MConn         net.Conn          //客户端连接
	Event         Event
}

// doResult 收敛do的返回值
type doResult struct {
	resp    *http.Response //响应
	conn    net.Conn       //连接
	err     error          //错误
	closeFn func()         //成功时归还
}

// 移除Content-Length并在defer中恢复(保持原逻辑)
func stripContentLengthHeader(req *http.Request) func() {
	if req == nil || req.Header == nil {
		return nil
	}

	ContentLengthName := ""
	var ContentLengthValue []string
	sName := "Content-Length"

	for k, v := range req.Header {
		if strings.EqualFold(k, sName) {
			ContentLengthName = k //保留原本的大小写名称
			ContentLengthValue = v
			break
		}
	}

	if ContentLengthName == "" {
		return nil
	}

	req.Header.Del(sName)
	return func() {
		req.Header.Del(sName)
		req.Header.SetArray(sName, ContentLengthValue)
	}
}

// 设置重定向策略(保持原逻辑)
func applyRedirectPolicy(client *clientPart, checkRedirect bool) {
	if checkRedirect {
		client.Client.CheckRedirect = public.HTTPAllowRedirect
	} else {
		client.Client.CheckRedirect = public.HTTPBanRedirect
	}
}

// HTTP2下删除TE头(保持原逻辑)
func stripTEForHTTP2(client *clientPart, req *http.Request) {
	if client.h2 && req != nil {
		req.Header.Del("TE")
	}
}

// 写入SunnyNetServerIpTags(保持原逻辑)
func setServerIPTag(req *http.Request, dialAddr string) {
	address, proxy, _ := net.SplitHostPort(dialAddr)
	if req != nil {
		ip := net.ParseIP(address)
		if ip == nil {
			req.SetContext(public.SunnyNetServerIpTags, dialAddr)
		} else {
			req.SetContext(public.SunnyNetServerIpTags, SunnyProxy.FormatIP(ip, proxy))
		}
	}
}

// 监控MConn断开并清理资源(保持原逻辑)
func watchClientConn(req *http.Request, client *clientPart, MConn net.Conn) func() {
	if MConn == nil {
		return func() {}
	}
	ticker := time.NewTicker(3 * time.Second)
	stop := make(chan struct{}) // 退出信号
	var mu sync.WaitGroup
	mu.Add(1) // 提前加 1，确保 Done() 被执行
	Cancel := req.WithCancel()

	go func() {
		defer mu.Done()
		ms := make([]byte, 1)
		for {
			select {
			case <-ticker.C:
				_ = MConn.SetReadDeadline(time.Now().Add(1 * time.Millisecond))
				_, er := MConn.Read(ms)
				if er != nil {
					if strings.Contains(er.Error(), "close") {
						if client.Conn != nil {
							Conn := client.Conn
							_ = Conn.Close()
						} else {
							Cancel()
						}
					}
				}
			case <-stop: // 监听退出信号
				return
			}
		}
	}()

	return func() {
		ticker.Stop()
		close(stop)
		mu.Wait()
		_ = MConn.SetDeadline(time.Time{})
	}
}

// 构建do的closeFn(保持原逻辑)
func buildCloseFn(client *clientPart, err error) func() {
	return func() {
		if err != nil {
			return
		}
		httpClientPop(client)
	}
}

// 处理HTTP2下某些方法的Body清理
func normalizeHTTP2Body(req *http.Request) {
	if req.ProtoMajor == 2 {
		switch req.Method {
		case public.HttpMethodHEAD, public.HttpMethodGET, public.HttpMethodTRACE, public.HttpMethodOPTIONS:
			if req.Body != nil {
				_ = req.Body.Close()
				req.Body = nil
			}
		}
	}
}

// 合并重复的Cookie头
func normalizeCookieHeader(req *http.Request) {
	if req != nil && req.Header != nil {
		Cookies := req.Header.GetArray("Cookie")
		if len(Cookies) > 1 {
			req.Header.Set("Cookie", strings.Join(Cookies, "; "))
		}
	}
}

// 克隆并按Scheme调整TLS配置
func buildTLSConfig(req *http.Request, opt Options) *tls.Config {
	cfg := opt.TLSConfig.Clone()
	if req.URL != nil && req.URL.Scheme != "http" {
		if cfg == nil {
			cfg = &tls.Config{}
		}
		cfg.InsecureSkipVerify = true
	}
	return cfg
}

// 按mustHTTP11缓存强制HTTP/1.1并刷新时间
func applyMustHTTP11(host string, cfg *tls.Config) uint32 {
	_hashCode := public.SumHashCode(host)
	_mustHTTP11_lock.Lock()
	if _mustHTTP11[_hashCode] != nil {
		cfg.NextProtos = public.HTTP1NextProtos
		x := time.Now()
		_mustHTTP11[_hashCode] = &x
	}
	_mustHTTP11_lock.Unlock()
	return _hashCode
}

// 每次循环动态刷新CipherSuites
func applyTLSValues(cfg *tls.Config, opt Options) {
	if cfg != nil && opt.GetTLSValues != nil {
		tv := opt.GetTLSValues()
		if len(tv) > 0 {
			cfg.CipherSuites = tv
		}
	}
}

// 处理错误：是否需要HTTP2->HTTP/1.1降级并重试
func needDowngradeHTTP11(err error, cfg *tls.Config, hashCode uint32) bool {
	ers := err.Error()
	if strings.Contains(ers, "stream error: stream ID") && len(cfg.NextProtos) == 2 {
		cfg.NextProtos = public.HTTP1NextProtos
		_mustHTTP11_lock.Lock()
		x := time.Now()
		_mustHTTP11[hashCode] = &x
		_mustHTTP11_lock.Unlock()
		return true
	}
	return false
}

// 处理握手/连接/EOF类错误：是否需要继续重试，以及是否清理HTTP2Config
func handleRetryableHandshakeError(req *http.Request, err error, handshakeCount *int) (shouldRetry bool, shouldReturn bool) {
	ers := err.Error()
	if strings.Contains(ers, "handshake") || strings.Contains(ers, "connection") || strings.Contains(ers, "EOF") {
		*handshakeCount++
		if *handshakeCount > 10 {
			return false, true
		}
		if strings.Contains(ers, "EOF") && *handshakeCount > 3 {
			if req.IsSetHTTP2Config() {
				req.SetHTTP2Config(nil)
			}
		}
		return true, false
	}
	return false, false
}

// 请求失败时关闭连接
func closeConnOnErr(conn net.Conn) {
	if conn != nil {
		_ = conn.Close()
	}
}
func getOutRouterIP(req *http.Request) *net.TCPAddr { //读取出口路由IP
	if req == nil {
		return nil
	}
	outRouterIP, _ := req.Context().Value(public.OutRouterIPKey).(*net.TCPAddr)
	return outRouterIP
}

func buildHTTPClientKey(req *http.Request, Proxy *SunnyProxy.Proxy, cfg *tls.Config, outRouterIP *net.TCPAddr) string { //构建缓存key
	s := dns.GetDnsServer()
	if outRouterIP != nil {
		s += outRouterIP.String() + "|"
	} else {
		s += "|"
	}
	if req != nil && req.URL != nil {
		s += req.URL.Host + "|" + req.Proto + "|" + req.URL.Scheme
	}
	s += "|" + Proxy.String() + "|"
	if cfg != nil {
		if len(cfg.NextProtos) < 1 {
			cfg.NextProtos = []string{http.H11Proto, http.H2Proto}
		}
		s += strings.Join(cfg.NextProtos, "-")
	}
	return s
}

func tryPopCachedClient(hash uint32, Proxy *SunnyProxy.Proxy, timeout time.Duration) *clientPart { //从缓存取client
	if clients, ok := httpClientMap[hash]; ok {
		if len(clients) > 0 {
			for key, client := range clients {
				delete(clients, key)
				refreshClientProxy(client, Proxy)          //刷新代理参数
				refreshClientConnDeadline(client, timeout) //刷新连接deadline与transport超时
				return client
			}
		}
	}
	return nil
}

func refreshClientProxy(client *clientPart, Proxy *SunnyProxy.Proxy) { //刷新RequestProxy
	var nproxy *SunnyProxy.Proxy
	if Proxy != nil {
		nproxy = Proxy.Clone()
	} else {
		nproxy = new(SunnyProxy.Proxy)
	}
	if client.RequestProxy != nil {
		nproxy.DialAddr = client.RequestProxy.DialAddr
	}
	client.RequestProxy = nproxy
}

func refreshClientConnDeadline(client *clientPart, timeout time.Duration) { //刷新连接deadline与transport超时
	if client.Conn != nil {
		Conn := client.Conn
		if timeout == 0 {
			_ = Conn.SetDeadline(time.Time{})
			_ = Conn.SetWriteDeadline(time.Time{})
			_ = Conn.SetDeadline(time.Time{})
		} else {
			_ = Conn.SetDeadline(time.Now().Add(timeout))
			_ = Conn.SetWriteDeadline(time.Now().Add(timeout))
			_ = Conn.SetDeadline(time.Now().Add(timeout))
		}
		client.Client.Timeout = 24 * time.Hour
		client.Transport.ResponseHeaderTimeout = 24 * time.Hour // 读取响应头超时
		client.Transport.IdleConnTimeout = 24 * time.Hour       // 空闲连接超时
		client.Transport.TLSHandshakeTimeout = 24 * time.Hour   // TLS 握手超时
	}
}

func applyTLSProtoCheck(cfg *tls.Config) { //设置TLS协议版本校验
	if cfg != nil {
		if len(cfg.NextProtos) > 0 {
			cfg.GetConfigForServer = func(info *tls.ServerHelloMsg) error {
				for _, proto := range cfg.NextProtos {
					if proto == http.H2Proto && info.SupportedVersion == 772 {
						return nil // 如果支持，则返回 nil
					}
					if proto == http.H11Proto && (info.SupportedVersion == 0 || info.Vers == 771) {
						return nil // 如果支持，则返回 nil
					}
				}
				ver := info.SupportedVersion
				if ver == 0 {
					ver = info.Vers
				}
				Proto, _ := http.ProtoVersions[info.Vers]
				if Proto == "" {
					return fmt.Errorf("服务器不支持您所选HTTP协议版本")
				}
				return fmt.Errorf("服务器不支持您所选HTTP协议版本: 需要协议[%s],请检查您的配置", strings.ToUpper(Proto))
			}
		}
	}
}

func newTransport(cfg *tls.Config, timeout time.Duration) *http.Transport { //创建Transport并设置超时
	Tr := &http.Transport{TLSClientConfig: cfg}
	if timeout == 0 {
		Tr.ResponseHeaderTimeout = 60 * time.Second // 读取响应头超时
		Tr.IdleConnTimeout = 60 * time.Second       // 空闲连接超时
		Tr.TLSHandshakeTimeout = 60 * time.Second   // TLS 握手超时
	} else {
		Tr.ResponseHeaderTimeout = timeout // 读取响应头超时
		Tr.IdleConnTimeout = timeout       // 空闲连接超时
		Tr.TLSHandshakeTimeout = timeout   // TLS 握手超时
	}
	return Tr
}

func configureH2IfNeeded(Tr *http.Transport, cfg *tls.Config) bool { //按NextProtos配置HTTP2
	h2 := false
	if cfg != nil {
		if len(cfg.NextProtos) < 1 {
			configureHTTP2Transport(Tr, cfg)
			h2 = true
		} else {
			for _, proto := range cfg.NextProtos {
				if proto == http.H2Proto {
					configureHTTP2Transport(Tr, cfg)
					h2 = true
					break
				}
			}
		}
	}
	return h2
}

func buildProxySet(Proxy *SunnyProxy.Proxy) (nproxy *SunnyProxy.Proxy, lookupProxy *SunnyProxy.Proxy, lookupDial func(network string, addr string, OutRouterIP *net.TCPAddr) (net.Conn, error), proxyHost string) { //构建代理与LookupIP拨号
	if Proxy != nil {
		nproxy = Proxy.Clone()
		lookupProxy = Proxy.Clone()
		lookupDial = lookupProxy.Dial
		proxyHost = Proxy.Host
	} else {
		nproxy = new(SunnyProxy.Proxy)
		lookupDial = lookupProxy.Dial
	}
	return
}

func newClientPart(cc http.Client, Tr *http.Transport, hash uint32, nproxy *SunnyProxy.Proxy, h2 bool, outRouterIP *net.TCPAddr) *clientPart { //创建clientPart
	res := &clientPart{Client: cc, key: hash, RequestProxy: nproxy, Transport: Tr, h2: h2}
	if outRouterIP != nil {
		res.outRouterIP = &net.TCPAddr{IP: outRouterIP.IP}
	}
	return res
}

// dialCtxArgs 收敛bindDialContext入参
type dialCtxArgs struct {
	Tr         *http.Transport                                                               //Transport对象
	Req        *http.Request                                                                 //请求对象
	Res        *clientPart                                                                   //客户端结构
	Timeout    time.Duration                                                                 //超时
	Event      Event                                                                         //事件回调
	ProxyHost  string                                                                        //代理Host
	LookupDial func(network string, addr string, OutRouterIP *net.TCPAddr) (net.Conn, error) //DNS回源拨号
}

// bindDialContext 参数收敛版(不改内部逻辑)
func bindDialContext(a dialCtxArgs) { //绑定拨号逻辑
	var ips []net.IP
	var isLookupIP bool
	var retries bool

	Tr := a.Tr
	req := a.Req
	res := a.Res
	timeout := a.Timeout
	event := a.Event
	proxyHost := a.ProxyHost
	lookupDial := a.LookupDial

	Tr.DialContext = func(ctx context.Context, network, addr string) (cnn net.Conn, _ error) {
		defer func() {
			if cnn != nil {
				attachConnAndTimeout(res, Tr, &res.Client, cnn, timeout) //绑定连接与超时设置
			}
		}()

		if conn, er, ok := tryDialRawServerIP(req, res, network, event); ok { //优先按上下文指定IP连接
			return conn, er
		}

		if dns.IsRemoteDnsServer() { //远程DNS直接拨号
			conn, er := res.RequestProxy.DialWithTimeout(network, addr, 3*time.Second, res.outRouterIP)
			return newHConn(event, conn), er
		}

		address, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		if conn, er, ok := tryDialDirectIP(res, network, address, port, event); ok { //addr就是IP时直接拨号
			return conn, er
		}

		if strings.ToLower(address) == "localhost" { //localhost强制走127.0.0.1
			r, e := res.RequestProxy.Dial(network, "127.0.0.1:"+port, res.outRouterIP)
			return newHConn(event, r), e
		}

		for { //域名解析后尝试多IP
			if !isLookupIP {
				isLookupIP = true
				first := dns.GetFirstIP(address, proxyHost)
				if first != nil {
					if first.To4() != nil {
						r, e := res.RequestProxy.Dial(network, fmt.Sprintf("%s:%s", first.String(), port), res.outRouterIP)
						return newHConn(event, r), e
					} else {
						r, e := res.RequestProxy.Dial(network, fmt.Sprintf("[%s]:%s", first.String(), port), res.outRouterIP)
						return newHConn(event, r), e
					}
				}
				ips, _ = dns.LookupIP(address, proxyHost, res.outRouterIP, lookupDial)
				if len(ips) < 1 {
					return nil, noIP
				}
			}

			if len(ips) < 1 {
				dns.SetFirstIP(address, proxyHost, nil)
				if retries {
					return nil, connectionFailed
				}
				isLookupIP = false
				retries = true
				continue
			}

			if allLocal127(ips) && len(ips) != 0 {
				return nil, errors.New(fmt.Sprintf("Address [%s] points to 127.0.0.1", address))
			}

			ip := extractAndRemoveIP(&ips)
			if ip != nil && ip.String() != "127.0.0.1" {
				if ip.To4() != nil {
					conn, er := res.RequestProxy.DialWithTimeout(network, fmt.Sprintf("%s:%s", ip.String(), port), 2*time.Second, res.outRouterIP)
					if conn != nil {
						dns.SetFirstIP(address, proxyHost, ip)
						return newHConn(event, conn), er
					}
					continue
				}
				conn, er := res.RequestProxy.DialWithTimeout(network, fmt.Sprintf("[%s]:%s", ip.String(), port), 2*time.Second, res.outRouterIP)
				if conn != nil {
					dns.SetFirstIP(address, proxyHost, ip)
					return newHConn(event, conn), er
				}
			}
		}
	}
}

func attachConnAndTimeout(res *clientPart, Tr *http.Transport, cc *http.Client, cnn net.Conn, timeout time.Duration) { //连接成功后设置超时并缓存
	res.Conn = cnn
	loop.Add(cnn)
	if timeout != 0 {
		_ = cnn.SetDeadline(time.Now().Add(timeout))
		_ = cnn.SetWriteDeadline(time.Now().Add(timeout))
		_ = cnn.SetDeadline(time.Now().Add(timeout))
	} else {
		_ = cnn.SetDeadline(time.Time{})
		_ = cnn.SetWriteDeadline(time.Time{})
		_ = cnn.SetDeadline(time.Time{})
	}
	Tr.ResponseHeaderTimeout = 24 * time.Hour // 读取响应头超时
	Tr.IdleConnTimeout = 24 * time.Hour       // 空闲连接超时
	Tr.TLSHandshakeTimeout = 24 * time.Hour   // TLS 握手超时
	cc.Timeout = 24 * time.Hour
}

func tryDialRawServerIP(req *http.Request, res *clientPart, network string, event Event) (net.Conn, error, bool) { //按上下文指定IP连接
	serveripFunc, ok := req.Context().Value(public.Connect_Raw_Address).(func() string)
	if ok && serveripFunc != nil {
		_serverIP_ := serveripFunc()
		if _serverIP_ != "" {
			address2, _, err2 := net.SplitHostPort(_serverIP_)
			if err2 == nil {
				ip := net.ParseIP(address2)
				if ip != nil {
					conn, er := res.RequestProxy.DialWithTimeout(network, _serverIP_, 3*time.Second, res.outRouterIP)
					if conn != nil {
						return newHConn(event, conn), er, true
					}
				}
			}
		}
	}
	return nil, nil, false
}

func tryDialDirectIP(res *clientPart, network, address, port string, event Event) (net.Conn, error, bool) { //addr本身是IP时直接拨号
	i := net.ParseIP(address)
	if i == nil {
		return nil, nil, false
	}
	if len(i) == net.IPv4len {
		r, e := res.RequestProxy.Dial(network, i.String()+":"+port, res.outRouterIP)
		return newHConn(event, r), e, true
	}
	r, e := res.RequestProxy.Dial(network, fmt.Sprintf("[%s]:%s", address, port), res.outRouterIP)
	return newHConn(event, r), e, true
}

func allLocal127(ips []net.IP) bool { //判断是否全是127.0.0.1
	for _, ip := range ips {
		if ip.String() != "127.0.0.1" {
			return false
		}
	}
	return true
}
