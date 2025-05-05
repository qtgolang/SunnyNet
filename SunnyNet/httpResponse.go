package SunnyNet

import (
	"github.com/qtgolang/SunnyNet/src/http"
	"io"
	"net"
	"strings"
)

type response struct {
	*http.Response
	rw       http.ResponseWriter
	Conn     net.Conn
	Close    func()
	ServerIP string
}

func isHeader(key string) bool {
	switch key {
	case "Transfer-Encoding":
		return true
	}
	return false
}
func (r *response) Done() {
	r.StatusCode = r.Response.StatusCode
	if r.StatusCode < 1 {
		r.StatusCode = 200
	}
	if len(r.Header) < 1 {
		if r.ProtoMajor == 2 {
			r.rw.Header().Set("connection", "Close")
		} else {
			r.rw.Header().Set("Connection", "Close")
		}
	} else {
		for k, v := range r.Header {
			if isHeader(k) {
				continue
			}
			r.rw.Header()[k] = v
		}
	}
	r.rw.WriteHeader(r.StatusCode)
	if r.Body != nil {
		bodyBytes, _ := io.ReadAll(r.Body)
		_, _ = r.rw.Write(bodyBytes)
	}
}
func (r *response) WriteHeader(DataLen ...string) []byte {
	contentLength := ""
	if len(DataLen) > 0 {
		contentLength = DataLen[0]
	}

	r.DelHeader("content-length")
	if contentLength != "-1" {
		if r.ProtoMajor == 2 {
			r.rw.Header().Set("content-length", contentLength)
		} else {
			r.rw.Header().Set("Content-Length", contentLength)
		}
	}
	for name, values := range r.Header {
		if strings.ToLower(name) == "content-type" {
			r.rw.Header()["Content-Type"] = values
			continue
		}
		r.rw.Header()[name] = values
	}
	r.rw.WriteHeader(r.StatusCode)
	return nil
}
func (r *response) Write(b []byte) (int, error) {
	return r.rw.Write(b)
}
func (r *response) DelHeader(keys ...string) {
	for _, key := range keys {
		k := strings.ToLower(key)
		for name, _ := range r.Header {
			if strings.ToLower(name) == k {
				r.Header.Del(name)
				continue
			}
		}
	}
}
