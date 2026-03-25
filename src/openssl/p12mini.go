package openssl

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

type p12Parsed struct {
	keyDER  []byte
	certDER []byte
	caDER   [][]byte
}

type p12miniParser struct {
	lastErr string
}

// setErr 设置错误消息（会覆盖旧值）。
func (p *p12miniParser) setErr(msg string) {
	if msg == "" {
		msg = "unknown error"
	}
	p.lastErr = msg
}

// setErrOnce 仅在尚未设置错误时设置错误消息（避免覆盖更具体的底层错误）。
func (p *p12miniParser) setErrOnce(msg string) {
	if p.lastErr == "" {
		p.setErr(msg)
	}
}

// bytesToHexPrefix 将前 n 个字节转成十六进制字符串（用于错误提示的前缀展示）。
func bytesToHexPrefix(b []byte, n int) string {
	if n <= 0 {
		return ""
	}
	if len(b) < n {
		n = len(b)
	}
	return fmt.Sprintf("%x", b[:n])
}

// oidEq 判断 OID 的原始内容字节是否与给定 OID der 内容一致（这里不解码 OID，只比对内容）。
func oidEq(oid Asn1View, der []byte) bool {
	return len(oid.b) == len(der) && bytes.Equal(oid.b, der)
}

var (
	OID_pkcs7_data         = []byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x07, 0x01}
	OID_pkcs7_encrypted    = []byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x07, 0x06}
	OID_bag_shroudedKey    = []byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x0c, 0x0a, 0x01, 0x02}
	OID_bag_certBag        = []byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x0c, 0x0a, 0x01, 0x03}
	OID_bag_safeContents   = []byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x0c, 0x0a, 0x01, 0x06}
	OID_cert_x509          = []byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x09, 0x16, 0x01}
	OID_pbe_sha1_rc2_40    = []byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x0c, 0x01, 0x06}
	OID_pbe_sha1_rc2_128   = []byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x0c, 0x01, 0x05}
	OID_pbe_sha1_3des_3key = []byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x0c, 0x01, 0x03}
	OID_pbe_sha1_3des_2key = []byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x0c, 0x01, 0x04}
)

// appendCA 追加一张 CA/链证书（DER），内部会拷贝一份避免外部引用变化。
func appendCA(out *p12Parsed, der []byte) {
	clone := make([]byte, len(der))
	copy(clone, der)
	out.caDER = append(out.caDER, clone)
}

// readOctetStringFlatten 读取 OCTET STRING；若是构造型（constructed），会把内部的若干 primitive OCTET STRING 拼接成一段连续字节。
// 这是为了兼容某些 PKCS#12 文件中使用的 “constructed OCTET STRING + BER indefinite length” 编码。
func readOctetStringFlatten(in *Asn1View) ([]byte, bool) {
	tag, content, ok := asn1ReadTlv(in)
	if !ok {
		return nil, false
	}
	if tag.cls != ASN1_CLASS_UNIVERSAL || tag.tagnum != 4 {
		return nil, false
	}
	if !tag.constructed {
		out := make([]byte, len(content.b))
		copy(out, content.b)
		return out, true
	}

	// Constructed OCTET STRING: concatenate inner primitive OCTET STRINGs.
	t := content
	var out []byte
	for len(t.b) > 0 {
		itag, icontent, ok := asn1ReadTlv(&t)
		if !ok {
			return nil, false
		}
		if itag.cls != ASN1_CLASS_UNIVERSAL || itag.tagnum != 4 || itag.constructed {
			return nil, false
		}
		out = append(out, icontent.b...)
	}
	return out, true
}

// readCtxImplicitOctetStringFlatten 读取 context-specific [tagnum] 的隐式 OCTET STRING。
// 若该字段是构造型（constructed），会拼接内部 primitive OCTET STRING。
func readCtxImplicitOctetStringFlatten(in *Asn1View, tagnum uint32) ([]byte, bool) {
	tag, content, ok := asn1ReadTlv(in)
	if !ok {
		return nil, false
	}
	if tag.cls != ASN1_CLASS_CONTEXT || tag.tagnum != tagnum {
		return nil, false
	}

	if !tag.constructed {
		out := make([]byte, len(content.b))
		copy(out, content.b)
		return out, true
	}

	// Constructed: concatenate inner primitive UNIVERSAL OCTET STRINGs.
	t := content
	var out []byte
	for len(t.b) > 0 {
		itag, icontent, ok := asn1ReadTlv(&t)
		if !ok {
			return nil, false
		}
		if itag.cls != ASN1_CLASS_UNIVERSAL || itag.tagnum != 4 || itag.constructed {
			return nil, false
		}
		out = append(out, icontent.b...)
	}
	return out, true
}

