package HttpCertificate

import (
	"crypto/x509"
	"github.com/qtgolang/SunnyNet/src/crypto/tls"
	"github.com/qtgolang/SunnyNet/src/public"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

var Lock sync.Mutex
var Map = make(map[string]*CertificateRequestManager)

type CertificateRequestManager struct {
	Rules  uint8
	Config *tls.Config
}

func (w *CertificateRequestManager) AddClientCAs(ClientCAs *x509.CertPool) {
	if w.Config == nil {
		w.Config = &tls.Config{ClientCAs: ClientCAs}
	} else {
		w.Config.ClientCAs = ClientCAs
	}
}
func (w *CertificateRequestManager) Load(ca, key string) bool {
	s, e := tls.X509KeyPair([]byte(ca), []byte(key))
	if e != nil {
		return false
	}
	if w.Config == nil {
		w.Config = &tls.Config{}

	}
	w.Config.Certificates = []tls.Certificate{s}
	return true
}
func GetTlsConfig(host string, Rules uint8) *tls.Config {
	if host == "" || host == "null" {
		return nil
	}
	RequestHost := ParsingHost(host)
	Lock.Lock()
	defer Lock.Unlock()
	for RulesHost, v := range Map {
		if v.Rules == Rules || v.Rules == public.CertificateRequestManagerRulesSendAndReceive {
			pattern := strings.ReplaceAll(strings.ReplaceAll(RulesHost, ".", "\\."), "*", ".*")
			re := regexp.MustCompile(pattern)
			if re.MatchString(RequestHost) {
				return v.Config
			}
		}
	}
	return nil
}
func ParsingHost(host string) string {
	m := host
	if len(m) < 6 {
		return host
	}
	if !strings.HasPrefix(m, "http:") {
		if !strings.HasPrefix(m, "https:") {
			m = "https://" + m
		}
	}
	a, b := url.Parse(m)
	if b != nil {
		return host
	}
	return a.Hostname()
}
