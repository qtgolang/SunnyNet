//go:build windows
// +build windows

package WinDivert

type Layer int

func (l Layer) String() string {
	switch l {
	case LayerNetwork:
		return "WINDIVERT_LAYER_NETWORK"
	case LayerNetworkForward:
		return "WINDIVERT_LAYER_NETWORK_FORWARD"
	case LayerFlow:
		return "WINDIVERT_LAYER_FLOW"
	case LayerSocket:
		return "WINDIVERT_LAYER_SOCKET"
	case LayerReflect:
		return "WINDIVERT_LAYER_REFLECT"
	//case LayerEthernet:
	//	return "WINDIVERT_LAYER_ETHERNET"
	default:
		return ""
	}
}

type Event int

func (e Event) String() string {
	switch e {
	case EventNetworkPacket:
		return "WINDIVERT_EVENT_NETWORK_PACKET"
	case EventFlowEstablished:
		return "WINDIVERT_EVENT_FLOW_ESTABLISHED"
	case EventFlowDeleted:
		return "WINDIVERT_EVENT_FLOW_DELETED"
	case EventSocketBind:
		return "WINDIVERT_EVENT_SOCKET_BIND"
	case EventSocketConnect:
		return "WINDIVERT_EVENT_SOCKET_CONNECT"
	case EventSocketListen:
		return "WINDIVERT_EVENT_SOCKET_LISTEN"
	case EventSocketAccept:
		return "WINDIVERT_EVENT_SOCKET_ACCEPT"
	case EventSocketClose:
		return "WINDIVERT_EVENT_SOCKET_CLOSE"
	case EventReflectOpen:
		return "WINDIVERT_EVENT_REFLECT_OPEN"
	case EventReflectClose:
		return "WINDIVERT_EVENT_REFLECT_CLOSE"
	//case EventEthernetFrame:
	//	return "WINDIVERT_EVENT_ETHERNET_FRAME"
	default:
		return ""
	}
}

type Shutdown int

func (s Shutdown) String() string {
	switch s {
	case ShutdownRecv:
		return "WINDIVERT_SHUTDOWN_RECV"
	case ShutdownSend:
		return "WINDIVERT_SHUTDOWN_SEND"
	case ShutdownBoth:
		return "WINDIVERT_SHUTDOWN_BOTH"
	default:
		return ""
	}
}

type Param int

func (p Param) String() string {
	switch p {
	case QueueLength:
		return "WINDIVERT_PARAM_QUEUE_LENGTH"
	case QueueTime:
		return "WINDIVERT_PARAM_QUEUE_TIME"
	case QueueSize:
		return "WINDIVERT_PARAM_QUEUE_SIZE"
	case VersionMajor:
		return "WINDIVERT_PARAM_VERSION_MAJOR"
	case VersionMinor:
		return "WINDIVERT_PARAM_VERSION_MINOR"
	default:
		return ""
	}
}
