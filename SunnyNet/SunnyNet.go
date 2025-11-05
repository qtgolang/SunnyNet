package SunnyNet

import (
	"bufio"
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/qtgolang/SunnyNet/src/Certificate"
	"github.com/qtgolang/SunnyNet/src/CrossCompiled"
	"github.com/qtgolang/SunnyNet/src/GoScriptCode"
	"github.com/qtgolang/SunnyNet/src/HttpCertificate"
	"github.com/qtgolang/SunnyNet/src/Interface"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/ProcessCheck"
	"github.com/qtgolang/SunnyNet/src/ReadWriteObject"
	"github.com/qtgolang/SunnyNet/src/Resource"
	"github.com/qtgolang/SunnyNet/src/SunnyProxy"
	"github.com/qtgolang/SunnyNet/src/crypto/tls"
	"github.com/qtgolang/SunnyNet/src/dns"
	"github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/httpClient"
	"github.com/qtgolang/SunnyNet/src/public"
	"github.com/qtgolang/SunnyNet/src/websocket"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func init() {
	//使用全部-1个CPU性能,例如你电脑CPU是4核心 那么就使用4-1 使用3核心的的CPU性能
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)
	CrossCompiled.SetNetworkConnectNumber()
}

// TargetInfo 请求连接信息
type TargetInfo struct {
	Host string //带端口号
	Port uint16
	IPV6 bool
}

func (s *TargetInfo) Clone() *TargetInfo {
	if s == nil {
		return nil
	}
	return &TargetInfo{
		Host: s.Host,
		Port: s.Port,
		IPV6: s.IPV6,
	}
}
func (s *TargetInfo) IsDomain() bool {
	if s == nil {
		return false
	}
	if s.IPV6 {
		return false
	}
	if ip := net.ParseIP(s.Host); ip != nil && (ip.To4() != nil || ip.To16() != nil) {
		return false
	}
	return true
}

// Remove 清除信息
func (s *TargetInfo) Remove() {
	s.Host = public.NULL
	s.Port = 0
}

// 解析IPV6地址
func parseIPv6Address(address string) (string, uint16, net.IP) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		// 没有端口号
		host = address
	}

	ip := net.ParseIP(host)
	if ip == nil {
		ipAddr1, err1 := net.ResolveIPAddr("ip", host)
		if err1 != nil {
			return "", 0, nil
		}
		ip = ipAddr1.IP
	}
	if ip == nil {
		return "", 0, nil
	} else if ip.To4() != nil {
		return "", 0, nil
	}
	var portNumber uint16
	if port != "" {
		portInt, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			return "", 0, nil
		}
		portNumber = uint16(portInt)
	}

	return ip.String(), portNumber, ip
}

// Parse 解析连接信息
func (s *TargetInfo) Parse(HostName string, Port interface{}, IPV6 ...bool) {
	//如果是8.8.8.8 则端口号不变
	//如果是8.8.8.8:8888 Host和端口都变
	//如果是Host="" Port=8888 Host不变,端口变
	//如果是8.8.8.8:8888 Port=8889 那么端口=8888
	if s == nil {
		return
	}
	Host := HostName
	p := uint16(0)
	s.IPV6 = len(IPV6) > 0
	if s.IPV6 {
		s.IPV6 = IPV6[0]
	}
	_s, _p, _ := parseIPv6Address(Host)
	if _s != "" {
		s.Host = _s
		p = _p
		s.IPV6 = true
	}
	if strings.Index(Host, ":") == -1 || s.IPV6 {
		switch v := Port.(type) {
		case string:
			a, _ := strconv.Atoi(v)
			if a > 0 {
				p = uint16(a)
			}
			break
		case uint16:
			p = v
			break
		default:
			a, _ := strconv.Atoi(fmt.Sprintf("%d", v))
			if a > 0 {
				p = uint16(a)
			}
			break
		}
	} else {
		arr := strings.Split(Host, ":")
		if len(arr) == 2 {
			Host = arr[0]
			a, _ := strconv.Atoi(arr[1])
			if a > 0 {
				p = uint16(a)
			}
		}
	}
	if p > 0 {
		s.Port = p
	}
	if Host != "" {
		if _s == "" {
			s.Host = Host
		}
	}
	if strings.ToLower(s.Host) == "localhost" {
		s.Host = "127.0.0.1"
	}
}

