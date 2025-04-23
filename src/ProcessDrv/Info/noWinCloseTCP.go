//go:build !windows
// +build !windows

package Info

const (
	AF_INET  = 2
	AF_INET6 = 23
)

// ClosePidTCP 关闭指定进程的所有TCP连接
func ClosePidTCP(PID int) {

}

// CloseNameTCP 关闭指定进程的所有TCP连接
func CloseNameTCP(processName string) {

}

// GetPIDByName 根据进程名获取进程 PID
func GetPIDByName(processName string) []int {
	return nil
}
