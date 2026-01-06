/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-18 11:15:09
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-19 14:53:46
 * @FilePath            : frp-web-testbackendinternalmiddlewaresecurity.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package middleware

import (
	"github.com/gin-gonic/gin"
)

func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// CSP 策略：允许内联脚本/样式（React需要）、WebSocket连接、data URI
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data: blob: https:; "+
				"font-src 'self' data:; "+
				"connect-src 'self' ws: wss: http: https:;")
		c.Next()
	}
}
