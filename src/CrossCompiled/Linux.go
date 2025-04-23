//go:build !windows && !darwin
// +build !windows,!darwin

package CrossCompiled

import "github.com/qtgolang/SunnyNet/src/ProcessDrv/Info"

func SetIeProxy(Off bool, Port int) bool {
	return false
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
func Pr_Install() bool {
	return false
}
func Pr_SetHandle(Handle any) bool {
	return false
}
func Drive_UnInstall() {
}
func Pr_IsInit() bool {
	return false
}
func NFapi_UdpSendReceiveFunc(udp func(Type int, Theoni int64, pid uint32, LocalAddress, RemoteAddress string, data []byte) []byte) func(Type int, Theoni int64, pid uint32, LocalAddress, RemoteAddress string, data []byte) []byte {
	return nil
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
