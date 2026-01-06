/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 16:25:18
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-30 17:08:34
 * @FilePath            : frp-web-testbackendinternalhandlerwebsocket_handler.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package handler

import (
	"frp-web-panel/internal/websocket"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	hub *websocket.Hub
}

func NewWebSocketHandler(hub *websocket.Hub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

// HandleConnection godoc
// @Summary 实时数据WebSocket连接
// @Description 建立WebSocket连接以接收实时数据推送，包括服务器状态、流量统计等
// @Tags WebSocket
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param token query string false "认证Token(可选，也可通过Authorization头传递)"
// @Success 101 {string} string "WebSocket连接建立成功"
// @Failure 401 {string} string "认证失败"
// @Router /api/ws/realtime [get]
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
	log.Println("[WebSocket] 收到连接请求")
	log.Println("[WebSocket] Token:", c.Query("token"))
	log.Println("[WebSocket] Authorization Header:", c.GetHeader("Authorization"))

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("[WebSocket] 升级失败:", err)
		return
	}

	log.Println("[WebSocket] 连接升级成功")

	client := &websocket.Client{
		Hub:  h.hub,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	h.hub.Register <- client
	log.Println("[WebSocket] 客户端已注册到Hub")

	go client.WritePump()
	go client.ReadPump()
}
