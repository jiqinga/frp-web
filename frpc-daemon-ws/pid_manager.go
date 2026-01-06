/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-29 15:54:58
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-29 15:55:12
 * @FilePath            : frp-web-testfrpc-daemon-wspid_manager.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// PIDManager 管理 frpc 进程的 PID 文件
type PIDManager struct {
	pidFile string
}

// NewPIDManager 创建 PID 管理器
func NewPIDManager(installDir string) *PIDManager {
	pidFile := filepath.Join(installDir, "frpc.pid")
	if installDir == "" {
		pidFile = "frpc.pid"
	}
	return &PIDManager{pidFile: pidFile}
}

// SavePID 保存 PID 到文件
func (p *PIDManager) SavePID(pid int) error {
	dir := filepath.Dir(p.pidFile)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建 PID 目录失败: %w", err)
		}
	}
	if err := os.WriteFile(p.pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("写入 PID 文件失败: %w", err)
	}
	log.Printf("[PID] 已保存 PID %d 到 %s", pid, p.pidFile)
	return nil
}

// ReadPID 从文件读取 PID
func (p *PIDManager) ReadPID() (int, error) {
	data, err := os.ReadFile(p.pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("读取 PID 文件失败: %w", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("解析 PID 失败: %w", err)
	}
	return pid, nil
}

// RemovePID 删除 PID 文件
func (p *PIDManager) RemovePID() error {
	if err := os.Remove(p.pidFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除 PID 文件失败: %w", err)
	}
	log.Printf("[PID] 已删除 PID 文件: %s", p.pidFile)
	return nil
}

// IsProcessRunning 检查指定 PID 的进程是否在运行
func (p *PIDManager) IsProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Windows 上 FindProcess 总是成功，需要尝试发送信号来检查
	// Signal(0) 不会真正发送信号，只检查进程是否存在
	err = process.Signal(os.Signal(nil))
	return err == nil
}

// GetPIDFilePath 获取 PID 文件路径
func (p *PIDManager) GetPIDFilePath() string {
	return p.pidFile
}
