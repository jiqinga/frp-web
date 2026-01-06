@echo off
setlocal enabledelayedexpansion

set OUTPUT_DIR=..\backend\data\daemon

if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

REM 获取当前日期和时间作为版本号
for /f "tokens=1-3 delims=/ " %%a in ('date /t') do (
    set YEAR=%%a
    set MONTH=%%b
    set DAY=%%c
)
for /f "tokens=1-2 delims=: " %%a in ('time /t') do (
    set HOUR=%%a
    set MINUTE=%%b
)
REM 获取秒数
for /f "tokens=3 delims=:." %%a in ('echo %time%') do set SECOND=%%a

REM 去除可能的空格
set YEAR=%YEAR: =%
set MONTH=%MONTH: =%
set DAY=%DAY: =%
set HOUR=%HOUR: =%
set MINUTE=%MINUTE: =%
set SECOND=%SECOND: =%

REM 补零处理
if "%HOUR:~1%"=="" set HOUR=0%HOUR%

set BUILD_TIME=%YEAR%%MONTH%%DAY%-%HOUR%%MINUTE%%SECOND%
set LDFLAGS=-s -w -X main.BuildTime=%BUILD_TIME%

echo 构建 frpc-daemon-ws
echo 编译时间: %BUILD_TIME%
echo 输出目录: %OUTPUT_DIR%
echo.

echo 构建 Linux AMD64...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="%LDFLAGS%" -o %OUTPUT_DIR%\frpc-daemon-ws-linux-amd64

echo 构建 Linux ARM64...
set GOOS=linux
set GOARCH=arm64
go build -ldflags="%LDFLAGS%" -o %OUTPUT_DIR%\frpc-daemon-ws-linux-arm64

echo 构建 Linux ARM...
set GOOS=linux
set GOARCH=arm
go build -ldflags="%LDFLAGS%" -o %OUTPUT_DIR%\frpc-daemon-ws-linux-arm

echo 构建 Windows AMD64...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="%LDFLAGS%" -o %OUTPUT_DIR%\frpc-daemon-ws-windows-amd64.exe

echo 构建 Windows 386...
set GOOS=windows
set GOARCH=386
go build -ldflags="%LDFLAGS%" -o %OUTPUT_DIR%\frpc-daemon-ws-windows-386.exe

echo 构建 macOS AMD64...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="%LDFLAGS%" -o %OUTPUT_DIR%\frpc-daemon-ws-darwin-amd64

echo 构建 macOS ARM64...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags="%LDFLAGS%" -o %OUTPUT_DIR%\frpc-daemon-ws-darwin-arm64

echo.
echo 构建完成!
echo 版本号: %BUILD_TIME%
echo 输出目录: %OUTPUT_DIR%
dir %OUTPUT_DIR%\frpc-daemon-ws-*

endlocal