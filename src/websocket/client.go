// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"bytes"
	"context"
	"errors"
	"github.com/qtgolang/SunnyNet/src/SunnyProxy"
	"github.com/qtgolang/SunnyNet/src/crypto/tls"
	"github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/http/httptrace"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"strings"
	"time"
)

// ErrBadHandshake is returned when the server response to opening handshake is
// invalid.
var ErrBadHandshake = errors.New("websocket: bad handshake")

var errInvalidCompression = errors.New("websocket: invalid compression negotiation")

// NewClient creates a new client connection using the given net connection.
// The URL u specifies the host and request URI. Use requestHeader to specify
// the origin (Origin), subprotocols (Sec-WebSocket-Protocol) and cookies
// (Cookie). Use the response.Header to get the selected subprotocol
// (Sec-WebSocket-Protocol) and cookies (Set-Cookie).
//
// If the WebSocket handshake fails, ErrBadHandshake is returned along with a
// non-nil *net.Response so that callers can handle redirects, authentication,
// etc.
//
// Deprecated: Use Dialer instead.
func NewClient(netConn net.Conn, u *url.URL, requestHeader http.Header, readBufSize, writeBufSize int, OutRouterIP *net.TCPAddr) (c *Conn, response *http.Response, serverIp string, err error) {
	d := Dialer{
		ReadBufferSize:  readBufSize,
		WriteBufferSize: writeBufSize,
		NetDial: func(net, addr string) (net.Conn, error) {
			return netConn, nil
		},
	}
	return d.Dial(u.String(), requestHeader, nil, OutRouterIP)
}

// A Dialer contains options for connecting to WebSocket server.
type Dialer struct {
	// NetDial specifies the dial function for creating TCP connections. If
	// NetDial is nil, net.Dial is used.
	NetDial func(network, addr string) (net.Conn, error)

	// NetDialContext specifies the dial function for creating TCP connections. If
	// NetDialContext is nil, net.DialContext is used.
	NetDialContext func(ctx context.Context, network, addr string) (net.Conn, error)

	// Proxy specifies a function to return a proxy for a given
	// Request. If the function returns a non-nil error, the
	// request is aborted with the provided error.
	// If Proxy is nil or returns a nil *URL, no proxy is used.

	// TLSClientConfig specifies the TLS configuration to use with tls.Client.
	// If nil, the default configuration is used.
	TLSClientConfig *tls.Config

	// HandshakeTimeout specifies the duration for the handshake to complete.
	HandshakeTimeout time.Duration

	// ReadBufferSize and WriteBufferSize specify I/O buffer sizes in bytes. If a buffer
	// size is zero, then a useful default size is used. The I/O buffer sizes
	// do not limit the size of the messages that can be sent or received.
	ReadBufferSize, WriteBufferSize int

	// WriteBufferPool is a pool of buffers for write operations. If the value
	// is not set, then write buffers are allocated to the connection for the
	// lifetime of the connection.
	//
	// A pool is most useful when the application has a modest volume of writes
	// across a large number of connections.
	//
	// Applications should use a single pool for each unique value of
	// WriteBufferSize.
	WriteBufferPool BufferPool

	// Subprotocols specifies the client's requested subprotocols.
	Subprotocols []string

	// EnableCompression specifies if the client should attempt to negotiate
	// per message compression (RFC 7692). Setting this value to true does not
	// guarantee that compression will be supported. Currently only "no context
	// takeover" modes are supported.
	EnableCompression bool

	// Jar specifies the cookie jar.
	// If Jar is nil, cookies are not sent in requests and ignored
	// in responses.
	Jar      http.CookieJar
	ProxyUrl *SunnyProxy.Proxy
}

func (d *Dialer) Dial(urlStr string, requestHeader http.Header, ProxyUrl *SunnyProxy.Proxy, OutRouterIP *net.TCPAddr, outTime ...int) (*Conn, *http.Response, string, error) {
	d.ProxyUrl = ProxyUrl
	t := 30 * 1000
	if len(outTime) > 0 {
		t = outTime[0]
	}
	return d.DialContext(context.Background(), urlStr, requestHeader, OutRouterIP, t)
}

var errMalformedURL = errors.New("malformed ws or wss URL")

