import { useState, useRef, useCallback, useEffect } from 'react';
import { toast } from '../../../components/ui/Toast';
import type { Client } from '../../../types';

export interface UpdateProgress {
  stage: string;
  progress: number;
  message: string;
  totalBytes: number;
  downloadedBytes: number;
}

export interface UpdateResult {
  success: boolean;
  version: string;
  message: string;
}

export interface UseUpdateWebSocketReturn {
  updateProgress: UpdateProgress | null;
  updateResult: UpdateResult | null;
  connectWebSocket: (client: Client, onSuccess: () => void) => () => void;
  resetProgress: () => void;
}

export function useUpdateWebSocket(): UseUpdateWebSocketReturn {
  const [updateProgress, setUpdateProgress] = useState<UpdateProgress | null>(null);
  const [updateResult, setUpdateResult] = useState<UpdateResult | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const clientRef = useRef<Client | null>(null);
  const onSuccessRef = useRef<(() => void) | null>(null);

  const resetProgress = useCallback(() => {
    setUpdateProgress(null);
    setUpdateResult(null);
  }, []);

  const connectWebSocket = useCallback((client: Client, onSuccess: () => void) => {
    clientRef.current = client;
    onSuccessRef.current = onSuccess;

    const token = localStorage.getItem('token');
    if (!token) return () => {};

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.host}/api/ws/realtime?token=${token}`;
    
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type === 'client_update_progress') {
          const progressData = data.data;
          if (clientRef.current && progressData.client_id === clientRef.current.id) {
            setUpdateProgress({
              stage: progressData.stage,
              progress: progressData.progress,
              message: progressData.message,
              totalBytes: progressData.total_bytes,
              downloadedBytes: progressData.downloaded_bytes,
            });
          }
        } else if (data.type === 'client_update_result') {
          const resultData = data.data;
          if (clientRef.current && resultData.client_id === clientRef.current.id) {
            setUpdateResult({
              success: resultData.success,
              version: resultData.version,
              message: resultData.message,
            });
            if (resultData.success) {
              toast.success(`客户端 ${clientRef.current.name} 更新成功`);
              onSuccessRef.current?.();
            } else {
              toast.error(`客户端 ${clientRef.current.name} 更新失败: ${resultData.message}`);
            }
          }
        }
      } catch (e) {
        console.error('解析WebSocket消息失败:', e);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket错误:', error);
    };

    ws.onclose = () => {
      console.log('WebSocket连接关闭');
    };

    return () => {
      ws.close();
      wsRef.current = null;
    };
  }, []);

  // 清理 WebSocket 连接
  useEffect(() => {
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, []);

  return {
    updateProgress,
    updateResult,
    connectWebSocket,
    resetProgress,
  };
}

// 更新阶段标签
export const STAGE_LABELS: Record<string, string> = {
  downloading: '下载中',
  stopping: '停止旧进程',
  replacing: '替换文件',
  starting: '启动新进程',
  completed: '完成',
  failed: '失败',
};