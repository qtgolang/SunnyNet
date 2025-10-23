//go:build linux && !android
// +build linux,!android

package Tun

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func RunCmd(c ...string) string {
	var arr []string
	var name string
	for _, vv := range c {
		a := strings.Split(strings.TrimSpace(vv), " ")
		for _, v := range a {
			if v != "" {
				if name == "" {
					name = v
					continue
				}
				arr = append(arr, v)
			}
		}
	}
	if len(arr) == 0 {
		return ""
	}
	r, _ := exec.Command(name, arr...).CombinedOutput()
	return string(r)
}

// Router 封装路由/iptables/TUN 的上下文
type TunRouter struct {
	tunName      string   // TUN 设备名，例如 tun0
	tunIP        string   // TUN 本地地址，例如 1.2.3.1
	tunGW        string   // TUN 虚拟对端网关，例如 1.2.3.2
	defGWIP      string   // 原默认网关 IP，例如 192.168.96.2
	ifaceName    string   // 原默认网卡名称，例如 ens33
	hostIP       string   // 本机在该网卡上的 IPv4，例如 192.168.96.134
	localCIDR    string   // 本地子网CIDR，例如 192.168.96.0/24
	iptablesRule []string // 记录待删除的 iptables 规则参数
}

// getDefaultRoute 解析默认网关与网卡、以及本机该网卡的 IPv4
func getDefaultRoute() (hostIP, defGWIP, localCIDR string, err error) {
	out := RunCmd("ip", "route", "show", "default")
	re := regexp.MustCompile(`default\s+via\s+(\S+)\s+dev\s+(\S+)`)
	m := re.FindStringSubmatch(out)
	if len(m) < 3 {
		err = fmt.Errorf("无法解析默认路由: %s", out)
		return
	}
	out2 := RunCmd("ip", "-4", "addr", "show", "dev", ifaceName)
	defGWIP, ifaceName = m[1], m[2]
	sc := bufio.NewScanner(strings.NewReader(out2))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "inet ") {
			// 形如：inet 192.168.96.134/24 brd ...
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				ipCidr := fields[1]
				localCIDR = ipCidr
				if i := strings.Index(ipCidr, "/"); i > 0 {
					hostIP = ipCidr[:i]
				}
				break
			}
		}
	}
	if hostIP == "" || localCIDR == "" {
		err = fmt.Errorf("无法解析网卡 %s 的 IPv4 地址: %s", ifaceName, out2)
		return
	}
	localCIDR, err = hostCIDRToNetworkCIDR(localCIDR) // 计算 network/prefix
	if err != nil {                                   // 如果失败
		return // 返回错误
	}
	return
}

// (r *Router) applyRouting 应用策略路由与 iptables 规则
func (r *TunRouter) applyRouting() error {
	// 放宽 rp_filter，避免策略路由下被 RPF 误杀
	RunCmd("sysctl", "-w", "net.ipv4.conf.all.rp_filter=2")                          // all=2
	RunCmd("sysctl", "-w", "net.ipv4.conf.default.rp_filter=2")                      // default=2
	RunCmd("sysctl", "-w", fmt.Sprintf("net.ipv4.conf.%s.rp_filter=2", r.ifaceName)) // 原网卡=2
	RunCmd("sysctl", "-w", fmt.Sprintf("net.ipv4.conf.%s.rp_filter=0", r.tunName))   // TUN=0

	// table 100：旁路（走原网关/原网卡）
	RunCmd("ip", "route", "replace", "default", "via", r.defGWIP, "dev", r.ifaceName, "table", "100") // 默认走原网关
	RunCmd("ip", "route", "replace", r.defGWIP+"/32", "dev", r.ifaceName, "table", "100")             // 网关/32 直连
	RunCmd("ip", "route", "replace", r.localCIDR, "dev", r.ifaceName, "table", "100")                 // 本地网段直连(网络前缀)

	// table 200：TUN（默认走 TUN）
	RunCmd("ip", "route", "replace", "default", "via", r.tunGW, "dev", r.tunName, "table", "200") // 默认走 TUN

	// ip rule（优先级从小到大）
	RunCmd("ip", "rule", "add", "priority", "90", "from", r.hostIP+"/32", "lookup", "100") // 本机源IP回包走原网卡
	RunCmd("ip", "rule", "add", "priority", "100", "fwmark", "1", "lookup", "100")         // 本进程(按UID标记)旁路
	RunCmd("ip", "rule", "add", "priority", "120", "to", r.localCIDR, "lookup", "main")    // 目的为本地网段走main
	RunCmd("ip", "rule", "add", "priority", "220", "lookup", "200")                        // 兜底全部走TUN

	// 刷新路由缓存
	RunCmd("ip", "-4", "route", "flush", "cache") // 清空IPv4路由缓存

	// 仅标记“本进程用户”的输出报文（无需 SO_MARK；使用 owner --uid-owner）
	uidStr := strconv.Itoa(int(_myPid))                                                                                                          // 转字符串
	r.iptablesRule = []string{"iptables", "-t", "mangle", "-A", "OUTPUT", "-m", "owner", "--uid-owner", uidStr, "-j", "MARK", "--set-mark", "1"} // 记录规则
	RunCmd(r.iptablesRule...)                                                                                                                    // 应用规则
	return nil                                                                                                                                   // 返回成功
}