// parsePbeParams 解析传统 PKCS#12 PBE 参数：SEQUENCE { salt OCTET STRING, iter INTEGER }。
func parsePbeParams(paramsAny Asn1View) (salt []byte, iter int, ok bool) {
	seq, ok := asn1ReadSequence(&paramsAny)
	if !ok {
		return nil, 0, false
	}
	t := seq
	saltView, ok := asn1ReadOctetString(&t)
	if !ok {
		return nil, 0, false
	}
	iterVal, ok := asn1ReadInt(&t)
	if !ok {
		return nil, 0, false
	}
	return saltView.b, iterVal, true
}

// pkcs7Unpad 对指定 blockSize 做 PKCS#7 去填充。
func pkcs7Unpad(buf []byte, blockSize int) ([]byte, bool) {
	if len(buf) == 0 {
		return nil, false
	}
	pad := int(buf[len(buf)-1])
	if pad <= 0 || pad > blockSize || pad > len(buf) {
		return nil, false
	}
	for i := 0; i < pad; i++ {
		if int(buf[len(buf)-1-i]) != pad {
			return nil, false
		}
	}
	return buf[:len(buf)-pad], true
}

// des3CBCDecryptPKCS7 使用 3DES-CBC 解密并做 PKCS#7 去填充。
// 兼容 2-key (16 字节) 与 3-key (24 字节) 两种 3DES key 形态。
func des3CBCDecryptPKCS7(ciphertext []byte, key []byte, iv []byte) ([]byte, bool) {
	if len(ciphertext) == 0 || len(ciphertext)%8 != 0 || len(iv) < 8 {
		return nil, false
	}
	if !(len(key) == 16 || len(key) == 24) {
		return nil, false
	}
	key24 := make([]byte, 24)
	if len(key) == 24 {
		copy(key24, key)
	} else {
		copy(key24[0:16], key)
		copy(key24[16:24], key[0:8]) // 2-key: K1||K2||K1
	}

	block, err := des.NewTripleDESCipher(key24)
	if err != nil {
		return nil, false
	}
	raw := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv[:8])
	mode.CryptBlocks(raw, ciphertext)
	return pkcs7Unpad(raw, 8)
}

// decryptPkcs12PbeSha1Rc2 解密 PKCS#12 传统 PBE: SHA1 + RC2-CBC（支持 RC2-40/RC2-128）。
func (p *p12miniParser) decryptPkcs12PbeSha1Rc2(password string, salt []byte, iter int, effectiveBits int, ct []byte) ([]byte, bool) {
	key := pkcs12KdfSHA1(password, salt, 1, iter, (effectiveBits+7)/16) // 40->5, 128->8? wait
	// RC2 key size:
	// - 40-bit RC2: 5 bytes
	// - 128-bit RC2: 16 bytes (effective bits 128)
	keyLen := 5
	if effectiveBits > 40 {
		keyLen = 16
	}
	key = pkcs12KdfSHA1(password, salt, 1, iter, keyLen)
	iv := pkcs12KdfSHA1(password, salt, 2, iter, 8)

	pt, ok := rc2CBCDecryptPKCS7(ct, key, effectiveBits, iv)
	if ok {
		return pt, true
	}

	// If padding failed, mirror debug behavior from C for RC2-40.
	raw, rawOK := rc2CBCDecryptRaw(ct, key, effectiveBits, iv)
	if rawOK && effectiveBits == 40 {
		p.setErr(fmt.Sprintf("RC2-40 decrypt produced bad padding (first16=%s)", bytesToHexPrefix(raw, 16)))
	}
	return nil, false
}

// decryptPkcs12PbeSha13des 解密 PKCS#12 传统 PBE: SHA1 + 2/3-key 3DES-CBC。
func (p *p12miniParser) decryptPkcs12PbeSha13des(password string, salt []byte, iter int, keyLen int, ct []byte) ([]byte, bool) {
	if keyLen != 16 && keyLen != 24 {
		return nil, false
	}
	key := pkcs12KdfSHA1(password, salt, 1, iter, keyLen)
	iv := pkcs12KdfSHA1(password, salt, 2, iter, 8)
	return des3CBCDecryptPKCS7(ct, key, iv)
}

