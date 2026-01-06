import { useState } from 'react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  BarChart,
  Bar,
  Legend
} from 'recharts';
import { Activity, BarChart3 } from 'lucide-react';
import { formatBytes } from '../../utils/websocket';
import { Card, CardHeader, CardContent, Tabs } from '../ui';
import { useThemeStore } from '../../store/theme';
import { getChartTheme, chartColors } from '../../utils/chartTheme';
import type { TrafficData } from '../../hooks/useRealtimeMonitor';

interface TrafficChartProps {
  chartHistory: { time: string; inRate: number; outRate: number }[];
  topProxies: TrafficData[];
}

type ChartMode = 'total' | 'top5';

// 自定义 Tooltip
const CustomTooltip = ({ active, payload, label }: { active?: boolean; payload?: Array<{ value: number; name: string; color: string }>; label?: string }) => {
  if (active && payload && payload.length) {
    return (
      <div className="rounded-lg border border-border bg-surface p-3 shadow-xl">
        <p className="mb-2 text-sm text-foreground-muted">{label}</p>
        {payload.map((entry, index) => (
          <p key={index} className="text-sm" style={{ color: entry.color }}>
            {entry.name}: {formatBytes(entry.value)}/s
          </p>
        ))}
      </div>
    );
  }
  return null;
};

// 自定义 Bar Tooltip
const BarTooltip = ({ active, payload }: { active?: boolean; payload?: Array<{ value: number; name: string; fill: string }> }) => {
  if (active && payload && payload.length) {
    return (
      <div className="rounded-lg border border-border bg-surface p-3 shadow-xl">
        {payload.map((entry, index) => (
          <p key={index} className="text-sm" style={{ color: entry.fill }}>
            {entry.name}: {formatBytes(entry.value)}/s
          </p>
        ))}
      </div>
    );
  }
  return null;
};

export function TrafficChart({ chartHistory, topProxies }: TrafficChartProps) {
  const { theme } = useThemeStore();
  const chartTheme = getChartTheme(theme === 'light');
  const [mode, setMode] = useState<ChartMode>('total');

  // 准备 Top5 数据
  const top5Data = topProxies.slice(0, 5).map(p => ({
    name: p.proxy_name,
    rate: p.bytes_in_rate + p.bytes_out_rate,
  }));

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Activity className="h-5 w-5 text-indigo-400" />
            <span>实时流量</span>
          </div>
          <Tabs
            items={[
              { key: 'total', label: '总流量', icon: <Activity className="h-4 w-4" />, children: null },
              { key: 'top5', label: 'Top 5', icon: <BarChart3 className="h-4 w-4" />, children: null },
            ]}
            activeKey={mode}
            onChange={(key) => setMode(key as ChartMode)}
            variant="pills"
          />
        </div>
      </CardHeader>
      <CardContent>
        <div className="h-72">
          {mode === 'total' ? (
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={chartHistory} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="inRateGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor={chartTheme.success} stopOpacity={chartTheme.gradientOpacityStart} />
                    <stop offset="95%" stopColor={chartTheme.success} stopOpacity={chartTheme.gradientOpacityEnd} />
                  </linearGradient>
                  <linearGradient id="outRateGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor={chartTheme.info} stopOpacity={chartTheme.gradientOpacityStart} />
                    <stop offset="95%" stopColor={chartTheme.info} stopOpacity={chartTheme.gradientOpacityEnd} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke={chartTheme.grid} vertical={false} />
                <XAxis
                  dataKey="time"
                  stroke={chartTheme.axis}
                  fontSize={12}
                  tickLine={false}
                  axisLine={false}
                />
                <YAxis
                  stroke={chartTheme.axis}
                  fontSize={12}
                  tickLine={false}
                  axisLine={false}
                  tickFormatter={(value) => formatBytes(value)}
                />
                <Tooltip content={<CustomTooltip />} />
                <Legend
                  wrapperStyle={{ paddingTop: 20 }}
                  formatter={(value) => <span className="text-foreground-secondary">{value}</span>}
                />
                <Area
                  type="monotone"
                  dataKey="inRate"
                  name="上传"
                  stroke={chartTheme.success}
                  strokeWidth={2}
                  fill="url(#inRateGradient)"
                />
                <Area
                  type="monotone"
                  dataKey="outRate"
                  name="下载"
                  stroke={chartTheme.info}
                  strokeWidth={2}
                  fill="url(#outRateGradient)"
                />
              </AreaChart>
            </ResponsiveContainer>
          ) : (
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={top5Data} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke={chartTheme.grid} vertical={false} />
                <XAxis
                  dataKey="name"
                  stroke={chartTheme.axis}
                  fontSize={12}
                  tickLine={false}
                  axisLine={false}
                  angle={-45}
                  textAnchor="end"
                  height={60}
                />
                <YAxis
                  stroke={chartTheme.axis}
                  fontSize={12}
                  tickLine={false}
                  axisLine={false}
                  tickFormatter={(value) => formatBytes(value)}
                />
                <Tooltip content={<BarTooltip />} />
                <Bar
                  dataKey="rate"
                  name="速率"
                  radius={[4, 4, 0, 0]}
                >
                  {top5Data.map((_, index) => (
                    <rect key={index} fill={chartColors[index % chartColors.length]} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>
      </CardContent>
    </Card>
  );
}