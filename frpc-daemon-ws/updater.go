/*
 * @Author              : Kilo Code
 * @Date                : 2025-12-01
 * @Description         : 客户端更新器 - 核心类型定义和进度上报
 */
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// UpdateType 更新类型
type UpdateType string

const (
	UpdateTypeFrpc   UpdateType = "frpc"
	UpdateTypeDaemon UpdateType = "daemon"
)

// UpdateStage 更新阶段
type UpdateStage string

const (
	StageDownloading UpdateStage = "downloading"
	StageStopping    UpdateStage = "stopping"
	StageReplacing   UpdateStage = "replacing"
	StageStarting    UpdateStage = "starting"
	StageCompleted   UpdateStage = "completed"
	StageFailed      UpdateStage = "failed"
)

// UpdateProgress 更新进度
type UpdateProgress struct {
	UpdateType      UpdateType  `json:"update_type"`
	Stage           UpdateStage `json:"stage"`
	Progress        int         `json:"progress"`
	Message         string      `json:"message"`
	TotalBytes      int64       `json:"total_bytes"`
	DownloadedBytes int64       `json:"downloaded_bytes"`
}

// UpdateResult 更新结果
type UpdateResult struct {
	UpdateType UpdateType `json:"update_type"`
	Success    bool       `json:"success"`
	Version    string     `json:"version"`
	Message    string     `json:"message"`
}

// Updater 更新器
type Updater struct {
	cfg          *Config
	frpcMgr      *FrpcManager
	wsClient     *WSClient
	progressChan chan UpdateProgress
	resultChan   chan UpdateResult
	done         chan struct{} // 用于退出 progressReporter 协程
	logFile      *os.File
	logFileMu    sync.Mutex // 保护 logFile 写入的互斥锁
}

// NewUpdater 创建更新器
func NewUpdater(cfg *Config, frpcMgr *FrpcManager, wsClient *WSClient) *Updater {
	return &Updater{
		cfg:          cfg,
		frpcMgr:      frpcMgr,
		wsClient:     wsClient,
		progressChan: make(chan UpdateProgress, 100),
		resultChan:   make(chan UpdateResult, 10),
		done:         make(chan struct{}),
	}
}

// Start 启动更新器的进度上报协程
func (u *Updater) Start() {
	go u.progressReporter()
}

// progressReporter 进度上报协程
func (u *Updater) progressReporter() {
	for {
		select {
		case <-u.done:
			return
		case progress := <-u.progressChan:
			u.wsClient.SendUpdateProgress(progress)
		case result := <-u.resultChan:
			u.wsClient.SendUpdateResult(result)
		}
	}
}

// Stop 停止更新器
func (u *Updater) Stop() {
	close(u.done)
}

// HandleUpdate 处理更新命令
func (u *Updater) HandleUpdate(updateType string, version string, downloadURL string, mirrorID uint) {
	log.Printf("[Updater] ========== 收到更新命令 ==========")
	log.Printf("[Updater] 更新类型: %s", updateType)
	log.Printf("[Updater] 目标版本: %s", version)
	log.Printf("[Updater] 下载地址: %s", downloadURL)
	log.Printf("[Updater] 镜像ID: %d", mirrorID)

	switch UpdateType(updateType) {
	case UpdateTypeFrpc:
		go u.updateFrpc(version, downloadURL)
	case UpdateTypeDaemon:
		go u.updateDaemon(version, downloadURL)
	default:
		log.Printf("[Updater] ❌ 不支持的更新类型: %s", updateType)
		u.reportResult(UpdateType(updateType), false, version, "不支持的更新类型")
	}
}

// reportProgress 上报进度
func (u *Updater) reportProgress(updateType UpdateType, stage UpdateStage, progress int, message string, totalBytes, downloadedBytes int64) {
	u.progressChan <- UpdateProgress{
		UpdateType:      updateType,
		Stage:           stage,
		Progress:        progress,
		Message:         message,
		TotalBytes:      totalBytes,
		DownloadedBytes: downloadedBytes,
	}
}

// reportResult 上报结果
func (u *Updater) reportResult(updateType UpdateType, success bool, version, message string) {
	u.resultChan <- UpdateResult{
		UpdateType: updateType,
		Success:    success,
		Version:    version,
		Message:    message,
	}
}

// writeUpdateLog 写入更新日志到文件
func (u *Updater) writeUpdateLog(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] %s\n", timestamp, message)

	log.Printf("[Updater] %s", message)

	u.logFileMu.Lock()
	defer u.logFileMu.Unlock()
	if u.logFile != nil {
		u.logFile.WriteString(logLine)
		u.logFile.Sync()
	}
}

// initUpdateLog 初始化更新日志文件
func (u *Updater) initUpdateLog(updateType UpdateType) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDir := filepath.Dir(exePath)

	logPath := filepath.Join(exeDir, fmt.Sprintf("update_%s.log", updateType))
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	u.logFileMu.Lock()
	u.logFile = f
	u.logFile.WriteString("\n")
	u.logFile.WriteString("========================================\n")
	u.logFile.WriteString(fmt.Sprintf("更新开始时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	u.logFile.WriteString(fmt.Sprintf("更新类型: %s\n", updateType))
	u.logFile.WriteString(fmt.Sprintf("操作系统: %s/%s\n", runtime.GOOS, runtime.GOARCH))
	u.logFile.WriteString("========================================\n")
	u.logFile.Sync()
	u.logFileMu.Unlock()

	return nil
}

// closeUpdateLog 关闭更新日志文件
func (u *Updater) closeUpdateLog() {
	u.logFileMu.Lock()
	defer u.logFileMu.Unlock()
	if u.logFile != nil {
		u.logFile.WriteString(fmt.Sprintf("更新结束时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
		u.logFile.WriteString("========================================\n\n")
		u.logFile.Close()
		u.logFile = nil
	}
}

// startFrpc 启动 frpc
func (u *Updater) startFrpc() error {
	if runtime.GOOS != "windows" && u.cfg.FrpcServiceName != "" {
		cmd := exec.Command("systemctl", "start", u.cfg.FrpcServiceName)
		return cmd.Run()
	}

	cmd := exec.Command(u.cfg.FrpcPath, "-c", u.cfg.FrpcConfig)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		cmd.Wait()
	}()

	return nil
}

// restartDaemonService 重启 daemon 服务
func (u *Updater) restartDaemonService() error {
	if runtime.GOOS == "windows" {
		log.Printf("[Updater] Windows 平台，直接退出进程")
		os.Exit(0)
		return nil
	}

	serviceName := u.cfg.DaemonServiceName
	if serviceName == "" {
		serviceName = "frpc-daemon"
	}

	log.Printf("[Updater] 执行 systemctl restart %s", serviceName)
	cmd := exec.Command("systemctl", "restart", serviceName)
	if err := cmd.Start(); err != nil {
		log.Printf("[Updater] systemctl restart 失败: %v，尝试直接退出", err)
		os.Exit(0)
	}

	return nil
}

// getArchiveExt 获取压缩包扩展名
func (u *Updater) getArchiveExt() string {
	if runtime.GOOS == "windows" {
		return ".zip"
	}
	return ".tar.gz"
}

// getExeExt 获取可执行文件扩展名
func (u *Updater) getExeExt() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
