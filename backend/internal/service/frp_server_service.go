/*
 * FrpServerService - 核心 CRUD 操作和结构体定义
 */
package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"frp-web-panel/internal/config"
	"frp-web-panel/internal/events"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"frp-web-panel/internal/util"
	"frp-web-panel/pkg/database"
	"sync"
)

type FrpServerService struct {
	repo            *repository.FrpServerRepository
	processManager  *ProcessManager
	downloadService *DownloadService
	eventBus        *events.EventBus
	runningTasks    map[uint]string
	taskMutex       sync.RWMutex
}

func NewFrpServerService() *FrpServerService {
	githubAPI := config.GlobalConfig.Frps.GithubAPI
	if githubAPI == "" {
		githubAPI = "https://api.github.com/repos/fatedier/frp"
	}
	return &FrpServerService{
		repo:            repository.NewFrpServerRepository(database.DB),
		processManager:  NewProcessManager(),
		downloadService: NewDownloadService(githubAPI),
		eventBus:        events.GetEventBus(),
		runningTasks:    make(map[uint]string),
	}
}

// PublishSSHLog 发布SSH日志事件
func (s *FrpServerService) PublishSSHLog(serverID uint, operation, log string) {
	s.eventBus.Publish(events.SSHLogEvent{
		ServerID:  serverID,
		Operation: operation,
		Log:       log,
	})
}

// PublishServerStatus 发布服务器状态事件
func (s *FrpServerService) PublishServerStatus(serverID uint, serverName, status string) {
	s.eventBus.Publish(events.ServerStatusEvent{
		ServerID:   serverID,
		ServerName: serverName,
		Status:     status,
	})
}

// 任务管理方法
func (s *FrpServerService) setTaskRunning(id uint, operation string) bool {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()
	if _, exists := s.runningTasks[id]; exists {
		return false
	}
	s.runningTasks[id] = operation
	return true
}

func (s *FrpServerService) clearTask(id uint) {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()
	delete(s.runningTasks, id)
}

func (s *FrpServerService) GetRunningTask(id uint) (string, bool) {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()
	operation, exists := s.runningTasks[id]
	return operation, exists
}

// CRUD 操作
func (s *FrpServerService) GetAll() ([]model.FrpServer, error) {
	return s.repo.GetAll()
}

func (s *FrpServerService) GetByID(id uint) (*model.FrpServer, error) {
	return s.repo.GetByID(id)
}

func (s *FrpServerService) Create(server *model.FrpServer) error {
	if server.ServerType == model.ServerTypeRemote {
		if server.Host == "" {
			server.Host = server.SSHHost
		}
		if server.DashboardPort == 0 {
			server.DashboardPort = 7500
		}
		if server.BindPort == 0 {
			server.BindPort = 7000
		}
		if server.InstallPath == "" {
			server.InstallPath = "/opt/frps"
		}
		if server.SSHPassword != "" {
			encrypted, err := util.Encrypt(server.SSHPassword)
			if err != nil {
				return fmt.Errorf("加密密码失败: %w", err)
			}
			server.SSHPassword = encrypted
		}
	}

	if server.DashboardUser == "" {
		server.DashboardUser = "admin"
	}
	if server.DashboardPwd == "" {
		server.DashboardPwd = generateRandomPassword(16)
	}

	return s.repo.Create(server)
}

func (s *FrpServerService) Update(server *model.FrpServer) error {
	if server.ServerType == model.ServerTypeRemote {
		if server.Host == "" {
			server.Host = server.SSHHost
		}
		if server.DashboardPort == 0 {
			server.DashboardPort = 7500
		}
		if server.BindPort == 0 {
			server.BindPort = 7000
		}
		if server.InstallPath == "" {
			server.InstallPath = "/opt/frps"
		}
		if server.SSHPassword != "" {
			existing, err := s.repo.GetByID(server.ID)
			if err != nil {
				return err
			}
			if server.SSHPassword != existing.SSHPassword {
				encrypted, err := util.Encrypt(server.SSHPassword)
				if err != nil {
					return fmt.Errorf("加密密码失败: %w", err)
				}
				server.SSHPassword = encrypted
			}
		}
	}
	return s.repo.Update(server)
}

func (s *FrpServerService) Delete(id uint, removeInstallation bool) error {
	if removeInstallation {
		server, err := s.repo.GetByID(id)
		if err != nil {
			return err
		}
		if server.ServerType == model.ServerTypeRemote && server.BinaryPath != "" {
			logFunc := func(msg string) {
				s.PublishSSHLog(id, "uninstall", msg)
			}
			installer, err := NewRemoteFrpsInstaller(server, logFunc)
			if err == nil {
				installer.Uninstall(server.InstallPath)
				installer.Close()
			}
		}
	}
	return s.repo.Delete(id)
}

// 辅助函数
func generateRandomPassword(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}
