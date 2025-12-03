//go:build !android
// +build !android

package JavaJni

import "net"

func LogError(msg string) {
}
func GetWifiAddr() []net.IP {

	return nil
}
