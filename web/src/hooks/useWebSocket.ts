import { useState, useEffect, useCallback } from 'react';
import { wsManager } from '../utils/websocket/WebSocketManager';
import type { ConnectionState, WebSocketMessage, MessageHandler } from '../utils/websocket/types';

interface UseWebSocketOptions {
  /** 是否自动连接 */
  autoConnect?: boolean;
}

interface UseWebSocketReturn {
  /** 当前连接状态 */
  state: ConnectionState;
  /** 是否已连接 */
  isConnected: boolean;
  /** 连接 WebSocket */
  connect: (url: string, token: string) => void;
  /** 断开连接 */
  disconnect: () => void;
  /** 发送消息 */
  send: <T>(message: WebSocketMessage<T>) => boolean;
  /** 订阅消息 */
  subscribe: <T = unknown>(type: string, handler: MessageHandler<T>) => () => void;
}

/**
 * WebSocket React Hook
 * 提供对全局 WebSocket 管理器的 React 封装
 */
export function useWebSocket(options: UseWebSocketOptions = {}): UseWebSocketReturn {
  const { autoConnect = false } = options;
  const [state, setState] = useState<ConnectionState>(wsManager.getState());

  useEffect(() => {
    return wsManager.onStateChange(setState);
  }, []);

  const connect = useCallback((url: string, token: string) => {
    wsManager.connect(url, token);
  }, []);

  const disconnect = useCallback(() => {
    wsManager.disconnect();
  }, []);

  const send = useCallback(<T,>(message: WebSocketMessage<T>) => {
    return wsManager.send(message);
  }, []);

  const subscribe = useCallback(<T = unknown,>(type: string, handler: MessageHandler<T>) => {
    return wsManager.subscribe(type, handler);
  }, []);

  // 自动连接（如果启用）
  useEffect(() => {
    if (autoConnect) {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${protocol}//${window.location.host}/api/ws/realtime`;
      const token = localStorage.getItem('token') || '';
      if (token) {
        connect(wsUrl, token);
      }
    }
  }, [autoConnect, connect]);

  return {
    state,
    isConnected: state === 'connected',
    connect,
    disconnect,
    send,
    subscribe,
  };
}

/**
 * 订阅特定类型消息的 Hook
 */
export function useWebSocketMessage<T = unknown>(
  type: string,
  handler: MessageHandler<T>,
  deps: React.DependencyList = []
): void {
  useEffect(() => {
    return wsManager.subscribe(type, handler);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [type, ...deps]);
}