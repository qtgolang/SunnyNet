package Api

import "C"
import (
	"encoding/json"
	"github.com/qtgolang/SunnyNet/src/protobuf"
	"strings"
)

func PbToJson(data []byte) string {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	var msg protobuf.Message
	msg.Unmarshal(data)
	b, e := json.Marshal(msg)
	if e != nil {
		return ""
	}
	PJson, _ := protobuf.ParseJson(string(b), "")
	s, _ := json.MarshalIndent(PJson, "", "\t")
	ss := string(s)
	ss = strings.ReplaceAll(ss, "\n", "\r\n")
	return ss
}

func JsonToPB(data string) []byte {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	b := protobuf.Marshal(data)
	return b
}
