package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ConfigSyncResult 配置同步结果
type ConfigSyncResult struct {
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	RolledBack bool   `json:"rolled_back"`
	Timestamp  string `json:"timestamp"`
}

type FrpcManager struct {
	cfg         *Config
	adminClient *FrpcAdminClient
	pidManager  *PIDManager
}

func NewFrpcManager(cfg *Config) *FrpcManager {
	// 创建 Admin API 客户端
	adminClient := NewFrpcAdminClient(
		cfg.FrpcAdminAddr,
		cfg.FrpcAdminPort,
		cfg.FrpcAdminUser,
		cfg.FrpcAdminPassword,
	)

	// 创建 PID 管理器
	pidManager := NewPIDManager(cfg.InstallDir)

	return &FrpcManager{
		cfg:         cfg,
		adminClient: adminClient,
		pidManager:  pidManager,
	}
}

func (m *FrpcManager) ApplyConfig(config string, version int) error {
	log.Printf("[FRPC] ========== 开始应用配置 ==========")
	log.Printf("[FRPC] 配置版本: %d", version)
	log.Printf("[FRPC] 配置文件路径: %s", m.cfg.FrpcConfig)

	// 第一层：语法验证
	log.Printf("[FRPC] 正在验证配置语法...")
	if err := ValidateConfigContent(m.cfg.FrpcPath, config); err != nil {
		log.Printf("[FRPC] ❌ 配置语法验证失败: %v", err)
		return fmt.Errorf("配置验证失败: %v", err)
	}
	log.Printf("[FRPC] ✅ 配置语法验证通过")

	// 备份当前配置
	backupPath, err := BackupConfig(m.cfg.FrpcConfig)
	if err != nil {
		log.Printf("[FRPC] ⚠️ 配置备份失败: %v", err)
	} else if backupPath != "" {
		log.Printf("[FRPC] ✅ 配置备份成功: %s", backupPath)
		defer CleanupBackup(backupPath) // 成功后清理备份
	}

	// 写入新配置
	log.Printf("[FRPC] 正在写入配置文件...")
	if err := os.WriteFile(m.cfg.FrpcConfig, []byte(config), 0644); err != nil {
		log.Printf("[FRPC] ❌ 写入配置失败: %v", err)
		return fmt.Errorf("写入配置失败: %v", err)
	}
	log.Printf("[FRPC] ✅ 配置已写入: %s (version=%d)", m.cfg.FrpcConfig, version)

	// 重载frpc
	log.Printf("[FRPC] 正在重载frpc...")
	if err := m.reloadFrpc(); err != nil {
		log.Printf("[FRPC] ❌ 重载frpc失败: %v", err)
		if backupPath != "" {
			RestoreConfig(backupPath, m.cfg.FrpcConfig)
		}
		return fmt.Errorf("重载frpc失败: %v", err)
	}

	// 第二层：健康检查
	log.Printf("[FRPC] 正在等待 frpc 启动并检查健康状态...")
	if err := m.WaitForHealthy(10 * time.Second); err != nil {
		log.Printf("[FRPC] ❌ 健康检查失败，开始回滚: %v", err)
		if backupPath != "" {
			if restoreErr := RestoreConfig(backupPath, m.cfg.FrpcConfig); restoreErr == nil {
				m.reloadFrpc()
				return fmt.Errorf("frpc 启动失败，已回滚: %v", err)
			}
		}
		return fmt.Errorf("frpc 启动失败: %v", err)
	}

	log.Printf("[FRPC] ✅ 配置已生效 (version=%d)", version)
	log.Printf("[FRPC] ========== 配置应用完成 ==========")
	return nil
}

// ApplyConfigWithResult 应用配置并返回详细结果
func (m *FrpcManager) ApplyConfigWithResult(config string, version int) ConfigSyncResult {
	result := ConfigSyncResult{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	err := m.ApplyConfig(config, version)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		// 检查是否包含回滚信息
		if strings.Contains(result.Error, "已回滚") || strings.Contains(result.Error, "rolled back") {
			result.RolledBack = true
		}
	} else {
		result.Success = true
	}

	return result
}

// WaitForHealthy 等待 frpc 健康状态
func (m *FrpcManager) WaitForHealthy(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	checkInterval := 1 * time.Second

	for time.Now().Before(deadline) {
		if m.IsFrpcAlive() {
			log.Printf("[FRPC] ✅ frpc 健康检查通过")
			return nil
		}
		log.Printf("[FRPC] 等待 frpc 启动...")
		time.Sleep(checkInterval)
	}

	return fmt.Errorf("frpc 在 %v 内未能启动", timeout)
}