// String 格式化信息返回格式127.0.0.1:8888
func (s *TargetInfo) String() string {
	if s.IPV6 {
		return fmt.Sprintf("[%s]:%d", s.Host, s.Port)
	}
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// 请求信息
type proxyRequest struct {
	Conn                  net.Conn                         //请求的原始TCP连接
	RwObj                 *ReadWriteObject.ReadWriteObject //读写对象
	Theology              int                              //中间件回调唯一ID
	Target                *TargetInfo                      //目标连接信息
	ProxyHost             string                           //请求之上的代理
	Pid                   string                           //s5连接过来的pid
	Global                *Sunny                           //继承全局中间件信息
	Request               *http.Request                    //要发送的请求体
	Response              response                         //HTTP响应体
	TCP                   public.TCP                       //TCP收发数据
	Proxy                 *SunnyProxy.Proxy                //设置指定代理
	HttpCall              int                              //http 请求回调地址
	TcpCall               int                              //TCP请求回调地址
	wsCall                int                              //ws回调地址
	HttpGoCall            func(ConnHTTP)                   //http 请求回调地址
	TcpGoCall             func(ConnTCP)                    //TCP请求回调地址
	wsGoCall              func(ConnWebSocket)              //ws回调地址
	NoRepairHttp          bool                             //不要纠正Http
	Lock                  sync.Mutex
	defaultScheme         string
	SendTimeout           time.Duration
	TlsConfig             *tls.Config
	_Display              bool //是否允许显示到列表，也就是是否调用Call
	_isRandomCipherSuites bool
	_SocksUser            string
	outRouterIP           *net.TCPAddr
	rawTarget             uint32
	_note                 string //注释上下文
}

var sUser = make(map[int]string)
var sL sync.Mutex

// 设置s5连接账号
func (s *proxyRequest) setSocket5User(user string) {
	sL.Lock()
	sUser[s.Theology] = user
	sL.Unlock()
}
func (s *proxyRequest) GetNote() string {
	return s._note
}

// 更新唯一ID以及s5连接账号
func (s *proxyRequest) updateSocket5User() {
	sL.Lock()
	user := sUser[s.Theology]
	delete(sUser, s.Theology)
	s.Theology = int(atomic.AddInt64(&public.Theology, 1))
	if user != "" {
		sUser[s.Theology] = user
		s._SocksUser = user
	}
	sL.Unlock()
}

// 清除唯一ID对应的s5连接账号
func (s *proxyRequest) delSocket5User() {
	sL.Lock()
	delete(sUser, s.Theology)
	sL.Unlock()
}

// GetSocket5User 获取唯一ID对应的s5连接账号
func GetSocket5User(TheologyId int) string {
	sL.Lock()
	user := sUser[TheologyId]
	sL.Unlock()
	return user
}

// AuthMethod S5代理鉴权
func (s *proxyRequest) AuthMethod() (bool, string) {
	av, err := s.RwObj.ReadByte()
	if err != nil || av != 1 {
		//fmt.Println(ID, "Socks5 auth version invalid")
		return false, public.NULL
	}

	uLen, err := s.RwObj.ReadByte()
	if err != nil || uLen <= 0 || uLen > 255 {
		//fmt.Println(ID, "Socks5 auth user length invalid")
		return false, public.NULL
	}

	uBuf := make([]byte, uLen)
	nr, err := s.RwObj.Read(uBuf)
	if err != nil || nr != int(uLen) {
		//fmt.Println(ID, "Socks5 auth user error", nr)
		return false, public.NULL
	}

	user := string(uBuf)

	pLen, err := s.RwObj.ReadByte()
	if err != nil || pLen <= 0 || pLen > 255 {
		//fmt.Println(ID, "Socks5 auth passwd length invalid", pLen)
		return false, public.NULL
	}

	pBuf := make([]byte, pLen)
	nr, err = s.RwObj.Read(pBuf)
	if err != nil || nr != int(pLen) {
		//fmt.Println(ID, "Socks5 auth passwd error", pLen, nr)
		return false, public.NULL
	}

	passwd := string(pBuf)
	if s.Global.socket5VerifyUser {
		if len(user) > 0 && len(passwd) > 0 {
			s.Global.socket5VerifyUserLock.Lock()
			if passwd == s.Global.socket5VerifyUserList[user] {
				s.Global.socket5VerifyUserLock.Unlock()
				_ = s.RwObj.WriteByte(0x01)
				_ = s.RwObj.WriteByte(0x00)
				s.setSocket5User(user)
				return true, passwd
			}
			s.Global.socket5VerifyUserLock.Unlock()
		}
	} else {
		if len(user) > -1 || len(passwd) > -1 {
			//fmt.Println(1, user, passwd)
			_ = s.RwObj.WriteByte(0x01)
			_ = s.RwObj.WriteByte(0x00)
			return true, passwd
		}
	}

	_ = s.RwObj.WriteByte(0x01)
	_ = s.RwObj.WriteByte(0x01)
	return false, public.NULL
}

// Socks5ProxyVerification S5代理验证
func (s *proxyRequest) Socks5ProxyVerification() bool {
	version, err := s.RwObj.ReadByte()
	if err != nil {
		return false
	}
	if version != public.Socks5Version {
		return false
	}
	methods, err := s.RwObj.ReadByte()
	if err != nil {
		return false
	}

	if methods < 0 || methods > 255 {
		return false
	}
	supportAuth := false
	method := public.Socks5AuthNone
	for i := 0; i < int(methods); i++ {
		method, err = s.RwObj.ReadByte()
		if err != nil {
			return false
		}
		if method == public.Socks5Auth {
			supportAuth = true
		}
	}

	err = s.RwObj.WriteByte(version)
	if err != nil {
		return false
	}

	// 支持加密, 则回复加密方法.
	if supportAuth {
		method = public.Socks5Auth
		err = s.RwObj.WriteByte(method)
		if err != nil {
			return false
		}
	} else {
		// 服务器不支持加密, 直接通过.
		method = public.Socks5AuthNone
		err = s.RwObj.WriteByte(method)
		if err != nil {
			return false
		}
	}
	_ = s.RwObj.Flush()
	ok := false
	// Auth mode, read user passwd.
	// 暂时没啥用 现在设置的不要密码或任意账号密码都通过
	if supportAuth {
		ok, _ = s.AuthMethod()
		if !ok {
			return false
		}
		_ = s.RwObj.Flush()
	} else if s.Global.socket5VerifyUser {
		return false
	}

	handshakeVersion, err := s.RwObj.ReadByte()
	if err != nil || handshakeVersion != public.Socks5Version {
		if err != nil {
		}
		return false
	}
	command, err := s.RwObj.ReadByte()
	if err != nil {
		//fmt.Println(ID, "Socks5 read command error", err.Error())
		return false
	}
	if command != public.Socks5CmdConnect &&
		command != public.Socks5CmdBind &&
		command != public.Socks5CmdUDP {
		return false
	}

	_, _ = s.RwObj.ReadByte() // rsv byte
	aTyp, err := s.RwObj.ReadByte()
	if err != nil {
		return false
	}
	if aTyp != public.Socks5typeDomainName &&
		aTyp != public.Socks5typeIpv4 &&
		aTyp != public.Socks5typeIpv6 {
		return false
	}

	hostname := public.NULL
	isV6 := false
	switch {
	case aTyp == public.Socks5typeIpv4:
		{
			IPv4Buf := make([]byte, 4)
			nr, err := s.RwObj.Read(IPv4Buf)
			if err != nil || nr != 4 {
				return false
			}

			ip := net.IP(IPv4Buf)
			hostname = ip.String()
		}
	case aTyp == public.Socks5typeIpv6:
		{
			IPv6Buf := make([]byte, 16)
			nr, err := s.RwObj.Read(IPv6Buf)
			if err != nil || nr != 16 {
				return false
			}

			ip := net.IP(IPv6Buf)
			hostname = ip.String()
			isV6 = true
		}
	case aTyp == public.Socks5typeDomainName:
		{
			dnLen, err := s.RwObj.ReadByte()
			if err != nil || int(dnLen) < 0 {
				return false
			}

			domain := make([]byte, dnLen)
			nr, err := s.RwObj.Read(domain)
			if err != nil || nr != int(dnLen) {
				return false
			}
			hostname = string(domain)
		}
	}
	portNum1, err := s.RwObj.ReadByte()
	if err != nil {
		return false
	}
	portNum2, err := s.RwObj.ReadByte()
	if err != nil {
		return false
	}
	port := uint16(portNum1)<<8 + uint16(portNum2)
	if isV6 {
		hostname = fmt.Sprintf("[%s]", hostname)
	}
	_ = s.RwObj.WriteByte(public.Socks5Version)

	if command == public.Socks5CmdUDP {
		ipArr := strings.Split(s.Conn.LocalAddr().String(), ":")
		_ = s.RwObj.WriteByte(0) // SOCKS5_SUCCEEDED
		_ = s.RwObj.WriteByte(0)
		if len(ipArr) != 2 {
			_ = s.RwObj.WriteByte(public.Socks5typeIpv4)
			_, _ = s.RwObj.Write(net.ParseIP("0.0.0.0").To4())
			_ = s.RwObj.WriteByte(portNum1)
			_ = s.RwObj.WriteByte(portNum2)
		} else {
			host := ipArr[0]
			if public.IsIPv4(host) {
				_ = s.RwObj.WriteByte(public.Socks5typeIpv4)
				_, _ = s.RwObj.Write(net.ParseIP(host).To4())
			} else if public.IsIPv6(host) {
				_ = s.RwObj.WriteByte(public.Socks5typeIpv6)
				_, _ = s.RwObj.Write(net.ParseIP(host).To16())
			} else {
				_ = s.RwObj.WriteByte(public.Socks5typeDomainName)
				_ = s.RwObj.WriteByte(byte(len(hostname)))
				_, _ = s.RwObj.WriteString(hostname)
			}
			portNum, _ := strconv.Atoi(ipArr[1])
			portNum1 = byte(portNum >> 8)
			portNum2 = byte(portNum)
			_ = s.RwObj.WriteByte(portNum1)
			_ = s.RwObj.WriteByte(portNum2)
		}
		_ = s.RwObj.Flush()
		for {
			b := make([]byte, 10)
			_, e := s.RwObj.Read(b)
			if e != nil {
				break
			}
		}
		return false
	}
	//var RemoteTCP net.Conn
	err = nil
	a := strings.Split(s.Conn.RemoteAddr().String(), ":")
	if len(a) >= 2 {
		hostname = strings.ReplaceAll(hostname, "127.0.0.1", a[0])
	}
	if err != nil {
		//fmt.Println(hostname, err)
		_ = s.RwObj.WriteByte(1) // SOCKS5_GENERAL_SOCKS_SERVER_FAILURE
	} else {
		_ = s.RwObj.WriteByte(0) // SOCKS5_SUCCEEDED
	}
	_ = s.RwObj.WriteByte(0)

	_ = s.RwObj.WriteByte(public.Socks5typeDomainName)
	_ = s.RwObj.WriteByte(byte(len(hostname)))
	_, _ = s.RwObj.WriteString(hostname)

	_ = s.RwObj.WriteByte(portNum1)
	_ = s.RwObj.WriteByte(portNum2)
	_ = s.RwObj.Flush()
	if err != nil {
		return false
	}
	s.Target.Parse(hostname, port)
	return true
}

var loopLock sync.Mutex
var linkMap = make(map[string]string)

func linkAdd(o, n string) {
	loopLock.Lock()
	defer loopLock.Unlock()
	linkMap[n] = o
}
func linkDel(n string) {
	loopLock.Lock()
	defer loopLock.Unlock()
	delete(linkMap, n)
}
func linkQuery(n string) string {
	loopLock.Lock()
	defer loopLock.Unlock()
	return linkMap[n]
}

// 封装连接逻辑
func dialTCP(proxyTools *SunnyProxy.Proxy, remoteAddr string, outRouterIP *net.TCPAddr) (net.Conn, error) {
	return proxyTools.DialWithTimeout("tcp", remoteAddr, 2*time.Second, outRouterIP)
}
func connectToTarget(s *proxyRequest, proxyTools *SunnyProxy.Proxy, outRouterIP *net.TCPAddr) (net.Conn, string) {
	if dns.IsRemoteDnsServer() {
		conn, _ := proxyTools.Dial("tcp", s.Target.String(), outRouterIP)
		return conn, s.Target.String()
	}
	ip := net.ParseIP(s.Target.Host)
	if ip != nil {
		remoteAddr := SunnyProxy.FormatIP(ip, fmt.Sprintf("%d", s.Target.Port))
		conn, _ := proxyTools.Dial("tcp", remoteAddr, outRouterIP)
		return conn, remoteAddr
	}

	var ProxyHost string
	var dial func(network string, addr string, outRouterIP *net.TCPAddr) (net.Conn, error)
	if proxyTools != nil {
		ProxyHost = proxyTools.Host
		dial = proxyTools.Dial
	}
	ip = dns.GetFirstIP(s.Target.Host, ProxyHost)
	if ip != nil {
		remoteAddr := SunnyProxy.FormatIP(ip, fmt.Sprintf("%d", s.Target.Port))
		conn, _ := dialTCP(proxyTools, remoteAddr, outRouterIP)
		if conn != nil {
			return conn, remoteAddr
		}
	}

	ips, _ := dns.LookupIP(s.Target.Host, ProxyHost, outRouterIP, dial)

	//优先尝试IPV4
	for _, ip2 := range ips {
		if ip4 := ip2.To4(); ip4 != nil {
			remoteAddr := SunnyProxy.FormatIP(ip2, fmt.Sprintf("%d", s.Target.Port))
			conn, _ := dialTCP(proxyTools, remoteAddr, outRouterIP)
			if conn != nil {
				dns.SetFirstIP(s.Target.Host, ProxyHost, ip2)
				return conn, remoteAddr
			}
		}
	}

	//最后尝试IPV6
	for _, ip2 := range ips {
		if ip6 := ip2.To16(); ip6 != nil {
			remoteAddr := SunnyProxy.FormatIP(ip2, fmt.Sprintf("%d", s.Target.Port))
			conn, _ := dialTCP(proxyTools, remoteAddr, outRouterIP)
			if conn != nil {
				dns.SetFirstIP(s.Target.Host, ProxyHost, ip2)
				return conn, remoteAddr
			}
		}
	}
	return nil, ""
}

// MustTcpProcessing 强制走TCP处理过程
// aheadData 提取获取的数据
func (s *proxyRequest) MustTcpProcessing(Tag string) {
	if s.Target == nil {
		return
	}
	if s.isLoop() {
		return
	}
	var err error
	var isClose = false
	as := &public.TcpMsg{}
	as.Data.WriteString(Tag)
	s.CallbackTCPRequest(public.SunnyNetMsgTypeTCPAboutToConnect, as, s.Target.String())
	if Tag != as.Data.String() {
		s.Target.Parse(as.Data.String(), 0)
	}
	var proxyTools *SunnyProxy.Proxy
	if as.Proxy != nil {
		proxyTools = as.Proxy
	} else if s.Global.proxy != nil {
		if !s.Global.proxyRules(s.Target.Host) {
			proxyTools = s.Global.proxy.Clone()
			if proxyTools != nil {
				proxyTools.Regexp = s.Global.proxyRules
			}
		}
	}
	RemoteTCP, RemoteAddr := connectToTarget(s, proxyTools, s.outRouterIP)
	if RemoteAddr != s.Target.String() {
		RemoteAddr = s.Target.String() + " -> " + RemoteAddr
	}
	defer func() {
		if !isClose {
			if RemoteTCP != nil {
				s.CallbackTCPRequest(public.SunnyNetMsgTypeTCPClose, nil, RemoteAddr)
			} else {
				s.CallbackTCPRequest(public.SunnyNetMsgTypeTCPClose, nil, s.Target.String())
			}
		}
		if as != nil {
			as.Data.Reset()
		}
		as = nil
		proxyTools = nil
		s.releaseTcp()
		if RemoteTCP != nil {
			_ = RemoteTCP.Close()
			linkDel(RemoteTCP.LocalAddr().String())
		}
	}()

	if RemoteTCP != nil {
		linkAdd(s.Conn.RemoteAddr().String(), RemoteTCP.LocalAddr().String())
	}
	if RemoteTCP != nil && Tag == public.TagTcpSSLAgreement {
		tlsConn := tls.Client(RemoteTCP, s.TlsConfig)
		err = tlsConn.Handshake()
		RemoteTCP = tlsConn
	}
	if err == nil && RemoteTCP != nil {
		tw := ReadWriteObject.NewReadWriteObject(RemoteTCP)
		{
			//构造结构体数据,主动发送，关闭等操作时需要用
			if s.TCP.Send == nil {
				s.TCP.Send = &public.TcpMsg{}
			}
			if s.TCP.Receive == nil {
				s.TCP.Receive = &public.TcpMsg{}
			}
			s.TCP.SendBw = s.RwObj.Writer
			s.TCP.ReceiveBw = tw.Writer
			s.TCP.ConnSend = s.Conn
			s.TCP.ConnServer = RemoteTCP
			TcpSceneLock.Lock()
			TcpStorage[s.Theology] = &s.TCP
			TcpSceneLock.Unlock()
		}
		as.Data.Reset()
		as.Data.Write([]byte(RemoteTCP.LocalAddr().String()))
		s.CallbackTCPRequest(public.SunnyNetMsgTypeTCPConnectOK, as, RemoteAddr)
		as.Data.Reset()
		isClose = s.TcpCallback(RemoteTCP, Tag, tw, RemoteAddr)
	} else {
		_ = s.Conn.Close()
	}
	return
}

// 释放tcp关联的数据
func (s *proxyRequest) releaseTcp() {
	//================================================================================================================================
	if s == nil {
		return
	}
	if s.TCP.Send != nil {
		s.TCP.Send.Data.Reset()
	}
	if s.TCP.Receive != nil {
		s.TCP.Receive.Data.Reset()
	}
	s.TCP.Send = nil
	s.TCP.SendBw = nil
	s.TCP.ConnSend = nil //=========================  释放相关数据
	s.TCP.Receive = nil
	s.TCP.ReceiveBw = nil
	s.TCP.ConnServer = nil
	TcpSceneLock.Lock()
	TcpStorage[s.Theology] = nil
	delete(TcpStorage, s.Theology)
	TcpSceneLock.Unlock()
	//================================================================================================================================
}

// TcpCallback TCP消息处理 返回 是否已经调用 通知 回调函数 TCP已经关闭
func (s *proxyRequest) TcpCallback(RemoteTCP net.Conn, Tag string, tw *ReadWriteObject.ReadWriteObject, RemoteAddr string) bool {
	if RemoteTCP == nil {
		return false
	}
	var wg sync.WaitGroup
	wg.Add(1)
	isHttpReq := false //是否纠正HTTP请求，可能由于某些原因 客户端发送数据不及时判断为了TCP请求，后续TCP处理时纠正为HTTP请求
	//读取客户端消息转发给服务端
	go func() {
		s.SocketForward(*tw.Writer, s.RwObj, public.SunnyNetMsgTypeTCPClientSend, s.Conn, RemoteTCP, &s.TCP, &isHttpReq, RemoteAddr)
		wg.Done()
	}()
	//读取服务器消息转发给客户端
	s.SocketForward(*s.RwObj.Writer, tw, public.SunnyNetMsgTypeTCPClientReceive, RemoteTCP, s.Conn, &s.TCP, &isHttpReq, RemoteAddr)
	wg.Wait()
	s.releaseTcp()
	if isHttpReq {
		//可能由于某些原因 客户端发送数据不及时判断为了TCP请求,此时纠正为HTTP请求
		s.CallbackTCPRequest(public.SunnyNetMsgTypeTCPClose, nil, RemoteAddr)
		s.updateSocket5User()
		//如果之前是HTTP请求识别错误 这里转由HTTP请求处理函数继续处理
		if Tag == public.TagTcpSSLAgreement {
			s.httpProcessing(nil, Tag)
		} else {
			s.httpProcessing(nil, Tag)
		}
		return true
	}
	return false
}

// transparentProcessing 透明代理请求 处理过程
func (s *proxyRequest) transparentProcessing() {
	//将数据全部取出，稍后重新放进去
	_bytes, _ := s.RwObj.Peek(s.RwObj.Reader.Buffered())
	//升级到TLS客户端
	fig := &tls.Config{InsecureSkipVerify: true}
	T := tls.Client(s.Conn, fig)
	//将数据重新写进去
	T.Reset(_bytes)
	//进行握手处理
	msg, serverName, e := T.ClientHello()
	if e == nil {
		//从握手信息中取出要连接的服务器域名
		if serverName == public.NULL {
			//如果没有取出 则按照连接地址处理
			serverName = s.Conn.LocalAddr().String()
		}
		//将地址写到请求中间件连接信息中
		s.Target.Parse(serverName, public.HttpsDefaultPort)
		var certificate *tls.Certificate
		var er error
		if s.isLoop() {
			certificate, _, er = WhoisLoopCache(s.Global, nil, s.Target.String(), s.Global.rootCa, s.Global.rootKey)
		} else {
			certificate, _, er = WhoisCache(s.Global, nil, "null", s.Target.String(), s.Global.rootCa, s.Global.rootKey)
		}
		//进行生成证书，用于服务器返回握手信息
		if er != nil {
			_ = T.Close()
			return
		}
		//将证书和域名信息设置到TLS客户端中
		cfg := &tls.Config{Certificates: []tls.Certificate{*certificate}, ServerName: HttpCertificate.ParsingHost(s.Target.String()), InsecureSkipVerify: true}
		T.SetServer(cfg)
		//进行与客户端握手
		e = T.ServerHandshake(msg)
		if e == nil {
			//如果握手过程中没有发生意外， 则重写客户端会话
			s.Conn = T                                      //将TLS会话替换原始会话
			s.RwObj = ReadWriteObject.NewReadWriteObject(T) //重新包装读写对象
			//开始按照HTTP请求流程处理
			s.TlsConfig = cfg
			s.httpProcessing(nil, public.TagTcpSSLAgreement)
		}
	} else {
		//如果握手失败 直接返回，不做任何处理
	}
}

// 请求是否环路
func (s *proxyRequest) isLoop() bool {
	_, port, _ := public.SplitHostPort(s.RwObj.RemoteAddr().String())
	ok := CrossCompiled.IsLoopRequest(port, s.Global.port)
	if ok {
		link := linkQuery(s.Conn.RemoteAddr().String())
		if link != "" {
			p := CrossCompiled.LoopRemotePort(link)
			if p < 1 {
				return false
			}
		}
	}
	return ok
}
func (s *proxyRequest) targetIsInterfaceAdders() bool {
	if int(s.Target.Port) != s.Global.port {
		return false
	}
	if s.Target.Host == public.CertDownloadHost1 {
		return true
	}
	if s.Target.Host == public.CertDownloadHost2 {
		return true
	}
	adders, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}
	for _, addr := range adders {
		ipNet, _ := addr.(*net.IPNet)
		if ipNet == nil {
			continue
		}
		a := ipNet.IP.String()
		if a == s.Target.Host || a == "["+s.Target.Host+"]" || s.Target.Host == "["+a+"]" {
			return true
		}
	}
	return false
}

