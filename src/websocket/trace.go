//go:build go1.8
// +build go1.8

package websocket

import (
	"github.com/qtgolang/SunnyNet/src/http/httptrace"
	"github.com/qtgolang/SunnyNet/src/crypto/tls"
)

func doHandshakeWithTrace(trace *httptrace.ClientTrace, tlsConn *tls.Conn, cfg *tls.Config) error {
	if trace.TLSHandshakeStart != nil {
		trace.TLSHandshakeStart()
	}
	err := doHandshake(tlsConn, cfg)
	if trace.TLSHandshakeDone != nil {
		trace.TLSHandshakeDone(tlsConn.ConnectionState(), err)
	}
	return err
}
