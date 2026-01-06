import { Cloud, Monitor } from 'lucide-react';
import { Badge } from '../../../components/ui/Badge';

export function getStatusBadge(status?: string) {
  const statusMap: Record<string, { variant: 'success' | 'default' | 'primary' | 'warning' | 'danger'; text: string; pulse: boolean }> = {
    running: { variant: 'success', text: '运行中', pulse: true },
    stopped: { variant: 'default', text: '已停止', pulse: false },
    starting: { variant: 'primary', text: '启动中', pulse: true },
    stopping: { variant: 'warning', text: '停止中', pulse: true },
    error: { variant: 'danger', text: '错误', pulse: false },
  };
  const config = statusMap[status || 'stopped'] || statusMap.stopped;
  return <Badge variant={config.variant} dot pulse={config.pulse}>{config.text}</Badge>;
}

export function getTypeBadge(serverType?: string) {
  if (serverType === 'remote') {
    return (
      <div className="inline-flex items-center gap-1.5 px-2 py-1 rounded-full bg-blue-500/20 border border-blue-500/30 whitespace-nowrap">
        <Cloud className="h-3.5 w-3.5 text-blue-400 flex-shrink-0" />
        <span className="text-xs font-medium text-blue-400">远程</span>
      </div>
    );
  }
  return (
    <div className="inline-flex items-center gap-1.5 px-2 py-1 rounded-full bg-emerald-500/20 border border-emerald-500/30 whitespace-nowrap">
      <Monitor className="h-3.5 w-3.5 text-emerald-400 flex-shrink-0" />
      <span className="text-xs font-medium text-emerald-400">本地</span>
    </div>
  );
}