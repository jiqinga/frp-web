/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-25 17:01:26
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-26 17:25:00
 * @FilePath            : frp-web-testfrpc-daemon-wsmain.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// BuildTime 编译时间，通过 -ldflags "-X main.BuildTime=xxx" 注入
var BuildTime = "dev"

func main() {
	configPath := flag.String("c", "daemon.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 设置日志 - 同时输出到文件和标准输出
	var logFile *os.File
	if cfg.LogFile != "" {
		var err error
		logFile, err = os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("打开日志文件失败: %v", err)
		} else {
			// 使用 MultiWriter 同时输出到文件和标准输出
			multiWriter := io.MultiWriter(os.Stdout, logFile)
			log.SetOutput(multiWriter)
		}
	}

	log.Println("======================================")
	log.Println("frpc-daemon-ws 启动")
	log.Printf("客户端ID: %d", cfg.ClientID)
	log.Printf("服务器地址: %s", cfg.ServerURL)
	log.Printf("frpc路径: %s", cfg.FrpcPath)
	log.Printf("frpc配置文件: %s", cfg.FrpcConfig)
	log.Printf("frpc Admin API: %s:%d", cfg.FrpcAdminAddr, cfg.FrpcAdminPort)
	if cfg.LogFile != "" {
		log.Printf("日志文件: %s", cfg.LogFile)
	}
	if cfg.InstallDir != "" {
		log.Printf("安装目录: %s", cfg.InstallDir)
	}
	log.Println("======================================")

	// 创建frpc管理器
	frpcMgr := NewFrpcManager(cfg)

	// 检查 Admin API 是否可用
	if frpcMgr.IsAdminAPIAvailable() {
		log.Println("[主程序] ✅ frpc Admin API 可用，将使用热重载方式更新配置")
	} else {
		log.Println("[主程序] ⚠️ frpc Admin API 不可用，将使用进程重启方式更新配置")
	}

	// 用于通知主程序退出的通道
	shutdownChan := make(chan struct{})

	// 创建WebSocket客户端
	var wsClient *WSClient
	wsClient = NewWSClient(cfg, func(config string, version int) {
		log.Printf("[主程序] 收到配置更新: version=%d", version)

		// 应用配置并获取详细结果
		result := frpcMgr.ApplyConfigWithResult(config, version)

		// 发送详细的配置同步结果
		wsClient.SendConfigSyncResult(result)

		// 同时发送旧格式的同步结果以保持兼容性
		if result.Success {
			wsClient.SendSyncResult(true, version, "配置已成功应用")
			log.Printf("[主程序] 配置同步成功: version=%d", version)
		} else {
			wsClient.SendSyncResult(false, version, result.Error)
			log.Printf("[主程序] 应用配置失败: %v", result.Error)
		}
	})

	// 设置停止命令回调
	wsClient.SetShutdownCallback(func() {
		log.Println("[主程序] ========== 收到服务器停止命令 ==========")

		// 先停止 frpc
		log.Println("[主程序] 正在停止 frpc...")
		if err := frpcMgr.Shutdown(); err != nil {
			log.Printf("[主程序] ⚠️ 停止 frpc 失败: %v", err)
		} else {
			log.Println("[主程序] ✅ frpc 已停止")
		}

		// 通知主程序退出
		log.Println("[主程序] 准备退出 daemon...")
		close(shutdownChan)
	})

	// 创建更新器
	updater := NewUpdater(cfg, frpcMgr, wsClient)
	updater.Start()

	// 设置更新命令回调
	wsClient.SetUpdateCallback(func(updateType string, version string, downloadURL string, mirrorID uint) {
		log.Printf("[主程序] 收到更新命令: type=%s, version=%s", updateType, version)
		updater.HandleUpdate(updateType, version, downloadURL, mirrorID)
	})

	// 设置证书同步回调
	wsClient.SetCertSyncCallback(func(domain string, certPEM string, keyPEM string) {
		log.Printf("[主程序] 收到证书同步: domain=%s", domain)
		if err := saveCertificate(cfg, domain, certPEM, keyPEM); err != nil {
			log.Printf("[主程序] ❌ 保存证书失败: %v", err)
		} else {
			log.Printf("[主程序] ✅ 证书保存成功: domain=%s", domain)
		}
	})

	// 设置证书删除回调
	wsClient.SetCertDeleteCallback(func(domain string) {
		log.Printf("[主程序] 收到证书删除: domain=%s", domain)
		if err := deleteCertificate(cfg, domain); err != nil {
			log.Printf("[主程序] ❌ 删除证书失败: %v", err)
		} else {
			log.Printf("[主程序] ✅ 证书删除成功: domain=%s", domain)
		}
	})

	// 创建日志流管理器
	logStreamMgr := NewLogStreamManager(cfg, func(logType LogType, line string) {
		wsClient.SendLogData(string(logType), line)
	})

	// 设置日志流命令回调
	wsClient.SetLogStreamCallback(func(logType string, action string, lines int) {
		log.Printf("[主程序] 收到日志流命令: type=%s, action=%s, lines=%d", logType, action, lines)
		switch action {
		case "start":
			if err := logStreamMgr.StartStream(LogType(logType), lines); err != nil {
				log.Printf("[主程序] ❌ 启动日志流失败: %v", err)
			} else {
				log.Printf("[主程序] ✅ 日志流已启动: type=%s", logType)
			}
		case "stop":
			logStreamMgr.StopStream(LogType(logType))
			log.Printf("[主程序] ✅ 日志流已停止: type=%s", logType)
		}
	})

	// 设置frpc控制命令回调
	wsClient.SetFrpcControlCallback(func(action string) {
		log.Printf("[主程序] 收到frpc控制命令: action=%s", action)
		var err error
		var message string
		switch action {
		case "start":
			err = frpcMgr.StartFrpc()
			message = "frpc启动成功"
		case "stop":
			err = frpcMgr.StopFrpc()
			message = "frpc停止成功"
		case "restart":
			err = frpcMgr.RestartFrpc()
			message = "frpc重启成功"
		default:
			err = fmt.Errorf("未知的控制操作: %s", action)
		}
		if err != nil {
			log.Printf("[主程序] ❌ frpc控制失败: %v", err)
			wsClient.SendFrpcControlResult(action, false, err.Error())
		} else {
			log.Printf("[主程序] ✅ %s", message)
			wsClient.SendFrpcControlResult(action, true, message)
		}
	})

	// 启动WebSocket客户端
	go wsClient.Run()

	// 连接成功后上报版本信息
	go func() {
		// 等待连接建立
		time.Sleep(3 * time.Second)
		reportVersionInfo(wsClient, cfg)
	}()

	// 启动 frpc 健康检查协程
	go startFrpcHealthChecker(wsClient, frpcMgr, cfg)

	// 等待中断信号或停止命令
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Println("收到系统退出信号，正在关闭...")
	case <-shutdownChan:
		log.Println("收到服务器停止命令，正在关闭...")
	}

	// 停止更新器
	updater.Stop()

	// 停止日志流
	logStreamMgr.StopAll()
	wsClient.Close()

	// 显式关闭日志文件（os.Exit 不会执行 defer）
	if logFile != nil {
		logFile.Close()
	}

	log.Println("frpc-daemon-ws 已停止")
}

