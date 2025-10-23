//go:build linux && !android
// +build linux,!android

package Tun // 定义包名为 Tun

import ( // 导入所需包
	"bytes"         // 用于接收子进程输出
	"fmt"           // 字符串格式化
	"io"            // IO 接口
	"os"            // 文件与进程操作
	"os/exec"       // 执行外部命令
	"path/filepath" // 构造临时文件路径
	"regexp"        // 正则解析后台 PID
	"strconv"       // 字符串转数字
	"strings"       // 构建脚本文本
	"syscall"       // 进程存活检测
	"time"          // 定时控制
)

// 全局脚本文本（由 CreateSh 生成）
var sh = "" // 存放生成后的 Shell 脚本文本

// 监控器状态（保持你原有语义）
var watchdogStarted bool // 标记监控脚本是否已启动
var watchdogPid int      // 记录后台化脚本 PID

// 脚本落盘路径（保持与原代码一致）
const scriptPath = "/usr/local/bin/SunnyTunCancel.sh" // 清理脚本的绝对路径

// writeFileAtomic 原子写文件（避免部分写入导致执行失败）
func writeFileAtomic(path string, data []byte, perm os.FileMode) error { // 定义原子写文件函数
	dir := filepath.Dir(path)                                 // 获取目录
	tmp := filepath.Join(dir, "."+filepath.Base(path)+".tmp") // 生成临时文件路径
	if err := os.WriteFile(tmp, data, perm); err != nil {     // 先写入临时文件
		return err // 返回错误
	}
	return os.Rename(tmp, path) // 重命名替换，达到原子写入效果
}

// CreateSh 生成守护/清理脚本（根据 r 的状态把命令具体化到脚本内）
func (r *TunRouter) CreateSh() { // 定义 CreateSh 方法
	var b strings.Builder // 使用高效的字符串构建器
	// 写入脚本头与后台化逻辑
	b.WriteString(`#!/bin/bash
set -u
set -o pipefail

# ==============================
# 🛠 SunnyTunCancel.sh
# 功能：后台监控目标进程，当进程退出时自动恢复默认网络配置
# ==============================

# 若未后台化，则后台化自己并仅输出后台 PID 一行
if [ -z "${DISOWNED-}" ]; then
  export DISOWNED=1
  nohup "$0" "$@" >/dev/null 2>&1 &
  printf "%s\n" "$!"
  exit 0
fi

# 目标 PID（第一个参数）
TARGET_PID="${1:-}"
if [ -z "${TARGET_PID}" ]; then
  echo "missing TARGET_PID" >&2
  exit 1
fi

echo "🕓 监控进程 PID: ${TARGET_PID}"

# 轮询检测目标进程是否仍在
while kill -0 "${TARGET_PID}" 2>/dev/null; do
  sleep 1
done

# === 以下为恢复/清理动作 ===
`)
	// 删除 iptables 规则（如果你进程曾添加过）
	if len(r.iptablesRule) > 0 { // 如果记录了 iptables 规则
		del := append([]string{}, r.iptablesRule...) // 拷贝一份
		if len(del) > 3 && del[3] == "-A" {          // 如果第四个参数为 -A
			del[3] = "-D" // 替换为 -D，实现删除
		}
		b.WriteString(strings.Join(del, " ")) // 拼接命令
		b.WriteByte('\n')                     // 换行
	}
	// 删除策略路由（按优先级顺序）
	b.WriteString("ip rule del priority 220 || true\n") // 删兜底
	b.WriteString("ip rule del priority 120 || true\n") // 删本地子网
	b.WriteString("ip rule del priority 100 || true\n") // 删 fwmark
	b.WriteString("ip rule del priority 90  || true\n") // 删 from hostIP

	// 清理路由表 200 与 100（仅删我们加的几条，避免误删用户其它配置）
	b.WriteString("ip route flush table 200 || true\n")                                                          // 清空 200
	b.WriteString(fmt.Sprintf("ip route del default via %s dev %s table 100 || true\n", r.defGWIP, r.ifaceName)) // 删默认
	b.WriteString(fmt.Sprintf("ip route del %s/32 dev %s table 100 || true\n", r.defGWIP, r.ifaceName))          // 删/32
	b.WriteString(fmt.Sprintf("ip route del %s dev %s table 100 || true\n", r.localCIDR, r.ifaceName))           // 删本地网段

	// 还原 rp_filter
	b.WriteString("sysctl -w net.ipv4.conf.all.rp_filter=1 || true\n")                          // all=1
	b.WriteString("sysctl -w net.ipv4.conf.default.rp_filter=1 || true\n")                      // default=1
	b.WriteString(fmt.Sprintf("sysctl -w net.ipv4.conf.%s.rp_filter=1 || true\n", r.ifaceName)) // iface=1
	b.WriteString(fmt.Sprintf("sysctl -w net.ipv4.conf.%s.rp_filter=1 || true\n", r.tunName))   // tun=1
	// 删除 TUN 接口（双保险）
	b.WriteString(fmt.Sprintf("ip link del %s || true\n", r.tunName)) // 删除 TUN
	// 收尾注释
	b.WriteString("echo '✅ SunnyTun 清理完成'\n") // 提示完成
	// 回填到全局 sh
	sh = b.String() // 将生成好的脚本文本保存到全局变量
}

