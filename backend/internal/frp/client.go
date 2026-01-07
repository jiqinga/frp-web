/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-19 17:07:58
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-07 11:00:45
 * @FilePath            : frp-web-testbackendinternalfrpclient.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package frp

import (
	"encoding/json"
	"fmt"
	"frp-web-panel/internal/logger"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

func NewClient(host string, port int, username, password string, timeout time.Duration) *Client {
	return &Client{
		baseURL:  fmt.Sprintf("http://%s:%d", host, port),
		username: username,
		password: password,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) doRequest(method, path string, result interface{}) error {
	fullURL := c.baseURL + path
	logger.Debugf("FRP API请求 - Method: %s, URL: %s", method, fullURL)

	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		logger.Debugf("创建HTTP请求失败: %v", err)
		return err
	}

	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
		logger.Debugf("使用Basic Auth - User: %s", c.username)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		logger.Debugf("HTTP请求执行失败 - 原始错误: %v", err)
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	logger.Debugf("HTTP响应 - StatusCode: %d, Status: %s", resp.StatusCode, resp.Status)

	if resp.StatusCode == http.StatusUnauthorized {
		logger.Debug("认证失败 - 401 Unauthorized")
		return ErrAuthFailed
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Debugf("读取响应体失败: %v", err)
		return err
	}

	logger.Debugf("响应体长度: %d bytes", len(body))
	if len(body) > 0 && len(body) < 1000 {
		logger.Debugf("响应体内容: %s", string(body))
	} else if len(body) >= 1000 {
		logger.Debugf("响应体内容(前500字符): %s...", string(body[:500]))
	}

	if resp.StatusCode != http.StatusOK {
		logger.Debugf("HTTP状态码异常: %d, 响应: %s", resp.StatusCode, string(body))
		return fmt.Errorf("HTTP状态码 %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			logger.Debugf("JSON解析失败: %v, 响应内容: %s", err, string(body))
			return fmt.Errorf("%w: %v", ErrInvalidResponse, err)
		}
		logger.Debug("JSON解析成功")
	}

	return nil
}

// doRequestRaw 执行HTTP请求并返回原始响应体字符串
func (c *Client) doRequestRaw(method, path string) (string, error) {
	fullURL := c.baseURL + path
	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		return "", err
	}

	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", ErrAuthFailed
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("%w: /metrics endpoint not available", ErrMetricsNotSupported)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP状态码 %d: %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}
