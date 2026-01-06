import request from './request';

export interface FrpsMetrics {
  client_counts: number;
  proxy_counts: Record<string, number>;
  total_proxies: number;
  traffic_in: number;
  traffic_out: number;
  cpu_seconds: number;
  memory_bytes: number;
  start_time: number;
  uptime: number;
  goroutines: number;
}

export interface ServerMetricsHistory {
  id: number;
  server_id: number;
  cpu_percent: number;
  memory_bytes: number;
  traffic_in: number;
  traffic_out: number;
  record_time: string;
}

export interface FrpServer {
  id?: number;
  name: string;
  server_type?: 'local' | 'remote';
  host: string;
  dashboard_port: number;
  dashboard_user: string;
  dashboard_pwd: string;
  bind_port: number;
  token?: string;
  ssh_host?: string;
  ssh_port?: number;
  ssh_user?: string;
  ssh_password?: string;
  install_path?: string;
  mirror_id?: number;
  enabled: boolean;
  status?: string;
  pid?: number;
  version?: string;
  binary_path?: string;
  config_path?: string;
  last_sync_time?: string;
  last_error?: string;
  created_at?: string;
  updated_at?: string;
}

export const frpServerApi = {
  getAll: () => request.get<FrpServer[]>('/frp-servers'),
  
  getById: (id: number) => request.get<FrpServer>(`/frp-servers/${id}`),
  
  create: (data: FrpServer) => request.post<FrpServer>('/frp-servers', data),
  
  update: (id: number, data: FrpServer) => request.put<FrpServer>(`/frp-servers/${id}`, data),
  
  delete: (id: number, removeInstallation?: boolean) => request.delete(`/frp-servers/${id}`, { params: { remove_installation: removeInstallation } }),
  
  testConnection: (data: FrpServer) => request.post('/frp-servers/test', data),
  
  start: (id: number) => request.post(`/frp-servers/${id}/start`),
  
  stop: (id: number) => request.post(`/frp-servers/${id}/stop`),
  
  restart: (id: number) => request.post(`/frp-servers/${id}/restart`),
  
  getStatus: (id: number) => request.get<{ status: string }>(`/frp-servers/${id}/status`),
  
  download: (id: number, version?: string) => request.post(`/frp-servers/${id}/download`, { version }),
  
  testSSH: (id: number) => request.post(`/frp-servers/${id}/test-ssh`),
  
  remoteInstall: (id: number, mirrorId?: number) => request.post(`/frp-servers/${id}/remote-install`, { mirror_id: mirrorId }),
  
  remoteStart: (id: number) => request.post(`/frp-servers/${id}/remote-start`),
  
  remoteStop: (id: number) => request.post(`/frp-servers/${id}/remote-stop`),
  
  remoteRestart: (id: number) => request.post(`/frp-servers/${id}/remote-restart`),
  
  remoteUninstall: (id: number) => request.post(`/frp-servers/${id}/remote-uninstall`),
  
  remoteGetLogs: (id: number, lines?: number) => request.get<{ logs: string }>(`/frp-servers/${id}/remote-logs`, { params: { lines } }),
  
  remoteGetVersion: (id: number) => request.get<{ version: string }>(`/frp-servers/${id}/remote-version`),
  
  getLocalVersion: (id: number) => request.get<{ version: string }>(`/frp-servers/${id}/local-version`),
  
  remoteReinstall: (id: number, regenerateAuth?: boolean, mirrorId?: number) => request.post(`/frp-servers/${id}/remote-reinstall`, { regenerate_auth: regenerateAuth, mirror_id: mirrorId }),
  
  remoteUpgrade: (id: number, version?: string, mirrorId?: number) => request.post(`/frp-servers/${id}/remote-upgrade`, { version, mirror_id: mirrorId }),
  
  getRunningTask: (id: number) => request.get<{ running: boolean; operation?: string }>(`/frp-servers/${id}/running-task`),
  
  parseConfig: (config: string) => request.post<{ bind_port: number; token: string; host: string; dashboard_port: number; dashboard_user: string; dashboard_pwd: string }>('/frp-servers/parse-config', { config }),
  
  getMetrics: (id: number) => request.get<FrpsMetrics>(`/frp-servers/${id}/metrics`),
  
  getMetricsHistory: (id: number, days?: number) => request.get<ServerMetricsHistory[]>(`/frp-servers/${id}/metrics-history`, { params: { days } }),
};