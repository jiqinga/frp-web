import { useState, useEffect, useCallback } from 'react';
import { Bell, Plus, AlertTriangle } from 'lucide-react';
import type { AlertRule } from '../../api/alert';
import { Button, Card, CardHeader, CardContent, Badge, Tabs } from '../../components/ui';
import { ConfirmModal } from '../../components/ui/ConfirmModal';
import { useAlerts } from './hooks';
import { AlertStatsCard, AlertRuleTable, AlertRuleFormModal, AlertLogsPanel, AlertTrendChart } from './components';

function AlertRules() {
  const {
    rules, logs, clients, servers, loading, logsLoading, stats,
    fetchRules, fetchLogs, loadClients, loadServers,
    createRule, updateRule, deleteRule, toggleRule,
  } = useAlerts();

  const [modalVisible, setModalVisible] = useState(false);
  const [editingRule, setEditingRule] = useState<AlertRule | null>(null);
  const [deleteConfirmVisible, setDeleteConfirmVisible] = useState(false);
  const [deletingRuleId, setDeletingRuleId] = useState<number | null>(null);
  const [togglingIds, setTogglingIds] = useState<Set<number>>(new Set());

  useEffect(() => {
    fetchRules();
    fetchLogs();
    loadClients();
    loadServers();
  }, [fetchRules, fetchLogs, loadClients, loadServers]);

  const handleAdd = () => {
    setEditingRule(null);
    setModalVisible(true);
  };

  const handleEdit = (rule: AlertRule) => {
    setEditingRule(rule);
    setModalVisible(true);
  };

  const handleDelete = (id: number) => {
    setDeletingRuleId(id);
    setDeleteConfirmVisible(true);
  };

  const confirmDelete = async () => {
    if (deletingRuleId !== null) {
      const success = await deleteRule(deletingRuleId);
      if (success) fetchRules();
    }
    setDeleteConfirmVisible(false);
    setDeletingRuleId(null);
  };

  const handleToggle = async (rule: AlertRule) => {
    const id = rule.id!;
    setTogglingIds(prev => new Set(prev).add(id));
    try {
      const success = await toggleRule(rule);
      if (success) fetchRules();
    } finally {
      setTogglingIds(prev => {
        const next = new Set(prev);
        next.delete(id);
        return next;
      });
    }
  };

  const isToggling = useCallback((id: number) => togglingIds.has(id), [togglingIds]);

  const handleSubmit = async (rule: AlertRule) => {
    const success = editingRule ? await updateRule(rule) : await createRule(rule);
    if (success) {
      setModalVisible(false);
      fetchRules();
    }
  };

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">告警管理</h1>
          <p className="mt-1 text-foreground-muted">配置流量告警和离线告警规则</p>
        </div>
        <Button onClick={handleAdd} icon={<Plus className="h-4 w-4" />}>
          新增规则
        </Button>
      </div>

      <AlertStatsCard stats={stats} />

      <AlertTrendChart logs={logs} />

      <Tabs
        items={[
          {
            key: 'rules',
            label: '规则管理',
            children: (
              <Card>
                <CardHeader>
                  <div className="flex items-center gap-2">
                    <Bell className="h-5 w-5 text-yellow-400" />
                    <span>告警规则</span>
                    <Badge variant="default">{rules.length} 条</Badge>
                  </div>
                </CardHeader>
                <CardContent className="p-0">
                  <AlertRuleTable
                    rules={rules}
                    clients={clients}
                    servers={servers}
                    loading={loading}
                    onEdit={handleEdit}
                    onDelete={handleDelete}
                    onToggle={handleToggle}
                    isToggling={isToggling}
                  />
                </CardContent>
              </Card>
            ),
          },
          {
            key: 'logs',
            label: '告警历史',
            children: <AlertLogsPanel logs={logs} loading={logsLoading} />,
          },
        ]}
        defaultActiveKey="rules"
      />

      <AlertRuleFormModal
        visible={modalVisible}
        editingRule={editingRule}
        clients={clients}
        servers={servers}
        onCancel={() => setModalVisible(false)}
        onSubmit={handleSubmit}
      />

      <ConfirmModal
        open={deleteConfirmVisible}
        onClose={() => setDeleteConfirmVisible(false)}
        onConfirm={confirmDelete}
        title="删除告警规则"
        content="确定删除此规则吗？删除后无法恢复。"
        type="warning"
        confirmText="删除"
        cancelText="取消"
      />

      <div className="flex items-start gap-3 rounded-lg border p-4 border-yellow-500/30 bg-yellow-500/10 dark:border-yellow-500/30 dark:bg-yellow-500/10">
        <AlertTriangle className="h-5 w-5 flex-shrink-0 text-yellow-500 dark:text-yellow-400" />
        <div className="text-sm">
          <p className="font-medium text-yellow-700 dark:text-yellow-200">告警规则说明</p>
          <ul className="mt-2 list-inside list-disc space-y-1 text-yellow-600 dark:text-yellow-200/80">
            <li>代理流量：监控代理的每日/每月流量或实时速率</li>
            <li>frpc离线：当客户端离线时触发告警</li>
            <li>frps离线：当FRP服务器状态异常时触发告警</li>
            <li>冷却时间：同一目标在冷却期内不会重复告警</li>
          </ul>
        </div>
      </div>
    </div>
  );
}

export function Component() {
  return <AlertRules />;
}

export default AlertRules;