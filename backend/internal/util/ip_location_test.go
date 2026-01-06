/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-03 16:29:00
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-12-03 16:30:02
 * @FilePath            : frp-web-testbackendinternalutilip_location_test.go
 * @Description         : IP归属地查询测试
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package util

import (
	"testing"
)

func TestGetIPLocation(t *testing.T) {
	// 初始化搜索器
	err := InitIPSearcher("../../data")
	if err != nil {
		t.Fatalf("初始化 IP 搜索器失败: %v", err)
	}
	defer CloseIPSearcher()

	tests := []struct {
		name     string
		ip       string
		expected string
	}{
		{
			name:     "内网IP - 192.168.x.x",
			ip:       "192.168.1.1",
			expected: "内网IP",
		},
		{
			name:     "内网IP - 10.x.x.x",
			ip:       "10.0.0.1",
			expected: "内网IP",
		},
		{
			name:     "内网IP - 127.0.0.1",
			ip:       "127.0.0.1",
			expected: "内网IP",
		},
		{
			name:     "公网IP - 百度",
			ip:       "220.181.38.148",
			expected: "", // 不检查具体值，只检查不为空
		},
		{
			name:     "公网IP - 谷歌DNS",
			ip:       "8.8.8.8",
			expected: "", // 不检查具体值，只检查不为空
		},
		{
			name:     "公网IP - 阿里云",
			ip:       "47.95.164.112",
			expected: "", // 不检查具体值，只检查不为空
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetIPLocation(tt.ip)
			t.Logf("IP: %s -> 归属地: %s", tt.ip, result)

			if tt.expected != "" {
				if result != tt.expected {
					t.Errorf("GetIPLocation(%s) = %s, want %s", tt.ip, result, tt.expected)
				}
			} else {
				// 对于公网IP，只检查结果不为空且不是错误信息
				if result == "" || result == "查询失败" || result == "未知" {
					t.Errorf("GetIPLocation(%s) = %s, expected valid location", tt.ip, result)
				}
			}
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.1", true},
		{"192.168.0.1", true},
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"127.0.0.1", true},
		{"169.254.1.1", true},
		{"8.8.8.8", false},
		{"220.181.38.148", false},
		{"1.1.1.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := isPrivateIP(tt.ip)
			if result != tt.expected {
				t.Errorf("isPrivateIP(%s) = %v, want %v", tt.ip, result, tt.expected)
			}
		})
	}
}

func BenchmarkGetIPLocation(b *testing.B) {
	// 初始化搜索器
	err := InitIPSearcher("../../data")
	if err != nil {
		b.Fatalf("初始化 IP 搜索器失败: %v", err)
	}
	defer CloseIPSearcher()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetIPLocation("220.181.38.148")
	}
}
