//go:build darwin
// +build darwin

package CrossCompiled

import (
	"fmt"
	"github.com/qtgolang/SunnyNet/src/ProcessDrv/Info"
	"os"
	"os/exec"
	"strings"
)

func runBatchWithPrivileges_lod(commands []string) (error, string) {
	// 用 && 拼接多个命令（注意引号转义）
	joined := strings.Join(commands, " && ")
	escaped := strings.ReplaceAll(joined, `"`, `\"`)
	script := fmt.Sprintf(`do shell script "%s" with administrator privileges`, escaped)
	out, err := exec.Command("osascript", "-e", script).CombinedOutput()
	s := string(out)
	return err, s
}
func runBatchWithPrivileges(commands []string) (error, string) {
	prompt := "SunnyNetV4-请求操作系统代理"
	// 拼接多条命令
	joined := strings.Join(commands, " && ")

	// 获取用户目录
	homeDir, _ := os.UserHomeDir()
	appSupportDir := homeDir + "/Library/Application Support/SunnyNetProV4"
	os.MkdirAll(appSupportDir, 0755)

	// 写入临时脚本
	bashPath := appSupportDir + "/run_as_admin.sh"
	bashScript := "#!/bin/bash\n" + joined + "\n"

	err := os.WriteFile(bashPath, []byte(bashScript), 0755)
	if err != nil {
		return fmt.Errorf("写入脚本失败: %v", err), ""
	}

	// 构造 AppleScript 脚本（自定义提示语）
	escapedPath := strings.ReplaceAll(bashPath, `"`, `\"`)
	escapedPrompt := strings.ReplaceAll(prompt, `"`, `\"`)
	appleScript := fmt.Sprintf(`do shell script "bash \"%s\"" with prompt "%s" with administrator privileges`, escapedPath, escapedPrompt)

	// 执行 osascript
	out, err := exec.Command("osascript", "-e", appleScript).CombinedOutput()
	return err, string(out)
}

type netInterface struct{}

// 获取所有网络接口名称
func (c *netInterface) getAllInterfaceNames() []string {
	cmd := exec.Command("networksetup", "-listallnetworkservices")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var interfaceNames []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "An asterisk (*) ") || strings.HasPrefix(line, " ") || line == "" {
			continue
		}
		interfaceNames = append(interfaceNames, line)
	}

	return interfaceNames
}

func (c *netInterface) SetProxy(proxyHost string, Port int) bool {
	AllInterfaceName := c.getAllInterfaceNames()
	if len(AllInterfaceName) < 1 {
		return false
	}
	var array []string
	for _, interfaceName := range AllInterfaceName {
		array = append(array, fmt.Sprintf("networksetup -setwebproxy \"%s\" %s %d", interfaceName, proxyHost, Port))
		array = append(array, fmt.Sprintf("networksetup -setsecurewebproxy \"%s\" %s %d", interfaceName, proxyHost, Port))
		array = append(array, fmt.Sprintf("networksetup -setsocksfirewallproxy \"%s\" %s %d", interfaceName, proxyHost, Port))
	}
	err, _ := runBatchWithPrivileges(array)
	return err == nil
}

func (c *netInterface) DisableProxy() bool {
	AllInterfaceName := c.getAllInterfaceNames()
	if len(AllInterfaceName) < 1 {
		return false
	}
	var array []string
	for _, interfaceName := range AllInterfaceName {
		array = append(array, fmt.Sprintf("networksetup -setwebproxystate \"%s\" off", interfaceName))
		array = append(array, fmt.Sprintf("networksetup -setsecurewebproxystate \"%s\" off", interfaceName))
		array = append(array, fmt.Sprintf("networksetup -setsocksfirewallproxystate \"%s\" off", interfaceName))
	}
	err, _ := runBatchWithPrivileges(array)
	return err == nil
}

func SetIeProxy(Off bool, Port int) bool {
	Inter := &netInterface{}
	if Off {
		return Inter.DisableProxy()
	}
	return Inter.SetProxy("127.0.0.1", Port)
}
func Drive_UnInstall() {
}
func NFapi_SunnyPointer(a ...uintptr) uintptr {
	return 0
}
func NFapi_IsInit(a ...bool) bool {
	return false
}
func NFapi_ProcessPortInt(a ...uint16) uint16 {
	return 0
}
func NFapi_ApiInit() bool {
	return false
}
func NFapi_MessageBox(caption, text string, style uintptr) (result int) {
	return 0
}
func NFapi_HookAllProcess(open, StopNetwork bool) {
}
func NFapi_ClosePidTCP(pid int) {
}
func NFapi_DelName(u string) {
}
func NFapi_AddName(u string) {
}
func NFapi_DelPid(pid uint32) {
}
func NFapi_AddPid(pid uint32) {
}
func NFapi_CloseNameTCP(u string) {
}
func NFapi_CancelAll() {
}
func NFapi_DelTcpConnectInfo(U uint16) {
}
func NFapi_GetTcpConnectInfo(U uint16) Info.DrvInfo {
	return nil
}
func NFapi_UdpSendReceiveFunc(udp func(Type int, Theoni int64, pid uint32, LocalAddress, RemoteAddress string, data []byte) []byte) func(Type int, Theoni int64, pid uint32, LocalAddress, RemoteAddress string, data []byte) []byte {
	return nil
}
func Pr_Install() bool {
	return false
}
func Pr_SetHandle(Handle any) bool {
	return false
}
func Pr_IsInit() bool {
	return false
}
func NFapi_Api_NfUdpPostSend(id uint64, remoteAddress any, buf []byte, option any) (int32, error) {
	return 0, nil
}

func SetNetworkConnectNumber() {
}

// CloseCurrentSocket  关闭指定进程的所有TCP连接
func CloseCurrentSocket(PID int, ulAf uint) {
}

// InstallCert 安装证书 将证书安装到Windows系统内
func InstallCert(certificates []byte) string {
	return "no Windows"
}

// 添加 Windows 防火墙规则
func AddFirewallRule() {

}
