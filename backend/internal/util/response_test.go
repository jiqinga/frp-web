/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-18 12:16:00
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-18 12:16:12
 * @FilePath            : frp-web-testbackendinternalutilresponse_test.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package util

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Success(c, "test data")

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"code":0`)
	assert.Contains(t, w.Body.String(), `"data":"test data"`)
}

func TestError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Error(c, 400, "error message")

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"code":400`)
	assert.Contains(t, w.Body.String(), `"message":"error message"`)
}

func TestSuccessResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	SuccessResponse(c, map[string]string{"key": "value"})

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"code":0`)
}
