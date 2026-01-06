package main

import (
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type WSClient struct {
	cfg           *Config
	conn          *websocket.Conn
	done          chan struct{}
	reconnect     chan struct{}
	writeMu       sync.Mutex // 保护 WebSocket 写操作的互斥锁
	onConfig      func(config string, version int)
	onShutdown    func()                                                                     // 收到停止命令时的回调
	onUpdate      func(updateType string, version string, downloadURL string, mirrorID uint) // 收到更新命令时的回调
	onCertSync    func(domain string, certPEM string, keyPEM string)                         // 收到证书同步时的回调
	onCertDelete  func(domain string)                                                        // 收到证书删除时的回调
	onLogStream   func(logType string, action string, lines int)                             // 收到日志流命令时的回调
	onFrpcControl func(action string)                                                        // 收到frpc控制命令时的回调
}

func NewWSClient(cfg *Config, onConfig func(string, int)) *WSClient {
	return &WSClient{
		cfg:       cfg,
		done:      make(chan struct{}),
		reconnect: make(chan struct{}, 1),
		onConfig:  onConfig,
	}
}

// SetUpdateCallback 设置更新命令回调函数
func (c *WSClient) SetUpdateCallback(callback func(updateType string, version string, downloadURL string, mirrorID uint)) {
	c.onUpdate = callback
}

// SetCertSyncCallback 设置证书同步回调函数
func (c *WSClient) SetCertSyncCallback(callback func(domain string, certPEM string, keyPEM string)) {
	c.onCertSync = callback
}

// SetCertDeleteCallback 设置证书删除回调函数
func (c *WSClient) SetCertDeleteCallback(callback func(domain string)) {
	c.onCertDelete = callback
}

// SetShutdownCallback 设置停止命令回调函数
func (c *WSClient) SetShutdownCallback(callback func()) {
	c.onShutdown = callback
}

// SetLogStreamCallback 设置日志流命令回调函数
func (c *WSClient) SetLogStreamCallback(callback func(logType string, action string, lines int)) {
	c.onLogStream = callback
}

// SetFrpcControlCallback 设置frpc控制命令回调函数
func (c *WSClient) SetFrpcControlCallback(callback func(action string)) {
	c.onFrpcControl = callback
}

func (c *WSClient) Connect() error {
	u, _ := url.Parse(c.cfg.ServerURL)
	u.Path = "/api/clients/daemon/ws"
	q := u.Query()
	q.Set("client_id", fmt.Sprintf("%d", c.cfg.ClientID))
	q.Set("token", c.cfg.Token)
	u.RawQuery = q.Encode()

	log.Printf("[WS客户端] 准备连接到服务器")
	log.Printf("[WS客户端] ServerURL: %s", c.cfg.ServerURL)
	log.Printf("[WS客户端] 完整URL: %s", u.String())
	log.Printf("[WS客户端] ClientID: %d", c.cfg.ClientID)
	// Token 脱敏处理，只显示前4个字符
	maskedToken := c.cfg.Token
	if len(maskedToken) > 4 {
		maskedToken = maskedToken[:4] + "****"
	}
	log.Printf("[WS客户端] Token: %s", maskedToken)

	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		if resp != nil {
			log.Printf("[WS客户端] ❌ 连接失败, HTTP状态码: %d", resp.StatusCode)
			log.Printf("[WS客户端] 响应头: %v", resp.Header)
		}
		log.Printf("[WS客户端] ❌ 连接错误: %v", err)
		return err
	}
	c.conn = conn
	log.Println("[WS客户端] ✅ 连接成功")
	return nil
}

func (c *WSClient) Run() {
	backoff := 5
	for {
		// 重连前确保关闭旧连接，防止连接泄漏
		if c.conn != nil {
			c.conn.Close()
			c.conn = nil
		}

		if err := c.Connect(); err != nil {
			log.Printf("[WS] 连接失败: %v, %d秒后重试", err, backoff)
			time.Sleep(time.Duration(backoff) * time.Second)
			if backoff < 60 {
				backoff *= 2
			}
			continue
		}
		backoff = 5

		go c.heartbeat()
		c.readLoop()

		select {
		case <-c.done:
			return
		case <-c.reconnect:
			log.Println("[WS] 准备重连...")
			time.Sleep(5 * time.Second)
		}
	}
}

