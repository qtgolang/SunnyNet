package SunnyNet

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/qtgolang/SunnyNet/src/HttpCertificate"
	"github.com/qtgolang/SunnyNet/src/SunnyProxy"
	"github.com/qtgolang/SunnyNet/src/crypto/tls"
	"github.com/qtgolang/SunnyNet/src/dns"
	"github.com/qtgolang/SunnyNet/src/public"
	"golang.org/x/sync/singleflight"
)

type Cache struct {
	mu    sync.RWMutex
	data  map[uint32]*entry
	ttl   time.Duration
	sunny *Sunny

	getSlots chan struct{}      // 控制Get并发
	sf       singleflight.Group // 控制同key的update只跑一次

	// janitor 可重启控制
	janMu       sync.Mutex
	janRunning  bool
	stopJanitor chan struct{}
	janWG       sync.WaitGroup
}

func newCache(sunny *Sunny) *Cache {
	cc := &Cache{
		data:     make(map[uint32]*entry),
		ttl:      time.Minute * 10,
		sunny:    sunny,
		getSlots: make(chan struct{}, 10),
	}
	return cc
}

func (c *Cache) StartJanitor() {
	c.janMu.Lock()
	if c.janRunning {
		c.janMu.Unlock()
		return
	}
	c.stopJanitor = make(chan struct{})
	c.janRunning = true
	stopCh := c.stopJanitor
	c.janWG.Add(1)
	c.janMu.Unlock()
	go func() {
		defer c.janWG.Done()
		ticker := time.NewTicker(c.ttl)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.cleanupExpired()
			case <-stopCh:
				return
			}
		}
	}()
}

func (c *Cache) StopJanitor() {
	c.janMu.Lock()
	if !c.janRunning {
		c.janMu.Unlock()
		return
	}
	stopCh := c.stopJanitor
	c.janRunning = false
	c.stopJanitor = nil
	c.janMu.Unlock()
	close(stopCh)
	c.janWG.Wait()
}

// entry 缓存内部条目
type entry struct {
	typeAt                 byte
	expireAt               time.Time
	netCert                *x509.Certificate //服务器响应证书
	cert                   *tls.Certificate
	serverName, targetAddr string
	DNSNames, nextProto    []string
	isLoop                 bool
}
type Result struct {
	Cert      *tls.Certificate
	DNSNames  []string
	NextProto []string
	NeedClose bool
	IsRules   bool
	At        byte
	HashCode  uint32
}

// Get 入口：最多10并发；同hashCode只允许一个update；返回结构体避免错位
func (c *Cache) Get(targetAddr *TargetInfo, serverName string, isLoopFunc func() bool) (Result, error) {
	hashCode := public.SumHashCode(targetAddr.String() + "/" + serverName)
	var res Result
	res.HashCode = hashCode
	res.NextProto = public.HTTP2NextProtos
	// 强制配置命中：不走缓存
	if in := c.getTlsConfig(targetAddr.String()); in != nil {
		res.Cert = in
		res.IsRules = c.sunny.tcpRules(serverName, targetAddr.Host)
		res.At = whoisHTTPS2
		return res, nil
	}
	if in := c.getTlsConfig(serverName); in != nil {
		res.Cert = in
		res.IsRules = c.sunny.tcpRules(serverName, targetAddr.Host)
		res.At = whoisHTTPS2
		return res, nil
	}
	nameKey := fmt.Sprintf("%s:%d", serverName, targetAddr.Port)
	if in := c.getTlsConfig(nameKey); in != nil {
		res.Cert = in
		res.IsRules = c.sunny.tcpRules(serverName, targetAddr.Host)
		res.At = whoisHTTPS2
		return res, nil
	}

	// 命中缓存（带过期判断）
	if e, ok := c.getValidEntry(hashCode); ok {
		return c.entryToResult(e, serverName, targetAddr.Host), nil
	}

	// 同一个hashCode只允许一个update
	key := strconv.FormatUint(uint64(hashCode), 10)
	v, err, _ := c.sf.Do(key, func() (any, error) {
		if e2, ok2 := c.getValidEntry(hashCode); ok2 {
			r := c.entryToResult(e2, serverName, targetAddr.Host)
			return r, nil
		}
		c.getSlots <- struct{}{}
		r, err := c.update(hashCode, targetAddr, serverName, isLoopFunc())
		<-c.getSlots
		return r, err
	})
	r := v.(Result)
	r.HashCode = hashCode
	return r, err
}