func (m *FrpcManager) reloadFrpc() error {
	// Windows 系统：使用进程管理方式
	if runtime.GOOS == "windows" {
		log.Printf("[FRPC] Windows 系统，使用进程管理方式重启 frpc")
		return m.restartFrpc()
	}

	// Linux 系统：必须配置 frpc_service_name
	if m.cfg.FrpcServiceName == "" {
		return fmt.Errorf("Linux 系统必须配置 frpc_service_name 才能重载配置")
	}

	// 注意：frpc 不支持 SIGHUP 信号热重载配置，必须使用 restart
	// 如果使用 systemctl reload（发送 HUP 信号），frpc 会直接退出而不是重新加载配置
	log.Printf("[FRPC] 使用 systemctl restart 重启服务: %s (frpc 不支持热重载)", m.cfg.FrpcServiceName)
	return m.restartFrpcViaSystemctl()
}

// GetProxyStatus 获取代理状态
func (m *FrpcManager) GetProxyStatus() (*AllProxyStatus, error) {
	if m.adminClient == nil {
		return nil, fmt.Errorf("Admin API 客户端未初始化")
	}
	return m.adminClient.GetStatus()
}

// StopProxy 停止指定代理
func (m *FrpcManager) StopProxy(names []string) error {
	if m.adminClient == nil {
		return fmt.Errorf("Admin API 客户端未初始化")
	}
	return m.adminClient.StopProxy(names)
}

// IsAdminAPIAvailable 检查 Admin API 是否可用
func (m *FrpcManager) IsAdminAPIAvailable() bool {
	if m.adminClient == nil {
		return false
	}
	return m.adminClient.IsAvailable()
}

// IsFrpcAlive 检查 frpc 进程是否存活（通过 /healthz 接口）
func (m *FrpcManager) IsFrpcAlive() bool {
	if m.adminClient == nil {
		return false
	}
	return m.adminClient.CheckFrpcAlive()
}

func (m *FrpcManager) restartFrpc() error {
	// Linux 系统：必须使用 systemctl
	if runtime.GOOS != "windows" {
		if m.cfg.FrpcServiceName == "" {
			return fmt.Errorf("Linux 系统必须配置 frpc_service_name 才能重启 frpc")
		}
		return m.restartFrpcViaSystemctl()
	}

	// Windows 系统：使用直接进程管理
	// 杀死旧进程
	if err := m.killFrpc(); err != nil {
		log.Printf("[FRPC] 停止进程失败: %v", err)
	}
	time.Sleep(2 * time.Second)

	// 启动新进程
	cmd := exec.Command(m.cfg.FrpcPath, "-c", m.cfg.FrpcConfig)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动frpc失败: %v", err)
	}

	// 保存 PID
	if err := m.pidManager.SavePID(cmd.Process.Pid); err != nil {
		log.Printf("[FRPC] ⚠️ 保存 PID 失败: %v", err)
	}

	// 后台运行
	go func() {
		cmd.Wait()
		m.pidManager.RemovePID()
	}()

	log.Printf("[FRPC] 进程已启动: PID=%d", cmd.Process.Pid)
	return nil
}

// restartFrpcViaSystemctl 使用 systemctl 重启 frpc 服务
func (m *FrpcManager) restartFrpcViaSystemctl() error {
	serviceName := m.cfg.FrpcServiceName
	log.Printf("[FRPC] 使用 systemctl 重启服务: %s", serviceName)

	cmd := exec.Command("systemctl", "restart", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[FRPC] systemctl restart 失败: %v, 输出: %s", err, string(output))
		return fmt.Errorf("systemctl restart %s 失败: %v", serviceName, err)
	}

	log.Printf("[FRPC] ✅ systemctl restart %s 成功", serviceName)

	// 等待服务启动
	time.Sleep(2 * time.Second)

	// 检查服务状态
	statusCmd := exec.Command("systemctl", "is-active", serviceName)
	statusOutput, _ := statusCmd.CombinedOutput()
	status := string(statusOutput)
	log.Printf("[FRPC] 服务状态: %s", status)

	return nil
}

func (m *FrpcManager) killFrpc() error {
	// Linux 系统：必须使用 systemctl
	if runtime.GOOS != "windows" {
		if m.cfg.FrpcServiceName == "" {
			return fmt.Errorf("Linux 系统必须配置 frpc_service_name 才能停止 frpc")
		}
		return m.stopFrpcViaSystemctl()
	}

	// Windows 系统：优先使用 PID 终止进程
	pid, err := m.pidManager.ReadPID()
	if err != nil {
		log.Printf("[FRPC] 读取 PID 失败: %v，将使用 taskkill /IM", err)
	}

	if pid > 0 {
		// 使用 PID 终止进程
		log.Printf("[FRPC] 使用 PID %d 终止进程", pid)
		cmd := exec.Command("taskkill", "/F", "/PID", fmt.Sprintf("%d", pid))
		log.Printf("[FRPC] 执行停止命令: %v", cmd.Args)
		if err := cmd.Run(); err != nil {
			log.Printf("[FRPC] 使用 PID 终止失败: %v，尝试使用进程名", err)
		} else {
			m.pidManager.RemovePID()
			return nil
		}
	}

	// 回退：使用进程名终止（仅当 PID 方式失败时）
	cmd := exec.Command("taskkill", "/F", "/IM", "frpc.exe")
	log.Printf("[FRPC] 执行停止命令: %v", cmd.Args)
	err = cmd.Run()
	if err != nil {
		log.Printf("[FRPC] 停止命令执行结果: %v (如果没有运行中的frpc进程，这是正常的)", err)
	}
	m.pidManager.RemovePID()
	return nil
}

