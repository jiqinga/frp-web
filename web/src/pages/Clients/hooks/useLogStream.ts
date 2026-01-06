import { useState, useCallback, useRef, useEffect } from 'react';
import { toast } from '../../../components/ui/Toast';
import { clientApi } from '../../../api/client';

export type LogType = 'frpc' | 'daemon';

export interface LogStreamState {
  logs: string[];
  isStreaming: boolean;
  logType: LogType | null;
  clientId: number | null;
}

export interface UseLogStreamReturn {
  logs: string[];
  isStreaming: boolean;
  logType: LogType | null;
  startLogStream: (clientId: number, logType: LogType, lines: number) => Promise<void>;
  stopLogStream: () => Promise<void>;
  clearLogs: () => void;
}

export function useLogStream(): UseLogStreamReturn {
  const [logs, setLogs] = useState<string[]>([]);
  const [isStreaming, setIsStreaming] = useState(false);
  const [logType, setLogType] = useState<LogType | null>(null);
  const clientIdRef = useRef<number | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // 清理WebSocket连接
  const cleanupWs = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
  }, []);

  // 连接WebSocket接收日志数据，返回Promise在连接成功时resolve
  const connectLogWs = useCallback((clientId: number): Promise<void> => {
    return new Promise((resolve, reject) => {
      cleanupWs();
      
      const token = localStorage.getItem('token');
      if (!token) {
        toast.error('未登录，无法连接日志流');
        reject(new Error('未登录'));
        return;
      }
      
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${protocol}//${window.location.host}/api/ws/logs/${clientId}?token=${token}`;
      
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        // WebSocket 连接成功，resolve Promise
        resolve();
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === 'log_data' && data.content) {
            setLogs(prev => [...prev, ...data.content.split('\n').filter((line: string) => line.trim())]);
          }
        } catch {
          // 如果不是JSON，直接作为日志内容
          if (event.data) {
            setLogs(prev => [...prev, event.data]);
          }
        }
      };

      ws.onerror = () => {
        // WebSocket 错误
        reject(new Error('WebSocket连接失败'));
      };

      ws.onclose = () => {
        // 如果还在streaming状态，尝试重连
        if (isStreaming && clientIdRef.current) {
          reconnectTimeoutRef.current = setTimeout(() => {
            connectLogWs(clientIdRef.current!);
          }, 3000);
        }
      };
    });
  }, [cleanupWs, isStreaming]);

  // 开始日志流
  const startLogStream = useCallback(async (clientId: number, type: LogType, lines: number) => {
    try {
      clientIdRef.current = clientId;
      setLogType(type);
      setLogs([]);
      
      // 先连接WebSocket，等待连接成功
      await connectLogWs(clientId);
      
      // WebSocket连接成功后，再发送开始日志流请求
      await clientApi.startLogStream(clientId, { log_type: type, lines });
      setIsStreaming(true);
      toast.success(`开始接收${type === 'frpc' ? 'frpc' : 'daemon'}日志`);
    } catch {
      toast.error('启动日志流失败');
      cleanupWs();
    }
  }, [connectLogWs, cleanupWs]);

  // 停止日志流
  const stopLogStream = useCallback(async () => {
    try {
      if (clientIdRef.current && logType) {
        await clientApi.stopLogStream(clientIdRef.current, logType);
      }
    } catch {
      // 静默处理停止日志流的错误
    } finally {
      cleanupWs();
      setIsStreaming(false);
      setLogType(null);
      clientIdRef.current = null;
    }
  }, [logType, cleanupWs]);

  // 清空日志
  const clearLogs = useCallback(() => {
    setLogs([]);
  }, []);

  // 组件卸载时清理
  useEffect(() => {
    return () => {
      cleanupWs();
    };
  }, [cleanupWs]);

  return {
    logs,
    isStreaming,
    logType,
    startLogStream,
    stopLogStream,
    clearLogs,
  };
}