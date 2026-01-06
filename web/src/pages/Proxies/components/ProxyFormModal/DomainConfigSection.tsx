import { HelpCircle } from 'lucide-react';
import { Input } from '../../../../components/ui/Input';
import { Select } from '../../../../components/ui/Select';
import { Switch } from '../../../../components/ui/Switch';
import { Tooltip } from '../../../../components/ui/Tooltip';
import type { ProxyType, DNSProvider } from '../../../../types';
import type { Certificate } from '../../../../api/certificate';
import { FIELD_LABELS, FIELD_TOOLTIPS, FIELD_PLACEHOLDERS } from '../../constants';
import type { FormData } from './types';
import { generateDomainsFromPrefixes, checkCertificateMatch } from './utils';

interface DomainConfigSectionProps {
  formData: FormData;
  currentProxyType: ProxyType;
  certificates: Certificate[];
  loadingCertificates: boolean;
  dnsProviders: DNSProvider[];
  providerDomains: string[];
  loadingDomains: boolean;
  errors: Record<string, string>;
  shouldShowField: (field: string) => boolean;
  getMatchingCertificates: () => Certificate[];
  updateField: (field: keyof FormData, value: string | number | boolean | undefined) => void;
  handleCertificateChange: (certId: number | undefined) => void;
  handleDnsSyncChange: (checked: boolean) => void;
  handleDnsProviderChange: (providerId: number | undefined) => void;
  handleRootDomainChange: (rootDomain: string) => void;
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

export function DomainConfigSection({
  formData,
  currentProxyType,
  certificates,
  loadingCertificates,
  dnsProviders,
  providerDomains,
  loadingDomains,
  errors,
  shouldShowField,
  getMatchingCertificates,
  updateField,
  handleCertificateChange,
  handleDnsSyncChange,
  handleDnsProviderChange,
  handleRootDomainChange,
}: DomainConfigSectionProps) {
  const matchingCerts = getMatchingCertificates();

  return (
    <>
      {currentProxyType === 'https' && (
        <div className="space-y-1.5">
          {renderLabel('SSL 证书', 'HTTPS 代理必须选择证书，选择后可自动填充 DNS 同步配置')}
          <Select
            value={formData.cert_id}
            onChange={(value) => handleCertificateChange(value as number | undefined)}
            options={matchingCerts.length > 0
              ? matchingCerts.map(cert => ({
                  value: cert.id,
                  label: `${cert.domain} (${cert.status === 'active' ? '有效' : cert.status})`,
                }))
              : certificates.map(cert => ({
                  value: cert.id,
                  label: `${cert.domain} (${cert.status === 'active' ? '有效' : cert.status})`,
                }))
            }
            placeholder={loadingCertificates ? '加载中...' : '请选择证书'}
            disabled={loadingCertificates}
            error={errors.cert_id}
          />
          {certificates.length === 0 && !loadingCertificates && (
            <p className="text-xs text-warning">暂无可用证书，请先在证书管理中申请证书</p>
          )}
          {certificates.length > 0 && matchingCerts.length === 0 && (
            <p className="text-xs text-warning">没有匹配当前域名的证书，显示所有可用证书</p>
          )}
        </div>
      )}

      {shouldShowField('custom_domains') && (
        <div className="space-y-3">
          {renderLabel('DNS 自动同步', '启用后，选择 DNS 提供商和根域名，只需填写域名前缀即可自动生成完整域名')}
          <div className="flex items-center gap-2">
            <Switch
              checked={formData.enable_dns_sync}
              onChange={handleDnsSyncChange}
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
                onChange={(value) => handleDnsProviderChange(value as number | undefined)}
                options={dnsProviders.map(p => ({ value: p.id, label: `${p.name} (${p.type})` }))}
                placeholder="请选择 DNS 提供商"
              />
              
              {formData.dns_provider_id && (
                <Select
                  label="根域名"
                  value={formData.dns_root_domain}
                  onChange={(value) => handleRootDomainChange(value as string)}
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
                        onChange={(e) => updateField('domain_prefixes', e.target.value)}
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
        </div>
      )}

      {shouldShowField('custom_domains') && !formData.enable_dns_sync && (
        <div className="space-y-1.5">
          {renderLabel(FIELD_LABELS.custom_domains, FIELD_TOOLTIPS.custom_domains)}
          <Input
            value={formData.custom_domains}
            onChange={(e) => updateField('custom_domains', e.target.value)}
            placeholder={FIELD_PLACEHOLDERS.custom_domains}
            error={errors.custom_domains}
          />
        </div>
      )}
      
      {shouldShowField('subdomain') && (
        <div className="space-y-1.5">
          {renderLabel(FIELD_LABELS.subdomain, FIELD_TOOLTIPS.subdomain)}
          <Input
            value={formData.subdomain}
            onChange={(e) => updateField('subdomain', e.target.value)}
            placeholder={FIELD_PLACEHOLDERS.subdomain}
          />
        </div>
      )}
    </>
  );
}