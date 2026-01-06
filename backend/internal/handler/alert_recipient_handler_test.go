package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"frp-web-panel/internal/model"
	"frp-web-panel/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupRecipientTestDB 创建内存数据库
func setupRecipientTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)

	err = db.AutoMigrate(&model.AlertRecipient{}, &model.AlertRecipientGroup{}, &model.AlertGroupRecipient{}, &model.OperationLog{})
	require.NoError(t, err)

	database.DB = db
	return db
}

// setupRecipientTestRouter 创建测试路由
func setupRecipientTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAlertRecipientHandler()

	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	})

	alerts := r.Group("/api/alerts")
	{
		alerts.GET("/recipients", handler.GetAllRecipients)
		alerts.POST("/recipients", handler.CreateRecipient)
		alerts.PUT("/recipients/:id", handler.UpdateRecipient)
		alerts.DELETE("/recipients/:id", handler.DeleteRecipient)
		alerts.GET("/groups", handler.GetAllGroups)
		alerts.POST("/groups", handler.CreateGroup)
		alerts.PUT("/groups/:id", handler.UpdateGroup)
		alerts.DELETE("/groups/:id", handler.DeleteGroup)
		alerts.PUT("/groups/:id/recipients", handler.SetGroupRecipients)
	}

	return r
}

// TestCreateRecipient 测试创建接收人
func TestCreateRecipient(t *testing.T) {
	setupRecipientTestDB(t)
	router := setupRecipientTestRouter()

	tests := []struct {
		name     string
		body     map[string]interface{}
		wantCode int
	}{
		{
			name:     "创建接收人成功",
			body:     map[string]interface{}{"name": "测试用户", "email": "test@example.com", "enabled": true},
			wantCode: 0,
		},
		{
			name:     "创建第二个接收人",
			body:     map[string]interface{}{"name": "管理员", "email": "admin@example.com", "enabled": true},
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/api/alerts/recipients", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			assert.Equal(t, float64(tt.wantCode), resp["code"])
		})
	}
}

// TestCreateRecipient_InvalidParams 测试创建接收人参数错误
func TestCreateRecipient_InvalidParams(t *testing.T) {
	setupRecipientTestDB(t)
	router := setupRecipientTestRouter()

	req, _ := http.NewRequest("POST", "/api/alerts/recipients", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(4001), resp["code"])
}

// TestGetAllRecipients 测试获取接收人列表
func TestGetAllRecipients(t *testing.T) {
	db := setupRecipientTestDB(t)
	router := setupRecipientTestRouter()

	// 清理旧数据
	db.Exec("DELETE FROM alert_recipients")
	db.Create(&model.AlertRecipient{Name: "用户1", Email: "user1@test.com", Enabled: true})
	db.Create(&model.AlertRecipient{Name: "用户2", Email: "user2@test.com", Enabled: true})

	t.Run("获取列表成功", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/alerts/recipients", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].([]interface{})
		assert.Equal(t, 2, len(data))
	})

	t.Run("空列表", func(t *testing.T) {
		db.Exec("DELETE FROM alert_recipients")
		req, _ := http.NewRequest("GET", "/api/alerts/recipients", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})
}

