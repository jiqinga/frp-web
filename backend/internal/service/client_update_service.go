/*
 * @Author              : Kilo Code
 * @Date                : 2025-12-01
 * @Description         : 客户端更新服务
 */
package service

import (
	"fmt"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"frp-web-panel/internal/websocket"
	"frp-web-panel/pkg/database"
	"log"
	"strings"
)

// UpdateType 更新类型
type UpdateType string

const (
	UpdateTypeFrpc   UpdateType = "frpc"
	UpdateTypeDaemon UpdateType = "daemon"
)

// UpdateRequest 更新请求
type UpdateRequest struct {
	ClientID   uint       `json:"client_id"`
	UpdateType UpdateType `json:"update_type"`
	Version    string     `json:"version,omitempty"`   // 可选，不指定则使用关联frps的版本
	MirrorID   *uint      `json:"mirror_id,omitempty"` // 可选，下载镜像源
}

// BatchUpdateRequest 批量更新请求
type BatchUpdateRequest struct {
	ClientIDs  []uint     `json:"client_ids"`
	UpdateType UpdateType `json:"update_type"`
	Version    string     `json:"version,omitempty"`
	MirrorID   *uint      `json:"mirror_id,omitempty"`
}

// ClientUpdateService 客户端更新服务
type ClientUpdateService struct {
	clientRepo       *repository.ClientRepository
	frpServerRepo    *repository.FrpServerRepository
	githubMirrorRepo *repository.GithubMirrorRepository
	realtimeService  *RealtimeService
}

// NewClientUpdateService 创建客户端更新服务
func NewClientUpdateService(realtimeService *RealtimeService) *ClientUpdateService {
	svc := &ClientUpdateService{
		clientRepo:       repository.NewClientRepository(),
		frpServerRepo:    repository.NewFrpServerRepository(database.DB),
		githubMirrorRepo: repository.NewGithubMirrorRepository(),
		realtimeService:  realtimeService,
	}

	// 设置回调函数
	svc.setupCallbacks()

	return svc
}

// setupCallbacks 设置WebSocket回调函数
func (s *ClientUpdateService) setupCallbacks() {
	// 设置更新进度回调
	websocket.ClientDaemonHubInstance.SetUpdateProgressCallback(func(clientID uint, updateType string, stage string, progress int, message string, totalBytes int64, downloadedBytes int64) {
		log.Printf("[ClientUpdateService] 收到更新进度: client_id=%d, type=%s, stage=%s, progress=%d%%", clientID, updateType, stage, progress)
		if s.realtimeService != nil {
			s.realtimeService.BroadcastUpdateProgress(clientID, updateType, stage, progress, message, totalBytes, downloadedBytes)
		}
	})

	// 设置更新结果回调
	websocket.ClientDaemonHubInstance.SetUpdateResultCallback(func(clientID uint, updateType string, success bool, version string, message string) {
		log.Printf("[ClientUpdateService] 收到更新结果: client_id=%d, type=%s, success=%v, version=%s", clientID, updateType, success, version)
		if s.realtimeService != nil {
			s.realtimeService.BroadcastUpdateResult(clientID, updateType, success, version, message)
		}

		// 更新成功后，更新数据库中的版本信息
		if success && version != "" {
			if updateType == string(UpdateTypeFrpc) {
				s.clientRepo.UpdateVersionInfo(clientID, version, "", "", "")
			} else if updateType == string(UpdateTypeDaemon) {
				s.clientRepo.UpdateVersionInfo(clientID, "", version, "", "")
			}
		}
	})

	// 设置版本上报回调
	websocket.ClientDaemonHubInstance.SetVersionReportCallback(func(clientID uint, frpcVersion string, daemonVersion string, os string, arch string) {
		log.Printf("[ClientUpdateService] 收到版本上报: client_id=%d, frpc=%s, daemon=%s, os=%s, arch=%s", clientID, frpcVersion, daemonVersion, os, arch)
		s.clientRepo.UpdateVersionInfo(clientID, frpcVersion, daemonVersion, os, arch)
	})

	log.Printf("[ClientUpdateService] WebSocket回调函数已设置")
}

// UpdateClient 更新单个客户端
func (s *ClientUpdateService) UpdateClient(req *UpdateRequest) error {
	// 检查客户端是否存在
	client, err := s.clientRepo.FindByID(req.ClientID)
	if err != nil {
		return fmt.Errorf("客户端不存在: %v", err)
	}

	// 检查客户端是否在线
	if !websocket.ClientDaemonHubInstance.IsClientOnline(req.ClientID) {
		return fmt.Errorf("客户端 %s 不在线", client.Name)
	}

	// 获取目标版本和下载URL
	version, downloadURL, err := s.getUpdateInfoForClient(client, req.UpdateType, req.Version, req.MirrorID)
	if err != nil {
		return err
	}

	// 获取镜像ID
	mirrorID := uint(0)
	if req.MirrorID != nil {
		mirrorID = *req.MirrorID
	}

	// 发送更新命令
	err = websocket.ClientDaemonHubInstance.SendUpdateCommand(req.ClientID, string(req.UpdateType), version, downloadURL, mirrorID)
	if err != nil {
		return fmt.Errorf("发送更新命令失败: %v", err)
	}

	log.Printf("[ClientUpdateService] ✅ 已向客户端 %s (ID=%d) 发送更新命令: type=%s, version=%s", client.Name, req.ClientID, req.UpdateType, version)
	return nil
}

