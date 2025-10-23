


@echo off
set NDK=E:\init\windows-ndk-x86_64
set CGO_ENABLED=1


set GOOS=windows
set GOARCH=386
set tmpPath=%~dp0
cd %tmpPath:~0,1%:
for %%I in ("%tmpPath%..\") do set "parentPath=%%~fI"
cd %parentPath%
echo [Full]_Build_x86_DLL
go build -trimpath -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%Library\Full\windows\SunnyNet.dll"
echo [Mini]_Build_x86_DLL
go build -trimpath -tags mini -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%Library\Mini\windows\SunnyNet.dll"


set GOOS=windows
set GOARCH=amd64
echo [Full]_Build_x64_DLL
go build -trimpath -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%Library\Full\windows\SunnyNet64.dll"
echo [Mini]_Build_x64_DLL
go build -trimpath -tags mini -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%Library\Mini\windows\SunnyNet64.dll"


set GOOS=android
set GOARCH=arm64
set CC=%NDK%\bin\aarch64-linux-android21-clang
echo [Full]_Build_Android_arm64-v8a.so
go build -trimpath  -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%Library/Full/Android/arm64-v8a/libSunnyNet.so"
echo [Mini]_Build_Android_arm64-v8a.so
go build -trimpath  -tags mini -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%Library/Mini/Android/arm64-v8a/libSunnyNet.so"

set GOOS=android
set GOARCH=arm
set CC=%NDK%\bin\armv7a-linux-androideabi21-clang
echo [Full]_Build_Android_armeabi-v7a.so
go build -trimpath  -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%Library/Full/Android/armeabi-v7a/libSunnyNet.so"
echo [Mini]_Build_Android_armeabi-v7a.so
go build -trimpath  -tags mini -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%Library/Mini/Android/armeabi-v7a/libSunnyNet.so"

set GOOS=android
set GOARCH=386
set CC=%NDK%\bin\x86_64-linux-android21-clang
echo [Full]_Build_Android_x86.so
go build -trimpath  -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%Library/Full/Android/x86/libSunnyNet.so"
echo [Mini]_Build_Android_x86.so
go build -trimpath  -tags mini -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%Library/Mini/Android/x86/libSunnyNet.so"

set GOOS=android
set GOARCH=386
set CC=%NDK%\bin\x86_64-linux-android21-clang
echo [Full]_Build_Android_x86_64.so
go build -trimpath  -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%Library/Full/Android/x86_64/libSunnyNet.so"
echo [Mini]_Build_Android_x86_64.so
go build -trimpath  -tags mini -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%Library/Mini/Android/x86_64/libSunnyNet.so"


set GOOS=android
set GOARCH=386
set CC=%NDK%\bin\i686-linux-android16-clang
echo [Full]_Build_Android_x86.so
go build -trimpath  -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%Library/Full/Android/x86/libSunnyNet.so"
echo [Mini]_Build_Android_x86.so
go build -trimpath  -tags mini -buildmode=c-shared  -ldflags "-s -w" -o "%tmpPath%Library/Mini/Android/x86/libSunnyNet.so"
@echo on
