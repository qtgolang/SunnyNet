@echo off
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=386
set tmpPath=%~dp0
cd %tmpPath:~0,1%:
for %%I in ("%tmpPath%..\") do set "parentPath=%%~fI"
cd %parentPath%
go build -trimpath -buildmode=c-shared  -tags mini  -ldflags "-s -w" -o "G:\Sunny\SunnyNet\BuildLibrary\SunnyNet.dll"
@echo on