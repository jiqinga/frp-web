import { Server, RefreshCw } from 'lucide-react';
import type { FrpServer } from '../../../api/frpServer';
import { StatCard, type StatCardColor } from '../../../components/ui';

interface StatsCardsProps {
  servers: FrpServer[];
}

export function StatsCards({ servers }: StatsCardsProps) {
  const stats: { label: string; value: number; icon: React.ReactNode; color: StatCardColor }[] = [
    { label: '服务器总数', value: servers.length, icon: <Server className="h-6 w-6 text-indigo-400" />, color: 'indigo' },
    { label: '本地服务器', value: servers.filter(s => s.server_type === 'local').length, icon: <Server className="h-6 w-6 text-emerald-400" />, color: 'emerald' },
    { label: '远程服务器', value: servers.filter(s => s.server_type === 'remote').length, icon: <Server className="h-6 w-6 text-cyan-400" />, color: 'cyan' },
    { label: '运行中', value: servers.filter(s => s.status === 'running').length, icon: <RefreshCw className="h-6 w-6 text-green-400" />, color: 'green' },
  ];

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      {stats.map(({ label, value, icon, color }) => (
        <StatCard key={label} title={label} value={value} icon={icon} color={color} />
      ))}
    </div>
  );
}