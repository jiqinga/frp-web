package service

import (
	"context"
	"fmt"
	"frp-web-panel/internal/logger"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

// TencentDNSService 腾讯云 DNS 服务 (DNSPod)
// 使用腾讯云官方 SDK (Apache-2.0 License)
// 文档: https://cloud.tencent.com/document/product/1427
type TencentDNSService struct {
	client *dnspod.Client
}

// NewTencentDNSService 创建腾讯云 DNS 服务实例
// secretID: 腾讯云 SecretId
// secretKey: 腾讯云 SecretKey
func NewTencentDNSService(secretID, secretKey string) (*TencentDNSService, error) {
	credential := common.NewCredential(secretID, secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"

	client, err := dnspod.NewClient(credential, "", cpf)
	if err != nil {
		return nil, fmt.Errorf("创建腾讯云 DNS 客户端失败: %w", err)
	}

	return &TencentDNSService{
		client: client,
	}, nil
}

// extractDomainParts 从完整域名中提取主域名和子域名
// 例如: app.example.com -> domain: example.com, subdomain: app
func extractDomainParts(fullDomain string) (domain, subdomain string) {
	parts := strings.Split(fullDomain, ".")
	if len(parts) >= 2 {
		domain = strings.Join(parts[len(parts)-2:], ".")
		if len(parts) > 2 {
			subdomain = strings.Join(parts[:len(parts)-2], ".")
		} else {
			subdomain = "@"
		}
	} else {
		domain = fullDomain
		subdomain = "@"
	}
	return
}

// AddRecord 添加 DNS A 记录
func (s *TencentDNSService) AddRecord(ctx context.Context, fullDomain, ip string) (string, error) {
	domain, subdomain := extractDomainParts(fullDomain)

	// 先检查记录是否已存在
	existingRecordID, err := s.findRecord(ctx, domain, subdomain, ip)
	if err == nil && existingRecordID != "" {
		logger.Infof("腾讯云DNS DNS 记录已存在: %s -> %s (RecordID: %s)", fullDomain, ip, existingRecordID)
		return existingRecordID, nil
	}

	// 创建新记录
	request := dnspod.NewCreateRecordRequest()
	request.Domain = common.StringPtr(domain)
	request.SubDomain = common.StringPtr(subdomain)
	request.RecordType = common.StringPtr("A")
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(ip)
	request.TTL = common.Uint64Ptr(600)

	response, err := s.client.CreateRecord(request)
	if err != nil {
		return "", fmt.Errorf("创建 DNS 记录失败: %w", err)
	}

	recordID := fmt.Sprintf("%d", *response.Response.RecordId)
	logger.Infof("腾讯云DNS DNS 记录已创建: %s -> %s (RecordID: %s)", fullDomain, ip, recordID)
	return recordID, nil
}

// findRecord 查找已存在的记录
func (s *TencentDNSService) findRecord(ctx context.Context, domain, subdomain, ip string) (string, error) {
	request := dnspod.NewDescribeRecordListRequest()
	request.Domain = common.StringPtr(domain)
	request.Subdomain = common.StringPtr(subdomain)
	request.RecordType = common.StringPtr("A")

	response, err := s.client.DescribeRecordList(request)
	if err != nil {
		return "", err
	}

	for _, record := range response.Response.RecordList {
		if *record.Value == ip {
			return fmt.Sprintf("%d", *record.RecordId), nil
		}
	}

	return "", fmt.Errorf("未找到匹配的记录")
}

// UpdateRecord 更新 DNS A 记录
func (s *TencentDNSService) UpdateRecord(ctx context.Context, fullDomain, ip, recordID string) error {
	domain, subdomain := extractDomainParts(fullDomain)

	request := dnspod.NewModifyRecordRequest()
	request.Domain = common.StringPtr(domain)
	request.SubDomain = common.StringPtr(subdomain)
	request.RecordType = common.StringPtr("A")
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(ip)
	request.TTL = common.Uint64Ptr(600)

	// 将 recordID 转换为 uint64
	var recordIDUint uint64
	fmt.Sscanf(recordID, "%d", &recordIDUint)
	request.RecordId = common.Uint64Ptr(recordIDUint)

	_, err := s.client.ModifyRecord(request)
	if err != nil {
		return fmt.Errorf("更新 DNS 记录失败: %w", err)
	}

	logger.Infof("腾讯云DNS DNS 记录已更新: %s -> %s", fullDomain, ip)
	return nil
}

// DeleteRecord 删除 DNS A 记录
func (s *TencentDNSService) DeleteRecord(ctx context.Context, fullDomain, recordID string) error {
	domain, _ := extractDomainParts(fullDomain)

	request := dnspod.NewDeleteRecordRequest()
	request.Domain = common.StringPtr(domain)

	// 将 recordID 转换为 uint64
	var recordIDUint uint64
	fmt.Sscanf(recordID, "%d", &recordIDUint)
	request.RecordId = common.Uint64Ptr(recordIDUint)

	_, err := s.client.DeleteRecord(request)
	if err != nil {
		return fmt.Errorf("删除 DNS 记录失败: %w", err)
	}

	logger.Infof("腾讯云DNS DNS 记录已删除: %s (RecordID: %s)", fullDomain, recordID)
	return nil
}

// TestConnection 测试腾讯云 DNS API 连接
func (s *TencentDNSService) TestConnection(ctx context.Context) error {
	// 尝试获取域名列表来验证凭证
	request := dnspod.NewDescribeDomainListRequest()
	request.Limit = common.Int64Ptr(1)

	_, err := s.client.DescribeDomainList(request)
	if err != nil {
		return fmt.Errorf("API 凭证验证失败: %w", err)
	}

	logger.Info("腾讯云DNS API 凭证验证成功")
	return nil
}

// ListDomains 获取该提供商下托管的所有域名列表
func (s *TencentDNSService) ListDomains(ctx context.Context) ([]string, error) {
	request := dnspod.NewDescribeDomainListRequest()
	request.Limit = common.Int64Ptr(100)

	response, err := s.client.DescribeDomainList(request)
	if err != nil {
		return nil, fmt.Errorf("获取域名列表失败: %w", err)
	}

	domains := make([]string, 0)
	if response.Response != nil && response.Response.DomainList != nil {
		for _, domain := range response.Response.DomainList {
			if domain.Name != nil {
				domains = append(domains, *domain.Name)
			}
		}
	}
	return domains, nil
}

// AddTXTRecord 添加 TXT 记录（用于 ACME DNS-01 验证）
func (s *TencentDNSService) AddTXTRecord(ctx context.Context, fullDomain, value string) (string, error) {
	domain, subdomain := extractDomainParts(fullDomain)

	request := dnspod.NewCreateRecordRequest()
	request.Domain = common.StringPtr(domain)
	request.SubDomain = common.StringPtr(subdomain)
	request.RecordType = common.StringPtr("TXT")
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(value)
	request.TTL = common.Uint64Ptr(120)

	response, err := s.client.CreateRecord(request)
	if err != nil {
		return "", fmt.Errorf("创建 TXT 记录失败: %w", err)
	}

	recordID := fmt.Sprintf("%d", *response.Response.RecordId)
	logger.Infof("腾讯云DNS TXT 记录已创建: %s -> %s (RecordID: %s)", fullDomain, value, recordID)
	return recordID, nil
}

// DeleteTXTRecord 删除 TXT 记录
func (s *TencentDNSService) DeleteTXTRecord(ctx context.Context, fullDomain, recordID string) error {
	domain, _ := extractDomainParts(fullDomain)

	request := dnspod.NewDeleteRecordRequest()
	request.Domain = common.StringPtr(domain)

	var recordIDUint uint64
	fmt.Sscanf(recordID, "%d", &recordIDUint)
	request.RecordId = common.Uint64Ptr(recordIDUint)

	_, err := s.client.DeleteRecord(request)
	if err != nil {
		return fmt.Errorf("删除 TXT 记录失败: %w", err)
	}

	logger.Infof("腾讯云DNS TXT 记录已删除: %s (RecordID: %s)", fullDomain, recordID)
	return nil
}
