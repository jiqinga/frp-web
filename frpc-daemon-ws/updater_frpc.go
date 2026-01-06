/*
 * @Author              : Kilo Code
 * @Date                : 2025-12-01
 * @Description         : 客户端更新器 - frpc 更新逻辑
 */
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// updateFrpc 更新 frpc
func (u *Updater) updateFrpc(version string, downloadURL string) {
	log.Printf("[Updater] 开始更新 frpc 到版本 %s", version)
	updateType := UpdateTypeFrpc

	// 阶段1: 下载 (0-60%)
	u.reportProgress(updateType, StageDownloading, 0, "开始下载 frpc...", 0, 0)

	tempDir := os.TempDir()
	archivePath := filepath.Join(tempDir, "frpc_update"+u.getArchiveExt())

	totalBytes, err := u.downloadFile(downloadURL, archivePath, func(downloaded, total int64) {
		progress := int(float64(downloaded) / float64(total) * 60)
		u.reportProgress(updateType, StageDownloading, progress, fmt.Sprintf("下载中... %.1f%%", float64(downloaded)/float64(total)*100), total, downloaded)
	})
	if err != nil {
		log.Printf("[Updater] ❌ 下载失败: %v", err)
		u.reportProgress(updateType, StageFailed, 0, "下载失败: "+err.Error(), 0, 0)
		u.reportResult(updateType, false, version, "下载失败: "+err.Error())
		return
	}
	log.Printf("[Updater] ✅ 下载完成，文件大小: %d bytes", totalBytes)

	// 阶段2: 停止 frpc (60-70%)
	u.reportProgress(updateType, StageStopping, 60, "正在停止 frpc...", totalBytes, totalBytes)

	if err := u.frpcMgr.Shutdown(); err != nil {
		log.Printf("[Updater] ⚠️ 停止 frpc 失败: %v，继续更新", err)
	}
	time.Sleep(2 * time.Second)
	u.reportProgress(updateType, StageStopping, 70, "frpc 已停止", totalBytes, totalBytes)

	// 阶段3: 替换文件 (70-80%)
	u.reportProgress(updateType, StageReplacing, 70, "正在替换 frpc 文件...", totalBytes, totalBytes)

	// 备份旧文件
	backupPath := u.cfg.FrpcPath + ".backup"
	if _, err := os.Stat(u.cfg.FrpcPath); err == nil {
		if err := os.Rename(u.cfg.FrpcPath, backupPath); err != nil {
			log.Printf("[Updater] ⚠️ 备份旧文件失败: %v", err)
		}
	}

	// 解压新文件
	newFrpcPath, err := u.extractFrpc(archivePath, filepath.Dir(u.cfg.FrpcPath))
	if err != nil {
		log.Printf("[Updater] ❌ 解压失败: %v", err)
		if _, err := os.Stat(backupPath); err == nil {
			os.Rename(backupPath, u.cfg.FrpcPath)
		}
		u.reportProgress(updateType, StageFailed, 0, "解压失败: "+err.Error(), 0, 0)
		u.reportResult(updateType, false, version, "解压失败: "+err.Error())
		return
	}

	// 如果解压出的文件名与配置的不同，重命名
	if newFrpcPath != u.cfg.FrpcPath {
		if err := os.Rename(newFrpcPath, u.cfg.FrpcPath); err != nil {
			log.Printf("[Updater] ⚠️ 重命名文件失败: %v", err)
		}
	}

	// 设置执行权限
	if runtime.GOOS != "windows" {
		os.Chmod(u.cfg.FrpcPath, 0755)
	}

	u.reportProgress(updateType, StageReplacing, 80, "文件替换完成", totalBytes, totalBytes)

	// 阶段4: 启动 frpc (80-95%)
	u.reportProgress(updateType, StageStarting, 80, "正在启动 frpc...", totalBytes, totalBytes)

	if err := u.startFrpc(); err != nil {
		log.Printf("[Updater] ❌ 启动 frpc 失败: %v", err)
		if _, err := os.Stat(backupPath); err == nil {
			os.Remove(u.cfg.FrpcPath)
			os.Rename(backupPath, u.cfg.FrpcPath)
			u.startFrpc()
		}
		u.reportProgress(updateType, StageFailed, 0, "启动失败: "+err.Error(), 0, 0)
		u.reportResult(updateType, false, version, "启动失败: "+err.Error())
		return
	}

	u.reportProgress(updateType, StageStarting, 95, "frpc 已启动", totalBytes, totalBytes)

	// 清理
	os.Remove(archivePath)
	os.Remove(backupPath)

	// 完成
	u.reportProgress(updateType, StageCompleted, 100, "更新完成", totalBytes, totalBytes)
	u.reportResult(updateType, true, version, "frpc 更新成功")
	log.Printf("[Updater] ✅ frpc 更新完成，版本: %s", version)
}
