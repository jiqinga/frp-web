import request from './request';

export type AlertTargetType = 'proxy' | 'frpc' | 'frps' | 'system';

// 系统级告警规则类型
export type SystemRuleType =
  | 'cert_apply_success' | 'cert_apply_failed'
  | 'cert_expiring' | 'cert_expired'
  | 'cert_renew_success' | 'cert_renew_failed'
  | 'dns_sync_success' | 'dns_sync_failed'
  | 'login_failed' | 'config_changed';

export interface AlertRule {
  id?: number;
  target_type: AlertTargetType;
  target_id: number;
  proxy_id: number; // 保留兼容
  rule_type: string; // daily, monthly, rate, offline, 或系统级规则类型
  threshold_value: number;
  threshold_unit: string;
  cooldown_minutes: number;
  offline_delay_seconds: number; // 离线延迟确认时间（秒）
  notify_on_recovery: boolean; // 恢复时是否通知
  enabled: boolean;
  notify_recipient_ids?: string; // 接收人ID列表，逗号分隔
  notify_group_ids?: string; // 分组ID列表，逗号分隔
  notify_webhook?: string;
  created_at?: string;
  updated_at?: string;
}

export interface AlertLog {
  id: number;
  rule_id: number;
  target_type: AlertTargetType;
  target_id: number;
  proxy_id: number;
  alert_type: string;
  current_value: number;
  threshold_value: number;
  message: string;
  event_data?: string; // 系统告警事件详情JSON
  notified: boolean;
  created_at: string;
}

// 系统规则类型名称映射
export const systemRuleTypeNames: Record<SystemRuleType, string> = {
  cert_apply_success: '证书申请成功',
  cert_apply_failed: '证书申请失败',
  cert_expiring: '证书即将过期',
  cert_expired: '证书已过期',
  cert_renew_success: '证书续签成功',
  cert_renew_failed: '证书续签失败',
  dns_sync_success: 'DNS同步成功',
  dns_sync_failed: 'DNS同步失败',
  login_failed: '登录失败',
  config_changed: '配置变更',
};

export const alertApi = {
  createRule: (data: AlertRule) => request.post('/alerts/rules', data),
  getAllRules: () => request.get('/alerts/rules'),
  getRulesByProxyID: (proxyId: number) => request.get(`/alerts/rules/proxy/${proxyId}`),
  updateRule: (data: AlertRule) => request.put('/alerts/rules', data),
  deleteRule: (id: number) => request.delete(`/alerts/rules/${id}`),
  getAlertLogs: (limit?: number) => request.get('/alerts/logs', { params: { limit } }),
};