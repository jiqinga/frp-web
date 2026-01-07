/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-17 16:16:48
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-07 10:56:43
 * @FilePath            : frp-web-testbackendinternalwebsocketclient.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package websocket

import (
	"frp-web-panel/internal/logger"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 512
)

type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseNormalClosure,    // 1000 - 正常关闭
				websocket.CloseGoingAway,        // 1001 - 页面离开
				websocket.CloseNoStatusReceived, // 1005 - 无状态关闭（浏览器直接关闭）
				websocket.CloseAbnormalClosure,  // 1006 - 异常关闭
			) {
				logger.Errorf("WebSocket错误: %v", err)
			}
			break
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
