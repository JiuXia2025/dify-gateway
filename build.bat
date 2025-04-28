@echo off

setlocal enabledelayedexpansion

echo 开始编译 Dify Gateway...

:: 创建输出目录
if not exist "bin" mkdir bin

:: 设置编译时间
set BUILD_TIME=%date:~0,4%%date:~5,2%%date:~8,2%_%time:~0,2%%time:~3,2%%time:~6,2%
set BUILD_TIME=%BUILD_TIME: =0%

:: 编译 Windows 64位
echo 正在编译 Windows 64位版本...
set GOOS=windows
set GOARCH=amd64
go build -o "bin\dify-gateway-windows-amd64-%BUILD_TIME%.exe" -ldflags "-s -w" main.go

:: 编译 Linux 64位
echo 正在编译 Linux 64位版本...
set GOOS=linux
set GOARCH=amd64
go build -o "bin\dify-gateway-linux-amd64-%BUILD_TIME%" -ldflags "-s -w" main.go

:: 编译 macOS 64位
echo 正在编译 macOS 64位版本...
set GOOS=darwin
set GOARCH=amd64
go build -o "bin\dify-gateway-darwin-amd64-%BUILD_TIME%" -ldflags "-s -w" main.go

:: 编译 macOS ARM64
echo 正在编译 macOS ARM64版本...
set GOOS=darwin
set GOARCH=arm64
go build -o "bin\dify-gateway-darwin-arm64-%BUILD_TIME%" -ldflags "-s -w" main.go

:: 复制配置文件
echo 正在复制配置文件...
copy app.yaml bin\

echo 编译完成！
echo 输出目录: %cd%\bin
echo.
echo 编译结果:
dir /b bin\

pause 