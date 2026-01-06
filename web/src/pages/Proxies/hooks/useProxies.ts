import { useState, useCallback, useMemo, useEffect } from 'react';
import { toast } from '../../../components/ui/Toast';
import { proxyApi } from '../../../api/proxy';
import { clientApi } from '../../../api/client';
import type { Proxy, Client } from '../../../types';

export interface UseProxiesReturn {
  // 数据状态
  proxies: Proxy[];
  clients: Client[];
  loading: boolean;
  selectedClient: number | undefined;
  onlineClientIds: Set<number>;
  
  // 操作方法
  setSelectedClient: (clientId: number | undefined) => void;
  fetchClients: () => Promise<void>;
  fetchProxies: (clientId?: number) => Promise<void>;
  createProxy: (values: Partial<Proxy>) => Promise<boolean>;
  updateProxy: (id: number, values: Partial<Proxy>) => Promise<boolean>;
  deleteProxy: (id: number, deleteDNS?: boolean) => Promise<boolean>;
  toggleProxy: (id: number) => Promise<boolean>;
  getClientById: (clientId: number) => Client | undefined;
  getAccessUrl: (proxy: Proxy) => { url: string; isClickable: boolean };
  getServerAddrForProxy: (proxy: Proxy) => string;
  isClientOnline: (clientId: number) => boolean;
}

export function useProxies(): UseProxiesReturn {
  const [proxies, setProxies] = useState<Proxy[]>([]);
  const [clients, setClients] = useState<Client[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedClient, setSelectedClient] = useState<number | undefined>();
  const [onlineClientIds, setOnlineClientIds] = useState<Set<number>>(new Set());

  // 获取在线客户端列表
  const fetchOnlineClients = useCallback(async () => {
    try {
      const res = await clientApi.getOnlineClients();
      setOnlineClientIds(new Set(res.online_client_ids));
    } catch {
      // 静默失败，不影响主流程
    }
  }, []);

  const fetchClients = useCallback(async () => {
    try {
      const res = await clientApi.getClients({ page: 1, page_size: 100 });
      setClients(res.list);
    } catch {
      toast.error('获取客户端列表失败');
    }
  }, []);

  const fetchProxies = useCallback(async (clientId?: number) => {
    setLoading(true);
    try {
      // 如果没有选择客户端，获取所有代理；否则获取指定客户端的代理
      const data = clientId
        ? await proxyApi.getProxiesByClient(clientId)
        : await proxyApi.getAllProxies();
      setProxies(data);
    } catch {
      toast.error('获取代理列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  const createProxy = useCallback(async (values: Partial<Proxy>): Promise<boolean> => {
    try {
      await proxyApi.createProxy(values);
      toast.success('创建成功');
      await fetchProxies(selectedClient);
      return true;
    } catch {
      toast.error('创建失败');
      return false;
    }
  }, [fetchProxies, selectedClient]);

  const updateProxy = useCallback(async (id: number, values: Partial<Proxy>): Promise<boolean> => {
    try {
      await proxyApi.updateProxy(id, values);
      toast.success('更新成功');
      await fetchProxies(selectedClient);
      return true;
    } catch {
      toast.error('更新失败');
      return false;
    }
  }, [fetchProxies, selectedClient]);

  const deleteProxy = useCallback(async (id: number, deleteDNS: boolean = true): Promise<boolean> => {
    try {
      await proxyApi.deleteProxy(id, deleteDNS);
      toast.success('删除成功');
      await fetchProxies(selectedClient);
      return true;
    } catch {
      toast.error('删除失败');
      return false;
    }
  }, [fetchProxies, selectedClient]);

  const toggleProxy = useCallback(async (id: number): Promise<boolean> => {
    try {
      const updatedProxy = await proxyApi.toggleProxy(id);
      toast.success(updatedProxy.enabled ? '代理已启用' : '代理已禁用');
      await fetchProxies(selectedClient);
      return true;
    } catch {
      toast.error('切换状态失败');
      return false;
    }
  }, [fetchProxies, selectedClient]);

  // 初始化加载客户端列表和在线状态
  useEffect(() => {
    fetchClients();
    fetchOnlineClients();
  }, [fetchClients, fetchOnlineClients]);

  // 定期刷新在线状态（每30秒）
  useEffect(() => {
    const interval = setInterval(fetchOnlineClients, 30000);
    return () => clearInterval(interval);
  }, [fetchOnlineClients]);

  // 当 selectedClient 变化时重新获取代理列表
  useEffect(() => {
    fetchProxies(selectedClient);
  }, [fetchProxies, selectedClient]);

  // 根据 client_id 获取客户端信息
  const getClientById = useCallback((clientId: number) => {
    return clients.find(c => c.id === clientId);
  }, [clients]);

  // 获取当前选中客户端
  const currentClient = useMemo(() => {
    return clients.find(c => c.id === selectedClient);
  }, [clients, selectedClient]);

  // 根据代理类型生成访问地址
  const getAccessUrl = useCallback((proxy: Proxy): { url: string; isClickable: boolean } => {
    // 如果选择了特定客户端，使用当前客户端的 server_addr
    // 否则根据代理的 client_id 查找对应客户端的 server_addr
    const client = selectedClient ? currentClient : getClientById(proxy.client_id);
    const serverAddr = client?.server_addr || '';
    
    // 检查是否是静态文件插件 - 静态文件插件使用 HTTP 协议访问
    if (proxy.plugin_type === 'static_file') {
      if (proxy.remote_port) {
        try {
          const config = JSON.parse(proxy.plugin_config || '{}');
          const stripPrefix = config.stripPrefix || '';
          const path = stripPrefix ? `/${stripPrefix}/` : '/';
          return { url: `http://${serverAddr}:${proxy.remote_port}${path}`, isClickable: true };
        } catch {
          return { url: `http://${serverAddr}:${proxy.remote_port}/`, isClickable: true };
        }
      }
      return { url: '端口未分配', isClickable: false };
    }
    
    switch (proxy.type?.toLowerCase()) {
      case 'tcp':
      case 'udp':
        if (proxy.remote_port) {
          return { url: `${serverAddr}:${proxy.remote_port}`, isClickable: false };
        }
        return { url: '端口未分配', isClickable: false };
      case 'http':
        if (proxy.custom_domains) {
          return { url: `http://${proxy.custom_domains}`, isClickable: true };
        }
        return { url: '未配置域名', isClickable: false };
      case 'https':
        if (proxy.custom_domains) {
          return { url: `https://${proxy.custom_domains}`, isClickable: true };
        }
        return { url: '未配置域名', isClickable: false };
      case 'stcp':
        return { url: 'STCP (无外部地址)', isClickable: false };
      default:
        return { url: '-', isClickable: false };
    }
  }, [selectedClient, currentClient, getClientById]);

  // 获取代理对应的服务器地址
  const getServerAddrForProxy = useCallback((proxy: Proxy): string => {
    const client = getClientById(proxy.client_id);
    return client?.server_addr || '';
  }, [getClientById]);

  // 判断客户端是否在线
  const isClientOnline = useCallback((clientId: number): boolean => {
    return onlineClientIds.has(clientId);
  }, [onlineClientIds]);

  return {
    proxies,
    clients,
    loading,
    selectedClient,
    onlineClientIds,
    setSelectedClient,
    fetchClients,
    fetchProxies,
    createProxy,
    updateProxy,
    deleteProxy,
    toggleProxy,
    getClientById,
    getAccessUrl,
    getServerAddrForProxy,
    isClientOnline,
  };
}