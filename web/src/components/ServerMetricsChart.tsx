import { useEffect, useState } from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { Spinner, Tabs } from './ui';
import { frpServerApi, type ServerMetricsHistory } from '../api/frpServer';
import { useThemeStore } from '../store/theme';
import { getChartTheme } from '../utils/chartTheme';
import { formatBytes } from '../utils/websocket';

interface Props {
  serverId: number;
}

const timeRangeOptions = [
  { key: '1', value: 1, label: '最近24小时' },
  { key: '3', value: 3, label: '最近3天' },
  { key: '7', value: 7, label: '最近7天' },
];

export default function ServerMetricsChart({ serverId }: Props) {
  const { theme } = useThemeStore();
  const isLight = theme === 'light';
  const chartTheme = getChartTheme(isLight);
  const [days, setDays] = useState(1);
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<ServerMetricsHistory[]>([]);

  useEffect(() => {
    const loadData = async () => {
      setLoading(true);
      try {
        const res = await frpServerApi.getMetricsHistory(serverId, days);
        setData(res || []);
      } catch {
        setData([]);
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, [serverId, days]);

  // 转换数据格式
  const chartData = data.map(d => ({
    time: new Date(d.record_time).toLocaleString('zh-CN', { 
      month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' 
    }),
    cpu: parseFloat(d.cpu_percent.toFixed(2)),
    memory: parseFloat((d.memory_bytes / 1024 / 1024).toFixed(2)),
    trafficIn: d.traffic_in,
    trafficOut: d.traffic_out,
  }));

  const tabItems = timeRangeOptions.map(option => ({
    key: option.key,
    label: option.label,
    children: null,
  }));

  return (
    <div className="space-y-4">
      {/* 时间范围选择 */}
      <div className="flex justify-center">
        <Tabs
          items={tabItems}
          activeKey={String(days)}
          onChange={(key) => setDays(Number(key))}
          variant="pills"
        />
      </div>

      {/* 图表内容 */}
      {loading ? (
        <div className="flex h-96 items-center justify-center">
          <Spinner size="lg" />
        </div>
      ) : data.length === 0 ? (
        <div className="flex h-96 flex-col items-center justify-center text-foreground-muted">
          <svg className="mb-4 h-16 w-16" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
          <span>暂无历史数据</span>
        </div>
      ) : (
        <div className="space-y-6">
          {/* CPU 和内存合并图表 */}
          <div>
            <h4 className="mb-2 text-sm font-medium text-foreground-secondary">CPU 使用率 / 内存占用</h4>
            <div className="h-40">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke={chartTheme.grid} />
                  <XAxis
                    dataKey="time"
                    stroke={chartTheme.axis}
                    tick={{ fill: chartTheme.tick, fontSize: 10 }}
                    tickLine={{ stroke: chartTheme.tickLine }}
                  />
                  <YAxis
                    yAxisId="left"
                    stroke={chartTheme.warning}
                    tick={{ fill: chartTheme.tick, fontSize: 10 }}
                    tickLine={{ stroke: chartTheme.tickLine }}
                    domain={[0, 100]}
                    tickFormatter={(v) => `${v}%`}
                  />
                  <YAxis
                    yAxisId="right"
                    orientation="right"
                    stroke={chartTheme.purple}
                    tick={{ fill: chartTheme.tick, fontSize: 10 }}
                    tickLine={{ stroke: chartTheme.tickLine }}
                    tickFormatter={(v) => `${v}MB`}
                  />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: chartTheme.tooltipBg,
                      border: `1px solid ${chartTheme.tooltipBorder}`,
                      borderRadius: '8px',
                    }}
                    labelStyle={{ color: chartTheme.tooltipText }}
                    formatter={(value: number, name: string) => [
                      name === 'cpu' ? `${value}%` : `${value} MB`,
                      name === 'cpu' ? 'CPU' : '内存'
                    ]}
                  />
                  <Legend
                    formatter={(value) => (
                      <span className="text-foreground-secondary">
                        {value === 'cpu' ? 'CPU 使用率' : '内存占用'}
                      </span>
                    )}
                  />
                  <Line
                    yAxisId="left"
                    type="monotone"
                    dataKey="cpu"
                    stroke={chartTheme.warning}
                    strokeWidth={2}
                    dot={false}
                  />
                  <Line
                    yAxisId="right"
                    type="monotone"
                    dataKey="memory"
                    stroke={chartTheme.purple}
                    strokeWidth={2}
                    dot={false}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </div>

          {/* 流量图表 */}
          <div>
            <h4 className="mb-2 text-sm font-medium text-foreground-secondary">流量统计</h4>
            <div className="h-32">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke={chartTheme.grid} />
                  <XAxis
                    dataKey="time"
                    stroke={chartTheme.axis}
                    tick={{ fill: chartTheme.tick, fontSize: 10 }}
                    tickLine={{ stroke: chartTheme.tickLine }}
                  />
                  <YAxis
                    stroke={chartTheme.axis}
                    tick={{ fill: chartTheme.tick, fontSize: 10 }}
                    tickLine={{ stroke: chartTheme.tickLine }}
                    tickFormatter={(v) => formatBytes(v)}
                  />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: chartTheme.tooltipBg,
                      border: `1px solid ${chartTheme.tooltipBorder}`,
                      borderRadius: '8px',
                    }}
                    labelStyle={{ color: chartTheme.tooltipText }}
                    formatter={(value: number, name: string) => [
                      formatBytes(value),
                      name === 'trafficIn' ? '入站' : '出站'
                    ]}
                  />
                  <Legend
                    formatter={(value) => (
                      <span className="text-foreground-secondary">
                        {value === 'trafficIn' ? '入站流量' : '出站流量'}
                      </span>
                    )}
                  />
                  <Line
                    type="monotone"
                    dataKey="trafficIn"
                    stroke={chartTheme.success}
                    strokeWidth={2}
                    dot={false}
                  />
                  <Line
                    type="monotone"
                    dataKey="trafficOut"
                    stroke={chartTheme.info}
                    strokeWidth={2}
                    dot={false}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}