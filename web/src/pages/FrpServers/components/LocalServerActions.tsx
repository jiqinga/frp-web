import { CheckCircle, Key, RefreshCw, BarChart3 } from 'lucide-react';
import { Button } from '../../../components/ui/Button';
import { Tooltip } from '../../../components/ui/Tooltip';
import type { FrpServer } from '../../../api/frpServer';
import type { OperationType } from '../hooks/useOperationLoading';

interface LocalServerActionsProps {
  server: FrpServer;
  onTestConnection: (server: FrpServer) => void;
  onViewAuth: (server: FrpServer) => void;
  onRefreshLocalVersion: (id: number) => void;
  onViewMetrics: (id: number) => void;
  isLoading?: (serverId: number, operation: OperationType) => boolean;
}

export function LocalServerActions({
  server, onTestConnection, onViewAuth, onRefreshLocalVersion, onViewMetrics, isLoading,
}: LocalServerActionsProps) {
  const loading = (op: OperationType) => isLoading?.(server.id!, op) ?? false;
  return (
    <div className="flex flex-wrap items-center gap-1">
      <Tooltip content="测试本地连接">
        <Button variant="ghost" size="sm" icon={<CheckCircle className="h-3.5 w-3.5" />}
          onClick={() => onTestConnection(server)} loading={loading('testConnection')} className="text-foreground-muted hover:text-indigo-400">测试</Button>
      </Tooltip>
      <Tooltip content="查看Dashboard认证信息">
        <Button variant="ghost" size="sm" icon={<Key className="h-3.5 w-3.5" />}
          onClick={() => onViewAuth(server)} className="text-foreground-muted hover:text-indigo-400">认证</Button>
      </Tooltip>
      <Tooltip content="刷新frps版本信息">
        <Button variant="ghost" size="sm" icon={<RefreshCw className="h-3.5 w-3.5" />}
          onClick={() => onRefreshLocalVersion(server.id!)} loading={loading('refreshLocalVersion')} className="text-foreground-muted hover:text-indigo-400">版本</Button>
      </Tooltip>
      {server.status === 'running' && (
        <Tooltip content="查看服务器运行指标">
          <Button variant="ghost" size="sm" icon={<BarChart3 className="h-3.5 w-3.5" />}
            onClick={() => onViewMetrics(server.id!)} className="text-foreground-muted hover:text-indigo-400">指标</Button>
        </Tooltip>
      )}
    </div>
  );
}