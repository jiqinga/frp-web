package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"frp-web-panel/internal/model"
	"frp-web-panel/internal/repository"
	"frp-web-panel/internal/service"
	"frp-web-panel/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupAlertTestDB 创建内存数据库并迁移表结构
func setupAlertTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err, "创建内存数据库失败")

	// 限制连接数，确保使用同一个内存数据库
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)

	err = db.AutoMigrate(&model.AlertRule{}, &model.AlertLog{}, &model.Proxy{}, &model.OperationLog{})
	require.NoError(t, err, "数据库迁移失败")

	// 设置全局数据库供 LogService 使用
	database.DB = db
	return db
}

// setupAlertTestRouter 创建测试路由
func setupAlertTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	alertRepo := repository.NewAlertRepo(db)
	trafficRepo := &repository.TrafficRepository{}
	proxyRepo := &repository.ProxyRepository{}
	alertService := service.NewAlertService(alertRepo, trafficRepo, proxyRepo)
	alertHandler := NewAlertHandler(alertService)

	// 模拟认证中间件
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	})

	alerts := r.Group("/api/alerts")
	{
		alerts.POST("/rules", alertHandler.CreateRule)
		alerts.GET("/rules", alertHandler.GetAllRules)
		alerts.GET("/rules/proxy/:id", alertHandler.GetRulesByProxyID)
		alerts.PUT("/rules", alertHandler.UpdateRule)
		alerts.DELETE("/rules/:id", alertHandler.DeleteRule)
		alerts.GET("/logs", alertHandler.GetAlertLogs)
	}

	return r
}

// TestCreateAlertRule_Success 测试创建告警规则成功场景
func TestCreateAlertRule_Success(t *testing.T) {
	db := setupAlertTestDB(t)
	router := setupAlertTestRouter(db)

	tests := []struct {
		name     string
		rule     map[string]interface{}
		wantCode int
	}{
		{
			name: "创建每日流量告警规则",
			rule: map[string]interface{}{
				"target_type": "proxy", "target_id": 1, "proxy_id": 1,
				"rule_type": "daily", "threshold_value": 1024, "threshold_unit": "MB", "enabled": true,
			},
			wantCode: 0,
		},
		{
			name: "创建每月流量告警规则",
			rule: map[string]interface{}{
				"target_type": "proxy", "target_id": 2, "proxy_id": 2,
				"rule_type": "monthly", "threshold_value": 10240, "threshold_unit": "GB", "enabled": true,
			},
			wantCode: 0,
		},
		{
			name: "创建实时速率告警规则",
			rule: map[string]interface{}{
				"target_type": "proxy", "target_id": 3, "proxy_id": 3,
				"rule_type": "rate", "threshold_value": 100, "threshold_unit": "MB", "enabled": true,
			},
			wantCode: 0,
		},
		{
			name: "创建frpc离线告警规则",
			rule: map[string]interface{}{
				"target_type": "frpc", "target_id": 1, "rule_type": "offline",
				"offline_delay_seconds": 60, "notify_on_recovery": true, "enabled": true,
			},
			wantCode: 0,
		},
		{
			name: "创建frps离线告警规则",
			rule: map[string]interface{}{
				"target_type": "frps", "target_id": 1, "rule_type": "offline",
				"offline_delay_seconds": 120, "notify_on_recovery": true, "enabled": true,
			},
			wantCode: 0,
		},
		{
			name: "创建系统告警规则-证书过期",
			rule: map[string]interface{}{
				"target_type": "system", "target_id": 0, "rule_type": "cert_expiring",
				"cooldown_minutes": 60, "enabled": true,
			},
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.rule)
			req, _ := http.NewRequest("POST", "/api/alerts/rules", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, float64(tt.wantCode), resp["code"])
		})
	}
}

// TestCreateAlertRule_InvalidParams 测试创建告警规则参数验证失败
func TestCreateAlertRule_InvalidParams(t *testing.T) {
	db := setupAlertTestDB(t)
	router := setupAlertTestRouter(db)

	tests := []struct {
		name     string
		body     string
		wantCode int
	}{
		{"空请求体", "", 4001},
		{"无效JSON", "{invalid}", 4001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/api/alerts/rules", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			assert.Equal(t, float64(tt.wantCode), resp["code"])
		})
	}
}

