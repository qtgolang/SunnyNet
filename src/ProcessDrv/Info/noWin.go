//go:build !windows
// +build !windows

package Info

import (
	"sync"
)

func GetSystemDirectory() string {
	return ""
}

func Wow64DisableWow64FsRedirection() uintptr {
	return 0
}

func Wow64RevertWow64FsRedirection(oldValue uintptr) bool {
	return false
}

var (
	WindowsDirectory = GetWindowsDirectory()
)

// WindowsX64 当前进程是否64位进程
const WindowsX64 = 4<<(^uintptr(0)>>63) == 8

// Is64Windows 系统是否是 64位 系统
var Is64Windows = IsX64CPU()

func IsX64CPU() bool {
	return false
}
func GetWindowsDirectory() string {
	return ""
}

// Exists 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	return false
}

func WriteFile(path string, data []byte) {
}
func checkFileIsExist(filename string) bool {
	return false
}
func ExecCommand(commandName string, params []string) string {
	return ""
}

type DrvInfo interface {
	GetRemoteAddress() string
	GetRemotePort() uint16
	GetPid() string
	IsV6() bool
	ID() uint64
	Close()
}

var Name = make(map[string]bool)
var Pid = make(map[uint32]bool)
var Proxy = make(map[uint16]DrvInfo)
var Lock sync.Mutex

var HookProcess bool

func HookAllProcess(open, StopNetwork bool) {

}

func GetTcpConnectInfo(u uint16) DrvInfo {
	return nil
}
func DelTcpConnectInfo(u uint16) {

}
func AddName(u string) bool {

	return true
}
func DelName(u string) bool {

	return true
}
func AddPid(u uint32) bool {

	return true
}
func DelPid(u uint32) bool {

	return true
}

func CancelAll() bool {

	return true
}
