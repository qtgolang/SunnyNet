//go:build darwin
// +build darwin

package Tun

// sh1 是恢复默认网关的方式，较快，稳定性未知
const sh1 = `
#!/bin/bash

# ==============================
# 🛠 SunnyTunCancel.sh
# 功能：后台监控目标进程，当进程退出时自动恢复默认网关
# ==============================

# ----------------------------------
# 🌀 如果当前脚本未被后台化，则重新后台运行自己，避免被主进程信号杀掉
# ----------------------------------
if [ -z "$DISOWNED" ]; then 
export DISOWNED=1                              # 标记为已后台化，防止递归 
nohup "$0" "$@" >/dev/null 2>&1 &              # 重新启动自己并脱离主进程 
echo $!										   # 重点：输出后台进程 PID
exit 0                                         # 退出当前这份脚本
fi

# ----------------------------------
# 📌 获取目标 PID
# ----------------------------------
TARGET_PID=$1
ORIGINAL_GW=$2
ORIGINAL_Name=$3
if [ -z "$TARGET_PID" ]; then
echo "❌ 请传入一个 PID"
exit 1
fi

# ----------------------------------
# 🔐 检查是否拥有 ROOT 权限
# ----------------------------------
if [ "$EUID" -eq 0 ]; then
echo "✅ 当前脚本以 ROOT 权限运行 (EUID=0)"
else
echo "❌ 当前脚本没有 ROOT 权限 (EUID=$EUID)"
fi

# ----------------------------------
# 👀 循环检测目标进程是否仍然存在
# kill -0 不发送实际信号，只用于判断进程是否存在
# ----------------------------------
echo "🕓 正在监控进程 PID: $TARGET_PID"
while kill -0 "$TARGET_PID" 2>/dev/null; do
sleep 1
done
sudo route delete default
sudo route -n delete -inet default -ifscope $ORIGINAL_Name
sudo route add default $ORIGINAL_GW
`

// sh2 是重启网卡的方式，较慢，稳定性应该姣好,先用sh1，这个也先保留
const sh2 = `
#!/bin/bash

# ==============================
# 🛠 SunnyTunCancel.sh
# 功能：后台监控目标进程，当进程退出时自动重启所有物理网卡
# ==============================

# ----------------------------------
# 🌀 如果当前脚本未被后台化，则重新后台运行自己，避免被主进程信号杀掉
# ----------------------------------
if [ -z "$DISOWNED" ]; then 
export DISOWNED=1                              # 标记为已后台化，防止递归 
nohup "$0" "$@" >/dev/null 2>&1 &              # 重新启动自己并脱离主进程 
echo $!										   # 重点：输出后台进程 PID
exit 0                                         # 退出当前这份脚本
fi

# ----------------------------------
# 📌 获取目标 PID
# ----------------------------------
TARGET_PID=$1
if [ -z "$TARGET_PID" ]; then
echo "❌ 请传入一个 PID"
exit 1
fi

# ----------------------------------
# 🔐 检查是否拥有 ROOT 权限
# ----------------------------------
if [ "$EUID" -eq 0 ]; then
echo "✅ 当前脚本以 ROOT 权限运行 (EUID=0)"
else
echo "❌ 当前脚本没有 ROOT 权限 (EUID=$EUID)"
fi

# ----------------------------------
# 👀 循环检测目标进程是否仍然存在
# kill -0 不发送实际信号，只用于判断进程是否存在
# ----------------------------------
echo "🕓 正在监控进程 PID: $TARGET_PID"
while kill -0 "$TARGET_PID" 2>/dev/null; do
sleep 1
done
echo "⚠️ 进程 $TARGET_PID 已退出，开始重启所有网卡..."

# ----------------------------------
# 🌐 获取所有网卡，排除 lo0 和 utun 接口
# ----------------------------------
interfaces=$(ifconfig -l | tr ' ' '\n' | grep -vE '^lo0$' | grep -vE '^utun')

if [ -z "$interfaces" ]; then
echo "❌ 未检测到可重启的物理网卡"
exit 0
fi

echo "🔍 检测到网卡: $interfaces"

# ----------------------------------
# 🔁 依次重启所有网卡
# ----------------------------------
for iface in $interfaces; do
echo "🚧 正在重启网卡: $iface"
ifconfig "$iface" down >/dev/null 2>&1
sleep 1
ifconfig "$iface" up >/dev/null 2>&1
echo "✅ $iface 已重启"
done

# ----------------------------------
# 🏁 结束
# ----------------------------------
echo "🎉 所有物理网卡重启完成"
`
