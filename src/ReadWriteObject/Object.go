package ReadWriteObject

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"time"
)

// NewReadWriteObject 构建读写对象
func NewReadWriteObject(c net.Conn) *ReadWriteObject {
	r := bufio.NewReader(c)
	w := bufio.NewWriter(c)
	return &ReadWriteObject{ReadWriter: bufio.NewReadWriter(r, w), c: c}
}

// ReadWriteObject 数据读写流
type ReadWriteObject struct {
	*bufio.ReadWriter
	c    net.Conn
	Hook *bytes.Buffer
}

func (w *ReadWriteObject) ReadSlice(delim byte) ([]byte, error) {
	var lineBuf []byte
	for {
		bs, e := w.Reader.ReadSlice(delim)
		lineBuf = append(lineBuf, bs...) // 直接追加
		if e == nil {                    // 正常读完一行
			return lineBuf, nil
		}
		if errors.Is(e, bufio.ErrBufferFull) {
			continue // 行太长，继续读
		}
		// 其他错误，返回当前已读数据
		return lineBuf, e
	}
}
func (w *ReadWriteObject) LocalAddr() net.Addr {
	return w.c.LocalAddr()
}
func (w *ReadWriteObject) Conn() net.Conn {
	return w.c
}
func (w *ReadWriteObject) Close() error {
	return w.c.Close()
}
func (w *ReadWriteObject) SetWriteDeadline(t time.Time) error {
	return w.c.SetWriteDeadline(t)
}
func (w *ReadWriteObject) SetReadDeadline(t time.Time) error {
	return w.c.SetReadDeadline(t)
}
func (w *ReadWriteObject) SetDeadline(t time.Time) error {
	return w.c.SetDeadline(t)
}

func (w *ReadWriteObject) Write(b []byte) (nn int, err error) {
	i, e := w.Writer.Write(b)
	w.Writer.Flush()
	return i, e
}
func (w *ReadWriteObject) RemoteAddr() net.Addr {
	return w.c.RemoteAddr()
}

func (w *ReadWriteObject) WriteString(b string) (nn int, err error) {
	i, e := w.Writer.Write([]byte(b))
	e = w.Flush()
	return i, e
}
func (w *ReadWriteObject) Read(b []byte) (nn int, err error) {
	i, e := w.ReadWriter.Read(b)
	if w.Hook != nil {
		w.Hook.Write(b[:i])
	}
	return i, e
}

func (w *ReadWriteObject) Buffered() int {
	return w.ReadWriter.Reader.Buffered()
}
