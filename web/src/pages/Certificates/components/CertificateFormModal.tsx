import { useState, useEffect, useMemo } from 'react';
import { RefreshCw } from 'lucide-react';
import { Modal, Select, Button, Input, Switch } from '../../../components/ui';
import { toast } from '../../../components/ui';
import { dnsApi } from '../../../api/dns';
import type { DNSProvider } from '../../../types';

interface CertificateFormModalProps {
  visible: boolean;
  providers: DNSProvider[];
  onCancel: () => void;
  onSubmit: (data: { proxy_id: number; domain: string; dns_provider_id: number; auto_renew?: boolean }) => Promise<boolean>;
}

export function CertificateFormModal({ visible, providers, onCancel, onSubmit }: CertificateFormModalProps) {
  const [rootDomain, setRootDomain] = useState('');
  const [subdomain, setSubdomain] = useState('');
  const [providerId, setProviderId] = useState<number | ''>('');
  const [autoRenew, setAutoRenew] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [domains, setDomains] = useState<string[]>([]);
  const [domainsLoading, setDomainsLoading] = useState(false);

  const providerOptions = providers.map(p => ({ value: p.id, label: p.name }));
  const domainOptions = domains.map(d => ({ value: d, label: d }));

  // 计算最终域名
  const finalDomain = useMemo(() => {
    if (!rootDomain) return '';
    return subdomain ? `${subdomain}.${rootDomain}` : rootDomain;
  }, [rootDomain, subdomain]);

  // 当选择DNS提供商时，获取该提供商下的域名列表
  useEffect(() => {
    if (!providerId) {
      setDomains([]);
      setRootDomain('');
      setSubdomain('');
      return;
    }

    const fetchDomains = async () => {
      setDomainsLoading(true);
      setRootDomain('');
      setSubdomain('');
      try {
        const data = await dnsApi.getProviderDomains(providerId as number);
        setDomains(data || []);
      } catch {
        toast.error('获取域名列表失败');
        setDomains([]);
      } finally {
        setDomainsLoading(false);
      }
    };

    fetchDomains();
  }, [providerId]);

  const handleSubmit = async () => {
    if (!subdomain || !finalDomain || !providerId) return;
    setSubmitting(true);
    const success = await onSubmit({ proxy_id: 0, domain: finalDomain, dns_provider_id: providerId as number, auto_renew: autoRenew });
    setSubmitting(false);
    if (success) {
      setRootDomain('');
      setSubdomain('');
      setProviderId('');
      setAutoRenew(true);
      onCancel();
    }
  };

  const handleClose = () => {
    setRootDomain('');
    setSubdomain('');
    setProviderId('');
    setAutoRenew(true);
    setDomains([]);
    onCancel();
  };

  return (
    <Modal open={visible} onClose={handleClose} title="申请SSL证书">
      <div className="space-y-4">
        <Select
          label="DNS提供商"
          value={providerId}
          onChange={(value) => setProviderId(Number(value))}
          options={providerOptions}
          placeholder="选择DNS提供商"
        />
        <div>
          <label className="block text-sm font-medium text-foreground-secondary mb-1.5">域名</label>
          <div className="grid grid-cols-[1fr_auto_1fr] items-center gap-1">
            <Input
              value={subdomain}
              onChange={(e) => setSubdomain(e.target.value.replace(/[^a-zA-Z0-9-]/g, ''))}
              placeholder="前缀（必填）"
              disabled={!providerId || domainsLoading}
              required
            />
            <span className="text-foreground-secondary text-center">.</span>
            <Select
              value={rootDomain}
              onChange={(value) => setRootDomain(value as string)}
              options={domainOptions}
              placeholder={domainsLoading ? '加载中...' : (providerId ? '选择根域名' : '请先选择DNS提供商')}
              disabled={!providerId || domainsLoading}
            />
          </div>
          {domainsLoading && (
            <div className="flex items-center gap-2 mt-1 text-xs text-foreground-secondary">
              <RefreshCw className="h-3 w-3 animate-spin" />
              <span>正在获取域名列表...</span>
            </div>
          )}
          {!domainsLoading && providerId && domains.length === 0 && (
            <p className="mt-1 text-xs text-yellow-500">该DNS提供商下暂无托管域名</p>
          )}
        </div>
        {finalDomain && (
          <div className="p-3 bg-background-secondary rounded-md">
            <p className="text-xs text-foreground-secondary mb-1">最终证书域名：</p>
            <p className="font-mono text-sm text-foreground">{finalDomain}</p>
          </div>
        )}
        <div className="flex items-center justify-between">
          <label className="text-sm font-medium text-foreground-secondary">自动续期</label>
          <Switch checked={autoRenew} onChange={setAutoRenew} />
        </div>
        <div className="flex justify-end gap-3 pt-4">
          <Button variant="secondary" onClick={handleClose}>取消</Button>
          <Button onClick={handleSubmit} disabled={submitting || !subdomain || !finalDomain || !providerId}>
            {submitting ? '申请中...' : '申请证书'}
          </Button>
        </div>
      </div>
    </Modal>
  );
}