package openssl

import (
	"crypto/sha1"
)

// pkcs12KdfSHA1 implements the PKCS#12 key derivation function with SHA-1
// (RFC 7292 Appendix B.2 style), matching the logic in the original C code.
// pkcs12KdfSHA1 用于 PKCS#12 传统 PBE：根据密码+salt 派生 key/iv（id=1 代表 key，id=2 代表 iv）。
func pkcs12KdfSHA1(password string, salt []byte, id int, iter int, outLen int) []byte {
	if iter <= 0 {
		iter = 1
	}
	if outLen <= 0 {
		return nil
	}
	if salt == nil {
		salt = []byte{}
	}

	pwBMP := passwordToBMPASCIIUTF16BE(password)
	// D = id repeated v times
	const u = 20 // SHA1 output size
	const v = 64

	D := make([]byte, v)
	for i := 0; i < v; i++ {
		D[i] = byte(id)
	}

	// S = salt repeated
	Slen := v * ((len(salt) + v - 1) / v)
	// In the original C implementation passwordToBMP always includes a 0x0000 terminator,
	// so pwBMP length is never zero.
	Plen := v * ((len(pwBMP) + v - 1) / v)
	Ilen := Slen + Plen

	I := make([]byte, Ilen)
	// Fill I = S || P
	if Slen > 0 {
		for i := 0; i < Slen; i++ {
			I[i] = salt[i%len(salt)]
		}
	}
	if Plen > 0 {
		for i := 0; i < Plen; i++ {
			I[Slen+i] = pwBMP[i%len(pwBMP)]
		}
	}

	out := make([]byte, outLen)
	n := outLen
	outp := 0

	// Ai = SHA1(D || I)
	Ai := make([]byte, u)
	for {
		h := sha1.New()
		h.Write(D)
		h.Write(I)
		sum := h.Sum(nil) // u bytes
		copy(Ai, sum)

		for r := 1; r < iter; r++ {
			h2 := sha1.New()
			h2.Write(Ai)
			copy(Ai, h2.Sum(nil))
		}

		toCopy := u
		if n < u {
			toCopy = n
		}
		copy(out[outp:outp+toCopy], Ai[:toCopy])
		outp += toCopy
		n -= toCopy
		if n <= 0 {
			break
		}

		// B = first u bytes of Ai repeated to v
		B := make([]byte, v)
		for j := 0; j < v; j++ {
			B[j] = Ai[j%u]
		}

		// Modify I: Ij = (Ij + B + 1) mod 2^16 (per v-byte block)
		for j := 0; j < Ilen; j += v {
			carry := uint16(1)
			for k := v; k > 0; {
				k--
				carry += uint16(I[j+k]) + uint16(B[k])
				I[j+k] = byte(carry)
				carry >>= 8
			}
		}
	}
	return out
}

// passwordToBMPASCIIUTF16BE 将密码按“UTF-16BE BMPString（仅按字节映射）+ 0x0000 结尾”的方式编码。
// 该行为与之前 C 版本保持一致（把 UTF-8 的字节直接当作字符码点低字节）。
func passwordToBMPASCIIUTF16BE(password string) []byte {
	// Matches C code: each byte of the input string is mapped to a BMP code unit
	// with big-endian UTF-16 encoding. The terminator is 0x0000.
	pw := []byte(password) // uses UTF-8 bytes, just like the C wrapper used UTF-8 bytes.
	// (n+1)*2: includes trailing 0x0000
	buf := make([]byte, (len(pw)+1)*2)
	for i := 0; i < len(pw); i++ {
		buf[2*i] = 0x00
		buf[2*i+1] = pw[i]
	}
	// buf already zeroed for terminator
	return buf
}

// bytesRepeatToLen 将 in 循环重复填充到指定长度；主要用于构造 KDF 的 S/P/I 缓冲区。
func bytesRepeatToLen(in []byte, outLen int) []byte {
	if outLen <= 0 {
		return []byte{}
	}
	if len(in) == 0 {
		return make([]byte, outLen)
	}
	out := make([]byte, outLen)
	for i := 0; i < outLen; i++ {
		out[i] = in[i%len(in)]
	}
	return out
}
