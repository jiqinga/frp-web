import { HelpCircle } from 'lucide-react';
import { Input } from '../../../../components/ui/Input';
import { Select, type SelectOption } from '../../../../components/ui/Select';
import { Switch } from '../../../../components/ui/Switch';
import { Tooltip } from '../../../../components/ui/Tooltip';
import type { PluginType, Client } from '../../../../types';
import { FIELD_LABELS, FIELD_PLACEHOLDERS, PROXY_TYPE_LABELS } from '../../constants';
import type { FormData } from './types';
import { PluginConfigSection } from '../PluginConfigSection';

interface BasicInfoSectionProps {
  formData: FormData;
  editingProxy: boolean;
  clients: Client[];
  onlineClientIds: Set<number>;
  pluginEnabled: boolean;
  currentPluginType: PluginType;
  pluginConfig: Record<string, string>;
  errors: Record<string, string>;
  updateField: (field: keyof FormData, value: string | number | boolean | undefined) => void;
  handleTypeChange: (value: string | number) => void;
  handlePluginEnabledChange: (checked: boolean) => void;
  handlePluginTypeChange: (value: PluginType) => void;
  handlePluginConfigChange: (field: string, value: string) => void;
}

const renderLabel = (label: string, tooltip?: string) => (
  <div className="flex items-center gap-1">
    <span>{label}</span>
    {tooltip && (
      <Tooltip content={tooltip}>
        <span className="cursor-help text-foreground-subtle">
          <HelpCircle className="h-3.5 w-3.5" />
        </span>
      </Tooltip>
    )}
  </div>
);

export function BasicInfoSection({
  formData,
  editingProxy,
  clients,
  onlineClientIds,
  pluginEnabled,
  currentPluginType,
  pluginConfig,
  errors,
  updateField,
  handleTypeChange,
  handlePluginEnabledChange,
  handlePluginTypeChange,
  handlePluginConfigChange,
}: BasicInfoSectionProps) {
  const clientOptions: SelectOption[] = clients.map(c => {
    const isOnline = onlineClientIds.has(c.id!);
    return {
      value: c.id!,
      label: isOnline ? c.name : `${c.name} (离线)`,
      disabled: !isOnline,
    };
  });

  const proxyTypeOptions: SelectOption[] = Object.entries(PROXY_TYPE_LABELS).map(([value, label]) => ({
    value,
    label,
  }));

  return (
    <>
      {!editingProxy && (
        <Select
          label="客户端"
          value={formData.client_id}
          onChange={(value) => updateField('client_id', value as number)}
          options={clientOptions}
          placeholder="请选择客户端"
          error={errors.client_id}
        />
      )}
      
      <Input
        label={FIELD_LABELS.name}
        value={formData.name}
        onChange={(e) => updateField('name', e.target.value)}
        placeholder={FIELD_PLACEHOLDERS.name}
        error={errors.name}
      />
      
      <Select
        label="类型"
        value={formData.type}
        onChange={handleTypeChange}
        options={proxyTypeOptions}
        disabled={pluginEnabled}
      />

      <div className="space-y-1.5">
        <div className="flex items-center gap-2">
          {renderLabel('启用插件', '启用插件后，代理类型将自动设为 TCP，本地地址配置将被插件配置替代')}
        </div>
        <div className="flex items-center gap-2">
          <Switch
            checked={pluginEnabled}
            onChange={handlePluginEnabledChange}
            size="sm"
          />
          <span className="text-sm text-foreground-muted">
            {pluginEnabled ? '启用' : '关闭'}
          </span>
        </div>
      </div>

      {pluginEnabled && (
        <PluginConfigSection
          pluginType={currentPluginType}
          pluginConfig={pluginConfig}
          onPluginTypeChange={handlePluginTypeChange}
          onPluginConfigChange={handlePluginConfigChange}
        />
      )}
      
      {!pluginEnabled && (
        <>
          <Input
            label={FIELD_LABELS.local_ip}
            value={formData.local_ip}
            onChange={(e) => updateField('local_ip', e.target.value)}
            placeholder={FIELD_PLACEHOLDERS.local_ip}
            error={errors.local_ip}
          />
          <Input
            label={FIELD_LABELS.local_port}
            type="number"
            value={formData.local_port?.toString() || ''}
            onChange={(e) => updateField('local_port', e.target.value ? parseInt(e.target.value) : undefined)}
            placeholder={FIELD_PLACEHOLDERS.local_port}
            error={errors.local_port}
          />
        </>
      )}
    </>
  );
}