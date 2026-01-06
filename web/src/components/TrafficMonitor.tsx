import { useState, useEffect } from 'react';
import { ArrowDownToLine, ArrowUpFromLine, Activity, Zap } from 'lucide-react';
import { trafficApi } from '../api/traffic';
import type { TrafficSummary } from '../types';
import { cn } from '../utils/cn';
import { formatBytes } from '../utils/websocket';

const formatRate = (rate: number): string => {
  return `${formatBytes(rate)}/s`;
};

interface StatCardProps {
  title: string;
  value: string;
  icon: React.ReactNode;
  iconColor: string;
  trend?: 'up' | 'down';
}

const StatCard = ({ title, value, icon, iconColor }: StatCardProps) => {
  return (
    <div className={cn(
      "relative overflow-hidden rounded-xl border p-4 backdrop-blur-sm",
      "border-border bg-surface"
    )}>
      {/* 背景装饰 */}
      <div className={cn("absolute -right-4 -top-4 h-24 w-24 rounded-full opacity-10 blur-2xl", iconColor)} />
      
      <div className="relative flex items-start justify-between">
        <div>
          <p className="text-sm text-foreground-muted">{title}</p>
          <p className="mt-2 text-2xl font-bold text-foreground">{value}</p>
        </div>
        <div className={cn("flex h-10 w-10 items-center justify-center rounded-lg", iconColor.replace('bg-', 'bg-opacity-20 '))}>
          {icon}
        </div>
      </div>
      
      {/* 底部装饰线 */}
      <div className={cn("absolute bottom-0 left-0 h-0.5 w-full", iconColor, "opacity-50")} />
    </div>
  );
};

export const TrafficMonitor = () => {
  const [summary, setSummary] = useState<TrafficSummary>({
    total_bytes_in: 0,
    total_bytes_out: 0,
    current_rate_in: 0,
    current_rate_out: 0,
    active_proxies: 0,
    total_proxies: 0
  });

  useEffect(() => {
    const fetchSummary = async () => {
      try {
        const data = await trafficApi.getSummary();
        setSummary(data);
      } catch {
        // 静默处理，流量统计失败不影响用户体验
      }
    };

    fetchSummary();
    const interval = setInterval(fetchSummary, 3000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <StatCard
        title="总入站流量"
        value={formatBytes(summary.total_bytes_in)}
        icon={<ArrowDownToLine className="h-5 w-5 text-cyan-400" />}
        iconColor="bg-cyan-500"
      />
      <StatCard
        title="总出站流量"
        value={formatBytes(summary.total_bytes_out)}
        icon={<ArrowUpFromLine className="h-5 w-5 text-purple-400" />}
        iconColor="bg-purple-500"
      />
      <StatCard
        title="当前入站速率"
        value={formatRate(summary.current_rate_in)}
        icon={<Activity className="h-5 w-5 text-green-400" />}
        iconColor="bg-green-500"
      />
      <StatCard
        title="当前出站速率"
        value={formatRate(summary.current_rate_out)}
        icon={<Zap className="h-5 w-5 text-orange-400" />}
        iconColor="bg-orange-500"
      />
    </div>
  );
};