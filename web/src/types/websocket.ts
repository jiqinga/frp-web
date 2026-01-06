export interface SSHLogMessage {
  type: 'ssh_log';
  server_id: number;
  operation: string;
  log: string;
  log_type: 'info' | 'progress' | 'error' | 'success';
  progress?: number;
  timestamp: string;
}

export interface EnhancedLogEntry {
  message: string;
  type: 'info' | 'progress' | 'error' | 'success';
  progress?: number;
  timestamp: string;
}

// 客户端更新进度消息
export interface ClientUpdateProgressMessage {
  type: 'client_update_progress';
  timestamp: string;
  data: {
    client_id: number;
    update_type: 'frpc' | 'daemon';
    stage: 'downloading' | 'stopping' | 'replacing' | 'starting' | 'completed' | 'failed';
    progress: number;
    message: string;
    total_bytes: number;
    downloaded_bytes: number;
  };
}

// 客户端更新结果消息
export interface ClientUpdateResultMessage {
  type: 'client_update_result';
  timestamp: string;
  data: {
    client_id: number;
    update_type: 'frpc' | 'daemon';
    success: boolean;
    version: string;
    message: string;
  };
}

// 更新阶段标签
export const UPDATE_STAGE_LABELS: Record<string, string> = {
  downloading: '下载中',
  stopping: '停止旧进程',
  replacing: '替换文件',
  starting: '启动新进程',
  completed: '完成',
  failed: '失败',
};