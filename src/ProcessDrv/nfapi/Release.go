//go:build windows
// +build windows

package NFapi

import "C"
import (
	_ "embed"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/Info"
	. "github.com/qtgolang/SunnyNet/src/ProcessDrv/Info"
	"github.com/qtgolang/SunnyNet/src/Resource"
	"os"
	"path/filepath"
	"strings"
)

// 删除旧的驱动文件
func deleteOldFiles() {
	OldFileName := System32Dir + "\\drivers\\SunnyFilter.sys"
	//复制到临时目录去系统重启后才可删除
	_ = MoveFileToTempDir(OldFileName, "Sunny_"+RandomLetters(32)+extensionsTemp)
	//删除临时目录下的所有sys 文件
	tempDir := os.TempDir()
	// 搜索所有 .sys 文件
	_ = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 检查文件是否是 .sys 文件
		if !info.IsDir() && filepath.Ext(path) == extensionsTemp {
			_ = os.Remove(path)
		}
		return nil
	})
}
func init() {
	deleteOldFiles()
}

// System32Dir C:\Windows\system32\
var System32Dir = GetSystemDirectory()
var extensionsTemp = ".tmpSys"
var DriverFile = System32Dir + "\\drivers\\" + NF_DriverName + ".sys"

func UnInstall() {
	//复制到临时目录去系统重启后才可删除
	_ = MoveFileToTempDir(DriverFile, "Sunny_"+RandomLetters(32)+extensionsTemp)
	DrDLL := Info.WindowsDirectory + NF_DLLName + "64.dll"
	_ = MoveFileToTempDir(DrDLL, "Sunny_"+RandomLetters(32)+extensionsTemp)
	DrDLL = Info.WindowsDirectory + NF_DLLName + "32.dll"
	_ = MoveFileToTempDir(DrDLL, "Sunny_"+RandomLetters(32)+extensionsTemp)
}
func Install() string {
	deleteOldFiles()
	//XP直接打开不程序，所以就直接忽略
	s := []string{"OS", "Get", "Caption"}
	IsWin7 := strings.Index(Info.ExecCommand("Wmic", s), "Windows 7") != -1
	var oldValue uintptr
	if Info.Is64Windows {
		//如果是32位进程 禁止文件重定向 驱动只能写到 system32 目录
		if !WindowsX64 {
			oldValue = Info.Wow64DisableWow64FsRedirection()
		}
	}
	if !Info.Exists(DriverFile) {
		if IsWin7 {
			if Info.Is64Windows {
				Info.WriteFile(DriverFile, Resource.TdiAmd64Netfilter2)

			} else {
				Info.WriteFile(DriverFile, Resource.TdiI386Netfilter2)
			}
		} else {
			if Info.Is64Windows {
				Info.WriteFile(DriverFile, Resource.WfpAmd64Netfilter2)
			} else {
				Info.WriteFile(DriverFile, Resource.WfpI386Netfilter2)
			}
		}
	}
	if Info.Is64Windows {
		//如果是32位进程 恢复文件重定向
		if !WindowsX64 {
			Info.Wow64RevertWow64FsRedirection(oldValue)
		}
	}

	DrDLL := ""
	if WindowsX64 {
		DrDLL = Info.WindowsDirectory + NF_DLLName + "64.dll"
		Info.WriteFile(DrDLL, Resource.NfapiX64Nfapi)
	} else {
		DrDLL = Info.WindowsDirectory + NF_DLLName + "32.dll"
		Info.WriteFile(DrDLL, Resource.NfapiWin32Nfapi)
	}
	return DrDLL
}
