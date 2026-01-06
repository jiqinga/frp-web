import { memo } from 'react';
import { Input } from '../../../../../components/ui/Input';
import { Select } from '../../../../../components/ui/Select';
import { Switch } from '../../../../../components/ui/Switch';
import { Tooltip } from '../../../../../components/ui/Tooltip';
import { HelpCircle } from 'lucide-react';
import type { FormData } from '../types';
import type { DNSProvider, ProxyType } from '../../../../../types';
import { generateDomainsFromPrefixes, checkCertificateMatch } from '../utils';
import type { Certificate } from '../../../../../api/certificate';
import { PROXY_TYPE_FIELDS } from '../../../constants';

interface DNSSyncSectionProps {
  formData: FormData;
  currentProxyType: ProxyType;
  dnsProviders: DNSProvider[];
  providerDomains: string[];
  loadingDomains: boolean;
  certificates: Certificate[];
  errors: Record<string, string>;
  onFieldChange: (field: keyof FormData, value: string | number | undefined | boolean) => void;
  onDnsSyncChange: (checked: boolean) => void;
  onDnsProviderChange: (providerId: number | undefined) => void;
  onRootDomainChange: (rootDomain: string) => void;
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

export const DNSSyncSection = memo(function DNSSyncSection({
  formData,
  currentProxyType,
  dnsProviders,
  providerDomains,
  loadingDomains,
  certificates,
  errors,
  onFieldChange,
  onDnsSyncChange,
  onDnsProviderChange,
  onRootDomainChange,
}: DNSSyncSectionProps) {
  const shouldShowField = (fieldName: string): boolean => {
    const fields = PROXY_TYPE_FIELDS[currentProxyType];
    return fields.includes(fieldName);
  };

  if (!shouldShowField('custom_domains')) {
    return null;
  }

  return (
    <div className="space-y-3">
      {renderLabel('DNS 自动同步', '启用后，选择 DNS 提供商和根域名，只需填写域名前缀即可自动生成完整域名')}
      <div className="flex items-center gap-2">
        <Switch
          checked={formData.enable_dns_sync}
          onChange={onDnsSyncChange}
          size="sm"
        />
        <span className="text-sm text-foreground-muted">
          {formData.enable_dns_sync ? '启用' : '关闭'}
        </span>
      </div>
      
      {formData.enable_dns_sync && (
        <div className="space-y-3 pl-4 border-l-2 border-border">
          <Select
            label="DNS 提供商"
            value={formData.dns_provider_id}
            onChange={(value) => onDnsProviderChange(value as number | undefined)}
            options={dnsProviders.map(p => ({ value: p.id, label: `${p.name} (${p.type})` }))}
            placeholder="请选择 DNS 提供商"
          />
          
          {formData.dns_provider_id && (
            <Select
              label="根域名"
              value={formData.dns_root_domain}
              onChange={(value) => onRootDomainChange(value as string)}
              options={providerDomains.map(d => ({ value: d, label: d }))}
              placeholder={loadingDomains ? '加载中...' : '请选择根域名'}
              disabled={loadingDomains || providerDomains.length === 0}
            />
          )}
          
          {formData.dns_provider_id && !loadingDomains && providerDomains.length === 0 && (
            <p className="text-xs text-warning">该提供商下没有托管的域名</p>
          )}
          
          {formData.dns_root_domain && (
            <div className="space-y-1.5">
              {renderLabel('域名前缀', '输入域名前缀，多个前缀用逗号分隔')}
              <div className="flex items-center gap-2">
                <div className="flex-1">
                  <Input
                    value={formData.domain_prefixes}
                    onChange={(e) => onFieldChange('domain_prefixes', e.target.value)}
                    placeholder="如: app, api, www"
                    error={errors.domain_prefixes}
                  />
                </div>
                <span className="text-sm text-foreground-muted whitespace-nowrap">
                  .{formData.dns_root_domain}
                </span>
              </div>
              {formData.domain_prefixes && (
                <p className="text-xs text-foreground-subtle">
                  将生成: {generateDomainsFromPrefixes(formData.domain_prefixes, formData.dns_root_domain).split(',').join(', ')}
                </p>
              )}
              {formData.cert_id && formData.dns_root_domain && (() => {
                const selectedCert = certificates.find(c => c.id === formData.cert_id);
                if (!selectedCert) return null;
                const { matches, warning } = checkCertificateMatch(
                  selectedCert.domain,
                  formData.dns_root_domain,
                  formData.domain_prefixes
                );
                if (!matches && warning) {
                  return <p className="text-xs text-warning">{warning}</p>;
                }
                return null;
              })()}
            </div>
          )}
        </div>
      )}

      {/* 自定义域名 - 未启用 DNS 同步时显示 */}
      {!formData.enable_dns_sync && (
        <div className="space-y-1.5">
          {renderLabel('自定义域名', '多个域名用逗号分隔')}
          <Input
            value={formData.custom_domains}
            onChange={(e) => onFieldChange('custom_domains', e.target.value)}
            placeholder="如: example.com, www.example.com"
            error={errors.custom_domains}
          />
        </div>
      )}
    </div>
  );
});