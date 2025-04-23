package hpkp

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
)

// Fingerprint returns the hpkp signature of an x509 certificate
func Fingerprint(c *x509.Certificate) string {
	digest := sha256.Sum256(c.RawSubjectPublicKeyInfo)
	return base64.StdEncoding.EncodeToString(digest[:])
}
