package openssl

import "fmt"

// Minimal DER/ASN.1 reader utilities.
// This is a Go port of the small C ASN.1 reader used by the p12mini parser.

type Asn1TagClass int

const (
	ASN1_CLASS_UNIVERSAL   Asn1TagClass = 0
	ASN1_CLASS_APPLICATION Asn1TagClass = 1
	ASN1_CLASS_CONTEXT     Asn1TagClass = 2
	ASN1_CLASS_PRIVATE     Asn1TagClass = 3
)

type Asn1Tag struct {
	cls         Asn1TagClass
	constructed bool
	tagnum      uint32
}

// Asn1View represents the remaining unread bytes.
// Sub-slices reference the same underlying array.
type Asn1View struct {
	b []byte
}

// remaining 返回当前还未读取的字节数。
func (v *Asn1View) remaining() int { return len(v.b) }

// readU8 读取并消耗 1 字节；失败返回 false。
func (v *Asn1View) readU8() (byte, bool) {
	if len(v.b) < 1 {
		return 0, false
	}
	out := v.b[0]
	v.b = v.b[1:]
	return out, true
}

// readTag 读取 ASN.1 标记头（类/是否构造/标签号），支持 high-tag-number 形式。
func readTag(in *Asn1View) (Asn1Tag, bool) {
	b, ok := in.readU8()
	if !ok {
		return Asn1Tag{}, false
	}
	tag := Asn1Tag{
		cls:         Asn1TagClass((b >> 6) & 0x3),
		constructed: (b & 0x20) != 0,
		tagnum:      uint32(b & 0x1f),
	}

	if tag.tagnum != 0x1f {
		return tag, true
	}

	// High-tag-number form.
	var tn uint32
	for {
		b2, ok := in.readU8()
		if !ok {
			return Asn1Tag{}, false
		}
		tn = (tn << 7) | uint32(b2&0x7f)
		if (b2 & 0x80) == 0 {
			break
		}
		if tn > 0x0fffffff {
			return Asn1Tag{}, false
		}
	}
	tag.tagnum = tn
	return tag, true
}

// readLen returns (length, indefinite).
// readLen 读取 ASN.1 长度字段；返回 (length, isIndefinite, ok)。
func readLen(in *Asn1View) (int, bool, bool) {
	b, ok := in.readU8()
	if !ok {
		return 0, false, false
	}
	if (b & 0x80) == 0 {
		return int(b), false, true
	}
	n := int(b & 0x7f)
	if n == 0 {
		return 0, true, true // BER indefinite length
	}
	// In C this is bounded by sizeof(size_t). Here we cap to avoid overflow.
	if n > 8 {
		return 0, false, false
	}
	if len(in.b) < n {
		return 0, false, false
	}
	v := 0
	for i := 0; i < n; i++ {
		v = (v << 8) | int(in.b[i])
	}
	in.b = in.b[n:]
	return v, false, true
}

// isEoc 判断是否为 BER 不定长内容的 End-Of-Contents（00 00）标记。
func isEoc(tag Asn1Tag, content Asn1View) bool {
	return tag.cls == ASN1_CLASS_UNIVERSAL && !tag.constructed && tag.tagnum == 0 && len(content.b) == 0
}

// asn1SkipTlv 跳过一个完整 TLV（含 BER 不定长构造类型的扫描跳过）。
func asn1SkipTlv(in *Asn1View) bool {
	tmp := *in
	tag, ok := readTag(&tmp)
	if !ok {
		return false
	}
	l, indef, ok := readLen(&tmp)
	if !ok {
		return false
	}
	if !indef {
		if len(tmp.b) < l {
			return false
		}
		tmp.b = tmp.b[l:]
		*in = tmp
		return true
	}
	if !tag.constructed {
		return false
	}
	// tmp now points at the start of indefinite contents.
	*in = tmp
	return asn1SkipIndefiniteContents(in)
}

// asn1SkipIndefiniteContents 扫描并跳过 BER 不定长构造内容，直到遇到 EOC。
func asn1SkipIndefiniteContents(in *Asn1View) bool {
	for {
		tmp := *in
		tag, content, ok := asn1ReadTlv(&tmp)
		if !ok {
			return false
		}
		if isEoc(tag, content) {
			*in = tmp // tmp is positioned after the EOC TLV.
			return true
		}
		if !asn1SkipTlv(in) {
			return false
		}
	}
}