// (r *Router) cleanup 还原策略路由与 iptables，删除 TUN
func (r *TunRouter) cleanup() {
	// 删除 iptables 规则
	if len(r.iptablesRule) > 0 { // 若存在记录
		del := append([]string{}, r.iptablesRule...) // 复制一份
		del[3] = "-D"                                // 将 -A 改为 -D
		RunCmd(del...)                               // 删除规则
	}

	// 删除策略路由（按优先级逐条删除）
	RunCmd("ip", "rule", "del", "priority", "220") // 删兜底
	RunCmd("ip", "rule", "del", "priority", "120") // 删本地网段
	RunCmd("ip", "rule", "del", "priority", "100") // 删 fwmark
	RunCmd("ip", "rule", "del", "priority", "90")  // 删 from hostIP

	// 清理 table200（TUN）与 table100（仅删我们加的三条）
	RunCmd("ip", "route", "flush", "table", "200")                                                // 清空200
	RunCmd("ip", "route", "del", "default", "via", r.defGWIP, "dev", r.ifaceName, "table", "100") // 删默认
	RunCmd("ip", "route", "del", r.defGWIP+"/32", "dev", r.ifaceName, "table", "100")             // 删/32
	RunCmd("ip", "route", "del", r.localCIDR, "dev", r.ifaceName, "table", "100")                 // 删本地网段

	// 还原 rp_filter（可按需保持放宽，这里演示复原）
	RunCmd("sysctl", "-w", "net.ipv4.conf.all.rp_filter=1")                          // all=1
	RunCmd("sysctl", "-w", "net.ipv4.conf.default.rp_filter=1")                      // default=1
	RunCmd("sysctl", "-w", fmt.Sprintf("net.ipv4.conf.%s.rp_filter=1", r.ifaceName)) // 原网卡=1
	RunCmd("sysctl", "-w", fmt.Sprintf("net.ipv4.conf.%s.rp_filter=1", r.tunName))   // TUN=1

	// 删除 TUN（双保险）
	RunCmd("ip", "link", "del", r.tunName) // 删除 TUN 接口
}

var rou *TunRouter

func init() {
	hostIP, defGWIP, localCIDR, _ := getDefaultRoute()
	rou = &TunRouter{
		tunIP:     "1.2.3.1", // 你可按需修改
		tunGW:     "1.2.3.2", // 你可按需修改
		defGWIP:   defGWIP,
		ifaceName: ifaceName,
		hostIP:    hostIP,
		localCIDR: localCIDR,
	}
}

// hostCIDRToNetworkCIDR 将诸如 "192.168.96.134/24" 规范化为 "192.168.96.0/24"
func hostCIDRToNetworkCIDR(hostCIDR string) (string, error) { // 返回网络前缀CIDR与错误
	ip, ipnet, err := net.ParseCIDR(hostCIDR) // 解析 CIDR
	if err != nil {                           // 如果解析失败
		return "", err // 返回错误
	}
	network := ip.Mask(ipnet.Mask)                                // 计算网络地址
	prefixLen, _ := ipnet.Mask.Size()                             // 取掩码长度
	return fmt.Sprintf("%s/%d", network.String(), prefixLen), nil // 组装 network/prefix
}
