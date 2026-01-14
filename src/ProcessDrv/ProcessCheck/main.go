package ProcessCheck

import (
	"strings"
	"sync"

	"github.com/qtgolang/SunnyNet/src/public"
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

var BlackName = make(map[string]bool)
var BlackPid = make(map[uint32]bool)

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
	if u == "" {
		return false
	}

	a1 := strings.ToLower(u)
	gbk, _ := public.Utf8ToGbk(u)
	utf8, _ := public.GbkToUtf8(u)
	a2 := strings.ToLower(gbk)
	a3 := strings.ToLower(utf8)

	if a2 == "" || a2 == a1 {
		a2 = ""
	}
	if a3 == "" || a3 == a1 || (a2 != "" && a3 == a2) {
		a3 = ""
	}

	Lock.Lock()
	Name[a1] = true
	if a2 != "" {
		Name[a2] = true
	}
	if a3 != "" {
		Name[a3] = true
	}
	Lock.Unlock()

	// 关闭连接一般不建议在锁内做
	CloseNameTCP(a1)
	if a2 != "" {
		CloseNameTCP(a2)
	}
	if a3 != "" {
		CloseNameTCP(a3)
	}

	return true
}

func DelName(u string) bool {
	if u == "" {
		return false
	}

	a1 := strings.ToLower(u)

	gbk, _ := public.Utf8ToGbk(u)
	utf8, _ := public.GbkToUtf8(u)

	a2 := strings.ToLower(gbk)
	a3 := strings.ToLower(utf8)

	if a2 == "" || a2 == a1 {
		a2 = ""
	}
	if a3 == "" || a3 == a1 || (a2 != "" && a3 == a2) {
		a3 = ""
	}

	Lock.Lock()
	delete(Name, a1)
	if a2 != "" {
		delete(Name, a2)
	}
	if a3 != "" {
		delete(Name, a3)
	}
	Lock.Unlock()

	CloseNameTCP(a1)
	if a2 != "" {
		CloseNameTCP(a2)
	}
	if a3 != "" {
		CloseNameTCP(a3)
	}

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

func AddBlackName(u string) bool {
	Lock.Lock()
	BlackName[strings.ToLower(u)] = true
	Lock.Unlock()
	CloseNameTCP(u)
	return true
}
func DelBlackName(u string) bool {
	Lock.Lock()
	delete(BlackName, strings.ToLower(u))
	Lock.Unlock()
	CloseNameTCP(u)
	return true
}
func AddBlackPid(u uint32) bool {
	Lock.Lock()
	BlackPid[u] = true
	Lock.Unlock()
	ClosePidTCP(int(u))
	return true
}
func DelBlackPid(u uint32) bool {
	Lock.Lock()
	delete(BlackPid, u)
	Lock.Unlock()
	ClosePidTCP(int(u))
	return true
}
func CancelBlackAll() bool {
	Lock.Lock()
	for u := range BlackName {
		CloseNameTCP(u)
		delete(BlackName, u)
	}
	for u := range BlackPid {
		ClosePidTCP(int(u))
		delete(BlackPid, u)
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
