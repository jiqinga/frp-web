/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-28
 * @Description         : frpc Admin API 客户端
 */
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// FrpcAdminClient frpc Admin API 客户端
type FrpcAdminClient struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

// ProxyStatus 代理状态
type ProxyStatus struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	LocalAddr  string `json:"localAddr"`
	RemoteAddr string `json:"remoteAddr"`
	Error      string `json:"error"`
}

// AllProxyStatus 所有代理状态
type AllProxyStatus struct {
	TCP   []ProxyStatus `json:"tcp"`
	UDP   []ProxyStatus `json:"udp"`
	HTTP  []ProxyStatus `json:"http"`
	HTTPS []ProxyStatus `json:"https"`
	STCP  []ProxyStatus `json:"stcp"`
	SUDP  []ProxyStatus `json:"sudp"`
	XTCP  []ProxyStatus `json:"xtcp"`
}

// NewFrpcAdminClient 创建新的 frpc Admin API 客户端
func NewFrpcAdminClient(addr string, port int, username, password string) *FrpcAdminClient {
	baseURL := fmt.Sprintf("http://%s:%d", addr, port)
	return &FrpcAdminClient{
		baseURL:  baseURL,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// doRequest 执行 HTTP 请求
func (c *FrpcAdminClient) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置 Basic Auth
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

// HealthCheck 健康检查（使用 /api/status 接口）
func (c *FrpcAdminClient) HealthCheck() error {
	log.Printf("[FrpcAdmin] 执行健康检查: %s", c.baseURL)
	log.Printf("[FrpcAdmin] 使用认证: user=%s, password=%s", c.username, maskPassword(c.password))

	resp, err := c.doRequest("GET", "/api/status", nil)
	if err != nil {
		return fmt.Errorf("健康检查请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("健康检查失败: 认证失败(401 Unauthorized)，请检查 frpc_admin_user 和 frpc_admin_password 配置是否与 frpc.toml 中的 webServer.user 和 webServer.password 一致")
		}
		return fmt.Errorf("健康检查失败, 状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	log.Printf("[FrpcAdmin] ✅ 健康检查通过")
	return nil
}

// CheckFrpcAlive 检查 frpc 进程是否存活（使用 /healthz 接口）
// 这个接口用于判断 frpc 进程本身是否在运行
func (c *FrpcAdminClient) CheckFrpcAlive() bool {
	resp, err := c.doRequest("GET", "/healthz", nil)
	if err != nil {
		log.Printf("[FrpcAdmin] frpc 健康检查失败: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[FrpcAdmin] frpc 健康检查失败, 状态码: %d", resp.StatusCode)
		return false
	}

	return true
}

// maskPassword 隐藏密码，只显示前两个字符
func maskPassword(password string) string {
	if len(password) <= 2 {
		return "***"
	}
	return password[:2] + "***"
}

// Reload 重载配置
func (c *FrpcAdminClient) Reload() error {
	log.Printf("[FrpcAdmin] 执行配置重载: %s/api/reload", c.baseURL)

	resp, err := c.doRequest("GET", "/api/reload", nil)
	if err != nil {
		return fmt.Errorf("重载请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("重载失败, 状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	log.Printf("[FrpcAdmin] ✅ 配置重载成功")
	return nil
}

// GetStatus 获取所有代理状态
func (c *FrpcAdminClient) GetStatus() (*AllProxyStatus, error) {
	log.Printf("[FrpcAdmin] 获取代理状态: %s/api/status", c.baseURL)

	resp, err := c.doRequest("GET", "/api/status", nil)
	if err != nil {
		return nil, fmt.Errorf("获取状态请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取状态失败, 状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var status AllProxyStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("解析状态响应失败: %v", err)
	}

	log.Printf("[FrpcAdmin] ✅ 获取代理状态成功")
	return &status, nil
}

// StopProxy 停止指定代理
func (c *FrpcAdminClient) StopProxy(names []string) error {
	log.Printf("[FrpcAdmin] 停止代理: %v", names)

	payload := map[string][]string{"names": names}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %v", err)
	}

	resp, err := c.doRequest("POST", "/api/stop", strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("停止代理请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("停止代理失败, 状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	log.Printf("[FrpcAdmin] ✅ 代理已停止: %v", names)
	return nil
}

// IsAvailable 检查 Admin API 是否可用
func (c *FrpcAdminClient) IsAvailable() bool {
	err := c.HealthCheck()
	return err == nil
}

// GetRunningProxyCount 获取运行中的代理数量
func (c *FrpcAdminClient) GetRunningProxyCount() (int, error) {
	status, err := c.GetStatus()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, p := range status.TCP {
		if p.Status == "running" {
			count++
		}
	}
	for _, p := range status.UDP {
		if p.Status == "running" {
			count++
		}
	}
	for _, p := range status.HTTP {
		if p.Status == "running" {
			count++
		}
	}
	for _, p := range status.HTTPS {
		if p.Status == "running" {
			count++
		}
	}
	for _, p := range status.STCP {
		if p.Status == "running" {
			count++
		}
	}
	for _, p := range status.SUDP {
		if p.Status == "running" {
			count++
		}
	}
	for _, p := range status.XTCP {
		if p.Status == "running" {
			count++
		}
	}

	return count, nil
}
