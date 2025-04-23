package http

import (
	"encoding/json"
	"github.com/qtgolang/SunnyNet/src/crypto/tls"
)

// PriorityParam are the stream prioritzation parameters.
type PriorityParam struct {
	// StreamDep is a 31-bit stream identifier for the
	// stream that this stream depends on. Zero means no
	// dependency.
	StreamDep uint32

	// Exclusive is whether the dependency is exclusive.
	Exclusive bool

	// Weight is the stream's zero-indexed weight. It should be
	// set together with StreamDep, or neither should be set. Per
	// the spec, "Add one to the value to obtain a weight between
	// 1 and 256."
	Weight uint8
}
type Priority struct {
	PriorityParam PriorityParam
	StreamID      uint32
}
type SettingID uint16

type H2Config struct {
	//clientHelloId     tls.ClientHelloID // 因为暂时用不上 ClientHelloID
	connectionFlow    uint32
	headerPriority    *PriorityParam
	priorities        []Priority
	pseudoHeaderOrder []string
	settings          map[SettingID]uint32
	settingsOrder     []SettingID
}
type h2Config struct {
	ConnectionFlow    uint32
	HeaderPriority    *PriorityParam
	Priorities        []Priority
	PseudoHeaderOrder []string
	Settings          map[SettingID]uint32
	SettingsOrder     []SettingID
}

func StringToH2Config(config string) (*H2Config, error) {
	var h h2Config
	e := json.Unmarshal([]byte(config), &h)
	if e != nil {
		return nil, e
	}
	return &H2Config{
		//clientHelloId:     clientHelloId,
		settings:          h.Settings,
		settingsOrder:     h.SettingsOrder,
		pseudoHeaderOrder: h.PseudoHeaderOrder,
		connectionFlow:    h.ConnectionFlow,
		priorities:        h.Priorities,
		headerPriority:    h.HeaderPriority,
	}, nil

}
func (c *H2Config) String() string {
	h := h2Config{
		ConnectionFlow:    c.connectionFlow,
		HeaderPriority:    c.headerPriority,
		Priorities:        c.priorities,
		PseudoHeaderOrder: c.pseudoHeaderOrder,
		Settings:          c.settings,
		SettingsOrder:     c.settingsOrder,
	}
	b, _ := json.Marshal(h)
	return string(b)
}
func NewH2Config(settings map[SettingID]uint32, settingsOrder []SettingID, pseudoHeaderOrder []string, connectionFlow uint32, priorities []Priority, headerPriority *PriorityParam) *H2Config {
	//clientHelloId tls.ClientHelloID,
	return &H2Config{
		//clientHelloId:     clientHelloId,
		settings:          settings,
		settingsOrder:     settingsOrder,
		pseudoHeaderOrder: pseudoHeaderOrder,
		connectionFlow:    connectionFlow,
		priorities:        priorities,
		headerPriority:    headerPriority,
	}
}

func (c H2Config) GetClientHelloSpec() (tls.ClientHelloSpec, error) {
	//return c.clientHelloId.ToSpec()
	return tls.ClientHelloSpec{}, nil
}

func (c H2Config) GetClientHelloStr() string {
	//return c.clientHelloId.Str()
	return ""
}

func (c H2Config) GetSettings() map[SettingID]uint32 {
	return c.settings
}

func (c H2Config) GetSettingsOrder() []SettingID {
	return c.settingsOrder
}

func (c H2Config) GetConnectionFlow() uint32 {
	return c.connectionFlow
}

func (c H2Config) GetPseudoHeaderOrder() []string {
	return c.pseudoHeaderOrder
}

func (c H2Config) GetHeaderPriority() *PriorityParam {
	return c.headerPriority
}

func (c H2Config) GetClientHelloId() tls.ClientHelloID {
	//return c.clientHelloId
	return tls.ClientHelloID{}
}

func (c H2Config) GetPriorities() []Priority {
	return c.priorities
}

const (
	SettingHeaderTableSize      SettingID = 0x1
	SettingEnablePush           SettingID = 0x2
	SettingMaxConcurrentStreams SettingID = 0x3
	SettingInitialWindowSize    SettingID = 0x4
	SettingMaxFrameSize         SettingID = 0x5
	SettingMaxHeaderListSize    SettingID = 0x6
)

var SettingName = map[SettingID]string{
	SettingHeaderTableSize:      "HEADER_TABLE_SIZE",
	SettingEnablePush:           "ENABLE_PUSH",
	SettingMaxConcurrentStreams: "MAX_CONCURRENT_STREAMS",
	SettingInitialWindowSize:    "INITIAL_WINDOW_SIZE",
	SettingMaxFrameSize:         "MAX_FRAME_SIZE",
	SettingMaxHeaderListSize:    "MAX_HEADER_LIST_SIZE",
}

