import { useState, useEffect } from 'react';
import type { AlertRule, AlertTargetType } from '../../../api/alert';
import { systemRuleTypeNames } from '../../../api/alert';
import type { Client } from '../../../types';
import type { FrpServer } from '../../../api/frpServer';
import { alertRecipientApi, type AlertRecipient, type AlertRecipientGroup } from '../../../api/alertRecipient';
import { Button, Input, Select, Switch, Modal, Checkbox, NumberStepper } from '../../../components/ui';

interface AlertRuleFormModalProps {
  visible: boolean;
  editingRule: AlertRule | null;
  clients: Client[];
  servers: FrpServer[];
  onCancel: () => void;
  onSubmit: (rule: AlertRule) => void;
}

const targetTypeOptions = [
  { value: 'proxy', label: '代理流量' },
  { value: 'frpc', label: 'frpc离线' },
  { value: 'frps', label: 'frps离线' },
  { value: 'system', label: '系统告警' },
];

const systemRuleTypeOptions = Object.entries(systemRuleTypeNames).map(([value, label]) => ({ value, label }));

const proxyRuleTypeOptions = [
  { value: 'daily', label: '每日流量' },
  { value: 'monthly', label: '每月流量' },
  { value: 'rate', label: '实时速率' },
];

const unitOptions = [
  { value: 'MB', label: 'MB' },
  { value: 'GB', label: 'GB' },
  { value: 'TB', label: 'TB' },
];

const convertToBytes = (value: number, unit: string): number => {
  const multipliers: Record<string, number> = {
    'bytes': 1, 'MB': 1024 * 1024, 'GB': 1024 * 1024 * 1024, 'TB': 1024 * 1024 * 1024 * 1024,
  };
  return value * (multipliers[unit] || 1);
};

const convertFromBytes = (bytes: number): { value: number; unit: string } => {
  if (bytes >= 1024 * 1024 * 1024 * 1024) return { value: bytes / (1024 * 1024 * 1024 * 1024), unit: 'TB' };
  if (bytes >= 1024 * 1024 * 1024) return { value: bytes / (1024 * 1024 * 1024), unit: 'GB' };
  if (bytes >= 1024 * 1024) return { value: bytes / (1024 * 1024), unit: 'MB' };
  return { value: bytes, unit: 'bytes' };
};

