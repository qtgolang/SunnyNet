//go:build darwin
// +build darwin

package Tun

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/qtgolang/SunnyNet/src/ProcessDrv/tun/tunPublic"
	CrossCompiled "github.com/qtgolang/SunnyNet/src/iphlpapi/net"
	"github.com/shirou/gopsutil/process"
	"github.com/songgao/water" // TUN 设备
)

func LogError(string) {}

// 判断是否是 IPv4
func isIPv4(ip net.IP) bool {
	return ip.To4() != nil
}

// 判断是否是 IPv6
func isIPv6(ip net.IP) bool {
	return ip.To16() != nil && ip.To4() == nil
}

// OpenTunDevice 创建并配置 TUN 设备
func OpenTunDevice(name, addr, gw, mask string) (io.ReadWriteCloser, error) {
	tunDev, err := water.New(water.Config{DeviceType: water.TUN})
	if err != nil {
		return nil, err
	}
	name = tunDev.Name()
	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, errors.New("无效的 IP 地址")
	}
	var params string
	if isIPv4(ip) {
		params = fmt.Sprintf("%s inet %s netmask %s %s", name, addr, mask, gw)
	} else if isIPv6(ip) {
		prefixlen, err := strconv.Atoi(mask)
		if err != nil {
			return nil, fmt.Errorf("解析 IPv6 前缀长度失败: %v", err)
		}
		params = fmt.Sprintf("%s inet6 %s/%d", name, addr, prefixlen)
	} else {
		return nil, errors.New("未知 IP 类型")
	}
	out, err := exec.Command("ifconfig", strings.Split(params, " ")...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("配置 IP 失败: %v, 输出: %s", err, string(out))
	}
	if gw != "" && isIPv4(ip) {
		routeOut, routeErr := exec.Command("route", "add", "default", gw).CombinedOutput()
		if routeErr != nil {
			return nil, fmt.Errorf("添加路由失败: %v, 输出: %s", routeErr, string(routeOut))
		}
	}
	return tunDev, nil
}

var (
	defaultGatewayIP, defaultGatewayIf, ifaceName = tunPublic.GetGatewayByDefault() // 获取默认网关IP与网卡
	watchdogStarted                               bool                              // Watchdog 是否已经启动
	watchdogPid                                   int                               // Watchdog 脚本进程 PID
)

// startWatchdog 启动路由恢复监控脚本，防止主进程异常退出，导致系统无网络
func startWatchdog() {
	// 如果 Watchdog 已经启动，则不重复启动
	if watchdogStarted {
		return
	}
	// 标记 Watchdog 已启动
	watchdogStarted = true
	// 将 shell 脚本内容写入 /usr/local/bin/SunnyTunCancel.sh 并赋予执行权限
	_ = os.Remove("/tmp/SunnyTunCancel.log")
	_ = os.Remove("/tmp/SunnyTunCancel.pid")
	_ = os.Remove("/usr/local/bin/SunnyTunCancel.sh")
	_ = os.WriteFile("/usr/local/bin/SunnyTunCancel.sh", []byte(sh1), 0777)
	// 启动一个后台 goroutine 持续监控
	go func() {
		// 获取当前主进程 PID（用于传给监控脚本）
		mainPid := os.Getpid()
		for {
			if watchdogPid > 0 {
				if syscall.Kill(watchdogPid, 0) == nil {
					// 如果进程存在，则等待 1 秒后继续检查
					time.Sleep(time.Second)
					continue
				}
			}
			// 启动监控脚本，脚本内部会监听主进程退出事件，并恢复默认路由
			cmd := exec.Command("/bin/sh", "/usr/local/bin/SunnyTunCancel.sh", fmt.Sprintf("%d", mainPid), defaultGatewayIf, ifaceName)
			// 将脚本输出重定向到黑洞（不输出到控制台）
			var buffer bytes.Buffer
			cmd.Stdout = &buffer
			cmd.Stderr = io.Discard
			// 启动脚本进程
			if err := cmd.Start(); err != nil {
				// 如果启动失败，将 PID 设为 0，下次循环会重试
				watchdogPid = 0
			} else {
				_ = cmd.Wait()
				watchdogPid, _ = strconv.Atoi(strings.TrimSpace(buffer.String()))
			}
			// 每 1 秒检查一次脚本状态
			time.Sleep(time.Second)
		}
	}()
}

var _gw = 10

