//go:build darwin
// +build darwin

package Tun

// sh1 是恢复默认网关的方式，较快，稳定性未知
const sh1 = `#!/bin/bash

# ==============================
# 🛠 SunnyTunCancel.sh（daemon 版）
# 功能：后台监控目标进程，当进程退出时自动恢复默认网关
# ==============================

LOG=/tmp/SunnyTunCancel.log                         # 日志
PIDFILE=/tmp/SunnyTunCancel.pid                     # PID 文件

# ----------------------------------
# 🌀 首次进入：用 daemon 真正守护（无 TTY、无 nohup 报错）
# ----------------------------------
if [ -z "$DISOWNED" ]; then                         # 如果未守护
    export DISOWNED=1                               # 标记，防递归
    /usr/sbin/daemon -f \                           # -f: 父进程立即退出
        -p "$PIDFILE" \                             # 写入 PID 文件
        -o "$LOG" \                                 # 日志输出
        /bin/bash "$0" "$@"                         # 以 bash 执行本脚本（进入下半段）
    exit 0                                          # 启动器退出
fi

# ----------------------------------
# 📌 参数
# ----------------------------------
TARGET_PID=$1                                       # 目标 PID
ORIGINAL_GW=$2                                      # 原默认网关 IP
ORIGINAL_IF=$3                                      # 原接口名
echo "[` + "`" + `date '+%F %T'` + "`" + `] started: EUID=$EUID args=$*" >>"$LOG"
if [ -z "$TARGET_PID" ]; then
    echo "[` + "`" + `date '+%F %T'` + "`" + `] ❌ 未传入 PID" >>"$LOG"
    exit 1
fi

# ----------------------------------
# 👀 监控
# ----------------------------------
echo "[` + "`" + `date '+%F %T'` + "`" + `] 🕓 监控 PID=$TARGET_PID" >>"$LOG"
while kill -0 "$TARGET_PID" 2>/dev/null; do
    sleep 1
done

# ----------------------------------
# ✅ 恢复路由
# ----------------------------------
echo "[` + "`" + `date '+%F %T'` + "`" + `] ⚠️ $TARGET_PID 退出，恢复路由..." >>"$LOG"
route delete default 2>>"$LOG" || true
if [ -n "$ORIGINAL_IF" ]; then
    route -n delete -inet default -ifscope "$ORIGINAL_IF" 2>>"$LOG" || true
fi
if [ -n "$ORIGINAL_GW" ]; then
    route add default "$ORIGINAL_GW" 2>>"$LOG" || true
    echo "[` + "`" + `date '+%F %T'` + "`" + `] ✅ 恢复默认网关 $ORIGINAL_GW" >>"$LOG"
else
    echo "[` + "`" + `date '+%F %T'` + "`" + `] ⚠️ 未提供 ORIGINAL_GW，跳过添加 default" >>"$LOG"
fi
echo "[` + "`" + `date '+%F %T'` + "`" + `] 🎯 完成" >>"$LOG"
exit 0

`

// sh2 是重启网卡的方式，较慢，稳定性应该姣好,先用sh1，这个也先保留
const sh2 = `#!/bin/bash

# ==============================
# 🛠 SunnyTunCancel.sh (daemon 版)
# 功能：后台监控目标进程，当进程退出时自动重启所有物理网卡
# ==============================

LOG=/tmp/SunnyTunCancel.log                         # 日志
PIDFILE=/tmp/SunnyTunCancel.pid                     # PID 文件

# ----------------------------------
# 🌀 守护进程启动逻辑（替代 nohup）
# ----------------------------------
if [ -z "$DISOWNED" ]; then
    export DISOWNED=1
    /usr/sbin/daemon -f \                    # -f 让父进程立刻退出，子进程成为真正守护进程
        -p "$PIDFILE" \                      # PID 写入文件
        -o "$LOG" \                          # 日志输出到 LOG
        /bin/bash "$0" "$@"                  # 重新以 bash 执行脚本本身
    exit 0
fi

echo "[\$(date '+%F %T')] ✅ 守护进程已启动，PID=\$\$，监控 PID=\$1" >>"\$LOG"

# ----------------------------------
# 📌 获取目标 PID
# ----------------------------------
TARGET_PID=\$1
if [ -z "\$TARGET_PID" ]; then
    echo "[\$(date '+%F %T')] ❌ 未传入 PID" >>"\$LOG"
    exit 1
fi

# ----------------------------------
# 🔐 检查是否拥有 ROOT 权限
# ----------------------------------
if [ "\$EUID" -eq 0 ]; then
    echo "[\$(date '+%F %T')] ✅ 当前脚本以 ROOT 权限运行 (EUID=0)" >>"\$LOG"
else
    echo "[\$(date '+%F %T')] ❌ 当前脚本没有 ROOT 权限 (EUID=\$EUID)" >>"\$LOG"
fi

# ----------------------------------
# 👀 循环检测目标进程是否仍然存在
# ----------------------------------
echo "[\$(date '+%F %T')] 🕓 正在监控进程 PID: \$TARGET_PID" >>"\$LOG"
while kill -0 "\$TARGET_PID" 2>/dev/null; do
    sleep 1
done
echo "[\$(date '+%F %T')] ⚠️ 进程 \$TARGET_PID 已退出，开始重启所有网卡..." >>"\$LOG"

# ----------------------------------
# 🌐 获取所有网卡，排除 lo0 和 utun 接口
# ----------------------------------
interfaces=\$(ifconfig -l | tr ' ' '\\n' | grep -vE '^lo0\$' | grep -vE '^utun')

if [ -z "\$interfaces" ]; then
    echo "[\$(date '+%F %T')] ❌ 未检测到可重启的物理网卡" >>"\$LOG"
    exit 0
fi

echo "[\$(date '+%F %T')] 🔍 检测到网卡: \$interfaces" >>"\$LOG"

# ----------------------------------
# 🔁 依次重启所有网卡
# ----------------------------------
for iface in \$interfaces; do
    echo "[\$(date '+%F %T')] 🚧 正在重启网卡: \$iface" >>"\$LOG"
    ifconfig "\$iface" down >/dev/null 2>&1
    # sleep 1
    ifconfig "\$iface" up >/dev/null 2>&1
    echo "[\$(date '+%F %T')] ✅ \$iface 已重启" >>"\$LOG"
done

# ----------------------------------
# 🏁 结束
# ----------------------------------
echo "[\$(date '+%F %T')] 🎉 所有物理网卡重启完成" >>"\$LOG"
`
