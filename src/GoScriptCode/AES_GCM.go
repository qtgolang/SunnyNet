//go:build !mini
// +build !mini

package GoScriptCode

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

type AesGCMEngine struct {
	key   []byte
	nonce []byte
	aead  cipher.AEAD
	block cipher.Block
}

func NewAesGCMEngine(key []byte) (*AesGCMEngine, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("AES-GCM KEY 必须是 16、24 或 32 字节")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}
	return &AesGCMEngine{
		key:   append([]byte(nil), key...),
		block: block,
	}, nil
}
func (g *AesGCMEngine) SetNonce(nonce []byte) bool {
	if len(nonce) != 12 && len(nonce) != 16 && len(nonce) != 24 && len(nonce) != 32 {
		return false
	}
	if g.block == nil {
		return false
	}
	g.nonce = append(g.nonce[:0], nonce...)
	ahead, err := cipher.NewGCMWithNonceSize(g.block, len(nonce))
	if err != nil {
		return false
	}

	g.aead = ahead
	return true
}
func (g *AesGCMEngine) Encrypt(plaintext, aad []byte) (ciphertext, tag []byte) {
	sealed := g.aead.Seal(nil, g.nonce, plaintext, aad)
	overhead := g.aead.Overhead()
	return append([]byte(nil), sealed[:len(sealed)-overhead]...), append([]byte(nil), sealed[len(sealed)-overhead:]...)
}

func (g *AesGCMEngine) ComputeTag(ciphertext, aad []byte) []byte {
	if g.aead == nil {
		return nil
	}
	block, _ := aes.NewCipher(g.key)
	var hashSubkey [16]byte
	block.Encrypt(hashSubkey[:], make([]byte, aes.BlockSize))

	j0 := makeJ0(hashSubkey, g.nonce)
	s := ghash(hashSubkey, aad, ciphertext)

	var tag [16]byte
	block.Encrypt(tag[:], j0[:])
	for i := range tag {
		tag[i] ^= s[i]
	}
	return append([]byte(nil), tag[:g.aead.Overhead()]...)
}

func (g *AesGCMEngine) Decrypt(ciphertext, aad []byte) []byte {
	if g.aead == nil {
		return nil
	}
	tag := g.ComputeTag(ciphertext, aad)
	sealed := make([]byte, 0, len(ciphertext)+len(tag))
	sealed = append(sealed, ciphertext...)
	sealed = append(sealed, tag...)
	plaintext, err := g.aead.Open(nil, g.nonce, sealed, aad)
	if err != nil {
		return nil
	}
	return plaintext
}

func multiply(x, y [16]byte) [16]byte {
	var z [16]byte
	v := y
	for i := 0; i < 128; i++ {
		if x[i/8]&(1<<uint(7-i%8)) != 0 {
			for j := range z {
				z[j] ^= v[j]
			}
		}
		lsb := v[15] & 1
		for j := 15; j > 0; j-- {
			v[j] = v[j]>>1 | v[j-1]<<7
		}
		v[0] >>= 1
		if lsb != 0 {
			v[0] ^= 0xe1
		}
	}
	return z
}

func ghashUpdate(y, h [16]byte, data []byte) [16]byte {
	for len(data) > 0 {
		var block [16]byte
		n := len(data)
		if n > len(block) {
			n = len(block)
		}
		copy(block[:], data[:n])
		for i := range y {
			y[i] ^= block[i]
		}
		y = multiply(y, h)
		data = data[n:]
	}
	return y
}

func ghash(h [16]byte, aad, ciphertext []byte) [16]byte {
	var y [16]byte
	y = ghashUpdate(y, h, aad)
	y = ghashUpdate(y, h, ciphertext)

	var lengths [16]byte
	putUint64(lengths[0:8], uint64(len(aad))*8)
	putUint64(lengths[8:16], uint64(len(ciphertext))*8)
	for i := range y {
		y[i] ^= lengths[i]
	}
	return multiply(y, h)
}

func makeJ0(h [16]byte, nonce []byte) [16]byte {
	if len(nonce) == 12 {
		var j0 [16]byte
		copy(j0[:12], nonce)
		j0[15] = 1
		return j0
	}

	var j0 [16]byte
	j0 = ghashUpdate(j0, h, nonce)
	var lengths [16]byte
	putUint64(lengths[8:16], uint64(len(nonce))*8)
	for i := range j0 {
		j0[i] ^= lengths[i]
	}
	return multiply(j0, h)
}

func putUint64(dst []byte, value uint64) {
	for i := len(dst) - 1; i >= 0; i-- {
		dst[i] = byte(value)
		value >>= 8
	}
}
func AES_GCM_Encrypt(key, nonce []byte, data []byte, aad []byte) (string, error) {
	a, e := NewAesGCMEngine(key)
	if e != nil {
		return "", e
	}
	a.SetNonce(nonce)
	t1, t2 := a.Encrypt(data, aad)
	return base64.StdEncoding.EncodeToString(append(t1, t2...)), nil
}

func AES_GCM_Decrypt(key, nonce []byte, data []byte, aad []byte, isTag bool) ([]byte, error) {
	a, e := NewAesGCMEngine(key)
	if e != nil {
		return nil, e
	}
	a.SetNonce(nonce)
	ciphertext := data
	if isTag {
		if len(data) < 16 {
			return nil, fmt.Errorf("AES-GCM 携带 Tag 的密文至少需要包含 16 字节认证标签")
		}
		ciphertext = data[:len(data)-16]
	}
	if len(ciphertext) == 0 && len(data) > 0 && !isTag {
		return nil, fmt.Errorf("AES-GCM 解密输入不能为空")
	}
	t1 := a.Decrypt(ciphertext, aad)
	return t1, nil
}
