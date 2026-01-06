import { useState, useEffect, useCallback } from 'react';
import { Copy, ExternalLink, Plug, HelpCircle, LineChart, Pencil, Trash2, Cloud, CloudOff, RefreshCw, AlertCircle } from 'lucide-react';
import { Table, type Column } from '../../../components/ui/Table';
import { TechTableContainer } from '../../../components/ui/TechTableContainer';
import { Button } from '../../../components/ui/Button';
import { Badge } from '../../../components/ui/Badge';
import { Tooltip } from '../../../components/ui/Tooltip';
import { ConfirmModal } from '../../../components/ui/ConfirmModal';
import { Switch } from '../../../components/ui/Switch';
import { toast } from '../../../components/ui/Toast';
import type { Proxy, Client, ProxyType, PluginType, DNSRecord, DNSRecordStatus } from '../../../types';
import { PROXY_TYPE_LABELS, PLUGIN_TYPE_LABELS } from '../constants';
import { dnsApi } from '../../../api/dns';

// 代理类型到 Badge variant 的映射
const proxyTypeVariantMap: Record<string, 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
  tcp: 'info',
  udp: 'success',
  http: 'warning',
  https: 'danger',
  stcp: 'primary',
};

// 插件类型到 Badge variant 的映射
const pluginTypeVariantMap: Record<string, 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
  http_proxy: 'info',
  socks5: 'primary',
  static_file: 'success',
  unix_domain_socket: 'warning',
  http2https: 'info',
  https2http: 'warning',
  https2https: 'danger',
};

interface ProxyTableProps {
  proxies: Proxy[];
  loading: boolean;
  getClientById: (clientId: number) => Client | undefined;
  getAccessUrl: (proxy: Proxy) => { url: string; isClickable: boolean };
  isClientOnline: (clientId: number) => boolean;
  onEdit: (proxy: Proxy) => void;
  onDelete: (id: number, deleteDNS?: boolean) => void;
  onToggle: (id: number) => void;
  isToggling: (id: number) => boolean;
  onShowPluginUsage: (proxy: Proxy) => void;
  onShowTrafficHistory: (proxy: Proxy) => void;
}

// DNS 状态图标组件
const DNSStatusIcon = ({ status, error }: { status?: DNSRecordStatus; error?: string }) => {
  if (!status) return null;
  
  const iconMap: Record<DNSRecordStatus, { icon: React.ReactNode; color: string; label: string }> = {
    pending: { icon: <RefreshCw className="h-3.5 w-3.5 animate-spin" />, color: 'text-yellow-400', label: '同步中' },
    synced: { icon: <Cloud className="h-3.5 w-3.5" />, color: 'text-green-400', label: 'DNS已同步' },
    failed: { icon: <AlertCircle className="h-3.5 w-3.5" />, color: 'text-red-400', label: `同步失败: ${error || '未知错误'}` },
    deleted: { icon: <CloudOff className="h-3.5 w-3.5" />, color: 'text-gray-400', label: 'DNS已删除' },
  };
  
  const { icon, color, label } = iconMap[status] || iconMap.pending;
  
  return (
    <Tooltip content={label}>
      <span className={`${color} cursor-help`}>{icon}</span>
    </Tooltip>
  );
};

