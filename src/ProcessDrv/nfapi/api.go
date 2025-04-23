//go:build windows
// +build windows

package NFapi

/*
#include <windows.h>
#include <stdio.h>
#include <stdlib.h>
int CGOMessageBox(char* text,char* caption,int style) {
     // 获取所需宽字符缓冲区的大小
    int textLength = MultiByteToWideChar(CP_UTF8, 0, text, -1, NULL, 0);
    int captionLength = MultiByteToWideChar(CP_UTF8, 0, caption, -1, NULL, 0);

    // 分配缓冲区
    wchar_t* wideText = (wchar_t*)malloc(textLength * sizeof(wchar_t));
    wchar_t* wideCaption = (wchar_t*)malloc(captionLength * sizeof(wchar_t));

    // 进行转换
    MultiByteToWideChar(CP_UTF8, 0, text, -1, wideText, textLength);
    MultiByteToWideChar(CP_UTF8, 0, caption, -1, wideCaption, captionLength);

    // 调用 MessageBoxW
    int result = MessageBoxW(NULL, wideText, wideCaption, style);

    // 释放分配的内存
    free(wideText);
    free(wideCaption);

    return result;
}
*/
import "C"
import (
	"fmt"
	. "github.com/qtgolang/SunnyNet/src/ProcessDrv/Info"
	"net"
	"strings"
	"unsafe"
)

var apiLoad bool
var apiNfInit bool

func MessageBox(caption, text string, style uintptr) int {
	a := C.CString(caption)
	b := C.CString(text)
	res := C.CGOMessageBox(b, a, C.int(style))
	C.free(unsafe.Pointer(a))
	C.free(unsafe.Pointer(b))
	return int(res)
}

func ApiInit() bool {
	if apiLoad == false {
		DLLPath := Install()
		//DLLPath := GetWindowsDirectory() + NF_DLLName + "64.dll"
		er := Api.Load(DLLPath)
		if er != nil {
			fmt.Println("LoadDLLPathErr=", er)
			return false
		}
		apiLoad = true
	}
	if apiNfInit == false {
		_, v := Api.NfRegisterDriver(NF_DriverName)
		if v != nil {
			errorText := v.Error()
			errorText = strings.ReplaceAll(errorText, "Windows cannot verify the digital signature for this file. A recent hardware or software change might have installed a file that is signed incorrectly or damaged, or that might be malicious software from an unknown source.", "Windows无法验证此驱动文件的数字签名。\r\n\r\n最近的硬件或软件更改可能安装了签名错误或损坏的文件，或者可能是来自未知来源的恶意软件。")
			errorText = strings.ReplaceAll(errorText, "This sys has been blocked from loading", "此驱动程序已被阻止加载。\r\n\r\n可能使用了和 Windows 位数不配的驱动文件。")
			errorText = strings.ReplaceAll(errorText, "The system cannot find the file specified.", "系统找不到指定的驱动文件。")
			errorText = strings.ReplaceAll(errorText, "The specified service has been marked for deletion.", "指定的服务已标记为删除。")
			fmt.Println("载入驱动失败：", errorText)
			return false
		}
		a, er := Api.NfInit()
		if er != nil {
			fmt.Println("NfInitErr=", er)
			return false
		}
		if a != 0 {
			fmt.Println("NfInitErr=", "可能已经有其他程序加载")
			return false
		}
		//_, _ = MoveFileToTempDir(DriverFile, "Sunny_"+randomLetters(32)+extensionsTemp)

		_, er = AddRule(false, IPPROTO_TCP, 0, D_OUT, 0, 0, AF_INET, "", "", "", "", NF_INDICATE_CONNECT_REQUESTS)  //TCP
		_, er = AddRule(false, IPPROTO_TCP, 0, D_OUT, 0, 0, AF_INET6, "", "", "", "", NF_INDICATE_CONNECT_REQUESTS) //TCP

		_, er = AddRule(false, IPPROTO_UDP, 0, D_OUT, 0, 0, AF_INET, "", "", "", "", NF_FILTER)  //UDP
		_, er = AddRule(false, IPPROTO_UDP, 0, D_OUT, 0, 0, AF_INET6, "", "", "", "", NF_FILTER) //UDP

		_, er = AddRule(false, IPPROTO_UDP, 0, D_IN, 0, 0, AF_INET, "", "", "", "", NF_FILTER)  //UDP
		_, er = AddRule(false, IPPROTO_UDP, 0, D_IN, 0, 0, AF_INET6, "", "", "", "", NF_FILTER) //UDP
		if er != nil {
			return false
		}
		apiNfInit = true
	}
	return true
}

func AddRule(toHead bool, _Protocol, pid int32, _Direction DIRECTION, _LocalPort, _RemotePort, family int16, LocalIp, LocalMask, RemoteIp, RemoteMask string, Flag FILTERING_FLAG) (NF_STATUS, error) {
	r := new(NF_RULE)
	var Protocol INT32
	Protocol.Set(_Protocol)
	r.Protocol = Protocol

	var processId UINT32
	processId.Set(uint32(pid))
	r.ProcessId = processId

	r.Direction = uint8(_Direction)

	var LocalPort UINT16
	LocalPort.Set(uint16(_LocalPort))
	r.LocalPort = LocalPort

	var RemotePort UINT16
	RemotePort.Set(uint16(_RemotePort))
	r.RemotePort = RemotePort

	var ipFamily INT16
	ipFamily.Set(family)
	r.IpFamily = ipFamily

	r.LocalIpAddress.SetIP(true, net.ParseIP(LocalIp))
	r.LocalIpAddressMask.SetIP(true, net.ParseIP(LocalMask))
	r.RemoteIpAddress.SetIP(true, net.ParseIP(RemoteIp))
	r.RemoteIpAddressMask.SetIP(true, net.ParseIP(RemoteMask))

	var FilteringFlag UINT32
	FilteringFlag.Set(uint32(Flag))
	r.FilteringFlag = FilteringFlag
	return Api.NfAddRule(r, toHead)
}
