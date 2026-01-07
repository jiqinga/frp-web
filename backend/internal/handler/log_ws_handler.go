/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-31 16:41:12
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-07 10:58:37
 * @FilePath            : frp-web-testbackendinternalhandlerlog_ws_handler.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package handler

import (
	"frp-web-panel/internal/logger"
	"frp-web-panel/internal/websocket"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

var logUpgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type LogWSHandler struct{}

func NewLogWSHandler() *LogWSHandler {
	return &LogWSHandler{}
}

// HandleConnection 处理日志 WebSocket 连接
// @Summary 日志流WebSocket连接
// @Description 建立WebSocket连接以接收指定客户端的实时日志
// @Tags WebSocket
// @Security BearerAuth
// @Param id path int true "客户端ID"
// @Success 101 {string} string "WebSocket连接建立成功"
// @Router /api/ws/logs/{id} [get]
func (h *LogWSHandler) HandleConnection(c *gin.Context) {
	clientIDStr := c.Param("id")
	clientID, err := strconv.ParseUint(clientIDStr, 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的客户端ID"})
		return
	}

	conn, err := logUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Errorf("[LogWS] 升级失败: %v", err)
		return
	}

	logger.Debugf("[LogWS] 前端连接成功，clientID=%d", clientID)

	client := &websocket.LogWSClient{
		ClientID: uint(clientID),
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      websocket.LogWSHubInstance,
	}

	websocket.LogWSHubInstance.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
