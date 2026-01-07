package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"frp-web-panel/internal/logger"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"frp-web-panel/pkg/database"
	"time"
)

type ClientRegisterService struct {
	tokenRepo        *repository.ClientRegisterTokenRepository
	clientRepo       *repository.ClientRepository
	githubMirrorRepo *repository.GithubMirrorRepository
	settingRepo      *repository.SettingRepository
	frpServerRepo    *repository.FrpServerRepository
}

func NewClientRegisterService() *ClientRegisterService {
	return &ClientRegisterService{
		tokenRepo:        repository.NewClientRegisterTokenRepository(),
		clientRepo:       repository.NewClientRepository(),
		githubMirrorRepo: repository.NewGithubMirrorRepository(),
		settingRepo:      repository.NewSettingRepository(),
		frpServerRepo:    repository.NewFrpServerRepository(database.DB),
	}
}

// getPublicURL 获取公网访问地址,优先从设置中读取,否则使用默认值
func (s *ClientRegisterService) getPublicURL() string {
	publicURL, err := s.settingRepo.GetOrCreate("public_url", "http://localhost:8080", "公网访问地址(用于生成客户端注册脚本)")
	if err != nil {
		return "http://localhost:8080"
	}
	return publicURL
}

// GenerateToken 生成注册Token
func (s *ClientRegisterService) GenerateToken(req *model.ClientRegisterToken, userID uint) (*model.ClientRegisterToken, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return nil, err
	}

	// 生成16位强随机admin密码
	adminPwd := make([]byte, 12)
	if _, err := rand.Read(adminPwd); err != nil {
		return nil, err
	}

	req.Token = hex.EncodeToString(token)
	req.AdminPassword = hex.EncodeToString(adminPwd)[:16]
	req.CreatedBy = userID
	req.ExpiresAt = time.Now().Add(24 * time.Hour) // 24小时有效期
	req.Used = false

	if err := s.tokenRepo.Create(req); err != nil {
		return nil, err
	}

	return req, nil
}

// GenerateScript 生成curl命令
func (s *ClientRegisterService) GenerateScript(token string, scriptType string, mirrorID uint) (string, error) {
	t, err := s.tokenRepo.FindByToken(token)
	if err != nil {
		return "", errors.New("token不存在")
	}

	if t.Used {
		return "", errors.New("token已被使用")
	}

	if time.Now().After(t.ExpiresAt) {
		return "", errors.New("token已过期")
	}

	publicURL := s.getPublicURL()
	installURL := fmt.Sprintf("%s/install/%s?type=%s&mirror=%d", publicURL, token, scriptType, mirrorID)

	if scriptType == "powershell" {
		return fmt.Sprintf("Invoke-WebRequest -Uri '%s' -UseBasicParsing | Invoke-Expression", installURL), nil
	}
	return fmt.Sprintf("bash <(curl -fsSL '%s')", installURL), nil
}