// httpProcessing http请求处理过程
func (s *proxyRequest) httpProcessing(aheadData []byte, Tag string) {
	var hh []byte
	var h2 []byte
	if len(aheadData) < 11 {
		h2, _ = s.RwObj.Peek(11 - len(aheadData))
	}
	hh = []byte(string(aheadData) + string(h2))

	if string(hh) == "PRI * HTTP/" {
		s.defaultScheme = "https"
		s.h2Request(aheadData)
		return
	}
	Method := public.GetMethod(hh)
	if public.IsHttpMethod(Method) {
		var buff bytes.Buffer
		buff.Write(aheadData)
		var isRules bool
		var host string
		var lineNumber int
		for {
			lineNumber++
			//找到HOST 进行匹配是否强制走 TCP
			bs, e := s.RwObj.ReadSlice('\n')
			ms := string(bs)
			buff.Write(bs)
			if lineNumber == 1 {
				isRules = !strings.Contains(strings.ToLower(buff.String()), "http/")
				if isRules {
					_ = s.RwObj.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
					Method = public.HttpMethodGET //防止 不是标准的 CONNECT 请求 ,防止在循环退出后出错
				}
			}
			if e != nil || len(bs) < 3 {
				break
			}
			arr := strings.SplitN(ms, ":", 2)
			if len(arr) > 1 && strings.ToLower(strings.TrimSpace(arr[0])) == "host" {
				host = strings.TrimSpace(arr[1])
				if !isRules {
					isRules = s.Global.tcpRules(host, s.Target.Host)
				}
				break
			}
		}
		if isRules && Method != public.HttpMethodCONNECT {
			// 每次最多读取 64 字节
			tmpBuf := make([]byte, 64)
			for {
				// 检查是否还有缓冲数据
				if s.RwObj.Buffered() == 0 {
					break
				}
				// 设置 10 毫秒超时
				_ = s.RwObj.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
				// 读取数据
				n, err := s.RwObj.Read(tmpBuf)
				if err != nil || n == 0 {
					break
				}
				// 写入缓冲区
				buff.Write(tmpBuf[:n])
			}
			_ = s.RwObj.SetReadDeadline(time.Time{})
			if s.Global.disableTCP {
				return
			}
			if s.Target.Host == "" {
				if host == "" {
					return
				}
				s.Target.Parse(host, 0)
			}
			if s.Target.Port == 0 {
				if Tag == public.TagTcpSSLAgreement {
					s.Target.Parse("", 443)
				} else {
					s.Target.Parse("", 80)
				}
			}
			if !(s.targetIsInterfaceAdders()) {
				s.NoRepairHttp = true
				s.RwObj = ReadWriteObject.NewReadWriteObject(newObjHook(s.RwObj, buff.Bytes()))
				s.MustTcpProcessing(Tag)
				return
			}
		}
		if Tag == public.TagTcpSSLAgreement {
			s.defaultScheme = "https"
		} else {
			s.defaultScheme = "http"
		}
		s.h1Request(buff.Bytes())
		return
	}
	//s.NoRepairHttp = true
	if len(aheadData) > 0 {
		s.RwObj = ReadWriteObject.NewReadWriteObject(newObjHook(s.RwObj, aheadData))
	}
	s.MustTcpProcessing(Tag)
	s.NoRepairHttp = false
	return
}

func (s *proxyRequest) isCerDownloadPage(request *http.Request) bool {
	i := int(s.Target.Port)
	if i == s.Global.port && (s.targetIsInterfaceAdders() || s.isLoop()) {
		if request != nil {
			if request.Header != nil {
				if request.Header.Get(public.HTTPClientTags) == "true" {
					request.Header.Del(public.HTTPClientTags)
					return false
				}
			}
			if request.URL != nil {
				defer func() { _ = s.Conn.Close() }()
				if request.URL.Path == "/favicon.ico" {
					_, _ = s.RwObj.Write(public.LocalBuildBody("image/x-icon", Resource.Icon))
					return true
				}

				if request.URL.Path == "/" || request.URL.Path == "/ssl" || request.URL.Path == public.NULL {
					_, _ = s.RwObj.Write(public.LocalBuildBody("text/html", `<html><head><meta http-equiv="Content-Type" content="text/html; charset=UTF-8"><title>证书安装</title></head><body style="font-family: arial,sans-serif;"><h1>[SunnyNet网络中间件] 证书安装</h1><br /><ul><li>您可以下载 <a href="SunnyRoot.cer">SunnyRoot 证书</a></ul><ul><li>您也可以 <a href="install.html">查看证书安装教程</a></ul></body></html>`))
					return true
				}
				if request.URL.Path == "/SunnyRoot.cer" || request.URL.Path == "SunnyRoot.cer" {
					_, _ = s.RwObj.Write(public.LocalBuildBody("application/x-x509-ca-cert", s.Global.ExportCert()))
					return true
				}
				if request.URL.Path == "/install.html" || request.URL.Path == "install.html" {
					bs := bytes.ReplaceAll(Resource.FrontendIndex, []byte(`/assets/index`), []byte(`install/assets/index`))
					_, _ = s.RwObj.Write(public.LocalBuildBody("text/html", bs))
					return true
				}
				if strings.HasPrefix(request.URL.Path, "/install/assets/") || strings.HasPrefix(request.URL.Path, "install/assets/") {
					data, err := Resource.ReadVueFile(strings.ReplaceAll(request.URL.Path, "/install/", ""))
					if err == nil {
						_FileType := strings.ToLower(request.URL.Path)
						_, _ = s.RwObj.WriteString("HTTP/1.1 200 OK\r\nCache-Control: no-cache, must-revalidate\r\nPragma: no-cache\r\nExpires: 0\r\nContent-Length: ")
						if strings.HasSuffix(_FileType, ".css") {
							mData := bytes.ReplaceAll(data, []byte("url(/assets/codicon"), []byte(strings.ReplaceAll("url("+"install/assets/codicon", "//", "/")))
							data = mData
							_, _ = s.RwObj.WriteString(fmt.Sprintf("%d\r\nContent-Type:  text/css\r\n\r\n", len(data)))
						}
						if strings.HasSuffix(_FileType, ".js") {
							_, _ = s.RwObj.WriteString(fmt.Sprintf("%d\r\nContent-Type:  application/x-javascript\r\n\r\n", len(data)))
						}
						if strings.HasSuffix(_FileType, ".wasm") {
							_, _ = s.RwObj.WriteString(fmt.Sprintf("%d\r\nContent-Type:  application/wasm\r\n\r\n", len(data)))
						}
						if strings.HasSuffix(_FileType, ".ttf") {
							_, _ = s.RwObj.WriteString(fmt.Sprintf("%d\r\nContent-Type:  application/application/x-font-ttf\r\n\r\n", len(data)))
						}
						_, _ = s.RwObj.Write(data)
						return true
					}
				}
				if !s.isUserScriptCodeEditRequest(request) {
					_, _ = s.RwObj.Write(public.LocalBuildBody("text/html", "404 Not Found"))
				}
				return true
			}
		}
	}
	return false
}
func (s *proxyRequest) Error(error error, _Display bool) {
	s._Display = _Display
	s.CallbackError(public.ProcessError(error))
	if errors.Is(error, public.ProvideForwardingServiceOnly) {
		return
	}
	if s.Response.Response != nil {
		if s.Response.Body != nil {
			_ = s.Response.Body.Close()
		}
		s.Response.Response = nil
	}
	if s.Response.rw == nil {
		s.Response.rw = &errorRW{conn: s.RwObj}
	}
	if !errors.Is(error, public.ProvideForwardingServiceOnly) {
		if s.Response.Response != nil {
			_ = s.Conn.SetDeadline(time.Now().Add(10 * time.Second))
			for k, v := range s.Response.Header {
				s.Response.rw.Header()[k] = v
			}
			s.Response.rw.WriteHeader(s.Response.StatusCode)
			if s.Response.Body != nil {
				bodyBytes, _ := ioutil.ReadAll(s.Response.Body)
				_, _ = s.Response.rw.Write(bodyBytes)
			}
			return
		}
	}
	if s.Request.Header.Get("ErrorClose") == "true" {
		return
	}
	er := []byte("")
	if error != nil {
		er = []byte(public.ProcessError(error))
	}
	if s.Response.rw == nil {
		return
	}
	_ = s.Conn.SetDeadline(time.Now().Add(10 * time.Second))
	s.Response.rw.Header().Set("Content-Length", fmt.Sprintf("%d", len(er)))
	s.Response.rw.Header().Set("Content-Type", "text/text; charset=utf-8")
	s.Response.rw.WriteHeader(http.StatusInternalServerError)
	_, _ = s.Response.rw.Write(er)
}

