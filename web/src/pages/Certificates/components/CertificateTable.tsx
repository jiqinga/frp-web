import { RefreshCw, Trash2, RotateCcw, Download } from 'lucide-react';
import { Table, Badge, Button, Tooltip, Switch } from '../../../components/ui';
import type { Certificate } from '../../../api/certificate';
import type { DNSProvider } from '../../../types';

interface CertificateTableProps {
  certificates: Certificate[];
  providers: DNSProvider[];
  loading: boolean;
  renewingId: number | null;
  reapplyingId: number | null;
  togglingAutoRenewId: number | null;
  onRenew: (id: number) => void;
  onReapply: (id: number) => void;
  onToggleAutoRenew: (id: number, autoRenew: boolean) => void;
  onDelete: (id: number) => void;
  onDownload: (id: number) => void;
}

const STATUS_MAP: Record<string, { label: string; variant: 'success' | 'warning' | 'danger' | 'default' | 'info' }> = {
  pending: { label: '申请中', variant: 'info' },
  active: { label: '有效', variant: 'success' },
  expiring: { label: '即将过期', variant: 'warning' },
  expired: { label: '已过期', variant: 'danger' },
  failed: { label: '失败', variant: 'danger' },
};

function formatDateTime(dateStr: string | null): string {
  if (!dateStr) return '-';
  const date = new Date(dateStr);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  });
}

export function CertificateTable({ certificates, providers, loading, renewingId, reapplyingId, togglingAutoRenewId, onRenew, onReapply, onToggleAutoRenew, onDelete, onDownload }: CertificateTableProps) {
  const getProviderName = (providerId: number) => {
    const provider = providers.find(p => p.id === providerId);
    return provider?.name || '-';
  };

  const columns = [
    {
      key: 'domain',
      title: '域名',
      align: 'center' as const,
      render: (_: unknown, record: Certificate) => (
        <span className="font-mono text-foreground">{record.domain}</span>
      ),
    },
    {
      key: 'status',
      title: '状态',
      align: 'center' as const,
      render: (_: unknown, record: Certificate) => {
        const status = STATUS_MAP[record.status] || { label: record.status, variant: 'default' as const };
        return <Badge variant={status.variant}>{status.label}</Badge>;
      },
    },
    {
      key: 'provider',
      title: 'DNS提供商',
      align: 'center' as const,
      render: (_: unknown, record: Certificate) => (
        <span className="text-foreground-secondary">{getProviderName(record.provider_id)}</span>
      ),
    },
    {
      key: 'validity',
      title: '有效期',
      align: 'center' as const,
      render: (_: unknown, record: Certificate) => (
        <div className="text-foreground-secondary text-xs space-y-1">
          <div>签发: {formatDateTime(record.not_before)}</div>
          <div>到期: {formatDateTime(record.not_after)}</div>
        </div>
      ),
    },
    {
      key: 'auto_renew',
      title: '自动续期',
      align: 'center' as const,
      render: (_: unknown, record: Certificate) => (
        <div className="flex justify-center">
          <Switch
            checked={record.auto_renew}
            onChange={(checked) => onToggleAutoRenew(record.id, checked)}
            disabled={togglingAutoRenewId === record.id || record.status === 'pending' || record.status === 'failed'}
          />
        </div>
      ),
    },
    {
      key: 'action',
      title: '操作',
      align: 'center' as const,
      render: (_: unknown, record: Certificate) => (
        <div className="flex items-center justify-center gap-1">
          {record.status === 'failed' ? (
            <Tooltip content={reapplyingId === record.id ? '重新申请中...' : '重新申请'}>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => onReapply(record.id)}
                disabled={reapplyingId === record.id}
              >
                <RotateCcw className={`h-4 w-4 text-orange-400 ${reapplyingId === record.id ? 'animate-spin' : ''}`} />
              </Button>
            </Tooltip>
          ) : (
            <Tooltip content={renewingId === record.id ? '续期中...' : '续期'}>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => onRenew(record.id)}
                disabled={renewingId === record.id || record.status === 'pending'}
              >
                <RefreshCw className={`h-4 w-4 ${renewingId === record.id ? 'animate-spin' : ''}`} />
              </Button>
            </Tooltip>
          )}
          <Tooltip content="下载">
            <Button
              size="sm"
              variant="ghost"
              onClick={() => onDownload(record.id)}
              disabled={record.status !== 'active' && record.status !== 'expiring'}
            >
              <Download className="h-4 w-4 text-blue-400" />
            </Button>
          </Tooltip>
          <Tooltip content="删除">
            <Button size="sm" variant="ghost" onClick={() => onDelete(record.id)}>
              <Trash2 className="h-4 w-4 text-red-400" />
            </Button>
          </Tooltip>
        </div>
      ),
    },
  ];

  return (
    <Table
      columns={columns}
      data={certificates}
      rowKey="id"
      loading={loading}
      emptyText="暂无证书"
    />
  );
}