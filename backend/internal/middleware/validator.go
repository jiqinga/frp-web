/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-18 11:18:31
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-26 16:41:48
 * @FilePath            : frp-web-testbackendinternalmiddlewarevalidator.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package middleware

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

var (
	sqlInjectionPattern = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|script|javascript|<script)`)
	xssPattern          = regexp.MustCompile(`(?i)(<script|javascript:|onerror=|onload=)`)
)

func InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, values := range c.Request.URL.Query() {
			for _, value := range values {
				if sqlInjectionPattern.MatchString(value) || xssPattern.MatchString(value) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "非法输入"})
					c.Abort()
					return
				}
			}
		}
		c.Next()
	}
}
