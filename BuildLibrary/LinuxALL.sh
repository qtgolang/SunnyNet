#!/usr/bin/env bash

set -u

echo ""
echo "请使用 Linux 环境编译，可以使用 WSL" > /dev/null
echo ""
echo "正在编译..."

export CGO_ENABLED=1

# 脚本所在目录
tmpPath="$(dirname "$(readlink -f "$0")")"
# 项目根目录
parentPath="$(dirname "$tmpPath")"
cd "$parentPath" || exit 1

# 这里按你的实际环境改
# 32 位这里改成 multilib 方式，不再依赖 i686-linux-gnu-gcc
LINUX_386_CC="gcc -m32"
# 当前机器如果本身就是 x64 Linux，直接用 gcc 就行
LINUX_AMD64_CC="gcc"
# arm 交叉编译器
LINUX_ARM_CC="arm-linux-gnueabihf-gcc"
# arm64 交叉编译器
LINUX_ARM64_CC="aarch64-linux-gnu-gcc"

# 如果配了 osxcross，就填上；没配就留空
DARWIN_AMD64_CC="${DARWIN_AMD64_CC:-o64-clang}"
DARWIN_ARM64_CC="${DARWIN_ARM64_CC:-aarch64-apple-darwin-clang}"

# 记录失败数量，最后统一汇总
FAIL_COUNT=0

build_one() {
  # 构建名称
  local name="$1"
  # 目标系统
  local goos="$2"
  # 目标架构
  local goarch="$3"
  # C 编译器命令，允许带参数
  local cc="$4"
  # 输出文件
  local out="$5"
  # 可选 tags
  local tags="${6:-}"
  # 可选 GOARM
  local goarm="${7:-}"
  # 取命令本体，给 command -v 检查用
  local cc_bin="${cc%% *}"

  # 创建输出目录
  mkdir -p "$(dirname "$out")"

  # 设置 Go 交叉编译环境
  export GOOS="$goos"
  export GOARCH="$goarch"
  export CC="$cc"

  # arm 单独设置 GOARM
  if [ -n "$goarm" ]; then
    export GOARM="$goarm"
  else
    unset GOARM 2>/dev/null || true
  fi

  # 检查编译器是否存在
  if ! command -v "$cc_bin" >/dev/null 2>&1; then
    echo ""
    echo "$name 编译失败：未找到交叉编译器 $cc_bin"
    FAIL_COUNT=$((FAIL_COUNT + 1))
    return 1
  fi

  echo ""
  echo "开始编译：$name"
  echo "GOOS=$GOOS GOARCH=$GOARCH CC=$CC"

  # 有 tags 就带上 tags
  if [ -n "$tags" ]; then
    go build -trimpath -tags "$tags" -buildmode=c-shared -ldflags "-s -w" -o "$out"
  else
    go build -trimpath -buildmode=c-shared -ldflags "-s -w" -o "$out"
  fi

  # 判断结果
  if [ $? -ne 0 ]; then
    echo ""
    echo "$name 编译失败！"
    FAIL_COUNT=$((FAIL_COUNT + 1))
    return 1
  fi

  echo "$name 编译完成！"
  return 0
}

# ================== Linux x86 ==================
build_one "Full Linux x86"  "linux" "386"   "$LINUX_386_CC"   "$tmpPath/Library/Full/Linux/x86/Sunny.so"
build_one "Mini Linux x86"  "linux" "386"   "$LINUX_386_CC"   "$tmpPath/Library/Mini/Linux/x86/Sunny.so"   "mini"

# ================== Linux x64 ==================
build_one "Full Linux x64"  "linux" "amd64" "$LINUX_AMD64_CC" "$tmpPath/Library/Full/Linux/x64/Sunny.so"
build_one "Mini Linux x64"  "linux" "amd64" "$LINUX_AMD64_CC" "$tmpPath/Library/Mini/Linux/x64/Sunny.so"   "mini"

# ================== Linux arm ==================
# GOARM 常见填 7，按目标设备也可以改成 6 或 5
build_one "Full Linux arm"  "linux" "arm"   "$LINUX_ARM_CC"   "$tmpPath/Library/Full/Linux/arm/Sunny.so"   ""      "7"
build_one "Mini Linux arm"  "linux" "arm"   "$LINUX_ARM_CC"   "$tmpPath/Library/Mini/Linux/arm/Sunny.so"   "mini"  "7"

# ================== Linux arm64 ==================
build_one "Full Linux arm64" "linux" "arm64" "$LINUX_ARM64_CC" "$tmpPath/Library/Full/Linux/arm64/Sunny.so"
build_one "Mini Linux arm64" "linux" "arm64" "$LINUX_ARM64_CC" "$tmpPath/Library/Mini/Linux/arm64/Sunny.so" "mini"

echo ""
echo "全部编译流程结束"

# 最后给个汇总，方便一眼看结果
if [ "$FAIL_COUNT" -gt 0 ]; then
  echo "有 $FAIL_COUNT 个目标编译失败"
  exit 1
else
  echo "全部目标编译成功"
  exit 0
fi