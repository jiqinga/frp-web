// 配置同步状态类型
export type ConfigSyncStatus = 'synced' | 'failed' | 'pending' | 'rolled_back';

// 配置同步状态标签
export const CONFIG_SYNC_STATUS_LABELS: Record<ConfigSyncStatus, string> = {
  synced: '已同步',
  failed: '同步失败',
  pending: '同步中',
  rolled_back: '已回滚',
};

// 配置同步状态颜色（对应 Badge variant）
export const CONFIG_SYNC_STATUS_VARIANTS: Record<ConfigSyncStatus, 'success' | 'danger' | 'warning' | 'info'> = {
  synced: 'success',
  failed: 'danger',
  pending: 'warning',
  rolled_back: 'info',
};

export interface Client {
  id: number;
  name: string;
  remark: string;
  server_addr: string;
  server_port: number;
  token: string;
  protocol: string;
  frpc_admin_host?: string;
  frpc_admin_port?: number;
  frpc_admin_user?: string;
  frpc_admin_pwd?: string;
  frp_server_id?: number;
  online_status?: string;
  last_heartbeat?: string;
  config_version?: number;
  ws_connected?: boolean;
  last_config_sync?: string;
  // 版本信息字段
  frpc_version?: string;
  daemon_version?: string;
  os?: string;
  arch?: string;
  // 配置同步状态字段
  config_sync_status?: ConfigSyncStatus;
  config_sync_error?: string;
  config_sync_time?: string;
  created_at: string;
  updated_at: string;
}