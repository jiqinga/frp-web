import { memo } from 'react';
import { Input } from '../../../../../components/ui/Input';
import { Tooltip } from '../../../../../components/ui/Tooltip';
import { HelpCircle } from 'lucide-react';
import type { FormData } from '../types';
import type { ProxyType } from '../../../../../types';
import { FIELD_LABELS, FIELD_TOOLTIPS, FIELD_PLACEHOLDERS, PROXY_TYPE_FIELDS } from '../../../constants';

interface NetworkConfigSectionProps {
  formData: FormData;
  currentProxyType: ProxyType;
  pluginEnabled: boolean;
  errors: Record<string, string>;
  onFieldChange: (field: keyof FormData, value: string | number | undefined) => void;
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

export const NetworkConfigSection = memo(function NetworkConfigSection({
  formData,
  currentProxyType,
  pluginEnabled,
  errors,
  onFieldChange,
}: NetworkConfigSectionProps) {
  const shouldShowField = (fieldName: string): boolean => {
    const fields = PROXY_TYPE_FIELDS[currentProxyType];
    return fields.includes(fieldName);
  };

  return (
    <>
      {/* 本地IP - 不使用插件时显示 */}
      {!pluginEnabled && (
        <Input
          label={FIELD_LABELS.local_ip}
          value={formData.local_ip}
          onChange={(e) => onFieldChange('local_ip', e.target.value)}
          placeholder={FIELD_PLACEHOLDERS.local_ip}
          error={errors.local_ip}
        />
      )}
      
      {/* 本地端口 - 不使用插件时显示 */}
      {!pluginEnabled && (
        <Input
          label={FIELD_LABELS.local_port}
          type="number"
          value={formData.local_port?.toString() || ''}
          onChange={(e) => onFieldChange('local_port', e.target.value ? parseInt(e.target.value) : undefined)}
          placeholder={FIELD_PLACEHOLDERS.local_port}
          error={errors.local_port}
        />
      )}
      
      {/* 远程端口 - 仅 TCP/UDP */}
      {shouldShowField('remote_port') && (
        <div className="space-y-1.5">
          {renderLabel(FIELD_LABELS.remote_port, FIELD_TOOLTIPS.remote_port)}
          <Input
            type="number"
            value={formData.remote_port?.toString() || ''}
            onChange={(e) => onFieldChange('remote_port', e.target.value ? parseInt(e.target.value) : undefined)}
            placeholder={FIELD_PLACEHOLDERS.remote_port}
          />
        </div>
      )}

      {/* 子域名 - 仅 HTTP/HTTPS */}
      {shouldShowField('subdomain') && (
        <div className="space-y-1.5">
          {renderLabel(FIELD_LABELS.subdomain, FIELD_TOOLTIPS.subdomain)}
          <Input
            value={formData.subdomain}
            onChange={(e) => onFieldChange('subdomain', e.target.value)}
            placeholder={FIELD_PLACEHOLDERS.subdomain}
          />
        </div>
      )}
      
      {/* URL路由 - 仅 HTTP */}
      {shouldShowField('locations') && (
        <div className="space-y-1.5">
          {renderLabel(FIELD_LABELS.locations, FIELD_TOOLTIPS.locations)}
          <Input
            value={formData.locations}
            onChange={(e) => onFieldChange('locations', e.target.value)}
            placeholder={FIELD_PLACEHOLDERS.locations}
          />
        </div>
      )}
      
      {/* Host重写 - 仅 HTTP/HTTPS */}
      {shouldShowField('host_header_rewrite') && (
        <div className="space-y-1.5">
          {renderLabel(FIELD_LABELS.host_header_rewrite, FIELD_TOOLTIPS.host_header_rewrite)}
          <Input
            value={formData.host_header_rewrite}
            onChange={(e) => onFieldChange('host_header_rewrite', e.target.value)}
            placeholder={FIELD_PLACEHOLDERS.host_header_rewrite}
          />
        </div>
      )}
      
      {/* HTTP用户名 - 仅 HTTP */}
      {shouldShowField('http_user') && (
        <div className="space-y-1.5">
          {renderLabel(FIELD_LABELS.http_user, FIELD_TOOLTIPS.http_user)}
          <Input
            value={formData.http_user}
            onChange={(e) => onFieldChange('http_user', e.target.value)}
            placeholder={FIELD_PLACEHOLDERS.http_user}
          />
        </div>
      )}
      
      {/* HTTP密码 - 仅 HTTP */}
      {shouldShowField('http_password') && (
        <div className="space-y-1.5">
          {renderLabel(FIELD_LABELS.http_password, FIELD_TOOLTIPS.http_password)}
          <Input
            type="password"
            value={formData.http_password}
            onChange={(e) => onFieldChange('http_password', e.target.value)}
            placeholder={FIELD_PLACEHOLDERS.http_password}
          />
        </div>
      )}
      
      {/* 密钥 - 仅 STCP */}
      {shouldShowField('secret_key') && (
        <div className="space-y-1.5">
          {renderLabel(FIELD_LABELS.secret_key, FIELD_TOOLTIPS.secret_key)}
          <Input
            type="password"
            value={formData.secret_key}
            onChange={(e) => onFieldChange('secret_key', e.target.value)}
            placeholder={FIELD_PLACEHOLDERS.secret_key}
            error={errors.secret_key}
          />
        </div>
      )}
      
      {/* 允许用户 - 仅 STCP */}
      {shouldShowField('allow_users') && (
        <div className="space-y-1.5">
          {renderLabel(FIELD_LABELS.allow_users, FIELD_TOOLTIPS.allow_users)}
          <Input
            value={formData.allow_users}
            onChange={(e) => onFieldChange('allow_users', e.target.value)}
            placeholder={FIELD_PLACEHOLDERS.allow_users}
          />
        </div>
      )}
    </>
  );
});