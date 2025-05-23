// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import "C"
import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/qtgolang/SunnyNet/src/ReadWriteObject"
	"github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/internal/textproto"
	"io"
	"net"
	"net/url"
	"strings"
	"time"
)

// HandshakeError describes an error with the handshake from the peer.
type HandshakeError struct {
	message string
}

func (e HandshakeError) Error() string { return e.message }

// Upgrader specifies parameters for upgrading an HTTP connection to a
// WebSocket connection.
type Upgrader struct {
	// HandshakeTimeout specifies the duration for the handshake to complete.
	HandshakeTimeout time.Duration

	// ReadBufferSize and WriteBufferSize specify I/O buffer sizes in bytes. If a buffer
	// size is zero, then buffers allocated by the HTTP server are used. The
	// I/O buffer sizes do not limit the size of the messages that can be sent
	// or received.
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

	// Subprotocols specifies the server's supported protocols in order of
	// preference. If this field is not nil, then the Upgrade method negotiates a
	// subprotocol by selecting the first match in this list with a protocol
	// requested by the client. If there's no match, then no protocol is
	// negotiated (the Sec-Websocket-Protocol header is not included in the
	// handshake response).
	Subprotocols []string

	// Error specifies the function for generating HTTP error responses. If Error
	// is nil, then net.Error is used to generate the HTTP response.
	Error func(w http.ResponseWriter, r *http.Request, status int, reason error)

	// CheckOrigin returns true if the request Origin header is acceptable. If
	// CheckOrigin is nil, then a safe default is used: return false if the
	// Origin request header is present and the origin host is not equal to
	// request Host header.
	//
	// A CheckOrigin function should carefully validate the request origin to
	// prevent cross-site request forgery.
	CheckOrigin func(r *http.Request) bool

	// EnableCompression specify if the server should attempt to negotiate per
	// message compression (RFC 7692). Setting this value to true does not
	// guarantee that compression will be supported. Currently only "no context
	// takeover" modes are supported.
	EnableCompression bool
}

func (u *Upgrader) returnError(w http.ResponseWriter, r *http.Request, status int, reason string) (*Conn, error) {
	err := HandshakeError{reason}
	if u.Error != nil {
		u.Error(w, r, status, err)
	} else {
		w.Header().Set("Sec-Websocket-Version", "13")
		http.Error(w, http.StatusText(status), status)
	}
	return nil, err
}

// checkSameOrigin returns true if the origin is not set or is equal to the request host.
func checkSameOrigin(r *http.Request) bool {
	origin := r.Header["Origin"]
	if len(origin) == 0 {
		return true
	}
	u, err := url.Parse(origin[0])
	if err != nil {
		return false
	}
	return equalASCIIFold(u.Host, r.Host)
}

func (u *Upgrader) selectSubprotocol(r *http.Request, responseHeader http.Header) string {
	if u.Subprotocols != nil {
		clientProtocols := Subprotocols(r)
		for _, serverProtocol := range u.Subprotocols {
			for _, clientProtocol := range clientProtocols {
				if clientProtocol == serverProtocol {
					return clientProtocol
				}
			}
		}
	} else if responseHeader != nil {
		return responseHeader.Get("Sec-Websocket-Protocol")
	}
	return ""
}

// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
//
// The responseHeader is included in the response to the client's upgrade
// request. Use the responseHeader to specify cookies (Set-Cookie) and the
// application negotiated subprotocol (Sec-WebSocket-Protocol).
//
// If the upgrade fails, then Upgrade replies to the client with an HTTP error
// response.
func (u *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header, c_ net.Conn, b_ *bufio.ReadWriter) (*Conn, error) {
	const badHandshake = "websocket: the client is not using the websocket protocol: "
	for name, values := range r.Header {
		Mikey := textproto.CanonicalMIMEHeaderKey(name)
		if Mikey != name {
			r.Header.Del(name)
			r.Header[Mikey] = values
		}
	}
	for name, values := range responseHeader {
		Mikey := textproto.CanonicalMIMEHeaderKey(name)
		if Mikey != name {
			responseHeader.Del(name)
			responseHeader[Mikey] = values
		}
	}
	if !tokenListContainsValue(r.Header, "Connection", "upgrade") {
		return u.returnError(w, r, http.StatusBadRequest, badHandshake+"'upgrade' token not found in 'Connection' header")
	}

	if !tokenListContainsValue(r.Header, "Upgrade", "websocket") {
		return u.returnError(w, r, http.StatusBadRequest, badHandshake+"'websocket' token not found in 'Upgrade' header")
	}

	if r.Method != "GET" {
		return u.returnError(w, r, http.StatusMethodNotAllowed, badHandshake+"request method is not GET")
	}

	if !tokenListContainsValue(r.Header, textproto.CanonicalMIMEHeaderKey("Sec-WebSocket-Version"), "13") {
		return u.returnError(w, r, http.StatusBadRequest, "websocket: unsupported version: 13 not found in 'Sec-Websocket-Version' header")
	}

	if _, ok := responseHeader[textproto.CanonicalMIMEHeaderKey("Sec-Websocket-Extensions")]; ok {
		return u.returnError(w, r, http.StatusInternalServerError, "websocket: application specific 'Sec-WebSocket-Extensions' headers are unsupported")
	}

	checkOrigin := u.CheckOrigin
	if checkOrigin == nil {
		checkOrigin = checkSameOrigin
	}
	if !checkOrigin(r) {
		return u.returnError(w, r, http.StatusForbidden, "websocket: request origin not allowed by Upgrader.CheckOrigin")
	}
	challengeKey := r.Header.Get(textproto.CanonicalMIMEHeaderKey("Sec-WebSocket-Key"))
	if challengeKey == "" {
		return u.returnError(w, r, http.StatusBadRequest, "websocket: not a websocket handshake: 'Sec-WebSocket-Key' header is missing or blank")
	}

	protocol := u.selectSubprotocol(r, responseHeader)

	// Negotiate PMCE
	var compress bool
	if u.EnableCompression {
		for _, ext := range ParseExtensions(r.Header) {
			if ext[""] != "permessage-deflate" {
				continue
			}
			compress = true
			break
		}
	}
	netConn := c_
	brw := b_
	if brw.Reader.Buffered() > 0 {
		netConn.Close()
		return nil, errors.New("websocket: client sent data before handshake is complete")
	}

	var br *bufio.Reader
	if u.ReadBufferSize == 0 && bufioReaderSize(netConn, brw.Reader) > 256 {
		// Reuse hijacked buffered reader as connection reader.
		br = brw.Reader
	}

	buf := bufioWriterBuffer(netConn, brw.Writer)

	var writeBuf []byte
	if u.WriteBufferPool == nil && u.WriteBufferSize == 0 && len(buf) >= maxFrameHeaderSize+256 {
		// Reuse hijacked write buffer as connection buffer.
		writeBuf = buf
	}

	c := newConn(netConn, true, u.ReadBufferSize, u.WriteBufferSize, u.WriteBufferPool, br, writeBuf)
	c.subprotocol = protocol

	if compress {
		c.newCompressionWriter = compressNoContextTakeover
		c.newDecompressionReader = decompressNoContextTakeover
	}

	// Use larger of hijacked buffer and connection write buffer for header.
	p := buf
	if len(c.writeBuf) > len(p) {
		p = c.writeBuf
	}
	p = p[:0]

	p = append(p, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: "...)
	p = append(p, computeAcceptKey(challengeKey)...)
	p = append(p, "\r\n"...)
	if c.subprotocol != "" {
		p = append(p, "Sec-WebSocket-Protocol: "...)
		p = append(p, c.subprotocol...)
		p = append(p, "\r\n"...)
	}
	if compress {
		p = append(p, "Sec-WebSocket-Extensions: permessage-deflate; server_no_context_takeover; client_no_context_takeover\r\n"...)
	}
	for k, vs := range responseHeader {
		if k == "Sec-Websocket-Protocol" {
			continue
		}
		for _, v := range vs {
			p = append(p, k...)
			p = append(p, ": "...)
			for i := 0; i < len(v); i++ {
				b := v[i]
				if b <= 31 {
					// prevent response splitting.
					b = ' '
				}
				p = append(p, b)
			}
			p = append(p, "\r\n"...)
		}
	}
	p = append(p, "\r\n"...)

	// Clear deadlines set by HTTP server.
	netConn.SetDeadline(time.Time{})

	if u.HandshakeTimeout > 0 {
		netConn.SetWriteDeadline(time.Now().Add(u.HandshakeTimeout))
	}
	if _, err := netConn.Write(p); err != nil {
		netConn.Close()
		return nil, err
	}
	if u.HandshakeTimeout > 0 {
		netConn.SetWriteDeadline(time.Time{})
	}

	return c, nil
}

func (u *Upgrader) UpgradeSunnyNetWebsocket(w *ReadWriteObject.ReadWriteObject, r *http.Request, responseHeader http.Header, c_ net.Conn, b_ *bufio.ReadWriter) (*Conn, error) {
	const badHandshake = "websocket: the client is not using the websocket protocol: "
	for name, values := range r.Header {
		Mikey := textproto.CanonicalMIMEHeaderKey(name)
		if Mikey != name {
			r.Header.Del(name)
			r.Header[Mikey] = values
		}
	}
	for name, values := range responseHeader {
		Mikey := textproto.CanonicalMIMEHeaderKey(name)
		if Mikey != name {
			responseHeader.Del(name)
			responseHeader[Mikey] = values
		}
	}
	if !tokenListContainsValue(r.Header, "Connection", "upgrade") {
		s := badHandshake + "'upgrade' token not found in 'Connection' header"
		_, _ = w.WriteString(fmt.Sprintf("HTTP/1.1 %d OK\r\nCache-Control: no-cache, must-revalidate\r\nPragma: no-cache\r\nContent-Length: %d\r\n\r\n%s", http.StatusBadRequest, len(s), s))
		return nil, HandshakeError{s}
	}

	if !tokenListContainsValue(r.Header, "Upgrade", "websocket") {
		s := badHandshake + "'websocket' token not found in 'Upgrade' header"
		_, _ = w.WriteString(fmt.Sprintf("HTTP/1.1 %d OK\r\nCache-Control: no-cache, must-revalidate\r\nPragma: no-cache\r\nContent-Length: %d\r\n\r\n%s", http.StatusBadRequest, len(s), s))
		return nil, HandshakeError{s}
	}

	if r.Method != "GET" {
		s := badHandshake + "request method is not GET"
		_, _ = w.WriteString(fmt.Sprintf("HTTP/1.1 %d OK\r\nCache-Control: no-cache, must-revalidate\r\nPragma: no-cache\r\nContent-Length: %d\r\n\r\n%s", http.StatusMethodNotAllowed, len(s), s))
		return nil, HandshakeError{s}
	}

	if !tokenListContainsValue(r.Header, textproto.CanonicalMIMEHeaderKey("Sec-WebSocket-Version"), "13") {
		s := "websocket: unsupported version: 13 not found in 'Sec-Websocket-Version' header"
		_, _ = w.WriteString(fmt.Sprintf("HTTP/1.1 %d OK\r\nCache-Control: no-cache, must-revalidate\r\nPragma: no-cache\r\nContent-Length: %d\r\n\r\n%s", http.StatusBadRequest, len(s), s))
		return nil, HandshakeError{s}
	}

	if _, ok := responseHeader[textproto.CanonicalMIMEHeaderKey("Sec-Websocket-Extensions")]; ok {
		s := "websocket: application specific 'Sec-WebSocket-Extensions' headers are unsupported"
		_, _ = w.WriteString(fmt.Sprintf("HTTP/1.1 %d OK\r\nCache-Control: no-cache, must-revalidate\r\nPragma: no-cache\r\nContent-Length: %d\r\n\r\n%s", http.StatusInternalServerError, len(s), s))
		return nil, HandshakeError{s}
	}

	checkOrigin := u.CheckOrigin
	if checkOrigin == nil {
		checkOrigin = checkSameOrigin
	}
	if !checkOrigin(r) {
		s := "websocket: request origin not allowed by Upgrader.CheckOrigin"
		_, _ = w.WriteString(fmt.Sprintf("HTTP/1.1 %d OK\r\nCache-Control: no-cache, must-revalidate\r\nPragma: no-cache\r\nContent-Length: %d\r\n\r\n%s", http.StatusForbidden, len(s), s))
		return nil, HandshakeError{s}
	}
	challengeKey := r.Header.Get(textproto.CanonicalMIMEHeaderKey("Sec-WebSocket-Key"))
	if challengeKey == "" {
		s := "websocket: not a websocket handshake: 'Sec-WebSocket-Key' header is missing or blank"
		_, _ = w.WriteString(fmt.Sprintf("HTTP/1.1 %d OK\r\nCache-Control: no-cache, must-revalidate\r\nPragma: no-cache\r\nContent-Length: %d\r\n\r\n%s", http.StatusBadRequest, len(s), s))
		return nil, HandshakeError{s}
	}

	protocol := u.selectSubprotocol(r, responseHeader)

	// Negotiate PMCE
	var compress bool
	if u.EnableCompression {
		for _, ext := range ParseExtensions(r.Header) {
			if ext[""] != "permessage-deflate" {
				continue
			}
			compress = true
			break
		}
	}
	netConn := c_
	brw := b_
	if brw.Reader.Buffered() > 0 {
		netConn.Close()
		return nil, errors.New("websocket: client sent data before handshake is complete")
	}

	var br *bufio.Reader
	if u.ReadBufferSize == 0 && bufioReaderSize(netConn, brw.Reader) > 256 {
		// Reuse hijacked buffered reader as connection reader.
		br = brw.Reader
	}

	buf := bufioWriterBuffer(netConn, brw.Writer)

	var writeBuf []byte
	if u.WriteBufferPool == nil && u.WriteBufferSize == 0 && len(buf) >= maxFrameHeaderSize+256 {
		// Reuse hijacked write buffer as connection buffer.
		writeBuf = buf
	}

	c := newConn(netConn, true, u.ReadBufferSize, u.WriteBufferSize, u.WriteBufferPool, br, writeBuf)
	c.subprotocol = protocol

	if compress {
		c.newCompressionWriter = compressNoContextTakeover
		c.newDecompressionReader = decompressNoContextTakeover
	}

	// Use larger of hijacked buffer and connection write buffer for header.
	p := buf
	if len(c.writeBuf) > len(p) {
		p = c.writeBuf
	}
	p = p[:0]

	p = append(p, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: "...)
	p = append(p, computeAcceptKey(challengeKey)...)
	p = append(p, "\r\n"...)
	if c.subprotocol != "" {
		p = append(p, "Sec-WebSocket-Protocol: "...)
		p = append(p, c.subprotocol...)
		p = append(p, "\r\n"...)
	}
	if compress {
		p = append(p, "Sec-WebSocket-Extensions: permessage-deflate; server_no_context_takeover; client_no_context_takeover\r\n"...)
	}
	for k, vs := range responseHeader {
		if k == "Sec-Websocket-Protocol" {
			continue
		}
		for _, v := range vs {
			p = append(p, k...)
			p = append(p, ": "...)
			for i := 0; i < len(v); i++ {
				b := v[i]
				if b <= 31 {
					// prevent response splitting.
					b = ' '
				}
				p = append(p, b)
			}
			p = append(p, "\r\n"...)
		}
	}
	p = append(p, "\r\n"...)

	// Clear deadlines set by HTTP server.
	netConn.SetDeadline(time.Time{})

	if u.HandshakeTimeout > 0 {
		netConn.SetWriteDeadline(time.Now().Add(u.HandshakeTimeout))
	}
	if _, err := netConn.Write(p); err != nil {
		netConn.Close()
		return nil, err
	}
	if u.HandshakeTimeout > 0 {
		netConn.SetWriteDeadline(time.Time{})
	}

	return c, nil
}

// UpgradeClient 自己改写的 将当前客户端的连接升级为Websocket会话
func (u *Upgrader) UpgradeClient(r *http.Request, Response *http.Response, netConn net.Conn) (*Conn, error) {
	protocol := u.selectSubprotocol(r, Response.Header)
	c := newConn(netConn, true, u.ReadBufferSize, u.WriteBufferSize, u.WriteBufferPool, nil, nil)
	c.subprotocol = protocol
	extensions := ParseExtensions(Response.Header)
	for _, ext := range extensions {
		if ext[""] != "permessage-deflate" {
			continue
		}
		c.WindowReader.Window_Size_Max = 1 << 15
		c.newDecompressionReader = decompressNoContextTakeover
		break
	}
	var bs bytes.Buffer
	bs.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
	for k, vs := range Response.Header {
		for _, v := range vs {
			bs.WriteString(strings.ReplaceAll(textproto.CanonicalMIMEHeaderKey(k), "-Websocket-", "-WebSocket-") + ": " + v + "\r\n")
		}
	}
	bs.WriteString("\r\n")

	if _, err := netConn.Write(bs.Bytes()); err != nil {
		_ = netConn.Close()
		return nil, err
	}
	_ = netConn.SetDeadline(time.Time{})
	return c, nil
}

// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
//
// Deprecated: Use websocket.Upgrader instead.
//
// Upgrade does not perform origin checking. The application is responsible for
// checking the Origin header before calling Upgrade. An example implementation
// of the same origin policy check is:
//
//	if req.Header.Get("Origin") != "net://"+req.Host {
//		net.Error(w, "Origin not allowed", net.StatusForbidden)
//		return
//	}
//
// If the endpoint supports subprotocols, then the application is responsible
// for negotiating the protocol used on the connection. Use the Subprotocols()
// function to get the subprotocols requested by the client. Use the
// Sec-Websocket-Protocol response header to specify the subprotocol selected
// by the application.
//
// The responseHeader is included in the response to the client's upgrade
// request. Use the responseHeader to specify cookies (Set-Cookie) and the
// negotiated subprotocol (Sec-Websocket-Protocol).
//
// The connection buffers IO to the underlying network connection. The
// readBufSize and writeBufSize parameters specify the size of the buffers to
// use. Messages can be larger than the buffers.
//
// If the request is not a valid WebSocket handshake, then Upgrade returns an
// error of type HandshakeError. Applications should handle this error by
// replying to the client with an HTTP error response.
func Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header, readBufSize, writeBufSize int, c_ net.Conn, b_ *bufio.ReadWriter) (*Conn, error) {
	u := Upgrader{ReadBufferSize: readBufSize, WriteBufferSize: writeBufSize}
	u.Error = func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		// don't return errors to maintain backwards compatibility
	}
	u.CheckOrigin = func(r *http.Request) bool {
		// allow all connections by default
		return true
	}
	return u.Upgrade(w, r, responseHeader, c_, b_)
}

