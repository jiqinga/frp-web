package service

import (
	"fmt"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/util"
	"runtime"
	"strings"
)

type RemoteFrpsInstaller struct {
	sshClient *SSHClient
	logFunc   func(string)
}

func NewRemoteFrpsInstaller(server *model.FrpServer, logFunc func(string)) (*RemoteFrpsInstaller, error) {
	password, err := util.Decrypt(server.SSHPassword)
	if err != nil {
		return nil, fmt.Errorf("解密密码失败: %w", err)
	}

	sshClient, err := NewSSHClient(SSHConfig{
		Host:     server.SSHHost,
		Port:     server.SSHPort,
		User:     server.SSHUser,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	return &RemoteFrpsInstaller{
		sshClient: sshClient,
		logFunc:   logFunc,
	}, nil
}

func (i *RemoteFrpsInstaller) log(msg string) {
	if i.logFunc != nil {
		i.logFunc(msg)
	}
}

func (i *RemoteFrpsInstaller) Install(server *model.FrpServer) error {
	i.log("开始安装frps...")

	arch, err := i.detectArchitecture()
	if err != nil {
		return fmt.Errorf("检测系统架构失败: %w", err)
	}
	i.log(fmt.Sprintf("检测到系统架构: %s", arch))

	version := server.Version
	if version == "" || version == "latest" {
		version = "0.65.0"
	}

	downloadURL := fmt.Sprintf("https://github.com/fatedier/frp/releases/download/v%s/frp_%s_linux_%s.tar.gz", version, version, arch)

	mirrorService := NewGithubMirrorService()
	downloadURL, err = mirrorService.ConvertGithubURL(downloadURL, server.MirrorID)
	if err != nil {
		return fmt.Errorf("转换下载地址失败: %w", err)
	}

	i.log(fmt.Sprintf("下载地址: %s", downloadURL))

	installPath := server.InstallPath
	if installPath == "" {
		installPath = "/opt/frps"
	}

	i.log(fmt.Sprintf("创建安装目录: %s", installPath))
	if err := i.executeCommand(fmt.Sprintf("sudo mkdir -p %s", installPath)); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	i.log("下载frps...")
	downloadCmd := fmt.Sprintf("cd /tmp && wget --progress=bar:force -O frps.tar.gz %s 2>&1", downloadURL)
	if err := i.executeCommand(downloadCmd); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	i.log("解压文件...")
	if err := i.executeCommand(fmt.Sprintf("cd /tmp && tar -xzf frps.tar.gz")); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	i.log("检查frps服务状态...")
	i.executeCommand("sudo systemctl is-active frps && sudo systemctl stop frps || true")

	i.log("复制文件到安装目录...")
	copyCmd := fmt.Sprintf("cd /tmp && sudo cp frp_%s_linux_%s/frps %s/ && sudo chmod +x %s/frps", version, arch, installPath, installPath)
	if err := i.executeCommand(copyCmd); err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}

	i.log("生成配置文件...")
	if err := i.generateConfig(server, installPath); err != nil {
		return fmt.Errorf("生成配置失败: %w", err)
	}

	i.log(fmt.Sprintf("Dashboard认证信息 - 用户名: %s, 密码: ***", server.DashboardUser))
	i.log(fmt.Sprintf("Dashboard访问地址: http://%s:%d", server.SSHHost, server.DashboardPort))
	i.log("客户端连接Token已配置 (请在服务器配置中查看)")

	i.log("创建systemd服务...")
	if err := i.createSystemdService(server, installPath); err != nil {
		return fmt.Errorf("创建服务失败: %w", err)
	}

	i.log("清理临时文件...")
	i.executeCommand("rm -rf /tmp/frps.tar.gz /tmp/frp_*")

	i.log("安装完成！")
	return nil
}

func (i *RemoteFrpsInstaller) Start(installPath string) error {
	i.log("启动frps服务...")
	if err := i.executeCommand("sudo systemctl start frps"); err != nil {
		return fmt.Errorf("启动失败: %w", err)
	}
	i.log("服务已启动")
	return nil
}

func (i *RemoteFrpsInstaller) Stop(installPath string) error {
	i.log("停止frps服务...")
	if err := i.executeCommand("sudo systemctl stop frps"); err != nil {
		return fmt.Errorf("停止失败: %w", err)
	}
	i.log("服务已停止")
	return nil
}

func (i *RemoteFrpsInstaller) Restart(installPath string) error {
	i.log("重启frps服务...")
	if err := i.executeCommand("sudo systemctl restart frps"); err != nil {
		return fmt.Errorf("重启失败: %w", err)
	}
	i.log("服务已重启")
	return nil
}

func (i *RemoteFrpsInstaller) Uninstall(installPath string) error {
	i.log("卸载frps...")

	i.log("停止服务...")
	i.executeCommand("sudo systemctl stop frps")

	i.log("禁用服务...")
	i.executeCommand("sudo systemctl disable frps")

	i.log("删除服务文件...")
	i.executeCommand("sudo rm -f /etc/systemd/system/frps.service")

	i.log("重载systemd...")
	i.executeCommand("sudo systemctl daemon-reload")

	i.log("删除安装目录...")
	if err := i.executeCommand(fmt.Sprintf("sudo rm -rf %s", installPath)); err != nil {
		return fmt.Errorf("删除目录失败: %w", err)
	}

	i.log("卸载完成！")
	return nil
}

func (i *RemoteFrpsInstaller) GetLogs(installPath string, lines int) (string, error) {
	logPath := fmt.Sprintf("%s/frps.log", installPath)

	// 检查日志文件是否存在
	checkCmd := fmt.Sprintf("test -f %s && echo 'exists' || echo 'not_found'", logPath)
	checkOutput, _ := i.sshClient.ExecuteCommand(checkCmd)

	if strings.TrimSpace(checkOutput) == "not_found" {
		return "", fmt.Errorf("日志文件不存在: %s\n提示: 请确保frps服务已启动并正在运行", logPath)
	}

	// 读取日志文件
	cmd := fmt.Sprintf("sudo tail -n %d %s", lines, logPath)
	output, err := i.sshClient.ExecuteCommand(cmd)
	if err != nil {
		return "", fmt.Errorf("读取日志文件失败: %w", err)
	}
	return output, nil
}

func (i *RemoteFrpsInstaller) GetStatus() (string, error) {
	output, err := i.sshClient.ExecuteCommand("sudo systemctl status frps")
	if err != nil {
		return output, nil
	}
	return output, nil
}

func (i *RemoteFrpsInstaller) GetVersion(installPath string) (string, error) {
	cmd := fmt.Sprintf("%s/frps --version", installPath)
	output, err := i.sshClient.ExecuteCommand(cmd)
	if err != nil {
		return "", fmt.Errorf("获取版本失败: %w", err)
	}
	return strings.TrimSpace(output), nil
}

func (i *RemoteFrpsInstaller) Close() error {
	return i.sshClient.Close()
}

func (i *RemoteFrpsInstaller) detectArchitecture() (string, error) {
	output, err := i.sshClient.ExecuteCommand("uname -m")
	if err != nil {
		return "", err
	}

	arch := strings.TrimSpace(output)
	switch arch {
	case "x86_64":
		return "amd64", nil
	case "aarch64", "arm64":
		return "arm64", nil
	case "armv7l":
		return "arm", nil
	default:
		return "amd64", nil
	}
}

func (i *RemoteFrpsInstaller) executeCommand(cmd string) error {
	return i.sshClient.ExecuteCommandWithCallback(cmd, func(output string) {
		i.log(output)
	})
}

func (i *RemoteFrpsInstaller) generateConfig(server *model.FrpServer, installPath string) error {
	config := fmt.Sprintf(`bindPort: %d
vhostHTTPPort: 80
vhostHTTPSPort: 443
enablePrometheus: true
auth:
  method: token
  token: "%s"
webServer:
  addr: "0.0.0.0"
  port: %d
  user: "%s"
  password: "%s"
log:
  to: "%s/frps.log"
  level: "info"
`, server.BindPort, server.Token, server.DashboardPort, server.DashboardUser, server.DashboardPwd, installPath)

	configPath := fmt.Sprintf("%s/frps.yaml", installPath)
	cmd := fmt.Sprintf("sudo tee %s > /dev/null << 'EOF'\n%s\nEOF", configPath, config)

	return i.executeCommand(cmd)
}

func (i *RemoteFrpsInstaller) createSystemdService(server *model.FrpServer, installPath string) error {
	serviceContent := fmt.Sprintf(`[Unit]
Description=FRP Server Service
After=network.target

[Service]
Type=simple
User=root
Restart=on-failure
RestartSec=5s
ExecStart=%s/frps -c %s/frps.yaml

[Install]
WantedBy=multi-user.target
`, installPath, installPath)

	cmd := fmt.Sprintf("sudo tee /etc/systemd/system/frps.service > /dev/null << 'EOF'\n%s\nEOF", serviceContent)
	if err := i.executeCommand(cmd); err != nil {
		return err
	}

	if err := i.executeCommand("sudo systemctl daemon-reload"); err != nil {
		return err
	}

	return i.executeCommand("sudo systemctl enable frps")
}

func init() {
	_ = runtime.GOOS
}
