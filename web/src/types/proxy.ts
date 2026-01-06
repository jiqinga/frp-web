// 代理类型枚举
export type ProxyType = 'tcp' | 'udp' | 'http' | 'https' | 'stcp';

// 代理类型字段配置
export const PROXY_TYPE_FIELDS: Record<ProxyType, string[]> = {
  tcp: ['name', 'local_ip', 'local_port', 'remote_port', 'bandwidth_limit'],
  udp: ['name', 'local_ip', 'local_port', 'remote_port', 'bandwidth_limit'],
  http: ['name', 'local_ip', 'local_port', 'custom_domains', 'subdomain', 'locations', 'host_header_rewrite', 'http_user', 'http_password', 'bandwidth_limit'],
  https: ['name', 'local_ip', 'local_port', 'custom_domains', 'subdomain', 'host_header_rewrite', 'bandwidth_limit'],
  stcp: ['name', 'local_ip', 'local_port', 'secret_key', 'allow_users', 'bandwidth_limit'],
};

// 代理类型必填字段配置
export const PROXY_REQUIRED_FIELDS: Record<ProxyType, string[]> = {
  tcp: ['name', 'local_ip', 'local_port'],
  udp: ['name', 'local_ip', 'local_port'],
  http: ['name', 'local_ip', 'local_port'],
  https: ['name', 'local_ip', 'local_port'],
  stcp: ['name', 'local_ip', 'local_port', 'secret_key'],
};

// 代理类型标签
export const PROXY_TYPE_LABELS: Record<ProxyType, string> = {
  tcp: 'TCP',
  udp: 'UDP',
  http: 'HTTP',
  https: 'HTTPS',
  stcp: 'STCP',
};

export interface Proxy {
  id: number;
  client_id: number;
  name: string;
  type: string;
  local_ip: string;
  local_port: number;
  remote_port?: number;
  custom_domains?: string;
  subdomain?: string;
  locations?: string;
  host_header_rewrite?: string;
  http_user?: string;
  http_password?: string;
  secret_key?: string;
  allow_users?: string;
  bandwidth_limit?: string;
  enabled?: boolean;
  // 插件配置字段
  plugin_type?: string;
  plugin_config?: string;
  // DNS 同步字段
  enable_dns_sync?: boolean;
  dns_provider_id?: number;
  dns_root_domain?: string;
  // 自动证书字段
  auto_cert?: boolean;
  cert_id?: number;
  total_bytes_in: number;
  total_bytes_out: number;
  current_bytes_in_rate: number;
  current_bytes_out_rate: number;
  last_online_time?: string;
  created_at: string;
  updated_at: string;
}