func (s *proxyRequest) doRequest() error {
	if s.Request == nil {
		return errors.New("request is nil")
	}
	if s.Request.URL == nil {
		return errors.New("request.url is nil")
	}
	var do *http.Response
	var n net.Conn
	var err error
	var Close func()
	do, n, err, Close = httpClient.Do(s.Request, s.Proxy, false, s.TlsConfig, s.SendTimeout, s.getTLSValues, s.Conn)
	if err == nil && do != nil {
		if s.rawTarget != 0 {
			whoisLock.Lock()
			obj := httpTypeMap[s.rawTarget]
			if obj != nil {
				if obj._type == whoisUndefined {
					obj._time = time.Now()
					if s.Request.URL.Scheme != "https" {
						obj._type = whoisNoHTTPS
					} else {
						if do.ProtoMajor == 2 {
							obj._type = whoisHTTPS2
						} else {
							obj._type = whoisHTTPS1
						}
					}
				}
			}
			whoisLock.Unlock()
		}
	}
	s.Response.Conn = n
	ip, _ := s.Request.Context().Value(public.SunnyNetServerIpTags).(string)
	if ip != "" {
		s.Response.ServerIP = ip
	} else {
		s.Response.ServerIP = "unknown"
	}
	s.Response.Response = do
	s.Response.Close = Close
	return err
}
func (s *proxyRequest) sendHttps(req *http.Request) {
	s.Target.Parse(req.Host, public.HttpsDefaultPort)
	if req.URL.Port() != public.NULL {
		Port, _ := strconv.Atoi(req.URL.Port())
		s.Target.Port = uint16(Port)
	}
	_, _ = s.RwObj.WriteString(public.TunnelConnectionEstablished)
	s.https()
}

func (s *proxyRequest) https() {
	//判断有没有连接信息，没有连接地址信息就直接返回
	if s.Target.Host == public.NULL || s.Target.Port < 1 {
		return
	}
	if s.Target.Port == 853 {
		sx := dns.GetDnsServer()
		if dns.GetDnsServer() != "localhost" {
			s.Target.Host = sx
		} else {
			s.Target.Host = "223.5.5.5"
		}
	}
	//是否开启了强制走TCP  And 如果是DNS请求则不用判断了，直接强制走TCP
	if (s.Global.isMustTcp || s.Target.Port == 853) && !s.targetIsInterfaceAdders() {
		if s.Global.disableTCP {
			return
		}
		s.NoRepairHttp = true
		//开启了强制走TCP，则按TCP流程处理
		s.MustTcpProcessing(public.TagMustTCP)
		return
	}
	var err error
	var serverName string
	var tlsConn *tls.Conn
	var HelloMsg *tls.ClientHelloMsg
	tlsConfig := &tls.Config{MaxVersion: tls.VersionTLS13, NextProtos: public.HTTP2NextProtos, InsecureSkipVerify: true}
	var hook bytes.Buffer
	s.RwObj.Hook = &hook
	tlsConn = tls.Server(s.RwObj, tlsConfig)
	defer func() {
		//函数退出时 清理TLS会话
		_ = tlsConn.Close()
		tlsConn = nil
	}()
	host := s.Target.String()
	//设置1秒的超时 来判断是否 https 请求 因为正常的非HTTPS TCP 请求也会进入到这里来，需要判断一下
	_ = tlsConn.SetDeadline(time.Now().Add(1 * time.Second))
	//取出第一个字节，判断是否TLS
	peek := tlsConn.Peek(1)
	if len(peek) == 1 && (peek[0] == 22 || peek[0] == 23) {
		//发送数据 如果 不是 HEX 16 或 17 那么肯定不是HTTPS 或TLS-TCP
		//HEX 16=ANSI 22 HEX 17=ANSI 23
		//如果是TLS请求设置3秒超时来处理握手信息
		_ = tlsConn.SetDeadline(time.Now().Add(3 * time.Second))
		//开始握手
		HelloMsg, serverName, err = tlsConn.ClientHello()
		s.RwObj.Hook = nil
		//得到握手信息后 恢复30秒的读写超时
		_ = tlsConn.SetDeadline(time.Now().Add(30 * time.Second))
		if err == nil {
			var certificate *tls.Certificate
			var DNSNames []string
			if !s.targetIsInterfaceAdders() {
				res, cert := ClientIsHttps(s.Target.String())
				if res == whoisUndefined {
					res, cert = ClientRequestIsHttps(s.Global, s.Target.String(), serverName)
				}
				if res == whoisNoHTTPS {
					_ = s.RwObj.Close()
					return
				} else if res == whoisUndefined {
					s.rawTarget = public.SumHashCode(s.Target.String())
				}
				if res == whoisHTTPS1 {
					tlsConfig.NextProtos = public.HTTP1NextProtos
				} else { //res == whoisHTTPS2
					tlsConfig.NextProtos = public.HTTP2NextProtos
				}
				name := ""
				if serverName != "" {
					name = fmt.Sprintf("%s:%d", serverName, s.Target.Port)
				}
				if s.isLoop() {
					certificate, DNSNames, _ = WhoisLoopCache(s.Global, cert, host, s.Global.rootCa, s.Global.rootKey)
				} else {
					certificate, DNSNames, _ = WhoisCache(s.Global, cert, name, host, s.Global.rootCa, s.Global.rootKey)
				}
				isRules := s.Global.tcpRules(serverName, s.Target.Host, DNSNames...)
				if isRules {
					if s.Global.disableTCP {
						return
					}
					s.NoRepairHttp = true
					s.RwObj = ReadWriteObject.NewReadWriteObject(newObjHook(s.RwObj, hook.Bytes()))
					s.MustTcpProcessing(public.TagMustTCP)
					return
				}
			} else {
				certificate, DNSNames, _ = WhoisLoopCache(s.Global, nil, host, s.Global.rootCa, s.Global.rootKey)
			}
			ServerName := s.Target.String()
			for _, v := range DNSNames {
				if ip := net.ParseIP(v); ip == nil {
					if !strings.Contains(v, "*") {
						ServerName = v
						//s.Target.Parse(v, 0)
					}
				}
			}
			if certificate == nil {
				err = noHttps
			} else {
				tlsConfig.Certificates = []tls.Certificate{*certificate}
				if serverName != "" {
					tlsConfig.ServerName = serverName
				} else {
					tlsConfig.ServerName = ServerName
				}
				tlsConfig.InsecureSkipVerify = true
				//tlsConfig.CipherSuites=
				//继续握手
				err = tlsConn.ServerHandshake(HelloMsg)
				if err != nil {
					s.Target.Parse(serverName, "")
					s.Request = new(http.Request)
					s.Request.URL, _ = url.Parse(public.HttpsRequestPrefix + strings.ReplaceAll(s.Target.Host, public.Space, public.NULL))
					s.Request.Host = strings.ReplaceAll(s.Target.Host, public.Space, public.NULL)
					ess := err.Error()
					if strings.Index(ess, "unknown certificate") != -1 ||
						strings.Index(ess, "An existing connection was forcibly closed by the remote host") != -1 ||
						strings.Index(ess, "An established connection was aborted by the software in your host machine") != -1 ||
						strings.Index(ess, "client offered only unsupported versions") != -1 {
						s.Error(err, true)
						return
					}
					s.Error(errors.New(fmt.Sprintf("%s [ %s ]", clientHandshakeFail, err.Error())), true)
					return
				}
				_ = tlsConn.SetDeadline(time.Now().Add(30 * time.Second))
			}

		}
	} else {
		isRules := s.Global.tcpRules(serverName, s.Target.Host)
		if isRules {
			if s.Global.disableTCP {
				return
			}
			s.NoRepairHttp = true
			s.RwObj = ReadWriteObject.NewReadWriteObject(newObjHook(s.RwObj, hook.Bytes()))
			s.MustTcpProcessing(public.TagMustTCP)
			return
		}
		err = noHttps
	}
	s.RwObj.Hook = nil
	if err != nil {
		//以上握手过程中 有错误产生 有错误则不是TLS
		//判断这些错误信息，是否还能继续处理
		if s.Global.isMustTcp == false && (err == io.EOF || strings.Index(err.Error(), "An existing connection was forcibly closed by the remote host.") != -1 || strings.Index(err.Error(), "An established connection was aborted by the software in your host machine") != -1) {
			s.Request = new(http.Request)
			s.Request.URL, _ = url.Parse(public.HttpsRequestPrefix + strings.ReplaceAll(s.Target.Host, public.Space, public.NULL))
			s.Request.Host = strings.ReplaceAll(s.Target.Host, public.Space, public.NULL)
			s._Display = true
			s.Error(errors.New("The client closes the connection "), true)
			return
		}
		//将TLS握手过程中的信息取出来
		bs := hook.Bytes()
		if len(bs) == 0 {
			//如果没有客户端没有主动发送数据的话
			//强制走TCP，按TCP流程处理
			s.MustTcpProcessing(public.TagTcpAgreement)
			return
		}
		//证书无效
		if s.Global.isMustTcp == false && strings.Index(err.Error(), "unknown certificate") != -1 || strings.Index(err.Error(), "client offered only unsupported versions") != -1 {
			s.Request = new(http.Request)
			if serverName == public.NULL {
				s.Request.URL, _ = url.Parse(public.HttpsRequestPrefix + s.Target.Host)
				s.Request.Host = strings.ReplaceAll(s.Target.Host, public.Space, public.NULL)
			} else {
				s.Request.URL, _ = url.Parse(public.HttpsRequestPrefix + serverName)
				s.Request.Host = strings.ReplaceAll(serverName, public.Space, public.NULL)
			}
			s.Error(err, true)
			return
		}
		//如果是其他错误，进行http处理流程，继续判断
		s.httpProcessing(bs, public.TagTcpAgreement)
		return
	}
	// 以上握手过程中 没有错误产生 说明是https 或TLS-TCP
	s.Conn = tlsConn                                      //重新保存TLS会话
	s.RwObj = ReadWriteObject.NewReadWriteObject(tlsConn) //重新包装读写对象
	//s.MustTcpProcessing(nil, public.TagTcpSSLAgreement)
	s.TlsConfig = tlsConfig
	s.httpProcessing(nil, public.TagTcpSSLAgreement)
}

