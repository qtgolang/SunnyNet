package protobuf

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/qtgolang/SunnyNet/src/protobuf/JSON"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
	"io"
	"math"
	"os"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Number = protowire.Number

type Type = protowire.Type

const (
	VarintType     Type = protowire.VarintType
	VarintTypeRaw  Type = protowire.Type(10)
	PbObjectType   Type = protowire.Type(33)
	Fixed32Type    Type = protowire.Fixed32Type
	Fixed64Type    Type = protowire.Fixed64Type
	BytesType      Type = protowire.BytesType
	BytesTypeRaw   Type = protowire.Type(22)
	StartGroupType Type = protowire.StartGroupType
	EndGroupType   Type = protowire.EndGroupType
)

type (
	Token   token
	Message []Token
	Tag     struct {
		Number Number
		Type   Type
	}
	Bool         bool
	Varint       uint64 //int64
	Svarint      uint64 //int64
	Uvarint      uint64
	Int32        uint64 //int32
	Uint32       uint64 //uint32
	Float32      float32
	Int64        uint64 //int64
	Uint64       uint64
	Float64      float64
	String       string
	Bytes        []byte
	LengthPrefix Message
	Denormalized struct {
		Count uint
		Value Token
	}
	Raw []byte
)

type token interface {
	isToken()
}

func (Message) isToken()      {}
func (Tag) isToken()          {}
func (Bool) isToken()         {}
func (Varint) isToken()       {}
func (Svarint) isToken()      {}
func (Uvarint) isToken()      {}
func (Int32) isToken()        {}
func (Uint32) isToken()       {}
func (Float32) isToken()      {}
func (Int64) isToken()        {}
func (Uint64) isToken()       {}
func (Float64) isToken()      {}
func (String) isToken()       {}
func (Bytes) isToken()        {}
func (LengthPrefix) isToken() {}
func (Denormalized) isToken() {}
func (Raw) isToken()          {}

func (m Message) Marshal() []byte {
	var out []byte
	for _, v := range m {
		switch v := v.(type) {
		case Message:
			out = append(out, v.Marshal()...)
		case Tag:
			out = protowire.AppendTag(out, v.Number, v.Type)
		case Bool:
			out = protowire.AppendVarint(out, protowire.EncodeBool(bool(v)))
		case Varint:
			out = protowire.AppendVarint(out, uint64(v))
		case Svarint:
			out = protowire.AppendVarint(out, protowire.EncodeZigZag(int64(v)))
		case Uvarint:
			out = protowire.AppendVarint(out, uint64(v))
		case Int32:
			out = protowire.AppendFixed32(out, uint32(v))
		case Uint32:
			out = protowire.AppendFixed32(out, uint32(v))
		case Float32:
			out = protowire.AppendFixed32(out, math.Float32bits(float32(v)))
		case Int64:
			out = protowire.AppendFixed64(out, uint64(v))
		case Uint64:
			out = protowire.AppendFixed64(out, uint64(v))
		case Float64:
			out = protowire.AppendFixed64(out, math.Float64bits(float64(v)))
		case String:
			out = protowire.AppendBytes(out, []byte(v))
		case Bytes:
			out = protowire.AppendBytes(out, []byte(v))
		case LengthPrefix:
			out = protowire.AppendBytes(out, Message(v).Marshal())
		case Denormalized:
			b := Message{v.Value}.Marshal()
			_, n := protowire.ConsumeVarint(b)
			out = append(out, b[:n]...)
			for i := uint(0); i < v.Count; i++ {
				out[len(out)-1] |= 0x80 // set continuation bit on previous
				out = append(out, 0)
			}
			out = append(out, b[n:]...)
		case Raw:
			out = append(out, v...)
		default:
			panic(fmt.Sprintf("unknown type: %T", v))
		}
	}
	return out
}

func (m *Message) Unmarshal(in []byte) bool {
	return m.unmarshal(in, nil, false)
}

func (m *Message) unmarshal(in []byte, desc protoreflect.MessageDescriptor, inferMessage bool) bool {
	p := parser{in: in, out: *m}
	p.parseMessage(desc, false, inferMessage)
	*m = p.out
	return p.invalid
}

