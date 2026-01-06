/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-19 17:07:34
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-09 15:04:17
 * @FilePath            : frp-web-testbackendinternalfrperrors.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package frp

import "errors"

var (
	ErrConnectionFailed    = errors.New("连接 FRP 服务器失败")
	ErrAuthFailed          = errors.New("认证失败")
	ErrTimeout             = errors.New("请求超时")
	ErrInvalidResponse     = errors.New("无效的响应数据")
	ErrServerNotFound      = errors.New("FRP 服务器不存在")
	ErrMetricsNotSupported = errors.New("服务器未开启 metrics 接口")
)
