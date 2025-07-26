
@echo off
set NDK=E:\init\windows-ndk-x86_64
set CGO_ENABLED=1

set tmpPath=%~dp0
cd %tmpPath:~0,1%:
for %%I in ("%tmpPath%..") do set "parentPath=%%~fI"
cd %parentPath%

set GOOS=android
set GOARCH=arm64
set CC=%NDK%\bin\aarch64-linux-android21-clang
echo [Full]_Build_Android_arm64-v8a.so
go build -ldflags "-s -w" -o "%tmpPath%Library/Full/Android/arm64-v8a/SunnyNet"
echo [Mini]_Build_Android_arm64-v8a.so
go build -tags mini  -ldflags "-s -w" -o "%tmpPath%Library/Mini/Android/arm64-v8a/SunnyNet"

@echo on
