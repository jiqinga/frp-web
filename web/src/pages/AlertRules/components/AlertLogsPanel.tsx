import { AlertTriangle, Clock, Monitor, Server, Activity } from 'lucide-react';
import type { AlertLog, AlertTargetType } from '../../../api/alert';
import { Card, CardHeader, CardContent, Badge } from '../../../components/ui';

interface AlertLogsPanelProps {
  logs: AlertLog[];
  loading: boolean;
}

const formatTime = (dateStr: string) => {
  const date = new Date(dateStr);
  return date.toLocaleString('zh-CN');
};

const getTargetIcon = (type: AlertTargetType) => {
  if (type === 'frpc') return <Monitor className="h-4 w-4" />;
  if (type === 'frps') return <Server className="h-4 w-4" />;
  return <Activity className="h-4 w-4" />;
};

export function AlertLogsPanel({ logs, loading }: AlertLogsPanelProps) {
  if (loading) {
    return (
      <Card>
        <CardContent className="py-12 text-center text-foreground-muted">
          加载中...
        </CardContent>
      </Card>
    );
  }

  if (logs.length === 0) {
    return (
      <Card>
        <CardContent className="py-12 text-center text-foreground-subtle">
          暂无告警记录
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <AlertTriangle className="h-5 w-5 text-yellow-400" />
          <span>告警历史记录</span>
          <Badge variant="default">{logs.length} 条</Badge>
        </div>
      </CardHeader>
      <CardContent className="p-0">
        <div className="divide-y divide-border max-h-96 overflow-y-auto">
          {logs.map((log) => (
            <div key={log.id} className="p-4 transition-colors hover:bg-surface-hover">
              <div className="flex items-start gap-3">
                <div className="flex-shrink-0 mt-0.5 text-yellow-400">
                  {getTargetIcon(log.target_type || 'proxy')}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-foreground-secondary">{log.message}</p>
                  <div className="flex items-center gap-3 mt-1 text-xs text-foreground-muted">
                    <span className="flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      {formatTime(log.created_at)}
                    </span>
                    <Badge variant={log.notified ? 'success' : 'default'} className="text-xs">
                      {log.notified ? '已通知' : '未通知'}
                    </Badge>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}