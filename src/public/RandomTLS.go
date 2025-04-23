package public

import (
	mrand "math/rand"
	"sync"
)

var RandomTLSValueArray = []uint16{0x0005, 0x000a, 0x002f, 0x0035, 0x003c, 0x009c, 0x009d, 0xc007, 0xc009, 0xc00a, 0xc011, 0xc012, 0xc013, 0xc014, 0xc023, 0xc027, 0xc02f, 0xc02b, 0xc030, 0xc02c, 0xcca8, 0xcca9, 0x1301, 0x1302, 0x1303, 0x5600}
var RandomTLSValueArrayLen = len(RandomTLSValueArray)

var _httpRandomTLSValue []uint16

func init() {
	_httpRandomTLSValue = make([]uint16, RandomTLSValueArrayLen)
	copy(_httpRandomTLSValue, RandomTLSValueArray)
}

var _RandomTLSLock sync.Mutex

func GetTLSValues() []uint16 {
	_RandomTLSLock.Lock()
	defer _RandomTLSLock.Unlock()
	n := mrand.Intn(RandomTLSValueArrayLen) + 1
	for i := RandomTLSValueArrayLen - 1; i > 0; i-- {
		j := mrand.Intn(i + 1)
		_httpRandomTLSValue[i], _httpRandomTLSValue[j] = _httpRandomTLSValue[j], _httpRandomTLSValue[i]
	}
	shuffledArray := make([]uint16, n)
	copy(shuffledArray, _httpRandomTLSValue[:n])
	return shuffledArray
}
