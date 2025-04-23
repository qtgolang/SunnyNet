package httpClient

import (
	"context"
	"errors"
	"fmt"
	"github.com/qtgolang/SunnyNet/src/SunnyProxy"
	tls "github.com/qtgolang/SunnyNet/src/crypto/tls"
	"github.com/qtgolang/SunnyNet/src/dns"
	"github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/public"
	"net"
	"strings"
	"sync"
	"time"
	"unsafe"
)

func Do(req *http.Request, RequestProxy *SunnyProxy.Proxy, CheckRedirect bool, config *tls.Config, outTime time.Duration, GetTLSValues func() []uint16, MConn net.Conn) (Response *http.Response, Conn net.Conn, err error, Close func()) {
	if req.ProtoMajor == 2 {
		Method := req.Method
		switch Method {
		case public.HttpMethodHEAD:
			fallthrough
		case public.HttpMethodGET:
			fallthrough
		case public.HttpMethodTRACE:
			fallthrough
		case public.HttpMethodOPTIONS:
			if req.Body != nil {
				_ = req.Body.Close()
				req.Body = nil
			}
		default:
			break
		}
	}
	{
		if req != nil && req.Header != nil {
			Cookies := req.Header.GetArray("Cookie")
			if len(Cookies) > 1 {
				req.Header.Set("Cookie", strings.Join(Cookies, "; "))
			}
		}
	}
	cfg := config.Clone()
	if req.URL != nil && req.URL.Scheme != "http" {
		if cfg == nil {
			cfg = &tls.Config{}
		}
		cfg.InsecureSkipVerify = true
	}
	handshakeCount := 0
	for {
		if cfg != nil && GetTLSValues != nil {
			tv := GetTLSValues()
			if len(tv) > 0 {
				cfg.CipherSuites = tv
			}
		}
		Response, Conn, err, Close = do(req, RequestProxy, CheckRedirect, cfg, outTime, MConn)
		if err != nil {
			if Conn != nil {
				_ = Conn.Close()
			}
			ers := err.Error()
			if strings.Contains(ers, "handshake") || strings.Contains(ers, "connection") || strings.Contains(ers, "EOF") {
				handshakeCount++
				if handshakeCount > 10 {
					Close = nil
					return
				}
				if strings.Contains(ers, "EOF") && handshakeCount > 3 {
					if req.IsSetHTTP2Config() {
						req.SetHTTP2Config(nil)
					}
				}
				continue
			}
		}
		return
	}
}
func do(req *http.Request, RequestProxy *SunnyProxy.Proxy, CheckRedirect bool, config *tls.Config, outTime time.Duration, MConn net.Conn) (*http.Response, net.Conn, error, func()) {

	SendTimeout := 30 * 1000 * time.Millisecond
	outTime = SendTimeout
	client := httpClientGet(req, RequestProxy, config, outTime)
	if CheckRedirect {
		client.Client.CheckRedirect = public.HTTPAllowRedirect
	} else {
		client.Client.CheckRedirect = public.HTTPBanRedirect
	}
	if MConn != nil {
		//防止客户端与 SunnyNet 断开连接，但是 SunnyNet 与 目标服务器 一直交互
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
								Conn := *client.Conn
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
		defer func() {
			ticker.Stop()
			close(stop)
			mu.Wait()
			_ = MConn.SetDeadline(time.Time{})
		}()
	}
	if client.h2 && req != nil {
		//部分HTTP2服务器不支持此协议头,导致出现协议错误
		req.Header.Del("TE")
	}
	reqs, err := client.Client.Do(req)
	if errors.Is(err, context.Canceled) {
		err = httpCancel
	}
	var rConn net.Conn
	if client.Conn != nil {
		rConn = *client.Conn
	}
	address, proxy, _ := net.SplitHostPort(client.RequestProxy.DialAddr)
	ip := net.ParseIP(address)
	if ip == nil {
		req.SetContext(public.SunnyNetServerIpTags, client.RequestProxy.DialAddr)
	} else {
		req.SetContext(public.SunnyNetServerIpTags, SunnyProxy.FormatIP(ip, proxy))
	}
	return reqs, rConn, err, func() { httpClientPop(client) }
}

