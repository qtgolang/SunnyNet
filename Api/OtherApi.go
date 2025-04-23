package Api

import (
	"bytes"
	"encoding/binary"
	"github.com/qtgolang/SunnyNet/src/public"
)
 
// BytesToInt 将Go int的Bytes 转为int
func BytesToInt(data uintptr, dataLen int) int {
	bys := public.CStringToBytes(data, dataLen)
	buff := bytes.NewBuffer(bys)
	var B int64
	_ = binary.Read(buff, binary.BigEndian, &B)
	return int(B)
}