func hostPortNoPort(u *url.URL) (hostPort, hostNoPort string) {
	hostPort = u.Host
	hostNoPort = u.Host
	if i := strings.LastIndex(u.Host, ":"); i > strings.LastIndex(u.Host, "]") {
		hostNoPort = hostNoPort[:i]
	} else {
		switch u.Scheme {
		case "wss":
			hostPort += ":443"
		case "https":
			hostPort += ":443"
		default:
			hostPort += ":80"
		}
	}
	return hostPort, hostNoPort
}

// DefaultDialer is a dialer with all fields set to the default values.
var DefaultDialer = &Dialer{
	HandshakeTimeout: 45 * time.Second,
}

// nilDialer is dialer to use when receiver is nil.
var nilDialer = *DefaultDialer

// DialContext creates a new client connection. Use requestHeader to specify the
// origin (Origin), subprotocols (Sec-WebSocket-Protocol) and cookies (Cookie).
// Use the response.Header to get the selected subprotocol
// (Sec-WebSocket-Protocol) and cookies (Set-Cookie).
//
// The context will be used in the request and in the Dialer.
//
// If the WebSocket handshake fails, ErrBadHandshake is returned along with a
// non-nil *net.Response so that callers can handle redirects, authentication,
// etcetera. The response body may not contain the entire response and does not
// need to be closed by the application.
func (d *Dialer) DialContext(ctx context.Context, urlStr string, requestHeader http.Header, OutRouterIP *net.TCPAddr, outTime ...int) (*Conn, *http.Response, string, error) {
	if d == nil {
		d = &nilDialer
	}
	challengeKey, err := generateChallengeKey()
	if err != nil {
		return nil, nil, "", err
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, nil, "", err
	}

	switch u.Scheme {
	case "ws":
		u.Scheme = "net"
	case "wss":
		u.Scheme = "https"
	case "http":
		u.Scheme = "net"
	case "https":
		u.Scheme = "https"
	default:
		return nil, nil, "", errMalformedURL
	}

	if u.User != nil {
		// User name and password are not allowed in websocket URIs.
		return nil, nil, "", errMalformedURL
	}

	req := &http.Request{
		Method:     "GET",
		URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       u.Host,
	}
	req = req.WithContext(ctx)

	// Set the cookies present in the cookie jar of the dialer
	if d.Jar != nil {
		for _, cookie := range d.Jar.Cookies(u) {
			req.AddCookie(cookie)
		}
	}
	req.Header = requestHeader
	// Set the request headers using the capitalization for names and values in
	// RFC examples. Although the capitalization shouldn't matter, there are
	// servers that depend on it. The Header.Set method is not used because the
	// method canonicalizes the header names.
	req.Header["Upgrade"] = []string{"websocket"}
	req.Header["Connection"] = []string{"Upgrade"}
	req.Header["Sec-WebSocket-Key"] = []string{challengeKey}
	req.Header["Sec-WebSocket-Version"] = []string{"13"}
	if len(d.Subprotocols) > 0 {
		req.Header["Sec-WebSocket-Protocol"] = []string{strings.Join(d.Subprotocols, ", ")}
	}
	if d.EnableCompression {
		req.Header["Sec-WebSocket-Extensions"] = []string{"permessage-deflate; server_no_context_takeover; client_no_context_takeover"}
	}

	if d.HandshakeTimeout != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, d.HandshakeTimeout)
		defer cancel()
	}

	_outTime := 30000
	if len(outTime) > 0 {
		_outTime = outTime[0]
		if _outTime < 1 {
			_outTime = 30000
		}
	}
	hostPort, hostNoPort := hostPortNoPort(u)
	trace := httptrace.ContextClientTrace(ctx)
	if trace != nil && trace.GetConn != nil {
		trace.GetConn(hostPort)
	}
	var proxy *SunnyProxy.Proxy
	if d.ProxyUrl == nil {
		proxy = new(SunnyProxy.Proxy)
	} else {
		proxy = d.ProxyUrl.Clone()
	}
	netConn, err := proxy.DialWithTimeout("tcp", hostPort, time.Duration(_outTime)*time.Millisecond, OutRouterIP)
	if trace != nil && trace.GotConn != nil {
		trace.GotConn(httptrace.GotConnInfo{
			Conn: netConn,
		})
	}
	if err != nil {
		return nil, nil, "", err
	}
	defer func() {
		if netConn != nil {
			_ = netConn.Close()
		}
	}()

	if u.Scheme == "https" {
		cfg := cloneTLSConfig(d.TLSClientConfig)
		if cfg.ServerName == "" {
			cfg.ServerName = hostNoPort
		}
		tlsConn := tls.Client(netConn, cfg)
		netConn = tlsConn

		var err error
		if trace != nil {

			err = doHandshakeWithTrace(trace, tlsConn, cfg)

		} else {

			err = doHandshake(tlsConn, cfg)

		}

		if err != nil {
			//fmt.Println("err1:=", err)
			return nil, nil, "", err
		}
	}

	conn := newConn(netConn, false, d.ReadBufferSize, d.WriteBufferSize, d.WriteBufferPool, nil, nil)

	if err = req.Write(netConn); err != nil {
		//fmt.Println("err2:=", err)
		return nil, nil, "", err
	}

	_ = conn.conn.SetDeadline(time.Now().Add(time.Duration(_outTime) * time.Millisecond))
	defer func() {
		if conn != nil {
			if conn.conn != nil {
				_ = conn.conn.SetDeadline(time.Time{})
			}
		}
	}()

	if trace != nil && trace.GotFirstResponseByte != nil {
		if peek, err := conn.br.Peek(1); err == nil && len(peek) == 1 {
			trace.GotFirstResponseByte()
		}
	}

	resp, err := http.ReadResponse(conn.br, req)
	if err != nil {
		//fmt.Println("err3:=", err)
		return nil, nil, "", err
	}

	if d.Jar != nil {
		if rc := resp.Cookies(); len(rc) > 0 {
			d.Jar.SetCookies(u, rc)
		}
	}

	if resp.StatusCode != 101 ||
		!strings.EqualFold(resp.Header.Get("Upgrade"), "websocket") ||
		!strings.EqualFold(resp.Header.Get("Connection"), "upgrade") ||
		resp.Header.Get("Sec-Websocket-Accept") == "" {
		// Before closing the network connection on return from this
		// function, slurp up some of the response to aid application
		// debugging.
		buf := make([]byte, 1024)
		n, _ := io.ReadFull(resp.Body, buf)
		resp.Body = ioutil.NopCloser(bytes.NewReader(buf[:n]))
		return nil, resp, "", ErrBadHandshake
	}
	extensions := ParseExtensions(resp.Header)
	for _, ext := range extensions {
		if ext[""] != "permessage-deflate" {
			continue
		}
		//conn.newCompressionWriter = compressNoContextTakeover
		conn.newDecompressionReader = decompressNoContextTakeover
		conn.WindowReader.Window_Size_Max = 1 << 15
		break
	}

	resp.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))
	conn.subprotocol = resp.Header.Get("Sec-Websocket-Protocol")

	netConn.SetDeadline(time.Time{})
	netConn = nil // to avoid close in defer.
	return conn, resp, proxy.DialAddr, nil
}