type config_http2 struct {
	Chrome_103        *H2Config
	Chrome_104        *H2Config
	Chrome_105        *H2Config
	Chrome_106        *H2Config
	Chrome_107        *H2Config
	Chrome_108        *H2Config
	Chrome_109        *H2Config
	Chrome_110        *H2Config
	Chrome_111        *H2Config
	Chrome_112        *H2Config
	Chrome_116_PSK    *H2Config
	Chrome_116_PSK_PQ *H2Config
	Chrome_117        *H2Config
	Chrome_120        *H2Config
	Chrome_124        *H2Config
	Safari_15_6_1     *H2Config
	Safari_16_0       *H2Config
	Safari_Ipad_15_6  *H2Config
	Safari_IOS_15_5   *H2Config
	Safari_IOS_15_6   *H2Config
	Safari_IOS_16_0   *H2Config
	Safari_IOS_17_0   *H2Config
	Opera_89          *H2Config
	Opera_91          *H2Config
	Opera_90          *H2Config
	Firefox_102       *H2Config
	Firefox_104       *H2Config
	Firefox_105       *H2Config
	Firefox_106       *H2Config
	Firefox_108       *H2Config
	Firefox_110       *H2Config
	Firefox_117       *H2Config
}

func (c *config_http2) String() string {
	s := ""
	s += "Chrome_103:" + c.Chrome_103.String() + "\r\n"
	s += "Chrome_104:" + c.Chrome_104.String() + "\r\n"
	s += "Chrome_105:" + c.Chrome_105.String() + "\r\n"
	s += "Chrome_106:" + c.Chrome_106.String() + "\r\n"
	s += "Chrome_107:" + c.Chrome_107.String() + "\r\n"
	s += "Chrome_108:" + c.Chrome_108.String() + "\r\n"
	s += "Chrome_109:" + c.Chrome_109.String() + "\r\n"
	s += "Chrome_110:" + c.Chrome_110.String() + "\r\n"
	s += "Chrome_111:" + c.Chrome_111.String() + "\r\n"
	s += "Chrome_112:" + c.Chrome_112.String() + "\r\n"
	s += "Chrome_116_PSK:" + c.Chrome_116_PSK.String() + "\r\n"
	s += "Chrome_116_PSK_PQ:" + c.Chrome_116_PSK_PQ.String() + "\r\n"
	s += "Chrome_117:" + c.Chrome_117.String() + "\r\n"
	s += "Chrome_120:" + c.Chrome_120.String() + "\r\n"
	s += "Chrome_124:" + c.Chrome_124.String() + "\r\n"
	s += "Safari_15_6_1:" + c.Safari_15_6_1.String() + "\r\n"
	s += "Safari_16_0:" + c.Safari_16_0.String() + "\r\n"
	s += "Safari_Ipad_15_6:" + c.Safari_Ipad_15_6.String() + "\r\n"
	s += "Safari_IOS_15_6:" + c.Safari_IOS_15_6.String() + "\r\n"
	s += "Safari_IOS_15_6:" + c.Safari_IOS_15_6.String() + "\r\n"
	s += "Safari_IOS_16_0:" + c.Safari_IOS_16_0.String() + "\r\n"
	s += "Safari_IOS_17_0:" + c.Safari_IOS_17_0.String() + "\r\n"
	s += "Opera_89:" + c.Opera_89.String() + "\r\n"
	s += "Opera_91:" + c.Opera_91.String() + "\r\n"
	s += "Opera_90:" + c.Opera_90.String() + "\r\n"
	s += "Firefox_102:" + c.Firefox_102.String() + "\r\n"
	s += "Firefox_104:" + c.Firefox_104.String() + "\r\n"
	s += "Firefox_105:" + c.Firefox_105.String() + "\r\n"
	s += "Firefox_106:" + c.Firefox_106.String() + "\r\n"
	s += "Firefox_108:" + c.Firefox_108.String() + "\r\n"
	s += "Firefox_110:" + c.Firefox_110.String() + "\r\n"
	s += "Firefox_117:" + c.Firefox_117.String() + "\r\n"
	return s
}

const h2ConfigKey = "h2Config"

var ConfigH2 config_http2

