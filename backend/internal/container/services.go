/*
 * Service 容器 - 集中管理所有 Service 实例
 */
package container

import (
	"frp-web-panel/internal/config"
	"frp-web-panel/internal/service"
	"frp-web-panel/internal/websocket"
)

// Services 包含所有 Service 实例
type Services struct {
	ACME                *service.ACMEService
	Alert               *service.AlertService
	AlertRecipient      *service.AlertRecipientService
	Auth                *service.AuthService
	CertRenewal         *service.CertRenewalScheduler
	Client              *service.ClientService
	ClientRegister      *service.ClientRegisterService
	ClientStatusChecker *service.ClientStatusChecker
	ClientUpdate        *service.ClientUpdateService
	DNS                 *service.DNSService
	Download            *service.DownloadService
	Email               *service.EmailService
	FrpServer           *service.FrpServerService
	FrpSync             *service.FrpSyncService
	GithubMirror        *service.GithubMirrorService
	Log                 *service.LogService
	MetricsCollector    *service.MetricsCollector
	Monitor             *service.MonitorService
	Proxy               *service.ProxyService
	Realtime            *service.RealtimeService
	Setting             *service.SettingService
	TaskManager         *service.TaskManager
	Traffic             *service.TrafficService
}

// NewServices 创建所有 Service 实例
func NewServices(repos *Repositories, hub *websocket.Hub, cfg *config.Config, clientDaemonHub *websocket.ClientDaemonHub) *Services {
	// 创建基础服务
	taskManager := service.NewTaskManager()
	emailService := service.NewEmailService()
	logService := service.NewLogService()
	settingService := service.NewSettingService()

	// 创建 FRP 相关服务
	githubAPI := cfg.Frps.GithubAPI
	if githubAPI == "" {
		githubAPI = "https://api.github.com/repos/fatedier/frp"
	}
	downloadService := service.NewDownloadService(githubAPI)
	frpServerService := service.NewFrpServerService()
	frpSyncService := service.NewFrpSyncService()

	// 创建实时服务
	realtimeService := service.NewRealtimeService()

	// 设置事件订阅器，将事件总线事件转发到 WebSocket Hub
	websocket.SetupEventSubscribers(hub)

	// 设置日志数据回调，将 Daemon 日志转发到前端 WebSocket
	clientDaemonHub.SetLogDataCallback(func(clientID uint, logType string, line string, timestamp int64) {
		websocket.LogWSHubInstance.BroadcastLog(clientID, logType, line, timestamp)
	})

	// 创建客户端服务（提前创建，供回调使用）
	clientServiceForCallback := service.NewClientService()

	// 设置配置同步结果回调，更新客户端配置同步状态
	clientDaemonHub.SetConfigSyncResultCallback(func(clientID uint, success bool, errorMsg string, rolledBack bool) {
		clientServiceForCallback.UpdateConfigSyncStatus(clientID, success, errorMsg, rolledBack)
	})

	// 创建指标采集服务
	metricsCollector := service.NewMetricsCollector(repos.FrpServer)

	// 创建告警服务
	alertService := service.NewAlertService(repos.Alert, repos.Traffic, repos.Proxy)
	alertService.SetClientRepo(repos.Client)
	alertService.SetFrpServerRepo(repos.FrpServer)
	alertRecipientService := service.NewAlertRecipientService()

	// 创建客户端相关服务
	clientService := service.NewClientService()
	clientRegisterService := service.NewClientRegisterService()
	clientStatusChecker := service.NewClientStatusChecker(clientService)
	clientUpdateService := service.NewClientUpdateService(realtimeService)

	// 创建证书相关服务
	acmeService := service.NewACMEService(false)
	acmeService.SetDaemonHub(clientDaemonHub)
	certRenewalScheduler := service.NewCertRenewalScheduler(acmeService)

	// 创建其他服务
	dnsService := service.NewDNSService()
	githubMirrorService := service.NewGithubMirrorService()
	monitorService := service.NewMonitorService()
	proxyService := service.NewProxyService()
	trafficService := service.NewTrafficService()
	authService := service.NewAuthService()

	return &Services{
		ACME:                acmeService,
		Alert:               alertService,
		AlertRecipient:      alertRecipientService,
		Auth:                authService,
		CertRenewal:         certRenewalScheduler,
		Client:              clientService,
		ClientRegister:      clientRegisterService,
		ClientStatusChecker: clientStatusChecker,
		ClientUpdate:        clientUpdateService,
		DNS:                 dnsService,
		Download:            downloadService,
		Email:               emailService,
		FrpServer:           frpServerService,
		FrpSync:             frpSyncService,
		GithubMirror:        githubMirrorService,
		Log:                 logService,
		MetricsCollector:    metricsCollector,
		Monitor:             monitorService,
		Proxy:               proxyService,
		Realtime:            realtimeService,
		Setting:             settingService,
		TaskManager:         taskManager,
		Traffic:             trafficService,
	}
}
