/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-19 17:08:54
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-19 17:09:06
 * @FilePath            : frp-web-testbackendinternalfrpfrpc_client.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package frp

import "time"

type FrpcClient struct {
	*Client
}

func NewFrpcClient(host string, port int, username, password string) *FrpcClient {
	return &FrpcClient{
		Client: NewClient(host, port, username, password, 10*time.Second),
	}
}

func (c *FrpcClient) GetStatus() (*ClientStatus, error) {
	var status ClientStatus
	if err := c.doRequest("GET", "/api/status", &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func (c *FrpcClient) Reload() error {
	return c.doRequest("GET", "/api/reload", nil)
}
