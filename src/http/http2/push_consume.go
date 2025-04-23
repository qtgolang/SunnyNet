// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package http2

import (
	"context"
	"errors"
	"net/url"

	 "github.com/qtgolang/SunnyNet/src/http"
)

var (
	errMissingHeaderMethod    = errors.New("http2: missing required request pseudo-header :method")
	errMissingHeaderScheme    = errors.New("http2: missing required request pseudo-header :scheme")
	errMissingHeaderPath      = errors.New("http2: missing required request pseudo-header :path")
	errMissingHeaderAuthority = errors.New("http2: missing required request pseudo-header :authority")
	errInvalidMethod          = errors.New("http2: method must be GET or HEAD")
	errInvalidScheme          = errors.New("http2: scheme must be http or https")
)

// DefaultPushHandler is a simple push handler for reading pushed responses
type DefaultPushHandler struct {
	promise       *http.Request
	origReqURL    *url.URL
	origReqHeader http.Header
	push          *http.Response
	pushErr       error
	done          chan struct{}
}

func (ph *DefaultPushHandler) HandlePush(r *PushedRequest) {
	ph.promise = r.Promise
	ph.origReqURL = r.OriginalRequestURL
	ph.origReqHeader = r.OriginalRequestHeader
	ph.push, ph.pushErr = r.ReadResponse(r.Promise.Context())
	if ph.done != nil {
		close(ph.done)
	}
}

// PushHandler consumes a pushed response.
type PushHandler interface {
	// HandlePush will be called once for every PUSH_PROMISE received
	// from the server. If HandlePush returns before the pushed stream
	// has completed, the pushed stream will be canceled.
	HandlePush(r *PushedRequest)
}

// PushedRequest describes a request that was pushed from the server.
type PushedRequest struct {
	// Promise is the HTTP/2 PUSH_PROMISE message. The promised
	// request does not have a body. Handlers should not modify Promise.
	//
	// Promise.RemoteAddr is the address of the server that started this push request.
	Promise *http.Request

	// OriginalRequestURL is the URL of the original client request that triggered the push.
	OriginalRequestURL *url.URL

	// OriginalRequestHeader contains the headers of the original client request that triggered the push.
	OriginalRequestHeader http.Header
	pushedStream          *clientStream
}

// ReadResponse reads the pushed response. If ctx is done before the
// response headers are fully received, ReadResponse will fail and PushedRequest
// will be cancelled.
func (pr *PushedRequest) ReadResponse(ctx context.Context) (*http.Response, error) {
	select {
	case <-ctx.Done():
		pr.Cancel()
		pr.pushedStream.bufPipe.CloseWithError(ctx.Err())
		return nil, ctx.Err()
	case <-pr.pushedStream.peerReset:
		return nil, pr.pushedStream.resetErr
	case resErr := <-pr.pushedStream.resc:
		if resErr.err != nil {
			pr.Cancel()
			pr.pushedStream.bufPipe.CloseWithError(resErr.err)
			return nil, resErr.err
		}
		resErr.res.Request = pr.Promise
		resErr.res.TLS = pr.pushedStream.cc.tlsState
		return resErr.res, resErr.err
	}
}

// Cancel tells the server that the pushed response stream should be terminated.
// See: https://tools.ietf.org/html/rfc7540#section-8.2.2
func (pr *PushedRequest) Cancel() {
	pr.pushedStream.cancelStream()
}

func pushedRequestToHTTPRequest(mppf *MetaPushPromiseFrame) (*http.Request, error) {
	method := mppf.PseudoValue("method")
	scheme := mppf.PseudoValue("scheme")
	authority := mppf.PseudoValue("authority")
	path := mppf.PseudoValue("path")
	// pseudo-headers required in all http2 requests
	if method == "" {
		return nil, errMissingHeaderMethod
	}
	if scheme == "" {
		return nil, errMissingHeaderScheme
	}
	if path == "" {
		return nil, errMissingHeaderPath
	}
	// authority is required for PUSH_PROMISE requests per RFC 7540 Section 8.2
	if authority == "" {
		return nil, errMissingHeaderAuthority
	}
	// "Promised requests MUST be cacheable (see [RFC7231], Section 4.2.3),
	// MUST be safe (see [RFC7231], Section 4.2.1)"
	// https://tools.ietf.org/html/rfc7540#section-8.2
	if method != "GET" && method != "HEAD" {
		return nil, errInvalidMethod
	}
	if scheme != "http" && scheme != "https" {
		return nil, errInvalidScheme
	}
	var headers http.Header
	for _, header := range mppf.RegularFields() {
		if len(headers) == 0 {
			headers = http.Header{}
		}
		headers.Add(header.Name, header.Value)
	}
	if err := checkValidPushPromiseRequestHeaders(headers); err != nil {
		return nil, err
	}
	if err := checkValidHTTP2RequestHeaders(headers); err != nil {
		return nil, err
	}
	reqUrl, err := url.ParseRequestURI(path)
	if err != nil {
		return nil, err
	}
	reqUrl.Host = authority
	reqUrl.Scheme = scheme
	return &http.Request{
		Method:     method,
		Proto:      "HTTP/2.0",
		ProtoMajor: 2,
		URL:        reqUrl,
		Header:     headers,
	}, nil
}

// handlePushEarlyReturnCancel handles the pushed request with the push handler.
// If PushHandler.HandlePush returns before the pushed stream has completed, the pushed
// stream is canceled.
func handlePushEarlyReturnCancel(pushHandler PushHandler, pushedRequest *PushedRequest) {
	handleReturned := make(chan struct{})
	go func() {
		defer close(handleReturned)
		pushHandler.HandlePush(pushedRequest)
	}()
	select {
	case <-handleReturned:
		pushedRequest.Cancel()
	case <-pushedRequest.pushedStream.done:
	}
}
