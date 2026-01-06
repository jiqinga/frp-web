import {
  CheckCircle, Cloud, Play, Square, RotateCcw, Key,
  RefreshCw, ArrowUp, FileText, BarChart3,
} from 'lucide-react';
import { Button } from '../../../components/ui/Button';
import { Tooltip } from '../../../components/ui/Tooltip';
import type { FrpServer } from '../../../api/frpServer';
import type { OperationType } from '../hooks/useOperationLoading';

interface RemoteServerActionsProps {
  server: FrpServer;
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
  isLoading?: (serverId: number, operation: OperationType) => boolean;
}

export function RemoteServerActions({
  server,
  onTestSSH,
  onRemoteInstall,
  onRemoteStart,
  onRemoteStop,
  onRemoteRestart,
  onViewAuth,
  onRefreshRemoteVersion,
  onViewMetrics,
  onRemoteReinstall,
  onRemoteUpgrade,
  onViewLogs,
  isLoading,
}: RemoteServerActionsProps) {
  const loading = (op: OperationType) => isLoading?.(server.id!, op) ?? false;
  return (
    <div className="flex flex-wrap items-center gap-1">
      <Tooltip content="测试SSH连接">
        <Button variant="ghost" size="sm" icon={<CheckCircle className="h-3.5 w-3.5" />}
          onClick={() => onTestSSH(server.id!)} loading={loading('testSSH')} className="text-foreground-muted hover:text-indigo-400">SSH</Button>
      </Tooltip>

      {!server.binary_path && (
        <Tooltip content="在远程服务器上安装frps">
          <Button variant="ghost" size="sm" icon={<Cloud className="h-3.5 w-3.5" />}
            onClick={() => onRemoteInstall(server.id!)} className="text-foreground-muted hover:text-green-400">安装</Button>
        </Tooltip>
      )}

      {server.binary_path && server.status === 'stopped' && (
        <Tooltip content="启动远程frps服务">
          <Button variant="ghost" size="sm" icon={<Play className="h-3.5 w-3.5" />}
            onClick={() => onRemoteStart(server.id!)} loading={loading('start')} className="text-foreground-muted hover:text-green-400">启动</Button>
        </Tooltip>
      )}

      {server.status === 'running' && (
        <>
          <Tooltip content="停止远程frps服务">
            <Button variant="ghost" size="sm" icon={<Square className="h-3.5 w-3.5" />}
              onClick={() => onRemoteStop(server.id!)} loading={loading('stop')} className="text-foreground-muted hover:text-orange-400">停止</Button>
          </Tooltip>
          <Tooltip content="重启远程frps服务">
            <Button variant="ghost" size="sm" icon={<RotateCcw className="h-3.5 w-3.5" />}
              onClick={() => onRemoteRestart(server.id!)} loading={loading('restart')} className="text-foreground-muted hover:text-blue-400">重启</Button>
          </Tooltip>
        </>
      )}

      {server.binary_path && (
        <>
          <Tooltip content="查看Dashboard认证信息">
            <Button variant="ghost" size="sm" icon={<Key className="h-3.5 w-3.5" />}
              onClick={() => onViewAuth(server)} className="text-foreground-muted hover:text-indigo-400">认证</Button>
          </Tooltip>
          <Tooltip content="刷新frps版本信息">
            <Button variant="ghost" size="sm" icon={<RefreshCw className="h-3.5 w-3.5" />}
              onClick={() => onRefreshRemoteVersion(server.id!)} loading={loading('refreshVersion')} className="text-foreground-muted hover:text-indigo-400">版本</Button>
          </Tooltip>
          {server.status === 'running' && (
            <Tooltip content="查看服务器运行指标">
              <Button variant="ghost" size="sm" icon={<BarChart3 className="h-3.5 w-3.5" />}
                onClick={() => onViewMetrics(server.id!)} className="text-foreground-muted hover:text-indigo-400">指标</Button>
            </Tooltip>
          )}
          <Tooltip content="重新安装frps">
            <Button variant="ghost" size="sm" icon={<RotateCcw className="h-3.5 w-3.5" />}
              onClick={() => onRemoteReinstall(server.id!)} className="text-foreground-muted hover:text-yellow-400">重装</Button>
          </Tooltip>
          <Tooltip content="升级frps版本">
            <Button variant="ghost" size="sm" icon={<ArrowUp className="h-3.5 w-3.5" />}
              onClick={() => onRemoteUpgrade(server.id!)} className="text-foreground-muted hover:text-green-400">升级</Button>
          </Tooltip>
          <Tooltip content="查看服务器日志">
            <Button variant="ghost" size="sm" icon={<FileText className="h-3.5 w-3.5" />}
              onClick={() => onViewLogs(server.id!)} loading={loading('viewLogs')} className="text-foreground-muted hover:text-cyan-400">日志</Button>
          </Tooltip>
        </>
      )}
    </div>
  );
}