/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-29 15:30:53
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-29 15:31:03
 * @FilePath            : frp-web-testbackendinternalmiddlewareerror_handler.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package middleware

import (
	"frp-web-panel/internal/errors"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware 统一错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] %v\n%s", r, debug.Stack())
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    errors.CodeInternal,
					"message": "服务器内部错误",
				})
				c.Abort()
			}
		}()

		c.Next()

		// 检查是否有 AppError 存储在 context 中
		if err, exists := c.Get("app_error"); exists {
			if appErr, ok := err.(*errors.AppError); ok {
				c.JSON(appErr.HTTPStatus(), gin.H{
					"code":    appErr.Code,
					"message": appErr.Message,
				})
				return
			}
		}
	}
}

// HandleAppError 将 AppError 设置到 context 并中止请求
func HandleAppError(c *gin.Context, err *errors.AppError) {
	c.Set("app_error", err)
	c.JSON(err.HTTPStatus(), gin.H{
		"code":    err.Code,
		"message": err.Message,
	})
	c.Abort()
}

// AbortWithAppError 直接返回 AppError 响应
func AbortWithAppError(c *gin.Context, err *errors.AppError) {
	c.JSON(err.HTTPStatus(), gin.H{
		"code":    err.Code,
		"message": err.Message,
	})
	c.Abort()
}