// getValidEntry：读取并校验过期；过期就删并返回 miss
func (c *Cache) getValidEntry(hashCode uint32) (*entry, bool) {
	now := time.Now()
	c.mu.RLock()
	e, ok := c.data[hashCode]
	c.mu.RUnlock()
	if !ok || e == nil {
		return nil, false
	}
	// 失效条件：expireAt - 1小时 < now
	if now.After(e.expireAt.Add(-time.Hour)) {
		// 双检删除，避免并发误删
		c.mu.Lock()
		e2, ok2 := c.data[hashCode]
		if ok2 && e2 != nil && now.After(e2.expireAt.Add(-time.Hour)) {
			delete(c.data, hashCode)
		}
		c.mu.Unlock()
		return nil, false
	}

	return e, true
}

// entryToResult：统一把缓存条目转成返回结果
func (c *Cache) entryToResult(e *entry, serverName, host string) Result {
	var r Result
	r.Cert = e.cert
	r.DNSNames = e.DNSNames
	r.At = e.typeAt

	if r.At == whoisHTTPS1 {
		r.NextProto = public.HTTP1NextProtos
	} else {
		r.NextProto = public.HTTP2NextProtos
	}

	r.IsRules = c.sunny.tcpRules(serverName, host, e.DNSNames...)
	return r
}

// update：真正生成/更新缓存
func (c *Cache) update(hashCode uint32, targetAddr *TargetInfo, serverName string, isLoop bool) (Result, error) {
	var (
		e    *entry
		name string
	)

	if !isLoop {
		cc := c.createCacheEntry(targetAddr.String(), serverName)

		// 不支持HTTPS：直接返回 needClose
		if cc.typeAt == whoisNoHTTPS {
			return Result{
				NeedClose: true,
				NextProto: public.HTTP1NextProtos,
				At:        whoisNoHTTPS,
			}, nil
		}

		e = &cc
		name = fmt.Sprintf("%s:%d", serverName, targetAddr.Port)
	} else {
		e = &entry{
			typeAt:     whoisHTTPS1,
			targetAddr: targetAddr.String(),
			serverName: serverName,
			isLoop:     true,
		}
		name = serverName
	}

	// whoisCache：你工程里已有（返回 cert、dnsNames、notAfter、error）
	cert, dnsNames, notAfter, err := c.whoisCache(e, name, targetAddr.String(), c.sunny.rootCa, c.sunny.rootKey)
	if err != nil {
		return Result{}, err
	}

	e.cert = cert
	e.DNSNames = dnsNames
	e.expireAt = notAfter

	// At/NextProto 决策统一放这里
	at := e.typeAt
	np := public.HTTP2NextProtos
	if at == whoisHTTPS1 {
		np = public.HTTP1NextProtos
	}
	e.nextProto = np

	isRules := c.sunny.tcpRules(serverName, targetAddr.Host, e.DNSNames...)

	// 写缓存
	c.mu.Lock()
	c.data[hashCode] = e
	c.mu.Unlock()

	return Result{
		Cert:      cert,
		DNSNames:  dnsNames,
		NextProto: np,
		NeedClose: false,
		IsRules:   isRules,
		At:        at,
	}, nil
}

func (c *Cache) updateType(hashCode uint32, Type byte) {
	c.mu.Lock()
	a := c.data[hashCode]
	if a != nil {
		if a.typeAt == whoisUndefined {
			a.typeAt = Type
		}
	}
	c.mu.Unlock()
	return
}

// cleanupExpired 扫描并清理过期项
func (c *Cache) cleanupExpired() {
	now := time.Now()

	c.mu.Lock()
	for k, e := range c.data {
		// 过期超过1小时才删除
		if now.After(e.expireAt.Add(time.Hour)) {
			delete(c.data, k)
		}
	}
	c.mu.Unlock()
}