type parser struct {
	in      []byte
	out     []Token
	invalid bool
}

func (p *parser) parseMessage(msgDesc protoreflect.MessageDescriptor, group, inferMessage bool) {
	for len(p.in) > 0 {
		v, n := protowire.ConsumeVarint(p.in)
		num, typ := protowire.DecodeTag(v)
		if n < 0 || num < 0 || v > math.MaxUint32 {
			p.out, p.in = append(p.out, Raw(p.in)), nil
			p.invalid = true
			return
		}
		if typ == EndGroupType && group {
			return // if inside a group, then stop
		}
		p.in = p.in[n:]
		/*
			_, nan := protowire.ConsumeVarint(p.in)
			if nan > 8 && typ == VarintType {
				p.out = append(p.out, Tag{num, VarintTypeRaw})
			} else {
				p.out = append(p.out, Tag{num, typ})
			}
		*/
		p.out = append(p.out, Tag{num, typ})
		if m := n - protowire.SizeVarint(v); m > 0 {
			p.out[len(p.out)-1] = Denormalized{uint(m), p.out[len(p.out)-1]}
		}

		// If descriptor is available, use it for more accurate parsing.
		var isPacked bool
		var kind protoreflect.Kind
		var subDesc protoreflect.MessageDescriptor
		if msgDesc != nil && !msgDesc.IsPlaceholder() {
			if fieldDesc := msgDesc.Fields().ByNumber(num); fieldDesc != nil {
				isPacked = fieldDesc.IsPacked()
				kind = fieldDesc.Kind()
				switch kind {
				case protoreflect.MessageKind, protoreflect.GroupKind:
					subDesc = fieldDesc.Message()
					if subDesc == nil || subDesc.IsPlaceholder() {
						kind = 0
					}
				}
			}
		}

		switch typ {
		case VarintType:
			p.parseVarint(kind)
		case Fixed32Type:
			p.parseFixed32(kind)
		case Fixed64Type:
			p.parseFixed64(kind)
		case BytesType:
			p.parseBytes(isPacked, kind, subDesc, inferMessage)
		case StartGroupType:
			p.parseGroup(num, subDesc, inferMessage)
		case EndGroupType:
			// Handled by p.parseGroup.
		default:
			p.out, p.in = append(p.out, Raw(p.in)), nil
			p.invalid = true
		}
	}
}

func (p *parser) parseVarint(kind protoreflect.Kind) {
	v, n := protowire.ConsumeVarint(p.in)
	if n < 0 {
		p.out, p.in = append(p.out, Raw(p.in)), nil
		p.invalid = true
		return
	}
	switch kind {
	case protoreflect.BoolKind:
		switch v {
		case 0:
			p.out, p.in = append(p.out, Bool(false)), p.in[n:]
		case 1:
			p.out, p.in = append(p.out, Bool(true)), p.in[n:]
		default:
			p.out, p.in = append(p.out, Uvarint(v)), p.in[n:]
		}
	default:
		p.out, p.in = append(p.out, String(fmt.Sprintf("%d", Varint(v)))), p.in[n:]
	}
	if m := n - protowire.SizeVarint(v); m > 0 {
		p.out[len(p.out)-1] = Denormalized{uint(m), p.out[len(p.out)-1]}
	}
}
func ConsumeFixed32(b []byte) (v float32, n int) {
	if len(b) < 4 {
		return 0, -1
	}
	vv := uint32(b[0])<<0 | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	kk := math.Float32frombits(vv)
	return kk, 4
}
func (p *parser) parseFixed32(kind protoreflect.Kind) {
	v, n := ConsumeFixed32(p.in)
	if n < 0 {
		p.out, p.in = append(p.out, Raw(p.in)), nil
		p.invalid = true
		return
	}
	p.out, p.in = append(p.out, Float32(v)), p.in[n:]
	/*
		switch kind {
		case protoreflect.FloatKind:
			p.out, p.in = append(p.out, Float32(math.Float32frombits(v))), p.in[n:]
		case protoreflect.Sfixed32Kind:
			p.out, p.in = append(p.out, Int32(v)), p.in[n:]
		default:
			p.out, p.in = append(p.out, Uint32(v)), p.in[n:]
		}
	*/
}

