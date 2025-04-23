package public

import "github.com/qtgolang/SunnyNet/src/http"

const (
	HTTP2_Fingerprint_Config_Firefox            = `{"ConnectionFlow":12517377,"HeaderPriority":{"StreamDep":13,"Exclusive":false,"Weight":41},"Priorities":[{"PriorityParam":{"StreamDep":0,"Exclusive":false,"Weight":200},"StreamID":3},{"PriorityParam":{"StreamDep":0,"Exclusive":false,"Weight":100},"StreamID":5},{"PriorityParam":{"StreamDep":0,"Exclusive":false,"Weight":0},"StreamID":7},{"PriorityParam":{"StreamDep":7,"Exclusive":false,"Weight":0},"StreamID":9},{"PriorityParam":{"StreamDep":3,"Exclusive":false,"Weight":0},"StreamID":11},{"PriorityParam":{"StreamDep":0,"Exclusive":false,"Weight":240},"StreamID":13}],"PseudoHeaderOrder":[":method",":path",":authority",":scheme"],"Settings":{"1":65536,"4":131072,"5":16384},"SettingsOrder":[1,4,5]}`
	HTTP2_Fingerprint_Config_Opera              = `{"ConnectionFlow":15663105,"HeaderPriority":null,"Priorities":null,"PseudoHeaderOrder":[":method",":authority",":scheme",":path"],"Settings":{"1":65536,"3":1000,"4":6291456,"6":262144},"SettingsOrder":[1,3,4,6]}`
	HTTP2_Fingerprint_Config_Safari_IOS_17_0    = `{"ConnectionFlow":10485760,"HeaderPriority":null,"Priorities":null,"PseudoHeaderOrder":[":method",":scheme",":path",":authority"],"Settings":{"2":0,"3":100,"4":2097152},"SettingsOrder":[2,4,3]}`
	HTTP2_Fingerprint_Config_Safari_IOS_16_0    = `{"ConnectionFlow":10485760,"HeaderPriority":null,"Priorities":null,"PseudoHeaderOrder":[":method",":scheme",":path",":authority"],"Settings":{"3":100,"4":2097152},"SettingsOrder":[4,3]}`
	HTTP2_Fingerprint_Config_Safari             = `{"ConnectionFlow":10485760,"HeaderPriority":null,"Priorities":null,"PseudoHeaderOrder":[":method",":scheme",":path",":authority"],"Settings":{"3":100,"4":4194304},"SettingsOrder":[4,3]}`
	HTTP2_Fingerprint_Config_Chrome_117_120_124 = `{"ConnectionFlow":15663105,"HeaderPriority":null,"Priorities":null,"PseudoHeaderOrder":[":method",":authority",":scheme",":path"],"Settings":{"1":65536,"2":0,"4":6291456,"6":262144},"SettingsOrder":[1,2,4,6]}`
	HTTP2_Fingerprint_Config_Chrome_106_116     = `{"ConnectionFlow":15663105,"HeaderPriority":null,"Priorities":null,"PseudoHeaderOrder":[":method",":authority",":scheme",":path"],"Settings":{"1":65536,"2":0,"3":1000,"4":6291456,"6":262144},"SettingsOrder":[1,2,3,4,6]}`
	HTTP2_Fingerprint_Config_Chrome_103_105     = `{"ConnectionFlow":15663105,"HeaderPriority":null,"Priorities":null,"PseudoHeaderOrder":[":method",":authority",":scheme",":path"],"Settings":{"1":65536,"3":1000,"4":6291456,"6":262144},"SettingsOrder":[1,3,4,6]}`
)

var HTTP2NextProtos = []string{http.H11Proto, http.H2Proto}
var HTTP1NextProtos = []string{http.H11Proto}
