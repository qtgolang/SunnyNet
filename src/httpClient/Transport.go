package httpClient

import (
	"context"
	"fmt"
	"github.com/qtgolang/SunnyNet/src/crypto/tls"
	"github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/http/http2"
	"net"
	"strings"
	"time"
)

type Transport struct {
	cachedTransports http.RoundTripper
	config           *tls.Config
	Profile          ClientProfile
	conn             net.Conn
	fidCount         int
	Timeout          time.Duration
}

func (rt *Transport) getDialTLSAddr(req *http.Request) string {
	host, port, err := net.SplitHostPort(req.URL.Host)
	if err == nil {
		return net.JoinHostPort(host, port)
	}
	return net.JoinHostPort(req.URL.Host, "443")
}
func (rt *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.fidCount = 0
	return rt.RoundTripDo(req)
}
func (rt *Transport) RoundTripDo(req *http.Request) (*http.Response, error) {
	addr := rt.getDialTLSAddr(req)
	if rt.cachedTransports == nil {
		if err := rt.getTransport(req, addr); err != nil {
			rt.fidCount++
			if rt.fidCount < 5 {
				if rt.conn != nil {
					_ = rt.conn.Close()
				}
				rt.conn = nil
				return rt.RoundTripDo(req)
			}
			return nil, err
		}
	}
	t := rt.cachedTransports
	res, err := t.RoundTrip(req)
	ok := false
	if err != nil {
		if strings.Contains(err.Error(), "use of closed network connection") {
			ok = true
		}
		if strings.Contains(err.Error(), "EOF") {
			ok = true
		}
	}
	if ok {
		rt.fidCount++
		if rt.fidCount < 5 {
			if rt.conn != nil {
				_ = rt.conn.Close()
			}
			rt.conn = nil
			return rt.RoundTripDo(req)
		}

	}
	if res != nil {
		//res.Request.SetContext("Encoding", res.Header.Get("Content-Encoding"))
		//res.Header.Del("Content-Encoding")

		/*
			修改这个文件处理自动解压缩 已经修改为了不要自动解压缩
			G:\Sunny\SunnyNetV4\src\http\transport.go
			func DecompressBody(res *Response) io.ReadCloser {
				return res.Body
				ce := res.Header.Get("Content-Encoding")
				res.ContentLength = -1
				res.Uncompressed = true
				return DecompressBodyByType(res.Body, ce)
			}
		*/

		res.Header.Del("Transfer-Encoding")
	}
	return res, err
}
func (rt *Transport) getTransport(req *http.Request, addr string) error {
	switch strings.ToLower(req.URL.Scheme) {
	case "http":
		rt.cachedTransports = rt.buildHttp1Transport()
		return nil
	case "https":
		break
	default:
		return fmt.Errorf("invalid URL scheme: [%v]", req.URL.Scheme)
	}
	_, err := rt.dialTLS(req.Context(), "tcp", addr)
	return err
}

func (rt *Transport) buildHttp1Transport() *http.Transport {
	t := &http.Transport{DialContext: rt.dial, DialTLSContext: rt.dialTLS, TLSClientConfig: rt.config}
	t.TLSHandshakeTimeout = rt.Timeout
	t.ResponseHeaderTimeout = rt.Timeout
	return t
}
func (rt *Transport) dialTLS(ctx context.Context, network, addr string) (net.Conn, error) {
	if rt.conn != nil {
		return rt.conn, nil
	}
	rawConn, err := rt.dial(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	conn := tls.Client(rawConn, rt.config)
	conn.SetDeadline(time.Now().Add(rt.Timeout))
	//conn := tls.UClient(rawConn, rt.config, rt.Profile.clientHelloId, false, false)
	if err = conn.HandshakeContext(ctx); err != nil {
		_ = conn.Close()
		return nil, err
	}
	switch conn.ConnectionState().NegotiatedProtocol {
	case http2.NextProtoTLS:
		rt.cachedTransports = &http.Transporth2{}
	default:
		rt.cachedTransports = rt.buildHttp1Transport()
	}
	conn.SetWriteDeadline(time.Now().Add(rt.Timeout))
	rt.conn = conn
	return conn, nil
}

func (rt *Transport) dial(ctx context.Context, network, addr string) (net.Conn, error) {
	var d net.Dialer
	d.Timeout = rt.Timeout
	if d.Timeout < 1 {
		d.Timeout = 30 * time.Second
	}
	return d.DialContext(ctx, network, addr)
}
