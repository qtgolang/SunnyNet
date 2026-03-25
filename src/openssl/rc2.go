package openssl

// RC2 implementation (CBC + PKCS#7), ported from the p12mini C code.

type Rc2Key struct {
	K [64]uint16
}

// KEY_TABLE matches OpenSSL crypto/rc2/rc2_skey.c
var rc2KeyTable = [256]uint8{
	217, 120, 249, 196, 25, 221, 181, 237, 40, 233, 253, 121, 74, 160, 216, 157,
	198, 126, 55, 131, 43, 118, 83, 142, 98, 76, 100, 136, 68, 139, 251, 162,
	23, 154, 89, 245, 135, 179, 79, 19, 97, 69, 109, 141, 9, 129, 125, 50,
	189, 143, 64, 235, 134, 183, 123, 11, 240, 149, 33, 34, 92, 107, 78, 130,
	84, 214, 101, 147, 206, 96, 178, 28, 115, 86, 192, 20, 167, 140, 241, 220,
	18, 117, 202, 31, 59, 190, 228, 209, 66, 61, 212, 48, 163, 60, 182, 38,
	111, 191, 14, 218, 70, 105, 7, 87, 39, 242, 29, 155, 188, 148, 67, 3,
	248, 17, 199, 246, 144, 239, 62, 231, 6, 195, 213, 47, 200, 102, 30, 215,
	8, 232, 234, 222, 128, 82, 238, 247, 132, 170, 114, 172, 53, 77, 106, 42,
	150, 26, 210, 113, 90, 21, 73, 116, 75, 159, 208, 94, 4, 24, 164, 236,
	194, 224, 65, 110, 15, 81, 203, 204, 36, 145, 175, 80, 161, 244, 112, 57,
	153, 124, 58, 133, 35, 184, 180, 122, 252, 2, 54, 91, 37, 85, 151, 49,
	45, 93, 250, 152, 227, 138, 146, 174, 5, 223, 41, 16, 103, 108, 186, 201,
	211, 0, 230, 207, 225, 158, 168, 44, 99, 22, 1, 63, 88, 226, 137, 169,
	13, 56, 52, 27, 171, 51, 255, 176, 187, 72, 12, 95, 185, 177, 205, 46,
	197, 243, 219, 71, 229, 165, 156, 119, 10, 166, 32, 104, 254, 127, 193, 173,
}

// loadLE32 以小端读取 4 字节为 uint32。
func loadLE32(p []byte) uint32 {
	return uint32(p[0]) |
		uint32(p[1])<<8 |
		uint32(p[2])<<16 |
		uint32(p[3])<<24
}

// storeLE32 以小端写入 uint32 到 4 字节。
func storeLE32(p []byte, v uint32) {
	p[0] = byte(v)
	p[1] = byte(v >> 8)
	p[2] = byte(v >> 16)
	p[3] = byte(v >> 24)
}

// rc2SetKey_ossl 按 OpenSSL 兼容方式生成 RC2 轮密钥（effective key bits 用于 RC2-40/RC2-128 等模式）。
func rc2SetKey_ossl(ks *Rc2Key, data []byte, bits int) {
	// Mirror OpenSSL's internal RC2_set_key_ossl logic.
	var k [128]uint8
	if len(data) > 128 {
		data = data[:128]
	}
	if bits <= 0 {
		bits = 1024
	}
	if bits > 1024 {
		bits = 1024
	}
	for i := 0; i < len(data); i++ {
		k[i] = data[i]
	}
	d := uint32(k[len(data)-1])
	j := 0
	for i := len(data); i < 128; i++ {
		d = uint32(rc2KeyTable[(k[j]+uint8(d))&0xff])
		k[i] = uint8(d)
		j++
	}

	j = (bits + 7) >> 3
	i := 128 - j
	c := uint8(0xff >> (uint((-bits) & 0x07)))
	d2 := rc2KeyTable[k[i]&c]
	k[i] = d2
	for i--; i >= 0; i-- {
		d2 = rc2KeyTable[k[i+j]^d2]
		k[i] = d2
		if i == 0 {
			break
		}
	}

	ki := 63
	for i = 127; i >= 0; i -= 2 {
		ks.K[ki] = (uint16(k[i]) << 8) | uint16(k[i-1])
		ki--
		if i <= 1 {
			break
		}
	}
}

