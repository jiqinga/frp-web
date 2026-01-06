/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-29 14:59:01
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-31 16:41:48
 * @FilePath            : frp-web-testbackendinternalcontainerhandlers.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
/*
 * Handler 容器 - 集中管理所有 Handler 实例
 */
package container

import (
	"frp-web-panel/internal/handler"
	"frp-web-panel/internal/websocket"
)

// Handlers 包含所有 Handler 实例
type Handlers struct {
	Alert          *handler.AlertHandler
	AlertRecipient *handler.AlertRecipientHandler
	Auth           *handler.AuthHandler
	Certificate    *handler.CertificateHandler
	Client         *handler.ClientHandler
	ClientDaemonWS *handler.ClientDaemonWSHandler
	ClientLog      *handler.ClientLogHandler
	DaemonDownload *handler.DaemonDownloadHandler
	DNS            *handler.DNSHandler
	FrpServer      *handler.FrpServerHandler
	GithubMirror   *handler.GithubMirrorHandler
	Log            *handler.LogHandler
	LogWS          *handler.LogWSHandler
	Monitor        *handler.MonitorHandler
	Proxy          *handler.ProxyHandler
	Setting        *handler.SettingHandler
	Traffic        *handler.TrafficHandler
	WebSocket      *handler.WebSocketHandler
}

// NewHandlers 创建所有 Handler 实例
func NewHandlers(services *Services, repos *Repositories, hub *websocket.Hub) *Handlers {
	return &Handlers{
		Alert:          handler.NewAlertHandler(services.Alert),
		AlertRecipient: handler.NewAlertRecipientHandler(),
		Auth:           handler.NewAuthHandler(),
		Certificate:    handler.NewCertificateHandler(repos.Certificate, services.ACME),
		Client:         handler.NewClientHandler(services.Client, services.ClientRegister, services.ClientUpdate, services.Log),
		ClientDaemonWS: handler.NewClientDaemonWSHandler(),
		ClientLog:      handler.NewClientLogHandler(),
		DaemonDownload: handler.NewDaemonDownloadHandler(),
		DNS:            handler.NewDNSHandler(),
		FrpServer:      handler.NewFrpServerHandler(services.FrpServer, services.Log, repos.ServerMetrics),
		GithubMirror:   handler.NewGithubMirrorHandler(),
		Log:            handler.NewLogHandler(),
		LogWS:          handler.NewLogWSHandler(),
		Monitor:        handler.NewMonitorHandler(),
		Proxy:          handler.NewProxyHandler(),
		Setting:        handler.NewSettingHandlerWithService(services.Realtime, services.MetricsCollector),
		Traffic:        handler.NewTrafficHandler(),
		WebSocket:      handler.NewWebSocketHandler(hub),
	}
}
