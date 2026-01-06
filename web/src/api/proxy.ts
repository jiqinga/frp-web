import request from './request';
import type { Proxy } from '../types';

export const proxyApi = {
  // 获取所有代理列表
  getAllProxies: () => request.get<Proxy[]>('/proxies'),

  // 获取指定客户端的代理列表
  getProxiesByClient: (clientId: number) =>
    request.get<Proxy[]>(`/clients/${clientId}/proxies`),

  createProxy: (data: Partial<Proxy>) => request.post<Proxy>('/proxies', data),

  updateProxy: (id: number, data: Partial<Proxy>) =>
    request.put<Proxy>(`/proxies/${id}`, data),

  // 删除代理
  // deleteDNS: 是否同时删除关联的 DNS 记录，默认为 true
  deleteProxy: (id: number, deleteDNS: boolean = true) =>
    request.delete(`/proxies/${id}?deleteDNS=${deleteDNS}`),

  // 切换代理启用/禁用状态
  toggleProxy: (id: number) => request.put<Proxy>(`/proxies/${id}/toggle`),
};