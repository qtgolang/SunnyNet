package SunnyNet

import (
	"crypto/x509"
	"github.com/qtgolang/SunnyNet/src/Certificate"
	"github.com/qtgolang/SunnyNet/src/HttpCertificate"
	"github.com/qtgolang/SunnyNet/src/crypto/tls"
)

const (
	//HTTPCertRules_Request 仅发送使用
	HTTPCertRules_Request = 1
	//HTTPCertRules_ResponseAndRequest 发送和解析都使用
	HTTPCertRules_ResponseAndRequest = 2
	//HTTPCertRules_Response 仅解析使用
	HTTPCertRules_Response = 3
)

func NewCertManager() *Certificate.CertManager {
	temp := &Certificate.CertManager{Tls: &tls.Config{}}
	temp.SetInsecureSkipVerify(true)
	return temp
}

// AddHttpCertificate 指定Host使用指定证书
func (s *Sunny) AddHttpCertificate(host string, Cert *Certificate.CertManager, Rules uint8) bool {
	HttpCertificate.Lock.Lock()
	defer HttpCertificate.Lock.Unlock()
	if Cert == nil {
		return false
	}
	ca := Cert.ExportCA()
	key := Cert.ExportKEY()
	cart := Cert.Cert
	var ClientCAs *x509.CertPool
	if Cert.Tls != nil {
		if Cert.Tls.ClientCAs != nil {
			ClientCAs = Cert.Tls.ClientCAs
		}
	}
	if (ca == "" || key == "") && cart == "" && ClientCAs != nil {
		c := &HttpCertificate.CertificateRequestManager{Rules: Rules}
		c.AddClientCAs(ClientCAs)
		HttpCertificate.Map[HttpCertificate.ParsingHost(host)] = c
		return true
	}
	if ca == "" && key == "" && cart == "" {
		return false
	}
	c := &HttpCertificate.CertificateRequestManager{Rules: Rules}
	if c.Load(ca, key) {
		c.AddClientCAs(ClientCAs)
		HttpCertificate.Map[HttpCertificate.ParsingHost(host)] = c
		return true
	}
	if len(Cert.Cert) > 1 {
		if c.Load(Cert.Cert, Cert.Cert) {
			HttpCertificate.Map[HttpCertificate.ParsingHost(host)] = c
			return true
		}
	}
	return false
}

// DelHttpCertificate 删除指定Host使用指定证书
func (s *Sunny) DelHttpCertificate(host string) {
	HttpCertificate.Lock.Lock()
	delete(HttpCertificate.Map, HttpCertificate.ParsingHost(host))
	HttpCertificate.Lock.Unlock()
}
