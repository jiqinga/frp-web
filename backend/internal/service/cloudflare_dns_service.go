/*
 * Cloudflare DNS 服务
 * 使用 cloudflare-go 官方 SDK
 * License: BSD-3-Clause
 */
package service

import (
	"context"
	"fmt"
	"frp-web-panel/internal/logger"
	"strings"

	"github.com/cloudflare/cloudflare-go"
)

// CloudflareDNSService Cloudflare DNS 服务
type CloudflareDNSService struct {
	api *cloudflare.API
}

// NewCloudflareDNSService 创建 Cloudflare DNS 服务
// apiToken: Cloudflare API Token (推荐使用 API Token 而非 API Key)
func NewCloudflareDNSService(apiToken string) (*CloudflareDNSService, error) {
	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		return nil, fmt.Errorf("创建 Cloudflare 客户端失败: %v", err)
	}
	return &CloudflareDNSService{api: api}, nil
}

// getZoneID 根据域名获取 Zone ID
// 智能匹配：从可用 Zones 中找到最长匹配的 Zone
// 例如：对于 opa.frps.de5.net，如果有 frps.de5.net 和 de5.net，选择 frps.de5.net
func (s *CloudflareDNSService) getZoneID(ctx context.Context, domain string) (string, error) {
	logger.Debugf("Cloudflare DNS getZoneID 调用: 完整域名=%s", domain)

	// 去除 FQDN 末尾的点（DNS-01 挑战返回的 FQDN 格式为 "_acme-challenge.example.com."）
	domain = strings.TrimSuffix(domain, ".")

	// 获取所有可用的 zones
	allZones, err := s.api.ListZones(ctx)
	if err != nil {
		return "", fmt.Errorf("获取 Zone 列表失败: %v", err)
	}

	logger.Debugf("Cloudflare DNS 当前 API Token 可访问的所有 Zones (%d 个):", len(allZones))
	for i, z := range allZones {
		logger.Debugf("Cloudflare DNS   [%d] Zone: %s (ID: %s)", i+1, z.Name, z.ID)
	}

	// 智能匹配：找到最长匹配的 Zone
	// 域名必须以 ".zoneName" 结尾，或者完全等于 zoneName
	var bestMatch cloudflare.Zone
	bestMatchLen := 0

	domainLower := strings.ToLower(domain)
	for _, zone := range allZones {
		zoneLower := strings.ToLower(zone.Name)
		// 检查域名是否以 .zoneName 结尾，或者完全等于 zoneName
		if domainLower == zoneLower || strings.HasSuffix(domainLower, "."+zoneLower) {
			if len(zone.Name) > bestMatchLen {
				bestMatch = zone
				bestMatchLen = len(zone.Name)
			}
		}
	}

	if bestMatchLen == 0 {
		return "", fmt.Errorf("未找到域名 %s 对应的 Zone，请确认该域名已在 Cloudflare 中托管且 API Token 有访问权限", domain)
	}

	logger.Debugf("Cloudflare DNS 智能匹配成功: 域名=%s -> Zone=%s (ID: %s)", domain, bestMatch.Name, bestMatch.ID)
	return bestMatch.ID, nil
}

// AddRecord 添加 DNS A 记录
func (s *CloudflareDNSService) AddRecord(ctx context.Context, domain, ip string) (string, error) {
	zoneID, err := s.getZoneID(ctx, domain)
	if err != nil {
		return "", err
	}

	rootDomain, rr := ParseDomain(domain)
	name := domain
	if rr == "@" {
		name = rootDomain
	}

	record := cloudflare.CreateDNSRecordParams{
		Type:    "A",
		Name:    name,
		Content: ip,
		TTL:     600, // 10 分钟
		Proxied: cloudflare.BoolPtr(false),
	}

	resp, err := s.api.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), record)
	if err != nil {
		return "", fmt.Errorf("添加 DNS 记录失败: %v", err)
	}

	logger.Infof("Cloudflare DNS 添加记录成功: %s -> %s, RecordId: %s", domain, ip, resp.ID)
	return resp.ID, nil
}

// UpdateRecord 更新 DNS A 记录
func (s *CloudflareDNSService) UpdateRecord(ctx context.Context, domain, ip, recordID string) error {
	zoneID, err := s.getZoneID(ctx, domain)
	if err != nil {
		return err
	}

	rootDomain, rr := ParseDomain(domain)
	name := domain
	if rr == "@" {
		name = rootDomain
	}

	record := cloudflare.UpdateDNSRecordParams{
		ID:      recordID,
		Type:    "A",
		Name:    name,
		Content: ip,
		TTL:     600,
		Proxied: cloudflare.BoolPtr(false),
	}

	_, err = s.api.UpdateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), record)
	if err != nil {
		return fmt.Errorf("更新 DNS 记录失败: %v", err)
	}

	logger.Infof("Cloudflare DNS 更新记录成功: %s -> %s, RecordId: %s", domain, ip, recordID)
	return nil
}

// DeleteRecord 删除 DNS A 记录
func (s *CloudflareDNSService) DeleteRecord(ctx context.Context, domain, recordID string) error {
	zoneID, err := s.getZoneID(ctx, domain)
	if err != nil {
		return err
	}

	err = s.api.DeleteDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), recordID)
	if err != nil {
		return fmt.Errorf("删除 DNS 记录失败: %v", err)
	}

	logger.Infof("Cloudflare DNS 删除记录成功: RecordId: %s", recordID)
	return nil
}

// TestConnection 测试 API 连接
func (s *CloudflareDNSService) TestConnection(ctx context.Context) error {
	// 使用 VerifyAPIToken 验证 Token 有效性，不需要额外权限
	_, err := s.api.VerifyAPIToken(ctx)
	if err != nil {
		return fmt.Errorf("连接测试失败: %v", err)
	}
	return nil
}

// ListDomains 获取该提供商下托管的所有域名列表
func (s *CloudflareDNSService) ListDomains(ctx context.Context) ([]string, error) {
	zones, err := s.api.ListZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取域名列表失败: %v", err)
	}

	domains := make([]string, 0, len(zones))
	for _, zone := range zones {
		domains = append(domains, zone.Name)
	}
	return domains, nil
}

// AddTXTRecord 添加 TXT 记录（用于 ACME DNS-01 验证）
func (s *CloudflareDNSService) AddTXTRecord(ctx context.Context, domain, value string) (string, error) {
	zoneID, err := s.getZoneID(ctx, domain)
	if err != nil {
		return "", err
	}

	record := cloudflare.CreateDNSRecordParams{
		Type:    "TXT",
		Name:    domain,
		Content: value,
		TTL:     120,
	}

	resp, err := s.api.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), record)
	if err != nil {
		return "", fmt.Errorf("添加 TXT 记录失败: %v", err)
	}

	logger.Infof("Cloudflare DNS 添加 TXT 记录成功: %s -> %s, RecordId: %s", domain, value, resp.ID)
	return resp.ID, nil
}

// DeleteTXTRecord 删除 TXT 记录
func (s *CloudflareDNSService) DeleteTXTRecord(ctx context.Context, domain, recordID string) error {
	zoneID, err := s.getZoneID(ctx, domain)
	if err != nil {
		return err
	}

	err = s.api.DeleteDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), recordID)
	if err != nil {
		return fmt.Errorf("删除 TXT 记录失败: %v", err)
	}

	logger.Infof("Cloudflare DNS 删除 TXT 记录成功: RecordId: %s", recordID)
	return nil
}
