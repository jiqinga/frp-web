import { memo, useCallback } from 'react';
import { Select } from '../../../../../components/ui/Select';
import { Tooltip } from '../../../../../components/ui/Tooltip';
import { HelpCircle } from 'lucide-react';
import type { FormData } from '../types';
import type { ProxyType } from '../../../../../types';
import type { Certificate } from '../../../../../api/certificate';
import { generateDomainsFromPrefixes } from '../utils';

interface CertificateSectionProps {
  formData: FormData;
  currentProxyType: ProxyType;
  certificates: Certificate[];
  loadingCertificates: boolean;
  errors: Record<string, string>;
  onCertificateChange: (certId: number | undefined) => void;
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

export const CertificateSection = memo(function CertificateSection({
  formData,
  currentProxyType,
  certificates,
  loadingCertificates,
  errors,
  onCertificateChange,
}: CertificateSectionProps) {
  const getMatchingCertificates = useCallback(() => {
    const domain = formData.enable_dns_sync && formData.dns_root_domain && formData.domain_prefixes
      ? generateDomainsFromPrefixes(formData.domain_prefixes, formData.dns_root_domain).split(',')[0]
      : formData.custom_domains?.split(',')[0]?.trim();
    
    if (!domain) return certificates;
    
    return certificates.filter(cert => {
      if (cert.domain === domain) return true;
      if (cert.domain.startsWith('*.')) {
        const wildcardBase = cert.domain.slice(2);
        const domainParts = domain.split('.');
        if (domainParts.length >= 2) {
          const domainBase = domainParts.slice(1).join('.');
          if (domainBase === wildcardBase) return true;
        }
      }
      return false;
    });
  }, [certificates, formData.custom_domains, formData.enable_dns_sync, formData.dns_root_domain, formData.domain_prefixes]);

  if (currentProxyType !== 'https') {
    return null;
  }

  const matchingCerts = getMatchingCertificates();

  return (
    <div className="space-y-1.5">
      {renderLabel('SSL 证书', 'HTTPS 代理必须选择证书，选择后可自动填充 DNS 同步配置')}
      <Select
        value={formData.cert_id}
        onChange={(value) => onCertificateChange(value as number | undefined)}
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
  );
});