// reportVersionInfo 上报版本信息
func reportVersionInfo(wsClient *WSClient, cfg *Config) {
	// 获取 frpc 版本
	frpcVersion := getFrpcVersion(cfg.FrpcPath)

	// daemon 版本（编译时注入的编译时间）
	daemonVersion := BuildTime

	// 获取操作系统和架构
	osName := runtime.GOOS
	arch := runtime.GOARCH

	log.Printf("[主程序] 上报版本信息: frpc=%s, daemon=%s, os=%s, arch=%s", frpcVersion, daemonVersion, osName, arch)
	wsClient.SendVersionReport(frpcVersion, daemonVersion, osName, arch)
}

// getFrpcVersion 获取 frpc 版本
func getFrpcVersion(frpcPath string) string {
	cmd := exec.Command(frpcPath, "-v")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[主程序] 获取 frpc 版本失败: %v", err)
		return "unknown"
	}
	version := strings.TrimSpace(string(output))
	// frpc 输出格式可能是 "frpc version 0.52.0" 或 "0.52.0"
	if strings.Contains(version, " ") {
		parts := strings.Fields(version)
		version = parts[len(parts)-1]
	}
	return version
}

// saveCertificate 保存证书到本地文件
func saveCertificate(cfg *Config, domain string, certPEM string, keyPEM string) error {
	// 验证 domain 不包含路径分隔符，防止路径遍历攻击
	if strings.ContainsAny(domain, "/\\") {
		return fmt.Errorf("无效的域名: 包含路径分隔符")
	}

	// 使用配置的安装目录，如果未配置则使用默认目录
	var certDir string
	if cfg.InstallDir != "" {
		certDir = filepath.Join(cfg.InstallDir, "certs")
	} else {
		// 跨平台默认目录
		if runtime.GOOS == "windows" {
			certDir = filepath.Join(os.Getenv("ProgramData"), "frpc", "certs")
		} else {
			certDir = "/opt/frpc/certs"
		}
	}

	// 创建证书目录
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("创建证书目录失败: %v", err)
	}

	// 保存证书文件
	certPath := filepath.Join(certDir, domain+".crt")
	if err := os.WriteFile(certPath, []byte(certPEM), 0644); err != nil {
		return fmt.Errorf("保存证书文件失败: %v", err)
	}
	log.Printf("[证书同步] 证书已保存: %s", certPath)

	// 保存私钥文件
	keyPath := filepath.Join(certDir, domain+".key")
	if err := os.WriteFile(keyPath, []byte(keyPEM), 0600); err != nil {
		return fmt.Errorf("保存私钥文件失败: %v", err)
	}
	log.Printf("[证书同步] 私钥已保存: %s", keyPath)

	return nil
}