func (c *WSClient) readLoop() {
	defer func() {
		c.conn.Close()
		select {
		case c.reconnect <- struct{}{}:
		default:
		}
	}()

	for {
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			log.Printf("[WS] ❌ 读取消息失败: %v", err)
			return
		}

		log.Printf("[WS] 收到消息类型: %s", msg.Type)

		switch msg.Type {
		case "config_update":
			log.Printf("[WS] ========== 收到配置更新消息 ==========")
			configRaw, ok := msg.Data["config"]
			if !ok {
				log.Printf("[WS] ❌ 配置更新消息中缺少 config 字段")
				continue
			}
			config, ok := configRaw.(string)
			if !ok {
				log.Printf("[WS] ❌ config 字段类型错误: %T", configRaw)
				continue
			}

			versionRaw, ok := msg.Data["version"]
			if !ok {
				log.Printf("[WS] ❌ 配置更新消息中缺少 version 字段")
				continue
			}
			version := int(versionRaw.(float64))

			log.Printf("[WS] ✅ 收到配置更新: version=%d", version)
			log.Printf("[WS] 配置内容长度: %d 字节", len(config))

			if c.onConfig != nil {
				log.Printf("[WS] 调用配置处理回调...")
				c.onConfig(config, version)
			} else {
				log.Printf("[WS] ⚠️ 配置处理回调未设置!")
			}
		case "shutdown":
			log.Printf("[WS] ========== 收到停止命令 ==========")
			log.Printf("[WS] 服务器请求停止 daemon 和 frpc")
			if c.onShutdown != nil {
				log.Printf("[WS] 调用停止回调...")
				c.onShutdown()
			} else {
				log.Printf("[WS] ⚠️ 停止回调未设置，直接退出")
				return
			}
		case "update":
			log.Printf("[WS] ========== 收到更新命令 ==========")
			updateTypeRaw, _ := msg.Data["update_type"].(string)
			versionRaw, _ := msg.Data["version"].(string)
			downloadURLRaw, _ := msg.Data["download_url"].(string)
			mirrorIDRaw := uint(0)
			if mid, ok := msg.Data["mirror_id"].(float64); ok {
				mirrorIDRaw = uint(mid)
			}
			log.Printf("[WS] 更新类型: %s, 版本: %s, 下载地址: %s", updateTypeRaw, versionRaw, downloadURLRaw)
			if c.onUpdate != nil {
				log.Printf("[WS] 调用更新回调...")
				c.onUpdate(updateTypeRaw, versionRaw, downloadURLRaw, mirrorIDRaw)
			} else {
				log.Printf("[WS] ⚠️ 更新回调未设置!")
			}
		case "cert_sync":
			log.Printf("[WS] ========== 收到证书同步命令 ==========")
			domain, _ := msg.Data["domain"].(string)
			certPEM, _ := msg.Data["cert_pem"].(string)
			keyPEM, _ := msg.Data["key_pem"].(string)
			log.Printf("[WS] 证书域名: %s, 证书长度: %d, 私钥长度: %d", domain, len(certPEM), len(keyPEM))
			if c.onCertSync != nil {
				log.Printf("[WS] 调用证书同步回调...")
				c.onCertSync(domain, certPEM, keyPEM)
			} else {
				log.Printf("[WS] ⚠️ 证书同步回调未设置!")
			}
		case "cert_delete":
			log.Printf("[WS] ========== 收到证书删除命令 ==========")
			domain, _ := msg.Data["domain"].(string)
			log.Printf("[WS] 要删除的证书域名: %s", domain)
			if c.onCertDelete != nil {
				log.Printf("[WS] 调用证书删除回调...")
				c.onCertDelete(domain)
			} else {
				log.Printf("[WS] ⚠️ 证书删除回调未设置!")
			}
		case "log_stream":
			log.Printf("[WS] ========== 收到日志流命令 ==========")
			logType, _ := msg.Data["log_type"].(string)
			action, _ := msg.Data["action"].(string)
			lines := 100
			if linesRaw, ok := msg.Data["lines"].(float64); ok {
				lines = int(linesRaw)
			}
			log.Printf("[WS] 日志类型: %s, 操作: %s, 行数: %d", logType, action, lines)
			if c.onLogStream != nil {
				log.Printf("[WS] 调用日志流回调...")
				c.onLogStream(logType, action, lines)
			} else {
				log.Printf("[WS] ⚠️ 日志流回调未设置!")
			}
		case "frpc_control":
			log.Printf("[WS] ========== 收到frpc控制命令 ==========")
			action, _ := msg.Data["action"].(string)
			log.Printf("[WS] 控制操作: %s", action)
			if c.onFrpcControl != nil {
				log.Printf("[WS] 调用frpc控制回调...")
				c.onFrpcControl(action)
			} else {
				log.Printf("[WS] ⚠️ frpc控制回调未设置!")
			}
		default:
			log.Printf("[WS] 未知消息类型: %s", msg.Type)
		}
	}
}

