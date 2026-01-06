/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-29 14:59:34
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-29 16:12:15
 * @FilePath            : frp-web-testbackendinternalcontainercontainer.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
/*
 * Container - 服务容器主入口
 * 集中管理所有 Repository、Service、Handler 实例的依赖注入
 */
package container

import (
	"frp-web-panel/internal/config"
	"frp-web-panel/internal/websocket"

	"gorm.io/gorm"
)

// Container 应用服务容器
type Container struct {
	DB              *gorm.DB
	Config          *config.Config
	Hub             *websocket.Hub
	ClientDaemonHub *websocket.ClientDaemonHub
	Repositories    *Repositories
	Services        *Services
	Handlers        *Handlers
}

// NewContainer 创建服务容器
func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	// 创建 WebSocket Hub
	hub := websocket.NewHub()

	// 获取 ClientDaemonHub 实例（已在 init 中初始化）
	clientDaemonHub := websocket.ClientDaemonHubInstance

	// 按依赖顺序创建各层实例
	repos := NewRepositories(db)
	services := NewServices(repos, hub, cfg, clientDaemonHub)
	handlers := NewHandlers(services, repos, hub)

	return &Container{
		DB:              db,
		Config:          cfg,
		Hub:             hub,
		ClientDaemonHub: clientDaemonHub,
		Repositories:    repos,
		Services:        services,
		Handlers:        handlers,
	}
}
