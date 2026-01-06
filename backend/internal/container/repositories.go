/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-29 14:56:44
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-29 14:57:58
 * @FilePath            : frp-web-testbackendinternalcontainerrepositories.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
/*
 * Repository 容器 - 集中管理所有 Repository 实例
 */
package container

import (
	"frp-web-panel/internal/repository"

	"gorm.io/gorm"
)

// Repositories 包含所有 Repository 实例
type Repositories struct {
	Alert          *repository.AlertRepo
	AlertRecipient *repository.AlertRecipientRepo
	Certificate    *repository.CertificateRepository
	Client         *repository.ClientRepository
	ClientRegToken *repository.ClientRegisterTokenRepository
	DNSProvider    *repository.DNSProviderRepository
	DNSRecord      *repository.DNSRecordRepository
	FrpServer      *repository.FrpServerRepository
	GithubMirror   *repository.GithubMirrorRepository
	Log            *repository.LogRepository
	Proxy          *repository.ProxyRepository
	ProxyMetrics   *repository.ProxyMetricsRepository
	ServerMetrics  *repository.ServerMetricsRepository
	Setting        *repository.SettingRepository
	Traffic        *repository.TrafficRepository
	User           *repository.UserRepository
}

// NewRepositories 创建所有 Repository 实例
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Alert:          repository.NewAlertRepo(db),
		AlertRecipient: repository.NewAlertRecipientRepo(),
		Certificate:    repository.NewCertificateRepository(),
		Client:         repository.NewClientRepository(),
		ClientRegToken: repository.NewClientRegisterTokenRepository(),
		DNSProvider:    repository.NewDNSProviderRepository(),
		DNSRecord:      repository.NewDNSRecordRepository(),
		FrpServer:      repository.NewFrpServerRepository(db),
		GithubMirror:   repository.NewGithubMirrorRepository(),
		Log:            repository.NewLogRepository(),
		Proxy:          repository.NewProxyRepository(),
		ProxyMetrics:   repository.NewProxyMetricsRepository(),
		ServerMetrics:  repository.NewServerMetricsRepository(),
		Setting:        repository.NewSettingRepository(),
		Traffic:        repository.NewTrafficRepository(),
		User:           repository.NewUserRepository(),
	}
}
