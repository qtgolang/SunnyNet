package hpkp

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestHeader_Matches(t *testing.T) {
	tests := []struct {
		name     string
		header   *Header
		pin      string
		expected bool
	}{
		{
			name:     "no match",
			header:   &Header{},
			pin:      "d6qzRu9zOECb90Uez27xWltNsj0e1Md7GkYYkVoZWmM=",
			expected: false,
		},
		{
			name: "match",
			header: &Header{
				Sha256Pins: []string{
					"d6qzRu9zOECb90Uez27xWltNsj0e1Md7GkYYkVoZWmM=",
					"E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g=",
				},
			},
			pin:      "E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g=",
			expected: true,
		},
	}

	for _, test := range tests {
		out := test.header.Matches(test.pin)
		if out != test.expected {
			t.Logf("want:%v", test.expected)
			t.Logf("got:%v", out)
			t.Fatalf("test case failed: %s", test.name)
		}
	}
}

func TestParseHeader(t *testing.T) {
	tests := []struct {
		name     string
		response *http.Response
		expected *Header
	}{
		{
			name:     "nil everything",
			response: nil,
			expected: nil,
		},
		{
			name: "no header",
			response: &http.Response{
				StatusCode: 200,
			},
			expected: nil,
		},
		{
			name: "hpkp header, but over http",
			response: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Public-Key-Pins": []string{`max-age=3000; pin-sha256="d6qzRu9zOECb90Uez27xWltNsj0e1Md7GkYYkVoZWmM="; pin-sha256="E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g="`},
				},
			},
			expected: nil,
		},
		{
			name: "multiple headers",
			response: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Public-Key-Pins": []string{
						`max-age=3000; pin-sha256="d6qzRu9zOECb90Uez27xWltNsj0e1Md7GkYYkVoZWmM="; pin-sha256="E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g="`,
						`max-age=3001; pin-sha256="bad header"`,
					},
				},
				TLS: &tls.ConnectionState{},
			},
			expected: &Header{
				MaxAge:            3000,
				IncludeSubDomains: false,
				Permanent:         false,
				Sha256Pins: []string{
					"d6qzRu9zOECb90Uez27xWltNsj0e1Md7GkYYkVoZWmM=",
					"E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g=",
				},
			},
		},
		// https://tools.ietf.org/html/rfc7469#section-2.1.5
		{
			name: "hpkp header (1)",
			response: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Public-Key-Pins": []string{`max-age=3000; pin-sha256="d6qzRu9zOECb90Uez27xWltNsj0e1Md7GkYYkVoZWmM="; pin-sha256="E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g="`},
				},
				TLS: &tls.ConnectionState{},
			},
			expected: &Header{
				MaxAge:            3000,
				IncludeSubDomains: false,
				Permanent:         false,
				Sha256Pins: []string{
					"d6qzRu9zOECb90Uez27xWltNsj0e1Md7GkYYkVoZWmM=",
					"E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g=",
				},
			},
		},
		{
			name: "hpkp header (2)",
			response: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Public-Key-Pins": []string{`max-age=2592000; pin-sha256="E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g="; pin-sha256="LPJNul+wow4m6DsqxbninhsWHlwfp0JecwQzYpOLmCQ="`},
				},
				TLS: &tls.ConnectionState{},
			},
			expected: &Header{
				MaxAge:            2592000,
				IncludeSubDomains: false,
				Permanent:         false,
				Sha256Pins: []string{
					"E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g=",
					"LPJNul+wow4m6DsqxbninhsWHlwfp0JecwQzYpOLmCQ=",
				},
			},
		},
		{
			name: "hpkp header (3)",
			response: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Public-Key-Pins": []string{`max-age=2592000; pin-sha256="E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g="; pin-sha256="LPJNul+wow4m6DsqxbninhsWHlwfp0JecwQzYpOLmCQ="; report-uri="http://example.com/pkp-report"`},
				},
				TLS: &tls.ConnectionState{},
			},
			expected: &Header{
				MaxAge:            2592000,
				IncludeSubDomains: false,
				Permanent:         false,
				Sha256Pins: []string{
					"E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g=",
					"LPJNul+wow4m6DsqxbninhsWHlwfp0JecwQzYpOLmCQ=",
				},
				ReportURI: "http://example.com/pkp-report",
			},
		},
		{
			name: "hpkp header (4)",
			response: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Public-Key-Pins-Report-Only": []string{`max-age=2592000; pin-sha256="E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g="; pin-sha256="LPJNul+wow4m6DsqxbninhsWHlwfp0JecwQzYpOLmCQ="; report-uri="http://example.com/pkp-report"`},
				},
				TLS: &tls.ConnectionState{},
			},
			expected: nil,
		},
		{
			name: "hpkp header (5)",
			response: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Public-Key-Pins": []string{`pin-sha256="d6qzRu9zOECb90Uez27xWltNsj0e1Md7GkYYkVoZWmM="; pin-sha256="LPJNul+wow4m6DsqxbninhsWHlwfp0JecwQzYpOLmCQ="; max-age=259200`},
				},
				TLS: &tls.ConnectionState{},
			},
			expected: &Header{
				MaxAge:            259200,
				IncludeSubDomains: false,
				Permanent:         false,
				Sha256Pins: []string{
					"d6qzRu9zOECb90Uez27xWltNsj0e1Md7GkYYkVoZWmM=",
					"LPJNul+wow4m6DsqxbninhsWHlwfp0JecwQzYpOLmCQ=",
				},
			},
		},
		{
			name: "hpkp header (6)",
			response: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Public-Key-Pins": []string{`pin-sha256="d6qzRu9zOECb90Uez27xWltNsj0e1Md7GkYYkVoZWmM="; pin-sha256="E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g="; pin-sha256="LPJNul+wow4m6DsqxbninhsWHlwfp0JecwQzYpOLmCQ="; max-age=10000; includeSubDomains`},
				},
				TLS: &tls.ConnectionState{},
			},
			expected: &Header{
				MaxAge:            10000,
				IncludeSubDomains: true,
				Permanent:         false,
				Sha256Pins: []string{
					"d6qzRu9zOECb90Uez27xWltNsj0e1Md7GkYYkVoZWmM=",
					"E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g=",
					"LPJNul+wow4m6DsqxbninhsWHlwfp0JecwQzYpOLmCQ=",
				},
			},
		},
	}

	for _, test := range tests {
		out := ParseHeader(test.response)
		if !equalHeaders(out, test.expected) {
			t.Logf("want:%v", test.expected)
			t.Logf("got:%v", out)
			t.Fatalf("test case failed: %s", test.name)
		}
	}
}

