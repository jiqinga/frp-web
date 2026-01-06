/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-17 16:19:07
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-30 14:49:16
 * @FilePath            : frp-web-testbackendinternalservicerealtime_service.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package service

import (
	"frp-web-panel/internal/events"
	"frp-web-panel/internal/repository"
	"frp-web-panel/pkg/database"
	"log"
	"strconv"
	"sync"
	"time"
)

type RealtimeService struct {
	eventBus         *events.EventBus
	proxyRepo        *repository.ProxyRepository
	trafficRepo      *repository.TrafficRepository
	clientRepo       *repository.ClientRepository
	frpServerService *FrpServerService
	settingRepo      *repository.SettingRepository
	trafficTicker    *time.Ticker
	statusTicker     *time.Ticker
	statusMutex      sync.RWMutex
}

func NewRealtimeService() *RealtimeService {
	return &RealtimeService{
		eventBus:         events.GetEventBus(),
		proxyRepo:        repository.NewProxyRepository(),
		trafficRepo:      repository.NewTrafficRepository(),
		clientRepo:       repository.NewClientRepository(),
		frpServerService: NewFrpServerService(),
		settingRepo:      repository.NewSettingRepository(),
	}
}

func (s *RealtimeService) Start() {
	log.Printf("[DEBUG RealtimeService] 启动实时服务")
	s.trafficTicker = time.NewTicker(1 * time.Second)
	go s.collectTrafficData()
	log.Printf("[DEBUG RealtimeService] 启动流量数据收集")
	go s.startServerStatusMonitor()
	log.Printf("[DEBUG RealtimeService] 启动服务器状态监控")
}

func (s *RealtimeService) Stop() {
	if s.trafficTicker != nil {
		s.trafficTicker.Stop()
	}
	if s.statusTicker != nil {
		s.statusTicker.Stop()
	}
}

func (s *RealtimeService) collectTrafficData() {
	for range s.trafficTicker.C {
		proxies, err := s.proxyRepo.FindAll()
		if err != nil {
			log.Printf("获取代理列表失败: %v", err)
			continue
		}

		// 获取所有客户端，建立 client_id -> client_name 的映射
		clients, err := s.clientRepo.GetAllForStatusCheck()
		if err != nil {
			log.Printf("获取客户端列表失败: %v", err)
			// 即使获取客户端失败，也继续处理，只是没有客户端名称
		}
		clientNameMap := make(map[uint]string)
		for _, client := range clients {
			clientNameMap[client.ID] = client.Name
		}

		var trafficData []map[string]interface{}
		onlineCount := 0
		offlineCount := 0
		nilTimeCount := 0
		for _, proxy := range proxies {
			online := false
			if proxy.LastOnlineTime != nil {
				timeSince := time.Since(*proxy.LastOnlineTime)
				online = timeSince < 2*time.Minute
				if online {
					onlineCount++
				} else {
					offlineCount++
					// 只打印前几个离线代理的详细信息，避免日志过多
					if offlineCount <= 3 {
						log.Printf("[DEBUG collectTrafficData] 代理 %s (ID:%d) 离线: LastOnlineTime=%v, 距今=%v",
							proxy.Name, proxy.ID, proxy.LastOnlineTime, timeSince)
					}
				}
			} else {
				nilTimeCount++
				offlineCount++
				// 只打印前几个没有 LastOnlineTime 的代理
				if nilTimeCount <= 3 {
					log.Printf("[DEBUG collectTrafficData] 代理 %s (ID:%d) LastOnlineTime 为 nil",
						proxy.Name, proxy.ID)
				}
			}

			// 获取客户端名称，如果不存在则使用空字符串（前端会显示默认名称）
			clientName := clientNameMap[proxy.ClientID]

			trafficData = append(trafficData, map[string]interface{}{
				"proxy_id":        proxy.ID,
				"proxy_name":      proxy.Name,
				"client_id":       proxy.ClientID,
				"client_name":     clientName,
				"bytes_in_rate":   proxy.CurrentBytesInRate,
				"bytes_out_rate":  proxy.CurrentBytesOutRate,
				"total_bytes_in":  proxy.TotalBytesIn,
				"total_bytes_out": proxy.TotalBytesOut,
				"online":          online,
			})
		}

		// 每10秒打印一次统计信息
		log.Printf("[DEBUG collectTrafficData] 代理统计: 总数=%d, 在线=%d, 离线=%d (其中 LastOnlineTime 为 nil: %d)",
			len(proxies), onlineCount, offlineCount, nilTimeCount)

		// 通过事件总线发布流量更新事件
		s.eventBus.Publish(events.TrafficUpdateEvent{
			Timestamp: time.Now(),
			Data:      trafficData,
		})
	}
}

