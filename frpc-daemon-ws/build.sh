#!/bin/sh
###
# @Author              : 寂情啊
# @Date                : 2025-11-25 17:02:22
# @LastEditors         : 寂情啊
# @LastEditTime        : 2025-12-01 17:00:15
# @FilePath            : frp-web-testfrpc-daemon-wsbuild.sh
# @Description         : 多平台构建脚本，支持外部传入 VERSION 和 OUTPUT_DIR
# @倾尽绿蚁花尽开，问潭底剑仙安在哉
###

# 支持外部传入参数，否则使用默认值
# VERSION: 版本号，默认使用当前时间戳
# OUTPUT_DIR: 输出目录，默认为 ../backend/data/daemon
VERSION="${VERSION:-$(date +%Y%m%d-%H%M%S)}"
OUTPUT_DIR="${OUTPUT_DIR:-../backend/data/daemon}"

mkdir -p "$OUTPUT_DIR"

LDFLAGS="-s -w -X main.BuildTime=${VERSION}"

echo "构建 frpc-daemon-ws"
echo "版本号: ${VERSION}"
echo "输出目录: ${OUTPUT_DIR}"
echo ""

# Linux AMD64
echo "构建 Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o "${OUTPUT_DIR}/frpc-daemon-ws-linux-amd64"

# Linux ARM64
echo "构建 Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o "${OUTPUT_DIR}/frpc-daemon-ws-linux-arm64"

# Linux ARM
echo "构建 Linux ARM..."
GOOS=linux GOARCH=arm go build -ldflags="${LDFLAGS}" -o "${OUTPUT_DIR}/frpc-daemon-ws-linux-arm"

# Windows AMD64
echo "构建 Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o "${OUTPUT_DIR}/frpc-daemon-ws-windows-amd64.exe"

# Windows 386
echo "构建 Windows 386..."
GOOS=windows GOARCH=386 go build -ldflags="${LDFLAGS}" -o "${OUTPUT_DIR}/frpc-daemon-ws-windows-386.exe"

# macOS AMD64
echo "构建 macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o "${OUTPUT_DIR}/frpc-daemon-ws-darwin-amd64"

# macOS ARM64
echo "构建 macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o "${OUTPUT_DIR}/frpc-daemon-ws-darwin-arm64"

echo ""
echo "✅ 构建完成!"
echo "版本号: ${VERSION}"
echo "输出目录: ${OUTPUT_DIR}"
ls -la "${OUTPUT_DIR}"/frpc-daemon-ws-* 2>/dev/null || echo "文件列表获取失败（可能在 Docker 环境中）"