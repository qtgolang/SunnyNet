package Proxifier

import (
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/Info"
	"github.com/qtgolang/SunnyNet/src/Resource"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

func Run(name string, arg ...string) int {
	cmd := exec.Command(name, arg...)
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	_ = cmd.Run()
	return cmd.ProcessState.ExitCode()
}
func UnInstall() {
	if Info.Is64Windows {
		var oldValue uintptr
		if !Info.WindowsX64 {
			oldValue = Info.Wow64DisableWow64FsRedirection()
		}
		unInstall64()
		unInstall32()
		if !Info.WindowsX64 {
			Info.Wow64RevertWow64FsRedirection(oldValue)
		}
	}
	return
}

func unInstall64() {
	BasePath := Info.WindowsDirectory + "System32\\"
	installFile := Info.WindowsDirectory + "installPrxer64.exe"
	baseUnInstall(BasePath, installFile, false)
}
func baseUnInstall(BasePath, installFile string, x86 bool) {
	lsp := BasePath + "PrxerDrv.dll"
	nsp := BasePath + "PrxerNsp.dll"
	if x86 {
		Info.WriteFile(installFile, Resource.X32InstallLSP)
	} else {
		Info.WriteFile(installFile, Resource.X64InstallLSP)
	}
	Resource.SetAdminRun(installFile)
	defer func() {
		os.Remove(installFile)
	}()
	Run(installFile, "un")
	//可能删除了不,可以有文件正在使用未卸载,那就移动到临时目录,等系统重启后，从临时目录删除
	_ = os.Remove(lsp)
	_ = os.Remove(nsp)
	_ = Info.MoveFileToTempDir(lsp, "Sunny_Prxer_"+Info.RandomLetters(32)+extensionsTemp)
	_ = Info.MoveFileToTempDir(nsp, "Sunny_Prxer_"+Info.RandomLetters(32)+extensionsTemp)
}

var extensionsTemp = ".tmpSys"

func unInstall32() {
	BasePath := Info.WindowsDirectory + "SysWOW64\\"
	/*
		//32位系统无法使用,不知道什么原因
		if !Info.Is64Windows {
			BasePath = Info.WindowsDirectory + "System32\\"
		}
	*/
	installFile := Info.WindowsDirectory + "installPrxer32.exe"
	baseUnInstall(BasePath, installFile, true)
}

func Install() bool {
	if Info.Is64Windows {
		var oldValue uintptr
		if !Info.WindowsX64 {
			oldValue = Info.Wow64DisableWow64FsRedirection()
		}
		a, b := Install64()
		c, d := Install32()
		if !Info.WindowsX64 {
			Info.Wow64RevertWow64FsRedirection(oldValue)
		}
		return a == true && b == true && c == true && d == true
	}
	//32位系统无法使用,不知道什么原因
	return false

	/*
		a, b := Install32()
		return a == true && b == true
	*/
}
func Install64() (bool, bool) {
	BasePath := Info.WindowsDirectory + "System32\\"
	installFile := Info.WindowsDirectory + "installPrxer64.exe"
	return baseInstall(BasePath, installFile, false)
}
func Install32() (bool, bool) {
	BasePath := Info.WindowsDirectory + "SysWOW64\\"
	/*
		//32位系统无法使用,不知道什么原因
		if !Info.Is64Windows {
			BasePath = Info.WindowsDirectory + "System32\\"
		}
	*/
	installFile := Info.WindowsDirectory + "installPrxer32.exe"
	return baseInstall(BasePath, installFile, true)
}
func baseInstall(BasePath, installFile string, x86 bool) (bool, bool) {
	lsp := BasePath + "PrxerDrv.dll"
	nsp := BasePath + "PrxerNsp.dll"
	if x86 {
		Info.WriteFile(installFile, Resource.X32InstallLSP)
	} else {
		Info.WriteFile(installFile, Resource.X64InstallLSP)
	}
	Resource.SetAdminRun(installFile)
	defer func() {
		os.Remove(installFile)
	}()
	a := installLsp(installFile, lsp, x86)
	b := installNsp(installFile, nsp, x86)
	return a, b
}
func installLsp(installFile, lsp string, x86 bool) bool {
	if Run(installFile, "il") == 1 {
		return true
	}
	if x86 {
		Info.WriteFile(lsp, Resource.X32PrxerDrv)
	} else {
		Info.WriteFile(lsp, Resource.X64PrxerDrv)
	}
	Run(installFile, "l")
	return Run(installFile, "il") == 1
}
func installNsp(installFile, nsp string, x86 bool) bool {
	if Run(installFile, "in") == 1 {
		return true
	}
	if x86 {
		Info.WriteFile(nsp, Resource.X32PrxerNsp)
	} else {
		Info.WriteFile(nsp, Resource.X64PrxerNsp)
	}
	Run(installFile, "n")
	return Run(installFile, "in") == 1
}