var clientHandshakeFail = `与客户端握手失败`
var noHttps = errors.New("No HTTPS ")

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
			if as.Mt != websocket.BinaryMessage {
				//发到服务器
				err = Server.WriteMessage(as.Mt, as.Data.Bytes())
			} else {
				err = Server.WriteFullMessage(as.Mt, as.Data.Bytes())
			}
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
func (s *Sunny) proxyRules(Host string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.proxyRegexp == nil {
		return false
	}
	if Host == "" {
		return false
	}
	x := s.proxyRegexp.MatchString(Host)
	//fmt.Println("proxyRegexp", Host, x)
	return x
}
func (s *Sunny) tcpRules(server, Host string, dns ...string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.isMustTcp {
		return true
	}
	//如果是DNS请求则不用判断了，直接强制走TCP
	if strings.HasSuffix(server, ":853") {
		return true
	}
	if s.mustTcpRulesAllow {
		//规则内走TCP
		{
			if s.mustTcpRegexp == nil {
				return false
			}
			if server != "" {
				if s.mustTcpRegexp.MatchString(server) {
					return true
				}
			}
			if Host != "" {
				if s.mustTcpRegexp.MatchString(Host) {
					return true
				}
			}

			for _, v := range dns {
				if v == "" {
					continue
				}
				x := s.mustTcpRegexp.MatchString(v)
				if x {
					return true
				}
			}
		}
		return false
	}
	//规则内不走TCP
	{
		if s.mustTcpRegexp == nil {
			return true
		}
		if s.mustTcpRegexp.MatchString(server) {
			return false
		}
		if s.mustTcpRegexp.MatchString(Host) {
			return false
		}
		for _, v := range dns {
			if v == "" {
				continue
			}
			x := s.mustTcpRegexp.MatchString(v)
			if x {
				return false
			}
		}
	}
	return true
}
func (s *proxyRequest) CompleteRequest(req *http.Request) {
	//储存 要发送的请求体
	s.Request = req
	defer func() {
		if s.Request != nil {
			if s.Request.Body != nil {
				_ = s.Request.Body.Close()
			}
			RawBody, isRawBody := s.Request.Context().Value(public.SunnyNetRawRequestBody).(io.ReadCloser)
			if isRawBody {
				_ = RawBody.Close()
				s.Request.SetContext(public.SunnyNetRawRequestBody, nil)
			}
		}
		s.Request = nil
		if s.Response.Response != nil {
			if s.Response.Body != nil {
				_ = s.Response.Body.Close()
			}
			s.Response.Response = nil
		}
		s.Proxy = nil
		req = nil
	}()
	//继承全局上游代理
	if !s.Global.proxyRules(s.Target.Host) {
		s.Proxy = s.Global.proxy.Clone()
		if s.Proxy != nil {
			s.Proxy.Regexp = s.Global.proxyRules
		}
	}
	if s.Request != nil && s.Request.URL != nil {
		if s.Request.URL.Scheme == "https" {
			s.TlsConfig = HttpCertificate.GetTlsConfig(s.Request.URL.Host, public.CertificateRequestManagerRulesSend).Clone()
			if s.TlsConfig == nil {
				s.TlsConfig = &tls.Config{InsecureSkipVerify: true}
			}
			tv := s.getTLSValues()
			if len(tv) > 0 {
				s.TlsConfig.CipherSuites = tv
			}
			s.TlsConfig.NextProtos = public.HTTP2NextProtos
		}
	}
	{
		//记录原始Body
		var RequestBody = s.Request.Body
		{
			if s.IsRequestRawBody() {
				s.Request.Body = io.NopCloser(bytes.NewBuffer(public.MaxUploadMsg)) //替换为提示信息在回调中显示
			}
		}
		//通知回调 即将开始发送请求
		s.CallbackBeforeRequest()
		{
			if s.outRouterIP != nil {
				req.SetContext(public.OutRouterIPKey, s.outRouterIP)
			}
			if s.IsRequestRawBody() {
				//当回调中处理完毕后,替换为原始Body
				if s.Request.Body != nil {
					_ = s.Request.Body.Close()
				}
				s.Request.Body = RequestBody
				RawRequestBodyLength, isRawRequestBodyLength := s.Request.Context().Value(public.SunnyNetRawRequestBodyLength).(int64)
				if isRawRequestBodyLength {
					s.Request.Header.Set("Content-Length", strconv.Itoa(int(RawRequestBodyLength)))
				}
			}
		}
	}

	//回调中设置 不发送 直接响应指定数据 或终止发送
	if s.Response.Response != nil {
		s.Response.ServerIP = fmt.Sprintf("%s:%d", "127.0.0.1", s.Global.port)
		s.Response.Response.ProtoMajor, s.Response.Response.ProtoMinor = s.Request.ProtoMajor, s.Request.ProtoMinor
		s.Response.Response.Proto = s.Request.Proto
		s.CallbackBeforeResponse()
		s.Response.Done()
		return
	}
	//验证处理是否websocket请求,如果是直接处理
	if s.handleWss() {
		return
	}
	//为了保证在请求完成时,还能获取到到请求的提交信息,先备份数据
	bakBytes := s.Request.GetData()
	err := s.doRequest()

	//为了保证在请求完成时,还能获取到到请求的提交信息,这里还原数据
	s.Request.SetData(bakBytes)
	defer func() {
		if s.Response.Close != nil {
			s.Response.Close()
		}
	}()
	if err != nil || s.Response.Response == nil {
		if s.Response.Response == nil && err == nil {
			err = errors.New("No data obtained. ")
		}
		s.Error(err, true)
		return
	}

	if s.Response.Header == nil {
		err = errors.New("Response.Header=null")
		s.Error(err, true)
		return
	}
	Length := -1
	Length_ := s.Response.Header.Get("Content-Length")
	if Length_ != "" {
		Length, _ = strconv.Atoi(Length_)
	}
	Method := ""
	if req != nil {
		Method = req.Method
	}
	s.copyBuffer(Method, Length)
	//	s.copyBuffer(s.Response.Body, s.Conn, s.ResponseConn, SetBodyValue, Length, SetReqHeadsValue, s.Response.Header.Get("Content-Type"), setOut, Method)
}
func (s *proxyRequest) RawRequestDataToFile(SaveFilePath string) bool {
	if s == nil {
		return false
	}
	s.Lock.Lock()
	defer s.Lock.Unlock()
	if s.Request == nil {
		return false
	}
	s.Request.SetContext(public.SunnyNetRawBodySaveFilePath, SaveFilePath)
	return true
}

// IsRequestRawBody 此请求是否为原始body 如果是 将无法修改提交的Body，请使用 RawRequestDataToFile 命令来储存到文件
func (s *proxyRequest) IsRequestRawBody() bool {
	if s == nil {
		return false
	}
	s.Lock.Lock()
	defer s.Lock.Unlock()
	if s.Request == nil {
		return false
	}
	return s.Request.IsRawBody
}

// CopyBuffer 转发数据
// rw http.ResponseWriter, src io.Reader, dstConn net.Conn, srcConn net.Conn, SetBodyValue func([]byte, error) []byte, ExpectLen int, SetReqHeadsValue func(string) []byte, ContentType string, setOut func(), Method string
func (s *proxyRequest) copyBuffer(Method string, ExpectLen int) {
	ContentType := s.Response.Header.Get("Content-Type")
	if ContentType == "" {
		for k, v := range s.Response.Header {
			if strings.EqualFold(k, "Content-Type") {
				if len(v) > 0 {
					ContentType = v[0]
					break
				}
			}
		}
	}

	dstConn := s.Response.Conn
	size := 512
	MaxSize := 5 * 1024 * 1024 //5M
	IsText := public.ContentTypeIsText(ContentType)
	if IsText && ExpectLen < 1 {
		MaxSize = 5 * 1024 * 1024 * 10 //50M
		size = 32 * 1024
	}
	buf := make([]byte, size)
	var buff bytes.Buffer
	defer func() {
		buff.Reset()
		buf = make([]byte, 0)
		buf = nil
	}()
	var isForward = false
	// 是否是大文件类型 是的话,不判断长度直接转发 并且长度大于指定值(5M) 则直接转发
	var ToIsForward = public.IsForward(ContentType) && (ExpectLen < 1 || ExpectLen > 5*1024*1024) //5M

	if Method == public.HttpMethodHEAD {
		s.CallbackBeforeResponse()
		s.Response.WriteHeader(strconv.Itoa(ExpectLen))
		_ = dstConn.SetDeadline(time.Now().Add(5 * time.Second))
		return
	}
	for {
		_ = s.Response.Conn.SetDeadline(time.Now().Add(time.Duration(30) * time.Second))
		nr, er := s.Response.Body.Read(buf)
		if nr > 0 {
			buff.Write(buf[0:nr])
			if ToIsForward || (isForward || ExpectLen > MaxSize || (ExpectLen < 1 && buff.Len() > MaxSize)) {
				_ = dstConn.SetDeadline(time.Now().Add(5 * time.Second))
				if isForward == false {
					isForward = true
					_ = dstConn.SetDeadline(time.Time{})
					s.Error(public.ProvideForwardingServiceOnly, s._Display)
					s.Response.WriteHeader(strconv.Itoa(ExpectLen))
				}
				nr = buff.Len()
				nw, ew := s.Response.Write(public.CopyBytes(buff.Bytes()))
				buff.Reset()
				buf = make([]byte, 40960)
				if nw < 0 || nr < nw {
					nw = 0
					if ew == nil {
						ew = errors.New("invalid write result")
					}
				}
				if ew != nil {
					return
				}
				if nr != nw {
					return
				}
				if er != nil {
					return
				}
				continue
			} else if ExpectLen > 0 && ExpectLen == buff.Len() {
				er = io.EOF
			}
		}
		if er != nil {
			if buff.Len() >= 0 {
				if s.Response.Body != nil {
					_ = s.Response.Body.Close()
				}
				s.Response.Body = ioutil.NopCloser(bytes.NewBuffer(buff.Bytes()))

				s.CallbackBeforeResponse()

				_body, _ := s.ReadAll(s.Response.Body)
				_ = s.Conn.SetDeadline(time.Time{})
				s.Response.WriteHeader(strconv.Itoa(len(_body)))
				_, _ = s.Response.Write(_body)
				_body = make([]byte, 0)
			}
			return
		}
	}
}

