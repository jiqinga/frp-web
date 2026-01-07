package service

import (
	"fmt"
	"frp-web-panel/internal/frp"
	"frp-web-panel/internal/logger"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"frp-web-panel/pkg/database"
	"strconv"
	"sync"
	"time"
)

// proxyTrafficCache 存储上次从 frps 获取的流量值，用于计算增量和检测重启
type proxyTrafficCache struct {
	TrafficIn  int64
	TrafficOut int64
}

type FrpSyncService struct {
	frpServerRepo *repository.FrpServerRepository
	clientRepo    *repository.ClientRepository
	proxyRepo     *repository.ProxyRepository
	settingRepo   *repository.SettingRepository
	stopChan      chan struct{}
	// 内存缓存：存储上次从 frps 获取的流量值
	// key: "serverID:proxyName"
	trafficCache sync.Map
}

func NewFrpSyncService() *FrpSyncService {
	return &FrpSyncService{
		frpServerRepo: repository.NewFrpServerRepository(database.DB),
		clientRepo:    repository.NewClientRepository(),
		proxyRepo:     repository.NewProxyRepository(),
		settingRepo:   repository.NewSettingRepository(),
		stopChan:      make(chan struct{}),
	}
}

func (s *FrpSyncService) Start() {
	// 从数据库读取间隔配置，默认值
	serverInfoInterval := s.getIntervalSetting("server_info_interval", 5)
	proxyStatusInterval := s.getIntervalSetting("proxy_status_interval", 10)

	go s.syncLoop(serverInfoInterval, s.syncServerInfo)
	go s.syncLoop(proxyStatusInterval, s.syncProxyStatus)

	logger.Info("FRP同步服务已启动")
}

