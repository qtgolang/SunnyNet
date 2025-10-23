package ProcessCheck

import (
	"strings"
	"sync"
)

type DrvInfo interface {
	GetRemoteAddress() string
	GetRemotePort() uint16
	GetPid() string
	IsV6() bool
	ID() uint64
	Close() error
}

var Name = make(map[string]bool)
var Pid = make(map[uint32]bool)
var Proxy = make(map[uint16]DrvInfo)
var Lock sync.Mutex

var HookProcess bool

func HookAllProcess(open, StopNetwork bool) {
	Lock.Lock()
	for u := range Name {
		delete(Name, u)
	}
	for u := range Pid {
		delete(Pid, u)
	}
	HookProcess = open
	Lock.Unlock()
	if StopNetwork {
		ClosePidTCP(-1)
	}
}
func GetTcpConnectInfo(u uint16) DrvInfo {
	Lock.Lock()
	k := Proxy[u]
	Lock.Unlock()
	if k == nil {
		return nil
	}
	return k
}
func DelTcpConnectInfo(u uint16) {
	Lock.Lock()
	delete(Proxy, u)
	Lock.Unlock()
}
func AddName(u string) bool {
	Lock.Lock()
	Name[strings.ToLower(u)] = true
	Lock.Unlock()
	CloseNameTCP(u)
	return true
}
func DelName(u string) bool {
	Lock.Lock()
	delete(Name, strings.ToLower(u))
	Lock.Unlock()
	CloseNameTCP(u)
	return true
}
func AddPid(u uint32) bool {
	Lock.Lock()
	Pid[u] = true
	Lock.Unlock()
	ClosePidTCP(int(u))
	return true
}
func DelPid(u uint32) bool {
	Lock.Lock()
	delete(Pid, u)
	Lock.Unlock()
	ClosePidTCP(int(u))
	return true
}
func CancelAll() bool {
	Lock.Lock()
	for u := range Name {
		CloseNameTCP(u)
		delete(Name, u)
	}
	for u := range Pid {
		ClosePidTCP(int(u))
		delete(Pid, u)
	}
	Lock.Unlock()
	return true
}

func AddDevObj(connPort uint16, info DrvInfo) {
	Lock.Lock()
	Proxy[connPort] = info
	Lock.Unlock()
}
func DelDevObj(connPort uint16) {
	Lock.Lock()
	delete(Proxy, connPort)
	Lock.Unlock()
}

// CheckPidByName 返回 true 表示不要继续,不在规则中
func CheckPidByName(pid int32, name string) bool {
	Lock.Lock()
	defer Lock.Unlock()
	if HookProcess {
		return false
	}
	if Name[strings.ToLower(name)] == false {
		if Pid[uint32(pid)] == false {
			return true
		}
	}
	return false
}
