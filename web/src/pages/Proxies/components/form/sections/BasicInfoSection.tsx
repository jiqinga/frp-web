import { memo } from 'react';
import { Input } from '../../../../../components/ui/Input';
import { Select, type SelectOption } from '../../../../../components/ui/Select';
import { Switch } from '../../../../../components/ui/Switch';
import { Tooltip } from '../../../../../components/ui/Tooltip';
import { HelpCircle } from 'lucide-react';
import type { FormData } from '../types';
import type { ProxyType } from '../../../../../types';
import { FIELD_LABELS, FIELD_PLACEHOLDERS, PROXY_TYPE_LABELS } from '../../../constants';

interface BasicInfoSectionProps {
  formData: FormData;
  currentProxyType: ProxyType;
  pluginEnabled: boolean;
  isEditing: boolean;
  clientOptions: SelectOption[];
  errors: Record<string, string>;
  onFieldChange: (field: keyof FormData, value: string | number | undefined) => void;
  onTypeChange: (value: string | number) => void;
  onPluginEnabledChange: (checked: boolean) => void;
}

const proxyTypeOptions: SelectOption[] = Object.entries(PROXY_TYPE_LABELS).map(([value, label]) => ({
  value,
  label,
}));

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

export const BasicInfoSection = memo(function BasicInfoSection({
  formData,
  pluginEnabled,
  isEditing,
  clientOptions,
  errors,
  onFieldChange,
  onTypeChange,
  onPluginEnabledChange,
}: BasicInfoSectionProps) {
  return (
    <>
      {/* 客户端选择 - 新增时显示 */}
      {!isEditing && (
        <Select
          label="客户端"
          value={formData.client_id}
          onChange={(value) => onFieldChange('client_id', value as number)}
          options={clientOptions}
          placeholder="请选择客户端"
          error={errors.client_id}
        />
      )}
      
      {/* 名称 */}
      <Input
        label={FIELD_LABELS.name}
        value={formData.name}
        onChange={(e) => onFieldChange('name', e.target.value)}
        placeholder={FIELD_PLACEHOLDERS.name}
        error={errors.name}
      />
      
      {/* 类型选择 */}
      <Select
        label="类型"
        value={formData.type}
        onChange={onTypeChange}
        options={proxyTypeOptions}
        disabled={pluginEnabled}
      />

      {/* 启用插件开关 */}
      <div className="space-y-1.5">
        <div className="flex items-center gap-2">
          {renderLabel('启用插件', '启用插件后，代理类型将自动设为 TCP，本地地址配置将被插件配置替代')}
        </div>
        <div className="flex items-center gap-2">
          <Switch
            checked={pluginEnabled}
            onChange={onPluginEnabledChange}
            size="sm"
          />
          <span className="text-sm text-foreground-muted">
            {pluginEnabled ? '启用' : '关闭'}
          </span>
        </div>
      </div>
    </>
  );
});