func (s *RealtimeService) startServerStatusMonitor() {
	interval := s.getCheckInterval()
	s.statusTicker = time.NewTicker(interval)

	for range s.statusTicker.C {
		s.checkServerStatus()
	}
}

func (s *RealtimeService) getCheckInterval() time.Duration {
	var value string
	if err := database.DB.Table("settings").Where("key = ?", "server_status_check_interval").Pluck("value", &value).Error; err != nil {
		return 10 * time.Second
	}

	seconds, err := strconv.Atoi(value)
	if err != nil || seconds < 5 {
		return 10 * time.Second
	}
	return time.Duration(seconds) * time.Second
}

func (s *RealtimeService) checkServerStatus() {
	log.Printf("[DEBUG checkServerStatus] 开始检查服务器状态")
	frpServerRepo := repository.NewFrpServerRepository(database.DB)
	servers, err := frpServerRepo.GetEnabled()
	if err != nil {
		log.Printf("获取已启用服务器失败: %v", err)
		return
	}
	log.Printf("[DEBUG checkServerStatus] 找到 %d 个启用的服务器", len(servers))

	var statusUpdates []map[string]interface{}
	for _, server := range servers {
		oldStatus := server.Status
		log.Printf("[DEBUG checkServerStatus] 服务器ID=%d, 数据库状态=%s", server.ID, oldStatus)

		newStatus, err := s.frpServerService.GetStatus(server.ID)
		if err != nil {
			log.Printf("[DEBUG checkServerStatus] 获取服务器ID=%d状态失败: %v", server.ID, err)
			continue
		}
		log.Printf("[DEBUG checkServerStatus] 服务器ID=%d, 实际状态=%s", server.ID, newStatus)

		if oldStatus != newStatus {
			log.Printf("[DEBUG checkServerStatus] 检测到状态变化: 服务器ID=%d, %s -> %s", server.ID, oldStatus, newStatus)
			statusUpdates = append(statusUpdates, map[string]interface{}{
				"server_id":       server.ID,
				"server_name":     server.Name,
				"status":          string(newStatus),
				"previous_status": string(oldStatus),
			})
		} else {
			log.Printf("[DEBUG checkServerStatus] 服务器ID=%d 状态未变化: %s", server.ID, oldStatus)
		}
	}

	if len(statusUpdates) > 0 {
		log.Printf("[DEBUG checkServerStatus] 准备推送 %d 个状态更新", len(statusUpdates))
		// 通过事件总线发布服务器状态更新
		for _, update := range statusUpdates {
			s.eventBus.Publish(events.ServerStatusEvent{
				ServerID:   update["server_id"].(uint),
				ServerName: update["server_name"].(string),
				Status:     update["status"].(string),
			})
		}
	} else {
		log.Printf("[DEBUG checkServerStatus] 没有状态变化，不推送")
	}
}

func (s *RealtimeService) UpdateCheckInterval(seconds int) {
	s.statusMutex.Lock()
	defer s.statusMutex.Unlock()

	if s.statusTicker != nil {
		s.statusTicker.Stop()
	}

	if seconds < 5 {
		seconds = 5
	}
	s.statusTicker = time.NewTicker(time.Duration(seconds) * time.Second)
	go func() {
		for range s.statusTicker.C {
			s.checkServerStatus()
		}
	}()
}

// BroadcastUpdateProgress 广播客户端更新进度
func (s *RealtimeService) BroadcastUpdateProgress(clientID uint, updateType string, stage string, progress int, message string, totalBytes int64, downloadedBytes int64) {
	log.Printf("[RealtimeService] 广播更新进度: client_id=%d, type=%s, stage=%s, progress=%d%%", clientID, updateType, stage, progress)
	s.eventBus.Publish(events.UpdateProgressEvent{
		ClientID:        clientID,
		UpdateType:      updateType,
		Stage:           stage,
		Progress:        progress,
		Message:         message,
		TotalBytes:      totalBytes,
		DownloadedBytes: downloadedBytes,
	})
}

// BroadcastUpdateResult 广播客户端更新结果
func (s *RealtimeService) BroadcastUpdateResult(clientID uint, updateType string, success bool, version string, message string) {
	log.Printf("[RealtimeService] 广播更新结果: client_id=%d, type=%s, success=%v, version=%s", clientID, updateType, success, version)
	s.eventBus.Publish(events.UpdateResultEvent{
		ClientID:   clientID,
		UpdateType: updateType,
		Success:    success,
		Version:    version,
		Message:    message,
	})
}
