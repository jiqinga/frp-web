import { useState } from 'react';
import { Cloud, Monitor, Edit, Trash2, Server } from 'lucide-react';
import { Table, type Column } from '../../../components/ui/Table';
import { Button } from '../../../components/ui/Button';
import { Tooltip } from '../../../components/ui/Tooltip';
import { Modal } from '../../../components/ui/Modal';
import type { FrpServer } from '../../../api/frpServer';
import { getStatusBadge, getTypeBadge } from './ServerBadges';
import { RemoteServerActions } from './RemoteServerActions';
import { LocalServerActions } from './LocalServerActions';
import { cn } from '../../../utils/cn';
import type { OperationType } from '../hooks/useOperationLoading';

interface ServerTableProps {
  servers: FrpServer[];
  loading: boolean;
  onTestSSH: (id: number) => void;
  onRemoteInstall: (id: number) => void;
  onRemoteStart: (id: number) => void;
  onRemoteStop: (id: number) => void;
  onRemoteRestart: (id: number) => void;
  onViewAuth: (server: FrpServer) => void;
  onRefreshRemoteVersion: (id: number) => void;
  onViewMetrics: (id: number) => void;
  onRemoteReinstall: (id: number) => void;
  onRemoteUpgrade: (id: number) => void;
  onViewLogs: (id: number) => void;
  onTestConnection: (server: FrpServer) => void;
  onRefreshLocalVersion: (id: number) => void;
  onEdit: (server: FrpServer) => void;
  onDelete: (server: FrpServer) => void;
  isLoading?: (serverId: number, operation: OperationType) => boolean;
}

export function ServerTable({
  servers, loading, onTestSSH, onRemoteInstall, onRemoteStart, onRemoteStop,
  onRemoteRestart, onViewAuth, onRefreshRemoteVersion, onViewMetrics,
  onRemoteReinstall, onRemoteUpgrade, onViewLogs, onTestConnection,
  onRefreshLocalVersion, onEdit, onDelete, isLoading,
}: ServerTableProps) {
  const [deleteConfirmVisible, setDeleteConfirmVisible] = useState(false);
  const [serverToDelete, setServerToDelete] = useState<FrpServer | null>(null);

  const handleDeleteClick = (server: FrpServer) => {
    setServerToDelete(server);
    setDeleteConfirmVisible(true);
  };

  const handleDeleteConfirm = () => {
    if (serverToDelete) onDelete(serverToDelete);
    setDeleteConfirmVisible(false);
    setServerToDelete(null);
  };

  const columns: Column<FrpServer>[] = [
    {
      key: 'name', title: '名称', dataIndex: 'name', width: 150,
      render: (value: unknown, record: FrpServer) => (
        <div className="flex items-center gap-2">
          {record.server_type === 'remote' ? <Cloud className="h-4 w-4 text-blue-400" /> : <Monitor className="h-4 w-4 text-green-400" />}
          <span className="font-medium text-foreground">{value as string}</span>
        </div>
      ),
    },
    { key: 'server_type', title: '类型', dataIndex: 'server_type', width: 100, render: (value: unknown) => getTypeBadge(value as string) },
    {
      key: 'host', title: '主机', dataIndex: 'host', width: 150,
      render: (_: unknown, record: FrpServer) => (
        <span className="font-mono text-sm text-foreground-secondary">{record.server_type === 'remote' ? record.ssh_host : record.host}</span>
      ),
    },
    { key: 'dashboard_port', title: '控制面板', dataIndex: 'dashboard_port', width: 100, render: (value: unknown) => <span className="font-mono text-foreground-secondary">{value as number}</span> },
    { key: 'bind_port', title: '绑定端口', dataIndex: 'bind_port', width: 100, render: (value: unknown) => <span className="font-mono text-foreground-secondary">{value as number}</span> },
    { key: 'status', title: '状态', dataIndex: 'status', width: 100, render: (value: unknown) => getStatusBadge(value as string) },
    { key: 'version', title: '版本', dataIndex: 'version', width: 100, render: (value: unknown) => <span className="font-mono text-sm text-foreground-muted">{(value as string) || '-'}</span> },
    {
      key: 'actions', title: '操作', width: 600,
      render: (_: unknown, record: FrpServer) => (
        <div className="flex items-center gap-2">
          {record.server_type === 'remote' ? (
            <RemoteServerActions server={record} onTestSSH={onTestSSH} onRemoteInstall={onRemoteInstall}
              onRemoteStart={onRemoteStart} onRemoteStop={onRemoteStop} onRemoteRestart={onRemoteRestart}
              onViewAuth={onViewAuth} onRefreshRemoteVersion={onRefreshRemoteVersion} onViewMetrics={onViewMetrics}
              onRemoteReinstall={onRemoteReinstall} onRemoteUpgrade={onRemoteUpgrade} onViewLogs={onViewLogs} isLoading={isLoading} />
          ) : (
            <LocalServerActions server={record} onTestConnection={onTestConnection} onViewAuth={onViewAuth}
              onRefreshLocalVersion={onRefreshLocalVersion} onViewMetrics={onViewMetrics} isLoading={isLoading} />
          )}
          <div className="w-px h-6 bg-border mx-1" />
          <Tooltip content="编辑服务器">
            <Button variant="ghost" size="sm" icon={<Edit className="h-3.5 w-3.5" />} onClick={() => onEdit(record)} className="text-foreground-muted hover:text-indigo-400" />
          </Tooltip>
          <Tooltip content="删除服务器">
            <Button variant="ghost" size="sm" icon={<Trash2 className="h-3.5 w-3.5" />} onClick={() => handleDeleteClick(record)} className="text-foreground-muted hover:text-red-400" />
          </Tooltip>
        </div>
      ),
    },
  ];

  return (
    <>
      <div className="relative">
        <div className="absolute -inset-0.5 bg-gradient-to-r from-indigo-500/20 via-purple-500/20 to-cyan-500/20 rounded-lg blur opacity-30" />
        <div className={cn(
          "relative backdrop-blur-sm rounded-lg border overflow-hidden bg-surface border-border"
        )}>
          <div className="absolute inset-0 overflow-hidden pointer-events-none">
            <div className="absolute inset-0 bg-gradient-to-b from-transparent via-indigo-500/5 to-transparent h-[200%] animate-scan" />
          </div>
          <Table columns={columns} data={servers} rowKey="id" loading={loading} emptyText="暂无服务器数据" />
          {!loading && servers.length === 0 && (
            <div className="flex flex-col items-center justify-center py-12 text-foreground-muted">
              <Server className="h-12 w-12 mb-4 opacity-50" />
              <p>暂无服务器数据</p>
              <p className="text-sm mt-1">点击"添加服务器"按钮创建新服务器</p>
            </div>
          )}
        </div>
      </div>
      <Modal open={deleteConfirmVisible} onClose={() => setDeleteConfirmVisible(false)} title="确认删除" size="sm"
        footer={
          <div className="flex justify-end gap-3">
            <Button variant="ghost" onClick={() => setDeleteConfirmVisible(false)}>取消</Button>
            <Button variant="danger" onClick={handleDeleteConfirm}>确认删除</Button>
          </div>
        }>
        <div className="py-4">
          <p className="text-foreground-secondary">确定要删除服务器 <span className="text-indigo-400 font-medium">"{serverToDelete?.name}"</span> 吗？</p>
          <p className="text-sm mt-2 text-foreground-muted">此操作不可恢复，删除后相关配置将丢失。</p>
        </div>
      </Modal>
    </>
  );
}