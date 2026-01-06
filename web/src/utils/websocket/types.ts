/** WebSocket 连接状态 */
export type ConnectionState = 'disconnected' | 'connecting' | 'connected' | 'reconnecting';

/** WebSocket 消息基础结构 */
export interface WebSocketMessage<T = unknown> {
  type: string;
  data?: T;
  timestamp?: string;
}

/** 消息处理器 */
export type MessageHandler<T = unknown> = (message: WebSocketMessage<T>) => void;

/** WebSocket 管理器配置 */
export interface WebSocketConfig {
  /** 重连间隔（毫秒） */
  reconnectInterval?: number;
  /** 最大重连次数 */
  maxReconnectAttempts?: number;
  /** 心跳间隔（毫秒） */
  heartbeatInterval?: number;
}

/** 连接状态变化回调 */
export type StateChangeCallback = (state: ConnectionState) => void;