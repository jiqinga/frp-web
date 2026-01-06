import { Bell, CheckCircle, AlertTriangle, Calendar } from 'lucide-react';
import { StatCard, type StatCardColor } from '../../../components/ui';
import { cn } from '../../../utils/cn';
import type { AlertStats } from '../hooks/useAlerts';

interface AlertStatsCardProps {
  stats: AlertStats;
  className?: string;
}

const statItems: { key: keyof AlertStats; label: string; icon: React.ReactNode; color: StatCardColor }[] = [
  { key: 'totalRules', label: '总规则数', icon: <Bell className="h-5 w-5 text-blue-400" />, color: 'blue' },
  { key: 'enabledRules', label: '启用规则', icon: <CheckCircle className="h-5 w-5 text-green-400" />, color: 'green' },
  { key: 'todayAlerts', label: '今日告警', icon: <AlertTriangle className="h-5 w-5 text-yellow-400" />, color: 'yellow' },
  { key: 'weekAlerts', label: '本周告警', icon: <Calendar className="h-5 w-5 text-purple-400" />, color: 'purple' },
];

export function AlertStatsCard({ stats, className }: AlertStatsCardProps) {
  return (
    <div className={cn("grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4", className)}>
      {statItems.map(({ key, label, icon, color }) => (
        <StatCard key={key} title={label} value={stats[key]} icon={icon} color={color} glow={false} />
      ))}
    </div>
  );
}