export function ProxyTable({
  proxies,
  loading,
  getClientById,
  getAccessUrl,
  isClientOnline,
  onEdit,
  onDelete,
  onToggle,
  isToggling,
  onShowPluginUsage,
  onShowTrafficHistory,
}: ProxyTableProps) {
  // 删除确认模态框状态
  const [deleteModalVisible, setDeleteModalVisible] = useState(false);
  const [proxyToDelete, setProxyToDelete] = useState<Proxy | null>(null);
  const [deleteDNSChecked, setDeleteDNSChecked] = useState(true); // 默认勾选删除DNS记录
  
  // DNS 记录状态缓存
  const [dnsRecords, setDnsRecords] = useState<Record<number, DNSRecord>>({});

  // 获取所有 DNS 记录
  const fetchDnsRecords = useCallback(async () => {
    try {
      const records = await dnsApi.getRecords();
      const recordMap: Record<number, DNSRecord> = {};
      records.forEach(record => {
        recordMap[record.proxy_id] = record;
      });
      setDnsRecords(recordMap);
    } catch {
      // 静默失败，DNS 记录显示不是关键功能
    }
  }, []);

  useEffect(() => {
    fetchDnsRecords();
  }, [fetchDnsRecords, proxies]);

  // 复制访问地址到剪贴板
  const copyToClipboard = (text: string) => {
    if (text && !text.includes('未') && text !== '-' && !text.includes('STCP')) {
      navigator.clipboard.writeText(text).then(() => {
        toast.success('已复制到剪贴板');
      }).catch(() => {
        toast.error('复制失败');
      });
    }
  };

  // 处理删除确认
  const handleDeleteClick = (proxy: Proxy) => {
    setProxyToDelete(proxy);
    setDeleteDNSChecked(true); // 重置为默认勾选
    setDeleteModalVisible(true);
  };

  const handleDeleteConfirm = () => {
    if (proxyToDelete) {
      onDelete(proxyToDelete.id, deleteDNSChecked);
    }
    setDeleteModalVisible(false);
    setProxyToDelete(null);
  };

  // 获取本地地址显示内容
  const getLocalAddrDisplay = (record: Proxy) => {
    if (record.plugin_type) {
      try {
        const config = JSON.parse(record.plugin_config || '{}');
        switch (record.plugin_type) {
          case 'static_file':
            return config.localPath || '静态文件';
          case 'unix_domain_socket':
            return config.unixPath || 'Unix套接字';
          default:
            return '插件模式';
        }
      } catch {
        return '插件模式';
      }
    }
    return `${record.local_ip}:${record.local_port}`;
  };

  const columns: Column<Proxy>[] = [
    {
      key: 'name',
      title: '名称',
      dataIndex: 'name',
      width: 120,
    },
    {
      key: 'client_name',
      title: '客户端',
      width: 120,
      render: (_value: unknown, record: Proxy) => {
        const client = getClientById(record.client_id);
        const online = isClientOnline(record.client_id);
        return client ? (
          <div className="flex items-center justify-center gap-1.5">
            <span className={`w-2 h-2 rounded-full ${online ? 'bg-green-400' : 'bg-gray-400'}`} />
            <Badge variant={online ? 'info' : 'default'} size="sm">{client.name}</Badge>
            {!online && <span className="text-xs text-gray-500">离线</span>}
          </div>
        ) : (
          <span className="text-foreground-muted">未知</span>
        );
      },
    },
    {
      key: 'type',
      title: '类型',
      dataIndex: 'type',
      width: 70,
      render: (value: unknown) => {
        const type = (value as string)?.toLowerCase();
        return (
          <Badge variant={proxyTypeVariantMap[type] || 'default'} size="sm">
            {PROXY_TYPE_LABELS[type as ProxyType] || (value as string)?.toUpperCase()}
          </Badge>
        );
      },
    },
    {
      key: 'plugin_type',
      title: '插件',
      dataIndex: 'plugin_type',
      width: 120,
      render: (value: unknown, record: Proxy) => {
        const pluginType = value as string;
        if (!pluginType) {
          return <span className="text-foreground-muted">-</span>;
        }
        return (
          <div className="flex items-center justify-center gap-1">
            <Tooltip content="点击查看使用说明">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onShowPluginUsage(record)}
                className="!p-0 h-auto"
              >
                <Badge variant={pluginTypeVariantMap[pluginType as PluginType] || 'default'} size="sm">
                  <Plug className="h-3 w-3" />
                  {PLUGIN_TYPE_LABELS[pluginType as PluginType] || pluginType}
                </Badge>
              </Button>
            </Tooltip>
            <Tooltip content="查看使用说明">
              <Button
                variant="ghost"
                size="sm"
                icon={<HelpCircle className="h-3.5 w-3.5" />}
                onClick={() => onShowPluginUsage(record)}
                className="!p-1"
              />
            </Tooltip>
          </div>
        );
      },
    },
    {
      key: 'local_addr',
      title: '本地地址',
      width: 140,
      render: (_value: unknown, record: Proxy) => {
        const display = getLocalAddrDisplay(record);
        const isPluginMode = !!record.plugin_type;
        return (
          <span className={isPluginMode ? 'text-foreground-muted' : 'text-foreground-secondary'}>
            {display}
          </span>
        );
      },
    },
    {
      key: 'access_url',
      title: '访问地址',
      width: 240,
      render: (_value: unknown, record: Proxy) => {
        const { url, isClickable } = getAccessUrl(record);
        const isValidUrl = url && !url.includes('未') && url !== '-' && !url.includes('STCP');
        const dnsRecord = dnsRecords[record.id];
        const showDnsStatus = record.enable_dns_sync && dnsRecord;
        
        return (
          <div className="flex items-center justify-center gap-1">
            {/* DNS 状态图标 */}
            {showDnsStatus && (
              <DNSStatusIcon status={dnsRecord.status} error={dnsRecord.last_error} />
            )}
            {isClickable ? (
              <Tooltip content={url}>
                <a
                  href={url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-indigo-400 hover:text-indigo-300 transition-colors truncate max-w-[160px] inline-block"
                >
                  {url}
                </a>
              </Tooltip>
            ) : (
              <Tooltip content={url}>
                <span
                  className={`truncate max-w-[160px] inline-block ${
                    isValidUrl ? 'text-indigo-400' : 'text-slate-500'
                  }`}
                >
                  {url}
                </span>
              </Tooltip>
            )}
            {isValidUrl && (
              <>
                <Tooltip content="复制">
                  <Button
                    variant="ghost"
                    size="sm"
                    icon={<Copy className="h-3.5 w-3.5" />}
                    onClick={() => copyToClipboard(url)}
                    className="!p-1"
                  />
                </Tooltip>
                {isClickable && (
                  <Tooltip content="在新窗口打开">
                    <Button
                      variant="ghost"
                      size="sm"
                      icon={<ExternalLink className="h-3.5 w-3.5" />}
                      onClick={() => window.open(url, '_blank', 'noopener,noreferrer')}
                      className="!p-1"
                    />
                  </Tooltip>
                )}
              </>
            )}
          </div>
        );
      },
    },
    {
      key: 'enabled',
      title: '状态',
      width: 80,
      render: (_value: unknown, record: Proxy) => {
        const isEnabled = record.enabled !== false;
        const toggling = isToggling(record.id);
        const online = isClientOnline(record.client_id);
        const tooltipContent = !online
          ? '客户端离线，无法操作'
          : (toggling ? '切换中...' : (isEnabled ? '点击禁用' : '点击启用'));
        return (
          <Tooltip content={tooltipContent}>
            <div>
              <Switch
                checked={isEnabled}
                onChange={() => onToggle(record.id)}
                size="sm"
                loading={toggling}
                disabled={!online}
              />
            </div>
          </Tooltip>
        );
      },
    },
    {
      key: 'action',
      title: '操作',
      width: 160,
      render: (_value: unknown, record: Proxy) => {
        const online = isClientOnline(record.client_id);
        return (
          <div className="flex items-center justify-center gap-1">
            <Tooltip content="流量历史">
              <Button
                variant="ghost"
                size="sm"
                icon={<LineChart className="h-4 w-4" />}
                onClick={() => onShowTrafficHistory(record)}
                className="!p-1.5"
              />
            </Tooltip>
            <Tooltip content={online ? '编辑' : '客户端离线，无法编辑'}>
              <span className="inline-block">
                <Button
                  variant="ghost"
                  size="sm"
                  icon={<Pencil className="h-4 w-4" />}
                  onClick={() => onEdit(record)}
                  className="!p-1.5"
                  disabled={!online}
                />
              </span>
            </Tooltip>
            <Tooltip content={online ? '删除' : '客户端离线，无法删除'}>
              <span className="inline-block">
                <Button
                  variant="ghost"
                  size="sm"
                  icon={<Trash2 className={`h-4 w-4 ${online ? 'text-red-400' : 'text-gray-400'}`} />}
                  onClick={() => handleDeleteClick(record)}
                  className="!p-1.5 hover:!bg-red-500/10"
                  disabled={!online}
                />
              </span>
            </Tooltip>
          </div>
        );
      },
    },
  ];

  return (
    <>
      <TechTableContainer>
        <Table
          columns={columns}
          data={proxies}
          rowKey="id"
          loading={loading}
          emptyText="暂无代理配置"
          size="sm"
        />
      </TechTableContainer>

      {/* 删除确认模态框 */}
      <ConfirmModal
        open={deleteModalVisible}
        onClose={() => setDeleteModalVisible(false)}
        onConfirm={handleDeleteConfirm}
        title="确认删除"
        type="warning"
        confirmText="确认删除"
        content={
          <>
            <p className="text-foreground-secondary">
              确定要删除代理 <span className="text-indigo-400 font-medium">{proxyToDelete?.name}</span> 吗？
            </p>
            <p className="text-sm text-foreground-muted mt-2">删除后将无法恢复</p>
            {/* DNS 删除选项 - 仅当代理启用了 DNS 同步时显示 */}
            {proxyToDelete?.enable_dns_sync && (
              <label className="flex items-center gap-2 mt-4 cursor-pointer">
                <input
                  type="checkbox"
                  checked={deleteDNSChecked}
                  onChange={(e) => setDeleteDNSChecked(e.target.checked)}
                  className="w-4 h-4 rounded border-border bg-surface text-indigo-500 focus:ring-indigo-500 focus:ring-offset-0"
                />
                <span className="text-sm text-foreground-secondary">
                  同时删除对应的 DNS 解析记录
                </span>
              </label>
            )}
          </>
        }
      />
    </>
  );
}