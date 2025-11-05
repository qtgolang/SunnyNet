echo ""  > /dev/null
echo "正在编译..."
set CGO_ENABLED=1
set GOOS=darwin
set GOARCH=arm64
tmpPath=$(dirname "$(readlink -f "$0")")
parentPath=$(dirname "$tmpPath")
cd "$parentPath"

go build  -trimpath -buildmode=c-shared  -ldflags "-s -w" -o "$tmpPath/Library/darwin/arm64/SunnyNet.dylib"
# 检查命令的退出状态码
if [ $? -ne 0 ]; then
  echo ""
  echo ""
  echo "编译失败！"
else
  echo ""
  echo ""
  echo "编译完成！"
fi
