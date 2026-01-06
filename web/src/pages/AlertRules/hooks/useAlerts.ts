import { useState, useCallback, useEffect } from 'react';
import { toast } from '../../../components/ui/Toast';
import { alertApi, type AlertRule, type AlertLog } from '../../../api/alert';
import { clientApi } from '../../../api/client';
import { frpServerApi, type FrpServer } from '../../../api/frpServer';
import type { Client } from '../../../types';

/** 分批处理，限制并发数 */
async function batchProcess<T, R>(
  items: T[],
  processor: (item: T) => Promise<R>,
  batchSize = 3
): Promise<R[]> {
  const results: R[] = [];
  for (let i = 0; i < items.length; i += batchSize) {
    const batch = items.slice(i, i + batchSize);
    const batchResults = await Promise.all(batch.map(processor));
    results.push(...batchResults);
  }
  return results;
}

export interface AlertStats {
  totalRules: number;
  enabledRules: number;
  todayAlerts: number;
  weekAlerts: number;
}

export interface UseAlertsReturn {
  rules: AlertRule[];
  logs: AlertLog[];
  clients: Client[];
  servers: FrpServer[];
  loading: boolean;
  logsLoading: boolean;
  stats: AlertStats;
  selectedRowKeys: number[];
  setSelectedRowKeys: (keys: number[]) => void;
  fetchRules: () => Promise<void>;
  fetchLogs: (limit?: number) => Promise<void>;
  loadClients: () => Promise<void>;
  loadServers: () => Promise<void>;
  createRule: (rule: AlertRule) => Promise<boolean>;
  updateRule: (rule: AlertRule) => Promise<boolean>;
  deleteRule: (id: number) => Promise<boolean>;
  toggleRule: (rule: AlertRule) => Promise<boolean>;
  batchToggleRules: (ids: number[], enabled: boolean) => Promise<boolean>;
}

export function useAlerts(): UseAlertsReturn {
  const [rules, setRules] = useState<AlertRule[]>([]);
  const [logs, setLogs] = useState<AlertLog[]>([]);
  const [clients, setClients] = useState<Client[]>([]);
  const [servers, setServers] = useState<FrpServer[]>([]);
  const [loading, setLoading] = useState(false);
  const [logsLoading, setLogsLoading] = useState(false);
  const [selectedRowKeys, setSelectedRowKeys] = useState<number[]>([]);

  const [stats, setStats] = useState<AlertStats>({ totalRules: 0, enabledRules: 0, todayAlerts: 0, weekAlerts: 0 });

  const fetchRules = useCallback(async () => {
    setLoading(true);
    try {
      const res = await alertApi.getAllRules();
      const rulesList = (res as AlertRule[]) || [];
      setRules(rulesList);
    } catch {
      toast.error('获取告警规则失败');
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchLogs = useCallback(async (limit = 100) => {
    setLogsLoading(true);
    try {
      const res = await alertApi.getAlertLogs(limit);
      const logsList = (res as AlertLog[]) || [];
      setLogs(logsList);
    } catch {
      toast.error('获取告警日志失败');
    } finally {
      setLogsLoading(false);
    }
  }, []);

  // 计算统计数据 - 当 rules 或 logs 变化时更新
  useEffect(() => {
    const now = new Date();
    const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
    const weekStart = todayStart - 6 * 24 * 60 * 60 * 1000;
    setStats({
      totalRules: rules.length,
      enabledRules: rules.filter(r => r.enabled).length,
      todayAlerts: logs.filter(l => new Date(l.created_at).getTime() >= todayStart).length,
      weekAlerts: logs.filter(l => new Date(l.created_at).getTime() >= weekStart).length,
    });
  }, [rules, logs]);

  const loadClients = useCallback(async () => {
    try {
      const res = await clientApi.getClients({ page: 1, page_size: 1000 });
      setClients(res.list || []);
    } catch { /* ignore */ }
  }, []);

  const loadServers = useCallback(async () => {
    try {
      const res = await frpServerApi.getAll();
      setServers(res || []);
    } catch { /* ignore */ }
  }, []);

  const createRule = useCallback(async (rule: AlertRule): Promise<boolean> => {
    try {
      await alertApi.createRule(rule);
      toast.success('创建成功');
      await fetchRules();
      return true;
    } catch {
      toast.error('创建失败');
      return false;
    }
  }, [fetchRules]);

  const updateRule = useCallback(async (rule: AlertRule): Promise<boolean> => {
    try {
      await alertApi.updateRule(rule);
      toast.success('更新成功');
      await fetchRules();
      return true;
    } catch {
      toast.error('更新失败');
      return false;
    }
  }, [fetchRules]);

  const deleteRule = useCallback(async (id: number): Promise<boolean> => {
    try {
      await alertApi.deleteRule(id);
      toast.success('删除成功');
      await fetchRules();
      return true;
    } catch {
      toast.error('删除失败');
      return false;
    }
  }, [fetchRules]);

  const toggleRule = useCallback(async (rule: AlertRule): Promise<boolean> => {
    try {
      await alertApi.updateRule({ ...rule, enabled: !rule.enabled });
      toast.success(rule.enabled ? '已禁用' : '已启用');
      await fetchRules();
      return true;
    } catch {
      toast.error('操作失败');
      return false;
    }
  }, [fetchRules]);

  const batchToggleRules = useCallback(async (ids: number[], enabled: boolean): Promise<boolean> => {
    try {
      const targetRules = rules.filter(r => ids.includes(r.id!));
      await batchProcess(targetRules, r => alertApi.updateRule({ ...r, enabled }), 3);
      toast.success(`已${enabled ? '启用' : '禁用'} ${ids.length} 条规则`);
      await fetchRules();
      return true;
    } catch {
      toast.error('批量操作失败');
      return false;
    }
  }, [rules, fetchRules]);

  return {
    rules,
    logs,
    clients,
    servers,
    loading,
    logsLoading,
    stats,
    selectedRowKeys,
    setSelectedRowKeys,
    fetchRules,
    fetchLogs,
    loadClients,
    loadServers,
    createRule,
    updateRule,
    deleteRule,
    toggleRule,
    batchToggleRules,
  };
}