// SetOutRouterIP 设置TCP/HTTP数据出口IP 请传入网卡对应的IP地址,用于指定网卡,例如 192.168.31.11
func (s *proxyRequest) SetOutRouterIP(RouterIP string) bool {
	if RouterIP == "" {
		s.outRouterIP = nil
		return true
	}
	ok, ip := public.IsLocalIP(RouterIP)
	if !ok {
		return false
	}
	if ip.To4() != nil {
		localAddr, err := net.ResolveTCPAddr("tcp", RouterIP+":0")
		if err != nil {
			return false
		}
		s.outRouterIP = localAddr
		return true
	}
	localAddr, err := net.ResolveTCPAddr("tcp", "["+RouterIP+"]:0")
	if err != nil {
		return false
	}
	s.outRouterIP = localAddr
	return true
}
func (s *proxyRequest) sendHttp(req *http.Request) {
	if req.URL == nil {
		return
	}
	if req.Method == public.HttpMethodCONNECT {
		s.sendHttps(req)
		return
	}
	if s.isCerDownloadPage(req) { // 安装移动端证书
		return
	}
	if req.URL.Scheme == "http" {
		if s.Target.Host == "" {
			if req.URL.Port() == "" {
				s.Target.Parse(req.Host, "80")
			} else {
				s.Target.Parse(req.Host, req.URL.Port())
			}
		}
		if s.Global.tcpRules(req.Host, s.Target.String()) {
			var buff bytes.Buffer
			_ = req.Write(&buff)
			s.NoRepairHttp = true
			s.RwObj = ReadWriteObject.NewReadWriteObject(newObjHook(s.RwObj, buff.Bytes()))
			s.MustTcpProcessing(public.TagMustTCP)
			return
		}
	}
	s.CompleteRequest(req)
}

func (s *proxyRequest) ReadAll(r io.Reader) ([]byte, error) {
	var bufBuffer bytes.Buffer
	b := make([]byte, 4096)
	defer func() {
		b = make([]byte, 0)
		bufBuffer.Reset()
		bufBuffer.Grow(0)
		b = nil
	}()
	for {
		n, err := r.Read(b[0:])
		bufBuffer.Write(b[0:n])
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return public.CopyBytes(bufBuffer.Bytes()), err
		}
	}
}

/*
SocketForward
MsgType ==1 dst=服务器端 src=客户端
MsgType ==2 dst=客户端 src=服务器端
*/
func (s *proxyRequest) SocketForward(dst bufio.Writer, src *ReadWriteObject.ReadWriteObject, MsgType int, t1, t2 net.Conn, TCP *public.TCP, isHttpReq *bool, RemoteAddr string) {
	as := &public.TcpMsg{}
	length := 4096
	MaxLength := 40960
	MaxMaxLength := MaxLength * 2
	MaxCount1 := 0
	MaxCount2 := 0
	buf := make([]byte, length)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("SocketForward 出了错：", err)
		}
		as.Data.Reset()
		//是否已经纠正为了HTTP请求
		if !*isHttpReq {
			//如果没有纠正，退出SocketForward函数时将关闭socket会话
			buf = nil
			if t1 != nil {
				_ = t1.Close()
			}
			if t2 != nil {
				_ = t2.Close()
			}
		}
		buf = make([]byte, 0)
	}()
	if t1 == nil {
		return
	}
	firstRequest := true //是否是首次接收请求
	if s.Global.isMustTcp || s.NoRepairHttp {
		firstRequest = false
	}
	for {
		TCP.L.Lock()
		_ = t1.SetDeadline(time.Now().Add(165 * time.Second))
		_ = t2.SetDeadline(time.Now().Add(165 * time.Second))
		TCP.L.Unlock()
		if firstRequest {
			//是否是客户端发送数据
			if MsgType == public.SunnyNetMsgTypeTCPClientSend {
				{
					//提取取出1个字节,
					peek, e := src.Peek(1)
					if e == nil {
						if len(peek) > 0 {
							//判断是否是HTTP请求
							if public.IsHTTPRequest(peek[0], src) {
								_ = t2.Close()
								//如果是，那么关闭本次连接服务器的socket，并且纠正为HTTP请求，后续交给HTTP请求处理函数继续处理
								*isHttpReq = true
								return
							}
						}
					}
					if s.Global.disableTCP {
						_ = t1.Close()
						_ = t2.Close()
						return
					}
				}
			}
			firstRequest = false
		}
		nr, er := src.Read(buf[0:]) // io.ReadAtLeast(src, buf[0:], 1)
		{
			//自动扩容，优化响应速度
			if nr == length {
				//如果连续10次接收大小为 4096 装满默认容器，那么就扩容到 40960
				MaxCount1++
				if MaxCount1 >= 10 {
					buf = resize(buf, MaxLength)
				}
			} else if nr == MaxLength {
				//如果连续10次接收大小为 40960 装满默认容器，那么就扩容到 81920，尽量不要扩容太大，否则可能会导致内存占用太高
				MaxCount2++
				if MaxCount2 >= 10 {
					buf = resize(buf, MaxMaxLength)
				}
			} else if MaxCount1 < 10 {
				MaxCount1 = 0
			} else if MaxCount2 < 10 {
				MaxCount2 = 0
			}
		}
		if nr > 0 {
			as.Data.Reset()
			as.Data.Write(buf[0:nr])
			s.CallbackTCPRequest(MsgType, as, RemoteAddr)
			if as.Data.Len() < 1 {
				continue
			}
			TCP.L.Lock()
			nw, ew := dst.Write(as.Data.Bytes())
			er = dst.Flush()
			TCP.L.Unlock()
			if nw != as.Data.Len() || ew != nil {
				break
			}
		}
		if er != nil {
			return
		}
	}
}
func resize(slice []byte, newLength int) []byte {
	if newLength <= cap(slice) {
		return slice[:newLength] // 如果容量足够，直接返回切片
	}

	// 创建一个新的切片，大小为 newLength
	newSlice := make([]byte, newLength)

	// 复制原始数据到新切片
	copy(newSlice, slice)

	// 释放原始切片
	slice = nil // 将原始切片设置为 nil，帮助垃圾回收器回收

	return newSlice
}

var divert ProcessDrv.Dev //使用的驱动
// Sunny  请使用 NewSunny 方法 请不要直接构造
type Sunny struct {
	disableTCP            bool              //禁止TCP连接
	disableUDP            bool              //禁止TCP连接
	certificates          []byte            //CA证书原始数据
	rootCa                *x509.Certificate //中间件CA证书
	rootKey               *rsa.PrivateKey   // 证书私钥
	initCertOK            bool              // 是否已经初始化证书
	port                  int               //启动的端口号
	Error                 error             //错误信息
	tcpSocket             *net.Listener     //TcpSocket服务器
	udpSocket             *net.UDPConn      //UdpSocket服务器
	outRouterIP           *net.TCPAddr
	connList              map[int64]net.Conn  //会话连接客户端、停止服务器时可以全部关闭
	lock                  sync.Mutex          //会话连接互斥锁
	socket5VerifyUser     bool                //S5代理是否需要验证账号密码
	socket5VerifyUserList map[string]string   //S5代理需要验证的账号密码列表
	socket5VerifyUserLock sync.Mutex          //S5代理验证时的锁
	isMustTcp             bool                //强制走TCP
	httpCallback          int                 //http 请求回调地址
	tcpCallback           int                 //TCP请求回调地址
	websocketCallback     int                 //ws请求回调地址
	udpCallback           int                 //udp请求回调地址
	goHttpCallback        func(ConnHTTP)      //http请求GO回调地址
	goTcpCallback         func(ConnTCP)       //TCP请求GO回调地址
	goWebsocketCallback   func(ConnWebSocket) //ws请求GO回调地址
	goUdpCallback         func(ConnUDP)       //UDP请求GO回调地址
	proxy                 *SunnyProxy.Proxy   //全局上游代理
	proxyRegexp           *regexp.Regexp      //上游代理使用规则
	mustTcpRegexp         *regexp.Regexp      //强制走TCP规则,如果 isMustTcp 打开状态,本功能则无效
	mustTcpRulesAllow     bool                // true 表示 mustTcpRegexp 规则内的强制走TCP，反之不在规则内的强制都TCP
	isRun                 bool                //是否在运行中
	SunnyContext          int
	isRandomTLS           bool   //是否随机使用TLS指纹
	userScriptCode        []byte //用户脚本代码
	httpMaxBodyLen        int64  //最大的用户提交数据长度
	connHijack            func(Hijack) bool
	script                struct {
		http         GoScriptCode.GoScriptTypeHTTP  //脚本代码	HTTP		事件入口函数
		tcp          GoScriptCode.GoScriptTypeTCP   //脚本代码	TCP			事件入口函数
		udp          GoScriptCode.GoScriptTypeUDP   //脚本代码	UDP			事件入口函数
		websocket    GoScriptCode.GoScriptTypeWS    //脚本代码	Websocket	事件入口函数
		SaveCallback GoScriptCode.SaveFuncInterface //保存代码执行的回调函数
		LogCallback  GoScriptCode.LogFuncInterface  //日志输出执行的回调函数
		AdminPage    string                         //管理页面
	}
}

func (s *Sunny) scriptHTTPCall(arg Interface.ConnHTTPScriptCall) {
	s.lock.Lock()
	_call := s.script.http
	s.lock.Unlock()
	if _call != nil {
		defer func() {
			if err := recover(); err != nil {
				//fmt.Println("script HTTP Call 出了错：", err)
			}
		}()
		_call(arg)
	}
}
func (s *Sunny) scriptTCPCall(arg Interface.ConnTCPScriptCall) {
	s.lock.Lock()
	_call := s.script.tcp
	s.lock.Unlock()
	if _call != nil {
		defer func() {
			if err := recover(); err != nil {
				//fmt.Println("script TCP Call 出了错：", err)
			}
		}()
		_call(arg)
	}
}

func (s *Sunny) scriptUDPCall(arg Interface.ConnUDPScriptCall) {
	s.lock.Lock()
	_call := s.script.udp
	s.lock.Unlock()
	if _call != nil {
		defer func() {
			if err := recover(); err != nil {
				//fmt.Println("script UDP Call 出了错：", err)
			}
		}()
		_call(arg)
	}

}

func (s *Sunny) scriptWebsocketCall(arg Interface.ConnWebSocketScriptCall) {
	s.lock.Lock()
	_call := s.script.websocket
	s.lock.Unlock()
	if _call != nil {
		defer func() {
			if err := recover(); err != nil {
				//fmt.Println("script Websocket Call 出了错：", err)
			}
		}()
		_call(arg)
	}
}

// SetRandomTLS 是否使用随机TLS指纹
func (s *Sunny) SetRandomTLS(open bool) {
	if s == nil {
		return
	}
	s.lock.Lock()
	s.isRandomTLS = open
	s.lock.Unlock()
}

var defaultManager = func() int {
	i := Certificate.CreateCertificate()
	c := Certificate.LoadCertificateContext(i)
	if c == nil {
		panic(errors.New("创建证书管理器错误！！"))
	}
	c.LoadX509Certificate(public.NULL, public.RootCa, public.RootKey)
	return i
}()

