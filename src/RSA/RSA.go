package RSA

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
)

// Rsa2PubVerifySign RSA2公钥验证签名
func Rsa2PubVerifySign(signContent, sign []byte, publicKey *rsa.PublicKey, hash crypto.Hash) bool {
	h := hash.New()
	h.Write(signContent)
	hashed := h.Sum(nil)
	err := rsa.VerifyPKCS1v15(publicKey, hash, hashed[:], sign)
	if err != nil {
		return false
	}
	return true
}

// RsaPrivateSign RSA2私钥签名
func RsaPrivateSign(sign, Ciphertext []byte, publicKey *rsa.PrivateKey, hash crypto.Hash) bool {
	h := hash.New()
	h.Write(sign)
	hashed := h.Sum(nil)
	Ret, err := rsa.SignPKCS1v15(rand.Reader, publicKey, hash, hashed)
	if err != nil {
		return false
	}
	if bytes.Equal(Ret, Ciphertext) {
		return true
	}
	return false
}
func FormatRSAPrivateKey(key []byte) string {
	arr := strings.Split(strings.ReplaceAll(strings.TrimSpace(string(key)), "\r", ""), "\n")
	o := ""
	for _, str := range arr {
		if strings.HasSuffix(str, "----") {
			if strings.Contains(str, "---END") {
				break
			}
			continue
		}
		o += str
	}
	return o
}

// RsaPubKeyEncrypt 公钥加密 [请使用私钥解密]
func RsaPubKeyEncrypt(pemKey, data []byte) []byte {
	_key_ := ParseKey(pemKey)
	if _key_ == nil {
		return nil
	}
	key, ok := _key_.(*rsa.PublicKey)
	if !ok {
		return nil
	}
	output := bytes.NewBuffer(nil)
	err, _ := PubKeyIO(key, bytes.NewReader(data), output, true)
	if err != nil {
		return nil
	}
	return output.Bytes()
}

// RsaPriKeyDecrypt 私钥解密 [请使用公钥加密]
func RsaPriKeyDecrypt(pemKey, data []byte) []byte {
	_key_ := ParseKey(pemKey)
	if _key_ == nil {
		return nil
	}
	key, ok := _key_.(*rsa.PrivateKey)
	if !ok {
		return nil
	}
	output := bytes.NewBuffer(nil)
	err, _ := PriKeyIO(key, bytes.NewReader(data), output, false)
	if err != nil {
		return nil
	}
	return output.Bytes()
}

// RsaPubKeyDecrypt 公钥解密 [请使用私钥加密]
func RsaPubKeyDecrypt(pemKey, data []byte) []byte {
	_key_ := ParseKey(pemKey)
	if _key_ == nil {
		return nil
	}
	key, ok := _key_.(*rsa.PublicKey)
	if !ok {
		return nil
	}
	output := bytes.NewBuffer(nil)
	err, _ := PubKeyIO(key, bytes.NewReader(data), output, false)
	if err != nil {
		return nil
	}
	return output.Bytes()
}

// RsaPriKeyEncrypt 私钥加密 [请使用公钥解密]
func RsaPriKeyEncrypt(pemKey, data []byte) []byte {
	_key_ := ParseKey(pemKey)
	if _key_ == nil {
		return nil
	}
	key, ok := _key_.(*rsa.PrivateKey)
	if !ok {
		return nil
	}
	output := bytes.NewBuffer(nil)
	err, _ := PriKeyIO(key, bytes.NewReader(data), output, true)
	if err != nil {
		return nil
	}
	return output.Bytes()
}

// ParseKey 解析 PEM 或 DER 格式的密钥
func ParseKey(keyBytes []byte) any {
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		// 尝试解析为 DER 格式
		return parseDERKey(keyBytes)
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		p, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err == nil {
			return p
		}
		return nil
	case "PRIVATE KEY":
		p, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err == nil {
			return p
		}
		return nil
	case "RSA PUBLIC KEY":
		p, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err == nil {
			return p
		}
		return nil
	case "PUBLIC KEY":
		p, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err == nil {
			return p
		}
		return nil
	default:
		return nil
	}
}

// parseDERKey 解析 DER 格式的密钥
func parseDERKey(keyBytes []byte) any {
	// 尝试直接解析 DER 格式
	if p, err := x509.ParsePKCS1PrivateKey(keyBytes); err == nil {
		return p
	}
	if p, err := x509.ParsePKCS8PrivateKey(keyBytes); err == nil {
		return p
	}
	if p, err := x509.ParsePKCS1PublicKey(keyBytes); err == nil {
		return p
	}
	if p, err := x509.ParsePKIXPublicKey(keyBytes); err == nil {
		return p
	}

	// 尝试将 Base64 解码为 DER 格式
	decodedBytes, err := base64.StdEncoding.DecodeString(string(keyBytes))
	if err != nil {
		return nil
	}
	if p, err1 := x509.ParsePKCS1PrivateKey(decodedBytes); err1 == nil {
		return p
	}
	if p, err1 := x509.ParsePKCS8PrivateKey(decodedBytes); err1 == nil {
		return p
	}
	if p, err1 := x509.ParsePKCS1PublicKey(decodedBytes); err1 == nil {
		return p
	}
	if p, err1 := x509.ParsePKIXPublicKey(decodedBytes); err1 == nil {
		return p
	}

	return nil
}
