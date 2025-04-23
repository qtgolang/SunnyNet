package http2_test

import (
	"bytes"
	gtls "crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	ghttp "net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/qtgolang/SunnyNet/src/http/cookiejar"
	"github.com/qtgolang/SunnyNet/src/http/httptest"
	"github.com/qtgolang/SunnyNet/src/crypto/tls"
	"golang.org/x/net/publicsuffix"

	 "github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/http/http2"
)

// Tests if connection settings are written correctly
func TestConnectionSettings(t *testing.T) {
	settings := []http2.Setting{
		{ID: http2.SettingHeaderTableSize, Val: 65536},
		{ID: http2.SettingMaxConcurrentStreams, Val: 1000},
		{ID: http2.SettingInitialWindowSize, Val: 6291456},
		{ID: http2.SettingMaxFrameSize, Val: 16384},
		{ID: http2.SettingMaxHeaderListSize, Val: 262144},
	}
	buf := new(bytes.Buffer)
	fr := http2.NewFramer(buf, buf)
	err := fr.WriteSettings(settings...)

	if err != nil {
		t.Fatalf(err.Error())
	}

	f, err := fr.ReadFrame()
	if err != nil {
		t.Fatal(err.Error())
	}

	sf := f.(*http2.SettingsFrame)
	n := sf.NumSettings()
	if n != len(settings) {
		t.Fatalf("Expected %d settings, got %d", len(settings), n)
	}

	for i := 0; i < n; i++ {
		s := sf.Setting(i)
		var err error
		switch s.ID {
		case http2.SettingHeaderTableSize:
			err = compareSettings(s.ID, s.Val, 65536)
		case http2.SettingMaxConcurrentStreams:
			err = compareSettings(s.ID, s.Val, 1000)
		case http2.SettingInitialWindowSize:
			err = compareSettings(s.ID, s.Val, 6291456)
		case http2.SettingMaxFrameSize:
			err = compareSettings(s.ID, s.Val, 16384)
		case http2.SettingMaxHeaderListSize:
			err = compareSettings(s.ID, s.Val, 262144)
		}

		if err != nil {
			t.Fatal(err.Error())
		}
	}
}

func compareSettings(ID http2.SettingID, output uint32, expected uint32) error {
	if output != expected {
		return errors.New(fmt.Sprintf("Setting %v, expected %d got %d", ID, expected, output))
	}
	return nil
}

// Round trip test, makes sure that the changes made doesn't break the library
func TestRoundTrip(t *testing.T) {
	settings := map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxFrameSize:         16384,
		http2.SettingMaxHeaderListSize:    262144,
	}

	settingsOrder := []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxFrameSize,
		http2.SettingMaxHeaderListSize,
	}

	tr := http2.Transport{
		Settings:      settings,
		SettingsOrder: settingsOrder,
	}
	req, err := http.NewRequest("GET", "www.google.com", nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	tr.RoundTrip(req)
}

// Tests if content-length header is present in request headers during POST
func TestContentLength(t *testing.T) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hdr, ok := r.Header["Content-Length"]; ok {
			if len(hdr) != 1 {
				t.Fatalf("Got %v content-length headers, should only be 1", len(hdr))
			}
			return
		}
		log.Printf("Proto: %v", r.Proto)
		for name, value := range r.Header {
			log.Printf("%v: %v", name, value)
		}
		t.Fatalf("Could not find content-length header")
	}))
	ts.EnableHTTP2 = true
	ts.StartTLS()
	defer ts.Close()

	u := ts.URL
	form := url.Values{}
	form.Add("Hello", "World")
	req, err := http.NewRequest("POST", u, strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf(err.Error())
	}
	req.Header.Add("user-agent", "Go Testing")

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer resp.Body.Close()
}

// TestClient_Cookies tests whether set cookies are being sent
func TestClient_SendsCookies(t *testing.T) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("cookie")
		if err != nil {
			t.Fatalf(err.Error())
		}
		if cookie.Value == "" {
			t.Fatalf("Cookie value is empty")
		}
	}))
	ts.EnableHTTP2 = true
	ts.StartTLS()
	defer ts.Close()
	c := ts.Client()
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
	c.Jar = jar
	ur := ts.URL
	u, err := url.Parse(ur)
	if err != nil {
		t.Fatalf(err.Error())
	}
	cookies := []*http.Cookie{{Name: "cookie", Value: "Hello world"}}
	jar.SetCookies(u, cookies)
	resp, err := c.Get(ur)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer resp.Body.Close()
}

// TestClient_Load is a dumb man's load test with charles :P
func TestClient_Load(t *testing.T) {
	u, err := url.Parse("http://localhost:8888")
	if err != nil {
		t.Fatalf(err.Error())
	}

	pool, err := getCharlesCert()
	if err != nil {
		t.Fatalf(err.Error())
	}
	c := http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2: true,
			Proxy:             http.ProxyURL(u),
			TLSClientConfig: &tls.Config{
				RootCAs: pool,
			},
		},
	}
	req, err := http.NewRequest("GET", "https://golang.org/pkg/net/mail/#Address", nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	for i := 0; i < 10; i++ {
		resp, err := c.Do(req)
		if err != nil {
			t.Fatalf(err.Error())
		}
		resp.Body.Close()
	}
}

func TestGClient_Load(t *testing.T) {
	u, err := url.Parse("http://localhost:8888")
	if err != nil {
		t.Fatalf(err.Error())
	}

	pool, err := getCharlesCert()
	if err != nil {
		t.Fatalf(err.Error())
	}
	c := ghttp.Client{
		Transport: &ghttp.Transport{
			ForceAttemptHTTP2: true,
			Proxy:             ghttp.ProxyURL(u),
			TLSClientConfig: &gtls.Config{
				RootCAs: pool,
			},
		},
	}
	req, err := ghttp.NewRequest("GET", "https://golang.org/pkg/net/mail/#Address", nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	for i := 0; i < 10; i++ {
		err := do(&c, req)
		if err != nil {
			t.Fatalf(err.Error())
		}
	}
}

func do(c *ghttp.Client, req *ghttp.Request) error {
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func getCharlesCert() (*x509.CertPool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	caCert, err := os.ReadFile(fmt.Sprintf("%v/charles_cert.pem", home))
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)
	return certPool, nil
}
