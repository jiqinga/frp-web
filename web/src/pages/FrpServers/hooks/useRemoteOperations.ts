import { useState, useCallback, useRef, useEffect } from 'react';
import { toast } from '../../../components/ui/Toast';
import { frpServerApi } from '../../../api/frpServer';
import { WebSocketClient } from '../../../utils/websocket';
import type { SSHLogMessage, EnhancedLogEntry } from '../../../types';

/**
 * 远程操作管理 Hook
 * 
 * 功能：
 * - 远程安装/重装/升级
 * - WebSocket 日志监听
 * - 远程启动/停止/重启
 * - 查看日志和认证信息
 */
export function useRemoteOperations() {
  const [installLogs, setInstallLogs] = useState<EnhancedLogEntry[]>([]);
  const [downloadProgress, setDownloadProgress] = useState(0);
  const wsClient = useRef<WebSocketClient | null>(null);
  const logEndRef = useRef<HTMLDivElement>(null);

  // 自动滚动到日志底部
  useEffect(() => {
    if (logEndRef.current) {
      logEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [installLogs]);

  // 连接 WebSocket 并监听 SSH 日志
  const connectWebSocket = useCallback((serverId: number, operation: string, onComplete?: () => void) => {
    setInstallLogs([{
      message: `正在${operation}...`,
      type: 'info',
      timestamp: new Date().toLocaleTimeString('zh-CN', { hour12: false })
    }]);

    const token = localStorage.getItem('token') || '';
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/api/ws/realtime`;
    
    // 先断开旧连接再创建新连接
    if (wsClient.current) {
      wsClient.current.disconnect();
    }
    wsClient.current = new WebSocketClient();
    wsClient.current.connect(wsUrl, token);
    
    wsClient.current.onMessage('ssh_log', (data: unknown) => {
      const msg = data as SSHLogMessage;
      
      if (msg.server_id === serverId) {
        const entry: EnhancedLogEntry = {
          message: msg.log,
          type: msg.log_type,
          progress: msg.progress,
          timestamp: msg.timestamp
        };
        
        setInstallLogs(prev => [...prev, entry]);
        
        if (msg.log_type === 'progress' && msg.progress) {
          setDownloadProgress(msg.progress);
        }
        
        if (msg.log.includes('完成') || msg.log.includes('失败')) {
          setTimeout(() => {
            setDownloadProgress(0);
            if (onComplete) {
              onComplete();
            }
          }, 1000);
        }
      }
    });
  }, []);

  // 断开 WebSocket 连接
  const disconnectWebSocket = useCallback(() => {
    if (wsClient.current) {
      wsClient.current.disconnect();
      wsClient.current = null;
    }
  }, []);

  // 清空日志
  const clearLogs = useCallback(() => {
    setInstallLogs([]);
    setDownloadProgress(0);
  }, []);

  // 检查是否有正在运行的任务
  const checkRunningTask = useCallback(async (serverId: number) => {
    return await frpServerApi.getRunningTask(serverId);
  }, []);

  // 远程安装
  const remoteInstall = useCallback(async (serverId: number, mirrorId?: number, onComplete?: () => void) => {
    connectWebSocket(serverId, '安装', onComplete);
    try {
      await frpServerApi.remoteInstall(serverId, mirrorId);
    } catch (error) {
      const err = error as { response?: { data?: { message?: string } } };
      setInstallLogs(prev => [...prev, {
        message: `错误: ${err.response?.data?.message || '安装失败'}`,
        type: 'error',
        timestamp: new Date().toLocaleTimeString('zh-CN', { hour12: false })
      }]);
    }
  }, [connectWebSocket]);

  // 远程重装
  const remoteReinstall = useCallback(async (
    serverId: number, 
    regenerateAuth: boolean, 
    mirrorId?: number,
    onComplete?: () => void
  ) => {
    connectWebSocket(serverId, '重装', onComplete);
    try {
      await frpServerApi.remoteReinstall(serverId, regenerateAuth, mirrorId);
    } catch (error) {
      const err = error as { response?: { data?: { message?: string } } };
      setInstallLogs(prev => [...prev, {
        message: `错误: ${err.response?.data?.message || '重装失败'}`,
        type: 'error',
        timestamp: new Date().toLocaleTimeString('zh-CN', { hour12: false })
      }]);
    }
  }, [connectWebSocket]);

  // 远程升级
  const remoteUpgrade = useCallback(async (
    serverId: number,
    version: string,
    mirrorId?: number,
    onComplete?: () => void
  ) => {
    connectWebSocket(serverId, '升级', onComplete);
    try {
      await frpServerApi.remoteUpgrade(serverId, version, mirrorId);
    } catch (error) {
      const err = error as { response?: { data?: { message?: string } } };
      setInstallLogs(prev => [...prev, {
        message: `错误: ${err.response?.data?.message || '升级失败'}`,
        type: 'error',
        timestamp: new Date().toLocaleTimeString('zh-CN', { hour12: false })
      }]);
    }
  }, [connectWebSocket]);

  // 远程启动
  const remoteStart = useCallback(async (serverId: number) => {
    await frpServerApi.remoteStart(serverId);
    toast.success('远程启动成功');
  }, []);

  // 远程停止
  const remoteStop = useCallback(async (serverId: number) => {
    await frpServerApi.remoteStop(serverId);
    toast.success('远程停止成功');
  }, []);

  // 远程重启
  const remoteRestart = useCallback(async (serverId: number) => {
    await frpServerApi.remoteRestart(serverId);
    toast.success('远程重启成功');
  }, []);

  // 查看远程日志
  const getRemoteLogs = useCallback(async (serverId: number, lines = 100) => {
    return await frpServerApi.remoteGetLogs(serverId, lines);
  }, []);

  // 获取日志颜色
  const getLogColor = useCallback((type: string) => {
    switch (type) {
      case 'error': return '#ff4d4f';
      case 'success': return '#52c41a';
      case 'warning': return '#faad14';
      case 'progress': return '#722ed1';
      case 'info': return '#1890ff';
      default: return '#1890ff';
    }
  }, []);

  return {
    installLogs,
    downloadProgress,
    logEndRef,
    connectWebSocket,
    disconnectWebSocket,
    clearLogs,
    checkRunningTask,
    remoteInstall,
    remoteReinstall,
    remoteUpgrade,
    remoteStart,
    remoteStop,
    remoteRestart,
    getRemoteLogs,
    getLogColor,
  };
}