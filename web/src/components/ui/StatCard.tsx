import { ArrowUpRight, ArrowDownRight } from 'lucide-react';
import { cn } from '../../utils/cn';

// 预定义颜色配置,避免 Tailwind 动态类名问题
const colorConfig = {
  indigo: { bg: 'bg-indigo-500', bgLight: 'bg-indigo-500/20', text: 'text-indigo-400' },
  cyan: { bg: 'bg-cyan-500', bgLight: 'bg-cyan-500/20', text: 'text-cyan-400' },
  green: { bg: 'bg-green-500', bgLight: 'bg-green-500/20', text: 'text-green-400' },
  purple: { bg: 'bg-purple-500', bgLight: 'bg-purple-500/20', text: 'text-purple-400' },
  yellow: { bg: 'bg-yellow-500', bgLight: 'bg-yellow-500/20', text: 'text-yellow-400' },
  red: { bg: 'bg-red-500', bgLight: 'bg-red-500/20', text: 'text-red-400' },
  blue: { bg: 'bg-blue-500', bgLight: 'bg-blue-500/20', text: 'text-blue-400' },
  emerald: { bg: 'bg-emerald-500', bgLight: 'bg-emerald-500/20', text: 'text-emerald-400' },
} as const;

export type StatCardColor = keyof typeof colorConfig;

export interface StatCardProps {
  title: string;
  value: number | string;
  icon: React.ReactNode;
  color?: StatCardColor;
  trend?: { value: number; isUp: boolean };
  loading?: boolean;
  glow?: boolean;
  className?: string;
}

export function StatCard({
  title,
  value,
  icon,
  color = 'indigo',
  trend,
  loading,
  glow = true,
  className,
}: StatCardProps) {
  const colors = colorConfig[color];

  return (
    <div className={cn(
      "relative overflow-hidden rounded-xl border border-border bg-surface/80 p-6 backdrop-blur-sm transition-all duration-300 hover:border-border-hover hover:bg-surface",
      className
    )}>
      {glow && (
        <div className={cn("absolute -right-8 -top-8 h-32 w-32 rounded-full opacity-20 blur-3xl", colors.bg)} />
      )}
      
      <div className="relative">
        <div className="flex items-center justify-between">
          <div className={cn("flex h-12 w-12 items-center justify-center rounded-xl", colors.bgLight)}>
            {icon}
          </div>
          {trend && (
            <div className={cn(
              "flex items-center gap-1 rounded-full px-2 py-1 text-xs font-medium",
              trend.isUp ? "bg-green-500/20 text-green-400" : "bg-red-500/20 text-red-400"
            )}>
              {trend.isUp ? <ArrowUpRight className="h-3 w-3" /> : <ArrowDownRight className="h-3 w-3" />}
              {trend.value}%
            </div>
          )}
        </div>
        
        <div className="mt-4">
          <p className="text-sm text-foreground-muted">{title}</p>
          {loading ? (
            <div className="mt-2 h-8 w-20 animate-pulse rounded bg-surface-hover" />
          ) : (
            <p className="mt-1 text-3xl font-bold text-foreground">{value}</p>
          )}
        </div>
      </div>
    </div>
  );
}