// NewSunny 创建一个中间件
func NewSunny() *Sunny {
	SunnyContext := NewMessageId()
	a, _ := regexp.Compile("ALL")
	s := &Sunny{SunnyContext: SunnyContext, connList: make(map[int64]net.Conn), socket5VerifyUserList: make(map[string]string), proxyRegexp: a, httpMaxBodyLen: public.MaxUploadLength, mustTcpRegexp: a, mustTcpRulesAllow: true}
	s.userScriptCode = GoScriptCode.DefaultCode
	s.script.AdminPage = "SunnyNetScriptEdit"
	_, s.script.http, s.script.websocket, s.script.tcp, s.script.udp = GoScriptCode.RunCode(SunnyContext, s.userScriptCode, nil)
	s.SetCert(defaultManager)
	SunnyStorageLock.Lock()
	SunnyStorage[s.SunnyContext] = s
	SunnyStorageLock.Unlock()
	return s
}

// SetMustTcpRegexp 设置强制走TCP规则,如果 打开了全部强制走TCP状态,本功能则无效 Rules=false 规则之外走TCP  Rules=true 规则之内走TCP
func (s *Sunny) SetMustTcpRegexp(RegexpList string, Rules bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	r := strings.ReplaceAll("^"+strings.ReplaceAll(RegexpList, " ", "")+"$", "\r", "")
	r = strings.ReplaceAll(r, "\t", "")
	r = strings.ReplaceAll(r, "\n", ";")
	r = strings.ReplaceAll(r, ";", "$|^")
	r = strings.ReplaceAll(r, ".", "\\.")
	r = strings.ReplaceAll(r, "*", ".*.?")
	if r == "" {
		r = "ALL"
	}
	a, e := regexp.Compile(r)
	s.mustTcpRulesAllow = Rules
	if e == nil {
		s.mustTcpRegexp = a
	} else {
		s.mustTcpRegexp = nil
	}
	return e
}

// CompileProxyRegexp 创建上游代理使用规则
func (s *Sunny) CompileProxyRegexp(Regexp string) error {
	r := strings.ReplaceAll("^"+strings.ReplaceAll(Regexp, " ", "")+"$", "\r", "")
	r = strings.ReplaceAll(r, "\t", "")
	r = strings.ReplaceAll(r, "\n", ";")
	r = strings.ReplaceAll(r, ";", "$|^")
	r = strings.ReplaceAll(r, ".", "\\.")
	r = strings.ReplaceAll(r, "*", ".*.?")
	if r == "" {
		r = "ALL" //让其全部匹配失败，也就是全部使用上游代理代理
	}
	a, e := regexp.Compile(r)
	s.lock.Lock()
	defer s.lock.Unlock()
	if e == nil {
		s.proxyRegexp = a
	} else {
		a1, _ := regexp.Compile("ALL")
		s.proxyRegexp = a1 //让其全部匹配失败，也就是全部使用上游代理代理
	}
	return e
}

// MustTcp 设置是否强制全部走TCP
func (s *Sunny) MustTcp(open bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.isMustTcp = open
}

// SetOutRouterIP 设置TCP/HTTP数据出口IP 请传入网卡对应的IP地址,用于指定网卡,例如 192.168.31.11
func (s *Sunny) SetOutRouterIP(RouterIP string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if RouterIP == "" {
		s.outRouterIP = nil
		return true
	}
	ok, ip := public.IsLocalIP(RouterIP)
	if !ok {
		return false
	}
	if ip.To4() != nil {
		localAddr, err := net.ResolveTCPAddr("tcp", RouterIP+":0")
		if err != nil {
			return false
		}
		s.outRouterIP = localAddr
		return true
	}
	localAddr, err := net.ResolveTCPAddr("tcp", "["+RouterIP+"]:0")
	if err != nil {
		return false
	}
	s.outRouterIP = localAddr
	return true
}

// Socket5VerifyUser S5代理是否需要验证账号密码
func (s *Sunny) Socket5VerifyUser(n bool) *Sunny {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.socket5VerifyUser = n
	return s
}

// Socket5AddUser S5代理添加需要验证的账号密码
func (s *Sunny) Socket5AddUser(u, p string) *Sunny {
	s.socket5VerifyUserLock.Lock()
	s.socket5VerifyUserList[u] = p
	s.socket5VerifyUserLock.Unlock()
	return s
}

// Socket5DelUser S5代理删除需要验证的账号
func (s *Sunny) Socket5DelUser(u string) *Sunny {
	s.socket5VerifyUserLock.Lock()
	delete(s.socket5VerifyUserList, u)
	s.socket5VerifyUserLock.Unlock()
	return s
}

// ExportCert 获取证书原内容
func (s *Sunny) ExportCert() []byte {
	ar := strings.Split(strings.ReplaceAll(string(s.certificates), "\r", public.NULL), "\n")
	var b bytes.Buffer
	for _, v := range ar {
		if strings.Index(v, ": ") == -1 && len(v) > 0 {
			b.WriteString(v + "\r\n")
		}
	}
	return public.CopyBytes(b.Bytes())
}

// SetIEProxy 设置IE代理 设置后请使用 CancelIEProxy 取消设置的IE代理
func (s *Sunny) SetIEProxy() bool {
	return CrossCompiled.SetIeProxy(false, s.Port())
}

// CancelIEProxy 取消设置的IE代理
func (s *Sunny) CancelIEProxy() bool {
	return CrossCompiled.SetIeProxy(true, s.Port())
}

// SetGlobalProxy 设置全局上游代理 仅支持Socket5和http 例如 socket5://admin:123456@127.0.0.1:8888 或 http://admin:123456@127.0.0.1:8888
func (s *Sunny) SetGlobalProxy(ProxyUrl string, outTime int) bool {
	s.proxy, _ = SunnyProxy.ParseProxy(ProxyUrl, outTime)
	return s.proxy != nil
}

// InstallCert 安装证书 将证书安装到Windows系统内
func (s *Sunny) InstallCert() string {
	return CrossCompiled.InstallCert(s.certificates)
}

// SetCert 设置证书
func (s *Sunny) SetCert(ManagerContext int) *Sunny {
	Manager := Certificate.LoadCertificateContext(ManagerContext)
	if Manager == nil {
		s.Error = errors.New("CertificateManager invalid ")
		return s
	}
	var err error
	s.initCertOK = false
	p, _ := pem.Decode([]byte(Manager.ExportCA()))
	s.certificates = nil
	s.rootCa, err = x509.ParseCertificate(p.Bytes)
	if err != nil {
		s.Error = err
		return s
	}
	s.certificates = []byte(Manager.ExportCA())
	p1, _ := pem.Decode([]byte(Manager.ExportKEY()))
	if p1 == nil {
		s.Error = errors.New("Key证书解析失败 ")
		return s
	}
	s.rootKey, err = x509.ParsePKCS1PrivateKey(p1.Bytes)
	if err != nil {
		k, e := x509.ParsePKCS8PrivateKey(p1.Bytes)
		if e != nil {
			s.Error = errors.New(err.Error() + " or " + e.Error())
			return s
		}
		kk := k.(*rsa.PrivateKey)
		if kk == nil {
			s.Error = err
			return s
		}
		s.rootKey = kk
	}
	s.initCertOK = true
	return s
}

// SetPort 设置端口号
func (s *Sunny) SetPort(Port int) *Sunny {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.port = Port
	return s
}

// IsScriptCodeSupported 当前SDK是否支持脚本代码
func (s *Sunny) IsScriptCodeSupported() bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.SetScriptPage("") != "no"
}

// SetHTTPRequestMaxUpdateLength 设置HTTP请求,提交数据,最大的长度
func (s *Sunny) SetHTTPRequestMaxUpdateLength(max int64) *Sunny {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.httpMaxBodyLen = max
	return s
}

// DisableTCP 禁用TCP
func (s *Sunny) DisableTCP(disable bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.disableTCP = disable
}

// DisableUDP 禁用UDP
func (s *Sunny) DisableUDP(disable bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.disableUDP = disable
}

// Port 获取端口号
func (s *Sunny) Port() int {
	return s.port
}

// SetCallback 设置回调地址
func (s *Sunny) SetCallback(httpCall, tcpCall, wsCall, udpCall int) *Sunny {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.httpCallback = httpCall
	s.tcpCallback = tcpCall
	s.websocketCallback = wsCall
	s.udpCallback = udpCall
	return s
}

// SetGoCallback 设置Go回调地址
func (s *Sunny) SetGoCallback(httpCall func(ConnHTTP), tcpCall func(ConnTCP), wsCall func(ConnWebSocket), udpCall func(ConnUDP)) *Sunny {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.goHttpCallback = httpCall
	s.goTcpCallback = tcpCall
	s.goWebsocketCallback = wsCall
	s.goUdpCallback = udpCall
	return s
}

// UnDrive 卸载驱动，仅Windows 有效【需要管理权限】执行成功后会立即重启系统,若函数执行后没有重启系统表示没有管理员权限
func (s *Sunny) UnDrive() {
	CrossCompiled.NFAPI{}.UnInstall()
	CrossCompiled.Tun{}.UnInstall()
	//一定要将Pr放到最后,因为写了自动重启
	CrossCompiled.Pr{}.UnInstall()
}

// OpenDrive 开始进程代理 会自动安装所需驱动文件
// DevMode 0=Proxifier,1=NFAPI,2=Tun
func (s *Sunny) OpenDrive(DevMode int) bool {
	if divert != nil {
		fmt.Println("你已选择另一个模式,不可切换")
		return false
	}
	if DevMode == CrossCompiled.DrvNF {
		divert = &CrossCompiled.NFAPI{TCP: s.handleClientConn, UDP: s.udpSendReceive, Sunny: s}
	} else if DevMode == CrossCompiled.DrvTun {
		divert = &CrossCompiled.Tun{TCP: s.handleClientConn, UDP: s.udpSendReceive, Sunny: s}
	} else if DevMode == CrossCompiled.DrvPr {
		//不支持UDP
		divert = &CrossCompiled.Pr{TCP: s.handleClientConn, UDP: s.udpSendReceive, Sunny: s}
	} else {
		return false
	}
	if !divert.Install() {
		divert = nil
		return false
	}
	divert.SetHandle()
	if divert.IsRun() {
		return true
	}
	return divert.Run()
}

// ProcessALLName 是否允许所有进程通过 所有 SunnyNet 通用,
// StopNetwork 是否对所有进程执行一次断网操作
// 请注意GoLang调试时候，StopNetwork请不要设置true
// 因为如果不断开的一次的话,已经建立的TCP链接无法抓包。
// Go程序调试，是通过TCP连接的，若使用此命令将无法调试。
func (s *Sunny) ProcessALLName(open, StopNetwork bool) *Sunny {
	ProcessCheck.HookAllProcess(open, StopNetwork)
	return s
}

// ProcessDelName 删除进程名  所有 SunnyNet 通用
func (s *Sunny) ProcessDelName(name string) *Sunny {
	ProcessCheck.DelName(name)
	//CrossCompiled.NFapi_CloseNameTCP(name)
	return s
}

// ProcessAddName 进程代理 添加进程名 所有 SunnyNet 通用
func (s *Sunny) ProcessAddName(Name string) *Sunny {
	ProcessCheck.AddName(Name)
	//CrossCompiled.NFapi_CloseNameTCP(Name)
	return s
}

// ProcessDelPid 删除PID  所有 SunnyNet 通用
func (s *Sunny) ProcessDelPid(Pid int) *Sunny {
	ProcessCheck.DelPid(uint32(Pid))
	//CrossCompiled.NFapi_ClosePidTCP(Pid)
	return s
}

// ProcessAddPid 进程代理 添加PID 所有 SunnyNet 通用
func (s *Sunny) ProcessAddPid(Pid int) *Sunny {
	ProcessCheck.AddPid(uint32(Pid))
	//CrossCompiled.NFapi_ClosePidTCP(Pid)
	return s
}

