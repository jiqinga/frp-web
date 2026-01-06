package service

import (
	"fmt"
	"frp-web-panel/internal/frp"
	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"frp-web-panel/pkg/database"
	"log"
	"strconv"
	"sync"
	"time"
)

// proxyTrafficSample 隧道流量采样数据
type proxyTrafficSample struct {
	TrafficIn  int64
	TrafficOut int64
	Timestamp  time.Time
}

type MetricsCollector struct {
	serverRepo       *repository.FrpServerRepository
	metricsRepo      *repository.ServerMetricsRepository
	proxyMetricsRepo *repository.ProxyMetricsRepository
	settingRepo      *repository.SettingRepository
	stopChan         chan struct{}
	interval         time.Duration
	ticker           *time.Ticker
	mutex            sync.RWMutex

	// 内存缓存：存储上一次采样值，用于计算速率
	// key: "serverID:proxyName"
	trafficCache sync.Map
}

func NewMetricsCollector(serverRepo *repository.FrpServerRepository) *MetricsCollector {
	settingRepo := repository.NewSettingRepository()
	// 从数据库读取采集间隔，默认30秒
	interval := 30 * time.Second
	if val, err := settingRepo.GetSetting("traffic_interval"); err == nil {
		if seconds, err := strconv.Atoi(val); err == nil && seconds > 0 {
			interval = time.Duration(seconds) * time.Second
		}
	}

	return &MetricsCollector{
		serverRepo:       serverRepo,
		metricsRepo:      repository.NewServerMetricsRepository(),
		proxyMetricsRepo: repository.NewProxyMetricsRepository(),
		settingRepo:      settingRepo,
		stopChan:         make(chan struct{}),
		interval:         interval,
	}
}

// Start 启动指标采集和数据清理
func (c *MetricsCollector) Start() {
	go c.collectLoop()
	go c.cleanupLoop()
	log.Printf("[MetricsCollector] 指标采集服务已启动，采集间隔: %v", c.interval)
}

// Stop 停止采集
func (c *MetricsCollector) Stop() {
	close(c.stopChan)
}

// UpdateInterval 动态更新采集间隔
func (c *MetricsCollector) UpdateInterval(seconds int) {
	if seconds < 5 {
		seconds = 5
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.interval = time.Duration(seconds) * time.Second
	if c.ticker != nil {
		c.ticker.Reset(c.interval)
	}
	log.Printf("[MetricsCollector] 采集间隔已更新为: %v", c.interval)
}

func (c *MetricsCollector) collectLoop() {
	c.mutex.Lock()
	c.ticker = time.NewTicker(c.interval)
	c.mutex.Unlock()
	defer c.ticker.Stop()

	// 启动时立即执行一次
	c.collectAll()

	for {
		select {
		case <-c.ticker.C:
			c.collectAll()
		case <-c.stopChan:
			return
		}
	}
}

func (c *MetricsCollector) collectAll() {
	servers, err := c.serverRepo.GetAll()
	if err != nil {
		log.Printf("[MetricsCollector] 获取服务器列表失败: %v", err)
		return
	}

	for _, server := range servers {
		if server.Status != model.StatusRunning {
			continue
		}
		go c.collectOne(&server)
	}
}

func (c *MetricsCollector) collectOne(server *model.FrpServer) {
	host := server.Host
	if server.ServerType == model.ServerTypeRemote && (host == "" || host == "0.0.0.0") {
		host = server.SSHHost
	}

	client := frp.NewFrpsClient(host, server.DashboardPort, server.DashboardUser, server.DashboardPwd)
	metrics, err := client.GetMetrics()
	if err != nil {
		log.Printf("[MetricsCollector] 采集服务器 %d 指标失败: %v", server.ID, err)
		return
	}

	now := time.Now()

	// 计算CPU百分比
	cpuPercent := 0.0
	if metrics.Uptime > 0 {
		cpuPercent = (metrics.CpuSeconds / float64(metrics.Uptime)) * 100
	}

	// 保存服务器级别指标
	record := &model.ServerMetricsHistory{
		ServerID:    server.ID,
		CpuPercent:  cpuPercent,
		MemoryBytes: metrics.MemoryBytes,
		TrafficIn:   metrics.TrafficIn,
		TrafficOut:  metrics.TrafficOut,
		RecordTime:  now,
	}

	if err := c.metricsRepo.Create(record); err != nil {
		log.Printf("[MetricsCollector] 保存服务器 %d 指标失败: %v", server.ID, err)
	}

	// 处理每个隧道的流量数据
	c.processProxyTraffics(server.ID, metrics.ProxyTraffics, now)
}

func (c *MetricsCollector) cleanupLoop() {
	// 计算到下一个凌晨3点的时间
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
	if now.After(next) {
		next = next.Add(24 * time.Hour)
	}

	timer := time.NewTimer(next.Sub(now))
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			c.cleanup()
			timer.Reset(24 * time.Hour)
		case <-c.stopChan:
			return
		}
	}
}

