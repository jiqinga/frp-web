// DNS 提供商类型
export type DNSProviderType = 'aliyun' | 'cloudflare' | 'tencent';

// DNS 提供商类型标签
export const DNS_PROVIDER_TYPE_LABELS: Record<DNSProviderType, string> = {
  aliyun: '阿里云 DNS',
  cloudflare: 'Cloudflare',
  tencent: '腾讯云 DNS',
};

// DNS 提供商认证字段配置
export const DNS_PROVIDER_AUTH_FIELDS: Record<DNSProviderType, {
  accessKeyLabel: string;
  accessKeyPlaceholder: string;
  secretKeyLabel: string;
  secretKeyPlaceholder: string;
  secretKeyRequired: boolean;
}> = {
  aliyun: {
    accessKeyLabel: 'AccessKey ID',
    accessKeyPlaceholder: '输入阿里云 AccessKey ID',
    secretKeyLabel: 'AccessKey Secret',
    secretKeyPlaceholder: '输入阿里云 AccessKey Secret',
    secretKeyRequired: true,
  },
  cloudflare: {
    accessKeyLabel: 'API Token',
    accessKeyPlaceholder: '输入 Cloudflare API Token',
    secretKeyLabel: 'API Token (确认)',
    secretKeyPlaceholder: '再次输入 API Token 确认',
    secretKeyRequired: false,
  },
  tencent: {
    accessKeyLabel: 'SecretId',
    accessKeyPlaceholder: '输入腾讯云 SecretId',
    secretKeyLabel: 'SecretKey',
    secretKeyPlaceholder: '输入腾讯云 SecretKey',
    secretKeyRequired: true,
  },
};

// DNS 提供商
export interface DNSProvider {
  id: number;
  name: string;
  type: DNSProviderType;
  access_key: string;
  secret_key?: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

// DNS 记录状态
export type DNSRecordStatus = 'pending' | 'synced' | 'failed' | 'deleted';

// DNS 记录状态标签
export const DNS_RECORD_STATUS_LABELS: Record<DNSRecordStatus, string> = {
  pending: '待同步',
  synced: '已同步',
  failed: '同步失败',
  deleted: '已删除',
};

// DNS 记录状态颜色
export const DNS_RECORD_STATUS_COLORS: Record<DNSRecordStatus, string> = {
  pending: 'orange',
  synced: 'green',
  failed: 'red',
  deleted: 'gray',
};

// DNS 记录
export interface DNSRecord {
  id: number;
  proxy_id: number;
  provider_id: number;
  domain: string;
  root_domain: string;
  record_type: string;
  record_value: string;
  record_id: string;
  status: DNSRecordStatus;
  last_error?: string;
  created_at: string;
  updated_at: string;
}

// 创建 DNS 提供商请求
export interface CreateDNSProviderRequest {
  name: string;
  type: DNSProviderType;
  access_key: string;
  secret_key: string;
  enabled?: boolean;
}

// 更新 DNS 提供商请求
export interface UpdateDNSProviderRequest {
  name?: string;
  type?: DNSProviderType;
  access_key?: string;
  secret_key?: string;
  enabled?: boolean;
}

// 测试 DNS 提供商连接请求
export interface TestDNSProviderRequest {
  type: DNSProviderType;
  access_key: string;
  secret_key?: string;
}