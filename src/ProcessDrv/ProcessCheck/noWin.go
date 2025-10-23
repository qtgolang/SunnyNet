//go:build !windows
// +build !windows

package ProcessCheck

var ClosePidTCP = func(PID int) {
}
var CloseNameTCP = func(processName string) {
}
