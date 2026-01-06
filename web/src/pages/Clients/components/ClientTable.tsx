import { Info, RefreshCw, Edit2, Eye, Trash2, FileText, Play, Square, RotateCcw } from 'lucide-react';
import { Table, type Column } from '../../../components/ui/Table';
import { Button } from '../../../components/ui/Button';
import { Badge } from '../../../components/ui/Badge';
import { Tooltip } from '../../../components/ui/Tooltip';
import { Pagination } from '../../../components/ui/Pagination';
import type { Client } from '../../../types';
import { CONFIG_SYNC_STATUS_LABELS, CONFIG_SYNC_STATUS_VARIANTS, type ConfigSyncStatus } from '../../../types';
import type { FrpcAction } from '../hooks/useFrpcControl';

interface ClientTableProps {
  clients: Client[];
  loading: boolean;
  page: number;
  total: number;
  onPageChange: (page: number) => void;
  onEdit: (client: Client) => void;
  onDelete: (id: number) => void;
  onViewConfig: (client: Client) => void;
  onUpdate: (client: Client) => void;
  onViewLogs: (client: Client) => void;
  onFrpcStart: (client: Client) => void;
  onFrpcStop: (client: Client) => void;
  onFrpcRestart: (client: Client) => void;
  frpcLoadingMap: Record<number, FrpcAction | null>;
}

// frpc 状态配置
const frpcStatusConfig: Record<string, { variant: 'success' | 'danger' | 'default'; text: string }> = {
  online: { variant: 'success', text: 'frpc在线' },
  offline: { variant: 'danger', text: 'frpc离线' },
  unknown: { variant: 'default', text: 'frpc未知' }
};

