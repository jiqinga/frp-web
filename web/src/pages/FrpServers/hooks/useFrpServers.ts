import { useState, useEffect, useCallback } from 'react';
import { toast } from '../../../components/ui/Toast';
import { frpServerApi, type FrpServer } from '../../../api/frpServer';
import { githubMirrorApi, type GithubMirror } from '../../../api/githubMirror';

/**
 * FRP 服务器数据管理 Hook
 * 
 * 功能：
 * - 服务器列表加载
 * - GitHub 镜像源加载
 * - 服务器 CRUD 操作
 * - 连接测试
 * - 版本刷新
 */
export function useFrpServers() {
  const [servers, setServers] = useState<FrpServer[]>([]);
  const [mirrors, setMirrors] = useState<GithubMirror[]>([]);
  const [loading, setLoading] = useState(false);

  // 加载服务器列表
  const loadServers = useCallback(async () => {
    setLoading(true);
    try {
      const res = await frpServerApi.getAll();
      setServers(res || []);
    } catch {
      toast.error('加载服务器列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  // 加载 GitHub 镜像源
  const loadMirrors = useCallback(async () => {
    try {
      const res = await githubMirrorApi.getAll();
      setMirrors(res || []);
    } catch {
      // ignore
    }
  }, []);

  // 创建服务器
  const createServer = useCallback(async (values: Partial<FrpServer>) => {
    await frpServerApi.create(values as FrpServer);
    toast.success('创建成功');
    await loadServers();
  }, [loadServers]);

  // 更新服务器
  const updateServer = useCallback(async (id: number, values: Partial<FrpServer>) => {
    await frpServerApi.update(id, values as FrpServer);
    toast.success('更新成功');
    await loadServers();
  }, [loadServers]);

  // 删除服务器
  const deleteServer = useCallback(async (id: number, removeInstallation = false) => {
    await frpServerApi.delete(id, removeInstallation);
    toast.success('删除成功');
    await loadServers();
  }, [loadServers]);

  // 测试连接
  const testConnection = useCallback(async (server: FrpServer) => {
    await frpServerApi.testConnection(server);
    toast.success('连接测试成功');
    await loadServers();
  }, [loadServers]);

  // 测试 SSH 连接
  const testSSH = useCallback(async (id: number) => {
    await frpServerApi.testSSH(id);
  }, []);

  // 刷新本地版本
  const refreshLocalVersion = useCallback(async (id: number) => {
    const res = await frpServerApi.getLocalVersion(id);
    toast.success(`版本: ${res.version}`);
    await loadServers();
  }, [loadServers]);

  // 刷新远程版本
  const refreshRemoteVersion = useCallback(async (id: number) => {
    const res = await frpServerApi.remoteGetVersion(id);
    toast.success(`版本: ${res.version}`);
    await loadServers();
  }, [loadServers]);

  // 生成随机 Token
  const generateToken = useCallback(() => {
    const charset = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
    let token = '';
    for (let i = 0; i < 48; i++) {
      token += charset.charAt(Math.floor(Math.random() * charset.length));
    }
    return token;
  }, []);

  // 解析配置文件
  const parseConfig = useCallback(async (configContent: string) => {
    return await frpServerApi.parseConfig(configContent);
  }, []);

  // 初始化加载
  useEffect(() => {
    loadServers();
    loadMirrors();
  }, [loadServers, loadMirrors]);

  return {
    servers,
    mirrors,
    loading,
    loadServers,
    createServer,
    updateServer,
    deleteServer,
    testConnection,
    testSSH,
    refreshLocalVersion,
    refreshRemoteVersion,
    generateToken,
    parseConfig,
  };
}