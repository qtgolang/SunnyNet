package Api

import "sync"

// 储存管理 MessageId
//  ---------------------------------------------

var MessageIdLock sync.Mutex
var messageId = 1000

//创建新的 messageId
func newMessageId() int {
	MessageIdLock.Lock()
	defer MessageIdLock.Unlock()
	messageId++
	t := messageId
	if t < 0 || t > 2147483640 {
		t = 9999
		messageId = 1000
	}
	return t
}
