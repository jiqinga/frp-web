import { Edit, Trash2, Monitor, Server, Activity, Shield } from 'lucide-react';
import type { AlertRule, AlertTargetType } from '../../../api/alert';
import { systemRuleTypeNames, type SystemRuleType } from '../../../api/alert';
import type { Client } from '../../../types';
import type { FrpServer } from '../../../api/frpServer';
import { Button, Table, Badge, Switch } from '../../../components/ui';

interface AlertRuleTableProps {
  rules: AlertRule[];
  clients: Client[];
  servers: FrpServer[];
  loading: boolean;
  onEdit: (rule: AlertRule) => void;
  onDelete: (id: number) => void;
  onToggle: (rule: AlertRule) => void;
  isToggling: (id: number) => boolean;
}

const convertFromBytes = (bytes: number): { value: number; unit: string } => {
  if (bytes >= 1024 * 1024 * 1024 * 1024) return { value: bytes / (1024 * 1024 * 1024 * 1024), unit: 'TB' };
  if (bytes >= 1024 * 1024 * 1024) return { value: bytes / (1024 * 1024 * 1024), unit: 'GB' };
  if (bytes >= 1024 * 1024) return { value: bytes / (1024 * 1024), unit: 'MB' };
  return { value: bytes, unit: 'bytes' };
};

const getTargetTypeLabel = (type: AlertTargetType) => {
  const labels: Record<string, string> = { proxy: '代理流量', frpc: 'frpc离线', frps: 'frps离线', system: '系统告警' };
  return labels[type] || type;
};

const getRuleTypeLabel = (type: string) => {
  const labels: Record<string, string> = { daily: '每日', monthly: '每月', rate: '实时速率', offline: '离线' };
  // 检查是否为系统规则类型
  if (type in systemRuleTypeNames) {
    return systemRuleTypeNames[type as SystemRuleType];
  }
  return labels[type] || type;
};

export function AlertRuleTable({
  rules, clients, servers, loading,
  onEdit, onDelete, onToggle, isToggling,
}: AlertRuleTableProps) {
  const getTargetName = (rule: AlertRule) => {
    const targetType = rule.target_type || 'proxy';
    const targetId = rule.target_id || rule.proxy_id;
    if (targetType === 'system') {
      return '全局';
    }
    if (targetType === 'frpc') {
      const client = clients.find(c => c.id === targetId);
      return client?.name || `客户端#${targetId}`;
    }
    if (targetType === 'frps') {
      const server = servers.find(s => s.id === targetId);
      return server?.name || `服务器#${targetId}`;
    }
    return `代理#${targetId}`;
  };

  const columns = [
    {
      key: 'target',
      title: '告警目标',
      render: (_: unknown, record: AlertRule) => {
        const targetType = record.target_type || 'proxy';
        const icon = targetType === 'system' ? <Shield className="h-4 w-4" /> :
                     targetType === 'frpc' ? <Monitor className="h-4 w-4" /> :
                     targetType === 'frps' ? <Server className="h-4 w-4" /> :
                     <Activity className="h-4 w-4" />;
        return (
          <div className="flex items-center justify-center gap-2">
            {icon}
            <span className="text-foreground">{getTargetName(record)}</span>
            <Badge variant="info">{getTargetTypeLabel(targetType)}</Badge>
          </div>
        );
      }
    },
    {
      key: 'rule_type',
      title: '规则类型',
      render: (_: unknown, record: AlertRule) => (
        <Badge variant="primary">{getRuleTypeLabel(record.rule_type)}</Badge>
      )
    },
    {
      key: 'threshold_value',
      title: '阈值/冷却',
      render: (_: unknown, record: AlertRule) => {
        if (record.rule_type === 'offline') {
          return <span className="text-foreground-muted">冷却 {record.cooldown_minutes || 60} 分钟</span>;
        }
        const { value, unit } = convertFromBytes(record.threshold_value);
        return <span className="text-foreground">{value.toFixed(1)} {unit}</span>;
      }
    },
    {
      key: 'notify',
      title: '通知方式',
      render: (_: unknown, record: AlertRule) => {
        const hasRecipients = record.notify_recipient_ids && record.notify_recipient_ids.length > 0;
        const hasGroups = record.notify_group_ids && record.notify_group_ids.length > 0;
        return (
          <span className="text-sm text-foreground-muted">
            {hasRecipients || hasGroups ? '邮件' : '-'}
          </span>
        );
      }
    },
    {
      key: 'enabled',
      title: '状态',
      render: (_: unknown, record: AlertRule) => (
        <div className="flex justify-center">
          <Switch
            checked={record.enabled}
            onChange={() => onToggle(record)}
            loading={isToggling(record.id!)}
          />
        </div>
      )
    },
    {
      key: 'action',
      title: '操作',
      render: (_: unknown, record: AlertRule) => (
        <div className="flex items-center justify-center gap-1">
          <Button size="sm" variant="ghost" onClick={() => onEdit(record)}>
            <Edit className="h-4 w-4" />
          </Button>
          <Button size="sm" variant="ghost" onClick={() => onDelete(record.id!)}>
            <Trash2 className="h-4 w-4 text-red-400" />
          </Button>
        </div>
      )
    },
  ];

  return (
    <Table
      columns={columns}
      data={rules}
      rowKey="id"
      loading={loading}
      emptyText="暂无告警规则"
    />
  );
}