const SunnyNetServerIpTags = "ServerAddr"

// ConnDialContext 自己改写的 返回Websocket.Conn 和httpResponse
func (d *Dialer) ConnDialContext(request *http.Request, ProxyUrl *SunnyProxy.Proxy, OutRouterIP *net.TCPAddr) (*Conn, *http.Response, error) {
	if d == nil {
		d = &nilDialer
	}
	d.ProxyUrl = ProxyUrl
	ctx := context.Background()
	req := &http.Request{
		Method:     "GET",
		URL:        request.URL,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       request.URL.Host,
	}
	req = req.WithContext(ctx)
	challengeKey := ""
	// Set the cookies present in the cookie jar of the dialer
	if d.Jar != nil {
		for _, cookie := range d.Jar.Cookies(request.URL) {
			req.AddCookie(cookie)
		}
	}
	if len(d.Subprotocols) > 0 {
		req.Header["Sec-WebSocket-Protocol"] = []string{strings.Join(d.Subprotocols, ", ")}
	}
	for k, vs := range request.Header {
		switch {
		case k == "Host":
			if len(vs) > 0 {
				req.Host = vs[0]
			}
			continue
		case strings.EqualFold(k, "Sec-Websocket-Key"):
			req.Header[k] = vs
			challengeKey = vs[0]
			continue
		default:
			kk := k + ""
			kk = strings.ReplaceAll(kk, "-Websocket-", "-WebSocket-")
			req.Header[kk] = vs
		}
	}
	if d.HandshakeTimeout != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, d.HandshakeTimeout)
		defer cancel()
	}
	hostPort, hostNoPort := hostPortNoPort(request.URL)
	trace := httptrace.ContextClientTrace(ctx)
	if trace != nil && trace.GetConn != nil {
		trace.GetConn(hostPort)
	}

	var proxy *SunnyProxy.Proxy
	if d.ProxyUrl == nil {
		proxy = new(SunnyProxy.Proxy)
	} else {
		proxy = d.ProxyUrl.Clone()
	}
	netConn, err := proxy.Dial("tcp", hostPort, OutRouterIP)
	defer func() {
		address, p, _ := net.SplitHostPort(proxy.DialAddr)
		ip := net.ParseIP(address)
		if ip == nil {
			request.SetContext(SunnyNetServerIpTags, proxy.DialAddr)
		} else {
			request.SetContext(SunnyNetServerIpTags, SunnyProxy.FormatIP(ip, p))
		}
	}()

	if trace != nil && trace.GotConn != nil {
		trace.GotConn(httptrace.GotConnInfo{
			Conn: netConn,
		})
	}
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		if netConn != nil {
			netConn.Close()
		}
	}()

	if request.URL.Scheme == "https" {
		cfg := cloneTLSConfig(d.TLSClientConfig)
		if cfg.ServerName == "" {
			cfg.ServerName = hostNoPort
		}
		tlsConn := tls.Client(netConn, cfg)
		netConn = tlsConn

		var err error
		if trace != nil {
			err = doHandshakeWithTrace(trace, tlsConn, cfg)
		} else {
			err = doHandshake(tlsConn, cfg)
		}

		if err != nil {
			return nil, nil, err
		}
	}

	conn := newConn(netConn, false, d.ReadBufferSize, d.WriteBufferSize, d.WriteBufferPool, nil, nil)

	if err := req.Write(netConn); err != nil {
		return nil, nil, err
	}

	if trace != nil && trace.GotFirstResponseByte != nil {
		if peek, err := conn.br.Peek(1); err == nil && len(peek) == 1 {
			trace.GotFirstResponseByte()
		}
	}

	resp, err := http.ReadResponse(conn.br, req)
	if err != nil {
		return nil, nil, err
	}

	if d.Jar != nil {
		if rc := resp.Cookies(); len(rc) > 0 {
			d.Jar.SetCookies(request.URL, rc)
		}
	}
	var _Upgrade string
	var _Connection string
	for k, v := range resp.Header {
		if strings.EqualFold(k, "Upgrade") {
			if len(v) > 0 {
				_Upgrade = v[0]

			}
			continue
		}
		if strings.EqualFold(k, "Connection") {
			if len(v) > 0 {
				_Connection = v[0]
			}
			continue
		}
		if _Upgrade != "" && _Connection != "" {
			break
		}
	}
	ok1 := !strings.EqualFold(_Upgrade, "websocket") && !strings.EqualFold(_Upgrade, "upgrade")
	ok2 := !strings.EqualFold(_Connection, "websocket") && !strings.EqualFold(_Connection, "upgrade")
	aa := computeAcceptKey(challengeKey)
	bb := resp.Header.Get("Sec-Websocket-Accept")
	if bb == "" {
		bb = resp.Header.Get("Sec-WebSocket-Accept")
		if bb == "" {
			for k, v := range resp.Header {
				if strings.EqualFold(k, "Sec-Websocket-Accept") {
					if len(v) > 0 {
						bb = v[0]
					}
					break
				}
			}
		}
	}
	if resp.StatusCode != 101 || ok1 || ok2 ||
		bb != aa {
		// Before closing the network connection on return from this
		// function, slurp up some of the response to aid application
		// debugging.
		buf := make([]byte, 1024)
		n, _ := io.ReadFull(resp.Body, buf)
		resp.Body = ioutil.NopCloser(bytes.NewReader(buf[:n]))
		return nil, resp, ErrBadHandshake
	}
	extensions := ParseExtensions(resp.Header)
	for _, ext := range extensions {
		if ext[""] != "permessage-deflate" {
			continue
		}
		conn.newDecompressionReader = decompressNoContextTakeover
		conn.WindowReader.Window_Size_Max = 1 << 15
		break
	}
	resp.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))
	conn.subprotocol = resp.Header.Get("Sec-Websocket-Protocol")
	netConn.SetDeadline(time.Time{})
	netConn = nil // to avoid close in defer.
	return conn, resp, nil
}
func doHandshake(tlsConn *tls.Conn, cfg *tls.Config) error {
	if err := tlsConn.Handshake(); err != nil {
		return err
	}
	if !cfg.InsecureSkipVerify {
		if err := tlsConn.VerifyHostname(cfg.ServerName); err != nil {
			return err
		}
	}
	return nil
}
