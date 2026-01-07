package handler

import (
	"frp-web-panel/internal/logger"
	"frp-web-panel/internal/service"
	"frp-web-panel/internal/websocket"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

var daemonUpgrader = ws.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type ClientDaemonWSHandler struct {
	clientService *service.ClientService
}

func NewClientDaemonWSHandler() *ClientDaemonWSHandler {
	return &ClientDaemonWSHandler{
		clientService: service.NewClientService(),
	}
}

// HandleConnection godoc
// @Summary Daemon WebSocket 连接
// @Description 客户端守护进程 WebSocket 连接接口，用于实时通信（无需认证，使用 Token 验证）
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Param client_id query int true "客户端ID"
// @Param token query string true "客户端Token"
// @Success 101 {string} string "WebSocket 连接成功"
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Router /api/clients/daemon/ws [get]
func (h *ClientDaemonWSHandler) HandleConnection(c *gin.Context) {
	logger.Debug("[守护进程WS] 收到WebSocket连接请求")
	logger.Debugf("[守护进程WS] 请求路径: %s", c.Request.URL.Path)
	logger.Debugf("[守护进程WS] 查询参数: %s", c.Request.URL.RawQuery)
	logger.Debugf("[守护进程WS] 客户端IP: %s", c.ClientIP())

	clientIDStr := c.Query("client_id")
	token := c.Query("token")

	logger.Debugf("[守护进程WS] 提取参数: client_id=%s, token=%s", clientIDStr, token)

	if clientIDStr == "" {
		logger.Warn("[守护进程WS] 错误: 缺少client_id参数")
		c.JSON(400, gin.H{"error": "缺少client_id参数"})
		return
	}

	clientID, err := strconv.ParseUint(clientIDStr, 10, 32)
	if err != nil {
		logger.Warnf("[守护进程WS] 错误: client_id格式错误: %v", err)
		c.JSON(400, gin.H{"error": "client_id格式错误"})
		return
	}

	logger.Debugf("[守护进程WS] 解析client_id: %d", clientID)

	client, err := h.clientService.GetClient(uint(clientID))
	if err != nil {
		logger.Warnf("[守护进程WS] 错误: 客户端不存在: ID=%d, 错误=%v", clientID, err)
		c.JSON(401, gin.H{"error": "客户端不存在"})
		return
	}

	logger.Debugf("[守护进程WS] 找到客户端: ID=%d, Name=%s", client.ID, client.Name)
	logger.Debugf("[守护进程WS] 客户端Token: %s", client.Token)
	logger.Debugf("[守护进程WS] 请求Token: %s", token)

	if client.Token != token {
		logger.Warn("[守护进程WS] 错误: Token验证失败")
		c.JSON(401, gin.H{"error": "Token验证失败"})
		return
	}

	logger.Debug("[守护进程WS] Token验证成功")

	logger.Debug("[守护进程WS] 开始升级为WebSocket连接...")
	conn, err := daemonUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Errorf("[守护进程WS] 错误: WebSocket升级失败: %v", err)
		return
	}
	logger.Debug("[守护进程WS] WebSocket升级成功")

	daemonConn := &websocket.DaemonConnection{
		ClientID: uint(clientID),
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      websocket.ClientDaemonHubInstance,
	}

	websocket.ClientDaemonHubInstance.Register <- daemonConn

	logger.Debugf("[守护进程WS] 更新客户端 %d 的WS连接状态...", clientID)
	if err := h.clientService.UpdateWSStatus(uint(clientID), true); err != nil {
		logger.Warnf("[守护进程WS] 警告: 更新WS状态失败: %v", err)
	} else {
		logger.Debug("[守护进程WS] 客户端 WS 连接状态已更新为 true")
	}

	logger.Debug("[守护进程WS] 启动读写协程...")
	go daemonConn.WritePump()
	go daemonConn.ReadPump(h)
}