func (s *FrpSyncService) getIntervalSetting(key string, defaultVal int) time.Duration {
	if val, err := s.settingRepo.GetSetting(key); err == nil {
		if seconds, err := strconv.Atoi(val); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return time.Duration(defaultVal) * time.Second
}

func (s *FrpSyncService) Stop() {
	close(s.stopChan)
}

func (s *FrpSyncService) syncLoop(interval time.Duration, syncFunc func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	syncFunc()

	for {
		select {
		case <-ticker.C:
			syncFunc()
		case <-s.stopChan:
			return
		}
	}
}

func (s *FrpSyncService) syncServerInfo() {
	servers, err := s.frpServerRepo.GetEnabled()
	if err != nil {
		logger.Errorf("获取已启用的FRP服务器失败: %v", err)
		return
	}

	logger.Debugf("同步服务获取到 %d 个启用的服务器", len(servers))
	for _, server := range servers {
		logger.Debugf("服务器 %s: enabled=%v, status=%s", server.Name, server.Enabled, server.Status)

		// 跳过已停止的服务器
		if server.Status == model.StatusStopped || server.Status == model.StatusStopping {
			logger.Debugf("跳过已停止的服务器: %s (status=%s)", server.Name, server.Status)
			continue
		}

		if err := s.syncSingleServer(&server); err != nil {
			logger.Errorf("同步服务器 %s 失败: %v", server.Name, err)
			server.LastError = err.Error()
		} else {
			server.LastError = ""
		}
		now := time.Now()
		server.LastSyncTime = &now
		s.frpServerRepo.Update(&server)
	}
}

func (s *FrpSyncService) syncSingleServer(server *model.FrpServer) error {
	// 对于远程服务器,使用SSHHost而不是Host
	host := server.Host
	if server.ServerType == model.ServerTypeRemote && server.SSHHost != "" {
		host = server.SSHHost
	}

	logger.Debugf("同步服务器 %s: type=%s, host=%s, dashboard_port=%d, user=%s",
		server.Name, server.ServerType, host, server.DashboardPort, server.DashboardUser)

	client := frp.NewFrpsClient(host, server.DashboardPort, server.DashboardUser, server.DashboardPwd)

	if err := client.HealthCheck(); err != nil {
		logger.Errorf("服务器 %s 健康检查失败: %v", server.Name, err)
		return err
	}

	info, err := client.GetServerInfo()
	if err != nil {
		return err
	}

	logger.Infof("服务器 %s: 客户端数=%d, 连接数=%d", server.Name, info.ClientCounts, info.CurConns)
	return nil
}

func (s *FrpSyncService) syncProxyStatus() {
	servers, err := s.frpServerRepo.GetEnabled()
	if err != nil {
		return
	}

	logger.Debugf("代理同步获取到 %d 个启用的服务器", len(servers))
	for _, server := range servers {
		logger.Debugf("代理同步 - 服务器 %s: enabled=%v, status=%s", server.Name, server.Enabled, server.Status)

		// 跳过已停止的服务器
		if server.Status == model.StatusStopped || server.Status == model.StatusStopping {
			logger.Debugf("跳过已停止的服务器代理同步: %s (status=%s)", server.Name, server.Status)
			continue
		}

		s.syncServerProxies(&server)
	}
}

func (s *FrpSyncService) syncServerProxies(server *model.FrpServer) {
	// 对于远程服务器,使用SSHHost而不是Host
	host := server.Host
	if server.ServerType == model.ServerTypeRemote && server.SSHHost != "" {
		host = server.SSHHost
	}

	logger.Debugf("syncServerProxies 开始同步服务器 %s (ID:%d) 的代理", server.Name, server.ID)

	// 首先获取关联到该服务器的所有客户端
	clients, err := s.clientRepo.FindByFrpServerID(server.ID)
	if err != nil {
		logger.Debugf("syncServerProxies 获取服务器 %s 关联的客户端失败: %v", server.Name, err)
		return
	}
	logger.Debugf("syncServerProxies 服务器 %s 关联了 %d 个客户端", server.Name, len(clients))

	if len(clients) == 0 {
		logger.Debugf("syncServerProxies 服务器 %s 没有关联的客户端，跳过同步", server.Name)
		return
	}

	// 构建客户端ID到客户端名称的映射
	clientIDToName := make(map[uint]string)
	clientIDs := make([]uint, len(clients))
	for i, c := range clients {
		clientIDs[i] = c.ID
		clientIDToName[c.ID] = c.Name
		logger.Debugf("syncServerProxies   - 客户端: %s (ID:%d)", c.Name, c.ID)
	}

	// 获取这些客户端的代理
	var dbProxies []model.Proxy
	for _, clientID := range clientIDs {
		proxies, err := s.proxyRepo.FindByClientID(clientID)
		if err != nil {
			logger.Debugf("syncServerProxies 获取客户端 %d 的代理失败: %v", clientID, err)
			continue
		}
		dbProxies = append(dbProxies, proxies...)
	}
	logger.Debugf("syncServerProxies 数据库中这些客户端共有 %d 个代理", len(dbProxies))

	// 构建多种格式的代理名称到代理的映射
	// FRP 服务器上的代理名称格式可能是:
	// 1. 直接使用代理名称: "proxy_name"
	// 2. 带客户端前缀: "client_name.proxy_name"
	proxyMap := make(map[string]*model.Proxy)
	for i := range dbProxies {
		proxy := &dbProxies[i]
		clientName := clientIDToName[proxy.ClientID]

		// 添加直接名称映射
		proxyMap[proxy.Name] = proxy

		// 添加带客户端前缀的名称映射 (client_name.proxy_name)
		fullName := clientName + "." + proxy.Name
		proxyMap[fullName] = proxy

		logger.Debugf("syncServerProxies   - 数据库代理: %s (ID:%d, ClientID:%d, 全名:%s)",
			proxy.Name, proxy.ID, proxy.ClientID, fullName)
	}

	// 从 FRP 服务器获取代理状态
	client := frp.NewFrpsClient(host, server.DashboardPort, server.DashboardUser, server.DashboardPwd)

	// 使用 GetAllProxies 获取所有类型的代理
	allProxies, err := client.GetAllProxies()
	if err != nil {
		logger.Errorf("从服务器 %s 获取所有代理失败: %v", server.Name, err)
		return
	}

	logger.Debugf("syncServerProxies 服务器 %s 获取到代理类型分布:", server.Name)
	totalProxies := 0
	for proxyType, proxies := range allProxies {
		logger.Debugf("syncServerProxies   - %s: %d 个代理", proxyType, len(proxies))
		totalProxies += len(proxies)
	}
	logger.Debugf("syncServerProxies FRP服务器返回总计: %d 个代理", totalProxies)

	updatedCount := 0
	onlineCount := 0
	for proxyType, proxies := range allProxies {
		for _, proxyInfo := range proxies {
			// 尝试匹配代理名称（支持直接名称和带前缀的名称）
			proxy, exists := proxyMap[proxyInfo.Name]
			if !exists {
				logger.Debugf("syncServerProxies 代理 %s (类型:%s) 不属于该服务器的客户端，跳过", proxyInfo.Name, proxyType)
				continue
			}

			proxy.FrpStatus = proxyInfo.Status
			proxy.FrpCurConns = proxyInfo.CurConns

			// 计算流量增量（处理 frps 重启的情况）
			deltaIn, deltaOut := s.calculateTrafficDelta(server.ID, proxyInfo.Name, proxyInfo.TodayTrafficIn, proxyInfo.TodayTrafficOut)
			proxy.TotalBytesIn += deltaIn
			proxy.TotalBytesOut += deltaOut

			now := time.Now()
			proxy.LastTrafficUpdate = &now

			if proxyInfo.Status == "online" {
				proxy.LastOnlineTime = &now
				onlineCount++

				// 获取历史流量数据计算速率
				trafficData, err := client.GetProxyTraffic(proxyInfo.Name)
				if err != nil {
					logger.Debugf("syncServerProxies 获取代理 %s 流量数据失败: %v", proxyInfo.Name, err)
				} else {
					// 计算速率：使用最后两个数据点的差值
					inRate, outRate := calculateRateFromHistory(trafficData.TrafficIn, trafficData.TrafficOut)
					proxy.CurrentBytesInRate = inRate
					proxy.CurrentBytesOutRate = outRate
					logger.Debugf("syncServerProxies 代理 %s 速率: in=%d B/s, out=%d B/s", proxyInfo.Name, inRate, outRate)
				}
			} else {
				// 离线代理速率为0
				proxy.CurrentBytesInRate = 0
				proxy.CurrentBytesOutRate = 0
			}

			s.proxyRepo.Update(proxy)
			updatedCount++
			logger.Debugf("syncServerProxies 更新代理 %s -> %s: 类型=%s, 状态=%s, LastOnlineTime=%v",
				proxyInfo.Name, proxy.Name, proxyType, proxyInfo.Status, proxy.LastOnlineTime)
		}
	}
	logger.Debugf("syncServerProxies 服务器 %s 同步完成: 更新了 %d 个代理, 其中 %d 个在线",
		server.Name, updatedCount, onlineCount)
}

// calculateRateFromHistory 从历史流量数据计算速率
// FRP 的 /api/traffic 接口返回的是每分钟的流量数据数组
func calculateRateFromHistory(trafficIn, trafficOut []int64) (int64, int64) {
	var inRate, outRate int64

	// 使用最后两个数据点计算速率
	if len(trafficIn) >= 2 {
		diff := trafficIn[len(trafficIn)-1] - trafficIn[len(trafficIn)-2]
		if diff > 0 {
			inRate = diff / 60 // 每分钟数据，转换为每秒
		}
	} else if len(trafficIn) >= 1 {
		inRate = trafficIn[len(trafficIn)-1] / 60
	}

	if len(trafficOut) >= 2 {
		diff := trafficOut[len(trafficOut)-1] - trafficOut[len(trafficOut)-2]
		if diff > 0 {
			outRate = diff / 60
		}
	} else if len(trafficOut) >= 1 {
		outRate = trafficOut[len(trafficOut)-1] / 60
	}

	return inRate, outRate
}

// calculateTrafficDelta 计算流量增量，处理 frps 重启的情况
// 当检测到当前值 < 上次记录值时，说明 frps 重启了，此时增量 = 当前值
func (s *FrpSyncService) calculateTrafficDelta(serverID uint, proxyName string, currentIn, currentOut int64) (deltaIn, deltaOut int64) {
	cacheKey := fmt.Sprintf("%d:%s", serverID, proxyName)

	// 从缓存获取上次的流量值
	if lastVal, ok := s.trafficCache.Load(cacheKey); ok {
		last := lastVal.(*proxyTrafficCache)

		// 计算增量
		deltaIn = currentIn - last.TrafficIn
		deltaOut = currentOut - last.TrafficOut

		// 检测 frps 重启：当前值 < 上次值
		if deltaIn < 0 {
			logger.Infof("TrafficDelta 检测到 frps 重启 (proxy=%s): TrafficIn 从 %d 变为 %d，使用当前值作为增量",
				proxyName, last.TrafficIn, currentIn)
			deltaIn = currentIn
		}
		if deltaOut < 0 {
			logger.Infof("TrafficDelta 检测到 frps 重启 (proxy=%s): TrafficOut 从 %d 变为 %d，使用当前值作为增量",
				proxyName, last.TrafficOut, currentOut)
			deltaOut = currentOut
		}
	} else {
		// 首次采集，不累加（避免把历史累计值当作增量）
		deltaIn = 0
		deltaOut = 0
		logger.Debugf("TrafficDelta 首次采集 proxy=%s: 记录基准值 in=%d, out=%d", proxyName, currentIn, currentOut)
	}

	// 更新缓存
	s.trafficCache.Store(cacheKey, &proxyTrafficCache{
		TrafficIn:  currentIn,
		TrafficOut: currentOut,
	})

	return deltaIn, deltaOut
}
