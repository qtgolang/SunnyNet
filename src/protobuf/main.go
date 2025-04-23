package protobuf

import "C"
import (
	"encoding/json"
	"github.com/qtgolang/SunnyNet/src/protobuf/JSON"
	"strings"
)

func ToJson(data []byte) string {
	var msg Message
	msg.Unmarshal(data)
	b, e := json.Marshal(msg)
	if e != nil {
		return ""
	}
	PJson, _ := ParseJson(string(b), "")
	s, _ := json.MarshalIndent(PJson, "", "\t")
	ss := string(s)
	ss = strings.ReplaceAll(ss, "\n", "\r\n")
	if ss == "null" {
		return ""
	}
	if ss == "\"null\"" {
		return ""
	}
	return ss
}

func JsonToPB(data string) []byte {
	b := Marshal(data)
	if len(b) < 1 {
		return []byte{}
	}
	return b
}
func JsonParse(data string) *JSON.SyJson {
	//构造一个JSON对象
	obj := JSON.NewSyJson()
	//解析JSON
	obj.Parse(data)
	return obj
}
