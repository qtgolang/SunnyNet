// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

// Common HTTP methods.
//
// Unless otherwise noted, these are defined in RFC 7231 section 4.3.
const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH" // RFC 5789
	MethodDelete  = "DELETE"
	MethodConnect = "CONNECT"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
	H10Proto      = "http/1.0"
	H11Proto      = "http/1.1"
	H2Proto       = "h2"
)

var ProtoVersions = map[uint16]string{
	770: H10Proto, // HTTP/1.0
	771: H11Proto, // HTTP/1.1
	772: H2Proto,  // HTTP/2.0
}
