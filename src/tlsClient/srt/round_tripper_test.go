package srt

import (
	tlsclient "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"strings"
	"testing"
)

func TestFile(t *testing.T) {
	tr, err := NewSpoofedRoundTripper(tlsclient.WithClientProfile(profiles.Chrome_120))
	assert.Nil(t, err)
	client := resty.New().SetTransport(tr)
	resp, err := client.R().SetFileReader("file", "file.txt", strings.NewReader("abc")).
		Post("https://httpbin.org/anything")
	assert.Nil(t, err)
	result := gjson.ParseBytes(resp.Body())
	assert.Equal(t, "abc", result.Get("files.file").String())
}
func TestRawBody(t *testing.T) {
	tr, err := NewSpoofedRoundTripper(tlsclient.WithClientProfile(profiles.Chrome_120))
	assert.Nil(t, err)
	client := resty.New().SetTransport(tr)
	for _, body := range []any{"abc", strings.NewReader("abc")} {
		resp, err := client.R().SetBody(body).Post("https://httpbin.org/anything")
		assert.Nil(t, err)
		result := gjson.ParseBytes(resp.Body())
		assert.Equal(t, "abc", result.Get("data").String())
	}
}
func TestQueryString(t *testing.T) {
	tr, err := NewSpoofedRoundTripper(tlsclient.WithClientProfile(profiles.Chrome_120))
	assert.Nil(t, err)
	client := resty.New().SetTransport(tr)
	resp, err := client.R().SetQueryParams(map[string]string{
		"a": "1",
		"b": "2",
	}).Post("https://httpbin.org/anything")
	assert.Nil(t, err)
	result := gjson.ParseBytes(resp.Body())
	assert.Equal(t, "1", result.Get("args.a").String())
	assert.Equal(t, "2", result.Get("args.b").String())
}
func TestFormData(t *testing.T) {
	tr, err := NewSpoofedRoundTripper(tlsclient.WithClientProfile(profiles.Chrome_120))
	assert.Nil(t, err)
	client := resty.New().SetTransport(tr)
	resp, err := client.R().SetFormData(map[string]string{
		"a": "1",
		"b": "2",
	}).Post("https://httpbin.org/anything")
	assert.Nil(t, err)
	result := gjson.ParseBytes(resp.Body())
	assert.Equal(t, "1", result.Get("form.a").String())
	assert.Equal(t, "2", result.Get("form.b").String())
}
func TestHeaders(t *testing.T) {
	tr, err := NewSpoofedRoundTripper(tlsclient.WithClientProfile(profiles.Chrome_120))
	assert.Nil(t, err)
	client := resty.New().SetTransport(tr)
	resp, err := client.R().SetHeaders(map[string]string{
		"User-Agent": "Chrome/120",
		"X-Custom":   "custom",
	}).Get("https://httpbin.org/anything")
	assert.Nil(t, err)
	result := gjson.ParseBytes(resp.Body())
	assert.Equal(t, "Chrome/120", result.Get("headers.User-Agent").String())
	assert.Equal(t, "custom", result.Get("headers.X-Custom").String())
}
func TestFingerprint(t *testing.T) {
	testFingerprint(t, profiles.Chrome_120, "1d9a054bac1eef41f30d370f9bbb2ad2",
		"t13d1516h2_8daaf6152771_b1ff8ab2d16f", "90224459f8bf70b7d0a8797eb916dbc9")
	testFingerprint(t, profiles.Chrome_103, "cd08e31494f9531f560d64c695473da9",
		"t13d1516h2_8daaf6152771_5fb3489db586", "7ad845f20fc17cc8088a0d9312b17da1")
	testFingerprint(t, profiles.Firefox_102, "579ccef312d18482fc42e2b822ca2430",
		"t13d1715h2_5b57614c22b0_5a7a167d0339", "fd4f649c50a64e33cc9e2407055bafbe")
}
func testFingerprint(t *testing.T, profile profiles.ClientProfile, ja3 string, ja4 string, akamai string) {
	tr, err := NewSpoofedRoundTripper(tlsclient.WithClientProfile(profile))
	assert.Nil(t, err)
	client := resty.New().SetTransport(tr)
	resp, err := client.R().Get("https://tls.peet.ws/api/all")
	assert.Nil(t, err)
	result := gjson.ParseBytes(resp.Body())
	assert.Equal(t, ja3, result.Get("tls.ja3_hash").String())
	assert.Equal(t, ja4, result.Get("tls.ja4").String())
	assert.Equal(t, akamai, result.Get("http2.akamai_fingerprint_hash").String())
}