func (c *Cache) createCacheEntry(targetAddr, serverName string) entry {
	res := entry{typeAt: whoisUndefined, targetAddr: targetAddr, serverName: serverName}

	// 建连
	conn := c.dialTarget(targetAddr)
	if conn == nil {
		return res
	}
	defer func() { _ = conn.Close() }()

	// 超时设置
	c.setConnDeadline(conn)

	// TLS 探测
	var hello *tls.ServerHelloMsg
	var certificate *x509.Certificate

	config := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         serverName,
	}
	config.GetConfigForServer = func(msg *tls.ServerHelloMsg) error {
		hello = msg
		return nil
	}
	config.VerifyServerCertificate = func(_certificate *x509.Certificate) error {
		certificate = _certificate
		return io.EOF
	}

	cc := tls.Client(conn, config)
	err := cc.Handshake()

	if hello == nil {
		if err != nil && strings.Contains(err.Error(), "close") {
			res.typeAt = whoisNoHTTPS
			return res
		}
		res.typeAt = whoisUndefined
		return res
	}

	isVer := whoisHTTPS1
	if hello.SupportedVersion == 772 {
		isVer = whoisHTTPS2
	}
	res.typeAt = byte(isVer)
	res.netCert = certificate
	return res
}

// dialTarget 负责根据你的策略建立连接：remoteDNS直接拨；否则解析优先IPv4最后IPv6
func (c *Cache) dialTarget(targetAddr string) net.Conn {
	// 如果是DNS解析服务器，直接本地拨号即可
	if dns.IsRemoteDnsServer() {
		conn, _ := c.sunny.proxy.DialWithTimeout("tcp", targetAddr, 2*time.Second, c.sunny.outRouterIP)
		return conn
	}

	proxyHost, proxyPort, err := net.SplitHostPort(targetAddr)
	if err != nil {
		return nil
	}

	// 目标本身就是IP，直接拨
	if ip := net.ParseIP(proxyHost); ip != nil {
		conn, _ := c.sunny.proxy.DialWithTimeout("tcp", SunnyProxy.FormatIP(ip, proxyPort), 3*time.Second, c.sunny.outRouterIP)
		return conn
	}

	// 先用缓存的首选IP试一次
	if first := dns.GetFirstIP(proxyHost, ""); first != nil {
		conn, _ := c.sunny.proxy.DialWithTimeout("tcp", SunnyProxy.FormatIP(first, proxyPort), 3*time.Second, c.sunny.outRouterIP)
		if conn != nil {
			return conn
		}
	}

	// 再做一次完整解析
	var (
		proxyUpstream string
		dial          func(network string, addr string, outRouterIP *net.TCPAddr) (net.Conn, error)
	)
	if c.sunny.proxy != nil {
		proxyUpstream = c.sunny.proxy.Host
		dial = c.sunny.proxy.Dial
	}

	ips, _ := dns.LookupIP(proxyHost, proxyUpstream, c.sunny.outRouterIP, dial)

	// 优先尝试IPv4
	for _, ip := range ips {
		if ip.To4() == nil {
			continue
		}
		conn, _ := c.sunny.proxy.DialWithTimeout("tcp", SunnyProxy.FormatIP(ip, proxyPort), 2*time.Second, c.sunny.outRouterIP)
		if conn != nil {
			dns.SetFirstIP(proxyHost, "", ip)
			return conn
		}
	}

	// 最后尝试IPv6
	for _, ip := range ips {
		if ip.To16() == nil || ip.To4() != nil {
			continue
		}
		conn, _ := c.sunny.proxy.DialWithTimeout("tcp", SunnyProxy.FormatIP(ip, proxyPort), 2*time.Second, c.sunny.outRouterIP)
		if conn != nil {
			dns.SetFirstIP(proxyHost, "", ip)
			return conn
		}
	}

	return nil
}

func (c *Cache) setConnDeadline(conn net.Conn) {
	// 原逻辑保持：有代理Host则3秒；没代理或Host空则1秒
	if c.sunny.proxy != nil && c.sunny.proxy.Host != "" {
		_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
		return
	}
	_ = conn.SetDeadline(time.Now().Add(1 * time.Second))
}