// TestGetAllRules 测试获取所有告警规则
func TestGetAllRules(t *testing.T) {
	db := setupAlertTestDB(t)
	router := setupAlertTestRouter(db)

	// 创建测试规则
	db.Create(&model.AlertRule{TargetType: "proxy", TargetID: 1, ProxyID: 1, RuleType: "daily", Enabled: true})
	db.Create(&model.AlertRule{TargetType: "frpc", TargetID: 1, RuleType: "offline", Enabled: true})

	t.Run("获取规则列表成功", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/alerts/rules", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)

		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].([]interface{})
		assert.Equal(t, 2, len(data))
	})

	t.Run("空列表场景", func(t *testing.T) {
		db.Exec("DELETE FROM alert_rules")
		req, _ := http.NewRequest("GET", "/api/alerts/rules", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].([]interface{})
		assert.Equal(t, 0, len(data))
	})
}

// TestUpdateRule 测试更新告警规则
func TestUpdateRule(t *testing.T) {
	db := setupAlertTestDB(t)
	router := setupAlertTestRouter(db)

	rule := model.AlertRule{TargetType: "proxy", TargetID: 1, ProxyID: 1, RuleType: "daily", ThresholdValue: 1024, Enabled: true}
	db.Create(&rule)

	tests := []struct {
		name     string
		update   map[string]interface{}
		wantCode int
	}{
		{
			name: "更新阈值",
			update: map[string]interface{}{
				"id": rule.ID, "target_type": "proxy", "target_id": 1, "proxy_id": 1,
				"rule_type": "daily", "threshold_value": 2048, "threshold_unit": "MB", "enabled": true,
			},
			wantCode: 0,
		},
		{
			name: "禁用规则",
			update: map[string]interface{}{
				"id": rule.ID, "target_type": "proxy", "target_id": 1, "proxy_id": 1,
				"rule_type": "daily", "threshold_value": 2048, "enabled": false,
			},
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.update)
			req, _ := http.NewRequest("PUT", "/api/alerts/rules", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			assert.Equal(t, float64(tt.wantCode), resp["code"])
		})
	}
}

// TestUpdateRule_InvalidParams 测试更新规则参数验证失败
func TestUpdateRule_InvalidParams(t *testing.T) {
	db := setupAlertTestDB(t)
	router := setupAlertTestRouter(db)

	req, _ := http.NewRequest("PUT", "/api/alerts/rules", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(4001), resp["code"])
}

// TestDeleteRule 测试删除告警规则
func TestDeleteRule(t *testing.T) {
	db := setupAlertTestDB(t)
	router := setupAlertTestRouter(db)

	rule := model.AlertRule{TargetType: "proxy", TargetID: 1, ProxyID: 1, RuleType: "daily", Enabled: true}
	db.Create(&rule)

	t.Run("正常删除", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/alerts/rules/%d", rule.ID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])

		var count int64
		db.Model(&model.AlertRule{}).Where("id = ?", rule.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("删除不存在的规则", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/alerts/rules/99999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})
}

// TestGetAlertLogs 测试获取告警日志
func TestGetAlertLogs(t *testing.T) {
	db := setupAlertTestDB(t)
	router := setupAlertTestRouter(db)

	db.Create(&model.AlertLog{RuleID: 1, TargetType: "proxy", TargetID: 1, AlertType: "daily", Message: "测试告警1"})
	db.Create(&model.AlertLog{RuleID: 1, TargetType: "proxy", TargetID: 1, AlertType: "daily", Message: "测试告警2"})
	db.Create(&model.AlertLog{RuleID: 2, TargetType: "frpc", TargetID: 1, AlertType: "offline", Message: "离线告警"})

	t.Run("获取日志列表", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/alerts/logs", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].([]interface{})
		assert.Equal(t, 3, len(data))
	})

	t.Run("使用limit参数", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/alerts/logs?limit=2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].([]interface{})
		assert.Equal(t, 2, len(data))
	})
}

// TestGetRulesByProxyID 测试根据代理ID获取规则
func TestGetRulesByProxyID(t *testing.T) {
	db := setupAlertTestDB(t)
	router := setupAlertTestRouter(db)

	db.Create(&model.AlertRule{TargetType: "proxy", TargetID: 1, ProxyID: 1, RuleType: "daily", Enabled: true})
	db.Create(&model.AlertRule{TargetType: "proxy", TargetID: 1, ProxyID: 1, RuleType: "monthly", Enabled: true})
	db.Create(&model.AlertRule{TargetType: "proxy", TargetID: 2, ProxyID: 2, RuleType: "daily", Enabled: true})

	t.Run("获取指定代理的规则", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/alerts/rules/proxy/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].([]interface{})
		assert.Equal(t, 2, len(data))
	})

	t.Run("获取不存在代理的规则", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/alerts/rules/proxy/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].([]interface{})
		assert.Equal(t, 0, len(data))
	})
}