// decryptAlgid 根据 AlgorithmIdentifier（OID + 参数）选择对应 PBE 解密算法，输出明文。
func (p *p12miniParser) decryptAlgid(algidAny Asn1View, password string, ct []byte) ([]byte, bool) {
	algidSeq, ok := asn1ReadSequence(&algidAny)
	if !ok {
		p.setErr("AlgorithmIdentifier is not a SEQUENCE")
		return nil, false
	}
	t := algidSeq
	oid, ok := asn1ReadOID(&t)
	if !ok {
		p.setErr("AlgorithmIdentifier missing OID")
		return nil, false
	}

	switch {
	case oidEq(oid, OID_pbe_sha1_rc2_40):
		salt, iter, ok := parsePbeParams(t)
		if !ok {
			p.setErr("PBE params parse failed")
			return nil, false
		}
		pt, ok := p.decryptPkcs12PbeSha1Rc2(password, salt, iter, 40, ct)
		if !ok {
			p.setErrOnce("PBE decrypt failed (wrong password or unsupported PBE/cipher?)")
			return nil, false
		}
		return pt, true

	case oidEq(oid, OID_pbe_sha1_rc2_128):
		salt, iter, ok := parsePbeParams(t)
		if !ok {
			p.setErr("PBE params parse failed")
			return nil, false
		}
		pt, ok := p.decryptPkcs12PbeSha1Rc2(password, salt, iter, 128, ct)
		if !ok {
			p.setErrOnce("PBE decrypt failed (wrong password or unsupported PBE/cipher?)")
			return nil, false
		}
		return pt, true

	case oidEq(oid, OID_pbe_sha1_3des_3key):
		salt, iter, ok := parsePbeParams(t)
		if !ok {
			p.setErr("PBE params parse failed")
			return nil, false
		}
		pt, ok := p.decryptPkcs12PbeSha13des(password, salt, iter, 24, ct)
		if !ok {
			p.setErrOnce("PBE decrypt failed (wrong password or unsupported PBE/cipher?)")
			return nil, false
		}
		return pt, true

	case oidEq(oid, OID_pbe_sha1_3des_2key):
		salt, iter, ok := parsePbeParams(t)
		if !ok {
			p.setErr("PBE params parse failed")
			return nil, false
		}
		pt, ok := p.decryptPkcs12PbeSha13des(password, salt, iter, 16, ct)
		if !ok {
			p.setErrOnce("PBE decrypt failed (wrong password or unsupported PBE/cipher?)")
			return nil, false
		}
		return pt, true
	}

	p.setErr(fmt.Sprintf("unsupported encryption algorithm OID (%s)", bytesToHexPrefix(oid.b, len(oid.b))))
	return nil, false
}

// parseContentInfoSeq 解析 PKCS#7 ContentInfo: SEQUENCE { contentType OID, [0] EXPLICIT content }。
func parseContentInfoSeq(ciSeqContent Asn1View) (oid Asn1View, any0 Asn1View, ok bool) {
	t := ciSeqContent
	oid, ok = asn1ReadOID(&t)
	if !ok {
		return Asn1View{}, Asn1View{}, false
	}
	any0, ok = asn1ReadCtxExplicit(&t, 0)
	if !ok {
		return Asn1View{}, Asn1View{}, false
	}
	return oid, any0, true
}

// parseSafecontents 解析 SafeContents: SEQUENCE OF SafeBag，并把私钥/证书提取到 out。
func (p *p12miniParser) parseSafecontents(safecontentsDer Asn1View, password string, out *p12Parsed) bool {
	// Need failure prefix from the original input.
	orig := safecontentsDer
	seq, ok := asn1ReadSequence(&safecontentsDer)
	if !ok {
		p.setErr(fmt.Sprintf("SafeContents not SEQUENCE (first16=%s)", bytesToHexPrefix(orig.b, 16)))
		return false
	}

	t := seq
	for len(t.b) > 0 {
		if !p.parseSafebag(&t, password, out) {
			return false
		}
	}
	return true
}

