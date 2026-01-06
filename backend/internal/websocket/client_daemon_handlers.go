package websocket

import "log"

// HandleUpdateProgress 处理更新进度消息
func (h *ClientDaemonHub) HandleUpdateProgress(clientID uint, data map[string]interface{}) {
	h.mu.RLock()
	callback := h.updateProgressCallback
	h.mu.RUnlock()

	if callback == nil {
		log.Printf("[ClientDaemonHub] 更新进度回调未设置，忽略消息")
		return
	}

	updateType, _ := data["update_type"].(string)
	stage, _ := data["stage"].(string)
	progress := int(data["progress"].(float64))
	message, _ := data["message"].(string)
	totalBytes := int64(0)
	downloadedBytes := int64(0)
	if tb, ok := data["total_bytes"].(float64); ok {
		totalBytes = int64(tb)
	}
	if db, ok := data["downloaded_bytes"].(float64); ok {
		downloadedBytes = int64(db)
	}

	go callback(clientID, updateType, stage, progress, message, totalBytes, downloadedBytes)
}

// HandleUpdateResult 处理更新结果消息
func (h *ClientDaemonHub) HandleUpdateResult(clientID uint, data map[string]interface{}) {
	h.mu.RLock()
	callback := h.updateResultCallback
	h.mu.RUnlock()

	if callback == nil {
		log.Printf("[ClientDaemonHub] 更新结果回调未设置，忽略消息")
		return
	}

	updateType, _ := data["update_type"].(string)
	success, _ := data["success"].(bool)
	version, _ := data["version"].(string)
	message, _ := data["message"].(string)

	go callback(clientID, updateType, success, version, message)
}

// HandleVersionReport 处理版本上报消息
func (h *ClientDaemonHub) HandleVersionReport(clientID uint, data map[string]interface{}) {
	h.mu.RLock()
	callback := h.versionReportCallback
	h.mu.RUnlock()

	if callback == nil {
		log.Printf("[ClientDaemonHub] 版本上报回调未设置，忽略消息")
		return
	}

	frpcVersion, _ := data["frpc_version"].(string)
	daemonVersion, _ := data["daemon_version"].(string)
	os, _ := data["os"].(string)
	arch, _ := data["arch"].(string)

	go callback(clientID, frpcVersion, daemonVersion, os, arch)
}

// HandleLogData 处理日志数据消息
func (h *ClientDaemonHub) HandleLogData(clientID uint, data map[string]interface{}) {
	log.Printf("[ClientDaemonHub] 收到客户端 %d 的日志数据: %+v", clientID, data)

	h.mu.RLock()
	callback := h.logDataCallback
	h.mu.RUnlock()

	if callback == nil {
		log.Printf("[ClientDaemonHub] ⚠️ 日志数据回调未设置，忽略消息")
		return
	}

	logType, _ := data["log_type"].(string)
	line, _ := data["line"].(string)
	timestamp := int64(0)
	if ts, ok := data["timestamp"].(float64); ok {
		timestamp = int64(ts)
	}

	log.Printf("[ClientDaemonHub] 转发日志: clientID=%d, logType=%s, line=%s", clientID, logType, line)
	go callback(clientID, logType, line, timestamp)
}

// HandleFrpcControlResult 处理frpc控制结果消息
func (h *ClientDaemonHub) HandleFrpcControlResult(clientID uint, data map[string]interface{}) {
	action, _ := data["action"].(string)
	success, _ := data["success"].(bool)
	message, _ := data["message"].(string)

	log.Printf("[ClientDaemonHub] 处理frpc控制结果: clientID=%d, action=%s, success=%v", clientID, action, success)

	// 通知等待的请求
	h.NotifyFrpcControlResult(clientID, success, message)

	// 同时调用回调（如果设置了）
	h.mu.RLock()
	callback := h.frpcControlResultCallback
	h.mu.RUnlock()

	if callback != nil {
		go callback(clientID, action, success, message)
	}
}

// HandleConfigSyncResult 处理配置同步结果消息
func (h *ClientDaemonHub) HandleConfigSyncResult(clientID uint, data map[string]interface{}) {
	h.mu.RLock()
	callback := h.configSyncResultCallback
	h.mu.RUnlock()

	if callback == nil {
		log.Printf("[ClientDaemonHub] 配置同步结果回调未设置，忽略消息")
		return
	}

	success, _ := data["success"].(bool)
	errorMsg, _ := data["error"].(string)
	rolledBack, _ := data["rolled_back"].(bool)

	log.Printf("[ClientDaemonHub] 收到客户端 %d 配置同步结果: success=%v, error=%s, rolled_back=%v", clientID, success, errorMsg, rolledBack)
	go callback(clientID, success, errorMsg, rolledBack)
}