func (c *WSClient) heartbeat() {
	ticker := time.NewTicker(time.Duration(c.cfg.HeartbeatSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			msg := Message{Type: "heartbeat"}
			if err := c.writeJSON(msg); err != nil {
				log.Printf("[WS] 心跳发送失败: %v, 触发重连", err)
				// 心跳失败时触发重连
				select {
				case c.reconnect <- struct{}{}:
				default:
				}
				return
			}
		case <-c.done:
			return
		}
	}
}

// writeJSON 线程安全的 JSON 写入方法
func (c *WSClient) writeJSON(msg interface{}) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteJSON(msg)
}

func (c *WSClient) SendSyncResult(success bool, version int, message string) {
	msg := Message{
		Type: "sync_result",
		Data: map[string]interface{}{
			"success": success,
			"version": version,
			"message": message,
		},
	}
	if err := c.writeJSON(msg); err != nil {
		log.Printf("[WS] 发送同步结果失败: %v", err)
	}
}

func (c *WSClient) Close() {
	close(c.done)
	if c.conn != nil {
		c.conn.Close()
	}
}

// SendUpdateProgress 发送更新进度
func (c *WSClient) SendUpdateProgress(progress UpdateProgress) {
	msg := Message{
		Type: "update_progress",
		Data: map[string]interface{}{
			"update_type":      string(progress.UpdateType),
			"stage":            string(progress.Stage),
			"progress":         progress.Progress,
			"message":          progress.Message,
			"total_bytes":      progress.TotalBytes,
			"downloaded_bytes": progress.DownloadedBytes,
		},
	}
	if err := c.writeJSON(msg); err != nil {
		log.Printf("[WS] 发送更新进度失败: %v", err)
	}
}

// SendUpdateResult 发送更新结果
func (c *WSClient) SendUpdateResult(result UpdateResult) {
	msg := Message{
		Type: "update_result",
		Data: map[string]interface{}{
			"update_type": string(result.UpdateType),
			"success":     result.Success,
			"version":     result.Version,
			"message":     result.Message,
		},
	}
	if err := c.writeJSON(msg); err != nil {
		log.Printf("[WS] 发送更新结果失败: %v", err)
	}
}

// SendVersionReport 发送版本上报
func (c *WSClient) SendVersionReport(frpcVersion, daemonVersion, os, arch string) {
	msg := Message{
		Type: "version_report",
		Data: map[string]interface{}{
			"frpc_version":   frpcVersion,
			"daemon_version": daemonVersion,
			"os":             os,
			"arch":           arch,
		},
	}
	if err := c.writeJSON(msg); err != nil {
		log.Printf("[WS] 发送版本上报失败: %v", err)
	}
}

// SendFrpcHealthStatus 发送 frpc 健康状态
func (c *WSClient) SendFrpcHealthStatus(alive bool) {
	msg := Message{
		Type: "frpc_health",
		Data: map[string]interface{}{
			"alive": alive,
		},
	}
	if err := c.writeJSON(msg); err != nil {
		log.Printf("[WS] 发送 frpc 健康状态失败: %v", err)
	}
}

// SendLogData 发送日志数据
func (c *WSClient) SendLogData(logType string, line string) {
	msg := Message{
		Type: "log_data",
		Data: map[string]interface{}{
			"log_type":  logType,
			"line":      line,
			"timestamp": time.Now().Unix(),
		},
	}
	if err := c.writeJSON(msg); err != nil {
		log.Printf("[WS] 发送日志数据失败: %v", err)
	}
}

// SendFrpcControlResult 发送frpc控制结果
func (c *WSClient) SendFrpcControlResult(action string, success bool, message string) {
	msg := Message{
		Type: "frpc_control_result",
		Data: map[string]interface{}{
			"action":  action,
			"success": success,
			"message": message,
		},
	}
	if err := c.writeJSON(msg); err != nil {
		log.Printf("[WS] 发送frpc控制结果失败: %v", err)
	}
}

// SendConfigSyncResult 发送配置同步结果
func (c *WSClient) SendConfigSyncResult(result ConfigSyncResult) {
	msg := Message{
		Type: "config_sync_result",
		Data: map[string]interface{}{
			"success":     result.Success,
			"error":       result.Error,
			"rolled_back": result.RolledBack,
			"timestamp":   result.Timestamp,
		},
	}
	if err := c.writeJSON(msg); err != nil {
		log.Printf("[WS] 发送配置同步结果失败: %v", err)
	} else {
		log.Printf("[WS] ✅ 配置同步结果已发送: success=%v, rolled_back=%v", result.Success, result.RolledBack)
	}
}
