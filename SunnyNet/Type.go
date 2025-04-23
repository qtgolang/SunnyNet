package SunnyNet

import (
	"fmt"
	"github.com/qtgolang/SunnyNet/src/ReadWriteObject"
	"github.com/qtgolang/SunnyNet/src/http"
)

type errorRW struct {
	conn *ReadWriteObject.ReadWriteObject
	ok   bool
	h    http.Header
	s    int
}

func (w *errorRW) Header() http.Header {
	if w.h == nil {
		w.h = http.Header{}
	}
	return w.h
}
func (w *errorRW) WriteHeader(statusCode int) {
	w.s = statusCode
}

func (w *errorRW) Write(b []byte) (int, error) {
	if !w.ok {
		w.ok = true
		_, _ = w.conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", w.s, http.StatusText(w.s))))
		for k, v := range w.h {
			for _, vv := range v {
				_, _ = w.conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, vv)))
			}
		}
		_, _ = w.conn.Write([]byte("\r\n"))
	}
	return w.conn.Write(b)
}
