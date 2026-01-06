import { WEBSOCKET } from '../../constants/app';
import type {
  ConnectionState,
  WebSocketConfig,
  WebSocketMessage,
  MessageHandler,
  StateChangeCallback,
} from './types';

const DEFAULT_CONFIG: Required<WebSocketConfig> = {
  reconnectInterval: WEBSOCKET.RECONNECT_INTERVAL,
  maxReconnectAttempts: WEBSOCKET.MAX_RECONNECT_ATTEMPTS,
  heartbeatInterval: WEBSOCKET.HEARTBEAT_INTERVAL,
};

/**
 * 全局 WebSocket 管理器（单例模式）
 */
class WebSocketManager {
  private static instance: WebSocketManager | null = null;
  
  private ws: WebSocket | null = null;
  private config: Required<WebSocketConfig>;
  private state: ConnectionState = 'disconnected';
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  
  private url = '';
  private token = '';
  
  private messageHandlers = new Map<string, Set<MessageHandler>>();
  private stateCallbacks = new Set<StateChangeCallback>();

  private constructor(config?: WebSocketConfig) {
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  static getInstance(config?: WebSocketConfig): WebSocketManager {
    if (!WebSocketManager.instance) {
      WebSocketManager.instance = new WebSocketManager(config);
    }
    return WebSocketManager.instance;
  }

  /** 获取当前连接状态 */
  getState(): ConnectionState {
    return this.state;
  }

  /** 连接 WebSocket */
  connect(url: string, token: string): void {
    if (this.state === 'connected' || this.state === 'connecting') {
      return;
    }

    this.url = url;
    this.token = token;
    this.setState('connecting');
    this.createConnection();
  }

  /** 断开连接 */
  disconnect(): void {
    this.clearTimers();
    this.reconnectAttempts = 0;
    
    if (this.ws) {
      this.ws.onclose = null;
      if (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING) {
        this.ws.close();
      }
      this.ws = null;
    }
    
    this.setState('disconnected');
  }

  /** 订阅消息 */
  subscribe<T = unknown>(type: string, handler: MessageHandler<T>): () => void {
    if (!this.messageHandlers.has(type)) {
      this.messageHandlers.set(type, new Set());
    }
    this.messageHandlers.get(type)!.add(handler as MessageHandler);
    
    return () => {
      this.messageHandlers.get(type)?.delete(handler as MessageHandler);
    };
  }

  /** 监听状态变化 */
  onStateChange(callback: StateChangeCallback): () => void {
    this.stateCallbacks.add(callback);
    return () => {
      this.stateCallbacks.delete(callback);
    };
  }

  /** 发送消息 */
  send<T>(message: WebSocketMessage<T>): boolean {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
      return true;
    }
    return false;
  }

  private createConnection(): void {
    const wsUrl = `${this.url}?token=${this.token}`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
      this.setState('connected');
      this.startHeartbeat();
    };

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data) as WebSocketMessage;
        this.messageHandlers.get(message.type)?.forEach((handler) => handler(message));
      } catch {
        // 忽略解析错误
      }
    };

    this.ws.onclose = () => {
      this.stopHeartbeat();
      if (this.state !== 'disconnected') {
        this.attemptReconnect();
      }
    };

    this.ws.onerror = () => {
      // 错误由 onclose 处理
    };
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.config.maxReconnectAttempts) {
      this.setState('disconnected');
      return;
    }

    this.setState('reconnecting');
    this.reconnectAttempts++;
    
    this.reconnectTimer = setTimeout(() => {
      this.createConnection();
    }, this.config.reconnectInterval);
  }

  private startHeartbeat(): void {
    this.heartbeatTimer = setInterval(() => {
      this.send({ type: 'ping' });
    }, this.config.heartbeatInterval);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  private clearTimers(): void {
    this.stopHeartbeat();
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  private setState(state: ConnectionState): void {
    if (this.state !== state) {
      this.state = state;
      this.stateCallbacks.forEach((cb) => cb(state));
    }
  }
}

export { WebSocketManager };
export const wsManager = WebSocketManager.getInstance();