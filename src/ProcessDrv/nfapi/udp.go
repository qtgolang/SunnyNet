package NFapi

import (
	"sync"
)

type NfOPT struct {
	Id            uint64
	RemoteAddress *SockaddrInx
	options       *NF_UDP_OPTIONS
}
type udpItem struct {
	Receive *NfOPT
	Send    *NfOPT
	Theoni  int64
}

func (u udpItem) ToClient(data []byte) bool {
	if u.Receive != nil {
		r, _ := Api.NfUdpPostReceive(u.Receive.Id, u.Receive.RemoteAddress, data, u.Receive.options)
		return r == 0
	}
	return false
}

func (u udpItem) ToServer(data []byte) bool {
	if u.Send != nil {
		r, _ := Api.NfUdpPostSend(u.Send.Id, u.Send.RemoteAddress, data, u.Send.options)
		return r == 0
	}
	return false
}

var mu sync.Mutex
var list = make(map[uint64]*udpItem)
