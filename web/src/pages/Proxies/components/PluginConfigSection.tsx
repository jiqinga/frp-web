import { HelpCircle } from 'lucide-react';
import { Input } from '../../../components/ui/Input';
import { Select, type SelectOption } from '../../../components/ui/Select';
import { Tooltip } from '../../../components/ui/Tooltip';
import type { PluginType } from '../../../types';
import {
  PLUGIN_TYPE_LABELS,
  PLUGIN_TYPE_FIELDS,
  PLUGIN_REQUIRED_FIELDS,
  PLUGIN_FIELD_LABELS,
  PLUGIN_FIELD_TOOLTIPS,
  PLUGIN_FIELD_PLACEHOLDERS,
} from '../constants';

interface PluginConfigSectionProps {
  pluginType: PluginType;
  pluginConfig: Record<string, string>;
  onPluginTypeChange: (type: PluginType) => void;
  onPluginConfigChange: (field: string, value: string) => void;
}

export function PluginConfigSection({
  pluginType,
  pluginConfig,
  onPluginTypeChange,
  onPluginConfigChange,
}: PluginConfigSectionProps) {
  // 判断插件字段是否必填
  const isPluginFieldRequired = (fieldName: string): boolean => {
    const requiredFields = PLUGIN_REQUIRED_FIELDS[pluginType];
    return requiredFields.includes(fieldName);
  };

  // 插件类型选项
  const pluginTypeOptions: SelectOption[] = Object.entries(PLUGIN_TYPE_LABELS).map(([value, label]) => ({
    value,
    label,
  }));

  // 渲染带提示的标签
  const renderLabel = (label: string, tooltip?: string, required?: boolean) => (
    <div className="flex items-center gap-1">
      <span>{label}</span>
      {required && <span className="text-red-400">*</span>}
      {tooltip && (
        <Tooltip content={tooltip}>
          <span className="text-slate-500 cursor-help">
            <HelpCircle className="h-3.5 w-3.5" />
          </span>
        </Tooltip>
      )}
    </div>
  );

  return (
    <div className="space-y-4">
      {/* 分隔线 - 插件配置 */}
      <div className="relative">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-border" />
        </div>
        <div className="relative flex justify-center">
          <span className="px-3 text-sm bg-surface text-foreground-muted">插件配置</span>
        </div>
      </div>
      
      {/* 插件类型选择 */}
      <Select
        label="插件类型"
        value={pluginType}
        onChange={(value) => onPluginTypeChange(value as PluginType)}
        options={pluginTypeOptions}
      />

      {/* 根据插件类型动态显示配置字段 */}
      {PLUGIN_TYPE_FIELDS[pluginType].map(field => {
        const isRequired = isPluginFieldRequired(field);
        const isPassword = field.toLowerCase().includes('password');
        
        return (
          <div key={field} className="space-y-1.5">
            {renderLabel(
              PLUGIN_FIELD_LABELS[field] || field,
              PLUGIN_FIELD_TOOLTIPS[field],
              isRequired
            )}
            <Input
              type={isPassword ? 'password' : 'text'}
              value={pluginConfig[field] || ''}
              onChange={(e) => onPluginConfigChange(field, e.target.value)}
              placeholder={PLUGIN_FIELD_PLACEHOLDERS[field]}
            />
          </div>
        );
      })}

      {/* 分隔线 - 代理配置 */}
      <div className="relative">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-border" />
        </div>
        <div className="relative flex justify-center">
          <span className="px-3 text-sm bg-surface text-foreground-muted">代理配置</span>
        </div>
      </div>
    </div>
  );
}