// startWatchdog 启动路由恢复监控脚本（防止主进程异常退出导致网络不恢复）
func startWatchdog() {   // 定义 startWatchdog 函数
	if watchdogStarted { // 如果已启动
		return // 直接返回
	}
	watchdogStarted = true // 标记已启动
	rou.CreateSh()
	_ = os.Remove(scriptPath)                          // 先尝试删除旧脚本（忽略错误）
	_ = writeFileAtomic(scriptPath, []byte(sh), 0o777) // 原子写入新脚本并赋可执行权限

	go func() { // 启动后台 goroutine 持续监控
		mainPid := os.Getpid() // 获取当前主进程 PID（被监控对象）
		for {                  // 循环监控
			// 若已有后台脚本在运行，则无需拉起
			if watchdogPid > 0 && syscall.Kill(watchdogPid, 0) == nil { // 若 PID 存在
				time.Sleep(time.Second) // 休眠 1 秒后继续检测
				continue                // 进入下次循环
			}

			// 运行脚本：使用 /bin/bash，保障 $EUID 等 Bash 变量可用
			// 兼容你原来的三参形式：传 (PID, defaultGatewayIf, ifaceName)
			// 即使脚本未使用这两个参数也不影响（保持你的外部调用兼容）
			cmd := exec.Command("/bin/bash", scriptPath, strconv.Itoa(mainPid), defaultGatewayIf, ifaceName) // 构造命令
			var buffer bytes.Buffer                                                                          // 创建输出缓冲区
			cmd.Stdout = &buffer                                                                             // 捕获标准输出（脚本首次运行仅输出后台 PID）
			cmd.Stderr = io.Discard                                                                          // 丢弃标准错误（避免污染 PID 解析）
			if err := cmd.Start(); err != nil {                                                              // 启动脚本失败
				watchdogPid = 0         // PID 置 0，稍后重试
				time.Sleep(time.Second) // 等待 1 秒
				continue                // 下一轮重试
			}
			_ = cmd.Wait() // 等待脚本前台进程退出（它会很快退出，只打印后台 PID）
			// 解析输出中的第一个十进制数字作为 PID（更鲁棒，避免杂讯）
			out := strings.TrimSpace(buffer.String()) // 去掉首尾空白
			re := regexp.MustCompile(`\b\d+\b`)       // 匹配第一个整数
			pidStr := re.FindString(out)              // 提取 PID 字符串
			if pidStr == "" {                         // 若未匹配到
				watchdogPid = 0         // 置 0 以便重试
				time.Sleep(time.Second) // 等待 1 秒
				continue                // 重试
			}
			pid, err := strconv.Atoi(pidStr) // 转换为整数
			if err != nil || pid <= 0 {      // 若解析失败或不合理
				watchdogPid = 0         // 置 0
				time.Sleep(time.Second) // 等待 1 秒
				continue                // 重试
			}
			watchdogPid = pid // 记录后台脚本 PID

			// 正常情况下，此处后台脚本常驻；定期确认其存活
			time.Sleep(time.Second) // 休眠 1 秒再继续下一轮检查
		}
	}() // 启动 goroutine 结束
}