// Subprotocols returns the subprotocols requested by the client in the
// Sec-Websocket-Protocol header.
func Subprotocols(r *http.Request) []string {
	h := strings.TrimSpace(r.Header.Get("Sec-Websocket-Protocol"))
	if h == "" {
		return nil
	}
	protocols := strings.Split(h, ",")
	for i := range protocols {
		protocols[i] = strings.TrimSpace(protocols[i])
	}
	return protocols
}

// IsWebSocketUpgrade returns true if the client requested upgrade to the
// WebSocket protocol.
func IsWebSocketUpgrade(r *http.Request) bool {
	return tokenListContainsValue(r.Header, "Connection", "upgrade") &&
		tokenListContainsValue(r.Header, "Upgrade", "websocket")
}

// bufioReaderSize size returns the size of a bufio.Reader.
func bufioReaderSize(originalReader io.Reader, br *bufio.Reader) int {
	// This code assumes that peek on a reset reader returns
	// bufio.Reader.buf[:0].
	// TODO: Use bufio.Reader.Size() after Go 1.10
	br.Reset(originalReader)
	if p, err := br.Peek(0); err == nil {
		return cap(p)
	}
	return 0
}

// writeHook is an io.Writer that records the last slice passed to it vio
// io.Writer.Write.
type writeHook struct {
	p []byte
}

func (wh *writeHook) Write(p []byte) (int, error) {
	wh.p = p
	return len(p), nil
}

// bufioWriterBuffer grabs the buffer from a bufio.Writer.
func bufioWriterBuffer(originalWriter io.Writer, bw *bufio.Writer) []byte {
	// This code assumes that bufio.Writer.buf[:1] is passed to the
	// bufio.Writer's underlying writer.
	var wh writeHook
	bw.Reset(&wh)
	bw.WriteByte(0)
	bw.Flush()

	bw.Reset(originalWriter)

	return wh.p[:cap(wh.p)]
}