// stopFrpcViaSystemctl 使用 systemctl 停止 frpc 服务
func (m *FrpcManager) stopFrpcViaSystemctl() error {
	serviceName := m.cfg.FrpcServiceName
	log.Printf("[FRPC] 使用 systemctl 停止服务: %s", serviceName)

	cmd := exec.Command("systemctl", "stop", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[FRPC] systemctl stop 失败: %v, 输出: %s", err, string(output))
		// 即使停止失败也返回 nil，因为目标是确保 frpc 不在运行
		return nil
	}

	log.Printf("[FRPC] ✅ systemctl stop %s 成功", serviceName)
	return nil
}

// Shutdown 停止 frpc 进程（用于响应服务器的停止命令）
func (m *FrpcManager) Shutdown() error {
	log.Printf("[FRPC] ========== 开始停止 frpc ==========")
	if err := m.killFrpc(); err != nil {
		log.Printf("[FRPC] ❌ 停止 frpc 失败: %v", err)
		return err
	}
	log.Printf("[FRPC] ✅ frpc 已停止")
	log.Printf("[FRPC] ========== 停止 frpc 完成 ==========")
	return nil
}

// StartFrpc 启动 frpc 进程
func (m *FrpcManager) StartFrpc() error {
	log.Printf("[FRPC] ========== 开始启动 frpc ==========")

	// Linux 系统：使用 systemctl
	if runtime.GOOS != "windows" {
		if m.cfg.FrpcServiceName == "" {
			return fmt.Errorf("Linux 系统必须配置 frpc_service_name 才能启动 frpc")
		}
		return m.startFrpcViaSystemctl()
	}

	// Windows 系统：直接启动进程
	cmd := exec.Command(m.cfg.FrpcPath, "-c", m.cfg.FrpcConfig)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		log.Printf("[FRPC] ❌ 启动 frpc 失败: %v", err)
		return fmt.Errorf("启动frpc失败: %v", err)
	}

	// 保存 PID
	if err := m.pidManager.SavePID(cmd.Process.Pid); err != nil {
		log.Printf("[FRPC] ⚠️ 保存 PID 失败: %v", err)
	}

	// 后台运行
	go func() {
		cmd.Wait()
		m.pidManager.RemovePID()
	}()

	log.Printf("[FRPC] ✅ frpc 已启动: PID=%d", cmd.Process.Pid)
	log.Printf("[FRPC] ========== 启动 frpc 完成 ==========")
	return nil
}

// startFrpcViaSystemctl 使用 systemctl 启动 frpc 服务
func (m *FrpcManager) startFrpcViaSystemctl() error {
	serviceName := m.cfg.FrpcServiceName
	log.Printf("[FRPC] 使用 systemctl 启动服务: %s", serviceName)

	cmd := exec.Command("systemctl", "start", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[FRPC] systemctl start 失败: %v, 输出: %s", err, string(output))
		return fmt.Errorf("systemctl start %s 失败: %v", serviceName, err)
	}

	log.Printf("[FRPC] ✅ systemctl start %s 成功", serviceName)

	// 等待服务启动
	time.Sleep(2 * time.Second)

	// 检查服务状态
	statusCmd := exec.Command("systemctl", "is-active", serviceName)
	statusOutput, _ := statusCmd.CombinedOutput()
	status := string(statusOutput)
	log.Printf("[FRPC] 服务状态: %s", status)

	return nil
}

// StopFrpc 停止 frpc 进程（对外暴露的方法）
func (m *FrpcManager) StopFrpc() error {
	return m.Shutdown()
}

// RestartFrpc 重启 frpc 进程（对外暴露的方法）
func (m *FrpcManager) RestartFrpc() error {
	log.Printf("[FRPC] ========== 开始重启 frpc ==========")
	if err := m.restartFrpc(); err != nil {
		log.Printf("[FRPC] ❌ 重启 frpc 失败: %v", err)
		return err
	}
	log.Printf("[FRPC] ✅ frpc 已重启")
	log.Printf("[FRPC] ========== 重启 frpc 完成 ==========")
	return nil
}
