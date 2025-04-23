package hpkp

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"time"
)

// PinFailure hold fields required for POSTing a pin validation failure JSON message
// to a host's report-uri.
type PinFailure struct {
	DateTime                  string   `json:"date-time"`
	Hostname                  string   `json:"hostname"`
	Port                      int      `json:"port"`
	EffectiveExpirationDate   string   `json:"effective-expiration-date"`
	IncludeSubdomains         bool     `json:"include-subdomains"`
	NotedHostname             string   `json:"noted-hostname"`
	ServedCertificateChain    []string `json:"served-certificate-chain"`
	ValidatedCertificateChain []string `json:"validated-certificate-chain"`
	KnownPins                 []string `json:"known-pins"`
}

// NewPinFailure creates a struct to report information on failed hpkp connections
func NewPinFailure(host string, port int, h *Header, c tls.ConnectionState) (*PinFailure, string) {
	if h == nil {
		return nil, ""
	}

	verifiedChain := []*x509.Certificate{}
	if len(c.VerifiedChains) > 0 {
		verifiedChain = c.VerifiedChains[len(c.VerifiedChains)-1]
	}

	return &PinFailure{
		DateTime: time.Now().Format(time.RFC3339),
		Hostname: host,
		Port:     port,
		EffectiveExpirationDate:   time.Unix(h.Created+h.MaxAge, 0).UTC().Format(time.RFC3339),
		IncludeSubdomains:         h.IncludeSubDomains,
		NotedHostname:             c.ServerName,
		ServedCertificateChain:    encodeCertificatesPEM(c.PeerCertificates),
		ValidatedCertificateChain: encodeCertificatesPEM(verifiedChain),
		KnownPins:                 h.Sha256Pins,
	}, h.ReportURI
}

// encodeCertificatesPEM converts a slice of x509 certficates to a slice of PEM encoded strings
func encodeCertificatesPEM(certs []*x509.Certificate) []string {
	var pemCerts []string

	var buffer bytes.Buffer
	for _, cert := range certs {
		pem.Encode(&buffer, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		})
		pemCerts = append(pemCerts, string(buffer.Bytes()))
		buffer.Reset()
	}

	return pemCerts
}
