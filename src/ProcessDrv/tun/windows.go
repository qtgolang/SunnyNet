//go:build windows
// +build windows

package tun

import (
	"bufio"
	"bytes"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/Info"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/ProcessCheck"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/tun/Tun"
	divert "github.com/qtgolang/SunnyNet/src/ProcessDrv/tun/WinDivert"
	"github.com/qtgolang/SunnyNet/src/Resource"
	"io"
	"os/exec"
	"strings"
	"syscall"
)

var Divert = divert.NewWinDivert()

func IsRun() bool {
	return Divert.IsRunning()
}

func Install() bool {
	if checkWinDivert() {
		return true
	}
	var oldValue uintptr
	if Info.Is64Windows {
		//如果是32位进程 禁止文件重定向 驱动只能写到 system32 目录
		if !Info.WindowsX64 {
			oldValue = Info.Wow64DisableWow64FsRedirection()
		}
	}
	if Info.Is64Windows {
		Info.WriteFile(driver64File, Resource.WinDivert64)
		registerWinDivert(driver64File)
	} else {
		Info.WriteFile(driver32File, Resource.WinDivert32)
		registerWinDivert(driver32File)
	}
	if Info.Is64Windows {
		//如果是32位进程 恢复文件重定向
		if !Info.WindowsX64 {
			Info.Wow64RevertWow64FsRedirection(oldValue)
		}
	}
	return checkWinDivert()
}

func SetHandle(Handle Tun.TcpFunc, udpSendReceiveFunc Tun.UdpFunc, sunny Tun.Interface) bool {
	Divert.SetHandle(Handle, ProcessCheck.CheckPidByName, udpSendReceiveFunc)
	return true
}
func Run() bool {
	return Divert.Run()
}
func Close() bool {
	Divert.Close()
	return true
}
func Name() string {
	return "winDivert"
}

func UnInstall() bool {
	runCmd("sc", "stop", "WinDivert")
	runCmd("sc", "delete", "WinDivert")
	s := "Sunny_" + Info.RandomLetters(32) + extensionsTemp
	_ = Info.MoveFileToTempDir(driver32File, s)
	s = "Sunny_" + Info.RandomLetters(32) + extensionsTemp
	_ = Info.MoveFileToTempDir(driver64File, s)
	return true
}

// 安卓接口
func AndroidTunCreated(fd int) {}

var base = Info.GetSystemDirectory() + "\\drivers\\"
var driver64File = base + "WinDivert64.sys"
var driver32File = base + "WinDivert32.sys"
var extensionsTemp = ".tmpSys"

func registerWinDivert(path string) {
	runCmd("sc", "create", "WinDivert", "type=kernel", "start=demand", `binPath=`+path)
}

// checkServer 安装证书 将证书安装到Windows系统内
func checkWinDivert() bool {
	return strings.Contains(strings.ReplaceAll(runCmd("sc", "query", "WinDivert"), " ", ""), "SERVICE_NAME:WinDivert")
}

func runCmd(Command string, args ...string) (res string) {
	cmd := exec.Command(Command, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err.Error()
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Start()
	var Buff bytes.Buffer
	reader := bufio.NewReader(stdout)
	for {
		line, err2 := reader.ReadBytes('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		Buff.Write(line)
	}
	return Buff.String()
}

func SetFd(fd int) bool {
	return true
}
