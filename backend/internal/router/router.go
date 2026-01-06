/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-14 15:28:45
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-31 16:42:19
 * @FilePath            : frp-web-testbackendinternalrouterrouter.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package router

import (
	"frp-web-panel/internal/container"
	"frp-web-panel/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 使用服务容器设置路由
func SetupRoutes(r *gin.Engine, c *container.Container) {
	h := c.Handlers

	api := r.Group("/api")
	{
		api.GET("/health", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{"status": "正常"})
		})

		auth := api.Group("/auth")
		{
			auth.POST("/login", h.Auth.Login)
			auth.GET("/profile", middleware.AuthMiddleware(), h.Auth.GetProfile)
			auth.PUT("/password", middleware.AuthMiddleware(), h.Auth.ChangePassword)
		}

		clients := api.Group("/clients", middleware.AuthMiddleware())
		{
			clients.GET("", h.Client.GetClients)
			clients.POST("", h.Client.CreateClient)
			clients.GET("/:id", h.Client.GetClient)
			clients.PUT("/:id", h.Client.UpdateClient)
			clients.DELETE("/:id", h.Client.DeleteClient)
			clients.GET("/:id/proxies", h.Proxy.GetProxiesByClient)
			clients.GET("/:id/export", h.Proxy.ExportConfig)
			clients.POST("/register/token", h.Client.GenerateRegisterToken)
			clients.GET("/register/script", h.Client.GenerateRegisterScript)
			clients.POST("/parse-config", h.Client.ParseConfig)
			clients.POST("/:id/update", h.Client.UpdateClientSoftware)
			clients.GET("/:id/versions", h.Client.GetClientVersions)
			clients.POST("/batch-update", h.Client.BatchUpdateClients)
			clients.GET("/online", h.Client.GetOnlineClients)
			clients.POST("/:id/logs/start", h.ClientLog.StartLogStream)
			clients.POST("/:id/logs/stop", h.ClientLog.StopLogStream)
			clients.POST("/:id/frpc/control", h.ClientLog.ControlFrpc)
		}

		api.POST("/clients/register", h.Client.RegisterClient)
		api.POST("/clients/heartbeat", middleware.RateLimitMiddleware(60), h.Client.Heartbeat)
		api.GET("/clients/daemon/ws", h.ClientDaemonWS.HandleConnection)

		r.GET("/install/:token", h.Client.GetInstallScript)
		r.GET("/download/daemon/:os/:arch", h.DaemonDownload.Download)

		proxies := api.Group("/proxies", middleware.AuthMiddleware())
		{
			proxies.GET("", h.Proxy.GetAllProxies)
			proxies.POST("", h.Proxy.CreateProxy)
			proxies.PUT("/:id", h.Proxy.UpdateProxy)
			proxies.DELETE("/:id", h.Proxy.DeleteProxy)
			proxies.PUT("/:id/toggle", h.Proxy.ToggleProxy)
		}

		traffic := api.Group("/traffic", middleware.AuthMiddleware())
		{
			traffic.GET("/summary", h.Traffic.GetTrafficSummary)
			traffic.GET("/trend", h.Traffic.GetTrafficTrend)
			traffic.GET("/proxy/:id", h.Traffic.GetTrafficHistory)
			traffic.GET("/rates/:server_id", h.Traffic.GetProxyRates)
			traffic.GET("/rates/:server_id/:proxy_name", h.Traffic.GetProxyRateHistory)
			traffic.GET("/proxies/summary", h.Traffic.GetProxiesTrafficSummary)
		}

		logs := api.Group("/logs", middleware.AuthMiddleware())
		{
			logs.GET("", h.Log.GetLogs)
			logs.POST("", h.Log.CreateLog)
		}

		settings := api.Group("/settings", middleware.AuthMiddleware())
		{
			settings.GET("", h.Setting.GetSettings)
			settings.PUT("", h.Setting.UpdateSetting)
			settings.POST("/test-email", h.Setting.TestEmail)
		}

		monitor := api.Group("/monitor", middleware.AuthMiddleware())
		{
			monitor.GET("/overview", h.Monitor.GetOverview)
			monitor.GET("/stats", h.Monitor.GetStats)
		}

		ws := api.Group("/ws", middleware.AuthMiddleware())
		{
			ws.GET("/realtime", h.WebSocket.HandleConnection)
			ws.GET("/logs/:id", h.LogWS.HandleConnection)
		}

		alerts := api.Group("/alerts", middleware.AuthMiddleware())
		{
			alerts.POST("/rules", h.Alert.CreateRule)
			alerts.GET("/rules", h.Alert.GetAllRules)
			alerts.GET("/rules/proxy/:id", h.Alert.GetRulesByProxyID)
			alerts.PUT("/rules", h.Alert.UpdateRule)
			alerts.DELETE("/rules/:id", h.Alert.DeleteRule)
			alerts.GET("/logs", h.Alert.GetAlertLogs)
			alerts.GET("/recipients", h.AlertRecipient.GetAllRecipients)
			alerts.POST("/recipients", h.AlertRecipient.CreateRecipient)
			alerts.PUT("/recipients/:id", h.AlertRecipient.UpdateRecipient)
			alerts.DELETE("/recipients/:id", h.AlertRecipient.DeleteRecipient)
			alerts.GET("/groups", h.AlertRecipient.GetAllGroups)
			alerts.POST("/groups", h.AlertRecipient.CreateGroup)
			alerts.PUT("/groups/:id", h.AlertRecipient.UpdateGroup)
			alerts.DELETE("/groups/:id", h.AlertRecipient.DeleteGroup)
			alerts.PUT("/groups/:id/recipients", h.AlertRecipient.SetGroupRecipients)
		}

		frpServers := api.Group("/frp-servers", middleware.AuthMiddleware())
		{
			frpServers.GET("", h.FrpServer.GetAll)
			frpServers.POST("", h.FrpServer.Create)
			frpServers.GET("/:id", h.FrpServer.GetByID)
			frpServers.PUT("/:id", h.FrpServer.Update)
			frpServers.DELETE("/:id", h.FrpServer.Delete)
			frpServers.POST("/test", h.FrpServer.TestConnection)
			frpServers.POST("/parse-config", h.FrpServer.ParseConfig)
			frpServers.POST("/:id/start", h.FrpServer.Start)
			frpServers.POST("/:id/stop", h.FrpServer.Stop)
			frpServers.POST("/:id/restart", h.FrpServer.Restart)
			frpServers.GET("/:id/status", h.FrpServer.GetStatus)
			frpServers.POST("/:id/download", h.FrpServer.Download)
			frpServers.POST("/:id/test-ssh", h.FrpServer.TestSSH)
			frpServers.POST("/:id/remote-install", h.FrpServer.RemoteInstall)
			frpServers.POST("/:id/remote-start", h.FrpServer.RemoteStart)
			frpServers.POST("/:id/remote-stop", h.FrpServer.RemoteStop)
			frpServers.POST("/:id/remote-restart", h.FrpServer.RemoteRestart)
			frpServers.POST("/:id/remote-uninstall", h.FrpServer.RemoteUninstall)
			frpServers.GET("/:id/remote-logs", h.FrpServer.RemoteGetLogs)
			frpServers.GET("/:id/remote-version", h.FrpServer.RemoteGetVersion)
			frpServers.GET("/:id/local-version", h.FrpServer.GetLocalVersion)
			frpServers.POST("/:id/remote-reinstall", h.FrpServer.RemoteReinstall)
			frpServers.POST("/:id/remote-upgrade", h.FrpServer.RemoteUpgrade)
			frpServers.GET("/:id/running-task", h.FrpServer.GetRunningTask)
			frpServers.GET("/:id/metrics", h.FrpServer.GetMetrics)
			frpServers.GET("/:id/metrics-history", h.FrpServer.GetMetricsHistory)
		}

		githubMirrors := api.Group("/github-mirrors", middleware.AuthMiddleware())
		{
			githubMirrors.GET("", h.GithubMirror.GetAll)
			githubMirrors.POST("", h.GithubMirror.Create)
			githubMirrors.GET("/:id", h.GithubMirror.GetByID)
			githubMirrors.PUT("/:id", h.GithubMirror.Update)
			githubMirrors.DELETE("/:id", h.GithubMirror.Delete)
			githubMirrors.POST("/:id/set-default", h.GithubMirror.SetDefault)
		}

		dns := api.Group("/dns", middleware.AuthMiddleware())
		{
			dns.GET("/providers", h.DNS.GetProviders)
			dns.POST("/providers", h.DNS.CreateProvider)
			dns.PUT("/providers/:id", h.DNS.UpdateProvider)
			dns.DELETE("/providers/:id", h.DNS.DeleteProvider)
			dns.GET("/providers/:id/secret", h.DNS.GetProviderSecret)
			dns.GET("/providers/:id/domains", h.DNS.GetProviderDomains)
			dns.POST("/providers/test", h.DNS.TestProviderConfig)
			dns.POST("/providers/:id/test", h.DNS.TestProvider)
			dns.GET("/records", h.DNS.GetRecords)
		}

		certificates := api.Group("/certificates", middleware.AuthMiddleware())
		{
			certificates.GET("", h.Certificate.ListCertificates)
			certificates.GET("/:id", h.Certificate.GetCertificate)
			certificates.POST("", h.Certificate.RequestCertificate)
			certificates.POST("/:id/renew", h.Certificate.RenewCertificate)
			certificates.POST("/:id/reapply", h.Certificate.ReapplyCertificate)
			certificates.PUT("/:id/auto-renew", h.Certificate.UpdateAutoRenew)
			certificates.GET("/:id/download", h.Certificate.DownloadCertificate)
			certificates.DELETE("/:id", h.Certificate.DeleteCertificate)
			certificates.GET("/by-domain", h.Certificate.GetCertificatesByDomain)
			certificates.GET("/expiring", h.Certificate.GetExpiringCertificates)
			certificates.GET("/active", h.Certificate.GetActiveCertificates)
			certificates.GET("/match", h.Certificate.GetMatchingCertificates)
		}
	}
}
