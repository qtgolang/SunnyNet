//go:build linux
// +build linux

package tunPublic

import ( // 导入包

	"net" // 获取本机网卡地址
	"os/exec"
	"strings"
)

func GetGatewayByDefault() (string, string, string) {
	out, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		return "", "", ""
	}

	// 示例输出: "default via 192.168.31.1 dev eth0 proto dhcp metric 100"
	fields := strings.Fields(string(out))
	if len(fields) < 5 {
		return "", "", ""
	}

	gateway := fields[2]
	iface := fields[4]

	// 获取本地 IP
	ip := getInterfaceIPv4(iface)
	return ip, gateway, iface
}

func getInterfaceIPv4(ifaceName string) string {
	ifi, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return ""
	}
	addrs, _ := ifi.Addrs()
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return ""
}
