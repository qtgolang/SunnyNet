package Api

import (
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/nfapi"
)

func SetUdpData(MessageId int, data []byte) bool {
	NFapi.UdpSync.Lock()
	buff := NFapi.UdpMap[MessageId]
	if buff != nil {
		buff.Reset()
		buff.Write(data)
		NFapi.UdpSync.Unlock()
		return true
	}
	NFapi.UdpSync.Unlock()
	return false
}
func GetUdpData(MessageId int) []byte {
	NFapi.UdpSync.Lock()
	buff := NFapi.UdpMap[MessageId]
	if buff != nil {
		NFapi.UdpSync.Unlock()
		return buff.Bytes()
	}
	NFapi.UdpSync.Unlock()
	return nil
}

func UdpSendToServer(tid int, data []byte) bool {
	return NFapi.UdpSendToServer(int64(tid), data)
}
func UdpSendToClient(tid int, data []byte) bool {
	return NFapi.UdpSendToClient(int64(tid), data)
}
