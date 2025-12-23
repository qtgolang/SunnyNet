package SunnyNet

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/qtgolang/SunnyNet/src/ReadWriteObject"
	"github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/public"
)

type objHook struct {
	*ReadWriteObject.ReadWriteObject
	aheadData []byte
}

func newObjHook(obj *ReadWriteObject.ReadWriteObject, aheadData []byte) *objHook {
	var hookObj objHook
	hookObj.ReadWriteObject = obj
	hookObj.aheadData = aheadData
	return &hookObj
}
func (n *objHook) Read(p []byte) (int, error) {
	if len(n.aheadData) < 1 {
		return n.ReadWriteObject.Read(p)
	}
	// 确定可以复制的最大字节数
	copyLength := len(n.aheadData)
	if len(p) < copyLength {
		copyLength = len(p)
	}
	copy(p, n.aheadData[:copyLength])
	n.aheadData = n.aheadData[copyLength:]
	if copyLength == 1 && copyLength < len(p) {
		a, e := n.ReadWriteObject.Read(p[copyLength:])
		return a + copyLength, e
	}
	return copyLength, nil
}
func (s *proxyRequest) h2Request(aheadData []byte) {
	s._SocksUser = GetSocket5User(s.Theology)
	http.H2NewConn(newObjHook(s.RwObj, aheadData), s.httpCall)
}
func (s *proxyRequest) h1Request(aheadData []byte) {
	s._SocksUser = GetSocket5User(s.Theology)
	http.H1NewConn(newObjHook(s.RwObj, aheadData), s.httpCall)
}

type httpBody struct {
	Body io.ReadCloser
	c    net.Conn
	req  *http.Request
	file io.WriteCloser
	init bool
	lock *sync.Mutex
}