// TestUpdateRecipient 测试更新接收人
func TestUpdateRecipient(t *testing.T) {
	db := setupRecipientTestDB(t)
	router := setupRecipientTestRouter()

	recipient := model.AlertRecipient{Name: "原名称", Email: "old@test.com", Enabled: true}
	db.Create(&recipient)

	t.Run("更新成功", func(t *testing.T) {
		body := map[string]interface{}{"name": "新名称", "email": "new@test.com", "enabled": true}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/alerts/recipients/%d", recipient.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})

	t.Run("禁用接收人", func(t *testing.T) {
		body := map[string]interface{}{"name": "新名称", "email": "new@test.com", "enabled": false}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/alerts/recipients/%d", recipient.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})
}

// TestDeleteRecipient 测试删除接收人
func TestDeleteRecipient(t *testing.T) {
	db := setupRecipientTestDB(t)
	router := setupRecipientTestRouter()

	recipient := model.AlertRecipient{Name: "待删除", Email: "del@test.com", Enabled: true}
	db.Create(&recipient)

	t.Run("删除成功", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/alerts/recipients/%d", recipient.ID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])

		var count int64
		db.Model(&model.AlertRecipient{}).Where("id = ?", recipient.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("删除不存在的接收人", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/alerts/recipients/99999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})
}

// TestCreateGroup 测试创建分组
func TestCreateGroup(t *testing.T) {
	setupRecipientTestDB(t)
	router := setupRecipientTestRouter()

	tests := []struct {
		name     string
		body     map[string]interface{}
		wantCode int
	}{
		{
			name:     "创建分组成功",
			body:     map[string]interface{}{"name": "运维组", "description": "运维人员", "enabled": true},
			wantCode: 0,
		},
		{
			name:     "创建第二个分组",
			body:     map[string]interface{}{"name": "开发组", "description": "开发人员", "enabled": true},
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/api/alerts/groups", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			assert.Equal(t, float64(tt.wantCode), resp["code"])
		})
	}
}

// TestCreateGroup_InvalidParams 测试创建分组参数错误
func TestCreateGroup_InvalidParams(t *testing.T) {
	setupRecipientTestDB(t)
	router := setupRecipientTestRouter()

	req, _ := http.NewRequest("POST", "/api/alerts/groups", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(4001), resp["code"])
}

// TestGetAllGroups 测试获取分组列表
func TestGetAllGroups(t *testing.T) {
	db := setupRecipientTestDB(t)
	router := setupRecipientTestRouter()

	db.Create(&model.AlertRecipientGroup{Name: "分组1", Description: "描述1", Enabled: true})
	db.Create(&model.AlertRecipientGroup{Name: "分组2", Description: "描述2", Enabled: true})

	t.Run("获取列表成功", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/alerts/groups", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].([]interface{})
		assert.Equal(t, 2, len(data))
	})
}

// TestUpdateGroup 测试更新分组
func TestUpdateGroup(t *testing.T) {
	db := setupRecipientTestDB(t)
	router := setupRecipientTestRouter()

	group := model.AlertRecipientGroup{Name: "原分组", Description: "原描述", Enabled: true}
	db.Create(&group)

	t.Run("更新成功", func(t *testing.T) {
		body := map[string]interface{}{"name": "新分组", "description": "新描述", "enabled": true}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/alerts/groups/%d", group.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})
}

// TestDeleteGroup 测试删除分组
func TestDeleteGroup(t *testing.T) {
	db := setupRecipientTestDB(t)
	router := setupRecipientTestRouter()

	group := model.AlertRecipientGroup{Name: "待删除分组", Enabled: true}
	db.Create(&group)

	t.Run("删除成功", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/alerts/groups/%d", group.ID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})
}

// TestSetGroupRecipients 测试设置分组成员
func TestSetGroupRecipients(t *testing.T) {
	db := setupRecipientTestDB(t)
	router := setupRecipientTestRouter()

	group := model.AlertRecipientGroup{Name: "测试分组", Enabled: true}
	db.Create(&group)

	r1 := model.AlertRecipient{Name: "成员1", Email: "m1@test.com", Enabled: true}
	r2 := model.AlertRecipient{Name: "成员2", Email: "m2@test.com", Enabled: true}
	db.Create(&r1)
	db.Create(&r2)

	t.Run("设置分组成员", func(t *testing.T) {
		body := map[string]interface{}{"recipient_ids": []uint{r1.ID, r2.ID}}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/alerts/groups/%d/recipients", group.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])

		var count int64
		db.Model(&model.AlertGroupRecipient{}).Where("group_id = ?", group.ID).Count(&count)
		assert.Equal(t, int64(2), count)
	})

	t.Run("清空分组成员", func(t *testing.T) {
		body := map[string]interface{}{"recipient_ids": []uint{}}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/alerts/groups/%d/recipients", group.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])

		var count int64
		db.Model(&model.AlertGroupRecipient{}).Where("group_id = ?", group.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})
}

// TestSetGroupRecipients_InvalidParams 测试设置分组成员参数错误
func TestSetGroupRecipients_InvalidParams(t *testing.T) {
	db := setupRecipientTestDB(t)
	router := setupRecipientTestRouter()

	group := model.AlertRecipientGroup{Name: "测试分组", Enabled: true}
	db.Create(&group)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/alerts/groups/%d/recipients", group.ID), bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(4001), resp["code"])
}