// OnTunCreated 当 TUN 设备创建时执行
func (n *NewTun) OnTunCreated(_ int) bool {
	// 如果 TUN 已经在运行，则直接返回 true，避免重复启动
	if n.IsRunning {
		return true
	}
	// 先标记为未运行状态
	n.IsRunning = false
	// 如果默认网关、网关 IP 或 Sunny 对象为空，说明环境异常，直接返回
	if defaultGatewayIf == "" || defaultGatewayIP == "" || n.Sunny == nil {
		return n.IsRunning
	}
	_gw++
	if _gw > 200 {
		_gw = 10
	}
	// TUN 虚拟网关 IP 地址（模拟出口）
	gw := fmt.Sprintf("1.2.3.%d", _gw)
	// 创建 TUN 设备，指定本地 IP、网关 IP 和掩码
	tun, err := OpenTunDevice("", "1.2.3.1", gw, "255.255.255.0")
	if err != nil {
		// 创建失败，返回 false
		return n.IsRunning
	}
	// 保存 TUN 设备对象
	n.tun = tun
	// 分配一个 64KB 缓冲区用于读取 TUN 数据
	buf := make([]byte, 65535)
	// 再次赋值 TUN 对象（重复赋值，这里保留原逻辑）
	n.tun = tun
	// 设置 Sunny 的出口 IP 地址为原默认网关 IP
	if !n.Sunny.SetOutRouterIP(defaultGatewayIP) {
		// 如果设置失败，关闭 TUN 设备
		_ = tun.Close()
		return n.IsRunning
	}
	// 标记 TUN 已经成功运行
	n.IsRunning = true
	// 修改系统默认路由为 TUN 网关
	_, _ = exec.Command("sudo", "route", "delete", "default").CombinedOutput()
	_, _ = exec.Command("sudo", "route", "add", "default", gw).CombinedOutput()
	_, _ = exec.Command("sudo", "route", "-n", "add", "-inet", "default", "-ifscope", ifaceName, defaultGatewayIf).CombinedOutput()
	// 启动 watchdog，用于在进程退出时恢复默认路由
	startWatchdog()
	// 启动后台 goroutine，持续读取 TUN 数据包
	go func() {
		// 退出时恢复系统默认路由
		defer func() {
			if defaultGatewayIf != "" {
				_, _ = exec.Command("sudo", "route", "delete", "default").CombinedOutput()
				_, _ = exec.Command("sudo", "route", "-n", "delete", "-inet", "default", "-ifscope", ifaceName).CombinedOutput()
				_, _ = exec.Command("sudo", "route", "add", "default", defaultGatewayIf).CombinedOutput()
			}
			_ = tun.Close()
		}()
		for {
			// 从 TUN 设备读取数据
			nBytes, er := tun.Read(buf)
			// 如果 TUN 已经标记为未运行，退出循环
			if !n.IsRunning {
				return
			}
			// 如果读取出错或长度为 0，则稍等后重试
			if er != nil || nBytes <= 0 {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			// 将读取到的数据拷贝到新切片中
			var packet []byte
			packet = append(packet, buf[:nBytes]...)
			// 异步调用解析函数处理数据包
			go func() {
				n.parsePacket(packet)
			}()
		}
	}()
	// 返回当前运行状态
	return n.IsRunning
}

type expiry struct {
	pid    int32
	name   string
	expiry time.Time
}

var (
	pidCache   = make(map[uint16]expiry)
	pidCacheMu sync.RWMutex
	pidTTL     = 3 * time.Second
	pidExpiry  = make(map[uint16]time.Time)
)

func getPidByPort(kind string, port uint16) (int32, string) {
	pidCacheMu.Lock()
	defer pidCacheMu.Unlock()
	for k, _ := range pidCache {
		if !time.Now().Before(pidExpiry[k]) {
			delete(pidExpiry, k)
			delete(pidCache, k)
		}
	}
	if obj, ok := pidCache[port]; ok && time.Now().Before(pidExpiry[port]) {
		return obj.pid, obj.name
	}
	all, _ := CrossCompiled.Connections(kind)
	for _, conn := range all {
		if conn.Laddr.Port == uint32(port) {
			pid := conn.Pid
			p, _ := process.NewProcess(pid)
			ch := expiry{pid: pid}
			if p != nil {
				ch.name, _ = p.Name()
			}
			pidCache[port] = ch
			pidExpiry[port] = time.Now().Add(pidTTL)
			return ch.pid, ch.name
		}
	}
	return 0, ""
}
