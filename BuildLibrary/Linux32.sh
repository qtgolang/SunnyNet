echo ""  > /dev/null
echo "请使用Linux环境编译,可以使用WSL"  > /dev/null
echo ""  > /dev/null
echo ""
echo "正在编译..."
set CGO_ENABLED=1
set GOOS=linux
set GOARCH=386
tmpPath=$(dirname "$(readlink -f "$0")")
parentPath=$(dirname "$tmpPath")
cd "$parentPath"

go build -buildmode=c-shared  -ldflags "-s -w" -o "$tmpPath/Library/Linux/x86/Sunny.so"
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
