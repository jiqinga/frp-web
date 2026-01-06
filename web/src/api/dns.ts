import request from './request';
import type {
  DNSProvider,
  DNSRecord,
  CreateDNSProviderRequest,
  UpdateDNSProviderRequest,
  TestDNSProviderRequest,
} from '../types';

export const dnsApi = {
  // ==================== DNS 提供商 ====================

  // 获取所有 DNS 提供商
  getProviders: () => request.get<DNSProvider[]>('/dns/providers'),

  // 获取单个 DNS 提供商
  getProvider: (id: number) => request.get<DNSProvider>(`/dns/providers/${id}`),

  // 创建 DNS 提供商
  createProvider: (data: CreateDNSProviderRequest) =>
    request.post<DNSProvider>('/dns/providers', data),

  // 更新 DNS 提供商
  updateProvider: (id: number, data: UpdateDNSProviderRequest) =>
    request.put<DNSProvider>(`/dns/providers/${id}`, data),

  // 删除 DNS 提供商
  deleteProvider: (id: number) => request.delete(`/dns/providers/${id}`),

  // 测试 DNS 提供商连接（未保存的配置）
  testProvider: (data: TestDNSProviderRequest) =>
    request.post<{ success: boolean; message: string }>('/dns/providers/test', data),

  // 测试已保存的 DNS 提供商连接（通过 ID）
  testProviderById: (id: number) =>
    request.post<{ message: string }>(`/dns/providers/${id}/test`),

  // 获取 DNS 提供商下托管的域名列表
  getProviderDomains: (id: number) => request.get<string[]>(`/dns/providers/${id}/domains`),

  // 获取 DNS 提供商密钥（用于编辑时显示）
  getProviderSecret: (id: number) =>
    request.get<{ secret_key: string }>(`/dns/providers/${id}/secret`),

  // ==================== DNS 记录 ====================

  // 获取所有 DNS 记录
  getRecords: () => request.get<DNSRecord[]>('/dns/records'),

  // 获取指定代理的 DNS 记录
  getRecordByProxy: (proxyId: number) =>
    request.get<DNSRecord>(`/dns/records/proxy/${proxyId}`),

  // 手动同步 DNS 记录
  syncRecord: (proxyId: number) =>
    request.post<DNSRecord>(`/dns/records/sync/${proxyId}`),

  // 删除 DNS 记录
  deleteRecord: (proxyId: number) =>
    request.delete(`/dns/records/proxy/${proxyId}`),
};