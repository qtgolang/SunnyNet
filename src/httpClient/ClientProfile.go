package httpClient

import (
	"github.com/qtgolang/SunnyNet/src/crypto/tls"
	"github.com/qtgolang/SunnyNet/src/http/http2"
)

type ClientProfile struct {
	clientHelloId     tls.ClientHelloID
	connectionFlow    uint32
	headerPriority    *http2.PriorityParam
	priorities        []http2.Priority
	pseudoHeaderOrder []string
	settings          map[http2.SettingID]uint32
	settingsOrder     []http2.SettingID
}

func NewClientProfile(clientHelloId tls.ClientHelloID, settings map[http2.SettingID]uint32, settingsOrder []http2.SettingID, pseudoHeaderOrder []string, connectionFlow uint32, priorities []http2.Priority, headerPriority *http2.PriorityParam) ClientProfile {
	return ClientProfile{
		clientHelloId:     clientHelloId,
		settings:          settings,
		settingsOrder:     settingsOrder,
		pseudoHeaderOrder: pseudoHeaderOrder,
		connectionFlow:    connectionFlow,
		priorities:        priorities,
		headerPriority:    headerPriority,
	}
}

func (c ClientProfile) GetClientHelloSpec() (tls.ClientHelloSpec, error) {
	return c.clientHelloId.ToSpec()
}

func (c ClientProfile) GetClientHelloStr() string {
	return c.clientHelloId.Str()
}

func (c ClientProfile) GetSettings() map[http2.SettingID]uint32 {
	return c.settings
}

func (c ClientProfile) GetSettingsOrder() []http2.SettingID {
	return c.settingsOrder
}

func (c ClientProfile) GetConnectionFlow() uint32 {
	return c.connectionFlow
}

func (c ClientProfile) GetPseudoHeaderOrder() []string {
	return c.pseudoHeaderOrder
}

func (c ClientProfile) GetHeaderPriority() *http2.PriorityParam {
	return c.headerPriority
}

func (c ClientProfile) GetClientHelloId() tls.ClientHelloID {
	return c.clientHelloId
}

func (c ClientProfile) GetPriorities() []http2.Priority {
	return c.priorities
}
