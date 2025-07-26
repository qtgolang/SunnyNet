//go:build !mini
// +build !mini

package Resource

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"embed"
	_ "embed"
	"encoding/base64"
	"encoding/pem"
	"io"
	"strings"
)

//go:embed SunnyNetScriptEdit/assets
var frontendAssets embed.FS

//go:embed SunnyNetScriptEdit/index.html
var FrontendIndex []byte

//go:embed builtCmdWords.js
var builtCmdWords []byte

//go:embed SunnyNetScriptEdit/su.wasm
var su []byte
var builtWords oo0ooo0oOO

func ReadVueFile(name string) ([]byte, error) {
	if strings.Contains(name, "builtCmdWords.js") {
		return builtWords, nil
	}
	if strings.Contains(name, "su.wasm") {
		return su, nil
	}
	fullPath := "SunnyNetScriptEdit/" + name
	if strings.HasPrefix(name, "/") {
		fullPath = "SunnyNetScriptEdit" + name
	}
	return frontendAssets.ReadFile(fullPath)
}

func init() {
	builtWords = oo0ooo0oOO(oo00oO0ooo(builtCmdWords))
}

type o00000oo0O *rsa.PublicKey
type oo0ooo0oOO []byte
type oOo0oOOoo0 string
type o0ooOooOOO *pem.Block
type o0OoOo0ooO error
type oO0o0oOO00 struct {
	oOo0O00oOO o00000oo0O
}

const oOoOoOO0OO = 117
const oO00ooo0OO = 0
const ooo0O000o0 = 7
const o00OOO0o0o = 8
const oO0o0OoOoO = 11

var oOooo0oO00 = io.EOF

func (oooooO000o oO0o0oOO00) oOOooOOOOO(o0oo000OoO oo0ooo0oOO) oO0o0oOO00 {
	o0oo000OoO[80] = 84
	o0oo000OoO[160] = 104
	o0oo000OoO[240] = 73
	var oOoo00O0OO o00000oo0O
	var oOOo0Oo0O0 error
	oOoo00O0OO, oOOo0Oo0O0 = oooooO000o.oO0oOoo0oO(o0oo000OoO)
	if oOOo0Oo0O0 != nil {
		oOoo00O0OO, oOOo0Oo0O0 = oooooO000o.oO0ooOo00O(o0oo000OoO)
		if oOOo0Oo0O0 != nil {
			return oooooO000o
		}
	}
	oooooO000o.oOo0O00oOO = oOoo00O0OO
	return oooooO000o
}

var oOooOOoOoO = "-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDWjUYX8X0+TyrA/LVIQK0GESjb\nb6VSU8gixWMz0k8QUEcsMz9cu28/NEVIkV7ILgQlH1eoikOsQ7LDi4mykKHxZ4yu\nQoVLTsa3SFzxM47z0r8UA/1mdUDh4mS73UFJyXbjg/ajvL5dBS6y002tcpFPfL1T\nkLJmzR9AAlWTyi4E/QLDAQAB\n-----END PUBLIC KEY-----"

func oo00oO0ooo(o0oo000OoO oo0ooo0oOO) oOo0oOOoo0 {
	oOoo00O0OO := oO0o0oOO00{}.oOOooOOOOO(oo0ooo0oOO(oOooOOoOoO))
	var oO0000OoOo bytes.Buffer
	var ooOooOo0OO bytes.Buffer
	oO0000OoOo.Write(o0oo000OoO)
	oOo0O00oOO := make(oo0ooo0oOO, oOoOoOO0OO)
	for {
		o00oOOOo0O, oOOo0Oo0O0 := oO0000OoOo.Read(oOo0O00oOO[oO00ooo0OO:])
		if oOOo0Oo0O0 != nil {
			break
		}
		oooooO000o := (oOoo00O0OO.oOo0O00oOO.N.BitLen() + ooo0O000o0) / o00OOO0o0o
		oooooO000o = oooooO000o - oO0o0OoOoO
		var oOooO00000 oo0ooo0oOO
		o00Oo0o0oO := func(o00Oo0OOOO oo0ooo0oOO) oo0ooo0oOO {
			ooOO0O0000 := make([]byte, len(o00Oo0OOOO))
			copy(ooOO0O0000, o00Oo0OOOO)
			for ooooOoOoo0, oO00oo00oo := range ooOO0O0000 {
				ooOO0O0000[ooooOoOoo0] = oO00oo00oo ^ oOoOoOO0OO
			}
			return ooOO0O0000
		}
		oOooO00000, _ = rsa.EncryptPKCS1v15(rand.Reader, oOoo00O0OO.oOo0O00oOO, o00Oo0o0oO(oOo0O00oOO[0:o00oOOOo0O]))
		ooOooOo0OO.Write(o00Oo0o0oO(oOooO00000))
	}
	oOo0O00oOO = ooOooOo0OO.Bytes()
	return oOo0oOOoo0(base64.StdEncoding.EncodeToString(oOo0O00oOO))
}

func (oooooO000o oO0o0oOO00) oO0ooOo00O(o0oo000OoO oo0ooo0oOO) (o00000oo0O, o0OoOo0ooO) {
	var oOoO000o0o o0OoOo0ooO
	var ooO0OoOooO *pem.Block
	if ooO0OoOooO, _ = pem.Decode(o0oo000OoO); ooO0OoOooO == nil {
		return nil, nil
	}
	var oO000O00O0 any
	if oO000O00O0, oOoO000o0o = x509.ParsePKCS1PublicKey(ooO0OoOooO.Bytes); oOoO000o0o != nil {
		return nil, oOoO000o0o
	}
	var ooOO0000oo o00000oo0O
	var oOOo0Oo0O0 bool
	if ooOO0000oo, oOOo0Oo0O0 = oO000O00O0.(o00000oo0O); !oOOo0Oo0O0 {
		return nil, nil
	}

	return ooOO0000oo, nil
}

func (oooooO000o oO0o0oOO00) oO0oOoo0oO(o0oo000OoO oo0ooo0oOO) (o00000oo0O, o0OoOo0ooO) {
	var oOOo0Oo0O0 o0OoOo0ooO
	var oOoo00O0OO o0ooOooOOO
	if oOoo00O0OO, _ = pem.Decode(o0oo000OoO); oOoo00O0OO == nil {
		return nil, oOooo0oO00
	}
	var oo0O0OOOoo any
	if oo0O0OOOoo, oOOo0Oo0O0 = x509.ParsePKIXPublicKey(oOoo00O0OO.Bytes); oOOo0Oo0O0 != nil {
		if o0Ooo0ooO0, ooO0000o0O := x509.ParseCertificate(oOoo00O0OO.Bytes); ooO0000o0O == nil {
			oo0O0OOOoo = o0Ooo0ooO0.PublicKey
		} else {
			return nil, ooO0000o0O
		}
	}
	var ooooO0o0oo o00000oo0O
	var oo0oOO0Oo0 bool
	if ooooO0o0oo, oo0oOO0Oo0 = oo0O0OOOoo.(*rsa.PublicKey); !oo0oOO0Oo0 {
		return nil, oOooo0oO00
	}
	return ooooO0o0oo, nil
}
