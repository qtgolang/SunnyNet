//go:build windows
// +build windows

package Info

/*
#include <windows.h>
#include <stdlib.h>

char* getSystemDirectory() {
    char* buffer = (char*)malloc(MAX_PATH);
    if (buffer == NULL) {
        return NULL;
    }
    DWORD result = GetSystemDirectoryA(buffer, MAX_PATH);
    if (result == 0) {
        free(buffer);
        return NULL;
    }
    return buffer;
}
BOOL disableWow64FsRedirection(PVOID* oldValue) {
    return Wow64DisableWow64FsRedirection(oldValue);
}

BOOL revertWow64FsRedirection(PVOID oldValue) {
    return Wow64RevertWow64FsRedirection(oldValue);
}
*/
import "C"
import (
	"bufio"
	"fmt"
	"github.com/qtgolang/SunnyNet/src/public"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

func GetSystemDirectory() string {
	buffer := C.getSystemDirectory()
	if buffer == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(buffer))
	return C.GoString(buffer)
}

// Wow64DisableWow64FsRedirection 禁用调用线程的文件系统重定向，默认情况下启用文件系统重定向。此功能对于想要访问本机system32目录的32位应用程序很有用。
func Wow64DisableWow64FsRedirection() uintptr {
	var oldValue C.PVOID
	success := C.disableWow64FsRedirection(&oldValue)
	if success == 0 {
		fmt.Println("禁用文件系统重定向 失败")
		return 0
	}
	return uintptr(oldValue)
}

// Wow64RevertWow64FsRedirection 恢复调用线程的文件系统重定向。
func Wow64RevertWow64FsRedirection(oldValue uintptr) bool {
	success := 0
	if oldValue == 0 {
		var oldValues C.PVOID
		success = int(C.revertWow64FsRedirection(oldValues))
	} else {
		success = int(C.revertWow64FsRedirection(C.PVOID(oldValue)))
	}
	if success == 0 {
		fmt.Println("恢复文件系统重定向 失败")
		return false
	}

	return true
}

var (
	WindowsDirectory = GetWindowsDirectory()
)

// WindowsX64 当前进程是否64位进程
const WindowsX64 = 4<<(^uintptr(0)>>63) == 8

// Is64Windows 系统是否是 64位 系统
var Is64Windows = IsX64CPU()

func IsX64CPU() bool {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	GetSystemWow64DirectoryA := kernel32.NewProc("GetSystemWow64DirectoryA")
	Lstrcpyn := kernel32.NewProc("lstrcpyn")
	lpBuffer := make([]byte, 255)
	p := uintptr(unsafe.Pointer(&lpBuffer[0]))
	r, _, _ := Lstrcpyn.Call(p, p, 0)
	r, _, _ = GetSystemWow64DirectoryA.Call(r, 255)
	return r > 0
}
func GetWindowsDirectory() string {
	winDir := os.Getenv("windir")
	if winDir == "" {
		// 如果 windir 不存在，则获取 SystemRoot 环境变量
		winDir = os.Getenv("SystemRoot")
	}
	if winDir[len(winDir)-1:] != "\\" {
		winDir += "\\"
	}
	return winDir
}

// Exists 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true

}

func WriteFile(path string, data []byte) {
	if checkFileIsExist(path) {
		err := os.Remove(path)
		if err != nil {
			return
		}
	}
	f, err1 := os.Create(path) //创建文件
	if err1 == nil {
		_, err1 = f.Write(data)
		if err1 != nil {

			return
		}
		err1 = f.Close()
		if err1 != nil {

			return
		}
	} else {
		if err1 != nil {
			return
		}
	}
}
func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
func ExecCommand(commandName string, params []string) string {
	cmd := exec.Command(commandName, params...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err.Error()
	}
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	_ = cmd.Start()
	var s []byte
	reader := bufio.NewReader(stdout)
	for {
		line, err2 := reader.ReadBytes('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		s = public.BytesCombine(s, line)
	}
	return string(s)
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
func IsFilterRequests(fileName, addr string) bool {
	if strings.Index(strings.ToLower(fileName), "wechat.exe") != -1 && (strings.Contains(addr, "::1") || strings.Contains(addr, "127.0.0.1")) {
		//如果微信连接到本机的这个请求被拦截,小程序无法打开,目前不清楚原因
		return true
	}
	return false
}
