import { useState, useMemo } from 'react';
import { Search, List, LayoutGrid, ChevronDown, ChevronRight, ArrowUpFromLine, ArrowDownToLine } from 'lucide-react';
import { formatBytes } from '../../utils/websocket';
import { SparkLine } from './SparkLine';
import { Card, CardHeader, CardContent, Input, Select, Badge, Table, Tabs } from '../ui';
import type { TrafficData, ClientGroup, ProxyHistory } from '../../hooks/useRealtimeMonitor';

interface ProxyListProps {
  trafficData: TrafficData[];
  clientGroups: ClientGroup[];
  getProxyHistory: (proxyId: number) => ProxyHistory[];
  onProxyClick?: (proxy: TrafficData) => void;
}

type ViewMode = 'flat' | 'grouped';
type StatusFilter = 'all' | 'online' | 'offline';

// 状态筛选选项
const statusOptions = [
  { value: 'all', label: '全部' },
  { value: 'online', label: '在线' },
  { value: 'offline', label: '离线' },
];

// 可折叠的客户端组
interface ClientGroupItemProps {
  group: ClientGroup & { proxies: TrafficData[] };
  columns: Array<{ key: string; title: string; render?: (value: unknown, record: TrafficData) => React.ReactNode }>;
  getProxyHistory: (proxyId: number) => ProxyHistory[];
  onProxyClick?: (proxy: TrafficData) => void;
}

const ClientGroupItem = ({ group, columns, getProxyHistory, onProxyClick }: ClientGroupItemProps) => {
  const [expanded, setExpanded] = useState(true);

  return (
    <div className="border-b border-border last:border-b-0">
      <button
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center justify-between px-4 py-3 text-left hover:bg-surface-hover"
      >
        <div className="flex items-center gap-3">
          {expanded ? (
            <ChevronDown className="h-4 w-4 text-foreground-muted" />
          ) : (
            <ChevronRight className="h-4 w-4 text-foreground-muted" />
          )}
          <span className="font-medium text-foreground">{group.client_name}</span>
          <Badge variant="info">{group.onlineCount}/{group.proxies.length}</Badge>
        </div>
        <div className="flex items-center gap-4 text-sm">
          <span className="flex items-center gap-1 text-green-400">
            <ArrowUpFromLine className="h-3.5 w-3.5" />
            {formatBytes(group.totalInRate)}/s
          </span>
          <span className="flex items-center gap-1 text-blue-400">
            <ArrowDownToLine className="h-3.5 w-3.5" />
            {formatBytes(group.totalOutRate)}/s
          </span>
        </div>
      </button>
      {expanded && (
        <div className="border-t border-border">
          <Table
            columns={columns.map(col => ({
              ...col,
              render: col.key === 'trend' 
                ? (_: unknown, record: TrafficData) => <SparkLine data={getProxyHistory(record.proxy_id)} />
                : col.render
            }))}
            data={group.proxies}
            rowKey="proxy_id"
            onRowClick={onProxyClick}
          />
        </div>
      )}
    </div>
  );
};

export function ProxyList({ trafficData, clientGroups, getProxyHistory, onProxyClick }: ProxyListProps) {
  const [search, setSearch] = useState('');
  const [viewMode, setViewMode] = useState<ViewMode>('grouped');
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [clientFilter, setClientFilter] = useState<string>('all');

  const filteredData = useMemo(() => {
    return trafficData.filter(item => {
      if (search && !item.proxy_name.toLowerCase().includes(search.toLowerCase())) return false;
      if (statusFilter === 'online' && !item.online) return false;
      if (statusFilter === 'offline' && item.online) return false;
      if (clientFilter !== 'all' && item.client_id !== Number(clientFilter)) return false;
      return true;
    });
  }, [trafficData, search, statusFilter, clientFilter]);

  const filteredGroups = useMemo(() => {
    return clientGroups.map(group => ({
      ...group,
      proxies: group.proxies.filter(item => {
        if (search && !item.proxy_name.toLowerCase().includes(search.toLowerCase())) return false;
        if (statusFilter === 'online' && !item.online) return false;
        if (statusFilter === 'offline' && item.online) return false;
        return true;
      }),
    })).filter(group => group.proxies.length > 0);
  }, [clientGroups, search, statusFilter]);

  const columns = [
    {
      key: 'proxy_name',
      title: '代理名称',
      render: (_: unknown, record: TrafficData) => (
        <span className="font-medium text-foreground">{record.proxy_name}</span>
      )
    },
    {
      key: 'online',
      title: '状态',
      render: (_: unknown, record: TrafficData) => (
        <Badge variant={record.online ? 'success' : 'default'}>
          {record.online ? '在线' : '离线'}
        </Badge>
      ),
    },
    {
      key: 'bytes_in_rate',
      title: '上传',
      render: (_: unknown, record: TrafficData) => (
        <span className="text-green-400">{formatBytes(record.bytes_in_rate)}/s</span>
      ),
    },
    {
      key: 'bytes_out_rate',
      title: '下载',
      render: (_: unknown, record: TrafficData) => (
        <span className="text-blue-400">{formatBytes(record.bytes_out_rate)}/s</span>
      ),
    },
    {
      key: 'trend',
      title: '趋势',
      render: (_: unknown, record: TrafficData) => (
        <SparkLine data={getProxyHistory(record.proxy_id)} />
      ),
    },
    {
      key: 'total',
      title: '总流量',
      render: (_: unknown, record: TrafficData) => (
        <span className="text-foreground-secondary">{formatBytes(record.total_bytes_in + record.total_bytes_out)}</span>
      ),
    },
  ];

  const clientOptions = [
    { value: 'all', label: '全部客户端' },
    ...clientGroups.map(g => ({ value: String(g.client_id), label: g.client_name }))
  ];

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-2">
            <List className="h-5 w-5 text-indigo-400" />
            <span>代理列表</span>
            <Badge variant="default">{trafficData.length} 个代理</Badge>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              <Input
                placeholder="搜索代理"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-40 pl-9"
              />
            </div>
            <Select
              value={statusFilter}
              onChange={(v) => setStatusFilter(v as StatusFilter)}
              options={statusOptions}
              className="w-24"
            />
            <Select
              value={clientFilter}
              onChange={(v) => setClientFilter(String(v))}
              options={clientOptions}
              className="w-36"
            />
            <Tabs
              items={[
                { key: 'grouped', label: '分组', icon: <LayoutGrid className="h-3.5 w-3.5" />, children: null },
                { key: 'flat', label: '平铺', icon: <List className="h-3.5 w-3.5" />, children: null },
              ]}
              activeKey={viewMode}
              onChange={(key) => setViewMode(key as ViewMode)}
              variant="pills"
            />
          </div>
        </div>
      </CardHeader>
      <CardContent className="p-0">
        {viewMode === 'flat' ? (
          <Table
            columns={columns}
            data={filteredData}
            rowKey="proxy_id"
            onRowClick={onProxyClick}
            emptyText="暂无代理数据"
          />
        ) : (
          <div className="divide-y divide-border">
            {filteredGroups.length === 0 ? (
              <div className="py-12 text-center text-foreground-muted">
                暂无代理数据
              </div>
            ) : (
              filteredGroups.map(group => (
                <ClientGroupItem
                  key={group.client_id}
                  group={group}
                  columns={columns}
                  getProxyHistory={getProxyHistory}
                  onProxyClick={onProxyClick}
                />
              ))
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}