// deleteCertificate 删除本地证书文件
func deleteCertificate(cfg *Config, domain string) error {
	// 验证 domain 不包含路径分隔符，防止路径遍历攻击
	if strings.ContainsAny(domain, "/\\") {
		return fmt.Errorf("无效的域名: 包含路径分隔符")
	}

	// 使用配置的安装目录，如果未配置则使用默认目录
	var certDir string
	if cfg.InstallDir != "" {
		certDir = filepath.Join(cfg.InstallDir, "certs")
	} else {
		// 跨平台默认目录
		if runtime.GOOS == "windows" {
			certDir = filepath.Join(os.Getenv("ProgramData"), "frpc", "certs")
		} else {
			certDir = "/opt/frpc/certs"
		}
	}

	// 删除证书文件
	certPath := filepath.Join(certDir, domain+".crt")
	if err := os.Remove(certPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除证书文件失败: %v", err)
	}
	log.Printf("[证书删除] 证书已删除: %s", certPath)

	// 删除私钥文件
	keyPath := filepath.Join(certDir, domain+".key")
	if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除私钥文件失败: %v", err)
	}
	log.Printf("[证书删除] 私钥已删除: %s", keyPath)

	return nil
}

// startFrpcHealthChecker 启动 frpc 健康检查协程
// 定期检查 frpc 进程是否存活，并上报给服务器
func startFrpcHealthChecker(wsClient *WSClient, frpcMgr *FrpcManager, cfg *Config) {
	// 等待 WebSocket 连接建立
	time.Sleep(5 * time.Second)

	// 健康检查间隔（与心跳间隔一致）
	interval := time.Duration(cfg.HeartbeatSec) * time.Second
	if interval < 10*time.Second {
		interval = 10 * time.Second // 最小间隔 10 秒
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("[健康检查] 启动 frpc 健康检查，间隔: %v", interval)

	// 记录上一次的状态，避免重复上报相同状态
	var lastAlive *bool

	for {
		select {
		case <-ticker.C:
			alive := frpcMgr.IsFrpcAlive()

			// 只在状态变化时上报，或者首次检查时上报
			if lastAlive == nil || *lastAlive != alive {
				if alive {
					log.Printf("[健康检查] ✅ frpc 运行正常")
				} else {
					log.Printf("[健康检查] ❌ frpc 未运行或不可达")
				}
				wsClient.SendFrpcHealthStatus(alive)
				lastAlive = &alive
			}
		}
	}
}
