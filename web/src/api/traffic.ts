import request from './request';
import type { TrafficSummary, TrafficStats, ProxyTrafficSummary, TrafficTrendPoint } from '../types';

export const trafficApi = {
  getSummary: () => request.get<TrafficSummary>('/traffic/summary'),
  
  getTrend: (hours?: number) =>
    request.get<TrafficTrendPoint[]>('/traffic/trend', { params: { hours } }),
  
  getHistory: (proxyId: number, start?: string, end?: string) =>
    request.get<TrafficStats[]>(`/traffic/proxy/${proxyId}`, { params: { start, end } }),

  getProxiesTrafficSummary: (hours?: number) =>
    request.get<Record<string, ProxyTrafficSummary>>('/traffic/proxies/summary', { params: { hours } })
};