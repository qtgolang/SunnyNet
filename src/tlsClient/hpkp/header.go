package hpkp

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Header holds a domain's hpkp information
type Header struct {
	Created           int64
	MaxAge            int64
	IncludeSubDomains bool
	Permanent         bool
	Sha256Pins        []string
	ReportURI         string
}

// Matches checks whether the provided pin is in the header list
func (h *Header) Matches(pin string) bool {
	for i := range h.Sha256Pins {
		if h.Sha256Pins[i] == pin {
			return true
		}
	}
	return false
}

// ParseHeader parses the hpkp information from an http.Response.
func ParseHeader(resp *http.Response) *Header {
	if resp == nil {
		return nil
	}

	// only make a header when using TLS
	if resp.TLS == nil {
		return nil
	}

	v, ok := resp.Header["Public-Key-Pins"]
	if !ok {
		return nil
	}

	// use the first header per RFC
	return populate(&Header{}, v[0])
}

// ParseReportOnlyHeader parses the hpkp information from an http.Response.
// The resulting header information should not be cached as max_age is
// ignored on HPKP-RO headers per the RFC.
func ParseReportOnlyHeader(resp *http.Response) *Header {
	if resp == nil {
		return nil
	}

	// only make a header when using TLS
	if resp.TLS == nil {
		return nil
	}

	v, ok := resp.Header["Public-Key-Pins-Report-Only"]
	if !ok {
		return nil
	}

	// use the first header per RFC
	return populate(&Header{}, v[0])
}

func populate(h *Header, v string) *Header {
	h.Sha256Pins = []string{}

	for _, field := range strings.Split(v, ";") {
		field = strings.TrimSpace(field)

		i := strings.Index(field, "pin-sha256")
		if i >= 0 {
			h.Sha256Pins = append(h.Sha256Pins, field[i+12:len(field)-1])
			continue
		}

		i = strings.Index(field, "report-uri")
		if i >= 0 {
			h.ReportURI = field[i+12 : len(field)-1]
			continue
		}

		i = strings.Index(field, "max-age=")
		if i >= 0 {
			ma, err := strconv.Atoi(field[i+8:])
			if err == nil {
				h.MaxAge = int64(ma)
			}
			continue
		}

		if strings.Contains(field, "includeSubDomains") {
			h.IncludeSubDomains = true
			continue
		}
	}

	h.Created = time.Now().Unix()
	return h
}