// GetInstallScript 获取安装脚本内容
func (s *ClientRegisterService) GetInstallScript(token string, scriptType string, mirrorID uint) (string, error) {
	logger.Debug("[GetInstallScript] ========================================")
	logger.Debugf("[GetInstallScript] 请求参数 - Token: %s, Type: %s, MirrorID: %d", token, scriptType, mirrorID)

	t, err := s.tokenRepo.FindByToken(token)
	if err != nil {
		logger.Debugf("[GetInstallScript] ❌ Token查找失败: %v", err)
		return "", errors.New("token不存在")
	}
	logger.Debugf("[GetInstallScript] ✅ Token找到 - ID: %d, ClientName: %s, Used: %v, ExpiresAt: %v", t.ID, t.ClientName, t.Used, t.ExpiresAt)

	if t.Used {
		logger.Debug("[GetInstallScript] ❌ Token已被使用")
		return "", errors.New("token已被使用")
	}

	if time.Now().After(t.ExpiresAt) {
		logger.Debugf("[GetInstallScript] ❌ Token已过期 (当前时间: %v, 过期时间: %v)", time.Now(), t.ExpiresAt)
		return "", errors.New("token已过期")
	}

	logger.Debugf("[GetInstallScript] 查找镜像源 ID: %d", mirrorID)
	mirror, err := s.githubMirrorRepo.GetByID(mirrorID)
	if err != nil {
		logger.Debugf("[GetInstallScript] ❌ 镜像源查找失败: %v", err)
		return "", errors.New("镜像源不存在")
	}
	logger.Debugf("[GetInstallScript] ✅ 镜像源找到 - Name: %s, BaseURL: %s", mirror.Name, mirror.BaseURL)

	logger.Debugf("[GetInstallScript] 查找FRP服务器 ID: %d", t.FrpServerID)
	frpServer, err := s.frpServerRepo.GetByID(t.FrpServerID)
	if err != nil {
		logger.Debugf("[GetInstallScript] ❌ FRP服务器查找失败: %v", err)
		return "", errors.New("FRP服务器不存在")
	}
	logger.Debugf("[GetInstallScript] ✅ FRP服务器找到 - Name: %s, Version: %s", frpServer.Name, frpServer.Version)

	version := frpServer.Version
	if version == "" {
		version = "0.65.0"
	}

	apiURL := s.getPublicURL()

	if scriptType == "powershell" {
		return s.generatePowerShellScript(t, apiURL, mirror.BaseURL, version), nil
	}
	return s.generateBashScript(t, apiURL, mirror.BaseURL, version), nil
}

