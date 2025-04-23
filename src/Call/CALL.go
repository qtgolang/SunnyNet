package Call

/*
	MakeChanNum

初始化几个通知管道
*/
var MakeChanNum = 750

// 限制CALL通知函数的访问 避免耗尽资源导致崩溃，或卡顿
var ch = make(chan bool, MakeChanNum)