export function ClientTable({
  clients,
  loading,
  page,
  total,
  onPageChange,
  onEdit,
  onDelete,
  onViewConfig,
  onUpdate,
  onViewLogs,
  onFrpcStart,
  onFrpcStop,
  onFrpcRestart,
  frpcLoadingMap,
}: ClientTableProps) {
  const columns: Column<Client>[] = [
    {
      key: 'name',
      title: '名称',
      dataIndex: 'name',
    },
    {
      key: 'status',
      title: '状态',
      render: (_: unknown, record: Client) => {
        const frpcConfig = frpcStatusConfig[record.online_status || 'unknown'] || frpcStatusConfig.unknown;
        
        const tooltipContent = (
          <div className="space-y-1 text-xs">
            {record.last_heartbeat && (
              <div>最后心跳: {new Date(record.last_heartbeat).toLocaleString('zh-CN')}</div>
            )}
            {record.config_version && (
              <div>配置版本: v{record.config_version}</div>
            )}
            {record.last_config_sync && (
              <div>最后同步: {new Date(record.last_config_sync).toLocaleString('zh-CN')}</div>
            )}
            {record.frpc_version && (
              <div>frpc版本: {record.frpc_version}</div>
            )}
            {record.daemon_version && (
              <div>daemon版本: {record.daemon_version}</div>
            )}
            {record.os && record.arch && (
              <div>系统: {record.os}/{record.arch}</div>
            )}
          </div>
        );
        
        return (
          <div className="flex items-center justify-center gap-2">
            <Badge variant={frpcConfig.variant} size="sm" dot pulse={record.online_status === 'online'}>
              {frpcConfig.text}
            </Badge>
            <Badge
              variant={record.ws_connected ? 'info' : 'default'}
              size="sm"
              dot
              pulse={record.ws_connected}
            >
              WS{record.ws_connected ? '在线' : '离线'}
            </Badge>
            <Tooltip content={tooltipContent}>
              <Info className="h-4 w-4 text-indigo-400 hover:text-indigo-300 cursor-pointer transition-colors" />
            </Tooltip>
          </div>
        );
      },
    },
    {
      key: 'server_addr',
      title: '服务器地址',
      dataIndex: 'server_addr',
    },
    {
      key: 'server_port',
      title: '端口',
      dataIndex: 'server_port',
    },
    {
      key: 'config_sync',
      title: '配置状态',
      render: (_: unknown, record: Client) => {
        const status = record.config_sync_status as ConfigSyncStatus | undefined;
        if (!status) {
          return <span className="text-slate-500">-</span>;
        }
        
        const variant = CONFIG_SYNC_STATUS_VARIANTS[status];
        const label = CONFIG_SYNC_STATUS_LABELS[status];
        
        const badge = (
          <Badge variant={variant} size="sm" dot pulse={status === 'pending'}>
            {label}
          </Badge>
        );
        
        if (record.config_sync_error) {
          return (
            <Tooltip content={<span className="text-xs">{record.config_sync_error}</span>}>
              {badge}
            </Tooltip>
          );
        }
        
        if (record.config_sync_time) {
          return (
            <Tooltip content={<span className="text-xs">同步时间: {new Date(record.config_sync_time).toLocaleString('zh-CN')}</span>}>
              {badge}
            </Tooltip>
          );
        }
        
        return badge;
      },
    },
    {
      key: 'version',
      title: '版本',
      render: (_: unknown, record: Client) => (
        <div className="flex flex-col items-center gap-0.5 text-xs">
          {record.frpc_version && <span className="text-slate-400">frpc: {record.frpc_version}</span>}
          {record.daemon_version && <span className="text-slate-400">daemon: {record.daemon_version}</span>}
          {!record.frpc_version && !record.daemon_version && <span className="text-slate-500">-</span>}
        </div>
      ),
    },
    {
      key: 'remark',
      title: '备注',
      dataIndex: 'remark',
      render: (value: unknown) => (
        <span className="text-slate-400">{(value as string) || '-'}</span>
      ),
    },
    {
      key: 'action',
      title: '操作',
      render: (_: unknown, record: Client) => (
        <div className="flex items-center justify-center gap-1">
          <Tooltip content="编辑">
            <Button
              size="sm"
              variant="ghost"
              icon={<Edit2 className="h-3.5 w-3.5" />}
              onClick={() => onEdit(record)}
            />
          </Tooltip>
          <Tooltip content={record.ws_connected ? "查看日志" : "仅WS在线时可查看"}>
            <Button
              size="sm"
              variant="ghost"
              icon={<FileText className="h-3.5 w-3.5" />}
              onClick={() => onViewLogs(record)}
              disabled={!record.ws_connected}
            />
          </Tooltip>
          <Tooltip content={record.ws_connected ? "重启frpc" : "仅WS在线时可控制"}>
            <Button
              size="sm"
              variant="ghost"
              icon={<RotateCcw className={`h-3.5 w-3.5 ${frpcLoadingMap[record.id] === 'restart' ? 'animate-spin' : ''}`} />}
              onClick={() => onFrpcRestart(record)}
              disabled={!record.ws_connected || !!frpcLoadingMap[record.id]}
              loading={frpcLoadingMap[record.id] === 'restart'}
            />
          </Tooltip>
          {record.online_status === 'online' ? (
            <Tooltip content={record.ws_connected ? "停止frpc" : "仅WS在线时可控制"}>
              <Button
                size="sm"
                variant="ghost"
                icon={<Square className="h-3.5 w-3.5 text-red-400" />}
                onClick={() => onFrpcStop(record)}
                disabled={!record.ws_connected || !!frpcLoadingMap[record.id]}
                loading={frpcLoadingMap[record.id] === 'stop'}
                className="hover:bg-red-500/10"
              />
            </Tooltip>
          ) : (
            <Tooltip content={record.ws_connected ? "启动frpc" : "仅WS在线时可控制"}>
              <Button
                size="sm"
                variant="ghost"
                icon={<Play className="h-3.5 w-3.5 text-green-400" />}
                onClick={() => onFrpcStart(record)}
                disabled={!record.ws_connected || !!frpcLoadingMap[record.id]}
                loading={frpcLoadingMap[record.id] === 'start'}
                className="hover:bg-green-500/10"
              />
            </Tooltip>
          )}
          <Tooltip content={record.ws_connected ? "更新软件" : "仅WS在线时可更新"}>
            <Button
              size="sm"
              variant="ghost"
              icon={<RefreshCw className="h-3.5 w-3.5" />}
              onClick={() => onUpdate(record)}
              disabled={!record.ws_connected}
            />
          </Tooltip>
          <Tooltip content="查看配置">
            <Button
              size="sm"
              variant="ghost"
              icon={<Eye className="h-3.5 w-3.5" />}
              onClick={() => onViewConfig(record)}
            />
          </Tooltip>
          <Tooltip content="删除">
            <Button
              size="sm"
              variant="ghost"
              icon={<Trash2 className="h-3.5 w-3.5 text-red-400" />}
              onClick={() => onDelete(record.id)}
              className="hover:bg-red-500/10"
            />
          </Tooltip>
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-4">
      {/* 表格容器 - 添加科技风样式 */}
      <div className="relative">
        {/* 发光边框效果 */}
        <div className="absolute -inset-0.5 bg-gradient-to-r from-indigo-500/20 via-purple-500/20 to-indigo-500/20 rounded-xl blur opacity-30" />
        
        <div className="relative rounded-xl overflow-hidden backdrop-blur-sm border border-border bg-surface/80">
          {/* 表格 */}
          <Table
            columns={columns}
            data={clients}
            rowKey="id"
            loading={loading}
            rowClassName={() =>
              'transition-all duration-200 hover:bg-surface-hover hover:shadow-lg hover:shadow-indigo-500/5'
            }
          />
        </div>
      </div>

      {/* 分页 */}
      {total > 10 && (
        <Pagination
          current={page}
          total={total}
          pageSize={10}
          onChange={onPageChange}
          showTotal
        />
      )}
    </div>
  );
}