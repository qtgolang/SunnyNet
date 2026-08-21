


@echo off
set NDK=F:\init\windows-ndk-x86_64
set CGO_ENABLED=1


set tmpPath=G:\AndroidProject\SunnyNet\app\src\main\jniLibs\

set GOOS=android
set GOARCH=arm64
set CC=%NDK%\bin\aarch64-linux-android21-clang
echo [Full]_Build_Android_arm64-v8a.so
go build -trimpath  -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%arm64-v8a/libSunnyNet.so"
set GOOS=android
set GOARCH=arm
set CC=%NDK%\bin\armv7a-linux-androideabi21-clang
echo [Full]_Build_Android_armeabi-v7a.so
go build -trimpath  -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%armeabi-v7a/libSunnyNet.so"

set GOOS=android
set GOARCH=386
set CC=%NDK%\bin\x86_64-linux-android21-clang
echo [Full]_Build_Android_x86.so
go build -trimpath  -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%x86/libSunnyNet.so"

set GOOS=android
set GOARCH=amd64
set CC=%NDK%\bin\x86_64-linux-android21-clang
echo [Full]_Build_Android_x86_64.so
go build -trimpath  -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%x86_64/libSunnyNet.so"