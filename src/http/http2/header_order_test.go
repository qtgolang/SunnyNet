package http2

import (
	"bytes"
	"log"
	"strings"
	"testing"

	 "github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/http/httptrace"
)

func TestHeaderOrder(t *testing.T) {
	req, err := http.NewRequest("POST", "https://www.httpbin.org/headers", nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	req.Header = http.Header{
		"sec-ch-ua":        {"\" Not;A Brand\";v=\"99\", \"Google Chrome\";v=\"91\", \"Chromium\";v=\"91\""},
		"accept":           {"*/*"},
		"x-requested-with": {"XMLHttpRequest"},
		"sec-ch-ua-mobile": {"?0"},
		"user-agent":       {"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36\", \"I shouldn't be here"},
		"content-type":     {"application/json"},
		"origin":           {"https://www.size.co.uk/"},
		"sec-fetch-site":   {"same-origin"},
		"sec-fetch-mode":   {"cors"},
		"sec-fetch-dest":   {"empty"},
		"accept-language":  {"en-US,en;q=0.9"},
		"accept-encoding":  {"gzip, deflate, br"},
		"referer":          {"https://www.size.co.uk/product/white-jordan-air-1-retro-high/16077886/"},
		http.HeaderOrderKey: {
			"sec-ch-ua",
			"accept",
			"x-requested-with",
			"sec-ch-ua-mobile",
			"user-agent",
			"content-type",
			"sec-fetch-site",
			"sec-fetch-mode",
			"sec-fetch-dest",
			"referer",
			"accept-encoding",
			"accept-language",
		},
		http.PHeaderOrderKey: {
			":method",
			":authority",
			":scheme",
			":path",
		},
	}
	var b []byte
	buf := bytes.NewBuffer(b)
	err = req.Header.Write(buf)
	if err != nil {
		t.Fatalf(err.Error())
	}
	arr := strings.Split(buf.String(), "\n")
	var hdrs []string
	for _, v := range arr {
		a := strings.Split(v, ":")
		if a[0] == "" {
			continue
		}
		hdrs = append(hdrs, a[0])
	}

	for i := range req.Header[http.HeaderOrderKey] {
		if hdrs[i] != req.Header[http.HeaderOrderKey][i] {
			t.Errorf("want: %s\ngot: %s\n", req.Header[http.HeaderOrderKey][i], hdrs[i])
		}
	}
}

func TestHeaderOrder2(t *testing.T) {
	hk := ""
	trace := &httptrace.ClientTrace{
		WroteHeaderField: func(key string, values []string) {
			hk += key + " "
		},
	}
	req, err := http.NewRequest("GET", "https://httpbin.org/#/Request_inspection/get_headers", nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	req.Header.Add("experience", "pain")
	req.Header.Add("grind", "harder")
	req.Header.Add("live", "mas")
	req.Header[http.HeaderOrderKey] = []string{"grind", "experience", "live"}
	req.Header[http.PHeaderOrderKey] = []string{":method", ":authority", ":scheme", ":path"}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	tr := &Transport{}
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer resp.Body.Close()

	eq := strings.EqualFold(hk, ":method :authority :scheme :path grind experience live accept-encoding user-agent ")
	if !eq {
		t.Fatalf("Header order not set properly, \n Got %v \n Want: %v", hk, ":method :authority :scheme :path grind experience live accept-encoding user-agent")
	}
}

func TestHeaderOrder3(t *testing.T) {
	req, err := http.NewRequest("GET", "https://google.com", nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	req.Header = http.Header{
		http.HeaderOrderKey: {
			"sec-ch-ua",
			"accept",
			"x-requested-with",
			"sec-ch-ua-mobile",
			"user-agent",
			"sec-fetch-site",
			"sec-fetch-mode",
			"sec-fetch-dest",
			"referer",
			"accept-encoding",
			"accept-language",
			"cookie",
		},
	}
	req.Header.Add("accept", "text / html, application/xhtml + xml, application / xml;q = 0.9, image/avif, image/webp, image/apng, * /*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Add("accept-encoding", "gzip, deflate, br")
	req.Header.Add("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("pragma", "no-cache")
	req.Header.Add("referer", "https://www.offspring.co.uk/")
	req.Header.Add("sec-ch-ua", `" Not A;Brand"; v = "99", "Chromium"; v = "90", "Google Chrome"; v = "90"`)
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-fetch-dest", "document")
	req.Header.Add("sec-fetch-mode", "navigate")
	req.Header.Add("sec-fetch-site", "same-origin")
	req.Header.Add("sec-fetch-user", "?1")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("user-agent", "Mozilla")
	var hdrs string
	trace := &httptrace.ClientTrace{WroteHeaderField: func(key string, value []string) {
		hdrs += key + "\n"
	}}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	tr := Transport{}
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer resp.Body.Close()
	log.Println(hdrs)
}