// ProcessCancelAll 进程代理 取消全部已设置的进程名
func (s *Sunny) ProcessCancelAll() *Sunny {
	ProcessCheck.CancelAll()
	//CrossCompiled.NFapi_ClosePidTCP(-1)
	return s
}

// SetScriptCall 设置脚本代码的回调函数
func (s *Sunny) SetScriptCall(log GoScriptCode.LogFuncInterface, save GoScriptCode.SaveFuncInterface) {
	s.lock.Lock()
	s.script.SaveCallback = save
	s.script.LogCallback = log
	s.lock.Unlock()
}

// Start 开始启动  调用 Error 获取错误信息 成功=nil
func (s *Sunny) Start() *Sunny {
	if s.isRun {
		s.Error = errors.New("已在运行中")
		return s
	}
	if s.port == 0 {
		s.Error = errors.New("未设置的端口号")
		return s
	}
	if !s.initCertOK {
		return s
	}
	CrossCompiled.AddFirewallRule()
	tcpListen, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(s.port))
	if err != nil {
		s.Error = err
		return s
	}
	udpListenAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:"+strconv.Itoa(s.port))
	if err != nil {
		s.Error = err
		_ = tcpListen.Close()
		return s
	}
	udpListen, err := net.ListenUDP("udp", udpListenAddr)
	if err != nil {
		s.Error = err
		_ = tcpListen.Close()
		return s
	}
	s.udpSocket = udpListen
	s.tcpSocket = &tcpListen
	s.Error = err
	s.isRun = true
	if divert != nil {
		if divert.Install() {
			divert.SetHandle()
			if !divert.IsRun() {
				divert.Run()
			}
		}
	}
	go s.listenTcpGo()
	go s.listenUdpGo()
	return s
}

// Close 关闭服务器
func (s *Sunny) Close() *Sunny {
	if s.tcpSocket != nil {
		_ = (*s.tcpSocket).Close()
	}
	if s.udpSocket != nil {
		_ = s.udpSocket.Close()
	}
	s.lock.Lock()
	for k, conn := range s.connList {
		_ = conn.Close()
		delete(s.connList, k)
	}
	if divert != nil {
		divert.Close()
	}
	s.lock.Unlock()
	return s
}

// listenTcpGo 循环监听
func (s *Sunny) listenTcpGo() {
	defer func() {
		if s.tcpSocket != nil || s.udpSocket != nil {
			s.Close()
		}
	}()
	defer func() { s.isRun = false }()
	for {
		c, err := (*s.tcpSocket).Accept()
		if err != nil && strings.Index(err.Error(), "timeout") == -1 {
			s.Error = err
			break
		}
		if err == nil {
			go s.handleClientConn(c)
		}
	}
}

func (s *proxyRequest) clone() *proxyRequest {
	req := &proxyRequest{
		Global:        s.Global,
		TcpCall:       s.Global.tcpCallback,
		HttpCall:      s.Global.httpCallback,
		wsCall:        s.Global.websocketCallback,
		TcpGoCall:     s.Global.goTcpCallback,
		HttpGoCall:    s.Global.goHttpCallback,
		wsGoCall:      s.Global.goWebsocketCallback,
		Theology:      s.Theology,
		Conn:          s.Conn,
		RwObj:         s.RwObj,
		Target:        s.Target.Clone(),
		ProxyHost:     s.ProxyHost,
		Pid:           s.Pid,
		Request:       s.Request,
		Response:      response{},
		Proxy:         s.Proxy,
		NoRepairHttp:  s.NoRepairHttp,
		defaultScheme: s.defaultScheme,
		SendTimeout:   s.SendTimeout,
		rawTarget:     s.rawTarget,
	}
	if s.outRouterIP != nil {
		req.outRouterIP = &net.TCPAddr{IP: s.outRouterIP.IP}
	}
	req.updateSocket5User()

	Theoni := int64(req.Theology)

	{
		sL.Lock()
		user := sUser[s.Theology]
		if user == "" {
			sUser[req.Theology] = s._SocksUser
			req._SocksUser = s._SocksUser
		}
		sL.Unlock()
	}

	s.Global.lock.Lock()
	s.Global.connList[Theoni] = s.Conn
	delete(s.Global.connList, Theoni)
	s.Global.lock.Unlock()
	return req
}
func (s *proxyRequest) free() {
	if s == nil {
		return
	}
	if s.Global == nil {
		return
	}
	s.delSocket5User()
	s.Global.lock.Lock()
	delete(s.Global.connList, int64(s.Theology))
	s.Global.lock.Unlock()
	//当 handleClientConn 函数 即将退出时 销毁 请求中间件 中的一些信息，避免内存泄漏
	s.RwObj = nil
	s.Conn = nil
	s.Global = nil
	s.Response.Response = nil
	s.Request = nil
	s.Target = nil
}
func (s *proxyRequest) isDriveConn() (ProcessCheck.DrvInfo, uint16) {
	if s == nil {
		return nil, 0
	}
	if divert == nil {
		return nil, 0
	}
	addr, ok := s.Conn.RemoteAddr().(*net.TCPAddr)
	if ok {
		u := uint16(addr.Port)
		info := ProcessCheck.GetTcpConnectInfo(u)
		if info == nil {
			addr, ok = s.Conn.LocalAddr().(*net.TCPAddr)
			if ok {
				u = uint16(addr.Port)
				info = ProcessCheck.GetTcpConnectInfo(u)
				return info, u
			}
			return nil, 0
		}
		return info, u
	}
	return nil, 0
}
func (s *proxyRequest) RandomCipherSuites() {
	s._isRandomCipherSuites = true
}

// getTLSValues 获取固定的TLS指纹列表或随机TLS指纹列表,如果未开启使用随机TLS指纹,并且未设置固定TLS指纹,则返回nil
func (s *proxyRequest) getTLSValues() []uint16 {
	if s == nil {
		return nil
	}
	if !s.Global.isRandomTLS && !s._isRandomCipherSuites {
		return nil
	}
	return public.GetTLSValues()
}

func (s *Sunny) handleClientConn(conn net.Conn) {
	req := &proxyRequest{Global: s, TcpCall: s.tcpCallback, HttpCall: s.httpCallback, wsCall: s.websocketCallback, TcpGoCall: s.goTcpCallback, HttpGoCall: s.goHttpCallback, wsGoCall: s.goWebsocketCallback, SendTimeout: 0} //原始请求对象

	Theoni := atomic.AddInt64(&public.Theology, 1)
	//存入会话列表 方便停止时，将所以连接断开
	s.lock.Lock()
	s.connList[Theoni] = conn
	//构造一个请求中间件
	if s.outRouterIP != nil {
		req.outRouterIP = &net.TCPAddr{IP: s.outRouterIP.IP}
	}
	s.lock.Unlock()

	defer func() {
		//当 handleClientConn 函数 即将退出时 从会话列表中删除当前会话
		_ = conn.Close()
		s.lock.Lock()
		delete(s.connList, Theoni)
		s.lock.Unlock()
		req.free()
		conn = nil
	}()
	//请求中间件一些必要参数赋值
	req.Conn = conn                                      //请求会话
	req.Target = &TargetInfo{}                           //构建一个请求连接信息，后续解析到值后会进行赋值
	req.RwObj = ReadWriteObject.NewReadWriteObject(conn) //构造客户端读写对象
	req.Theology = int(Theoni)                           //当前请求唯一ID
	req.Response = response{}
	info, DrivePort := req.isDriveConn()
	if info != nil {
		req.setSocket5User("驱动程序")
		//如果是 通过 NFapi 驱动进来的数据 对连接信息进行赋值
		req.Pid = info.GetPid()
		req.Target.Parse(info.GetRemoteAddress(), info.GetRemotePort(), info.IsV6())
		if s.connHijack != nil {
			if s.connHijack(&_hijack{req}) {
				return
			}
		}
		//然后进行数据处理,按照HTTPS数据进行处理
		req.https()
		_ = info.Close()
		ProcessCheck.DelTcpConnectInfo(DrivePort)
		return
	}
	req.Pid = CrossCompiled.GetTcpInfoPID(conn.RemoteAddr().String(), s.port)
	//若不是 通过 NFapi 驱动进来的数据 那么就是通过代理传递过来的数据
	//进行预读1个字节的数据
	peek, err := req.RwObj.Peek(1)
	if err != nil {
		//读取1个字节失败直接返回
		return
	}
	//如果第一个字节是0x05 说明是通过S5代理连接的
	if peek[0] == 0x05 {
		//进行S5鉴权
		if req.Socks5ProxyVerification() == false {
			return
		}
		if s.connHijack != nil {
			if s.connHijack(&_hijack{req}) {
				return
			}
		}
		if s.isMustTcp && !req.targetIsInterfaceAdders() {
			if s.disableTCP {
				return
			}
			//如果开启了强制走TCP ，则按TCP处理流程处理
			req.MustTcpProcessing(public.TagMustTCP)
			return
		}
		//如果没有开启强制走TCP，则按https 数据进行处理
		req.https()
		return
	}
	if s.connHijack != nil {
		if s.connHijack(&_hijack{req}) {
			return
		}
	}
	//如果没有开启用户身份验证 且 第一个字节是 22 或 23 说明可能是透明代理
	if s.socket5VerifyUser == false && (peek[0] == 22 || peek[0] == 23) {
		//按透明代理处理流程处理
		req.transparentProcessing()
		return
	}
	//如果没有开启用户身份验证 且 第一个字节符合HTTP/S 请求头
	if s.socket5VerifyUser == false && public.IsHTTPRequest(peek[0], req.RwObj) {
		//按照http请求处理
		req.httpProcessing(nil, public.TagTcpAgreement)
	}
}
func (s *Sunny) SetDnsServer(server string) {
	dns.SetDnsServer(server)
}

// SetHijack 设置劫持函数 函数返回一个bool 如果返回true 表示，此连接过程已由您自行处理SunnyNet不再处理该连接
func (s *Sunny) SetHijack(fn func(hijack Hijack) bool) {
	s.lock.Lock()
	s.connHijack = fn
	s.lock.Unlock()
}

type Hijack interface {
	Conn() net.Conn     //劫持的会话
	Pid() int           //如果是远程连接PID=0
	Username() string   //如果是socks连接并且传递了账号密码
	RemoteAddr() string //远端地址 以这个为准 Conn的RemoteAddr可能不准确
	LocalAddr() string  //来源地址 和 Conn 的 LocalAddr 一致
}
type _hijack struct {
	*proxyRequest
}

// Pid 如果是远程连接PID=0
func (t *_hijack) Pid() int {
	o, _ := strconv.Atoi(t.proxyRequest.Pid)
	return o
}

// Conn 劫持的会话
func (t *_hijack) Conn() net.Conn {
	return t.RwObj
}

// Username 如果是socks连接并且传递了账号密码
func (t *_hijack) Username() string {
	return t.proxyRequest._SocksUser
}

// RemoteAddr 远端地址 以这个为准 Conn的RemoteAddr可能不准确
func (t *_hijack) RemoteAddr() string {
	if t.Target == nil {
		return t.RwObj.RemoteAddr().String()
	}
	return t.Target.String()
}

// LocalAddr 来源地址 和 Conn 的 LocalAddr 一致
func (t *_hijack) LocalAddr() string {
	return t.RwObj.LocalAddr().String()
}
