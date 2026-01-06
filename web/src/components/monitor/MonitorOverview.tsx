import { ArrowUpFromLine, ArrowDownToLine, CheckCircle, Wifi, WifiOff } from 'lucide-react';
import { formatBytes } from '../../utils/websocket';
import { cn } from '../../utils/cn';

interface MonitorOverviewProps {
  totalInRate: number;
  totalOutRate: number;
  onlineCount: number;
  totalCount: number;
  connected: boolean;
}

interface StatCardProps {
  title: string;
  value: string;
  suffix?: string;
  icon: React.ReactNode;
  iconColor: string;
  valueColor?: string;
  extra?: React.ReactNode;
}

const StatCard = ({ title, value, suffix, icon, iconColor, valueColor, extra }: StatCardProps) => {
  return (
    <div className="relative overflow-hidden rounded-xl border border-border bg-surface/80 p-4 backdrop-blur-sm">
      <div className={cn("absolute -right-4 -top-4 h-20 w-20 rounded-full opacity-20 blur-2xl", iconColor)} />
      
      <div className="relative">
        <div className="flex items-center justify-between">
          <span className="text-sm text-foreground-muted">{title}</span>
          <div className={cn("flex h-8 w-8 items-center justify-center rounded-lg", iconColor, "bg-opacity-20")}>
            {icon}
          </div>
        </div>
        <div className="mt-2">
          <span className={cn("text-2xl font-bold", valueColor || "text-foreground")}>
            {value}
          </span>
          {suffix && <span className="ml-1 text-sm text-foreground-muted">{suffix}</span>}
        </div>
        {extra && <div className="mt-2">{extra}</div>}
      </div>
    </div>
  );
};

export function MonitorOverview({ totalInRate, totalOutRate, onlineCount, totalCount, connected }: MonitorOverviewProps) {
  const onlinePercent = totalCount > 0 ? ((onlineCount / totalCount) * 100).toFixed(1) : '0';

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <StatCard
        title="上传速率"
        value={formatBytes(totalInRate)}
        suffix="/s"
        icon={<ArrowUpFromLine className="h-4 w-4 text-green-400" />}
        iconColor="bg-green-500"
        valueColor="text-green-400"
      />
      
      <StatCard
        title="下载速率"
        value={formatBytes(totalOutRate)}
        suffix="/s"
        icon={<ArrowDownToLine className="h-4 w-4 text-blue-400" />}
        iconColor="bg-blue-500"
        valueColor="text-blue-400"
      />
      
      <StatCard
        title="在线代理"
        value={String(onlineCount)}
        suffix={`/ ${totalCount}`}
        icon={<CheckCircle className="h-4 w-4 text-emerald-400" />}
        iconColor="bg-emerald-500"
        valueColor={onlineCount > 0 ? "text-emerald-400" : "text-slate-400"}
        extra={
          <div className="text-xs text-foreground-muted">
            在线率: {onlinePercent}%
          </div>
        }
      />
      
      <StatCard
        title="连接状态"
        value={connected ? '已连接' : '未连接'}
        icon={connected ? <Wifi className="h-4 w-4 text-green-400" /> : <WifiOff className="h-4 w-4 text-red-400" />}
        iconColor={connected ? "bg-green-500" : "bg-red-500"}
        valueColor={connected ? "text-green-400" : "text-red-400"}
        extra={
          <div className="flex items-center gap-2">
            <div className={cn(
              "h-2 w-2 rounded-full",
              connected ? "bg-green-500 animate-pulse" : "bg-red-500"
            )} />
            <span className="text-xs text-foreground-muted">
              {connected ? '实时更新中' : '等待连接'}
            </span>
          </div>
        }
      />
    </div>
  );
}