import { useState, useCallback } from 'react';
import { toast } from '../../../components/ui/Toast';
import { clientApi } from '../../../api/client';
import { frpServerApi, type FrpServer } from '../../../api/frpServer';
import { githubMirrorApi, type GithubMirror } from '../../../api/githubMirror';
import type { Client } from '../../../types';

export interface UseClientsReturn {
  // 数据状态
  clients: Client[];
  total: number;
  loading: boolean;
  page: number;
  keyword: string;
  frpServers: FrpServer[];
  githubMirrors: GithubMirror[];
  selectedRowKeys: number[];
  
  // 操作方法
  setPage: (page: number) => void;
  setKeyword: (keyword: string) => void;
  setSelectedRowKeys: (keys: number[]) => void;
  fetchClients: () => Promise<void>;
  loadFrpServers: () => Promise<void>;
  loadGithubMirrors: () => Promise<void>;
  createClient: (values: Partial<Client>) => Promise<boolean>;
  updateClient: (id: number, values: Partial<Client>) => Promise<boolean>;
  deleteClient: (id: number) => Promise<boolean>;
  getConfig: (clientId: number) => Promise<string | null>;
  parseConfig: (configContent: string) => Promise<{
    server_addr?: string;
    server_port?: number;
    token?: string;
    frpc_admin_host?: string;
    frpc_admin_port?: number;
    frpc_admin_user?: string;
    frpc_admin_pwd?: string;
  } | null>;
}

export function useClients(): UseClientsReturn {
  const [clients, setClients] = useState<Client[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [frpServers, setFrpServers] = useState<FrpServer[]>([]);
  const [githubMirrors, setGithubMirrors] = useState<GithubMirror[]>([]);
  const [selectedRowKeys, setSelectedRowKeys] = useState<number[]>([]);

  const fetchClients = useCallback(async () => {
    setLoading(true);
    try {
      const res = await clientApi.getClients({ page, page_size: 10, keyword });
      setClients(res.list);
      setTotal(res.total);
    } catch {
      toast.error('获取客户端列表失败');
    } finally {
      setLoading(false);
    }
  }, [page, keyword]);

  const loadFrpServers = useCallback(async () => {
    try {
      const servers = await frpServerApi.getAll();
      setFrpServers(servers);
    } catch {
      toast.error('加载FRP服务器列表失败');
    }
  }, []);

  const loadGithubMirrors = useCallback(async () => {
    try {
      const mirrors = await githubMirrorApi.getAll();
      setGithubMirrors(mirrors.filter(m => m.enabled));
    } catch {
      // 静默处理，镜像源加载失败不影响主要功能
    }
  }, []);

  const createClient = useCallback(async (values: Partial<Client>): Promise<boolean> => {
    try {
      await clientApi.createClient(values);
      toast.success('创建成功');
      await fetchClients();
      return true;
    } catch {
      toast.error('创建失败');
      return false;
    }
  }, [fetchClients]);

  const updateClient = useCallback(async (id: number, values: Partial<Client>): Promise<boolean> => {
    try {
      await clientApi.updateClient(id, values);
      toast.success('更新成功');
      await fetchClients();
      return true;
    } catch {
      toast.error('更新失败');
      return false;
    }
  }, [fetchClients]);

  const deleteClient = useCallback(async (id: number): Promise<boolean> => {
    try {
      await clientApi.deleteClient(id);
      toast.success('删除成功');
      await fetchClients();
      return true;
    } catch {
      toast.error('删除失败');
      return false;
    }
  }, [fetchClients]);

  const getConfig = useCallback(async (clientId: number): Promise<string | null> => {
    try {
      const config = await clientApi.getConfig(clientId) as string;
      return config;
    } catch {
      toast.error('获取配置失败');
      return null;
    }
  }, []);

  const parseConfig = useCallback(async (configContent: string) => {
    try {
      const res = await clientApi.parseConfig(configContent) as {
        server_addr?: string;
        server_port?: number;
        token?: string;
        frpc_admin_host?: string;
        frpc_admin_port?: number;
        frpc_admin_user?: string;
        frpc_admin_pwd?: string;
      };
      toast.success('配置导入成功');
      return res;
    } catch (error) {
      const err = error as { response?: { data?: { message?: string } } };
      toast.error(err.response?.data?.message || '配置解析失败,请检查配置格式');
      return null;
    }
  }, []);

  return {
    clients,
    total,
    loading,
    page,
    keyword,
    frpServers,
    githubMirrors,
    selectedRowKeys,
    setPage,
    setKeyword,
    setSelectedRowKeys,
    fetchClients,
    loadFrpServers,
    loadGithubMirrors,
    createClient,
    updateClient,
    deleteClient,
    getConfig,
    parseConfig,
  };
}