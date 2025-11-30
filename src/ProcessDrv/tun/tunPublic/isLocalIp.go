package tunPublic

import (
	"net"
)

func IsLocalIp(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		//MacOS utun 网关
		if ip4[0] == 1 && ip4[1] == 2 && ip4[2] == 3 && ip4[3] == 1 {
			return true
		}
		//Android Vpn 网关
		if ip4[0] == 10 && ip4[1] == 0 && ip4[2] == 0 && ip4[3] == 2 {
			return true
		}
	}
	return false
}
