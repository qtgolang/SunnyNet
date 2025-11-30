package SunnyProxy

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/qtgolang/SunnyNet/src/crypto/tls"
	"golang.org/x/net/proxy"
)

var dnsConfig = &tls.Config{
	ClientSessionCache: tls.NewLRUClientSessionCache(32),
	InsecureSkipVerify: true,
}
var invalidProxy = fmt.Errorf("invalid host")

type Proxy struct {
	*url.URL
	timeout  time.Duration
	Regexp   func(Host string) bool
	DialAddr string
}

func ParseProxy(u string, timeout ...int) (*Proxy, error) {
	var err error
	p := &Proxy{}
	p.URL, err = url.Parse(u)
	if err != nil {
		return nil, err
	}
	if p.URL == nil {
		return nil, invalidProxy
	}
	Scheme := strings.ToLower(p.URL.Scheme)
	if Scheme != "http" && Scheme != "https" && Scheme != "socket" && Scheme != "sock" && Scheme != "socket5" && Scheme != "socks5" && Scheme != "socks" {
		return nil, fmt.Errorf("invalid scheme: %s", p.URL.Scheme)
	}
	if Scheme == "socket" || Scheme == "sock" || Scheme == "socket5" || Scheme == "socks5" || Scheme == "socks" {
		p.URL.Scheme = "socks5"
	}
	if p.Host == "null" {
		p.Host = ""
		p.URL = nil
		return p, nil
	}
	if len(p.Host) < 3 {
		return nil, invalidProxy
	}

	p.timeout = 30 * time.Second
	if len(timeout) > 0 {
		if timeout[0] > 0 {
			p.timeout = time.Duration(timeout[0]) * time.Millisecond
		}
	}
	return p, err
}
func (p *Proxy) IsSocksType() bool {
	if p == nil {
		return false
	}
	if p.URL == nil {
		return false
	}
	return p.URL.Scheme == "socks5"
}
func (p *Proxy) String() string {
	if p == nil {
		return ""
	}
	if p.URL == nil {
		return ""
	}
	return p.URL.String()
}
func (p *Proxy) User() string {
	if p == nil {
		return ""
	}
	if p.URL == nil {
		return ""
	}
	if p.URL.User == nil {
		return ""
	}
	return p.URL.User.Username()
}
func (p *Proxy) Pass() string {
	if p == nil {
		return ""
	}
	if p.URL == nil {
		return ""
	}
	if p.URL.User == nil {
		return ""
	}
	pass, _ := p.URL.User.Password()
	return pass
}
func (p *Proxy) Clone() *Proxy {
	if p == nil {
		return nil
	}
	if p.URL == nil {
		return nil
	}
	if len(p.Host) < 3 {
		return nil
	}
	n := &Proxy{}

	n.URL, _ = url.Parse(p.URL.String())
	n.timeout = p.timeout
	n.Regexp = p.Regexp
	n.DialAddr = p.DialAddr
	return n
}
func (p *Proxy) SetTimeout(d time.Duration) {
	if p == nil {
		return
	}
	p.timeout = d
	return
}
func (p *Proxy) getTimeout() time.Duration {
	if p == nil || p.timeout == 0 {
		return 15 * time.Second
	}
	return p.timeout
}
func (p *Proxy) getSocksAuth() *proxy.Auth {
	if p.User() == "" {
		return nil
	}
	return &proxy.Auth{
		User:     p.User(),
		Password: p.Pass(),
	}
}
func (p *Proxy) DialWithTimeout(network, addr string, Timeout time.Duration, OutRouterIP *net.TCPAddr) (net.Conn, error) {
	pp := p.Clone()
	if pp == nil {
		pp = &Proxy{}
	}
	defer func() {
		if p != nil {
			p.DialAddr = addr
		}
	}()

	pp.timeout = Timeout
	return pp.Dial(network, addr, OutRouterIP)
}
func (p *Proxy) Dial(network, addr string, OutRouterIP *net.TCPAddr) (net.Conn, error) {
	var directDialer = direct{timeout: p.getTimeout(), OutRouterIP: OutRouterIP}
	addrHost, _, _ := net.SplitHostPort(addr)
	if p == nil {
		a, e := directDialer.Dial(network, addr)
		return a, e
	}
	p.DialAddr = addrHost

	if p.URL == nil {
		a, e := directDialer.Dial(network, addr)
		if a != nil {
			p.DialAddr = a.RemoteAddr().String()
		}
		return a, e
	}
	if p.Regexp != nil {
		if addrHost != "" && p.Regexp(addrHost) {
			a, e := directDialer.Dial(network, addr)
			if a != nil {
				p.DialAddr = a.RemoteAddr().String()
			}
			return a, e
		}
	}
	var e error
	var conn net.Conn
	if p.IsSocksType() {
		d, err1 := proxy.SOCKS5("tcp", p.Host, p.getSocksAuth(), directDialer)
		if err1 != nil {
			return nil, err1
		}
		conn, e = d.Dial(network, addr)
		if conn != nil {
			p.DialAddr = addr
		}
		return conn, e
	}
	p.DialAddr = p.Host
	conn, e = directDialer.Dial(network, p.DialAddr)
	if e != nil {
		return nil, e
	}
	us := ""
	if p.User() != "" {
		ns := base64.StdEncoding.EncodeToString([]byte(p.User() + ":" + p.Pass()))
		us = "Authorization: Basic " + ns + "\r\n"
		//部分HTTP代理 需要 Proxy-Authorization
		us += "Proxy-Authorization: Basic " + ns + "\r\n"
	}
	//部分HTTP代理 需要 Proxy-Connection
	us += "Proxy-Connection: Keep-Alive\r\n"
	_, e = conn.Write([]byte("CONNECT " + addr + " HTTP/1.1\r\nHost: " + addr + "\r\n" + us + "\r\n"))
	if e != nil {
		return nil, e
	}
	b := make([]byte, 128)
	n, er := conn.Read(b)
	if n < 13 {
		_ = conn.Close()
		return nil, er
	}
	s := string(b[:12])
	if s != "HTTP/1.1 200" && s != "HTTP/1.0 200" {
		return nil, fmt.Errorf(string(b))
	}
	b = make([]byte, 128)
	var ms error
	for {
		_ = conn.SetDeadline(time.Now().Add(100 * time.Millisecond))
		n, ms = conn.Read(b)
		if ms != nil {
			break
		}
	}
	_ = conn.SetDeadline(time.Time{})
	return conn, er
}
 