// rc2DecryptBlockOSSL 解密一个 RC2 64-bit 分组（与 OpenSSL RC2_decrypt 兼容）。
func rc2DecryptBlockOSSL(d0 *uint32, d1 *uint32, ks *Rc2Key) {
	// Mirrors RC2_decrypt_ossl
	l0 := *d0
	x0 := uint16(l0 & 0xffff)
	x1 := uint16((l0 >> 16) & 0xffff)

	l1 := *d1
	x2 := uint16(l1 & 0xffff)
	x3 := uint16((l1 >> 16) & 0xffff)

	n := uint32(3)
	i := uint32(5)
	p0 := 63
	for {
		t := (uint32(x3)<<11 | uint32(x3)>>5) & 0xffff
		x3 = uint16((t - (uint32(x0) &^ uint32(x2)) - (uint32(x1) & uint32(x2)) - uint32(ks.K[p0])) & 0xffff)
		p0--

		t = (uint32(x2)<<13 | uint32(x2)>>3) & 0xffff
		x2 = uint16((t - (uint32(x3) &^ uint32(x1)) - (uint32(x0) & uint32(x1)) - uint32(ks.K[p0])) & 0xffff)
		p0--

		t = (uint32(x1)<<14 | uint32(x1)>>2) & 0xffff
		x1 = uint16((t - (uint32(x2) &^ uint32(x0)) - (uint32(x3) & uint32(x0)) - uint32(ks.K[p0])) & 0xffff)
		p0--

		t = (uint32(x0)<<15 | uint32(x0)>>1) & 0xffff
		x0 = uint16((t - (uint32(x1) &^ uint32(x3)) - (uint32(x2) & uint32(x3)) - uint32(ks.K[p0])) & 0xffff)
		p0--

		i--
		if i == 0 {
			if n == 0 {
				break
			}
			n--
			if n == 0 {
				break
			}
			if n == 2 {
				i = 6
			} else {
				i = 5
			}

			// p1 is ks.K[0..]
			x3 = uint16((uint32(x3) - uint32(ks.K[x2&0x3f])) & 0xffff)
			x2 = uint16((uint32(x2) - uint32(ks.K[x1&0x3f])) & 0xffff)
			x1 = uint16((uint32(x1) - uint32(ks.K[x0&0x3f])) & 0xffff)
			x0 = uint16((uint32(x0) - uint32(ks.K[x3&0x3f])) & 0xffff)
		}
	}

	*d0 = uint32(x0) | uint32(x1)<<16
	*d1 = uint32(x2) | uint32(x3)<<16
}

// rc2ECBDecryptBlock 对单个 8 字节分组做 RC2-ECB 解密（原地修改）。
func rc2ECBDecryptBlock(ks *Rc2Key, block []byte) {
	if len(block) < 8 {
		return
	}
	d0 := loadLE32(block[0:4])
	d1 := loadLE32(block[4:8])
	rc2DecryptBlockOSSL(&d0, &d1, ks)
	storeLE32(block[0:4], d0)
	storeLE32(block[4:8], d1)
}

// pkcs7UnpadBlockSize 对指定分组大小做 PKCS#7 去填充；成功返回去填充后的切片视图。
func pkcs7UnpadBlockSize(buf []byte, blockSize int) ([]byte, bool) {
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

// rc2CBCDecryptRaw 使用 RC2-CBC 解密（不做 PKCS#7 去填充）。
func rc2CBCDecryptRaw(ciphertext []byte, key []byte, effectiveBits int, iv []byte) ([]byte, bool) {
	if len(ciphertext) == 0 || len(ciphertext)%8 != 0 || len(iv) < 8 {
		return nil, false
	}
	var ks Rc2Key
	rc2SetKey_ossl(&ks, key, effectiveBits)

	pt := make([]byte, len(ciphertext))
	prev := make([]byte, 8)
	copy(prev, iv[:8])

	for off := 0; off < len(ciphertext); off += 8 {
		block := make([]byte, 8)
		copy(block, ciphertext[off:off+8])
		ctmp := make([]byte, 8)
		copy(ctmp, block)
		rc2ECBDecryptBlock(&ks, block)
		for i := 0; i < 8; i++ {
			block[i] ^= prev[i]
		}
		copy(pt[off:off+8], block)
		copy(prev, ctmp)
	}
	return pt, true
}

// rc2CBCDecryptPKCS7 使用 RC2-CBC 解密并做 PKCS#7 去填充。
func rc2CBCDecryptPKCS7(ciphertext []byte, key []byte, effectiveBits int, iv []byte) ([]byte, bool) {
	raw, ok := rc2CBCDecryptRaw(ciphertext, key, effectiveBits, iv)
	if !ok {
		return nil, false
	}
	return pkcs7UnpadBlockSize(raw, 8)
}
