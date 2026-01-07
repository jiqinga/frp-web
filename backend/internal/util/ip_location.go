/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-03 16:06:00
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-07 10:57:47
 * @FilePath            : frp-web-testbackendinternalutilip_location.go
 * @Description         : IP归属地查询工具 - 使用 ip2region 离线数据库
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package util

import (
	"frp-web-panel/internal/logger"
	"net"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

var (
	ipSearcher     *xdb.Searcher
	ipSearcherOnce sync.Once
	ipSearcherErr  error
)

// InitIPSearcher 初始化 IP 搜索器
// 使用完全基于内存的查询方式，性能最佳
func InitIPSearcher(dataDir string) error {
	ipSearcherOnce.Do(func() {
		// 构建 xdb 文件路径
		dbPath := filepath.Join(dataDir, "ip2region_v4.xdb")

		// 从文件加载整个 xdb 数据到内存
		cBuff, err := xdb.LoadContentFromFile(dbPath)
		if err != nil {
			ipSearcherErr = err
			logger.Errorf("[IP2Region] 加载数据文件失败: %v", err)
			return
		}

		// 获取 IPv4 版本
		version, err := xdb.VersionFromName("v4")
		if err != nil {
			ipSearcherErr = err
			logger.Errorf("[IP2Region] 获取版本信息失败: %v", err)
			return
		}

		// 创建基于内存的搜索器
		ipSearcher, ipSearcherErr = xdb.NewWithBuffer(version, cBuff)
		if ipSearcherErr != nil {
			logger.Errorf("[IP2Region] 创建搜索器失败: %v", ipSearcherErr)
			return
		}

		logger.Infof("[IP2Region] 初始化成功，数据文件: %s", dbPath)
	})

	return ipSearcherErr
}

// CloseIPSearcher 关闭 IP 搜索器，释放资源
func CloseIPSearcher() {
	if ipSearcher != nil {
		ipSearcher.Close()
		logger.Info("[IP2Region] 搜索器已关闭")
	}
}

// GetIPLocation 查询IP归属地
// 使用 ip2region 离线数据库进行查询
func GetIPLocation(ip string) string {
	// 检查是否为内网IP或本地IP
	if isPrivateIP(ip) {
		return "内网IP"
	}

	// 检查搜索器是否已初始化
	if ipSearcher == nil {
		// 尝试延迟初始化
		if err := InitIPSearcher("./data"); err != nil {
			return "查询失败"
		}
	}

	// 执行查询
	region, err := ipSearcher.SearchByStr(ip)
	if err != nil {
		logger.Errorf("[IP2Region] 查询IP %s 失败: %v", ip, err)
		return "查询失败"
	}

	// 解析并格式化结果
	// ip2region 返回格式: 国家|区域|省份|城市|ISP
	return formatIPRegion(region)
}

// formatIPRegion 格式化 ip2region 返回的结果
// ip2region v4 返回格式: 国家|区域|省份|城市|ISP (5个字段)
// 或者: 国家|省份|城市|ISP (4个字段，某些版本)
// 输出格式: 国家 省份 城市 ISP (去除空值和重复值)
func formatIPRegion(region string) string {
	parts := strings.Split(region, "|")

	var country, province, city, isp string

	switch len(parts) {
	case 5:
		// 标准格式: 国家|区域|省份|城市|ISP
		country = parts[0]
		// area := parts[1]    // 区域，通常为 0，不使用
		province = parts[2]
		city = parts[3]
		isp = parts[4]
	case 4:
		// 简化格式: 国家|省份|城市|ISP
		country = parts[0]
		province = parts[1]
		city = parts[2]
		isp = parts[3]
	default:
		// 无法解析，直接返回原始字符串
		return region
	}

	// 构建结果，去除空值（0 表示空）
	var result []string

	if country != "0" && country != "" {
		result = append(result, country)
	}

	if province != "0" && province != "" && province != country {
		result = append(result, province)
	}

	if city != "0" && city != "" && city != province {
		result = append(result, city)
	}

	if isp != "0" && isp != "" {
		result = append(result, isp)
	}

	if len(result) == 0 {
		return "未知"
	}

	return strings.Join(result, " ")
}

// isPrivateIP 检查是否为内网IP
func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// 检查是否为回环地址
	if ip.IsLoopback() {
		return true
	}

	// 检查是否为私有地址
	if ip.IsPrivate() {
		return true
	}

	// 检查是否为链路本地地址
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// 检查常见的内网IP段
	privateBlocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	}

	for _, block := range privateBlocks {
		_, cidr, err := net.ParseCIDR(block)
		if err != nil {
			continue
		}
		if cidr.Contains(ip) {
			return true
		}
	}

	return false
}