type direct struct {
	timeout     time.Duration
	OutRouterIP *net.TCPAddr
}

func (ps direct) Dial(network, addr string) (net.Conn, error) {
	return ps.DialContext(context.Background(), network, addr)
}

func (ps direct) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	var d net.Dialer                  // 底层 net.Dialer
	d.Timeout = ps.timeout            // 使用自定义超时
	if d.Timeout < time.Millisecond { // 防止超时时间太小
		d.Timeout = 5 * time.Second
	}

	// 本地回环地址直接走系统默认，不做网卡绑定
	if !strings.Contains(addr, "127.0.0.1") && !strings.Contains(addr, "[::1]") && ps.OutRouterIP != nil {
		// 根据 OutRouterIP 找到对应网卡的 v4/v6 地址
		if mip := RouterIPInspect(ps.OutRouterIP); mip != nil {
			host, _, err := net.SplitHostPort(addr) // 从 "host:port" 里拆出 host
			if err == nil {
				// 去掉 IPv6 字面量的方括号，例如 "[240e:...::1]"
				if len(host) > 2 && host[0] == '[' && host[len(host)-1] == ']' {
					host = host[1 : len(host)-1]
				}

				// 尝试把 host 当作 IP 字面量解析
				if ip := net.ParseIP(host); ip != nil {
					// ---- addr 是 IP 字面量：不依赖 network 判断 v4 / v6 ----

					// 根据 network 前缀判断是 TCP 还是 UDP，因为 LocalAddr 类型必须匹配
					isTCP := strings.HasPrefix(network, "tcp")
					isUDP := strings.HasPrefix(network, "udp")

					if ip4 := ip.To4(); ip4 != nil {
						// 目标是 IPv4
						if mip.IPv4 != nil {
							if isTCP {
								// TCP 使用 *net.TCPAddr
								d.LocalAddr = &net.TCPAddr{
									IP:   mip.IPv4,
									Port: 0,
								}
							} else if isUDP {
								// UDP 使用 *net.UDPAddr
								d.LocalAddr = &net.UDPAddr{
									IP:   mip.IPv4,
									Port: 0,
								}
							}
						}
					} else {
						// 目标是 IPv6
						if mip.IPv6 != nil {
							if isTCP {
								d.LocalAddr = &net.TCPAddr{
									IP:   mip.IPv6.IP,
									Port: 0,
									Zone: mip.IPv6.Zone,
								}
							} else if isUDP {
								d.LocalAddr = &net.UDPAddr{
									IP:   mip.IPv6.IP,
									Port: 0,
									Zone: mip.IPv6.Zone,
								}
							}
						}
					}
				} else {
					// ---- addr 是域名：按 network 里是否带 4/6 来决定优先绑 v4 或 v6 ----
					isTCP := strings.HasPrefix(network, "tcp")
					isUDP := strings.HasPrefix(network, "udp")

					// 明确要求 v4 的情况：tcp4 / udp4
					if strings.Contains(network, "4") {
						if mip.IPv4 != nil {
							if isTCP {
								d.LocalAddr = &net.TCPAddr{
									IP:   mip.IPv4,
									Port: 0,
								}
							} else if isUDP {
								d.LocalAddr = &net.UDPAddr{
									IP:   mip.IPv4,
									Port: 0,
								}
							}
						}
					} else if strings.Contains(network, "6") {
						// 明确要求 v6 的情况：tcp6 / udp6
						if mip.IPv6 != nil {
							if isTCP {
								d.LocalAddr = &net.TCPAddr{
									IP:   mip.IPv6.IP,
									Port: 0,
									Zone: mip.IPv6.Zone,
								}
							} else if isUDP {
								d.LocalAddr = &net.UDPAddr{
									IP:   mip.IPv6.IP,
									Port: 0,
									Zone: mip.IPv6.Zone,
								}
							}
						}
					} else {
						// network 既没写 4 也没写 6（例如 "tcp" / "udp"）
						// 这种情况下系统会自己选 v4/v6，我们只是尽量给它一个匹配的 LocalAddr
						if mip.IPv4 != nil {
							if isTCP {
								d.LocalAddr = &net.TCPAddr{
									IP:   mip.IPv4,
									Port: 0,
								}
							} else if isUDP {
								d.LocalAddr = &net.UDPAddr{
									IP:   mip.IPv4,
									Port: 0,
								}
							}
						} else if mip.IPv6 != nil {
							if isTCP {
								d.LocalAddr = &net.TCPAddr{
									IP:   mip.IPv6.IP,
									Port: 0,
									Zone: mip.IPv6.Zone,
								}
							} else if isUDP {
								d.LocalAddr = &net.UDPAddr{
									IP:   mip.IPv6.IP,
									Port: 0,
									Zone: mip.IPv6.Zone,
								}
							}
						}
					}
				}
			}
		}
	}

	// 给 DialContext 再包一层超时上下文
	ctx, cancel := context.WithTimeout(ctx, d.Timeout)
	defer cancel()

	return d.DialContext(ctx, network, addr)
}

