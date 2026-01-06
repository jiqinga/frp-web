// 应用全局常量配置

// 分页配置
export const PAGINATION = {
  DEFAULT_PAGE_SIZE: 10,
  PAGE_SIZE_OPTIONS: [10, 20, 50, 100],
} as const;

// 实时监控配置
export const REALTIME_MONITOR = {
  MAX_CHART_HISTORY: 30,
  MAX_PROXY_HISTORY: 20,
  REFRESH_INTERVAL: 1000,
} as const;

// WebSocket 配置
export const WEBSOCKET = {
  RECONNECT_INTERVAL: 3000,
  MAX_RECONNECT_ATTEMPTS: 5,
  HEARTBEAT_INTERVAL: 30000,
} as const;

// 请求配置
export const REQUEST = {
  TIMEOUT: 10000,
  RETRY_COUNT: 3,
  RETRY_DELAY: 1000,
} as const;

// 表单验证
export const VALIDATION = {
  MIN_PASSWORD_LENGTH: 6,
  MAX_NAME_LENGTH: 50,
  MAX_REMARK_LENGTH: 200,
} as const;

// 流量单位
export const TRAFFIC_UNITS = {
  KB: 1024,
  MB: 1024 * 1024,
  GB: 1024 * 1024 * 1024,
} as const;

// 状态刷新间隔
export const REFRESH_INTERVALS = {
  CLIENT_STATUS: 5000,
  PROXY_STATUS: 3000,
  SERVER_STATUS: 10000,
} as const;

// 列表虚拟化阈值
export const VIRTUALIZATION = {
  THRESHOLD: 100,
  ITEM_HEIGHT: 48,
  OVERSCAN: 5,
} as const;