// asn1ReadTlv 读取一个 TLV，返回 (tag, contentView, ok)，并推进输入游标。
// 对 BER 不定长构造类型，会扫描 EOC 并返回去掉 EOC 的 content 视图。
func asn1ReadTlv(in *Asn1View) (Asn1Tag, Asn1View, bool) {
	tmp := *in
	tag, ok := readTag(&tmp)
	if !ok {
		return Asn1Tag{}, Asn1View{}, false
	}
	l, indef, ok := readLen(&tmp)
	if !ok {
		return Asn1Tag{}, Asn1View{}, false
	}
	if !indef {
		if len(tmp.b) < l {
			return Asn1Tag{}, Asn1View{}, false
		}
		content := Asn1View{b: tmp.b[:l]}
		tmp.b = tmp.b[l:]
		*in = tmp
		return tag, content, true
	}

	if !tag.constructed {
		return Asn1Tag{}, Asn1View{}, false
	}

	// Handle BER indefinite-length constructed types by scanning for EOC.
	scan := tmp
	scan2 := scan
	if !asn1SkipIndefiniteContents(&scan2) {
		return Asn1Tag{}, Asn1View{}, false
	}
	consumed := len(scan.b) - len(scan2.b) // content bytes + 2 bytes EOC
	if consumed < 2 {
		return Asn1Tag{}, Asn1View{}, false
	}
	content := Asn1View{b: scan.b[:consumed-2]} // strip EOC
	*in = scan2
	return tag, content, true
}

// asn1ExpectUniversal 读取并校验下一个 TLV 必须是 UNIVERSAL 指定 tagnum（例如 SEQUENCE=16）。
func asn1ExpectUniversal(in *Asn1View, tagnum uint32) (Asn1View, bool) {
	tag, content, ok := asn1ReadTlv(in)
	if !ok {
		return Asn1View{}, false
	}
	if tag.cls != ASN1_CLASS_UNIVERSAL || tag.tagnum != tagnum {
		return Asn1View{}, false
	}
	return content, true
}

// asn1ReadInt 读取非负 INTEGER（最多 4 字节）并转成 int。
func asn1ReadInt(in *Asn1View) (int, bool) {
	c, ok := asn1ExpectUniversal(in, 2)
	if !ok {
		return 0, false
	}
	if len(c.b) == 0 {
		return 0, false
	}
	if (c.b[0] & 0x80) != 0 {
		return 0, false
	}
	if len(c.b) > 4 {
		return 0, false
	}
	v := 0
	for _, bb := range c.b {
		v = (v << 8) | int(bb)
	}
	return v, true
}

// asn1ReadOctetString 读取 OCTET STRING 的内容视图（不拷贝）。
func asn1ReadOctetString(in *Asn1View) (Asn1View, bool) {
	return asn1ExpectUniversal(in, 4)
}

// asn1ReadOID 读取 OBJECT IDENTIFIER 的内容视图（不做 OID 解码，返回原始字节）。
func asn1ReadOID(in *Asn1View) (Asn1View, bool) {
	return asn1ExpectUniversal(in, 6)
}

// asn1ReadSequence 读取 SEQUENCE 的内容视图（不拷贝）。
func asn1ReadSequence(in *Asn1View) (Asn1View, bool) {
	return asn1ExpectUniversal(in, 16)
}

// asn1ReadCtxExplicit 读取显式的 context-specific [tagnum]（必须是 constructed），返回其内容视图。
func asn1ReadCtxExplicit(in *Asn1View, tagnum uint32) (Asn1View, bool) {
	tag, c, ok := asn1ReadTlv(in)
	if !ok {
		return Asn1View{}, false
	}
	if tag.cls != ASN1_CLASS_CONTEXT || !tag.constructed || tag.tagnum != tagnum {
		return Asn1View{}, false
	}
	return c, true
}

// asn1ReadCtxImplicitOctetString 读取隐式的 context-specific [tagnum] OCTET STRING（必须是 primitive），返回其内容视图。
func asn1ReadCtxImplicitOctetString(in *Asn1View, tagnum uint32) (Asn1View, bool) {
	tag, c, ok := asn1ReadTlv(in)
	if !ok {
		return Asn1View{}, false
	}
	if tag.cls != ASN1_CLASS_CONTEXT || tag.constructed || tag.tagnum != tagnum {
		return Asn1View{}, false
	}
	return c, true
}

// asn1DupNextTlv copies the next TLV (including header) into a new byte slice.
// asn1DupNextTlv 将下一个 TLV（含 tag/len 头）整体拷贝出来，并推进输入游标。
func asn1DupNextTlv(in *Asn1View) ([]byte, bool) {
	orig := *in
	origLen := len(orig.b)
	// Consume once on a copy to compute TLV length.
	tmp := *in
	_, _, ok := asn1ReadTlv(&tmp)
	if !ok {
		return nil, false
	}
	consumed := origLen - len(tmp.b)
	if consumed <= 0 || consumed > origLen {
		return nil, false
	}
	out := make([]byte, consumed)
	copy(out, in.b[:consumed])
	*in = tmp
	return out, true
}

// String 用于调试显示当前视图长度。
func (v Asn1View) String() string {
	return fmt.Sprintf("Asn1View{len=%d}", len(v.b))
}
