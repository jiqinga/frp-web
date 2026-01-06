/*
 * @Author              : 寂情�?
 * @Date                : 2025-11-28 15:09:14
 * @LastEditors         : 寂情�?
 * @LastEditTime        : 2025-12-30 16:31:28
 * @FilePath            : frp-web-testbackendinternalhandlerdaemon_download_handler.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在�?
 */
package handler

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type DaemonDownloadHandler struct {
	daemonDir string
}

func NewDaemonDownloadHandler() *DaemonDownloadHandler {
	// 默认守护程序存储目录
	daemonDir := "./data/daemon"
	return &DaemonDownloadHandler{
		daemonDir: daemonDir,
	}
}

// Download godoc
// @Summary 下载Daemon守护程序
// @Description 下载指定操作系统和架构的frpc-daemon-ws守护程序二进制文�?
// @Tags Daemon下载
// @Produce application/octet-stream
// @Param os path string true "操作系统" Enums(linux, windows, darwin)
// @Param arch path string true "CPU架构" Enums(amd64, arm64, arm, 386)
// @Success 200 {file} binary "守护程序二进制文�?
// @Failure 400 {object} map[string]interface{} "不支持的平台"
// @Failure 404 {object} map[string]interface{} "守护程序文件不存�?
// @Router /download/daemon/{os}/{arch} [get]
func (h *DaemonDownloadHandler) Download(c *gin.Context) {
	osName := c.Param("os")
	arch := c.Param("arch")

	// 验证参数
	if !h.isValidPlatform(osName, arch) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的平台: " + osName + "/" + arch,
		})
		return
	}

	// 构建文件�?
	fileName := h.buildFileName(osName, arch)
	filePath := filepath.Join(h.daemonDir, fileName)

	// 检查文件是否存�?
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "守护程序文件不存在，请先构建: " + fileName,
		})
		return
	}

	// 设置响应�?
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/octet-stream")

	// 发送文�?
	c.File(filePath)
}

// isValidPlatform 验证平台是否支持
func (h *DaemonDownloadHandler) isValidPlatform(osName, arch string) bool {
	validPlatforms := map[string][]string{
		"linux":   {"amd64", "arm64", "arm"},
		"windows": {"amd64", "386"},
		"darwin":  {"amd64", "arm64"},
	}

	archs, ok := validPlatforms[osName]
	if !ok {
		return false
	}

	for _, a := range archs {
		if a == arch {
			return true
		}
	}
	return false
}

// buildFileName 构建文件�?
func (h *DaemonDownloadHandler) buildFileName(osName, arch string) string {
	baseName := "frpc-daemon-ws"
	if osName == "windows" {
		return baseName + "-" + osName + "-" + arch + ".exe"
	}
	return baseName + "-" + osName + "-" + arch
}
