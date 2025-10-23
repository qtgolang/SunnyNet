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
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func (N NFAPI) UnInstall() bool {
	NFapi2.UnInstall()
	return true
}

func (N NFAPI) Install() bool {
	return NFapi2.ApiInit()
}

func (N NFAPI) IsRun() bool {
	return NFapi2.IsInit
}

func (N NFAPI) SetHandle() bool {
	NFapi2.ProcessPortInt = uint16(N.Sunny.Port())
	return true
}

func (N NFAPI) Run() bool {
	NFapi2.UdpSendReceiveFunc = N.UDP
	NFapi2.IsInit = NFapi2.ApiInit()
	return NFapi2.IsInit
}

func (N NFAPI) Close() bool {
	NFapi2.ProcessPortInt = 0
	NFapi2.IsInit = false
	return true
}

func (N NFAPI) Name() string {
	return "NFAPI"
}

func (p Pr) Install() bool {
	return Proxifier.Install()
}

func (p Pr) IsRun() bool {
	return Proxifier.IsInit()
}

func (p Pr) SetHandle() bool {
	return Proxifier.SetHandle(p.TCP)
}
func (p Pr) Run() bool {
	//安装后自动就启动了
	return true
}

func (p Pr) Close() bool {
	return Proxifier.SetHandle(nil)
}

func (p Pr) Name() string {
	return "Proxifier"
}

func (p Pr) UnInstall() bool {
	Proxifier.UnInstall()
	return true
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
