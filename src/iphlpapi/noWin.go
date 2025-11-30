//go:build !windows
// +build !windows

package iphlpapi

import "C"

func CloseCurrentSocket(PID int, ulAf uint) {

}
func IsPortListening(port int) bool {
	return false
}