func TestParseReportOnlyHeader(t *testing.T) {
	tests := []struct {
		name     string
		response *http.Response
		expected *Header
	}{
		{
			name:     "nil everything",
			response: nil,
			expected: nil,
		},
		{
			name: "no header",
			response: &http.Response{
				StatusCode: 200,
			},
			expected: nil,
		},
		{
			name: "hpkp header, but over http",
			response: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Public-Key-Pins": []string{`max-age=3000; pin-sha256="d6qzRu9zOECb90Uez27xWltNsj0e1Md7GkYYkVoZWmM="; pin-sha256="E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g="`},
				},
			},
			expected: nil,
		},
		{
			name: "multiple headers",
			response: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Public-Key-Pins-Report-Only": []string{
						`max-age=3000; pin-sha256="d6qzRu9zOECb90Uez27xWltNsj0e1Md7GkYYkVoZWmM="; pin-sha256="E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g="`,
						`max-age=3001; pin-sha256="bad header"`,
					},
				},
				TLS: &tls.ConnectionState{},
			},
			expected: &Header{
				MaxAge:            3000,
				IncludeSubDomains: false,
				Permanent:         false,
				Sha256Pins: []string{
					"d6qzRu9zOECb90Uez27xWltNsj0e1Md7GkYYkVoZWmM=",
					"E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g=",
				},
			},
		},
		// https://tools.ietf.org/html/rfc7469#section-2.1.5
		{
			name: "hpkp header (1)",
			response: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Public-Key-Pins": []string{`max-age=3000; pin-sha256="d6qzRu9zOECb90Uez27xWltNsj0e1Md7GkYYkVoZWmM="; pin-sha256="E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g="`},
				},
				TLS: &tls.ConnectionState{},
			},
			expected: nil,
		},
		{
			name: "hpkp header (4)",
			response: &http.Response{
				StatusCode: 200,
				Header: map[string][]string{
					"Public-Key-Pins-Report-Only": []string{`max-age=2592000; pin-sha256="E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g="; pin-sha256="LPJNul+wow4m6DsqxbninhsWHlwfp0JecwQzYpOLmCQ="; report-uri="http://example.com/pkp-report"`},
				},
				TLS: &tls.ConnectionState{},
			},
			expected: &Header{
				MaxAge:            2592000,
				IncludeSubDomains: false,
				Permanent:         false,
				Sha256Pins: []string{
					"E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g=",
					"LPJNul+wow4m6DsqxbninhsWHlwfp0JecwQzYpOLmCQ=",
				},
				ReportURI: "http://example.com/pkp-report",
			},
		},
	}

	for _, test := range tests {
		out := ParseReportOnlyHeader(test.response)
		if !equalHeaders(out, test.expected) {
			t.Logf("want:%v", test.expected)
			t.Logf("got:%v", out)
			t.Fatalf("test case failed: %s", test.name)
		}
	}
}

func equalHeaders(a, b *Header) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if a.IncludeSubDomains != b.IncludeSubDomains {
		return false
	}

	if a.MaxAge != b.MaxAge {
		return false
	}

	if a.Permanent != b.Permanent {
		return false
	}

	if a.ReportURI != b.ReportURI {
		return false
	}

	if len(a.Sha256Pins) != len(b.Sha256Pins) {
		return false
	}

	for i := range a.Sha256Pins {
		if a.Sha256Pins[i] != b.Sha256Pins[i] {
			return false
		}
	}

	return true
}
