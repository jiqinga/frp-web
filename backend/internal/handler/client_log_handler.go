package handler

import (
	"frp-web-panel/internal/service"
	"frp-web-panel/internal/util"
	"frp-web-panel/internal/websocket"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ClientLogHandler struct {
	clientService *service.ClientService
	logService    *service.LogService
}

func NewClientLogHandler() *ClientLogHandler {
	return &ClientLogHandler{
		clientService: service.NewClientService(),
		logService:    service.NewLogService(),
	}
}

// StartLogStream godoc
// @Summary 开始日志流
// @Description 开始接收指定客户端的实时日志流
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "客户端ID"
// @Param request body object{log_type=string,lines=int} true "日志配置，log_type: frpc/daemon，lines: 初始行数"
// @Success 200 {object} util.Response{data=object{message=string}}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/clients/{id}/logs/start [post]
func (h *ClientLogHandler) StartLogStream(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		LogType string `json:"log_type" binding:"required"` // frpc 或 daemon
		Lines   int    `json:"lines"`                       // 初始行数，默认100
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	if req.LogType != "frpc" && req.LogType != "daemon" {
		util.Error(c, 400, "无效的日志类型，必须是 frpc 或 daemon")
		return
	}

	if req.Lines <= 0 {
		req.Lines = 100
	}

	if !websocket.ClientDaemonHubInstance.IsClientOnline(uint(id)) {
		util.Error(c, 400, "客户端未连接")
		return
	}

	if err := websocket.ClientDaemonHubInstance.SendLogStreamCommand(uint(id), req.LogType, "start", req.Lines); err != nil {
		util.Error(c, 500, "发送日志流命令失败: "+err.Error())
		return
	}

	log.Printf("[日志流] 已向客户端 %d 发送开始日志流命令: type=%s, lines=%d", id, req.LogType, req.Lines)
	util.Success(c, gin.H{"message": "日志流已启动"})
}

// StopLogStream godoc
// @Summary 停止日志流
// @Description 停止接收指定客户端的实时日志流
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "客户端ID"
// @Param request body object{log_type=string} true "日志类型，log_type: frpc/daemon"
// @Success 200 {object} util.Response{data=object{message=string}}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/clients/{id}/logs/stop [post]
func (h *ClientLogHandler) StopLogStream(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		LogType string `json:"log_type" binding:"required"` // frpc 或 daemon
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	if req.LogType != "frpc" && req.LogType != "daemon" {
		util.Error(c, 400, "无效的日志类型，必须是 frpc 或 daemon")
		return
	}

	if !websocket.ClientDaemonHubInstance.IsClientOnline(uint(id)) {
		util.Error(c, 400, "客户端未连接")
		return
	}

	if err := websocket.ClientDaemonHubInstance.SendLogStreamCommand(uint(id), req.LogType, "stop", 0); err != nil {
		util.Error(c, 500, "发送停止日志流命令失败: "+err.Error())
		return
	}

	log.Printf("[日志流] 已向客户端 %d 发送停止日志流命令: type=%s", id, req.LogType)
	util.Success(c, gin.H{"message": "日志流已停止"})
}

// ControlFrpc godoc
// @Summary 控制 frpc
// @Description 控制客户端的 frpc 进程（启动/停止/重启），同步等待执行结果
// @Tags 客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "客户端ID"
// @Param request body object{action=string} true "操作类型，action: start/stop/restart"
// @Success 200 {object} util.Response{data=object{success=bool,message=string}}
// @Failure 400 {object} util.Response
// @Failure 500 {object} util.Response
// @Router /api/clients/{id}/frpc/control [post]
func (h *ClientLogHandler) ControlFrpc(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		Action string `json:"action" binding:"required"` // start, stop, restart
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	if req.Action != "start" && req.Action != "stop" && req.Action != "restart" {
		util.Error(c, 400, "无效的操作类型，必须是 start, stop 或 restart")
		return
	}

	if !websocket.ClientDaemonHubInstance.IsClientOnline(uint(id)) {
		util.Error(c, 400, "客户端未连接")
		return
	}

	// 同步等待结果，超时30秒
	result, err := websocket.ClientDaemonHubInstance.SendFrpcControlCommandAndWait(uint(id), req.Action, 30*time.Second)
	if err != nil {
		util.Error(c, 500, "frpc控制失败: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	client, _ := h.clientService.GetClient(uint(id))
	clientName := ""
	if client != nil {
		clientName = client.Name
	}
	h.logService.CreateLogAsync(userID.(uint), "frpc_control", "client", uint(id),
		"frpc控制: "+clientName+" (操作: "+req.Action+", 结果: "+result.Message+")", c.ClientIP())

	log.Printf("[frpc控制] 客户端 %d 控制完成: action=%s, success=%v", id, req.Action, result.Success)

	if result.Success {
		// 更新客户端的 frpc 在线状态
		var newStatus string
		if req.Action == "stop" {
			newStatus = "offline"
		} else {
			newStatus = "online"
		}
		if err := h.clientService.UpdateOnlineStatusDirectly(uint(id), newStatus); err != nil {
			log.Printf("[frpc控制] ⚠️ 更新客户端状态失败: %v", err)
		}
		util.Success(c, gin.H{"success": true, "message": result.Message})
	} else {
		util.Error(c, 500, result.Message)
	}
}