func (c *Cache) whoisCache(Entry *entry, serverName, host string, parent *x509.Certificate, priv *rsa.PrivateKey) (*tls.Certificate, []string, time.Time, error) {
	cc, d, n := c.createLocalCert(Entry, serverName, host, parent, priv)
	if cc != nil {
		return cc, d, n, nil
	}
	return nil, nil, n, _GetIpCertError
}
func (c *Cache) createLocalCert(e *entry, serverName, host string, parent *x509.Certificate, priv *rsa.PrivateKey) (*tls.Certificate, []string, time.Time) {
	hasSNI := serverName != "" && serverName != "null"
	if e != nil && e.netCert != nil {
		mHost := host
		if hasSNI {
			mHost = serverName
		}
		certByte, priByte, err := generatePem(e.netCert, mHost, parent, priv)
		if err == nil {
			cert, err2 := tls.X509KeyPair(certByte, priByte)
			if err2 == nil {
				DNSNames := make([]string, 0, len(e.netCert.DNSNames)+len(e.netCert.IPAddresses))
				DNSNames = append(DNSNames, e.netCert.DNSNames...)
				for _, ip := range e.netCert.IPAddresses {
					DNSNames = append(DNSNames, ip.String())
				}
				return &cert, DNSNames, e.netCert.NotAfter
			}
		}
	}

	var (
		mHost string
		err   error
	)

	if hasSNI {
		mHost, _, err = public.SplitHostPort(serverName)
	} else {
		if !strings.HasSuffix(host, ":853") {
			cert, dns, not := c.tryCreateNetCert(e, host, parent, priv)
			if cert != nil {
				return cert, dns, not
			}
		}
		mHost, _, err = public.SplitHostPort(host)
	}
	if err != nil {
		return nil, nil, time.Time{}
	}
	certByte, priByte, dns, not, err := generatePemTemp(mHost, parent, priv)
	if err != nil {
		return nil, nil, time.Time{}
	}
	cert, err := tls.X509KeyPair(certByte, priByte)
	if err != nil {
		return nil, nil, time.Time{}
	}
	return &cert, dns, not
}

func (c *Cache) tryCreateNetCert(e *entry, host string, parent *x509.Certificate, priv *rsa.PrivateKey) (*tls.Certificate, []string, time.Time) {
	cert, DNSNames, not, _ := c.createNetCert(e, host, parent, priv)
	if cert == nil {
		return nil, nil, time.Time{}
	}
	return cert, DNSNames, not
}

func (c *Cache) createNetCert(Entry *entry, host string, parent *x509.Certificate, priv *rsa.PrivateKey) (*tls.Certificate, []string, time.Time, error) {
	mHost, _, err := public.SplitHostPort(host)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	if ip := net.ParseIP(mHost); ip != nil {
		var rr *x509.Certificate
		if Entry.netCert != nil {
			rr = Entry.netCert
		}
		if rr == nil {
			for i := 0; i < 5; i++ {
				rr, err = c.getIpAddressHost(host)
				if rr != nil {
					break
				}
			}
		}
		if rr == nil {
			return nil, nil, time.Time{}, _GetIpCertError
		}
		not := time.Now().AddDate(0, 0, 365)
		certByte, priByte, er := generatePem(rr, mHost, parent, priv)
		if er != nil {
			return nil, nil, not, er
		}
		certificate, er := tls.X509KeyPair(certByte, priByte)
		if er != nil {
			return nil, nil, not, er
		}
		DNSNames := rr.DNSNames
		for _, v := range rr.IPAddresses {
			DNSNames = append(DNSNames, v.String())
		}
		return &certificate, DNSNames, not, nil
	}
	return nil, nil, time.Time{}, _ParseIPError
}
func (c *Cache) getTlsConfig(host string) *tls.Certificate {
	in := HttpCertificate.GetTlsConfig(host, public.CertificateRequestManagerRulesReceive)
	if in != nil {
		if len(in.Certificates) > 0 {
			return &in.Certificates[0]
		}
	}
	return nil
}
func (c *Cache) getIpAddressHost(ipAddress string) (*x509.Certificate, error) {
	config := &tls.Config{InsecureSkipVerify: true}
	var x *x509.Certificate
	config.VerifyServerCertificate = func(certificate *x509.Certificate) error {
		x = certificate
		return io.EOF
	}
	var conn net.Conn
	var err error
	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()
	conn, err = c.sunny.proxy.Dial("tcp", ipAddress, c.sunny.outRouterIP)
	if err != nil {
		return nil, err
	}
	t := tls.Client(conn, config)
	err = t.Handshake()
	if x != nil {
		err = nil
	}
	return x, err
}

