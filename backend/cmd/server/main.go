/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:28:16
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-08 14:05:24
 * @FilePath            : frp-web-testbackendcmdservermain.go
 * @Description         : 程序入口
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
// @title FRP 管理面板 API
// @version 1.0
// @description FRP 管理面板后端 API 文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"context"
	"fmt"
	"frp-web-panel/internal/logger"
	"frp-web-panel/internal/middleware"
	"frp-web-panel/internal/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "frp-web-panel/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// 启动初始化
	result, err := Bootstrap("./configs/config.yaml")
	if err != nil {
		logger.Fatalf("启动初始化失败: %v", err)
	}
	defer Cleanup()

	c := result.Container

	// 注册回调和定时任务
	RegisterCallbacks(c)
	RegisterScheduledTasks(c)

	// 启动后台服务
	StartServices(c)

	// 设置 HTTP 服务器
	gin.SetMode(result.Config.Server.Mode)
	r := gin.Default()

	// 添加条件Gzip中间件，跳过WebSocket请求
	r.Use(middleware.ConditionalGzip())
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimitMiddleware(100))
	r.Use(middleware.InputValidationMiddleware())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.SetupRoutes(r, c)

	addr := fmt.Sprintf(":%d", result.Config.Server.Port)
	logger.Infof("服务器启动于 %s", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("正在关闭服务器...")

	// 停止后台服务
	StopServices(c, 10*time.Second)

	// 关闭 HTTP 服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown")
	}

	logger.Info("服务器已退出")
}
