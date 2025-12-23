// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/qtgolang/SunnyNet/src/http/httptrace"
	"github.com/qtgolang/SunnyNet/src/http/internal/ascii"
	"github.com/qtgolang/SunnyNet/src/internal/textproto"

	"golang.org/x/net/http/httpguts"
)

// A Header represents the key-value pairs in an HTTP header.
//
// The keys should be in canonical form, as returned by
// CanonicalHeaderKey.
type Header map[string][]string

// Add adds the key, value pair to the header.
// It appends to any existing values associated with key.
// The key is case insensitive; it is canonicalized by
// CanonicalHeaderKey.
func (h Header) Add(key, value string) {
	for k, _ := range h {
		if strings.EqualFold(k, key) {
			textproto.MIMEHeader(h).Add(k, value)
			return
		}
	}
	textproto.MIMEHeader(h).Add(key, value)
}

// Set sets the header entries associated with key to the
// single element value. It replaces any existing values
// associated with key. The key is case insensitive; it is
// canonicalized by textproto.CanonicalMIMEHeaderKey.
// To use non-canonical keys, assign to the map directly.
func (h Header) Set(key, value string) {
	for k, _ := range h {
		if strings.EqualFold(k, key) {
			textproto.MIMEHeader(h).Set(k, value)
			return
		}
	}
	textproto.MIMEHeader(h).Set(key, value)

}