export function AlertRuleFormModal({
  visible, editingRule, clients, servers, onCancel, onSubmit,
}: AlertRuleFormModalProps) {
  const [formData, setFormData] = useState({
    target_type: 'proxy' as AlertTargetType,
    target_id: '',
    rule_type: 'daily',
    threshold_value: '',
    threshold_unit: 'GB',
    cooldown_minutes: '60',
    offline_delay_seconds: '60',
    notify_on_recovery: true,
    notify_recipient_ids: [] as number[],
    notify_group_ids: [] as number[],
    enabled: true
  });
  const [recipients, setRecipients] = useState<AlertRecipient[]>([]);
  const [groups, setGroups] = useState<AlertRecipientGroup[]>([]);

  useEffect(() => {
    if (visible) {
      alertRecipientApi.getRecipients().then(r => setRecipients(r || []));
      alertRecipientApi.getGroups().then(g => setGroups(g || []));
    }
  }, [visible]);

  useEffect(() => {
    if (editingRule) {
      const { value, unit } = convertFromBytes(editingRule.threshold_value);
      setFormData({
        target_type: editingRule.target_type || 'proxy',
        target_id: String(editingRule.target_id || editingRule.proxy_id),
        rule_type: editingRule.rule_type,
        threshold_value: String(value),
        threshold_unit: unit,
        cooldown_minutes: String(editingRule.cooldown_minutes || 60),
        offline_delay_seconds: String(editingRule.offline_delay_seconds || 60),
        notify_on_recovery: editingRule.notify_on_recovery ?? true,
        notify_recipient_ids: editingRule.notify_recipient_ids ? editingRule.notify_recipient_ids.split(',').map(Number).filter(Boolean) : [],
        notify_group_ids: editingRule.notify_group_ids ? editingRule.notify_group_ids.split(',').map(Number).filter(Boolean) : [],
        enabled: editingRule.enabled
      });
    } else {
      setFormData({
        target_type: 'proxy',
        target_id: '',
        rule_type: 'daily',
        threshold_value: '',
        threshold_unit: 'GB',
        cooldown_minutes: '60',
        offline_delay_seconds: '60',
        notify_on_recovery: true,
        notify_recipient_ids: [],
        notify_group_ids: [],
        enabled: true
      });
    }
  }, [editingRule, visible]);

  const handleSubmit = () => {
    const isSystem = formData.target_type === 'system';
    const isProxy = formData.target_type === 'proxy';
    const data: AlertRule = {
      ...(editingRule?.id ? { id: editingRule.id } : {}),
      target_type: formData.target_type,
      target_id: isSystem ? 0 : Number(formData.target_id),
      proxy_id: isProxy ? Number(formData.target_id) : 0,
      rule_type: isProxy ? formData.rule_type : (isSystem ? formData.rule_type : 'offline'),
      threshold_value: isProxy ? convertToBytes(Number(formData.threshold_value), formData.threshold_unit) : 0,
      threshold_unit: formData.threshold_unit,
      cooldown_minutes: Number(formData.cooldown_minutes),
      offline_delay_seconds: Number(formData.offline_delay_seconds),
      notify_on_recovery: formData.notify_on_recovery,
      notify_recipient_ids: formData.notify_recipient_ids.join(','),
      notify_group_ids: formData.notify_group_ids.join(','),
      enabled: formData.enabled,
    };
    onSubmit(data);
  };

  const getTargetOptions = () => {
    if (formData.target_type === 'frpc') {
      return clients.map(c => ({ value: String(c.id), label: c.name }));
    }
    if (formData.target_type === 'frps') {
      return servers.map(s => ({ value: String(s.id), label: s.name }));
    }
    return [];
  };

  return (
    <Modal open={visible} onClose={onCancel} title={editingRule ? '编辑告警规则' : '新增告警规则'}>
      <div className="space-y-4">
        <div className="space-y-2">
          <label className="text-sm font-medium text-foreground-secondary">告警类型</label>
          <Select
            value={formData.target_type}
            onChange={(value) => setFormData(prev => ({ ...prev, target_type: value as AlertTargetType, target_id: '', rule_type: value === 'proxy' ? 'daily' : (value === 'system' ? 'cert_expiring' : 'offline') }))}
            options={targetTypeOptions}
          />
        </div>

        {formData.target_type === 'system' ? (
          <div className="space-y-2">
            <label className="text-sm font-medium text-foreground-secondary">系统事件类型</label>
            <Select
              value={formData.rule_type}
              onChange={(value) => setFormData(prev => ({ ...prev, rule_type: String(value) }))}
              options={systemRuleTypeOptions}
            />
          </div>
        ) : formData.target_type === 'proxy' ? (
          <Input
            label="代理ID"
            type="number"
            value={formData.target_id}
            onChange={(e) => setFormData(prev => ({ ...prev, target_id: e.target.value }))}
            placeholder="输入代理ID"
            required
          />
        ) : (
          <div className="space-y-2">
            <label className="text-sm font-medium text-foreground-secondary">
              {formData.target_type === 'frpc' ? '选择客户端' : '选择服务器'}
            </label>
            <Select
              value={formData.target_id}
              onChange={(value) => setFormData(prev => ({ ...prev, target_id: String(value) }))}
              options={getTargetOptions()}
              placeholder={formData.target_type === 'frpc' ? '选择客户端' : '选择服务器'}
            />
          </div>
        )}

        {formData.target_type === 'proxy' && (
          <>
            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground-secondary">规则类型</label>
              <Select
                value={formData.rule_type}
                onChange={(value) => setFormData(prev => ({ ...prev, rule_type: String(value) }))}
                options={proxyRuleTypeOptions}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground-secondary">阈值</label>
              <div className="flex gap-2 items-center">
                <NumberStepper
                  value={formData.threshold_value}
                  onChange={(value) => setFormData(prev => ({ ...prev, threshold_value: value }))}
                  min={0}
                  step={1}
                />
                <Select
                  value={formData.threshold_unit}
                  onChange={(value) => setFormData(prev => ({ ...prev, threshold_unit: String(value) }))}
                  options={unitOptions}
                  className="w-24"
                />
              </div>
            </div>
          </>
        )}

        <Input
          label="冷却时间(分钟)"
          type="number"
          value={formData.cooldown_minutes}
          onChange={(e) => setFormData(prev => ({ ...prev, cooldown_minutes: e.target.value }))}
          placeholder="告警冷却时间"
        />

        {formData.target_type !== 'proxy' && (
          <>
            <Input
              label="延迟确认时间(秒)"
              type="number"
              value={formData.offline_delay_seconds}
              onChange={(e) => setFormData(prev => ({ ...prev, offline_delay_seconds: e.target.value }))}
              placeholder="离线多久后才发送告警，防止网络波动误报"
            />
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-foreground-secondary">恢复时通知</span>
              <Switch checked={formData.notify_on_recovery} onChange={(checked) => setFormData(prev => ({ ...prev, notify_on_recovery: checked }))} />
            </div>
          </>
        )}

        <div className="space-y-2">
          <label className="text-sm font-medium text-foreground-secondary">通知接收人</label>
          <div className="max-h-32 overflow-y-auto p-2 rounded space-y-1 bg-surface">
            {recipients.length === 0 ? <span className="text-foreground-subtle text-sm">暂无接收人，请先在设置中添加</span> : recipients.map(r => {
              const toggleRecipient = () => setFormData(f => ({ ...f, notify_recipient_ids: f.notify_recipient_ids.includes(r.id!) ? f.notify_recipient_ids.filter(id => id !== r.id) : [...f.notify_recipient_ids, r.id!] }));
              return (
                <div key={r.id} className="flex items-center gap-2 p-1 rounded cursor-pointer hover:bg-surface-hover" onClick={toggleRecipient}>
                  <Checkbox checked={formData.notify_recipient_ids.includes(r.id!)} onChange={toggleRecipient} size="sm" />
                  <span className="text-sm text-foreground">{r.name}</span>
                  <span className="text-foreground-muted text-xs">{r.email}</span>
                </div>
              );
            })}
          </div>
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium text-foreground-secondary">通知分组</label>
          <div className="max-h-32 overflow-y-auto p-2 rounded space-y-1 bg-surface">
            {groups.length === 0 ? <span className="text-foreground-subtle text-sm">暂无分组，请先在设置中添加</span> : groups.map(g => {
              const toggleGroup = () => setFormData(f => ({ ...f, notify_group_ids: f.notify_group_ids.includes(g.id!) ? f.notify_group_ids.filter(id => id !== g.id) : [...f.notify_group_ids, g.id!] }));
              return (
                <div key={g.id} className="flex items-center gap-2 p-1 rounded cursor-pointer hover:bg-surface-hover" onClick={toggleGroup}>
                  <Checkbox checked={formData.notify_group_ids.includes(g.id!)} onChange={toggleGroup} size="sm" />
                  <span className="text-sm text-foreground">{g.name}</span>
                </div>
              );
            })}
          </div>
        </div>

        <div className="flex items-center justify-between">
          <span className="text-sm font-medium text-foreground-secondary">启用</span>
          <Switch checked={formData.enabled} onChange={(checked) => setFormData(prev => ({ ...prev, enabled: checked }))} />
        </div>

        <div className="flex justify-end gap-3 pt-4">
          <Button variant="secondary" onClick={onCancel}>取消</Button>
          <Button onClick={handleSubmit}>{editingRule ? '更新' : '创建'}</Button>
        </div>
      </div>
    </Modal>
  );
}