// ConsumeFixed64 parses b as a little-endian uint64, reporting its length.
// This returns a negative length upon an error (see ParseError).
func ConsumeFixed64(b []byte) (v float64, n int) {
	if len(b) < 8 {
		return 0, -1
	}
	bits := binary.LittleEndian.Uint64(b)
	return math.Float64frombits(bits), 8
}
func (p *parser) parseFixed64(kind protoreflect.Kind) {
	v, n := ConsumeFixed64(p.in)
	if n < 0 {
		p.out, p.in = append(p.out, Raw(p.in)), nil
		p.invalid = true
		return
	}
	switch kind {
	case protoreflect.DoubleKind:
		p.out, p.in = append(p.out, Float32(v)), p.in[n:]
	case protoreflect.Sfixed64Kind:
		p.out, p.in = append(p.out, Float64(v)), p.in[n:]
	default:
		p.out, p.in = append(p.out, Float64(v)), p.in[n:]
	}
}

func (p *parser) parseBytes(isPacked bool, kind protoreflect.Kind, desc protoreflect.MessageDescriptor, inferMessage bool) {
	v, n := protowire.ConsumeVarint(p.in)
	if n < 0 {
		p.out, p.in = append(p.out, Raw(p.in)), nil
		p.invalid = true
		return
	}
	p.out, p.in = append(p.out, Uvarint(v)), p.in[n:]
	if m := n - protowire.SizeVarint(v); m > 0 {
		p.out[len(p.out)-1] = Denormalized{uint(m), p.out[len(p.out)-1]}
	}
	if v > uint64(len(p.in)) {
		p.out, p.in = append(p.out, Raw(p.in)), nil
		p.invalid = true
		return
	}
	p.out = p.out[:len(p.out)-1] // subsequent tokens contain prefix-length

	if isPacked {
		p.parsePacked(int(v), kind)
	} else {
		switch kind {
		case protoreflect.MessageKind:
			p2 := parser{in: p.in[:v]}
			p2.parseMessage(desc, false, inferMessage)
			p.out, p.in = append(p.out, LengthPrefix(p2.out)), p.in[v:]
		case protoreflect.StringKind:
			p.out, p.in = append(p.out, String(p.in[:v])), p.in[v:]
		case protoreflect.BytesKind:
			p.out, p.in = append(p.out, Bytes(p.in[:v])), p.in[v:]
		default:
			if inferMessage {
				// Check whether this is a syntactically valid message.
				p2 := parser{in: p.in[:v]}
				p2.parseMessage(nil, false, inferMessage)
				if !p2.invalid {
					p.out, p.in = append(p.out, LengthPrefix(p2.out)), p.in[v:]
					break
				}
			}
			p.out, p.in = append(p.out, Bytes(p.in[:v])), p.in[v:]
		}
	}
	if m := n - protowire.SizeVarint(v); m > 0 {
		p.out[len(p.out)-1] = Denormalized{uint(m), p.out[len(p.out)-1]}
	}
}

func (p *parser) parsePacked(n int, kind protoreflect.Kind) {
	p2 := parser{in: p.in[:n]}
	for len(p2.in) > 0 {
		switch kind {
		case protoreflect.BoolKind, protoreflect.EnumKind,
			protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind,
			protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind:
			p2.parseVarint(kind)
		case protoreflect.Fixed32Kind, protoreflect.Sfixed32Kind, protoreflect.FloatKind:
			p2.parseFixed32(kind)
		case protoreflect.Fixed64Kind, protoreflect.Sfixed64Kind, protoreflect.DoubleKind:
			p2.parseFixed64(kind)
		default:
			panic(fmt.Sprintf("invalid packed kind: %v", kind))
		}
	}
	p.out, p.in = append(p.out, LengthPrefix(p2.out)), p.in[n:]
}

