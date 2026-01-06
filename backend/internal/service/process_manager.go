/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-20 14:25:42
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-20 14:25:58
 * @FilePath            : frp-web-testbackendinternalserviceprocess_manager.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package service

import (
	"fmt"
	"frp-web-panel/internal/model"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

type ProcessManager struct{}

func NewProcessManager() *ProcessManager {
	return &ProcessManager{}
}

func (pm *ProcessManager) Start(server *model.FrpServer) error {
	if pm.IsRunning(server.PID) {
		return fmt.Errorf("进程已在运行中")
	}

	if server.BinaryPath == "" || server.ConfigPath == "" {
		return fmt.Errorf("二进制文件或配置文件路径未设置")
	}

	cmd := exec.Command(server.BinaryPath, "-c", server.ConfigPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动进程失败: %v", err)
	}

	server.PID = cmd.Process.Pid
	server.Status = model.StatusRunning

	pidFile := filepath.Join(filepath.Dir(server.ConfigPath), fmt.Sprintf("frps_%d.pid", server.ID))
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(server.PID)), 0644); err != nil {
		return fmt.Errorf("写入PID文件失败: %v", err)
	}

	return nil
}

func (pm *ProcessManager) Stop(server *model.FrpServer) error {
	if !pm.IsRunning(server.PID) {
		server.Status = model.StatusStopped
		server.PID = 0
		return nil
	}

	process, err := os.FindProcess(server.PID)
	if err != nil {
		return fmt.Errorf("查找进程失败: %v", err)
	}

	if err := process.Kill(); err != nil {
		return fmt.Errorf("停止进程失败: %v", err)
	}

	server.Status = model.StatusStopped
	server.PID = 0

	pidFile := filepath.Join(filepath.Dir(server.ConfigPath), fmt.Sprintf("frps_%d.pid", server.ID))
	os.Remove(pidFile)

	return nil
}

func (pm *ProcessManager) Restart(server *model.FrpServer) error {
	if err := pm.Stop(server); err != nil {
		return err
	}
	return pm.Start(server)
}

func (pm *ProcessManager) IsRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func (pm *ProcessManager) GetStatus(server *model.FrpServer) model.FrpServerStatus {
	if pm.IsRunning(server.PID) {
		return model.StatusRunning
	}
	return model.StatusStopped
}
