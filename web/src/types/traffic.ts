export interface TrafficStats {
  id: number;
  proxy_id: number;
  bytes_in: number;
  bytes_out: number;
  current_rate_in: number;
  current_rate_out: number;
  record_time: string;
}

export interface TrafficSummary {
  total_bytes_in: number;
  total_bytes_out: number;
  current_rate_in: number;
  current_rate_out: number;
  active_proxies: number;
  total_proxies: number;
}

export interface ProxyTrafficSummary {
  total_in: number;
  total_out: number;
}

export interface TrafficTrendPoint {
  time: string;
  inbound: number;
  outbound: number;
}