import request from './request';

// 监控概览数据
export interface OverviewData {
  total_clients: number;
  total_proxies: number;
  active_proxies: number;
  total_bytes_in: number;
  total_bytes_out: number;
  current_rate_in: number;
  current_rate_out: number;
}

// 统计数据
export interface StatsData {
  proxy_type_stats: Record<string, number>;
  recent_logs: unknown[];
}

// 实时流量数据（来自 WebSocket）
export interface TrafficData {
  proxy_id: number;
  proxy_name: string;
  client_id: number;
  bytes_in_rate: number;
  bytes_out_rate: number;
  total_bytes_in: number;
  total_bytes_out: number;
  online: boolean;
}

export const monitorApi = {
  // 获取监控概览
  getOverview: () => request.get<OverviewData>('/monitor/overview'),
  
  // 获取统计数据
  getStats: () => request.get<StatsData>('/monitor/stats'),
};