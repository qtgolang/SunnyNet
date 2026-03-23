#!/usr/bin/env bash

set -u

echo ""
echo "请使用 Linux 环境编译，可以使用 WSL" > /dev/null
echo ""
echo "正在编译..."

export CGO_ENABLED=1

tmpPath="$(dirname "$(readlink -f "$0")")"
parentPath="$(dirname "$tmpPath")"
cd "$parentPath" || exit 1

# 这里按你的实际安装路径改
LINUX_386_CC="i686-linux-gnu-gcc"
LINUX_AMD64_CC="x86_64-linux-gnu-gcc"
LINUX_ARM_CC="arm-linux-gnueabihf-gcc"
LINUX_ARM64_CC="aarch64-linux-gnu-gcc"

# 如果配了 osxcross，就填上；没配就留空
DARWIN_AMD64_CC="${DARWIN_AMD64_CC:-o64-clang}"
DARWIN_ARM64_CC="${DARWIN_ARM64_CC:-aarch64-apple-darwin-clang}"

build_one() {
  local name="$1"
  local goos="$2"
  local goarch="$3"
  local cc="$4"
  local out="$5"
  local tags="${6:-}"
  local goarm="${7:-}"

  mkdir -p "$(dirname "$out")"

  export GOOS="$goos"
  export GOARCH="$goarch"
  export CC="$cc"

  if [ -n "$goarm" ]; then
    export GOARM="$goarm"
  else
    unset GOARM 2>/dev/null || true
  fi

  if ! command -v "$CC" >/dev/null 2>&1; then
    echo ""
    echo "$name 编译失败：未找到交叉编译器 $CC"
    return 1
  fi

  if [ -n "$tags" ]; then
    go build -trimpath -tags "$tags" -buildmode=c-shared -ldflags "-s -w" -o "$out"
  else
    go build -trimpath -buildmode=c-shared -ldflags "-s -w" -o "$out"
  fi

  if [ $? -ne 0 ]; then
    echo ""
    echo "$name 编译失败！"
    return 1
  else
    echo "$name 编译完成！"
    return 0
  fi
}

# ================== Linux x86 ==================
build_one "Full Linux x86"  "linux" "386"   "$LINUX_386_CC"   "$tmpPath/Library/Full/Linux/x86/Sunny.so"
build_one "Mini Linux x86"  "linux" "386"   "$LINUX_386_CC"   "$tmpPath/Library/Mini/Linux/x86/Sunny.so"   "mini"

# ================== Linux x64 ==================
build_one "Full Linux x64"  "linux" "amd64" "$LINUX_AMD64_CC" "$tmpPath/Library/Full/Linux/x64/Sunny.so"
build_one "Mini Linux x64"  "linux" "amd64" "$LINUX_AMD64_CC" "$tmpPath/Library/Mini/Linux/x64/Sunny.so"  "mini"

# ================== Linux arm ==================
# GOARM 常见是 7，你也可以按目标改成 6/5
build_one "Full Linux arm"  "linux" "arm"   "$LINUX_ARM_CC"   "$tmpPath/Library/Full/Linux/arm/Sunny.so"  ""      "7"
build_one "Mini Linux arm"  "linux" "arm"   "$LINUX_ARM_CC"   "$tmpPath/Library/Mini/Linux/arm/Sunny.so"  "mini"  "7"

# ================== Linux arm64 ==================
build_one "Full Linux arm64" "linux" "arm64" "$LINUX_ARM64_CC" "$tmpPath/Library/Full/Linux/arm64/Sunny.so"
build_one "Mini Linux arm64" "linux" "arm64" "$LINUX_ARM64_CC" "$tmpPath/Library/Mini/Linux/arm64/Sunny.so" "mini"

echo ""
echo "全部编译流程结束"