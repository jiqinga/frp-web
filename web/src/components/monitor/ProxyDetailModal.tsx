import { useMemo } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { Activity, ArrowUpFromLine, ArrowDownToLine, Clock } from 'lucide-react';
import { Modal, Badge } from '../ui';
import { formatBytes } from '../../utils/websocket';
import { useThemeStore } from '../../store/theme';
import { getChartTheme } from '../../utils/chartTheme';
import type { TrafficData, ProxyHistory } from '../../hooks/useRealtimeMonitor';

interface ProxyDetailModalProps {
  open: boolean;
  onClose: () => void;
  proxy: TrafficData | null;
  history: ProxyHistory[];
}

export function ProxyDetailModal({ open, onClose, proxy, history }: ProxyDetailModalProps) {
  const { theme } = useThemeStore();
  const chartTheme = getChartTheme(theme === 'light');

  const chartData = useMemo(() => {
    return history.map(item => ({
      time: item.time,
      upload: item.inRate,
      download: item.outRate,
    }));
  }, [history]);

  if (!proxy) return null;

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={
        <div className="flex items-center gap-3">
          <Activity className="h-5 w-5 text-indigo-400" />
          <span>{proxy.proxy_name}</span>
          <Badge variant={proxy.online ? 'success' : 'default'}>
            {proxy.online ? '在线' : '离线'}
          </Badge>
        </div>
      }
      size="lg"
    >
      <div className="space-y-6">
        {/* 实时统计 */}
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          <div className="rounded-lg border border-border bg-surface/80 p-4">
            <div className="flex items-center gap-2 text-sm text-foreground-muted">
              <ArrowUpFromLine className="h-4 w-4 text-green-400" />
              上传速率
            </div>
            <div className="mt-1 text-xl font-semibold text-green-400">
              {formatBytes(proxy.bytes_in_rate)}/s
            </div>
          </div>
          <div className="rounded-lg border border-border bg-surface/80 p-4">
            <div className="flex items-center gap-2 text-sm text-foreground-muted">
              <ArrowDownToLine className="h-4 w-4 text-blue-400" />
              下载速率
            </div>
            <div className="mt-1 text-xl font-semibold text-blue-400">
              {formatBytes(proxy.bytes_out_rate)}/s
            </div>
          </div>
          <div className="rounded-lg border border-border bg-surface/80 p-4">
            <div className="flex items-center gap-2 text-sm text-foreground-muted">
              <ArrowUpFromLine className="h-4 w-4 text-emerald-400" />
              总上传
            </div>
            <div className="mt-1 text-xl font-semibold text-emerald-400">
              {formatBytes(proxy.total_bytes_in)}
            </div>
          </div>
          <div className="rounded-lg border border-border bg-surface/80 p-4">
            <div className="flex items-center gap-2 text-sm text-foreground-muted">
              <ArrowDownToLine className="h-4 w-4 text-cyan-400" />
              总下载
            </div>
            <div className="mt-1 text-xl font-semibold text-cyan-400">
              {formatBytes(proxy.total_bytes_out)}
            </div>
          </div>
        </div>

        {/* 流量趋势图 */}
        <div className="rounded-lg border border-border bg-surface-hover/30 p-4">
          <div className="mb-4 flex items-center gap-2 text-sm font-medium text-foreground-secondary">
            <Clock className="h-4 w-4 text-indigo-400" />
            流量趋势（最近 60 秒）
          </div>
          <div className="h-64">
            {chartData.length > 0 ? (
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke={chartTheme.grid} />
                  <XAxis
                    dataKey="time"
                    stroke={chartTheme.axis}
                    tick={{ fill: chartTheme.tick, fontSize: 11 }}
                    tickLine={{ stroke: chartTheme.tickLine }}
                  />
                  <YAxis
                    stroke={chartTheme.axis}
                    tick={{ fill: chartTheme.tick, fontSize: 11 }}
                    tickLine={{ stroke: chartTheme.tickLine }}
                    tickFormatter={(value) => formatBytes(value)}
                  />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: chartTheme.tooltipBg,
                      border: `1px solid ${chartTheme.tooltipBorder}`,
                      borderRadius: '8px',
                      boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.3)',
                    }}
                    labelStyle={{ color: chartTheme.tooltipText }}
                    formatter={(value: number, name: string) => [
                      `${formatBytes(value)}/s`,
                      name === 'upload' ? '上传' : '下载'
                    ]}
                  />
                  <Legend
                    formatter={(value) => (
                      <span className="text-foreground-secondary">
                        {value === 'upload' ? '上传' : '下载'}
                      </span>
                    )}
                  />
                  <Line
                    type="monotone"
                    dataKey="upload"
                    stroke={chartTheme.success}
                    strokeWidth={2}
                    dot={false}
                    activeDot={{ r: 4, fill: chartTheme.success }}
                  />
                  <Line
                    type="monotone"
                    dataKey="download"
                    stroke={chartTheme.info}
                    strokeWidth={2}
                    dot={false}
                    activeDot={{ r: 4, fill: chartTheme.info }}
                  />
                </LineChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex h-full items-center justify-center text-foreground-muted">
                暂无历史数据
              </div>
            )}
          </div>
        </div>

        {/* 代理信息 */}
        <div className="rounded-lg border border-border bg-surface-hover/30 p-4">
          <div className="mb-3 text-sm font-medium text-foreground-secondary">代理信息</div>
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span className="text-foreground-muted">代理名称：</span>
              <span className="ml-2 text-foreground">{proxy.proxy_name || '-'}</span>
            </div>
            <div>
              <span className="text-foreground-muted">客户端：</span>
              <span className="ml-2 text-foreground">{proxy.client_name || '-'}</span>
            </div>
            <div>
              <span className="text-foreground-muted">代理 ID：</span>
              <span className="ml-2 text-foreground">{proxy.proxy_id}</span>
            </div>
            <div>
              <span className="text-foreground-muted">客户端 ID：</span>
              <span className="ml-2 text-foreground">{proxy.client_id}</span>
            </div>
          </div>
        </div>
      </div>
    </Modal>
  );
}