


@echo off
set NDK=F:\init\windows-ndk-x86_64
set CGO_ENABLED=1


set tmpPath=G:\AndroidProject\SunnyNet\app\src\main\jniLibs\

set GOOS=android
set GOARCH=arm64
set CC=%NDK%\bin\aarch64-linux-android21-clang
echo [Full]_Build_Android_arm64-v8a.so
go build -trimpath  -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%arm64-v8a/libSunnyNet.so"