func FormatIP(ip net.IP, port string) string {
	if ip.To4() != nil {
		return fmt.Sprintf("%s:%s", ip.String(), port)
	}
	return fmt.Sprintf("[%s]:%s", ip.String(), port)
}

// RouterIPs 保存某个出口网卡的 IPv4 / IPv6 信息  // 定义一个结构体保存 IPv4、IPv6 和 Zone
type RouterIPs struct {
	IPv4 net.IP      // 网卡上的 IPv4 地址（可能为 nil）
	IPv6 *net.IPAddr // 网卡上的 IPv6 地址（可能为 nil，Zone 会填接口名）
}

// RouterIPInspect 根据 addr.IP 找到对应网卡，并返回该网卡的 IPv4 和 IPv6 信息  // 通过目标 IP 反查对应的网卡和它的 v4/v6
func RouterIPInspect(addr *net.TCPAddr) *RouterIPs { // 参数类型保持不变，只改返回类型
	if addr == nil {                                 // 保护一下空指针
		return nil
	}

	interfaces, err := net.Interfaces() // 获取所有网卡
	if err != nil {                     // 如果获取失败直接返回
		return nil
	}

	for _, face := range interfaces { // 遍历每个网卡
		addrs, err1 := face.Addrs() // 取网卡上的所有地址
		if err1 != nil {            // 某个网卡取地址失败就跳过
			continue
		}

		// 先判断这个网卡的任何一个地址是否包含目标 addr.IP  // 判断这个网卡是不是我们要找的那张
		match := false            // 标记是否匹配
		for _, a := range addrs { // 遍历网卡地址
			if aspnet, ok := a.(*net.IPNet); ok { // 只处理 IPNet
				if aspnet.Contains(addr.IP) { // 目标 IP 在这个网段里
					match = true // 标记匹配
					break        // 找到了就可以退出这层循环
				}
			}
		}
		if !match { // 如果这个网卡不匹配目标 IP，跳过
			continue
		}

		// 走到这里说明 face 就是承载 addr.IP 的网卡  // 现在从这个网卡上收集它的 v4 和 v6
		var ip4 net.IP      // 保存网卡上的 IPv4
		var ip6 *net.IPAddr // 保存网卡上的 IPv6（带 Zone）

		for _, a := range addrs { // 再遍历这个网卡的地址
			aspnet, ok := a.(*net.IPNet) // 仍然只关心 IPNet
			if !ok {
				continue
			}

			ip := aspnet.IP // 实际 IP（可能是 v4 也可能是 v6）

			if ip4 == nil && ip.To4() != nil { // 第一次碰到 IPv4 时记录下来
				ip4 = ip
			}

			if ip6 == nil && ip.To4() == nil { // 第一次碰到 IPv6（真 v6，不是 v4 映射）时记录下来
				ip6 = &net.IPAddr{ // 用 IPAddr 是为了能带 Zone
					IP:   ip,        // IPv6 地址
					Zone: face.Name, // Zone 填网卡名，比如 "Ethernet"、"Wi-Fi"
				}
			}
		}

		// 如果这个网卡有任何一个地址，就返回  // 至少有一个 v4 或 v6 就算找到
		if ip4 != nil || ip6 != nil {
			return &RouterIPs{
				IPv4: ip4,
				IPv6: ip6,
			}
		}
	}
	return nil // 都没找到就返回 nil
}
