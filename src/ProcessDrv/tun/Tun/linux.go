//go:build linux && !android
// +build linux,!android

package Tun

import (
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/qtgolang/SunnyNet/src/ProcessDrv/tun/tunPublic"
	CrossCompiled "github.com/qtgolang/SunnyNet/src/iphlpapi/net"
	"github.com/shirou/gopsutil/process"
	"github.com/songgao/water" // TUN 设备
)

var defaultGatewayIP, defaultGatewayIf, ifaceName = tunPublic.GetGatewayByDefault()

// OpenTunDevice 创建并配置 TUN 设备
func OpenTunDevice(addr string) (io.ReadWriteCloser, string, error) {
	// 创建 TUN 设备
	tunDev, err := water.New(water.Config{DeviceType: water.TUN})
	if err != nil {
		return nil, "", err
	}
	name := tunDev.Name()
	// 配置 TUN 本地 IP 地址
	cmd := exec.Command("ip", "addr", "add", addr+"/24", "dev", name)
	if out, er := cmd.CombinedOutput(); er != nil {
		fmt.Println("❌ 配置 IP 失败:", string(out), er)
		_ = tunDev.Close()
		return nil, name, er
	}

	// 启用 TUN 设备
	cmd = exec.Command("ip", "link", "set", "dev", name, "up")
	if out, er := cmd.CombinedOutput(); er != nil {
		fmt.Println("❌ 启用接口失败:", string(out), er)
		_ = tunDev.Close()
		return nil, name, er
	}

	return tunDev, name, nil
}

// OnTunCreated 当 TUN 设备创建时执行
func (n *NewTun) OnTunCreated(_ int) bool {
	if n.IsRunning {
		return true
	}
	if defaultGatewayIP == "" || defaultGatewayIf == "" || ifaceName == "" || rou.localCIDR == "" {
		return false
	}
	n.IsRunning = false
	tun, devName, err := OpenTunDevice(rou.tunIP)
	if err != nil {
		return n.IsRunning
	}
	// 设置 Sunny 的出口 IP 地址为原默认网关 IP
	if !n.Sunny.SetOutRouterIP(defaultGatewayIP) {
		_ = tun.Close()
		return n.IsRunning
	}
	n.tun = tun
	buf := make([]byte, 65535)
	n.tun = tun
	n.IsRunning = true
	rou.tunName = devName
	if rou.applyRouting() != nil {
		_ = tun.Close()
		rou.cleanup()
		return false
	}
	startWatchdog()
	go func() {
		defer func() {
			_ = tun.Close()
			rou.cleanup()
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
				time.Sleep(100 * time.Millisecond)
				continue
			}
			// 将读取到的数据拷贝到新切片中
			var packet []byte
			packet = append(packet, buf[:nBytes]...)
			//我在这里输出，当局域网设备连接时，一点输出都没有
			go func() {
				n.parsePacket(packet)
			}()
		}
	}()
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