// HandleMessage 实现MessageHandler接口
func (h *ClientDaemonWSHandler) HandleMessage(clientID uint, msg *websocket.Message) {
	switch msg.Type {
	case "heartbeat":
		h.handleHeartbeat(clientID)
	case "sync_result":
		h.handleSyncResult(clientID, msg)
	case "update_progress":
		h.handleUpdateProgress(clientID, msg)
	case "update_result":
		h.handleUpdateResult(clientID, msg)
	case "version_report":
		h.handleVersionReport(clientID, msg)
	case "frpc_health":
		h.handleFrpcHealth(clientID, msg)
	case "log_data":
		h.handleLogData(clientID, msg)
	case "frpc_control_result":
		h.handleFrpcControlResult(clientID, msg)
	case "config_sync_result":
		h.handleConfigSyncResult(clientID, msg)
	}
}

func (h *ClientDaemonWSHandler) handleHeartbeat(clientID uint) {
	h.clientService.UpdateHeartbeat(clientID)
}

func (h *ClientDaemonWSHandler) handleSyncResult(clientID uint, msg *websocket.Message) {
	data := msg.Data
	success := data["success"].(bool)
	version := int(data["version"].(float64))
	message := data["message"].(string)

	logger.Debugf("[客户端 %d] 配置同步结果: version=%d, success=%v, message=%s",
		clientID, version, success, message)

	if success {
		now := time.Now()
		h.clientService.UpdateConfigSync(clientID, version, &now)
	}
}

func (h *ClientDaemonWSHandler) handleUpdateProgress(clientID uint, msg *websocket.Message) {
	logger.Debugf("[客户端 %d] 收到更新进度消息", clientID)
	websocket.ClientDaemonHubInstance.HandleUpdateProgress(clientID, msg.Data)
}

func (h *ClientDaemonWSHandler) handleUpdateResult(clientID uint, msg *websocket.Message) {
	logger.Debugf("[客户端 %d] 收到更新结果消息", clientID)
	websocket.ClientDaemonHubInstance.HandleUpdateResult(clientID, msg.Data)
}

func (h *ClientDaemonWSHandler) handleVersionReport(clientID uint, msg *websocket.Message) {
	logger.Debugf("[客户端 %d] 收到版本上报消息", clientID)
	websocket.ClientDaemonHubInstance.HandleVersionReport(clientID, msg.Data)
}

func (h *ClientDaemonWSHandler) handleFrpcHealth(clientID uint, msg *websocket.Message) {
	alive, ok := msg.Data["alive"].(bool)
	if !ok {
		logger.Warnf("[客户端 %d] frpc 健康状态消息格式错误", clientID)
		return
	}

	logger.Debugf("[客户端 %d] frpc 健康状态: alive=%v", clientID, alive)

	if alive {
		if err := h.clientService.UpdateOnlineStatusDirectly(clientID, "online"); err != nil {
			logger.Errorf("[客户端 %d] 更新在线状态失败: %v", clientID, err)
		}
	} else {
		if err := h.clientService.UpdateOnlineStatusDirectly(clientID, "offline"); err != nil {
			logger.Errorf("[客户端 %d] 更新离线状态失败: %v", clientID, err)
		}
	}
}

func (h *ClientDaemonWSHandler) handleLogData(clientID uint, msg *websocket.Message) {
	logger.Debugf("[客户端 %d] 收到日志数据消息: %+v", clientID, msg.Data)
	websocket.ClientDaemonHubInstance.HandleLogData(clientID, msg.Data)
}

func (h *ClientDaemonWSHandler) handleFrpcControlResult(clientID uint, msg *websocket.Message) {
	logger.Debugf("[客户端 %d] 收到frpc控制结果消息", clientID)
	websocket.ClientDaemonHubInstance.HandleFrpcControlResult(clientID, msg.Data)
}

func (h *ClientDaemonWSHandler) handleConfigSyncResult(clientID uint, msg *websocket.Message) {
	logger.Debugf("[客户端 %d] 收到配置同步结果消息", clientID)
	websocket.ClientDaemonHubInstance.HandleConfigSyncResult(clientID, msg.Data)
}