func (c *MetricsCollector) cleanup() {
	before := time.Now().AddDate(0, 0, -7)

	// 清理服务器指标历史
	deleted, err := c.metricsRepo.DeleteOlderThan(before)
	if err != nil {
		log.Printf("[MetricsCollector] 清理服务器历史数据失败: %v", err)
	} else {
		log.Printf("[MetricsCollector] 已清理 %d 条服务器过期指标数据", deleted)
	}

	// 清理隧道指标历史
	deletedProxy, err := c.proxyMetricsRepo.DeleteOlderThan(before)
	if err != nil {
		log.Printf("[MetricsCollector] 清理隧道历史数据失败: %v", err)
	} else {
		log.Printf("[MetricsCollector] 已清理 %d 条隧道过期指标数据", deletedProxy)
	}
}

// processProxyTraffics 处理每个隧道的流量数据，计算速率并保存
func (c *MetricsCollector) processProxyTraffics(serverID uint, traffics []frp.ProxyTrafficData, now time.Time) {
	if len(traffics) == 0 {
		return
	}

	intervalSecs := int64(c.interval.Seconds())
	if intervalSecs <= 0 {
		intervalSecs = 30
	}

	var proxyMetrics []model.ProxyMetricsHistory

	for _, pt := range traffics {
		cacheKey := fmt.Sprintf("%d:%s", serverID, pt.Name)

		var rateIn, rateOut int64

		// 从缓存获取上次采样值
		if lastVal, ok := c.trafficCache.Load(cacheKey); ok {
			last := lastVal.(*proxyTrafficSample)
			elapsed := now.Sub(last.Timestamp).Seconds()
			if elapsed > 0 {
				rateIn = int64(float64(pt.TrafficIn-last.TrafficIn) / elapsed)
				rateOut = int64(float64(pt.TrafficOut-last.TrafficOut) / elapsed)
				// 防止负值（服务器重启等情况）
				if rateIn < 0 {
					rateIn = 0
				}
				if rateOut < 0 {
					rateOut = 0
				}
			}
		}

		// 更新缓存
		c.trafficCache.Store(cacheKey, &proxyTrafficSample{
			TrafficIn:  pt.TrafficIn,
			TrafficOut: pt.TrafficOut,
			Timestamp:  now,
		})

		// 构建隧道指标记录
		proxyMetrics = append(proxyMetrics, model.ProxyMetricsHistory{
			ServerID:   serverID,
			ProxyName:  pt.Name,
			ProxyType:  pt.Type,
			TrafficIn:  pt.TrafficIn,
			TrafficOut: pt.TrafficOut,
			RateIn:     rateIn,
			RateOut:    rateOut,
			RecordTime: now,
		})

		// 更新 Proxy 表的实时速率（通过名称匹配）
		c.updateProxyRate(pt.Name, rateIn, rateOut)
	}

	// 批量保存隧道指标
	if err := c.proxyMetricsRepo.BatchCreate(proxyMetrics); err != nil {
		log.Printf("[MetricsCollector] 保存隧道指标失败: %v", err)
	}
}

// updateProxyRate 更新 Proxy 表的实时速率
func (c *MetricsCollector) updateProxyRate(proxyName string, rateIn, rateOut int64) {
	err := database.DB.Model(&model.Proxy{}).
		Where("name = ?", proxyName).
		Updates(map[string]interface{}{
			"current_bytes_in_rate":  rateIn,
			"current_bytes_out_rate": rateOut,
			"last_traffic_update":    time.Now(),
		}).Error
	if err != nil {
		log.Printf("[MetricsCollector] 更新隧道 %s 速率失败: %v", proxyName, err)
	}
}
