import { Shield, CheckCircle, AlertTriangle, XCircle, Clock, AlertOctagon } from 'lucide-react';
import { Card, CardContent } from '../../../components/ui';
import type { CertificateStats } from '../hooks';

interface CertificateStatsCardProps {
  stats: CertificateStats;
  className?: string;
}

export function CertificateStatsCard({ stats, className }: CertificateStatsCardProps) {
  const items = [
    { label: '总证书', value: stats.total, icon: Shield, color: 'text-blue-400' },
    { label: '有效', value: stats.active, icon: CheckCircle, color: 'text-green-400' },
    { label: '即将过期', value: stats.expiring, icon: AlertTriangle, color: 'text-yellow-400' },
    { label: '已过期', value: stats.expired, icon: XCircle, color: 'text-red-400' },
    { label: '申请中', value: stats.pending, icon: Clock, color: 'text-purple-400' },
    { label: '失败', value: stats.failed, icon: AlertOctagon, color: 'text-orange-400' },
  ];

  return (
    <Card className={className}>
      <CardContent className="p-4">
        <div className="grid grid-cols-3 gap-4 md:grid-cols-6">
          {items.map((item) => (
            <div key={item.label} className="flex flex-col items-center gap-1">
              <item.icon className={`h-5 w-5 ${item.color}`} />
              <span className="text-2xl font-bold text-foreground">{item.value}</span>
              <span className="text-xs text-foreground-muted">{item.label}</span>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}