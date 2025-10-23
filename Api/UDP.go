package Api

import (
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/SunnyNetUDP"
)

func SetUdpData(MessageId int, data []byte) bool {
	return SunnyNetUDP.SetMessage(MessageId, data)
}
func GetUdpData(MessageId int) []byte {
	return SunnyNetUDP.GetMessage(MessageId)
}

func UdpSendToServer(tid int, data []byte) bool {
	obj := SunnyNetUDP.GetUDPItem(int64(tid))
	if obj != nil {
		return obj.ToServer(data)
	}
	return false
}
func UdpSendToClient(tid int, data []byte) bool {
	obj := SunnyNetUDP.GetUDPItem(int64(tid))
	if obj != nil {
		return obj.ToClient(data)
	}
	return false
}