func (s *ClientRegisterService) generateBashScript(t *model.ClientRegisterToken, apiURL, baseURL, version string) string {
	downloadURL := baseURL + "/fatedier/frp/releases/download"
	wsURL := apiURL
	if len(wsURL) > 7 && wsURL[:7] == "http://" {
		wsURL = "ws://" + wsURL[7:]
	} else if len(wsURL) > 8 && wsURL[:8] == "https://" {
		wsURL = "wss://" + wsURL[8:]
	}

	return fmt.Sprintf(`#!/bin/bash
set -e

echo "╔════════════════════════════════════════════════════════════╗"
echo "║         FRP 客户端自动安装脚本 v2.0                      ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "📋 配置信息:"
echo "   客户端名称: %s"
echo "   服务器地址: %s:%d"
echo "   FRP 版本: %s"
echo ""

# 检测系统架构
echo "🔍 检测系统架构.."
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    armv7l) ARCH="arm" ;;
    *) echo "❌不支持的架构: $ARCH"; exit 1 ;;
esac
echo "✅系统架构: $ARCH"
echo ""

# 下载frpc
VERSION="%s"
DOWNLOAD_URL="%s/v${VERSION}/frp_${VERSION}_linux_${ARCH}.tar.gz"
echo "📦 [1/6] 下载 FRP 客户端.."
echo "   下载地址: $DOWNLOAD_URL"
wget -q --show-progress -O frp.tar.gz "$DOWNLOAD_URL" || {
    echo "❌下载失败，请检查网络连接"
    exit 1
}
echo "✅下载完成"
echo ""

echo "📂 解压文件..."
tar -xzf frp.tar.gz
echo "✅解压完成"
echo ""

# 安装frpc
INSTALL_DIR="/opt/frpc"
echo "⚙️  [2/6] 安装 frpc 到 ${INSTALL_DIR}..."
# 停止 systemd 服务
if sudo systemctl is-active --quiet frpc 2>/dev/null; then
    echo "   检测到 frpc 服务正在运行，正在停止.."
    sudo systemctl stop frpc
    sleep 2
fi
# 确保所有 frpc 进程都已停止（处理非 systemd 启动的情况）
if pgrep -x frpc > /dev/null 2>&1; then
    echo "   检测到 frpc 进程仍在运行，正在强制停止.."
    sudo pkill -x frpc || true
    sleep 2
fi
sudo mkdir -p ${INSTALL_DIR}
sudo cp frp_*_linux_${ARCH}/frpc ${INSTALL_DIR}/
sudo chown root:root ${INSTALL_DIR}/frpc
sudo chmod 755 ${INSTALL_DIR}/frpc
echo "✅安装完成"
echo ""

# 生成frpc配置 (TOML格式)
echo "📝 [3/6] 生成frpc配置..."
sudo tee ${INSTALL_DIR}/frpc.toml > /dev/null << 'EOF'
serverAddr = "%s"
serverPort = %d
user = "%s"

auth.token = "%s"

log.to = "/opt/frpc/frpc.log"
log.level = "info"
log.maxDays = 7

webServer.addr = "127.0.0.1"
webServer.port = 7400
webServer.user = "admin"
webServer.password = "%s"

EOF
sudo chown root:root ${INSTALL_DIR}/frpc.toml
sudo chmod 644 ${INSTALL_DIR}/frpc.toml
echo "✅配置完成"
echo ""

# 创建frpc服务
echo "🔧 [4/6] 配置 frpc systemd 服务..."
sudo tee /etc/systemd/system/frpc.service > /dev/null << EOF
[Unit]
Description=FRP Client Service
After=network.target

[Service]
Type=simple
User=root
Restart=on-failure
RestartSec=5s
ExecStart=${INSTALL_DIR}/frpc -c ${INSTALL_DIR}/frpc.toml
ExecReload=/bin/kill -HUP \$MAINPID
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF
sudo systemctl daemon-reload
sudo systemctl enable frpc > /dev/null 2>&1
sudo systemctl start frpc
sleep 2
if sudo systemctl is-active --quiet frpc; then
    echo "✅frpc服务启动成功"
else
    echo "⚠️  frpc服务启动失败"
fi
echo ""

# 注册客户端(初始状态为offline)
echo "📝 [5/6] 注册客户端到管理平台..."
echo "   注册后初始状态为离线,等待守护程序连接..."
REGISTER_RESPONSE=$(curl -s -X POST "%s/api/clients/register" \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"%s\"}")
CLIENT_ID=$(echo "$REGISTER_RESPONSE" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
if [ -n "$CLIENT_ID" ]; then
    echo "✅注册成功 (ClientID: $CLIENT_ID, 状态: offline)"
else
    echo "⚠️  注册失败，跳过守护程序安装"
    rm -rf frp.tar.gz frp_*_linux_${ARCH}
    exit 0
fi
echo ""

# 下载并安装守护程序
DAEMON_DIR="/opt/frpc-daemon"
echo "🔧 [6/6] 安装配置同步守护程序..."
echo "   守护程序连接成功后,客户端状态将自动更新为在线..."
if sudo systemctl is-active --quiet frpc-daemon 2>/dev/null; then
    echo "   检测到守护程序正在运行,正在停止.."
    sudo systemctl stop frpc-daemon
fi
echo "   下载守护程序..."
DAEMON_URL="%s/download/daemon/linux/${ARCH}"
sudo mkdir -p ${DAEMON_DIR}
sudo wget -q -O ${DAEMON_DIR}/frpc-daemon-ws "$DAEMON_URL" || {
    echo "⚠️  守护程序下载失败，跳过"
    rm -rf frp.tar.gz frp_*_linux_${ARCH}
    exit 0
}
sudo chmod +x ${DAEMON_DIR}/frpc-daemon-ws

# 生成守护程序配置
sudo tee ${DAEMON_DIR}/daemon.yaml > /dev/null << EOF
client_id: ${CLIENT_ID}
token: "%s"
server_url: "%s"
frpc_path: "${INSTALL_DIR}/frpc"
frpc_config: "${INSTALL_DIR}/frpc.toml"
frpc_admin_port: 7400
frpc_admin_user: "admin"
frpc_admin_password: "%s"
frpc_service_name: "frpc"
daemon_service_name: "frpc-daemon"
log_file: "${DAEMON_DIR}/frpc-daemon.log"
heartbeat_sec: 30
EOF

# 创建守护程序服务
sudo tee /etc/systemd/system/frpc-daemon.service > /dev/null << EOF
[Unit]
Description=frpc Daemon WebSocket Service
After=network.target frpc.service

[Service]
Type=simple
User=root
WorkingDirectory=${DAEMON_DIR}
ExecStart=${DAEMON_DIR}/frpc-daemon-ws -c ${DAEMON_DIR}/daemon.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable frpc-daemon > /dev/null 2>&1
sudo systemctl start frpc-daemon
sleep 2
if sudo systemctl is-active --quiet frpc-daemon; then
    echo "✅守护程序启动成功,正在连接服务器..."
    echo "   客户端状态将在连接成功后自动更新为在线"
else
    echo "⚠️  守护程序启动失败,客户端保持离线状态"
fi
echo ""

# 清理
rm -rf frp.tar.gz frp_*_linux_${ARCH}

echo "╔════════════════════════════════════════════════════════════╗"
echo "║                   安装完成                                ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "📁 安装信息:"
echo "   frpc目录: ${INSTALL_DIR}"
echo "   守护程序: ${DAEMON_DIR}"
echo "   配置文件: ${INSTALL_DIR}/frpc.toml"
echo ""
echo "🔧 常用命令:"
echo "   frpc状态: sudo systemctl status frpc"
echo "   守护程序: sudo systemctl status frpc-daemon"
echo "   查看日志: sudo journalctl -u frpc-daemon -f"
echo ""
`, t.ClientName, t.ServerAddr, t.ServerPort, version, version, downloadURL,
		t.ServerAddr, t.ServerPort, t.ClientName, t.TokenStr, t.AdminPassword,
		apiURL, t.Token, apiURL, t.TokenStr, wsURL, t.AdminPassword)
}