// parseSafebag 解析单个 SafeBag，并按 bag 类型处理：
// - certBag：提取证书
// - pkcs8ShroudedKeyBag：解密并提取私钥（PKCS#8 DER）
// - safeContentsBag：递归解析
func (p *p12miniParser) parseSafebag(in *Asn1View, password string, out *p12Parsed) bool {
	orig := *in
	bagSeq, ok := asn1ReadSequence(in)
	if !ok {
		p.setErr(fmt.Sprintf("SafeBag not SEQUENCE (first16=%s)", bytesToHexPrefix(orig.b, 16)))
		return false
	}
	t := bagSeq
	bagOID, ok := asn1ReadOID(&t)
	if !ok {
		p.setErr("SafeBag missing bagId OID")
		return false
	}

	bagValue, ok := asn1ReadCtxExplicit(&t, 0)
	if !ok {
		p.setErr(fmt.Sprintf("SafeBag missing [0] value (bagId=%s)", bytesToHexPrefix(bagOID.b, len(bagOID.b))))
		return false
	}

	switch {
	case oidEq(bagOID, OID_bag_certBag):
		certbagSeq, ok := asn1ReadSequence(&bagValue)
		if !ok {
			p.setErr("CertBag not SEQUENCE")
			return false
		}
		ct := certbagSeq
		certid, ok := asn1ReadOID(&ct)
		if !ok {
			p.setErr("CertBag missing certId OID")
			return false
		}
		if !oidEq(certid, OID_cert_x509) {
			p.setErr("unsupported CertBag certId (not x509Certificate)")
			return false
		}
		certvalExp, ok := asn1ReadCtxExplicit(&ct, 0)
		if !ok {
			p.setErr("CertBag missing [0] certValue")
			return false
		}
		certOct, ok := asn1ReadOctetString(&certvalExp)
		if !ok {
			p.setErr("CertBag certValue is not OCTET STRING")
			return false
		}

		if len(out.certDER) == 0 {
			out.certDER = append([]byte(nil), certOct.b...)
		} else {
			appendCA(out, certOct.b)
		}
		return true

	case oidEq(bagOID, OID_bag_safeContents):
		sc := bagValue
		if !p.parseSafecontents(sc, password, out) {
			p.setErr("nested SafeContentsBag parse failed")
			return false
		}
		return true

	case oidEq(bagOID, OID_bag_shroudedKey):
		epki, ok := asn1ReadSequence(&bagValue)
		if !ok {
			p.setErr("EncryptedPrivateKeyInfo not SEQUENCE")
			return false
		}
		ep := epki
		algidTlv, ok := asn1DupNextTlv(&ep)
		if !ok {
			return false
		}
		encBytes, ok := readOctetStringFlatten(&ep)
		if !ok {
			return false
		}

		pt, ok := p.decryptAlgid(Asn1View{b: algidTlv}, password, encBytes)
		if !ok {
			return false
		}

		if len(out.keyDER) == 0 {
			out.keyDER = pt
		}
		return true
	}

	// Unknown bag type: ignore.
	return true
}

// parseEncrypteddataToPlain 解析 PKCS#7 EncryptedData，并使用传统 PBE 解密得到明文 SafeContents DER。
func (p *p12miniParser) parseEncrypteddataToPlain(password string, encrypteddataAny Asn1View) ([]byte, bool) {
	orig := encrypteddataAny
	edSeq, ok := asn1ReadSequence(&encrypteddataAny)
	if !ok {
		p.setErr("EncryptedData not SEQUENCE")
		return nil, false
	}
	t := edSeq
	ver, ok := asn1ReadInt(&t)
	if !ok {
		p.setErr("EncryptedData missing version")
		return nil, false
	}
	_ = ver

	eci, ok := asn1ReadSequence(&t)
	if !ok {
		p.setErr("EncryptedData missing encryptedContentInfo")
		return nil, false
	}

	e := eci
	_, ok = asn1ReadOID(&e) // contentType OID ignored
	if !ok {
		p.setErr("encryptedContentInfo missing contentType OID")
		return nil, false
	}

	algidTlv, ok := asn1DupNextTlv(&e)
	if !ok {
		p.setErr("encryptedContentInfo missing contentEncryptionAlgorithm")
		return nil, false
	}

	encBytes, ok := readCtxImplicitOctetStringFlatten(&e, 0)
	if !ok {
		p.setErr("encryptedContentInfo missing encryptedContent [0]")
		return nil, false
	}

	pt, ok := p.decryptAlgid(Asn1View{b: algidTlv}, password, encBytes)
	if !ok {
		return nil, false
	}
	_ = orig
	return pt, true
}