// BatchUpdateClients 批量更新客户端
func (s *ClientUpdateService) BatchUpdateClients(req *BatchUpdateRequest) (successCount int, failedClients []string, err error) {
	if len(req.ClientIDs) == 0 {
		return 0, nil, fmt.Errorf("未指定要更新的客户端")
	}

	// 获取所有客户端信息
	clients, err := s.clientRepo.FindByIDs(req.ClientIDs)
	if err != nil {
		return 0, nil, fmt.Errorf("获取客户端信息失败: %v", err)
	}

	for _, client := range clients {
		updateReq := &UpdateRequest{
			ClientID:   client.ID,
			UpdateType: req.UpdateType,
			Version:    req.Version,
			MirrorID:   req.MirrorID,
		}

		if err := s.UpdateClient(updateReq); err != nil {
			log.Printf("[ClientUpdateService] ❌ 更新客户端 %s 失败: %v", client.Name, err)
			failedClients = append(failedClients, fmt.Sprintf("%s: %v", client.Name, err))
		} else {
			successCount++
		}
	}

	return successCount, failedClients, nil
}

// getUpdateInfoForClient 获取客户端的更新信息（内部使用，接收client对象）
func (s *ClientUpdateService) getUpdateInfoForClient(client *model.Client, updateType UpdateType, specifiedVersion string, mirrorID *uint) (version string, downloadURL string, err error) {
	switch updateType {
	case UpdateTypeFrpc:
		return s.getFrpcUpdateInfo(client.FrpServerID, specifiedVersion, mirrorID, client.OS, client.Arch)
	case UpdateTypeDaemon:
		return s.getDaemonUpdateInfo(client.OS, client.Arch)
	default:
		return "", "", fmt.Errorf("不支持的更新类型: %s", updateType)
	}
}

// GetUpdateInfoForClientByID 根据客户端ID获取更新信息（外部API使用）
func (s *ClientUpdateService) GetUpdateInfoForClientByID(clientID uint, updateType UpdateType, specifiedVersion string, mirrorID *uint) (version string, downloadURL string, err error) {
	client, err := s.clientRepo.FindByID(clientID)
	if err != nil {
		return "", "", fmt.Errorf("客户端不存在: %v", err)
	}

	return s.getUpdateInfoForClient(client, updateType, specifiedVersion, mirrorID)
}

// getFrpcUpdateInfo 获取frpc更新信息
func (s *ClientUpdateService) getFrpcUpdateInfo(frpServerID *uint, specifiedVersion string, mirrorID *uint, clientOS string, clientArch string) (version string, downloadURL string, err error) {
	// 确定版本
	if specifiedVersion != "" {
		version = specifiedVersion
	} else if frpServerID != nil {
		// 从关联的frps服务器获取版本
		server, err := s.frpServerRepo.GetByID(*frpServerID)
		if err != nil {
			return "", "", fmt.Errorf("获取关联FRP服务器失败: %v", err)
		}
		version = server.Version
		if version == "" {
			return "", "", fmt.Errorf("关联的FRP服务器未设置版本信息")
		}
	} else {
		return "", "", fmt.Errorf("未指定版本且客户端未关联FRP服务器")
	}

	// 确保版本号以v开头
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	// 构建下载URL
	// 格式: https://github.com/fatedier/frp/releases/download/v0.52.0/frp_0.52.0_linux_amd64.tar.gz
	versionNum := strings.TrimPrefix(version, "v")

	// 确定操作系统和架构
	osName := clientOS
	arch := clientArch
	if osName == "" {
		osName = "linux" // 默认
	}
	if arch == "" {
		arch = "amd64" // 默认
	}

	// 确定文件扩展名
	ext := "tar.gz"
	if osName == "windows" {
		ext = "zip"
	}

	downloadURL = fmt.Sprintf("https://github.com/fatedier/frp/releases/download/%s/frp_%s_%s_%s.%s", version, versionNum, osName, arch, ext)

	// 如果指定了镜像源，转换URL
	if mirrorID != nil {
		mirrorService := NewGithubMirrorService()
		convertedURL, err := mirrorService.ConvertGithubURL(downloadURL, mirrorID)
		if err != nil {
			log.Printf("[ClientUpdateService] ⚠️ 转换镜像URL失败: %v，使用原始URL", err)
		} else {
			downloadURL = convertedURL
		}
	}

	return version, downloadURL, nil
}

// getDaemonUpdateInfo 获取daemon更新信息
func (s *ClientUpdateService) getDaemonUpdateInfo(clientOS string, clientArch string) (version string, downloadURL string, err error) {
	// daemon从服务端下载，URL格式: /download/daemon/{os}/{arch}
	osName := clientOS
	arch := clientArch
	if osName == "" {
		osName = "linux"
	}
	if arch == "" {
		arch = "amd64"
	}

	// 这里返回相对路径，daemon会自动拼接服务器地址
	downloadURL = fmt.Sprintf("/download/daemon/%s/%s", osName, arch)
	version = "latest" // daemon版本由服务端决定

	return version, downloadURL, nil
}

// GetOnlineClients 获取所有在线客户端
func (s *ClientUpdateService) GetOnlineClients() ([]uint, error) {
	return websocket.ClientDaemonHubInstance.GetOnlineClientIDs(), nil
}

// GetClientVersions 获取客户端版本信息
func (s *ClientUpdateService) GetClientVersions(clientID uint) (frpcVersion string, daemonVersion string, os string, arch string, err error) {
	client, err := s.clientRepo.FindByID(clientID)
	if err != nil {
		return "", "", "", "", err
	}
	return client.FrpcVersion, client.DaemonVersion, client.OS, client.Arch, nil
}
