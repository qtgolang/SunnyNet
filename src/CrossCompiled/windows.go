//go:build windows
// +build windows

package CrossCompiled

import "C"
import (
	"bufio"
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/Trisia/gosysproxy"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/Info"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/Proxifier"
	NFapi2 "github.com/qtgolang/SunnyNet/src/ProcessDrv/nfapi"
	"github.com/qtgolang/SunnyNet/src/iphlpapi"
	"github.com/qtgolang/SunnyNet/src/public"
	"golang.org/x/sys/windows"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

func NFapi_SunnyPointer(a ...uintptr) uintptr {
	if len(a) > 0 {
		NFapi2.SunnyPointer = a[0]
	}
	return NFapi2.SunnyPointer
}
func NFapi_IsInit(a ...bool) bool {
	if len(a) > 0 {
		NFapi2.IsInit = a[0]
	}
	return NFapi2.IsInit
}
func Pr_Install() bool {
	return Proxifier.Install()
}
func Pr_IsInit() bool {
	return Proxifier.IsInit()
}

func Pr_SetHandle(Handle func(conn net.Conn)) bool {
	return Proxifier.SetHandle(Handle)
}
func NFapi_ProcessPortInt(a ...uint16) uint16 {
	if len(a) > 0 {
		NFapi2.ProcessPortInt = a[0]
	}
	return NFapi2.ProcessPortInt
}
func NFapi_ApiInit() bool {
	return NFapi2.ApiInit()
}
func NFapi_MessageBox(caption, text string, style uintptr) (result int) {
	return NFapi2.MessageBox(caption, text, style)
}
func Drive_UnInstall() {
	tmp := NFapi2.System32Dir + "\\tmp.tmp"
	if err := os.WriteFile(tmp, []byte("check"), 0777); err != nil {
		return
	}
	_ = os.Remove(tmp)
	NFapi2.UnInstall()
	Proxifier.UnInstall()
	Proxifier.Run("shutdown", "/r", "/f", "/t", "0")
	time.Sleep(2 * time.Second)
}
func NFapi_HookAllProcess(open, StopNetwork bool) {
	Info.HookAllProcess(open, StopNetwork)
}
func NFapi_ClosePidTCP(pid int) {
	Info.ClosePidTCP(pid)
}
func NFapi_DelName(u string) {
	a, e := public.GbkToUtf8(u)
	if e != nil {
		Info.AddName(a)
	}
	a, e = public.Utf8ToGbk(u)
	if e != nil {
		Info.AddName(a)
	}
	Info.DelName(u)
}
func NFapi_AddName(u string) {
	a, e := public.GbkToUtf8(u)
	if e != nil {
		Info.AddName(a)
	}
	a, e = public.Utf8ToGbk(u)
	if e != nil {
		Info.AddName(a)
	}
	Info.AddName(u)
}
func NFapi_DelPid(pid uint32) {
	Info.DelPid(pid)
}
func NFapi_AddPid(pid uint32) {
	Info.AddPid(pid)
}

func NFapi_CancelAll() {
	Info.CancelAll()
}
func NFapi_DelTcpConnectInfo(U uint16) {
	Info.DelTcpConnectInfo(U)
}
func NFapi_GetTcpConnectInfo(U uint16) Info.DrvInfo {
	return Info.GetTcpConnectInfo(U)
}

func NFapi_UdpSendReceiveFunc(udp func(Type int, Theoni int64, pid uint32, LocalAddress, RemoteAddress string, data []byte) []byte) func(Type int, Theoni int64, pid uint32, LocalAddress, RemoteAddress string, data []byte) []byte {
	NFapi2.UdpSendReceiveFunc = udp
	return NFapi2.UdpSendReceiveFunc
}
func NFapi_Api_NfUdpPostSend(id uint64, remoteAddress *NFapi2.SockaddrInx, buf []byte, option *NFapi2.NF_UDP_OPTIONS) (NFapi2.NF_STATUS, error) {
	return NFapi2.Api.NfUdpPostSend(id, remoteAddress, buf, option)
}

func SetIeProxy(Off bool, Port int) bool {
	// "github.com/Tri sia/gos ysp roxy"
	if Off {
		_ = gosysproxy.Off()
		return true
	}
	ies := "127.0.0.1:" + strconv.Itoa(Port)
	_ = gosysproxy.SetGlobalProxy("http="+ies+";https="+ies, "")
	return true
}

// InstallCert 安装证书 将证书安装到Windows系统内
func InstallCert(certificates []byte) (res string) {
	defer func() {
		CertificateName := public.GetCertificateName(certificates)
		if CertificateName != "" && isInstallSunnyNetCertificates(CertificateName) {
			res = "already in store"
		}
	}()
	tempDir := os.TempDir()
	err := public.WriteBytesToFile(certificates, tempDir+"\\SunnyNet.crt")
	if err != nil {
		return err.Error()
	}
	var args []string
	args = append(args, "-addstore")
	args = append(args, "root")
	args = append(args, tempDir+"\\SunnyNet.crt")
	defer func() { _ = public.RemoveFile(tempDir + "\\SunnyNet.crt") }()
	cmd := exec.Command("certutil", args...)
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

// InstallCert2 有感安装,会提示对话框安装
func InstallCert2(certPEM []byte) string {
	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return "Invalid certificate"
	}
	storeName, _ := syscall.UTF16PtrFromString("ROOT")
	store, err := windows.CertOpenStore(windows.CERT_STORE_PROV_SYSTEM, 0, 0, windows.CERT_SYSTEM_STORE_CURRENT_USER, uintptr(unsafe.Pointer(storeName)))
	if err != nil {
		return fmt.Sprintf("failed to open certificate store: %v", err)
	}
	defer windows.CertCloseStore(store, 0)
	certContext, err := windows.CertCreateCertificateContext(
		windows.X509_ASN_ENCODING|windows.PKCS_7_ASN_ENCODING,
		&block.Bytes[0],
		uint32(len(block.Bytes)),
	)
	if err != nil {
		return fmt.Sprintf("failed to create certificate context: %v", err)
	}
	defer windows.CertFreeCertificateContext(certContext)
	// 将证书添加到存储区
	if windows.CertAddCertificateContextToStore(
		store,
		certContext,
		windows.CERT_STORE_ADD_USE_EXISTING,
		nil,
	) != nil {
		return "安装证书失败：用户未授权安装证书"
	}
	return "already in store"
}

const (
	CERT_SYSTEM_STORE_CURRENT_USER  = uint32(1 << 16) // 当前用户证书存储
	CERT_SYSTEM_STORE_LOCAL_MACHINE = uint32(2 << 16) // 本地计算机证书存储
)

// 检查是否安装了包含 "SunnyNet" 的证书
func isInstallSunnyNetCertificates(CertificateName string) bool {
	if !_isInstallSunnyNetCertificates(CERT_SYSTEM_STORE_CURRENT_USER, CertificateName) {
		return _isInstallSunnyNetCertificates(CERT_SYSTEM_STORE_LOCAL_MACHINE, CertificateName)
	}
	return true
}

func _isInstallSunnyNetCertificates(CERT uint32, CertificateName string) bool {
	// 将 "ROOT" 转换为 UTF-16 指针
	storeName, err := syscall.UTF16PtrFromString("ROOT")
	if err != nil {
		return false // 转换失败，返回 false
	}

	// 打开当前用户的根证书存储
	store, err := windows.CertOpenStore(windows.CERT_STORE_PROV_SYSTEM, 0, 0, CERT, uintptr(unsafe.Pointer(storeName)))
	if store == 0 || err != nil {
		return false // 打开证书存储失败，返回 false
	}
	defer windows.CertCloseStore(store, 0) // 确保在函数结束时关闭证书存储

	var cert *windows.CertContext // 声明证书上下文
	for {
		// 枚举证书存储中的证书
		cert, _ = windows.CertEnumCertificatesInStore(store, cert)
		if cert == nil {
			break // 如果没有更多证书，退出循环
		}
		// 获取证书的字节数据
		certBytes := (*[1 << 20]byte)(unsafe.Pointer(cert.EncodedCert))[:cert.Length:cert.Length]
		// 解析证书
		parsedCert, er := x509.ParseCertificate(certBytes)
		if er != nil {
			continue // 如果解析失败，继续下一个证书
		}
		// 检查证书的主题名称是否包含 "CertificateName"
		if strings.Contains(parsedCert.Subject.CommonName, CertificateName) {
			return true // 找到匹配的证书，返回 true
		}
	}

	return false // 未找到匹配的证书，返回 false
}
func SetNetworkConnectNumber() {
	//https://blog.csdn.net/PYJcsdn/article/details/126251054
	//尽量避免这个问题
	var args []string
	args = append(args, "int")
	args = append(args, "ipv4")
	args = append(args, "set")
	args = append(args, "dynamicport")
	args = append(args, "tcp")
	args = append(args, "start=10000")
	args = append(args, "num=55000")
	Info.ExecCommand("netsh", args)
	var args1 []string
	args1 = append(args1, "int")
	args1 = append(args1, "ipv6")
	args1 = append(args1, "set")
	args1 = append(args1, "dynamicport")
	args1 = append(args1, "tcp")
	args1 = append(args1, "start=10000")
	args1 = append(args1, "num=55000")
	Info.ExecCommand("netsh", args1)
}

// CloseCurrentSocket  关闭指定进程的所有TCP连接
func CloseCurrentSocket(PID int, ulAf uint) {
	iphlpapi.CloseCurrentSocket(PID, ulAf)
}

// 添加 Windows 防火墙规则
func AddFirewallRule() {
	executablePath, _ := os.Executable()
	// 删除现有规则
	cmd := exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name=SunnyNet")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} // 隐藏窗口
	_ = cmd.Run()

	// 添加入站规则
	cmd = exec.Command("netsh", "advfirewall", "firewall", "add", "rule", "name=SunnyNet", "dir=in", "action=allow", "program="+executablePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} // 隐藏窗口
	_ = cmd.Run()

	// 添加出站规则
	cmd = exec.Command("netsh", "advfirewall", "firewall", "add", "rule", "name=SunnyNetOut", "dir=out", "action=allow", "program="+executablePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} // 隐藏窗口
	_ = cmd.Run()
}