func (p *parser) parseGroup(startNum protowire.Number, desc protoreflect.MessageDescriptor, inferMessage bool) {
	p2 := parser{in: p.in}
	p2.parseMessage(desc, true, inferMessage)
	if len(p2.out) > 0 {
		p.out = append(p.out, Message(p2.out))
	}
	p.invalid = p2.invalid
	p.in = p2.in

	// Append the trailing end group.
	v, n := protowire.ConsumeVarint(p.in)
	if endNum, typ := protowire.DecodeTag(v); typ == EndGroupType {
		if startNum != endNum {
			p.invalid = true
		}
		p.out, p.in = append(p.out, Tag{endNum, typ}), p.in[n:]
		if m := n - protowire.SizeVarint(v); m > 0 {
			p.out[len(p.out)-1] = Denormalized{uint(m), p.out[len(p.out)-1]}
		}
	}
}

func (m Message) Format(s fmt.State, r rune) {
	switch r {
	case 'x':
		io.WriteString(s, fmt.Sprintf("%x", m.Marshal()))
	case 'X':
		io.WriteString(s, fmt.Sprintf("%X", m.Marshal()))
	case 'v':
		switch {
		case s.Flag('#'):
			io.WriteString(s, m.format(true, true))
		case s.Flag('+'):
			io.WriteString(s, m.format(false, true))
		default:
			io.WriteString(s, m.format(false, false))
		}
	default:
		panic("invalid verb: " + string(r))
	}
}

func (m Message) format(source, multi bool) string {
	var ss []string
	var prefix, nextPrefix string
	for _, v := range m {
		// Ensure certain tokens have preceding or succeeding newlines.
		prefix, nextPrefix = nextPrefix, " "
		if multi {
			switch v := v.(type) {
			case Tag: // only has preceding newline
				prefix = "\n"
			case Denormalized: // only has preceding newline
				if _, ok := v.Value.(Tag); ok {
					prefix = "\n"
				}
			case Message, Raw: // has preceding and succeeding newlines
				prefix, nextPrefix = "\n", "\n"
			}
		}

		s := formatToken(v, source, multi)
		ss = append(ss, prefix+s+",")
	}

	var s string
	if len(ss) > 0 {
		s = strings.TrimSpace(strings.Join(ss, ""))
		if multi {
			s = "\n\t" + strings.Join(strings.Split(s, "\n"), "\n\t") + "\n"
		} else {
			s = strings.TrimSuffix(s, ",")
		}
	}
	s = fmt.Sprintf("%T{%s}", m, s)
	if !source {
		s = trimPackage(s)
	}
	return s
}

func formatToken(t Token, source, multi bool) (s string) {
	switch v := t.(type) {
	case Message:
		s = v.format(source, multi)
	case LengthPrefix:
		s = formatPacked(v, source, multi)
		if s == "" {
			ms := Message(v).format(source, multi)
			s = fmt.Sprintf("%T(%s)", v, ms)
		}
	case Tag:
		s = fmt.Sprintf("%T{%d, %s}", v, v.Number, formatType(v.Type, source))
	case Bool, Varint, Svarint, Uvarint, Int32, Uint32, Float32, Int64, Uint64, Float64:
		if source {
			// Print floats in a way that preserves exact precision.
			if f, _ := v.(Float32); math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
				switch {
				case f > 0:
					s = fmt.Sprintf("%T(math.Inf(+1))", v)
				case f < 0:
					s = fmt.Sprintf("%T(math.Inf(-1))", v)
				case math.Float32bits(float32(math.NaN())) == math.Float32bits(float32(f)):
					s = fmt.Sprintf("%T(math.NaN())", v)
				default:
					s = fmt.Sprintf("%T(math.Float32frombits(0x%08x))", v, math.Float32bits(float32(f)))
				}
				break
			}
			if f, _ := v.(Float64); math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
				switch {
				case f > 0:
					s = fmt.Sprintf("%T(math.Inf(+1))", v)
				case f < 0:
					s = fmt.Sprintf("%T(math.Inf(-1))", v)
				case math.Float64bits(float64(math.NaN())) == math.Float64bits(float64(f)):
					s = fmt.Sprintf("%T(math.NaN())", v)
				default:
					s = fmt.Sprintf("%T(math.Float64frombits(0x%016x))", v, math.Float64bits(float64(f)))
				}
				break
			}
		}
		s = fmt.Sprintf("%T(%v)", v, v)
	case String, Bytes, Raw:
		s = fmt.Sprintf("%s", v)
		s = fmt.Sprintf("%T(%s)", v, formatString(s))
	case Denormalized:
		s = fmt.Sprintf("%T{+%d, %v}", v, v.Count, formatToken(v.Value, source, multi))
	default:
		panic(fmt.Sprintf("unknown type: %T", v))
	}
	if !source {
		s = trimPackage(s)
	}
	return s
}