var httpCancel = errors.New("客户端取消请求")
var httpLock sync.Mutex
var httpClientMap map[uint32]clientList

type clientList map[uintptr]*clientPart

func hashCode(s string) uint32 {
	var hash int32 = 0
	for _, ch := range s {
		hash = 31*hash + ch
	}
	return uint32(hash)
}
func httpClientGet(req *http.Request, Proxy *SunnyProxy.Proxy, cfg *tls.Config, timeout time.Duration) *clientPart {
	httpLock.Lock()
	defer httpLock.Unlock()
	outRouterIP, _ := req.Context().Value(public.OutRouterIPKey).(*net.TCPAddr)
	s := ""
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
	hash := hashCode(s)
	if clients, ok := httpClientMap[hash]; ok {
		if len(clients) > 0 {
			for key, client := range clients {
				delete(clients, key)
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
				if client.Conn != nil {
					Conn := *client.Conn
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
				return client
			}
		}
	}
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
	var ips []net.IP
	var isLookupIP bool
	var ProxyHost string
	var LookupIPdial func(network string, addr string, OutRouterIP *net.TCPAddr) (net.Conn, error)
	var nproxy *SunnyProxy.Proxy
	var LookupIPproxy *SunnyProxy.Proxy
	if Proxy != nil {
		nproxy = Proxy.Clone()
		LookupIPproxy = Proxy.Clone()
		LookupIPdial = LookupIPproxy.Dial
		ProxyHost = Proxy.Host
	} else {
		nproxy = new(SunnyProxy.Proxy)
		LookupIPdial = LookupIPproxy.Dial
	}
	cc := http.Client{Transport: Tr, Timeout: timeout}
	res := &clientPart{Client: cc, key: hash, RequestProxy: nproxy, Transport: Tr, h2: h2}
	if outRouterIP != nil {
		res.outRouterIP = &net.TCPAddr{IP: outRouterIP.IP}
	}
	Tr.DialContext = func(ctx context.Context, network, addr string) (cnn net.Conn, _ error) {
		defer func() {
			if cnn != nil {
				res.Conn = &cnn
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
		}()
		_serverIP_func, ok := req.Context().Value(public.Connect_Raw_Address).(func() string)
		if ok && _serverIP_func != nil {
			_serverIP_ := _serverIP_func()
			if _serverIP_ != "" {
				address2, _, err2 := net.SplitHostPort(_serverIP_)
				if err2 == nil {
					ip := net.ParseIP(address2)
					if ip != nil {
						conn, er := res.RequestProxy.DialWithTimeout(network, _serverIP_, 3*time.Second, res.outRouterIP)
						if conn != nil {
							return conn, er
						}
					}
				}
			}
		}
		address, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		i := net.ParseIP(address)
		if i != nil {
			if len(i) == net.IPv4len {
				return res.RequestProxy.Dial(network, i.String()+":"+port, res.outRouterIP)
			}
			return res.RequestProxy.Dial(network, fmt.Sprintf("[%s]:%s", address, port), res.outRouterIP)
		}
		if strings.ToLower(address) == "localhost" {
			return res.RequestProxy.Dial(network, "127.0.0.1:"+port, res.outRouterIP)
		}

		var retries bool
		for {
			if !isLookupIP {
				isLookupIP = true
				first := dns.GetFirstIP(address, ProxyHost)
				if first != nil {
					if first.To4() != nil {
						return res.RequestProxy.Dial(network, fmt.Sprintf("%s:%s", first.String(), port), res.outRouterIP)
					} else {
						return res.RequestProxy.Dial(network, fmt.Sprintf("[%s]:%s", first.String(), port), res.outRouterIP)
					}
				}
				ips, _ = dns.LookupIP(address, ProxyHost, res.outRouterIP, LookupIPdial)
				if len(ips) < 1 {
					return nil, noIP
				}
			}
			if len(ips) < 1 {
				dns.SetFirstIP(address, ProxyHost, nil)
				if retries {
					return nil, connectionFailed
				}
				isLookupIP = false
				retries = true
				continue
			}
			var AllLocalIP = true
			for _, ip := range ips {
				if ip.String() != "127.0.0.1" {
					AllLocalIP = false
					break
				}
			}
			if AllLocalIP && len(ips) != 0 {
				return nil, errors.New(fmt.Sprintf("Address [%s] points to 127.0.0.1", address))
			}
			ip := extractAndRemoveIP(&ips)
			if ip != nil && ip.String() != "127.0.0.1" {
				if ip.To4() != nil {
					conn, er := res.RequestProxy.DialWithTimeout(network, fmt.Sprintf("%s:%s", ip.String(), port), 2*time.Second, res.outRouterIP)
					if conn != nil {
						dns.SetFirstIP(address, ProxyHost, ip)
						return conn, er
					}
					continue
				}
				conn, er := res.RequestProxy.DialWithTimeout(network, fmt.Sprintf("[%s]:%s", ip.String(), port), 2*time.Second, res.outRouterIP)
				if conn != nil {
					dns.SetFirstIP(address, ProxyHost, ip)
					return conn, er
				}
			}
		}
	}
	return res
}

// 优先使用IPV4
func extractAndRemoveIP(ips *[]net.IP) net.IP {
	for i, ip := range *ips {
		if ip.To4() != nil { // 检查是否为 IPv4
			// 找到 IPv4，删除并返回
			*ips = append((*ips)[:i], (*ips)[i+1:]...) // 删除
			return ip
		}
	}
	// 如果没有找到 IPv4，查找 IPv6
	for i, ip := range *ips {
		if ip.To16() != nil { // IPv6 的情况
			*ips = append((*ips)[:i], (*ips)[i+1:]...) // 删除
			return ip
		}
	}
	return nil
}

var noIP = errors.New("DNS解析失败,无可用IP地址")
var connectionFailed = errors.New("连接 DNS解析的所有IP地址 都失败了")

func configureHTTP2Transport(Tr *http.Transport, cfg *tls.Config) {
	// 检查是否配置了 HTTP/2.0 协议
	protoFound := false
	for _, proto := range cfg.NextProtos {
		if proto == http.H2Proto {
			protoFound = true
			break
		}
	}

	// 如果找到了 HTTP/2.0 协议，则配置 HTTP/2.0 传输
	if protoFound || len(cfg.NextProtos) == 0 {
		http.HTTP2configureTransport(Tr)
	}
}

type clientPart struct {
	Client       http.Client
	time         time.Time
	key          uint32
	Conn         *net.Conn
	Transport    *http.Transport
	RequestProxy *SunnyProxy.Proxy
	h2           bool
	outRouterIP  *net.TCPAddr
}

func httpClientPop(client *clientPart) {
	if client == nil || client.key == 0 {
		return
	}
	httpLock.Lock()
	defer httpLock.Unlock()
	client.time = time.Now()
	clients := httpClientMap[client.key]
	if clients == nil {
		httpClientMap[client.key] = make(clientList)
		clients = httpClientMap[client.key]
	}
	clients[uintptr(unsafe.Pointer(client))] = client
}
func httpClientClear() {
	httpLock.Lock()
	defer httpLock.Unlock()
	t := time.Now()
	o := 5 * time.Second
	for k, clients := range httpClientMap {
		for key, client := range clients {
			if t.Sub(client.time) > o {
				delete(clients, key)
			}
		}
		if len(clients) == 0 {
			delete(httpClientMap, k)
		}
	}
}
func init() {
	httpClientMap = make(map[uint32]clientList)
	go func() {
		for {
			time.Sleep(time.Second * 3)
			httpClientClear()
		}
	}()
}