// Get gets the first value associated with the given key. If
// there are no values associated with the key, Get returns "".
// It is case insensitive; textproto.CanonicalMIMEHeaderKey is
// used to canonicalize the provided key. Get assumes that all
// keys are stored in canonical form. To use non-canonical keys,
// access the map directly.
func (h Header) Get(key string) string {
	for k, v := range h {
		if strings.EqualFold(k, key) {
			if len(v) > 0 {
				if strings.EqualFold("host", key) {
					if strings.HasSuffix(v[0], ":-1") || strings.HasSuffix(v[0], ":0") {
						clean := strings.TrimSuffix(v[0], ":-1")
						clean = strings.TrimSuffix(clean, ":0")
						h[k] = []string{clean}
						return clean
					}
				}
				return v[0]
			}
		}
	}
	return ""
}
func (h Header) GetArray(key string) []string {
	for k, v := range h {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return nil
}
func (h Header) SetArray(key string, val []string) {
	for k, _ := range h {
		if strings.EqualFold(k, key) {
			h[k] = val
			return
		}
	}
	h[key] = val
	return
}

// Values returns all values associated with the given key.
// It is case insensitive; textproto.CanonicalMIMEHeaderKey is
// used to canonicalize the provided key. To use non-canonical
// keys, access the map directly.
// The returned slice is not a copy.
func (h Header) Values(key string) []string {
	return textproto.MIMEHeader(h).Values(key)
}

// get is like Get, but key must already be in CanonicalHeaderKey form.
func (h Header) get(key string) string {
	return h.Get(key)
}

// has reports whether h has the provided key defined, even if it's
// set to 0-length slice.
func (h Header) has(key string) bool {
	for k, _ := range h {
		if strings.EqualFold(k, key) {
			return true
		}
	}
	return false
}

// Del deletes the values associated with key.
// The key is case insensitive; it is canonicalized by
// CanonicalHeaderKey.
func (h Header) Del(key string) {
	for k, _ := range h {
		if strings.EqualFold(k, key) {
			delete(h, k)
			return
		}
	}
}

// Write writes a header in wire format.
func (h Header) Write(w io.Writer) error {
	return h.write(w, nil)
}

func (h Header) write(w io.Writer, trace *httptrace.ClientTrace) error {
	return h.writeSubset(w, nil, trace)
}

// Clone returns a copy of h or nil if h is nil.
func (h Header) Clone() Header {
	if h == nil {
		return nil
	}

	// Find total number of values.
	nv := 0
	for _, vv := range h {
		nv += len(vv)
	}
	sv := make([]string, nv) // shared backing array for headers' values
	h2 := make(Header, len(h))
	for k, vv := range h {
		if vv == nil {
			// Preserve nil values. ReverseProxy distinguishes
			// between nil and zero-length header values.
			h2[k] = nil
			continue
		}
		n := copy(sv, vv)
		h2[k] = sv[:n:n]
		sv = sv[n:]
	}
	return h2
}

var timeFormats = []string{
	TimeFormat,
	time.RFC850,
	time.ANSIC,
}

// ParseTime parses a time header (such as the Date: header),
// trying each of the three formats allowed by HTTP/1.1:
// TimeFormat, time.RFC850, and time.ANSIC.
func ParseTime(text string) (t time.Time, err error) {
	for _, layout := range timeFormats {
		t, err = time.Parse(layout, text)
		if err == nil {
			return
		}
	}
	return
}

var headerNewlineToSpace = strings.NewReplacer("\n", " ", "\r", " ")

// stringWriter implements WriteString on a Writer.
type stringWriter struct {
	w io.Writer
}

func (w stringWriter) WriteString(s string) (n int, err error) {
	return w.w.Write([]byte(s))
}

type KeyValues struct {
	Key    string
	Values []string
}

// A headerSorter implements sort.Interface by sorting a []keyValues
// by key. It's used as a pointer, so it can fit in a sort.Interface
// interface value without allocation.
type headerSorter struct {
	kvs   []KeyValues
	order map[string]int
}

func (s *headerSorter) Len() int           { return len(s.kvs) }
func (s *headerSorter) Swap(i, j int)      { s.kvs[i], s.kvs[j] = s.kvs[j], s.kvs[i] }
func (s *headerSorter) Less(i, j int) bool { return s.kvs[i].Key < s.kvs[j].Key }

var headerSorterPool = sync.Pool{
	New: func() any { return new(headerSorter) },
}

// sortedKeyValues returns h's keys sorted in the returned kvs
// slice. The headerSorter used to sort is also returned, for possible
// return to headerSorterCache.
func (h Header) sortedKeyValues(exclude map[string]bool) (kvs []KeyValues, hs *headerSorter) {
	hs = headerSorterPool.Get().(*headerSorter)
	if cap(hs.kvs) < len(h) {
		hs.kvs = make([]KeyValues, 0, len(h))
	}
	kvs = hs.kvs[:0]
	for k, vv := range h {
		if !exclude[k] {
			kvs = append(kvs, KeyValues{k, vv})
		}
	}
	hs.kvs = kvs
	sort.Sort(hs)
	return kvs, hs
}

// WriteSubset writes a header in wire format.
// If exclude is not nil, keys where exclude[key] == true are not written.
// Keys are not canonicalized before checking the exclude map.
func (h Header) WriteSubset(w io.Writer, exclude map[string]bool) error {
	return h.writeSubset(w, exclude, nil)
}

func (h Header) writeSubset(w io.Writer, exclude map[string]bool, trace *httptrace.ClientTrace) error {
	ws, ok := w.(io.StringWriter)
	if !ok {
		ws = stringWriter{w}
	}
	kvs, sorter := h.sortedKeyValues(exclude)
	var formattedVals []string
	for _, kv := range kvs {
		if !httpguts.ValidHeaderFieldName(kv.Key) {
			// This could be an error. In the common case of
			// writing response headers, however, we have no good
			// way to provide the error back to the server
			// handler, so just drop invalid headers instead.
			continue
		}
		for _, v := range kv.Values {
			v = headerNewlineToSpace.Replace(v)
			v = textproto.TrimString(v)
			for _, s := range []string{kv.Key, ": ", v, "\r\n"} {
				if _, err := ws.WriteString(s); err != nil {
					headerSorterPool.Put(sorter)
					return err
				}
			}
			if trace != nil && trace.WroteHeaderField != nil {
				formattedVals = append(formattedVals, v)
			}
		}
		if trace != nil && trace.WroteHeaderField != nil {
			trace.WroteHeaderField(kv.Key, formattedVals)
			formattedVals = nil
		}
	}
	headerSorterPool.Put(sorter)
	return nil
}

func (h Header) SortedKeyValues(exclude map[string]bool) (kvs []KeyValues, hs *headerSorter) {
	hs = headerSorterPool.Get().(*headerSorter)
	if cap(hs.kvs) < len(h) {
		hs.kvs = make([]KeyValues, 0, len(h))
	}
	kvs = hs.kvs[:0]
	for k, vv := range h {
		mutex.RLock()
		if !exclude[k] {
			kvs = append(kvs, KeyValues{k, vv})
		}
		mutex.RUnlock()
	}
	hs.kvs = kvs
	sort.Sort(hs)
	return kvs, hs
}

var mutex = &sync.RWMutex{}

func (h Header) SortedKeyValuesBy(order map[string]int, exclude map[string]bool) (kvs []KeyValues, hs *headerSorter) {
	hs = headerSorterPool.Get().(*headerSorter)
	if cap(hs.kvs) < len(h) {
		hs.kvs = make([]KeyValues, 0, len(h))
	}
	kvs = hs.kvs[:0]
	for k, vv := range h {
		mutex.RLock()
		if !exclude[k] {
			kvs = append(kvs, KeyValues{k, vv})
		}
		mutex.RUnlock()
	}
	hs.kvs = kvs
	hs.order = order
	sort.Sort(hs)

	return kvs, hs
}

// CanonicalHeaderKey returns the canonical format of the
// header key s. The canonicalization converts the first
// letter and any letter following a hyphen to upper case;
// the rest are converted to lowercase. For example, the
// canonical key for "accept-encoding" is "Accept-Encoding".
// If s contains a space or invalid header field bytes, it is
// returned without modifications.
func CanonicalHeaderKey(s string) string { return textproto.CanonicalMIMEHeaderKey(s) }

// hasToken reports whether token appears with v, ASCII
// case-insensitive, with space or comma boundaries.
// token must be all lowercase.
// v may contain mixed cased.
func hasToken(v, token string) bool {
	if len(token) > len(v) || token == "" {
		return false
	}
	if v == token {
		return true
	}
	for sp := 0; sp <= len(v)-len(token); sp++ {
		// Check that first character is good.
		// The token is ASCII, so checking only a single byte
		// is sufficient. We skip this potential starting
		// position if both the first byte and its potential
		// ASCII uppercase equivalent (b|0x20) don't match.
		// False positives ('^' => '~') are caught by EqualFold.
		if b := v[sp]; b != token[0] && b|0x20 != token[0] {
			continue
		}
		// Check that start pos is on a valid token boundary.
		if sp > 0 && !isTokenBoundary(v[sp-1]) {
			continue
		}
		// Check that end pos is on a valid token boundary.
		if endPos := sp + len(token); endPos != len(v) && !isTokenBoundary(v[endPos]) {
			continue
		}
		if ascii.EqualFold(v[sp:sp+len(token)], token) {
			return true
		}
	}
	return false
}

func isTokenBoundary(b byte) bool {
	return b == ' ' || b == ',' || b == '\t'
}