func generatePem(template *x509.Certificate, host string, parent *x509.Certificate, priv *rsa.PrivateKey) ([]byte, []byte, error) {
	mHost, _, err := net.SplitHostPort(host)
	if err != nil {
		mHost = host
	}
	template1 := x509.Certificate{
		SerialNumber:                template.SerialNumber,                // 序列号，CA 颁发的唯一序列号，通常为随机生成
		Subject:                     template.Subject,                     // 证书主题，包含持有者的信息（国家、组织等）
		NotBefore:                   template.NotBefore,                   // 证书开始生效时间
		NotAfter:                    template.NotAfter,                    // 证书到期时间
		KeyUsage:                    template.KeyUsage,                    // 密钥用法，指明证书可用于的操作（如签名、加密等）
		ExtKeyUsage:                 template.ExtKeyUsage,                 // 扩展密钥用法，指明证书的额外用途（如客户端认证、服务器认证等）
		EmailAddresses:              template.EmailAddresses,              // 证书持有者的电子邮件地址
		IPAddresses:                 template.IPAddresses,                 // 证书包含的 IP 地址列表
		DNSNames:                    template.DNSNames,                    // 证书关联的 DNS 域名列表
		Issuer:                      template.Issuer,                      // 证书颁发者的信息
		IssuingCertificateURL:       template.IssuingCertificateURL,       // 颁发者证书的 URL
		BasicConstraintsValid:       template.BasicConstraintsValid,       // 基础约束是否有效
		IsCA:                        template.IsCA,                        // 标识证书是否为证书颁发机构（CA）证书
		AuthorityKeyId:              template.AuthorityKeyId,              // CA 密钥标识符
		UnknownExtKeyUsage:          template.UnknownExtKeyUsage,          // 未知的扩展密钥用途列表
		ExtraExtensions:             template.ExtraExtensions,             // 额外的 X.509 扩展字段
		PermittedDNSDomainsCritical: template.PermittedDNSDomainsCritical, // 是否为关键的允许 DNS 域名
		PermittedDNSDomains:         template.PermittedDNSDomains,         // 允许的 DNS 域名列表
		PolicyIdentifiers:           template.PolicyIdentifiers,           // 策略标识符的列表
		MaxPathLen:                  template.MaxPathLen,                  // 最大路径长度，限制证书链的深度
		MaxPathLenZero:              template.MaxPathLenZero,              // 最大路径长度是否可以为零
		SubjectKeyId:                template.SubjectKeyId,                // 证书持有者密钥的标识符
	}
	if ip := net.ParseIP(mHost); ip != nil {
		template1.IPAddresses = append(template1.IPAddresses, ip)
	} else {
		template1.DNSNames = append(template1.DNSNames, mHost)
	}
	cer, err := x509.CreateCertificate(rand.Reader, &template1, parent, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}
	return pem.EncodeToMemory(&pem.Block{ // 证书
			Type:  "CERTIFICATE",
			Bytes: cer,
		}), pem.EncodeToMemory(&pem.Block{ // 私钥
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		}), err
}
func generatePemTemp(mHost string, parent *x509.Certificate, priv *rsa.PrivateKey) ([]byte, []byte, []string, time.Time, error) {
	serialNumber, _ := rand.Int(rand.Reader, public.MaxBig)
	not := time.Now().AddDate(0, 0, 365)
	template := x509.Certificate{
		SerialNumber: serialNumber, // SerialNumber 是 CA 颁布的唯一序列号，在此使用一个大随机数来代表它
		Subject: pkix.Name{ //Name代表一个X.509识别名。只包含识别名的公共属性，额外的属性被忽略。
			CommonName: mHost,
		},
		NotBefore:      time.Now().AddDate(0, 0, -1),
		NotAfter:       not,
		KeyUsage:       x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature, //KeyUsage 与 ExtKeyUsage 用来表明该证书是用来做服务器认证的
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},               // 密钥扩展用途的序列
		EmailAddresses: []string{"forward.nice.cp@gmail.com"},
	}
	{
		if ip := net.ParseIP(mHost); ip != nil {
			template.IPAddresses = []net.IP{ip}
			template.DNSNames = []string{ip.String()}
		} else if _, _, ip = parseIPv6Address(mHost); ip != nil {
			template.IPAddresses = []net.IP{ip}
			template.DNSNames = []string{ip.String()}
		} else {
			template.DNSNames = []string{mHost}
		}
	}

	DNSNames := template.DNSNames
	cer, err := x509.CreateCertificate(rand.Reader, &template, parent, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, nil, not, err
	}
	return pem.EncodeToMemory(&pem.Block{ // 证书
			Type:  "CERTIFICATE",
			Bytes: cer,
		}), pem.EncodeToMemory(&pem.Block{ // 私钥
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		}), DNSNames, not, err
}

var _GetIpCertError = fmt.Errorf("no success Get Certificate")

var _ParseIPError = errors.New("Not an IP address ")

const whoisUndefined = 0
const whoisNoHTTPS = 1
const whoisHTTPS1 = 2
const whoisHTTPS2 = 3
