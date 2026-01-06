export class WebSocketClient {
  private ws: WebSocket | null = null;
  private url: string = '';
  private token: string = '';
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  private messageHandlers: Map<string, ((data: unknown) => void)[]> = new Map();
  private isDisconnecting: boolean = false;
  private isConnecting: boolean = false;

  connect(url: string, token: string) {
    // 防止重复连接
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
      return;
    }
    
    this.url = url;
    this.token = token;
    this.isDisconnecting = false;
    this.isConnecting = true;
    
    const wsUrl = `${url}?token=${token}`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      this.isConnecting = false;
      if (!this.isDisconnecting) {
        this.startHeartbeat();
      }
    };

    this.ws.onmessage = (event) => {
      if (this.isDisconnecting) return;
      try {
        const data = JSON.parse(event.data);
        const handlers = this.messageHandlers.get(data.type) || [];
        handlers.forEach(handler => handler(data));
      } catch {
        // ignore parse errors
      }
    };

    this.ws.onclose = () => {
      this.isConnecting = false;
      this.stopHeartbeat();
      if (!this.isDisconnecting) {
        this.reconnect();
      }
    };

    this.ws.onerror = () => {
      this.isConnecting = false;
      // error handled by onclose
    };
  }

  private startHeartbeat() {
    this.heartbeatTimer = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: 'ping' }));
      }
    }, 30000);
  }

  private stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  private reconnect() {
    if (this.reconnectTimer) return;
    
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect(this.url, this.token);
    }, 5000);
  }

  onMessage(type: string, handler: (data: unknown) => void) {
    if (!this.messageHandlers.has(type)) {
      this.messageHandlers.set(type, []);
    }
    this.messageHandlers.get(type)!.push(handler);
  }

  disconnect() {
    this.isDisconnecting = true;
    this.isConnecting = false;
    this.stopHeartbeat();
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      // 只在连接已建立时关闭
      if (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING) {
        this.ws.close();
      }
      this.ws = null;
    }
    this.messageHandlers.clear();
  }
}

export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}