func formatPacked(v LengthPrefix, source, multi bool) string {
	var ss []string
	for _, v := range v {
		switch v.(type) {
		case Bool, Varint, Svarint, Uvarint, Int32, Uint32, Float32, Int64, Uint64, Float64, Denormalized, Raw:
			if v, ok := v.(Denormalized); ok {
				switch v.Value.(type) {
				case Bool, Varint, Svarint, Uvarint:
				default:
					return ""
				}
			}
			ss = append(ss, formatToken(v, source, multi))
		default:
			return ""
		}
	}
	s := fmt.Sprintf("%T{%s}", v, strings.Join(ss, ", "))
	if !source {
		s = trimPackage(s)
	}
	return s
}

func formatType(t Type, source bool) (s string) {
	switch t {
	case VarintType:
		s = pkg + ".VarintType"
	case Fixed32Type:
		s = pkg + ".Fixed32Type"
	case Fixed64Type:
		s = pkg + ".Fixed64Type"
	case BytesType:
		s = pkg + ".BytesType"
	case StartGroupType:
		s = pkg + ".StartGroupType"
	case EndGroupType:
		s = pkg + ".EndGroupType"
	default:
		s = fmt.Sprintf("Type(%d)", t)
	}
	if !source {
		s = strings.TrimSuffix(trimPackage(s), "Type")
	}
	return s
}

// formatString returns a quoted string for s.
func formatString(s string) string {
	// Use quoted string if it the same length as a raw string literal.
	// Otherwise, attempt to use the raw string form.
	qs := strconv.Quote(s)
	if len(qs) == 1+len(s)+1 {
		return qs
	}

	// Disallow newlines to ensure output is a single line.
	// Disallow non-printable runes for readability purposes.
	rawInvalid := func(r rune) bool {
		return r == '`' || r == '\n' || r == utf8.RuneError || !unicode.IsPrint(r)
	}
	if strings.IndexFunc(s, rawInvalid) < 0 {
		return "`" + s + "`"
	}
	return qs
}

var pkg = path.Base(reflect.TypeOf(Tag{}).PkgPath())

func trimPackage(s string) string {
	return strings.TrimPrefix(strings.TrimPrefix(s, pkg), ".")
}

const _Tag = "tag"
const _Type = "Type"
const _Text = "value"
const _Note = "note"
const _useNote = "useNote"
const _Path = "path"

// 两个Byte数组是否相同
func isBytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func Unmarshal(in []byte, rootpath string) ([]interface{}, error) {
	if func() bool {
		for k, v := range in {
			if v < 30 && k < 20 {
				if v == 10 || v == 13 {
					if CheckValid(in) == nil {
						return true
					}
				}
				return false
			}
		}

		return true
	}() {
		return nil, errors.New("in err")
	}

	//qt.File().WriteBytestoFile(in, "C:\\Users\\qinka\\Desktop\\xx1\\"+hex.EncodeToString(qt.Md5(in, ""))+".txt")
	var m Message
	if m.Unmarshal(in) {
		return nil, errors.New("invalid")
	}

	k, er := json.Marshal(m)
	if er != nil {
		return nil, errors.New("in Marshal err")
	}
	//WriteFile(in)
	return ParseJson(string(k), rootpath)
}

