package Interface

type EventFuncArgs struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type EventFunc struct {
	Name    string          `json:"name"`
	Args    []EventFuncArgs `json:"Args"`
	Returns []string        `json:"Returns"`
	Comment string          `json:"comment"`
}

type ExportEventInterface struct {
	HTTPScriptEvent      []EventFunc `json:"HTTPEvent"`
	TCPScriptEvent       []EventFunc `json:"TCPEvent"`
	UDPScriptEvent       []EventFunc `json:"UDPEvent"`
	WebSocketScriptEvent []EventFunc `json:"WebSocketEvent"`
}

var ExportEvent ExportEventInterface
