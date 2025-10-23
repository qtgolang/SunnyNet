//go:build darwin
// +build darwin

package tunPublic

import (
	"bufio"   // 用于按行读取命令输出
	"fmt"     // 格式化输出
	"net"     // 获取本机网卡地址
	"os/exec" // 执行系统命令
	"strings" // 处理字符串
)

// findInterfaceByIP 根据给定的 IPv4 字符串查找包含该地址的接口名
func findInterfaceByIP(ipStr string) (string, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil || ip.To4() == nil {
		return "", fmt.Errorf("无效的 IPv4 地址: %s", ipStr)
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, ifi := range ifaces {
		addrs, err := ifi.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			var curIP net.IP
			switch v := a.(type) {
			case *net.IPNet:
				curIP = v.IP
			case *net.IPAddr:
				curIP = v.IP
			}
			if curIP != nil && curIP.Equal(ip) {
				return ifi.Name, nil
			}
		}
	}
	return "", fmt.Errorf("未找到 IP %s 对应的网卡", ipStr)
}

// parseNetstatDefault 解析 netstat -rn -f inet 输出，找到 default 路由网关
func parseNetstatDefault(output, ifaceName string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		fields := strings.Fields(strings.TrimSpace(scanner.Text()))
		if len(fields) < 6 {
			continue
		}
		if fields[0] == "default" && fields[len(fields)-1] == ifaceName {
			return fields[1], nil
		}
	}
	return "", fmt.Errorf("在路由表中未找到接口 %s 的 default 路由", ifaceName)
}

// getGatewayByInterface 通过接口名查找 default 网关
func getGatewayByInterface(ifaceName string) (string, error) {
	// 优先使用 netstat
	cmd := exec.Command("netstat", "-rn", "-f", "inet")
	out, err := cmd.Output()
	if err == nil {
		if gw, e := parseNetstatDefault(string(out), ifaceName); e == nil {
			return gw, nil
		}
	}

	// 备用使用 route get
	cmd2 := exec.Command("route", "-n", "get", "default")
	out2, err2 := cmd2.Output()
	if err2 != nil {
		return "", fmt.Errorf("无法获取网关")
	}

	var gw, ifn string
	for _, line := range strings.Split(string(out2), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "gateway:"):
			gw = strings.TrimSpace(strings.TrimPrefix(line, "gateway:"))
		case strings.HasPrefix(line, "interface:"):
			ifn = strings.TrimSpace(strings.TrimPrefix(line, "interface:"))
		}
	}
	if ifn == ifaceName && gw != "" {
		return gw, nil
	}

	return "", fmt.Errorf("未找到接口 %s 对应的网关", ifaceName)
}

// GetGatewayByDefault 返回当前默认出口 IPv4 及其网关地址
func GetGatewayByDefault() (string, string, string) {
	// 建立 UDP 连接获取默认出口 IP
	conn, err := net.Dial("udp", "1.2.3.4:5")
	if err != nil {
		return "", "", ""
	}
	defer conn.Close()

	localIP := conn.LocalAddr().(*net.UDPAddr).IP.To4()
	if localIP == nil {
		return "", "", ""
	}

	ifaceName, err := findInterfaceByIP(localIP.String())
	if err != nil {
		return localIP.String(), "", ifaceName
	}

	gw, err := getGatewayByInterface(ifaceName)
	if err != nil {
		return localIP.String(), "", ifaceName
	}

	return localIP.String(), gw, ifaceName
}