func (h *httpBody) Read(p []byte) (n int, err error) {
	if !h.init {
		h.init = true
		SaveFilePath, ok := h.req.Context().Value(public.SunnyNetRawBodySaveFilePath).(string)
		if ok && SaveFilePath != "" {
			//防止多个请求写入同一个文件，导致闪退等问题
			_lock.Lock()
			lo := _lockfileMap[SaveFilePath]
			if lo == nil {
				lo = &sync.Mutex{}
				_lockfileMap[SaveFilePath] = lo
			}
			h.lock = lo
			_lock.Unlock()
			file, er1 := os.OpenFile(SaveFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
			if er1 == nil {
				h.file = file
			}
		}
	}
	if h.lock != nil {
		h.lock.Lock()
		defer h.lock.Unlock()
	}
	_ = h.c.SetReadDeadline(time.Now().Add(10 * time.Second))
	n, e := h.Body.Read(p)
	if h.file != nil && n > 0 {
		_, _ = h.file.Write(p[0:n])
	}
	if h.file != nil && e != nil {
		_ = h.file.Close()
		h.file = nil
	}
	return n, e
}
func (h *httpBody) Close() error {
	if h.lock != nil {
		h.lock.Lock()
		defer h.lock.Unlock()
	}
	if h.file != nil {
		_ = h.file.Close()
		h.file = nil
	}
	return h.Body.Close()
}

var _lock sync.Mutex
var _lockfileMap = make(map[string]*sync.Mutex)

func normalizeHostPort(h string) string {
	if h == "" {
		return ""
	}
	if strings.Contains(h, "[") && strings.Contains(h, ".") {
		host, port, _ := net.SplitHostPort(h)
		if host == "" {
			return h
		}
		if port != "" {
			return fmt.Sprintf("%s:%s", host, port)
		}
		return host
	}
	return h
}

func (s *proxyRequest) httpCall(rw http.ResponseWriter, req *http.Request) {
	if req == nil {
		return
	}

	r := s.clone()
	defer r.free()

	Target := r.Target.Clone()

	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), public.Connect_Raw_Address, Target.String))
	defer cancel()

	res := req.Clone(ctx)

	// body 处理
	if res.GetIsNullBody() {
		if res.Body != nil {
			_ = res.Body.Close()
		}
		res.Body = nil
		res.ContentLength = 0
	} else {
		res.Body = &httpBody{Body: req.Body, c: s.Conn, req: res}
	}

	// raw body 标记
	bodyLen := res.GetBodyLength()
	res.SetContext(public.SunnyNetRawRequestBodyLength, bodyLen)
	res.IsRawBody = bodyLen >= s.Global.httpMaxBodyLen

	// host 规范化
	res.Host = normalizeHostPort(res.Host)
	if res.URL != nil {
		res.URL.Host = normalizeHostPort(res.URL.Host)
	}

	// header host 规范化
	if mh := res.Header.Get("host"); mh != "" {
		nmh := normalizeHostPort(mh)
		if nmh != mh {
			res.Header.Del("host")
		}
	}

	res.RequestURI = ""

	// URL 处理
	if res.URL != nil {
		// scheme 决策
		if r.defaultScheme == "" || req.URL.Scheme == "https" {
			res.URL.Scheme = "https"
		} else {
			res.URL.Scheme = r.defaultScheme
		}

		// Target 默认端口
		if r.Target.Port == 0 {
			if req.URL.Scheme == "https" {
				r.Target.Parse("", 443)
			} else {
				r.Target.Parse("", 80)
			}
		}

		// Target.Host 为空：从 res.Host / header host 推断
		if r.Target.Host == "" {
			if res.Host != "" {
				r.Target.Parse(res.Host, 0)
			} else if h := req.Header.Get("host"); h != "" {
				r.Target.Parse(h, 0)
			}

			if r.Target.IsDomain() {
				res.Host = r.Target.String()
				res.URL.Host = res.Host
			} else if h := req.Header.Get("host"); h != "" {
				res.Host = h
				res.URL.Host = res.Host
			}
		}

		// 可能被重写了，再规范化一次
		res.Host = normalizeHostPort(req.Host)

		// res.Host 为空兜底
		if res.Host == "" {
			if h := req.Header.Get("host"); h != "" {
				res.URL.Host = h
				u, _ := url.Parse(res.URL.String())
				if u != nil {
					res.URL = u
					res.Host = u.Host
				}
				if r.Target.Host == "" && r.Target.Port == 0 {
					r.Target.Parse(res.Host, 0)
				}
			} else if r.Target.Host != "" {
				res.URL.Host = r.Target.String()
				u, _ := url.Parse(res.URL.String())
				if u != nil {
					res.URL = u
					res.Host = u.Host
				}
			}
		} else {
			// res.Host 有值：根据是否域名以及是否与当前目标一致决定 URL.Host
			aIP := TargetInfo{}
			aIP.Parse(res.Host, 0)
			if !aIP.IsDomain() && aIP.Host != s.Target.Host {
				res.URL.Host = r.Target.String()
			} else {
				res.URL.Host = res.Host
			}
		}

		// 标准端口时去掉 :80/:443
		p := res.URL.Port()
		if (p == "443" && res.URL.Scheme == "https") || (p == "80" && res.URL.Scheme == "http") {
			host, _, _ := net.SplitHostPort(res.Host)
			if host != "" {
				res.URL.Host = host
				res.Host = host
			} else {
				res.URL.Host = res.Host
			}
		}

		// URL端口与 Target端口不一致时，重写成 Target.String()
		_p, _ := strconv.Atoi(res.URL.Port())
		if _p != int(r.Target.Port) {
			if !((_p != 443 && r.Target.Port == 443) || (_p != 80 && r.Target.Port == 80)) {
				res.URL.Host = r.Target.String()
				u, _ := url.Parse(res.URL.String())
				if u != nil {
					res.URL = u
					res.Host = u.Host
					if res.Header.Get("host") != "" {
						res.Header.Set("host", u.Host)
					}
				}
			}
		}

		// IPv6 host 要加中括号
		ip := net.ParseIP(res.Host)
		if ip != nil {
			if ip.To4() == nil && len(ip) == net.IPv6len {
				res.URL.Host = "[" + res.Host + "]"
				res.Host = "[" + res.URL.Host + "]"
			}
		}
	}

	r.Response.rw = rw

	// HTTP/2 header key 规范化
	if res.ProtoMajor == 2 {
		reHeader := make(http.Header)
		for k, v := range res.Header {
			name := textproto.CanonicalMIMEHeaderKey(k)
			reHeader[name] = append(reHeader[name], v...)
		}
		res.Header = reHeader
	}

	Target.Parse(r.Target.String(), 0)
	res.TransferEncoding = nil

	t1 := time.Now()
	r.sendHttp(res)

	if time.Since(t1) > 5*time.Minute {
		_ = s.Conn.Close()
	}
}
