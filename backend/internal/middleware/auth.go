/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:26:53
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-20 16:30:38
 * @FilePath            : frp-web-testbackendinternalmiddlewareauth.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package middleware

import (
	"frp-web-panel/internal/config"
	"frp-web-panel/internal/util"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// 优先从Header获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}

		// 如果Header中没有，尝试从query参数获取（用于WebSocket）
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			util.ErrorWithStatus(c, 401, 401, "未提供认证令牌")
			c.Abort()
			return
		}

		claims, err := util.ParseToken(token, config.GlobalConfig.JWT.Secret)
		if err != nil {
			util.ErrorWithStatus(c, 401, 401, "认证令牌无效")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