func init() {
	ConfigH2.Safari_IOS_15_5 = &H2Config{
		//clientHelloId: tls.HelloIOS_15_5,
		settings: map[SettingID]uint32{
			SettingInitialWindowSize:    2097152,
			SettingMaxConcurrentStreams: 100,
		},
		settingsOrder: []SettingID{
			SettingInitialWindowSize,
			SettingMaxConcurrentStreams,
		},
		pseudoHeaderOrder: []string{
			":method",
			":scheme",
			":path",
			":authority",
		},
		connectionFlow: 10485760,
	}

	ConfigH2.Chrome_117 = &H2Config{
		/*
			clientHelloId: tls.ClientHelloID{
				Client:               "Chrome",
				RandomExtensionOrder: false,
				Version:              "117",
				Seed:                 nil,
				SpecFactory: func() (tls.ClientHelloSpec, error) {
					return tls.ClientHelloSpec{
						CipherSuites: []uint16{
							tls.GREASE_PLACEHOLDER,
							tls.TLS_AES_128_GCM_SHA256,
							tls.TLS_AES_256_GCM_SHA384,
							tls.TLS_CHACHA20_POLY1305_SHA256,
							tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
							tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
							tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
							tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
							tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
							tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
							tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
							tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
							tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
							tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
							tls.TLS_RSA_WITH_AES_128_CBC_SHA,
							tls.TLS_RSA_WITH_AES_256_CBC_SHA,
						},
						CompressionMethods: []uint8{
							tls.CompressionNone,
						},
						Extensions: []tls.TLSExtension{
							&tls.UtlsGREASEExtension{},
							&tls.PSKKeyExchangeModesExtension{[]uint8{
								tls.PskModeDHE,
							}},
							&tls.SNIExtension{},
							&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
							&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
								tls.ECDSAWithP256AndSHA256,
								tls.PSSWithSHA256,
								tls.PKCS1WithSHA256,
								tls.ECDSAWithP384AndSHA384,
								tls.PSSWithSHA384,
								tls.PKCS1WithSHA384,
								tls.PSSWithSHA512,
								tls.PKCS1WithSHA512,
							}},
							&tls.SupportedVersionsExtension{[]uint16{
								tls.GREASE_PLACEHOLDER,
								tls.VersionTLS13,
								tls.VersionTLS12,
							}},
							&tls.ApplicationSettingsExtension{SupportedProtocols: []string{"h2"}},
							&tls.SupportedCurvesExtension{[]tls.CurveID{
								tls.CurveID(tls.GREASE_PLACEHOLDER),
								tls.X25519,
								tls.CurveP256,
								tls.CurveP384,
							}},
							&tls.ExtendedMasterSecretExtension{},
							&tls.SessionTicketExtension{},
							&tls.UtlsCompressCertExtension{[]tls.CertCompressionAlgo{
								tls.CertCompressionBrotli,
							}},
							&tls.SCTExtension{},
							&tls.StatusRequestExtension{},
							&tls.KeyShareExtension{[]tls.KeyShare{
								{Group: tls.CurveID(tls.GREASE_PLACEHOLDER), Data: []byte{0}},
								{Group: tls.X25519},
							}},
							&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateOnceAsClient},
							&tls.SupportedPointsExtension{SupportedPoints: []byte{
								tls.PointFormatUncompressed,
							}},
							&tls.UtlsGREASEExtension{},
							&tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
						},
					}, nil
				},
			},
		*/
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:   65536,
			SettingEnablePush:        0,
			SettingInitialWindowSize: 6291456,
			SettingMaxHeaderListSize: 262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingEnablePush,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}

	ConfigH2.Chrome_124 = &H2Config{
		/*
			clientHelloId: tls.ClientHelloID{
				Client:               "Chrome",
				RandomExtensionOrder: false,
				Version:              "124",
				Seed:                 nil,
				SpecFactory: func() (tls.ClientHelloSpec, error) {
					return tls.ClientHelloSpec{
						CipherSuites: []uint16{
							tls.GREASE_PLACEHOLDER,
							tls.TLS_AES_128_GCM_SHA256,
							tls.TLS_AES_256_GCM_SHA384,
							tls.TLS_CHACHA20_POLY1305_SHA256,
							tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
							tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
							tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
							tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
							tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
							tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
							tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
							tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
							tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
							tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
							tls.TLS_RSA_WITH_AES_128_CBC_SHA,
							tls.TLS_RSA_WITH_AES_256_CBC_SHA,
						},
						CompressionMethods: []uint8{
							tls.CompressionNone,
						},
						Extensions: []tls.TLSExtension{
							&tls.UtlsGREASEExtension{},
							&tls.UtlsCompressCertExtension{[]tls.CertCompressionAlgo{
								tls.CertCompressionBrotli,
							}},
							&tls.SCTExtension{},
							&tls.ExtendedMasterSecretExtension{},
							&tls.ApplicationSettingsExtension{SupportedProtocols: []string{"h2"}},
							&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
							&tls.SupportedVersionsExtension{[]uint16{
								tls.GREASE_PLACEHOLDER,
								tls.VersionTLS13,
								tls.VersionTLS12,
							}},
							&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
								tls.ECDSAWithP256AndSHA256,
								tls.PSSWithSHA256,
								tls.PKCS1WithSHA256,
								tls.ECDSAWithP384AndSHA384,
								tls.PSSWithSHA384,
								tls.PKCS1WithSHA384,
								tls.PSSWithSHA512,
								tls.PKCS1WithSHA512,
							}},
							&tls.SupportedPointsExtension{SupportedPoints: []byte{
								tls.PointFormatUncompressed,
							}},
							&tls.SNIExtension{},
							&tls.SessionTicketExtension{},
							&tls.SupportedCurvesExtension{[]tls.CurveID{
								tls.GREASE_PLACEHOLDER,
								tls.X25519Kyber768Draft00,
								tls.X25519,
								tls.CurveP256,
								tls.CurveP384,
							}},
							tls.BoringGREASEECH(),
							&tls.StatusRequestExtension{},
							&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateOnceAsClient},
							&tls.PSKKeyExchangeModesExtension{[]uint8{
								tls.PskModeDHE,
							}},
							&tls.KeyShareExtension{[]tls.KeyShare{
								{Group: tls.CurveID(tls.GREASE_PLACEHOLDER), Data: []byte{0}},
								{Group: tls.X25519Kyber768Draft00},
								{Group: tls.X25519},
							}},
							&tls.UtlsGREASEExtension{},
						},
					}, nil
				},
			},
		*/
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:   65536,
			SettingEnablePush:        0,
			SettingInitialWindowSize: 6291456,
			SettingMaxHeaderListSize: 262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingEnablePush,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}

	ConfigH2.Chrome_120 = &H2Config{
		/*
			clientHelloId: tls.ClientHelloID{
				Client:               "Chrome",
				RandomExtensionOrder: false,
				Version:              "120",
				Seed:                 nil,
				SpecFactory: func() (tls.ClientHelloSpec, error) {
					return tls.ClientHelloSpec{
						CipherSuites: []uint16{
							tls.GREASE_PLACEHOLDER,
							tls.TLS_AES_128_GCM_SHA256,
							tls.TLS_AES_256_GCM_SHA384,
							tls.TLS_CHACHA20_POLY1305_SHA256,
							tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
							tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
							tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
							tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
							tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
							tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
							tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
							tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
							tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
							tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
							tls.TLS_RSA_WITH_AES_128_CBC_SHA,
							tls.TLS_RSA_WITH_AES_256_CBC_SHA,
						},
						CompressionMethods: []uint8{
							tls.CompressionNone,
						},
						Extensions: []tls.TLSExtension{
							&tls.UtlsGREASEExtension{},
							&tls.SNIExtension{},
							&tls.PSKKeyExchangeModesExtension{[]uint8{
								tls.PskModeDHE,
							}},
							&tls.SupportedVersionsExtension{[]uint16{
								tls.GREASE_PLACEHOLDER,
								tls.VersionTLS13,
								tls.VersionTLS12,
							}},
							&tls.StatusRequestExtension{},
							&tls.ExtendedMasterSecretExtension{},
							&tls.SessionTicketExtension{},
							&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
								tls.ECDSAWithP256AndSHA256,
								tls.PSSWithSHA256,
								tls.PKCS1WithSHA256,
								tls.ECDSAWithP384AndSHA384,
								tls.PSSWithSHA384,
								tls.PKCS1WithSHA384,
								tls.PSSWithSHA512,
								tls.PKCS1WithSHA512,
							}},
							&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateOnceAsClient},
							&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
							tls.BoringGREASEECH(),
							&tls.SCTExtension{},
							&tls.KeyShareExtension{[]tls.KeyShare{
								{Group: tls.CurveID(tls.GREASE_PLACEHOLDER), Data: []byte{0}},
								{Group: tls.X25519},
							}},
							&tls.SupportedCurvesExtension{[]tls.CurveID{
								tls.GREASE_PLACEHOLDER,
								tls.X25519,
								tls.CurveP256,
								tls.CurveP384,
							}},
							&tls.SupportedPointsExtension{SupportedPoints: []byte{
								tls.PointFormatUncompressed,
							}},
							&tls.ApplicationSettingsExtension{SupportedProtocols: []string{"h2"}},
							&tls.UtlsCompressCertExtension{[]tls.CertCompressionAlgo{
								tls.CertCompressionBrotli,
							}},
							&tls.UtlsGREASEExtension{},
						},
					}, nil
				},
			},
		*/
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:   65536,
			SettingEnablePush:        0,
			SettingInitialWindowSize: 6291456,
			SettingMaxHeaderListSize: 262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingEnablePush,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}

	ConfigH2.Chrome_112 = &H2Config{
		//clientHelloId: tls.HelloChrome_112,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:      65536,
			SettingEnablePush:           0,
			SettingMaxConcurrentStreams: 1000,
			SettingInitialWindowSize:    6291456,
			SettingMaxHeaderListSize:    262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingEnablePush,
			SettingMaxConcurrentStreams,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}

	ConfigH2.Chrome_116_PSK = &H2Config{
		//clientHelloId: tls.HelloChrome_112_PSK,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:      65536,
			SettingEnablePush:           0,
			SettingMaxConcurrentStreams: 1000,
			SettingInitialWindowSize:    6291456,
			SettingMaxHeaderListSize:    262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingEnablePush,
			SettingMaxConcurrentStreams,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}

	ConfigH2.Chrome_116_PSK_PQ = &H2Config{
		//clientHelloId: tls.HelloChrome_115_PQ_PSK,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:      65536,
			SettingEnablePush:           0,
			SettingMaxConcurrentStreams: 1000,
			SettingInitialWindowSize:    6291456,
			SettingMaxHeaderListSize:    262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingEnablePush,
			SettingMaxConcurrentStreams,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}

	ConfigH2.Chrome_111 = &H2Config{
		//clientHelloId: tls.HelloChrome_111,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:      65536,
			SettingEnablePush:           0,
			SettingMaxConcurrentStreams: 1000,
			SettingInitialWindowSize:    6291456,
			SettingMaxHeaderListSize:    262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingEnablePush,
			SettingMaxConcurrentStreams,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}

	ConfigH2.Chrome_110 = &H2Config{
		//clientHelloId: tls.HelloChrome_110,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:      65536,
			SettingEnablePush:           0,
			SettingMaxConcurrentStreams: 1000,
			SettingInitialWindowSize:    6291456,
			SettingMaxHeaderListSize:    262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingEnablePush,
			SettingMaxConcurrentStreams,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}

	ConfigH2.Chrome_109 = &H2Config{
		//clientHelloId: tls.HelloChrome_109,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:      65536,
			SettingEnablePush:           0,
			SettingMaxConcurrentStreams: 1000,
			SettingInitialWindowSize:    6291456,
			SettingMaxHeaderListSize:    262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingEnablePush,
			SettingMaxConcurrentStreams,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}

	ConfigH2.Chrome_108 = &H2Config{
		//clientHelloId: tls.HelloChrome_108,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:      65536,
			SettingEnablePush:           0,
			SettingMaxConcurrentStreams: 1000,
			SettingInitialWindowSize:    6291456,
			SettingMaxHeaderListSize:    262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingEnablePush,
			SettingMaxConcurrentStreams,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}

	ConfigH2.Chrome_107 = &H2Config{
		//clientHelloId: tls.HelloChrome_107,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:      65536,
			SettingEnablePush:           0,
			SettingMaxConcurrentStreams: 1000,
			SettingInitialWindowSize:    6291456,
			SettingMaxHeaderListSize:    262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingEnablePush,
			SettingMaxConcurrentStreams,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}

	ConfigH2.Chrome_106 = &H2Config{
		//clientHelloId: tls.HelloChrome_106,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:      65536,
			SettingEnablePush:           0,
			SettingMaxConcurrentStreams: 1000,
			SettingInitialWindowSize:    6291456,
			SettingMaxHeaderListSize:    262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingEnablePush,
			SettingMaxConcurrentStreams,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}

	ConfigH2.Chrome_105 = &H2Config{
		//clientHelloId: tls.HelloChrome_105,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:      65536,
			SettingMaxConcurrentStreams: 1000,
			SettingInitialWindowSize:    6291456,
			SettingMaxHeaderListSize:    262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingMaxConcurrentStreams,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}

	ConfigH2.Chrome_104 = &H2Config{
		//clientHelloId: tls.HelloChrome_104,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:      65536,
			SettingMaxConcurrentStreams: 1000,
			SettingInitialWindowSize:    6291456,
			SettingMaxHeaderListSize:    262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingMaxConcurrentStreams,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}

	ConfigH2.Chrome_103 = &H2Config{
		//clientHelloId: tls.HelloChrome_103,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:      65536,
			SettingMaxConcurrentStreams: 1000,
			SettingInitialWindowSize:    6291456,
			SettingMaxHeaderListSize:    262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingMaxConcurrentStreams,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}

	ConfigH2.Safari_15_6_1 = &H2Config{
		//clientHelloId: tls.HelloSafari_15_6_1,
		settings: map[SettingID]uint32{
			SettingInitialWindowSize:    4194304,
			SettingMaxConcurrentStreams: 100,
		},
		settingsOrder: []SettingID{
			SettingInitialWindowSize,
			SettingMaxConcurrentStreams,
		},
		pseudoHeaderOrder: []string{
			":method",
			":scheme",
			":path",
			":authority",
		},
		connectionFlow: 10485760,
	}

	ConfigH2.Safari_16_0 = &H2Config{
		//clientHelloId: tls.HelloSafari_16_0,
		settings: map[SettingID]uint32{
			SettingInitialWindowSize:    4194304,
			SettingMaxConcurrentStreams: 100,
		},
		settingsOrder: []SettingID{
			SettingInitialWindowSize,
			SettingMaxConcurrentStreams,
		},
		pseudoHeaderOrder: []string{
			":method",
			":scheme",
			":path",
			":authority",
		},
		connectionFlow: 10485760,
	}

	ConfigH2.Safari_Ipad_15_6 = &H2Config{
		//clientHelloId: tls.HelloIPad_15_6,
		settings: map[SettingID]uint32{
			SettingInitialWindowSize:    2097152,
			SettingMaxConcurrentStreams: 100,
		},
		settingsOrder: []SettingID{
			SettingInitialWindowSize,
			SettingMaxConcurrentStreams,
		},
		pseudoHeaderOrder: []string{
			":method",
			":scheme",
			":path",
			":authority",
		},
		connectionFlow: 10485760,
	}

	ConfigH2.Safari_IOS_17_0 = &H2Config{
		//clientHelloId: tls.HelloIOS_16_0,
		settings: map[SettingID]uint32{
			SettingEnablePush:           0,
			SettingInitialWindowSize:    2097152,
			SettingMaxConcurrentStreams: 100,
		},
		settingsOrder: []SettingID{
			SettingEnablePush,
			SettingInitialWindowSize,
			SettingMaxConcurrentStreams,
		},
		pseudoHeaderOrder: []string{
			":method",
			":scheme",
			":path",
			":authority",
		},
		connectionFlow: 10485760,
	}

	ConfigH2.Safari_IOS_16_0 = &H2Config{
		//clientHelloId: tls.HelloIOS_16_0,
		settings: map[SettingID]uint32{
			SettingInitialWindowSize:    2097152,
			SettingMaxConcurrentStreams: 100,
		},
		settingsOrder: []SettingID{
			SettingInitialWindowSize,
			SettingMaxConcurrentStreams,
		},
		pseudoHeaderOrder: []string{
			":method",
			":scheme",
			":path",
			":authority",
		},
		connectionFlow: 10485760,
	}

	ConfigH2.Safari_IOS_15_5 = &H2Config{
		//clientHelloId: tls.HelloIOS_15_5,
		settings: map[SettingID]uint32{
			SettingInitialWindowSize:    2097152,
			SettingMaxConcurrentStreams: 100,
		},
		settingsOrder: []SettingID{
			SettingInitialWindowSize,
			SettingMaxConcurrentStreams,
		},
		pseudoHeaderOrder: []string{
			":method",
			":scheme",
			":path",
			":authority",
		},
		connectionFlow: 10485760,
	}

	ConfigH2.Safari_IOS_15_6 = &H2Config{
		//clientHelloId: tls.HelloIOS_15_6,
		settings: map[SettingID]uint32{
			SettingInitialWindowSize:    2097152,
			SettingMaxConcurrentStreams: 100,
		},
		settingsOrder: []SettingID{
			SettingInitialWindowSize,
			SettingMaxConcurrentStreams,
		},
		pseudoHeaderOrder: []string{
			":method",
			":scheme",
			":path",
			":authority",
		},
		connectionFlow: 10485760,
	}

	ConfigH2.Firefox_117 = &H2Config{
		/*
			clientHelloId: tls.ClientHelloID{
				Client:               "Firefox",
				RandomExtensionOrder: false,
				Version:              "117",
				Seed:                 nil,
				SpecFactory: func() (tls.ClientHelloSpec, error) {
					return tls.ClientHelloSpec{
						CipherSuites: []uint16{
							tls.TLS_AES_128_GCM_SHA256,
							tls.TLS_CHACHA20_POLY1305_SHA256,
							tls.TLS_AES_256_GCM_SHA384,
							tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
							tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
							tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
							tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
							tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
							tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
							tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
							tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
							tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
							tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
							tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
							tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
							tls.TLS_RSA_WITH_AES_128_CBC_SHA,
							tls.TLS_RSA_WITH_AES_256_CBC_SHA,
						},
						CompressionMethods: []byte{
							tls.CompressionNone,
						},
						Extensions: []tls.TLSExtension{
							&tls.SNIExtension{},
							&tls.ExtendedMasterSecretExtension{},
							&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateOnceAsClient},
							&tls.SupportedCurvesExtension{[]tls.CurveID{
								tls.X25519,
								tls.CurveP256,
								tls.CurveP384,
								tls.CurveP521,
								tls.FAKEFFDHE2048,
								tls.FAKEFFDHE3072,
							}},
							&tls.SupportedPointsExtension{SupportedPoints: []byte{
								tls.PointFormatUncompressed,
							}},

							&tls.SessionTicketExtension{},
							&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
							&tls.StatusRequestExtension{},
							&tls.DelegatedCredentialsExtension{
								SupportedSignatureAlgorithms: []tls.SignatureScheme{
									tls.ECDSAWithP256AndSHA256,
									tls.ECDSAWithP384AndSHA384,
									tls.ECDSAWithP521AndSHA512,
									tls.ECDSAWithSHA1,
								},
							},
							&tls.KeyShareExtension{[]tls.KeyShare{
								{Group: tls.X25519},
								{Group: tls.CurveP256},
							}},
							&tls.SupportedVersionsExtension{[]uint16{
								tls.VersionTLS13,
								tls.VersionTLS12,
							}},
							&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
								tls.ECDSAWithP256AndSHA256,
								tls.ECDSAWithP384AndSHA384,
								tls.ECDSAWithP521AndSHA512,
								tls.PSSWithSHA256,
								tls.PSSWithSHA384,
								tls.PSSWithSHA512,
								tls.PKCS1WithSHA256,
								tls.PKCS1WithSHA384,
								tls.PKCS1WithSHA512,
								tls.ECDSAWithSHA1,
								tls.PKCS1WithSHA1,
							}},
							&tls.PSKKeyExchangeModesExtension{[]uint8{
								tls.PskModeDHE,
							}},
							&tls.FakeRecordSizeLimitExtension{0x4001},
							&tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
						}}, nil
				},
			},
		*/
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:   65536,
			SettingInitialWindowSize: 131072,
			SettingMaxFrameSize:      16384,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingInitialWindowSize,
			SettingMaxFrameSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":path",
			":authority",
			":scheme",
		},
		connectionFlow: 12517377,
		headerPriority: &PriorityParam{
			StreamDep: 13,
			Exclusive: false,
			Weight:    41,
		},
		priorities: []Priority{
			{StreamID: 3, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    200,
			}},
			{StreamID: 5, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    100,
			}},
			{StreamID: 7, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 9, PriorityParam: PriorityParam{
				StreamDep: 7,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 11, PriorityParam: PriorityParam{
				StreamDep: 3,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 13, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    240,
			}},
		},
	}

	ConfigH2.Firefox_110 = &H2Config{
		//clientHelloId: tls.HelloFirefox_110,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:   65536,
			SettingInitialWindowSize: 131072,
			SettingMaxFrameSize:      16384,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingInitialWindowSize,
			SettingMaxFrameSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":path",
			":authority",
			":scheme",
		},
		connectionFlow: 12517377,
		headerPriority: &PriorityParam{
			StreamDep: 13,
			Exclusive: false,
			Weight:    41,
		},
		priorities: []Priority{
			{StreamID: 3, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    200,
			}},
			{StreamID: 5, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    100,
			}},
			{StreamID: 7, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 9, PriorityParam: PriorityParam{
				StreamDep: 7,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 11, PriorityParam: PriorityParam{
				StreamDep: 3,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 13, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    240,
			}},
		},
	}

	ConfigH2.Firefox_108 = &H2Config{
		//clientHelloId: tls.HelloFirefox_108,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:   65536,
			SettingInitialWindowSize: 131072,
			SettingMaxFrameSize:      16384,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingInitialWindowSize,
			SettingMaxFrameSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":path",
			":authority",
			":scheme",
		},
		connectionFlow: 12517377,
		headerPriority: &PriorityParam{
			StreamDep: 13,
			Exclusive: false,
			Weight:    41,
		},
		priorities: []Priority{
			{StreamID: 3, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    200,
			}},
			{StreamID: 5, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    100,
			}},
			{StreamID: 7, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 9, PriorityParam: PriorityParam{
				StreamDep: 7,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 11, PriorityParam: PriorityParam{
				StreamDep: 3,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 13, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    240,
			}},
		},
	}

	ConfigH2.Firefox_106 = &H2Config{
		//clientHelloId: tls.HelloFirefox_106,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:   65536,
			SettingInitialWindowSize: 131072,
			SettingMaxFrameSize:      16384,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingInitialWindowSize,
			SettingMaxFrameSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":path",
			":authority",
			":scheme",
		},
		connectionFlow: 12517377,
		headerPriority: &PriorityParam{
			StreamDep: 13,
			Exclusive: false,
			Weight:    41,
		},
		priorities: []Priority{
			{StreamID: 3, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    200,
			}},
			{StreamID: 5, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    100,
			}},
			{StreamID: 7, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 9, PriorityParam: PriorityParam{
				StreamDep: 7,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 11, PriorityParam: PriorityParam{
				StreamDep: 3,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 13, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    240,
			}},
		},
	}

	ConfigH2.Firefox_105 = &H2Config{
		//clientHelloId: tls.HelloFirefox_105,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:   65536,
			SettingInitialWindowSize: 131072,
			SettingMaxFrameSize:      16384,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingInitialWindowSize,
			SettingMaxFrameSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":path",
			":authority",
			":scheme",
		},
		connectionFlow: 12517377,
		headerPriority: &PriorityParam{
			StreamDep: 13,
			Exclusive: false,
			Weight:    41,
		},
		priorities: []Priority{
			{StreamID: 3, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    200,
			}},
			{StreamID: 5, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    100,
			}},
			{StreamID: 7, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 9, PriorityParam: PriorityParam{
				StreamDep: 7,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 11, PriorityParam: PriorityParam{
				StreamDep: 3,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 13, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    240,
			}},
		},
	}

	ConfigH2.Firefox_104 = &H2Config{
		//clientHelloId: tls.HelloFirefox_104,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:   65536,
			SettingInitialWindowSize: 131072,
			SettingMaxFrameSize:      16384,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingInitialWindowSize,
			SettingMaxFrameSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":path",
			":authority",
			":scheme",
		},
		connectionFlow: 12517377,
		headerPriority: &PriorityParam{
			StreamDep: 13,
			Exclusive: false,
			Weight:    41,
		},
		priorities: []Priority{
			{StreamID: 3, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    200,
			}},
			{StreamID: 5, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    100,
			}},
			{StreamID: 7, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 9, PriorityParam: PriorityParam{
				StreamDep: 7,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 11, PriorityParam: PriorityParam{
				StreamDep: 3,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 13, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    240,
			}},
		},
	}

	ConfigH2.Firefox_102 = &H2Config{
		//clientHelloId: tls.HelloFirefox_102,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:   65536,
			SettingInitialWindowSize: 131072,
			SettingMaxFrameSize:      16384,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingInitialWindowSize,
			SettingMaxFrameSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":path",
			":authority",
			":scheme",
		},
		connectionFlow: 12517377,
		headerPriority: &PriorityParam{
			StreamDep: 13,
			Exclusive: false,
			Weight:    41,
		},
		priorities: []Priority{
			{StreamID: 3, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    200,
			}},
			{StreamID: 5, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    100,
			}},
			{StreamID: 7, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 9, PriorityParam: PriorityParam{
				StreamDep: 7,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 11, PriorityParam: PriorityParam{
				StreamDep: 3,
				Exclusive: false,
				Weight:    0,
			}},
			{StreamID: 13, PriorityParam: PriorityParam{
				StreamDep: 0,
				Exclusive: false,
				Weight:    240,
			}},
		},
	}

	ConfigH2.Opera_90 = &H2Config{
		//clientHelloId: tls.HelloOpera_90,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:      65536,
			SettingMaxConcurrentStreams: 1000,
			SettingInitialWindowSize:    6291456,
			SettingMaxHeaderListSize:    262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingMaxConcurrentStreams,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}

	ConfigH2.Opera_91 = &H2Config{
		//clientHelloId: tls.HelloOpera_91,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:      65536,
			SettingMaxConcurrentStreams: 1000,
			SettingInitialWindowSize:    6291456,
			SettingMaxHeaderListSize:    262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingMaxConcurrentStreams,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}

	ConfigH2.Opera_89 = &H2Config{
		//clientHelloId: tls.HelloOpera_89,
		settings: map[SettingID]uint32{
			SettingHeaderTableSize:      65536,
			SettingMaxConcurrentStreams: 1000,
			SettingInitialWindowSize:    6291456,
			SettingMaxHeaderListSize:    262144,
		},
		settingsOrder: []SettingID{
			SettingHeaderTableSize,
			SettingMaxConcurrentStreams,
			SettingInitialWindowSize,
			SettingMaxHeaderListSize,
		},
		pseudoHeaderOrder: []string{
			":method",
			":authority",
			":scheme",
			":path",
		},
		connectionFlow: 15663105,
	}
}
