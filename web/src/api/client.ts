import request from './request';
import type { Client, PaginationResponse } from '../types';

// 更新请求类型
export interface UpdateClientRequest {
  update_type: 'frpc' | 'daemon';
  version?: string;
  mirror_id?: number;
}

// 批量更新请求类型
export interface BatchUpdateRequest {
  client_ids: number[];
  update_type: 'frpc' | 'daemon';
  version?: string;
  mirror_id?: number;
}

// 批量更新响应类型
export interface BatchUpdateResponse {
  message: string;
  success_count: number;
  failed_clients: string[];
  total: number;
}

// 客户端版本信息
export interface ClientVersions {
  frpc_version: string;
  daemon_version: string;
  os: string;
  arch: string;
}

// 在线客户端响应
export interface OnlineClientsResponse {
  online_client_ids: number[];
  count: number;
}

// 日志流请求类型
export interface LogStreamRequest {
  log_type: 'frpc' | 'daemon';
  lines: number;
}

// frpc控制请求类型
export interface FrpcControlRequest {
  action: 'start' | 'stop' | 'restart';
}

// frpc控制响应类型
export interface FrpcControlResponse {
  message: string;
  success: boolean;
}

export const clientApi = {
  getClients: (params: { page: number; page_size: number; keyword?: string }) =>
    request.get<PaginationResponse<Client>>('/clients', { params }),

  getClient: (id: number) => request.get<Client>(`/clients/${id}`),

  createClient: (data: Partial<Client>) => request.post<Client>('/clients', data),

  updateClient: (id: number, data: Partial<Client>) =>
    request.put<Client>(`/clients/${id}`, data),

  deleteClient: (id: number) => request.delete(`/clients/${id}`),

  exportConfig: (id: number) => request.get(`/clients/${id}/export`, { responseType: 'blob' }),

  // 获取客户端配置内容（文本格式）
  getConfig: (id: number) => request.get<string>(`/clients/${id}/export`, { responseType: 'text' }),

  generateRegisterToken: (data: {
    client_name: string;
    frp_server_id: number;
    server_addr: string;
    server_port: number;
    token_str?: string;
    protocol?: string;
    remark?: string;
  }) => request.post('/clients/register/token', data),

  generateRegisterScript: (params: { token: string; type: string; mirror: string }) =>
    request.get('/clients/register/script', { params, responseType: 'text' }),

  parseConfig: (config: string) =>
    request.post('/clients/parse-config', { config }),

  // 更新客户端软件（frpc 或 daemon）
  updateClientSoftware: (id: number, data: UpdateClientRequest) =>
    request.post<{ message: string }>(`/clients/${id}/update`, data),

  // 批量更新客户端软件
  batchUpdateClients: (data: BatchUpdateRequest) =>
    request.post<BatchUpdateResponse>('/clients/batch-update', data),

  // 获取客户端版本信息
  getClientVersions: (id: number) =>
    request.get<ClientVersions>(`/clients/${id}/versions`),

  // 获取在线客户端列表
  getOnlineClients: () =>
    request.get<OnlineClientsResponse>('/clients/online'),

  // 开始日志流
  startLogStream: (id: number, data: LogStreamRequest) =>
    request.post<{ message: string }>(`/clients/${id}/logs/start`, data),

  // 停止日志流
  stopLogStream: (id: number, logType: 'frpc' | 'daemon') =>
    request.post<{ message: string }>(`/clients/${id}/logs/stop`, { log_type: logType }),

  // 控制frpc（启动/停止/重启）
  controlFrpc: (id: number, data: FrpcControlRequest) =>
    request.post<FrpcControlResponse>(`/clients/${id}/frpc/control`, data),
};