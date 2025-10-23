package SunnyNetUDP

import (
	"sync"
)

type SunnyNetUDP interface {
	ToClient([]byte) bool
	ToServer([]byte) bool
}

var mu sync.Mutex
var list = make(map[int64]SunnyNetUDP)

func AddUDPItem(id int64, Item SunnyNetUDP) {
	mu.Lock()
	list[id] = Item
	mu.Unlock()
}

func DelUDPItem(id int64) {
	mu.Lock()
	delete(list, id)
	mu.Unlock()
}

func GetUDPItem(id int64) SunnyNetUDP {
	mu.Lock()
	obj := list[id]
	mu.Unlock()
	return obj
}

var messageMu sync.Mutex

var message = make(map[int][]byte)

func SetMessage(MessageId int, buff []byte) bool {
	messageMu.Lock()
	defer messageMu.Unlock()
	if _, ok := message[MessageId]; ok {
		message[MessageId] = buff
		return true
	}
	return false
}
func ResetMessage(MessageId int, buff []byte) bool {
	messageMu.Lock()
	defer messageMu.Unlock()
	message[MessageId] = buff
	return true
}
func GetMessage(MessageId int) []byte {
	messageMu.Lock()
	defer messageMu.Unlock()
	if _, ok := message[MessageId]; ok {
		return message[MessageId]
	}
	return nil
}
func DelMessage(MessageId int) {
	messageMu.Lock()
	defer messageMu.Unlock()
	delete(message, MessageId)
}
