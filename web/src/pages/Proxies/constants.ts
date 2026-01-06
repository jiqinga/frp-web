// 代理类型颜色配置
export const PROXY_TYPE_COLORS: Record<string, string> = {
  tcp: 'blue',
  udp: 'green',
  http: 'orange',
  https: 'red',
  stcp: 'purple',
};

// 字段标签配置
export const FIELD_LABELS: Record<string, string> = {
  name: '名称',
  local_ip: '本地IP',
  local_port: '本地端口',
  remote_port: '远程端口',
  custom_domains: '自定义域名',
  subdomain: '子域名',
  locations: 'URL路由',
  host_header_rewrite: 'Host重写',
  http_user: 'HTTP用户名',
  http_password: 'HTTP密码',
  secret_key: '密钥',
  allow_users: '允许用户',
  bandwidth_limit: '带宽限制',
};

// 字段提示配置
export const FIELD_TOOLTIPS: Record<string, string> = {
  remote_port: 'TCP/UDP 类型的代理如果不填写，系统将自动分配一个可用端口（10000-65535）',
  custom_domains: '访问该代理的自定义域名，如 example.com。HTTP/HTTPS 类型可选，如果不填写可以使用子域名',
  subdomain: '使用 frps 配置的子域名前缀，如填写 test 则访问地址为 test.your-frps-domain.com',
  locations: 'URL路由匹配规则，多个路径用逗号分隔，如 /api,/admin',
  host_header_rewrite: '重写发送到后端服务的 Host 头，用于虚拟主机场景',
  http_user: 'HTTP Basic 认证用户名，留空表示不启用认证',
  http_password: 'HTTP Basic 认证密码',
  secret_key: 'STCP 代理的密钥，访问者需要使用相同的密钥才能连接',
  allow_users: '允许访问的用户列表，多个用户用逗号分隔，留空表示允许所有用户',
  bandwidth_limit: '带宽限制，如 1MB 或 500KB',
};

// 字段占位符配置
export const FIELD_PLACEHOLDERS: Record<string, string> = {
  name: '请输入代理名称',
  local_ip: '127.0.0.1',
  local_port: '请输入本地端口',
  remote_port: '留空自动分配',
  custom_domains: 'example.com',
  subdomain: 'test',
  locations: '/api,/admin',
  host_header_rewrite: 'internal.example.com',
  http_user: 'admin',
  http_password: '请输入密码',
  secret_key: '请输入密钥',
  allow_users: 'user1,user2',
  bandwidth_limit: '1MB',
};

// 带宽单位选项
export const BANDWIDTH_UNITS = [
  { value: 'KB', label: 'KB/s' },
  { value: 'MB', label: 'MB/s' },
];

// 解析带宽限制字符串，返回数值和单位
export const parseBandwidthLimit = (value?: string): { value: number | undefined; unit: string } => {
  if (!value) return { value: undefined, unit: 'MB' };
  const match = value.match(/^(\d+(?:\.\d+)?)\s*(KB|MB)?$/i);
  if (match) {
    return {
      value: parseFloat(match[1]),
      unit: (match[2] || 'MB').toUpperCase(),
    };
  }
  return { value: undefined, unit: 'MB' };
};

// 格式化带宽限制为字符串
export const formatBandwidthLimit = (value: number | undefined, unit: string): string | undefined => {
  if (value === undefined || value === null || value === 0) return undefined;
  return `${value}${unit}`;
};

// 重新导出 types 中的常量，方便使用
export {
  PROXY_TYPE_FIELDS,
  PROXY_REQUIRED_FIELDS,
  PROXY_TYPE_LABELS,
  PLUGIN_TYPE_FIELDS,
  PLUGIN_REQUIRED_FIELDS,
  PLUGIN_TYPE_LABELS,
  PLUGIN_TYPE_COLORS,
  PLUGIN_FIELD_LABELS,
  PLUGIN_FIELD_TOOLTIPS,
  PLUGIN_FIELD_PLACEHOLDERS,
} from '../../types';