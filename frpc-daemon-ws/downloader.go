/*
 * @Author              : Kilo Code
 * @Date                : 2025-12-01
 * @Description         : 客户端更新器 - 下载功能
 */
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// downloadFile 下载文件，支持进度回调
func (u *Updater) downloadFile(url string, destPath string, progressCallback func(downloaded, total int64)) (int64, error) {
	// 创建带超时的 HTTP 客户端，防止大文件下载无限阻塞
	client := &http.Client{
		Timeout: 30 * time.Minute, // 下载超时30分钟
	}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	totalSize := resp.ContentLength

	out, err := os.Create(destPath)
	if err != nil {
		return 0, fmt.Errorf("创建文件失败: %v", err)
	}
	defer out.Close()

	var downloaded int64
	buf := make([]byte, 32*1024)
	lastReport := time.Now()

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				return downloaded, fmt.Errorf("写入文件失败: %v", writeErr)
			}
			downloaded += int64(n)

			if time.Since(lastReport) > 100*time.Millisecond {
				if progressCallback != nil && totalSize > 0 {
					progressCallback(downloaded, totalSize)
				}
				lastReport = time.Now()
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return downloaded, fmt.Errorf("读取响应失败: %v", err)
		}
	}

	if progressCallback != nil && totalSize > 0 {
		progressCallback(downloaded, totalSize)
	}

	return downloaded, nil
}
