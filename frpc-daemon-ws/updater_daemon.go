/*
 * @Author              : Kilo Code
 * @Date                : 2025-12-01
 * @Description         : 客户端更新器 - daemon 更新逻辑
 */
package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// updateDaemon 更新 daemon 自身
func (u *Updater) updateDaemon(version string, downloadURL string) {
	updateType := UpdateTypeDaemon
	actualVersion := BuildTime

	if err := u.initUpdateLog(updateType); err != nil {
		log.Printf("[Updater] ⚠️ 无法创建更新日志文件: %v", err)
	}
	defer u.closeUpdateLog()

	u.writeUpdateLog("开始更新 daemon 到版本 %s", version)
	u.writeUpdateLog("下载地址: %s", downloadURL)

	// 构建完整的下载URL
	fullURL := downloadURL
	if strings.HasPrefix(downloadURL, "/") {
		fullURL = strings.TrimSuffix(u.cfg.ServerURL, "/") + downloadURL
		fullURL = strings.Replace(fullURL, "ws://", "http://", 1)
		fullURL = strings.Replace(fullURL, "wss://", "https://", 1)
	}
	u.writeUpdateLog("完整下载地址: %s", fullURL)

	// 阶段1: 下载 (0-60%)
	u.reportProgress(updateType, StageDownloading, 0, "开始下载 daemon...", 0, 0)
	u.writeUpdateLog("阶段1: 开始下载...")

	currentExe, err := os.Executable()
	if err != nil {
		u.writeUpdateLog("❌ 获取当前可执行文件路径失败: %v", err)
		u.reportProgress(updateType, StageFailed, 0, "获取路径失败: "+err.Error(), 0, 0)
		u.reportResult(updateType, false, version, "获取路径失败: "+err.Error())
		return
	}
	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		u.writeUpdateLog("⚠️ 解析符号链接失败: %v，使用原始路径", err)
	}
	u.writeUpdateLog("当前可执行文件: %s", currentExe)

	exeDir := filepath.Dir(currentExe)
	newDaemonPath := filepath.Join(exeDir, "frpc-daemon-ws-new"+u.getExeExt())
	u.writeUpdateLog("临时文件路径: %s", newDaemonPath)

	totalBytes, err := u.downloadFile(fullURL, newDaemonPath, func(downloaded, total int64) {
		progress := int(float64(downloaded) / float64(total) * 60)
		u.reportProgress(updateType, StageDownloading, progress, "下载中...", total, downloaded)
	})
	if err != nil {
		u.writeUpdateLog("❌ 下载失败: %v", err)
		u.reportProgress(updateType, StageFailed, 0, "下载失败: "+err.Error(), 0, 0)
		u.reportResult(updateType, false, version, "下载失败: "+err.Error())
		return
	}
	u.writeUpdateLog("✅ 下载完成，文件大小: %d bytes", totalBytes)

	if runtime.GOOS != "windows" {
		os.Chmod(newDaemonPath, 0755)
		u.writeUpdateLog("已设置执行权限")
	}

	u.reportProgress(updateType, StageReplacing, 70, "准备替换文件...", totalBytes, totalBytes)
	u.writeUpdateLog("阶段2: 准备替换文件...")

	u.reportProgress(updateType, StageReplacing, 80, "正在替换文件...", totalBytes, totalBytes)

	backupPath := currentExe + ".backup"
	if err := copyFile(currentExe, backupPath); err != nil {
		u.writeUpdateLog("⚠️ 备份当前文件失败: %v", err)
	} else {
		u.writeUpdateLog("✅ 已备份当前文件到: %s", backupPath)
	}

	if err := os.Rename(newDaemonPath, currentExe); err != nil {
		u.writeUpdateLog("❌ 替换文件失败: %v", err)
		u.reportProgress(updateType, StageFailed, 0, "替换文件失败: "+err.Error(), 0, 0)
		u.reportResult(updateType, false, version, "替换文件失败: "+err.Error())
		return
	}
	u.writeUpdateLog("✅ 文件替换成功（使用原子 rename 操作）")

	if runtime.GOOS != "windows" {
		os.Chmod(currentExe, 0755)
		u.writeUpdateLog("已设置新文件执行权限")
	}

	u.reportProgress(updateType, StageStarting, 90, "准备重启服务...", totalBytes, totalBytes)
	u.writeUpdateLog("阶段3: 准备重启服务...")

	u.reportProgress(updateType, StageCompleted, 100, "文件已更新，即将重启", totalBytes, totalBytes)
	u.writeUpdateLog("📤 已发送进度 100%%")

	u.reportResult(updateType, true, actualVersion, "daemon 更新成功，即将重启")
	u.writeUpdateLog("📤 已发送更新结果")

	u.writeUpdateLog("⏳ 等待 WebSocket 消息发送完成...")
	time.Sleep(2 * time.Second)

	u.writeUpdateLog("🔄 准备重启服务...")
	serviceName := u.cfg.DaemonServiceName
	if serviceName == "" {
		serviceName = "frpc-daemon"
	}
	u.writeUpdateLog("服务名称: %s", serviceName)

	if err := u.restartDaemonService(); err != nil {
		u.writeUpdateLog("❌ 重启服务失败: %v", err)
	} else {
		u.writeUpdateLog("✅ 重启命令已发送")
	}
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
