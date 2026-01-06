import request from './request';
import type { PaginationResponse } from '../types';

export interface OperationLog {
  id: number;
  user_id: number;
  username: string;
  operation_type: string;
  resource_type: string;
  resource_id: number;
  description: string;
  ip_address: string;
  ip_location: string;
  created_at: string;
}

export const logApi = {
  getLogs: (params: { page: number; page_size: number; operation_type?: string; resource_type?: string }) =>
    request.get<PaginationResponse<OperationLog>>('/logs', { params }),

  createLog: (data: Partial<OperationLog>) => request.post('/logs', data)
};