func typetostring(s Type) string {
	switch s {
	default:
		return "invalid"
	case BytesType:
		return "String"
	case BytesTypeRaw:
		return "StringRaw"
	case VarintType:
		return "Varint"
	case VarintTypeRaw:
		return "VarintRaw"
	case Fixed32Type:
		return "Fixed32"
	case Fixed64Type:
		return "Fixed64"
	case PbObjectType:
		return "Object"
	case StartGroupType:
		return "Group"
	case EndGroupType:
		return "EndGroup"

	}
}
func stringtotype(s string) Type {
	switch s {
	default:
		return BytesType
	case "EndGroup":
		return EndGroupType
	case "StringRaw":
		return BytesTypeRaw
	case "Varint":
		return VarintType
	case "VarintRaw":
		return VarintTypeRaw
	case "Fixed32":
		return Fixed32Type
	case "Fixed64":
		return Fixed64Type
	case "Object":
		return PbObjectType
	case "Group":
		return StartGroupType
	}
}

func WriteFile(w []byte) {
	im++
	file, err := os.Create("C:\\Users\\qinka\\Desktop\\x\\" + strconv.Itoa(im) + ".txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	// 写入文件内容
	writer := bufio.NewWriter(file)
	_, err = writer.Write(w)
	if err != nil {
		panic(err)
	}
	writer.Flush()
}

var im = 0

func ParseJson(k, rootpath string) ([]interface{}, error) {
	SyJson := JSON.NewSyJson()
	var Ret []interface{}
	SyJson.Parse(k)
	ElementLength := SyJson.GetNum("")
	var Target interface{}
	for i := 0; i < ElementLength; i++ {
		PATH := "[" + strconv.Itoa(i) + "]"
		TypeStr := SyJson.GetData(PATH + ".Type")
		if TypeStr != "" {
			Type1, _ := strconv.Atoi(TypeStr)
			_Type_ := protowire.Type(Type1)
			_Number_, _ := strconv.ParseInt(SyJson.GetData(PATH+".Number"), 10, 64)
			if _Type_ == EndGroupType {
				s3 := make(map[string]interface{})
				Target = s3
				s3[_Tag] = _Number_
				ty := typetostring(_Type_)
				s3[_Type] = ty
				s3[_Text] = make([]interface{}, 0)
				Ret = append(Ret, Target)
				continue
			}
			if _Type_ == StartGroupType {
				s3 := make(map[string]interface{})
				Target = s3
				s3[_Tag] = _Number_
				ty := typetostring(_Type_)
				if ty == "invalid" {
					return nil, errors.New("invalid type=" + TypeStr)
				}
				s3[_Type] = ty
				s3[_Text] = make([]interface{}, 0)
				Ret = append(Ret, Target)
				continue
			}
			s3 := make(map[string]interface{})
			Target = s3
			s3[_Tag] = _Number_
			ty := typetostring(_Type_)
			if ty == "invalid" {
				return nil, errors.New("invalid type=" + TypeStr)
			}
			s3[_Type] = ty
		} else {
			if Target == nil {
				return nil, errors.New("type err")
			}
			s4 := Target.(map[string]interface{})
			if s4 != nil {
				Text := SyJson.GetData("[" + strconv.Itoa(i) + "]")
				if s4[_Type] == typetostring(StartGroupType) {
					s4[_Text], _ = ParseJson(Text, getpath(rootpath, len(Ret)))

					continue
				}

				if s4[_Type] != typetostring(BytesType) {
					Text1, _ := strconv.ParseFloat(Text, 64)
					opty := stringtotype(s4[_Type].(string))
					switch opty {
					case Fixed64Type:
						s4[_Text] = Text1
						s4[_Path] = getpath(rootpath, len(Ret)) + ".value"
						break
					case Fixed32Type:
						s4[_Text] = Text1
						s4[_Path] = getpath(rootpath, len(Ret)) + ".value"

						break
					case VarintTypeRaw:
						var BytesToInt = func(bys []byte) int {
							bytebuff := bytes.NewBuffer(bys)
							var data int64
							binary.Read(bytebuff, binary.BigEndian, &data)
							return int(data)
						}
						s4[_Text] = Text
						b, _ := base64.StdEncoding.DecodeString(Text)
						s4[_Note] = BytesToInt(b)
						s4[_Path] = getpath(rootpath, len(Ret)) + ".value"

						break
					default:
						s4[_Text] = Text
						s4[_Path] = getpath(rootpath, len(Ret)) + ".value"
					}
					continue
				}
				n, e := base64.StdEncoding.DecodeString(Text)
				if e != nil {
					s4[_Text] = Text
					s4[_Path] = getpath(rootpath, len(Ret)) + ".value"
					continue
				}

				str1, e := Unmarshal(n, getpath(rootpath, len(Ret))+".value")
				//Type -> invalid
				if e != nil {
					ns := string(n)
					if IsChineseChar(ns) || strings.Contains(e.Error(), "invalid") || strings.Contains(e.Error(), "type err") {
						s4[_Text] = Text
						s4[_Note] = string(n)
						s4[_Type] = typetostring(BytesTypeRaw)
						s4[_Path] = getpath(rootpath, len(Ret)) + ".value"

						continue
					}
					s4[_Text] = ns //url.QueryEscape(ns)
					s4[_Path] = getpath(rootpath, len(Ret)) + ".value"

					continue
				}
				s4[_Text] = str1
				s4[_Note] = Text
				s4[_useNote] = false
				s4["INF"] = "如果您觉得此tag解析是错误的,请手动修改'Note'的Base64值,并且将'useNote'设置为 true 那么在json转PB时,将使用 Note 值作为属性值而不是 value"
				s4[_Type] = typetostring(PbObjectType)
				s4[_Path] = getpath(rootpath, len(Ret)) + ".note"
				//s4[_Path] = rootpath + ".value"
			}
			continue
		}
		Ret = append(Ret, Target)
	}
	return Ret, nil
}

func getpath(rootpath string, l int) string {
	if rootpath == "" {
		return "[" + strconv.Itoa(l-1) + "]"
	}
	return rootpath + ".[" + strconv.Itoa(l-1) + "]"
}
func IsChineseChar(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) || (regexp.MustCompile("[\u3002\uff1b\uff0c\uff1a\u201c\u201d\uff08\uff09\u3001\uff1f\u300a\u300b\ufffd]").MatchString(string(r))) {
			return true
		}
	}
	return false
}

