/*
 * @Author              : 寂情啊
 * @Date                : 2026-01-08 13:56:18
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-08 14:05:47
 * @FilePath            : frp-web-testbackendinternalmiddlewarewebsocket_debug.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package middleware

import (
	"strings"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// ConditionalGzip 条件Gzip中间件，跳过WebSocket请求
func ConditionalGzip() gin.HandlerFunc {
	gzipHandler := gzip.Gzip(gzip.DefaultCompression)
	return func(c *gin.Context) {
		// 检测WebSocket请求
		if strings.Contains(c.Request.URL.Path, "/ws/") || c.GetHeader("Upgrade") == "websocket" {
			c.Next()
			return
		}
		gzipHandler(c)
	}
}
