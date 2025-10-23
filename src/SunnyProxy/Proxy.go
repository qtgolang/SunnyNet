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
	fmt.Println("CONNECT " + addr + " HTTP/1.1\r\nHost: " + addr + "\r\n" + us + "\r\n")
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
	fmt.Println("->>>\r\n" + s)
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
	var m net.Dialer
	m.Timeout = ps.timeout
	if m.Timeout < time.Millisecond {
		m.Timeout = 5 * time.Second
	}
	if !strings.Contains(addr, "127.0.0.1") && !strings.Contains(addr, "[::1]") {
		mip := RouterIPInspect(ps.OutRouterIP)
		if mip != nil {
			m.LocalAddr = &net.TCPAddr{
				IP:   mip,
				Port: 0,
			}
		}
	}
	ctx, cancel := context.WithTimeout(ctx, m.Timeout)
	defer cancel()
	return m.DialContext(ctx, network, addr)
}
func FormatIP(ip net.IP, port string) string {
	if ip.To4() != nil {
		return fmt.Sprintf("%s:%s", ip.String(), port)
	}
	return fmt.Sprintf("[%s]:%s", ip.String(), port)
}
func RouterIPInspect(addr *net.TCPAddr) net.IP {
	if addr == nil {
		return nil
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, face := range interfaces {
		adders, err1 := face.Addrs()
		if err1 != nil {
			continue
		}
		for _, a := range adders {
			if aspnet, ok := a.(*net.IPNet); ok {
				if aspnet.Contains(addr.IP) {
					return addr.IP
				}
			}
		}
	}
	return nil
}