func Marshal(in string) []byte {
	var src interface{}
	j := "{\"/\":" + in + "}"
	var t []Token
	e := json.Unmarshal([]byte(j), &src)
	if e != nil {
		return []byte{}
	}
	g := src.(map[string]interface{})
	if g == nil {
		return []byte{}
	}
	if g["/"] == nil {
		return []byte{}
	}
	gg := g["/"].([]interface{})
	if g == nil {
		return []byte{}
	}
	for _, v := range gg {
		switch f := v.(type) {
		case map[string]interface{}:
			a1 := pmap(f)
			if len(a1) > 0 {
				for i := 0; i < len(a1); i++ {
					t = append(t, a1[i])
				}
			}
		case []interface{}:
		}
	}
	var message Message
	message = t
	return message.Marshal()
}
func par(p []interface{}) []byte {
	var t []Token
	for _, v := range p {
		switch f := v.(type) {
		case map[string]interface{}:
			a1 := pmap(f)
			if len(a1) > 0 {
				for i := 0; i < len(a1); i++ {
					t = append(t, a1[i])
				}
			}
		}
	}
	var message Message
	message = t
	return message.Marshal()
}

func pmap(p map[string]interface{}) []Token {

	Content := p[_Text]
	_Number_ := p[_Tag].(float64)
	_Type_ := stringtotype(p[_Type].(string))
	var ret []Token
	if _Type_ == PbObjectType {
		ret = append(ret, Tag{Number: protowire.Number(_Number_), Type: BytesType})
		isUseNoteObj := p[_useNote]
		if isUseNoteObj != nil {
			isUseNote := isUseNoteObj.(bool)
			if isUseNote {
				NoteObj := p[_Note]
				if NoteObj != nil {
					NoteStr := NoteObj.(string)
					bs, e := base64.StdEncoding.DecodeString(NoteStr)
					if e == nil {
						ret = append(ret, Bytes(bs))
						return ret
					}
				}
				ret = append(ret, Bytes(""))
				return ret
			}
		}
		if Content == nil {
			ret = append(ret, Bytes(""))
		} else {
			ret = append(ret, Bytes(par(Content.([]interface{}))))
		}

		return ret
	}
	if _Type_ == VarintTypeRaw {
		var s Token
		s = Tag{Number: protowire.Number(_Number_), Type: VarintType}
		ret = append(ret, s)
		aRaw, _ := base64.StdEncoding.DecodeString(Content.(string))
		ret = append(ret, Raw(aRaw))
		return ret
	}
	if _Type_ == BytesTypeRaw {
		var s Token
		s = Tag{Number: protowire.Number(_Number_), Type: BytesType}
		ret = append(ret, s)
		aRaw, _ := base64.StdEncoding.DecodeString(Content.(string))
		ret = append(ret, Bytes(aRaw))
		return ret
	}

	if _Type_ == StartGroupType {
		ret = append(ret, Tag{Number: protowire.Number(_Number_), Type: StartGroupType})
		if Content != nil {
			x := Content.([]interface{})
			for i := 0; i < len(x); i++ {
				xx := x[i].(map[string]interface{})
				if xx != nil {
					a1 := pmap(xx)
					for ix := 0; ix < len(a1); ix++ {
						ret = append(ret, a1[ix])
					}
				}
			}
		}
		//ret = append(ret, Tag{Number: protowire.Number(_Number_), Type: EndGroupType})
		return ret
	}
	if _Type_ == EndGroupType {
		ret = append(ret, Tag{Number: protowire.Number(_Number_), Type: EndGroupType})
		return ret
	}
	if _Type_ == BytesType {
		var s = Tag{Number: protowire.Number(_Number_), Type: BytesType}
		ret = append(ret, s)
		us := Content.(string)
		//us, _ = url.QueryUnescape(us)
		ret = append(ret, String(us))
		return ret
	}
	if _Type_ == VarintType {
		var s Token
		s = Tag{Number: protowire.Number(_Number_), Type: VarintType}
		ret = append(ret, s)
		Text1 := Varint(0)
		switch v := Content.(type) {
		case string:
			Text2, _ := strconv.ParseUint(v, 10, 64)
			Text1 = Varint(Text2)
		case float64:
			Text1 = Varint(v)
		case float32:
			Text1 = Varint(v)
		case nil:
			Text1 = 0
		default:
			Text2, _ := strconv.ParseUint(fmt.Sprintf("%d", Content), 10, 64)
			Text1 = Varint(Text2)
		}
		ret = append(ret, Text1)
		return ret
	}
	var ConVel = Float64(0)
	if Content != nil {
		switch v := Content.(type) {
		case string:
			Convex, _ := strconv.ParseUint(v, 10, 64)
			ConVel = Float64(Convex)
			break
		case float64:
			ConVel = Float64(v)
		case float32:
			ConVel = Float64(v)
		case int:
			ConVel = Float64(v)
		case int8:
			ConVel = Float64(v)
		case int16:
			ConVel = Float64(v)
		case int32:
			ConVel = Float64(v)
		case Float64:
			ConVel = Float64(v)
		case Float32:
			ConVel = Float64(v)
		case byte:
			ConVel = Float64(v)
		default:
			ConVel = Float64(v.(float64))
		}
	}
	if _Type_ == protowire.Fixed32Type {
		ret = append(ret, Tag{Number: protowire.Number(_Number_), Type: Fixed32Type})
		ret = append(ret, Float32(ConVel))
		return ret
	}
	if _Type_ == protowire.Fixed64Type {
		var s = Tag{Number: protowire.Number(_Number_), Type: Fixed64Type}
		ret = append(ret, s)
		ret = append(ret, Float64(ConVel))
		return ret
	}
	return ret
}
func CheckValid(data []byte) error {
	var d decodeState
	err := checkValid(data, &d.scan)
	if err != nil {
		return err
	}
	return nil
}

// decodeState represents the state while decoding a JSON value.
type decodeState struct {
	data         []byte
	off          int // read offset in data
	scan         scanner
	nextscan     scanner  // for calls to nextValue
	errorContext struct { // provides context for type errors
		Struct string
		Field  string
	}
	savedError error
	useNumber  bool
}
