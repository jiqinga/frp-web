import { useState, useEffect, useRef, useCallback } from 'react';
import { WebSocketClient, formatBytes } from '../utils/websocket';

export interface TrafficData {
  proxy_id: number;
  proxy_name: string;
  client_id: number;
  client_name?: string;
  bytes_in_rate: number;
  bytes_out_rate: number;
  total_bytes_in: number;
  total_bytes_out: number;
  online: boolean;
}

export interface ProxyHistory {
  time: string;
  inRate: number;
  outRate: number;
}

export interface ClientGroup {
  client_id: number;
  client_name: string;
  proxies: TrafficData[];
  totalInRate: number;
  totalOutRate: number;
  onlineCount: number;
}

export interface MonitorState {
  trafficData: TrafficData[];
  totalInRate: number;
  totalOutRate: number;
  onlineCount: number;
  totalCount: number;
  clientGroups: ClientGroup[];
  topProxies: TrafficData[];
  chartHistory: { time: string; inRate: number; outRate: number }[];
  proxyHistories: Map<number, ProxyHistory[]>;
  connected: boolean;
}

const MAX_HISTORY = 30;
const MAX_PROXY_HISTORY = 20;

export function useRealtimeMonitor() {
  const [state, setState] = useState<MonitorState>({
    trafficData: [],
    totalInRate: 0,
    totalOutRate: 0,
    onlineCount: 0,
    totalCount: 0,
    clientGroups: [],
    topProxies: [],
    chartHistory: [],
    proxyHistories: new Map(),
    connected: false,
  });

  const wsClient = useRef<WebSocketClient | null>(null);
  const proxyHistoriesRef = useRef<Map<number, ProxyHistory[]>>(new Map());

  const processTrafficData = useCallback((data: TrafficData[]) => {
    const time = new Date().toLocaleTimeString();
    const totalIn = data.reduce((sum, item) => sum + item.bytes_in_rate, 0);
    const totalOut = data.reduce((sum, item) => sum + item.bytes_out_rate, 0);
    const onlineCount = data.filter(item => item.online).length;

    // 按客户端分组
    const groupMap = new Map<number, ClientGroup>();
    data.forEach(item => {
      if (!groupMap.has(item.client_id)) {
        groupMap.set(item.client_id, {
          client_id: item.client_id,
          client_name: item.client_name || `客户端 ${item.client_id}`,
          proxies: [],
          totalInRate: 0,
          totalOutRate: 0,
          onlineCount: 0,
        });
      }
      const group = groupMap.get(item.client_id)!;
      group.proxies.push(item);
      group.totalInRate += item.bytes_in_rate;
      group.totalOutRate += item.bytes_out_rate;
      if (item.online) group.onlineCount++;
    });

    // Top 5 代理（按总速率排序）
    const topProxies = [...data]
      .sort((a, b) => (b.bytes_in_rate + b.bytes_out_rate) - (a.bytes_in_rate + a.bytes_out_rate))
      .slice(0, 5);

    // 更新每个代理的历史记录
    data.forEach(item => {
      const history = proxyHistoriesRef.current.get(item.proxy_id) || [];
      history.push({ time, inRate: item.bytes_in_rate, outRate: item.bytes_out_rate });
      if (history.length > MAX_PROXY_HISTORY) history.shift();
      proxyHistoriesRef.current.set(item.proxy_id, history);
    });

    setState(prev => {
      const newChartHistory = [...prev.chartHistory, { time, inRate: totalIn, outRate: totalOut }];
      if (newChartHistory.length > MAX_HISTORY) newChartHistory.shift();

      return {
        ...prev,
        trafficData: data,
        totalInRate: totalIn,
        totalOutRate: totalOut,
        onlineCount,
        totalCount: data.length,
        clientGroups: Array.from(groupMap.values()),
        topProxies,
        chartHistory: newChartHistory,
        proxyHistories: new Map(proxyHistoriesRef.current),
      };
    });
  }, []);

  useEffect(() => {
    let mounted = true;
    
    // 延迟连接以避免 React StrictMode 双重调用问题
    const connectTimeout = setTimeout(() => {
      if (!mounted) return;
      
      const token = localStorage.getItem('token') || '';
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.host;
      const wsUrl = `${protocol}//${host}/api/ws/realtime`;

      wsClient.current = new WebSocketClient();
      wsClient.current.connect(wsUrl, token);

      wsClient.current.onMessage('traffic_update', (data: unknown) => {
        if (!mounted) return;
        const message = data as { type: string; data: TrafficData[] };
        processTrafficData(message.data);
        setState(prev => ({ ...prev, connected: true }));
      });
    }, 100);

    return () => {
      mounted = false;
      clearTimeout(connectTimeout);
      wsClient.current?.disconnect();
    };
  }, [processTrafficData]);

  const getProxyHistory = useCallback((proxyId: number): ProxyHistory[] => {
    return state.proxyHistories.get(proxyId) || [];
  }, [state.proxyHistories]);

  return { ...state, getProxyHistory, formatBytes };
}