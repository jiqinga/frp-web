/*
 * @Author              : 寂情啊
 * @Date                : 2025-12-22 15:48:28
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2026-01-07 11:15:21
 * @FilePath            : frp-web-testbackendinternalservicealiyun_dns_service.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package service

import (
	"context"
	"fmt"
	"frp-web-panel/internal/logger"
	"strings"

	alidns "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

type AliyunDNSService struct {
	client *alidns.Client
}

func NewAliyunDNSService(accessKey, secretKey string) (*AliyunDNSService, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(accessKey),
		AccessKeySecret: tea.String(secretKey),
		Endpoint:        tea.String("alidns.cn-hangzhou.aliyuncs.com"),
	}
	client, err := alidns.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("创建阿里云DNS客户端失败: %v", err)
	}
	return &AliyunDNSService{client: client}, nil
}

// AddRecord 添加DNS记录
func (s *AliyunDNSService) AddRecord(domain, rr, recordType, value string) (string, error) {
	request := &alidns.AddDomainRecordRequest{
		DomainName: tea.String(domain),
		RR:         tea.String(rr),
		Type:       tea.String(recordType),
		Value:      tea.String(value),
	}
	response, err := s.client.AddDomainRecord(request)
	if err != nil {
		return "", fmt.Errorf("添加DNS记录失败: %v", err)
	}
	logger.Infof("阿里云DNS 添加记录成功: %s.%s -> %s, RecordId: %s", rr, domain, value, *response.Body.RecordId)
	return *response.Body.RecordId, nil
}

// DeleteRecord 删除DNS记录
func (s *AliyunDNSService) DeleteRecord(recordID string) error {
	request := &alidns.DeleteDomainRecordRequest{
		RecordId: tea.String(recordID),
	}
	_, err := s.client.DeleteDomainRecord(request)
	if err != nil {
		return fmt.Errorf("删除DNS记录失败: %v", err)
	}
	logger.Infof("阿里云DNS 删除记录成功: RecordId: %s", recordID)
	return nil
}

// TestConnection 测试连接
func (s *AliyunDNSService) TestConnection() error {
	request := &alidns.DescribeDomainsRequest{
		PageNumber: tea.Int64(1),
		PageSize:   tea.Int64(1),
	}
	_, err := s.client.DescribeDomains(request)
	if err != nil {
		return fmt.Errorf("连接测试失败: %v", err)
	}
	return nil
}

// ParseDomain 解析域名，返回根域名和子域名前缀
func ParseDomain(fullDomain string) (rootDomain, rr string) {
	// 去除 FQDN 末尾的点（如 "_acme-challenge.hts.jiqinga.top." -> "_acme-challenge.hts.jiqinga.top"）
	fullDomain = strings.TrimSuffix(fullDomain, ".")
	parts := strings.Split(fullDomain, ".")
	if len(parts) < 2 {
		return fullDomain, "@"
	}
	if len(parts) == 2 {
		return fullDomain, "@"
	}
	rootDomain = strings.Join(parts[len(parts)-2:], ".")
	rr = strings.Join(parts[:len(parts)-2], ".")
	return
}

// AliyunDNSServiceV2 实现 DNSOperator 接口的阿里云 DNS 服务
type AliyunDNSServiceV2 struct {
	*AliyunDNSService
}

// NewAliyunDNSServiceV2 创建实现 DNSOperator 接口的阿里云 DNS 服务
func NewAliyunDNSServiceV2(accessKey, secretKey string) (*AliyunDNSServiceV2, error) {
	base, err := NewAliyunDNSService(accessKey, secretKey)
	if err != nil {
		return nil, err
	}
	return &AliyunDNSServiceV2{AliyunDNSService: base}, nil
}

// AddRecord 实现 DNSOperator 接口 - 添加 DNS A 记录
func (s *AliyunDNSServiceV2) AddRecord(ctx context.Context, domain, ip string) (string, error) {
	rootDomain, rr := ParseDomain(domain)
	return s.AliyunDNSService.AddRecord(rootDomain, rr, "A", ip)
}

// UpdateRecord 实现 DNSOperator 接口 - 更新 DNS A 记录
func (s *AliyunDNSServiceV2) UpdateRecord(ctx context.Context, domain, ip, recordID string) error {
	// 阿里云更新记录需要先删除再添加
	if err := s.AliyunDNSService.DeleteRecord(recordID); err != nil {
		return err
	}
	rootDomain, rr := ParseDomain(domain)
	_, err := s.AliyunDNSService.AddRecord(rootDomain, rr, "A", ip)
	return err
}

// DeleteRecord 实现 DNSOperator 接口 - 删除 DNS A 记录
func (s *AliyunDNSServiceV2) DeleteRecord(ctx context.Context, domain, recordID string) error {
	return s.AliyunDNSService.DeleteRecord(recordID)
}

// TestConnection 实现 DNSOperator 接口 - 测试连接
func (s *AliyunDNSServiceV2) TestConnection(ctx context.Context) error {
	return s.AliyunDNSService.TestConnection()
}

// ListDomains 实现 DNSOperator 接口 - 获取域名列表
func (s *AliyunDNSServiceV2) ListDomains(ctx context.Context) ([]string, error) {
	request := &alidns.DescribeDomainsRequest{
		PageNumber: tea.Int64(1),
		PageSize:   tea.Int64(100),
	}
	response, err := s.client.DescribeDomains(request)
	if err != nil {
		return nil, fmt.Errorf("获取域名列表失败: %v", err)
	}

	domains := make([]string, 0)
	if response.Body != nil && response.Body.Domains != nil {
		for _, domain := range response.Body.Domains.Domain {
			if domain.DomainName != nil {
				domains = append(domains, *domain.DomainName)
			}
		}
	}
	return domains, nil
}

// AddTXTRecord 实现 DNSOperator 接口 - 添加 TXT 记录（用于 ACME DNS-01 验证）
func (s *AliyunDNSServiceV2) AddTXTRecord(ctx context.Context, domain, value string) (string, error) {
	rootDomain, rr := ParseDomain(domain)
	return s.AliyunDNSService.AddRecord(rootDomain, rr, "TXT", value)
}

// DeleteTXTRecord 实现 DNSOperator 接口 - 删除 TXT 记录
func (s *AliyunDNSServiceV2) DeleteTXTRecord(ctx context.Context, domain, recordID string) error {
	return s.AliyunDNSService.DeleteRecord(recordID)
}
