package httpClient

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
	"unsafe"

	"github.com/qtgolang/SunnyNet/src/SunnyProxy"
	tls "github.com/qtgolang/SunnyNet/src/crypto/tls"
	"github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/loop"
	"github.com/qtgolang/SunnyNet/src/public"
)

var _mustHTTP11 = make(map[uint32]*time.Time)
var _mustHTTP11_lock sync.Mutex

func init() {
	go func() {
		for {
			time.Sleep(time.Minute)
			_mustHTTP11_lock.Lock()
			Nox := time.Now()
			for key, value := range _mustHTTP11 {
				if Nox.Sub(*value) > time.Minute*3 {
					delete(_mustHTTP11, key)
				}
			}
			_mustHTTP11_lock.Unlock()
		}
	}()
}

func Do(req *http.Request, RequestProxy *SunnyProxy.Proxy, CheckRedirect bool, config *tls.Config, outTime time.Duration, GetTLSValues func() []uint16, MConn net.Conn) Result {
	return DoOptions(req, Options{
		RequestProxy:  RequestProxy,
		CheckRedirect: CheckRedirect,
		TLSConfig:     config,
		OutTime:       outTime,
		GetTLSValues:  GetTLSValues,
		MConn:         MConn,
	})
}

// DoOptions 优化参数/返回值与拆分函数
func DoOptions(req *http.Request, opt Options) (r Result) {
	normalizeHTTP2Body(req)    // HTTP/2 下清理不允许携带 Body 的方法
	normalizeCookieHeader(req) // 合并多个 Cookie 头为一个

	cfg := buildTLSConfig(req, Options(opt))    // 克隆并构建 TLS 配置
	_hashCode := applyMustHTTP11(req.Host, cfg) // 根据 host 命中缓存时强制 HTTP/1.1

	handshakeCount := 0 // 握手/连接类错误计数
	for {               // 重试循环开始
		applyTLSValues(cfg, opt) // 动态设置 CipherSuites

		dr := do(doArgs{ // 调用底层 do 执行一次请求
			req:           req,               // 请求对象
			RequestProxy:  opt.RequestProxy,  // 代理配置
			CheckRedirect: opt.CheckRedirect, // 是否允许重定向
			config:        cfg,               // TLS 配置
			outTime:       opt.OutTime,       // 超时时间
			MConn:         opt.MConn,         // 客户端连接
			Event:         opt.Event,
		})
		r.Response, r.Conn, r.Err, r.Close = dr.resp, dr.conn, dr.err, dr.closeFn // 拆包结果

		if r.Err != nil { // 如果请求出错
			closeConnOnErr(r.Conn) // 出错时关闭底层连接

			if needDowngradeHTTP11(r.Err, cfg, _hashCode) { // HTTP2 stream error 降级 HTTP/1.1
				continue // 继续下一次重试
			}

			shouldRetry, shouldReturn := handleRetryableHandshakeError(req, r.Err, &handshakeCount) // 处理握手/连接/EOF 错误
			if shouldReturn {                                                                       // 达到最大重试次数
				r.Close = nil // 禁用 Close 回调
				return        // 直接返回
			}
			if shouldRetry { // 仍可重试
				continue // 进入下一轮
			}
		}
		return // 成功或不可重试错误直接返回
	}
}

func do(a doArgs) (r doResult) {
	req := a.req                     // 请求对象
	RequestProxy := a.RequestProxy   // 代理配置
	CheckRedirect := a.CheckRedirect // 是否允许重定向
	config := a.config               // TLS 配置
	outTime := a.outTime             // 超时时间
	MConn := a.MConn                 // 客户端连接

	if restore := stripContentLengthHeader(req); restore != nil { // 移除 Content-Length 头
		defer restore() // 函数返回前恢复 Content-Length
	}
	if outTime < 100*time.Millisecond { // 超时时间过小
		outTime = 30 * time.Second // 使用默认 30 秒
	}

	client := httpClientGet(req, RequestProxy, config, outTime, a.Event) // 获取 HTTP 客户端
	applyRedirectPolicy(client, CheckRedirect)                           // 设置重定向策略

	cleanup := watchClientConn(req, client, MConn) // 启动客户端连接监控
	defer cleanup()                                // 函数返回时停止监控并清理

	stripTEForHTTP2(client, req) // HTTP/2 场景下删除 TE 头

	r.resp, r.err = client.Client.Do(req)   // 发起 HTTP 请求
	if errors.Is(r.err, context.Canceled) { // 判断是否为取消错误
		r.err = httpCancel // 转换为内部取消错误
	}

	if client.Conn != nil { // 如果底层连接存在
		r.conn = client.Conn // 返回该连接
	}

	setServerIPTag(req, client.RequestProxy.DialAddr) // 写入服务端 IP 信息到 context

	r.closeFn = buildCloseFn(client, r.err) // 构建 Close 回调函数
	return                                  // 返回 doResult
}

var httpCancel = errors.New("客户端取消请求")
var httpLock sync.Mutex
var httpClientMap map[uint32]clientList

type clientList map[uintptr]*clientPart

func httpClientGet(req *http.Request, Proxy *SunnyProxy.Proxy, cfg *tls.Config, timeout time.Duration, event Event) *clientPart {
	httpLock.Lock()
	defer httpLock.Unlock()

	outRouterIP := getOutRouterIP(req)                         //从上下文取出出口路由IP
	keyStr := buildHTTPClientKey(req, Proxy, cfg, outRouterIP) //构建客户端缓存key字符串
	hash := public.SumHashCode(keyStr)                         //计算hash

	if c := tryPopCachedClient(hash, Proxy, timeout); c != nil { //尝试从缓存取出可复用客户端
		return c
	}

	applyTLSProtoCheck(cfg) //设置TLS协议版本校验逻辑

	Tr := newTransport(cfg, timeout)                                   //创建Transport并设置超时
	h2 := configureH2IfNeeded(Tr, cfg)                                 //按NextProtos配置HTTP2
	nproxy, lookupProxy, lookupDial, proxyHost := buildProxySet(Proxy) //构建代理对象与LookupIP拨号器

	cc := http.Client{Transport: Tr, Timeout: timeout}          //创建HTTP客户端
	res := newClientPart(cc, Tr, hash, nproxy, h2, outRouterIP) //创建clientPart

	bindDialContext(dialCtxArgs{
		Tr:         Tr,
		Req:        req,
		Res:        res,
		Timeout:    timeout,
		Event:      event,
		ProxyHost:  proxyHost,
		LookupDial: lookupDial,
	})
	_ = lookupProxy //保持原有变量结构，不改变逻辑意图

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
		_, _ = http.HTTP2configureTransport(Tr)
	}
}

type clientPart struct {
	Client       http.Client
	time         time.Time
	key          uint32
	Conn         net.Conn
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
	if client.Conn != nil {
		loop.Un(client.Conn)
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
				if client.Conn != nil {
					loop.Un(client.Conn)
				}
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
