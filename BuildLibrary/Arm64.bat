
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
go build -trimpath  -buildmode=c-shared  -ldflags "-s -w" -o  "G:\AndroidProject\TunTest\app\src\main\jniLibs\arm64-v8a\libSunnyNet.so"
copy G:\AndroidProject\TunTest\app\src\main\jniLibs\arm64-v8a\libSunnyNet.so G:\AndroidProject\SunnyNetDemo\app\src\main\jniLibs\arm64-v8a\libSunnyNet.so
@echo on