func (s *ClientRegisterService) generatePowerShellScript(t *model.ClientRegisterToken, apiURL, baseURL, version string) string {
	downloadURL := baseURL + "/fatedier/frp/releases/download"
	wsURL := apiURL
	if len(wsURL) > 7 && wsURL[:7] == "http://" {
		wsURL = "ws://" + wsURL[7:]
	} else if len(wsURL) > 8 && wsURL[:8] == "https://" {
		wsURL = "wss://" + wsURL[8:]
	}

	// 使用 %%s 来转义 %s 在PowerShell中的特殊含义
	script := "# FRP 客户端自动安装脚本 v2.0\n" +
		"$ErrorActionPreference = \"Stop\"\n\n" +
		"Write-Host \"╔════════════════════════════════════════════════════════════╗\" -ForegroundColor Cyan\n" +
		"Write-Host \"║         FRP 客户端自动安装脚本 v2.0                      ║\" -ForegroundColor Cyan\n" +
		"Write-Host \"╚════════════════════════════════════════════════════════════╝\" -ForegroundColor Cyan\n" +
		"Write-Host \"\"\n" +
		fmt.Sprintf("Write-Host \"📋 配置信息:\" -ForegroundColor White\n") +
		fmt.Sprintf("Write-Host \"   客户端名称: %s\" -ForegroundColor Gray\n", t.ClientName) +
		fmt.Sprintf("Write-Host \"   服务器地址: %s:%d\" -ForegroundColor Gray\n", t.ServerAddr, t.ServerPort) +
		fmt.Sprintf("Write-Host \"   FRP 版本: %s\" -ForegroundColor Gray\n", version) +
		"Write-Host \"\"\n\n" +
		fmt.Sprintf("$VERSION = \"%s\"\n", version) +
		"$ARCH = if ([Environment]::Is64BitOperatingSystem) { \"amd64\" } else { \"386\" }\n" +
		fmt.Sprintf("$DOWNLOAD_URL = \"%s/v$VERSION/frp_${VERSION}_windows_$ARCH.zip\"\n", downloadURL) +
		"$INSTALL_DIR = \"$env:ProgramFiles\\frpc\"\n" +
		"$DAEMON_DIR = \"$env:ProgramFiles\\frpc-daemon\"\n\n" +
		"Write-Host \"🔍 检测系统架构..\" -ForegroundColor Yellow\n" +
		"Write-Host \"✅ 系统架构: $ARCH\" -ForegroundColor Green\n" +
		"Write-Host \"\"\n\n" +
		"Write-Host \"📦 [1/6] 下载 FRP 客户端..\" -ForegroundColor Yellow\n" +
		"try {\n" +
		"    Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile \"frp.zip\" -UseBasicParsing\n" +
		"    Write-Host \"✅下载完成\" -ForegroundColor Green\n" +
		"} catch {\n" +
		"    Write-Host \"❌下载失败: $_\" -ForegroundColor Red\n" +
		"    exit 1\n" +
		"}\n" +
		"Write-Host \"\"\n\n" +
		"Write-Host \"📂 解压文件...\" -ForegroundColor Yellow\n" +
		"Expand-Archive -Path \"frp.zip\" -DestinationPath \".\" -Force\n" +
		"$FrpDir = Get-ChildItem -Directory -Filter \"frp_*_windows_$ARCH\" | Select-Object -First 1\n" +
		"Write-Host \"✅解压完成\" -ForegroundColor Green\n" +

		"Write-Host \"\"\n\n" +
		"Write-Host \"⚙️  [2/6] 安装 frpc...\" -ForegroundColor Yellow\n" +
		"if (-not (Test-Path $INSTALL_DIR)) {\n" +
		"    New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null\n" +
		"}\n" +
		"Copy-Item -Path \"$($FrpDir.FullName)\\frpc.exe\" -Destination $INSTALL_DIR -Force\n" +
		"Write-Host \"✅安装完成\" -ForegroundColor Green\n" +
		"Write-Host \"\"\n\n" +
		"Write-Host \"📝 [3/6] 生成frpc配置...\" -ForegroundColor Yellow\n" +
		"$configContent = @\"\n" +
		fmt.Sprintf("serverAddr = \"%s\"\n", t.ServerAddr) +
		fmt.Sprintf("serverPort = %d\n", t.ServerPort) +
		fmt.Sprintf("user = \"%s\"\n\n", t.ClientName) +
		fmt.Sprintf("auth.token = \"%s\"\n\n", t.TokenStr) +
		"log.to = \"$INSTALL_DIR\\frpc.log\"\n" +
		"log.level = \"info\"\n" +
		"log.maxDays = 7\n\n" +
		"webServer.addr = \"127.0.0.1\"\n" +
		"webServer.port = 7400\n" +
		"webServer.user = \"admin\"\n" +
		fmt.Sprintf("webServer.password = \"%s\"\n\n", t.AdminPassword) +
		"\"@\n" +
		"$configContent | Out-File -FilePath \"$INSTALL_DIR\\frpc.toml\" -Encoding UTF8\n" +
		"Write-Host \"✅配置完成\" -ForegroundColor Green\n" +
		"Write-Host \"\"\n\n" +
		"Write-Host \"🔧 [4/6] 配置frpc服务...\" -ForegroundColor Yellow\n" +
		"$serviceName = \"frpc\"\n" +
		"$serviceExists = Get-Service -Name $serviceName -ErrorAction SilentlyContinue\n" +
		"if ($serviceExists) {\n" +
		"    Stop-Service -Name $serviceName -Force -ErrorAction SilentlyContinue\n" +
		"    sc.exe delete $serviceName | Out-Null\n" +
		"    Start-Sleep -Seconds 2\n" +
		"}\n" +
		"$binaryPath = \"\\\"`\"$INSTALL_DIR\\frpc.exe`\\\" -c `\\\"$INSTALL_DIR\\frpc.toml`\\\"\"\n" +
		"sc.exe create $serviceName binPath= $binaryPath start= auto DisplayName= \"FRP Client Service\" | Out-Null\n" +
		"sc.exe failure $serviceName reset= 86400 actions= restart/5000/restart/10000/restart/30000 | Out-Null\n" +
		"Start-Service -Name $serviceName\n" +
		"Start-Sleep -Seconds 2\n" +
		"if ((Get-Service -Name $serviceName).Status -eq \"Running\") {\n" +
		"    Write-Host \"✅frpc服务启动成功\" -ForegroundColor Green\n" +
		"} else {\n" +
		"    Write-Host \"⚠️  frpc服务启动失败\" -ForegroundColor Yellow\n" +
		"}\n" +
		"Write-Host \"\"\n\n" +
		"Write-Host \"📝 [5/6] 注册客户端...\" -ForegroundColor Yellow\n" +
		"try {\n" +
		fmt.Sprintf("    $body = @{token=\"%s\"} | ConvertTo-Json\n", t.Token) +
		fmt.Sprintf("    $response = Invoke-RestMethod -Uri \"%s/api/clients/register\" -Method Post -Body $body -ContentType \"application/json\"\n", apiURL) +
		"    $CLIENT_ID = $response.data.id\n" +
		"    Write-Host \"✅注册成功 (ClientID: $CLIENT_ID)\" -ForegroundColor Green\n" +
		"} catch {\n" +
		"    Write-Host \"⚠️  注册失败，跳过守护程序\" -ForegroundColor Yellow\n" +
		"    Remove-Item -Path \"frp.zip\" -Force -ErrorAction SilentlyContinue\n" +
		"    Remove-Item -Path $FrpDir.FullName -Recurse -Force -ErrorAction SilentlyContinue\n" +
		"    exit 0\n" +
		"}\n" +
		"Write-Host \"\"\n\n" +
		"Write-Host \"🔧 [6/6] 安装守护程序...\" -ForegroundColor Yellow\n" +
		"try {\n" +
		"    $daemonService = \"frpc-daemon\"\n" +
		"    $daemonExists = Get-Service -Name $daemonService -ErrorAction SilentlyContinue\n" +
		"    if ($daemonExists -and $daemonExists.Status -eq \"Running\") {\n" +
		"        Write-Host \"   检测到守护程序正在运行,正在停止..\" -ForegroundColor Yellow\n" +
		"        Stop-Service -Name $daemonService -Force -ErrorAction SilentlyContinue\n" +
		"        Start-Sleep -Seconds 2\n" +
		"    }\n" +
		fmt.Sprintf("    $DAEMON_URL = \"%s/download/daemon/windows/$ARCH\"\n", apiURL) +
		"    if (-not (Test-Path $DAEMON_DIR)) {\n" +
		"        New-Item -ItemType Directory -Path $DAEMON_DIR -Force | Out-Null\n" +
		"    }\n" +
		"    Invoke-WebRequest -Uri $DAEMON_URL -OutFile \"$DAEMON_DIR\\frpc-daemon-ws.exe\" -UseBasicParsing\n" +
		"    \n" +
		"    $daemonConfig = @\"\n" +
		"client_id: $CLIENT_ID\n" +
		fmt.Sprintf("token: \"%s\"\n", t.TokenStr) +
		fmt.Sprintf("server_url: \"%s\"\n", wsURL) +
		"frpc_path: \"$INSTALL_DIR\\frpc.exe\"\n" +
		"frpc_config: \"$INSTALL_DIR\\frpc.toml\"\n" +
		"frpc_admin_port: 7400\n" +
		"frpc_admin_user: \"admin\"\n" +
		fmt.Sprintf("frpc_admin_password: \"%s\"\n", t.AdminPassword) +
		"frpc_service_name: \"frpc\"\n" +
		"daemon_service_name: \"frpc-daemon\"\n" +
		"log_file: \"$DAEMON_DIR\\frpc-daemon.log\"\n" +
		"heartbeat_sec: 30\n" +
		"\"@\n" +
		"    $daemonConfig | Out-File -FilePath \"$DAEMON_DIR\\daemon.yaml\" -Encoding UTF8\n" +
		"    \n" +
		"    if ($daemonExists) {\n" +
		"        sc.exe delete $daemonService | Out-Null\n" +
		"        Start-Sleep -Seconds 2\n" +
		"    }\n" +
		"    $daemonBinary = \"\\\"`\"$DAEMON_DIR\\frpc-daemon-ws.exe`\\\" -c `\\\"$DAEMON_DIR\\daemon.yaml`\\\"\"\n" +
		"    sc.exe create $daemonService binPath= $daemonBinary start= auto DisplayName= \"frpc Daemon Service\" | Out-Null\n" +
		"    Start-Service -Name $daemonService\n" +
		"    Start-Sleep -Seconds 2\n" +
		"    if ((Get-Service -Name $daemonService).Status -eq \"Running\") {\n" +
		"        Write-Host \"✅守护程序启动成功\" -ForegroundColor Green\n" +
		"    } else {\n" +
		"        Write-Host \"⚠️  守护程序启动失败\" -ForegroundColor Yellow\n" +
		"    }\n" +
		"} catch {\n" +
		"    Write-Host \"⚠️  守护程序安装失败: $_\" -ForegroundColor Yellow\n" +
		"}\n" +
		"Write-Host \"\"\n\n" +
		"Remove-Item -Path \"frp.zip\" -Force -ErrorAction SilentlyContinue\n" +
		"Remove-Item -Path $FrpDir.FullName -Recurse -Force -ErrorAction SilentlyContinue\n\n" +
		"Write-Host \"╔════════════════════════════════════════════════════════════╗\" -ForegroundColor Cyan\n" +
		"Write-Host \"║                   安装完成                                ║\" -ForegroundColor Cyan\n" +
		"Write-Host \"╚════════════════════════════════════════════════════════════╝\" -ForegroundColor Cyan\n" +
		"Write-Host \"\"\n" +
		"Write-Host \"📁 安装信息:\" -ForegroundColor White\n" +
		"Write-Host \"   frpc目录: $INSTALL_DIR\" -ForegroundColor Gray\n" +
		"Write-Host \"   守护程序: $DAEMON_DIR\" -ForegroundColor Gray\n" +
		"Write-Host \"\"\n" +
		"Write-Host \"🔧 常用命令:\" -ForegroundColor White\n" +
		"Write-Host \"   frpc状态: Get-Service frpc\" -ForegroundColor Gray\n" +
		"Write-Host \"   守护程序: Get-Service frpc-daemon\" -ForegroundColor Gray\n" +
		"Write-Host \"\"\n"

	return script
}