// ParseP12File 读取并解析 .p12/.pfx 文件，返回提取到的私钥（PKCS#8 DER）、叶子证书与链证书（DER）。
func (p *p12miniParser) ParseP12File(path, password string) (*p12Parsed, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		p.setErr("failed to read p12 file")
		return nil, errors.New(p.lastErr)
	}

	out := &p12Parsed{}
	in := Asn1View{b: data}

	pfxSeq, ok := asn1ReadSequence(&in)
	if !ok {
		p.setErr("PFX is not a SEQUENCE")
		return nil, errors.New(p.lastErr)
	}

	pfx := pfxSeq
	version, ok := asn1ReadInt(&pfx)
	if !ok || version != 3 {
		p.setErr("unsupported PFX version (expected 3)")
		return nil, errors.New(p.lastErr)
	}

	authsafeCiSeq, ok := asn1ReadSequence(&pfx)
	if !ok {
		p.setErr("missing authSafe ContentInfo")
		return nil, errors.New(p.lastErr)
	}

	authOid, authAny, ok := parseContentInfoSeq(authsafeCiSeq)
	if !ok {
		p.setErr("bad authSafe ContentInfo")
		return nil, errors.New(p.lastErr)
	}
	if !oidEq(authOid, OID_pkcs7_data) {
		p.setErr("authSafe ContentInfo not pkcs7-data")
		return nil, errors.New(p.lastErr)
	}

	authBytes, ok := readOctetStringFlatten(&authAny)
	if !ok {
		p.setErr("authSafe content is not OCTET STRING")
		return nil, errors.New(p.lastErr)
	}

	asIn := Asn1View{b: authBytes}
	asSeq, ok := asn1ReadSequence(&asIn)
	if !ok {
		p.setErr(fmt.Sprintf("AuthenticatedSafe not SEQUENCE (first16=%s)", bytesToHexPrefix(authBytes, 16)))
		return nil, errors.New(p.lastErr)
	}

	items := asSeq
	for len(items.b) > 0 {
		ciSeq, ok := asn1ReadSequence(&items)
		if !ok {
			p.setErr("inner ContentInfo not SEQUENCE")
			return nil, errors.New(p.lastErr)
		}
		oid, any0, ok := parseContentInfoSeq(ciSeq)
		if !ok {
			p.setErr("bad inner ContentInfo")
			return nil, errors.New(p.lastErr)
		}

		switch {
		case oidEq(oid, OID_pkcs7_data):
			scBytes, ok := readOctetStringFlatten(&any0)
			if !ok {
				p.setErr("safe data not OCTET")
				return nil, errors.New(p.lastErr)
			}
			sc := Asn1View{b: scBytes}
			if !p.parseSafecontents(sc, password, out) {
				p.setErrOnce("parse SafeContents failed")
				return nil, errors.New(p.lastErr)
			}

		case oidEq(oid, OID_pkcs7_encrypted):
			plain, ok := p.parseEncrypteddataToPlain(password, any0)
			if !ok {
				p.setErrOnce("decrypt EncryptedData failed")
				return nil, errors.New(p.lastErr)
			}
			sc := Asn1View{b: plain}
			if !p.parseSafecontents(sc, password, out) {
				p.setErrOnce("parse decrypted SafeContents failed")
				return nil, errors.New(p.lastErr)
			}
		}
	}

	if len(out.keyDER) == 0 && len(out.certDER) == 0 && len(out.caDER) == 0 {
		p.setErr("no key/cert found")
		return nil, errors.New(p.lastErr)
	}
	return out, nil
}

// GetPrivateKey parses a .p12/.pfx file and returns PEM blocks (PRIVATE KEY / CERTIFICATE).
// GetPrivateKey 解析 .p12/.pfx 并返回 PEM block 列表：
// - PRIVATE KEY：PKCS#8 私钥 DER
// - CERTIFICATE：叶子证书 DER
// - CERTIFICATE：链/CA 证书 DER（若存在）
func GetPrivateKey(privateKeyName, privatePassword string) ([]*pem.Block, error) {
	p := &p12miniParser{}
	parsed, err := p.ParseP12File(privateKeyName, privatePassword)
	if err != nil {
		return nil, err
	}

	var blocks []*pem.Block
	if len(parsed.keyDER) > 0 {
		blocks = append(blocks, &pem.Block{Type: "PRIVATE KEY", Bytes: parsed.keyDER})
	}
	if len(parsed.certDER) > 0 {
		blocks = append(blocks, &pem.Block{Type: "CERTIFICATE", Bytes: parsed.certDER})
	}
	for _, c := range parsed.caDER {
		if len(c) == 0 {
			continue
		}
		blocks = append(blocks, &pem.Block{Type: "CERTIFICATE", Bytes: c})
	}
	if len(blocks) == 0 {
		return nil, fmt.Errorf("no pem blocks produced")
	}
	return blocks, nil
}