// RegisterClient 使用Token注册客户端
func (s *ClientRegisterService) RegisterClient(token string) (*model.Client, error) {
	logger.Debug("[客户端注册] ========================================")
	logger.Debugf("[客户端注册] 收到注册请求 - Token: %s", token)

	t, err := s.tokenRepo.FindByToken(token)
	if err != nil {
		logger.Debug("[客户端注册] ❌ Token不存在")
		return nil, errors.New("token不存在")
	}

	if t.Used {
		logger.Debug("[客户端注册] ❌ Token已被使用")
		return nil, errors.New("token已被使用")
	}

	if time.Now().After(t.ExpiresAt) {
		logger.Debug("[客户端注册] ❌ Token已过期")
		return nil, errors.New("token已过期")
	}

	// 创建客户端记录,初始状态为offline,等待守护程序WS连接后更新为online
	client := &model.Client{
		Name:          t.ClientName,
		ServerAddr:    t.ServerAddr,
		ServerPort:    t.ServerPort,
		Token:         t.TokenStr,
		Protocol:      t.Protocol,
		Remark:        t.Remark,
		FrpServerID:   &t.FrpServerID,
		OnlineStatus:  "offline", // 初始状态为离线,等待守护程序连接
		WsConnected:   false,     // WS未连接
		FrpcAdminHost: "127.0.0.1",
		FrpcAdminPort: 7400,
		FrpcAdminUser: "admin",
		FrpcAdminPwd:  t.AdminPassword, // 使用生成的随机密码
	}

	logger.Debug("[客户端注册] 创建客户端记录:")
	logger.Debugf("[客户端注册]   Name: %s", client.Name)
	logger.Debugf("[客户端注册]   Server: %s:%d", client.ServerAddr, client.ServerPort)
	logger.Debugf("[客户端注册]   FrpServerID: %d", t.FrpServerID)
	logger.Debugf("[客户端注册]   初始状态: %s (等待守护程序WS连接)", client.OnlineStatus)

	if err := s.clientRepo.Create(client); err != nil {
		logger.Errorf("[客户端注册] ❌ 创建失败: %v", err)
		return nil, err
	}

	if err := s.tokenRepo.MarkAsUsed(t.ID); err != nil {
		logger.Warnf("[客户端注册] ⚠️ 标记Token失败: %v", err)
		return nil, err
	}

	logger.Debugf("[客户端注册] ✅ 注册成功 - ClientID: %d", client.ID)
	logger.Debug("[客户端注册] ========================================")
	return client, nil
}
