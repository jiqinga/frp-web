import { useState, useEffect } from 'react';
import {
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  Area,
  ComposedChart,
  PieChart,
  Pie,
  Cell,
} from 'recharts';
import { ArrowUp, ArrowDown, LineChart, BarChart3 } from 'lucide-react';
import { Modal, Spinner, Button, Tabs } from './ui';
import { toast } from './ui/Toast';
import { trafficApi } from '../api/traffic';
import { formatBytes } from '../utils/websocket';
import { useThemeStore } from '../store/theme';
import { getChartTheme } from '../utils/chartTheme';
import type { TrafficStats, ProxyTrafficSummary } from '../types';

interface TrafficHistoryModalProps {
  visible: boolean;
  onClose: () => void;
  proxyId: number | null;
  proxyName: string;
}

// Tab 类型
type TabType = 'history' | 'summary';

// 时间范围预设 - 流量历史
const rangePresets = [
  { label: '最近1小时', hours: 1 },
  { label: '最近6小时', hours: 6 },
  { label: '最近24小时', hours: 24 },
  { label: '最近7天', hours: 24 * 7 },
  { label: '最近30天', hours: 24 * 30 },
];

// 时间范围预设 - 流量汇总
const summaryPresets = [
  { label: '最近1小时', hours: 1 },
  { label: '最近3小时', hours: 3 },
  { label: '最近6小时', hours: 6 },
  { label: '最近1天', hours: 24 },
  { label: '最近3天', hours: 24 * 3 },
  { label: '最近7天', hours: 24 * 7 },
  { label: '最近30天', hours: 24 * 30 },
];

// 格式化日期为带时区偏移的 ISO 字符串
// 例如: 2025-12-09T16:25:23+08:00
const formatDateToISO = (date: Date): string => {
  const tzOffset = -date.getTimezoneOffset();
  const sign = tzOffset >= 0 ? '+' : '-';
  const hours = String(Math.floor(Math.abs(tzOffset) / 60)).padStart(2, '0');
  const minutes = String(Math.abs(tzOffset) % 60).padStart(2, '0');
  
  return date.getFullYear() + '-' +
    String(date.getMonth() + 1).padStart(2, '0') + '-' +
    String(date.getDate()).padStart(2, '0') + 'T' +
    String(date.getHours()).padStart(2, '0') + ':' +
    String(date.getMinutes()).padStart(2, '0') + ':' +
    String(date.getSeconds()).padStart(2, '0') +
    sign + hours + ':' + minutes;
};

// 格式化日期显示
const formatDateDisplay = (dateStr: string): string => {
  const date = new Date(dateStr);
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
};

export function TrafficHistoryModal({ visible, onClose, proxyId, proxyName }: TrafficHistoryModalProps) {
  const { theme } = useThemeStore();
  const isLight = theme === 'light';
  const chartTheme = getChartTheme(isLight);
  const [activeTab, setActiveTab] = useState<TabType>('history');
  const [loading, setLoading] = useState(false);
  const [trafficData, setTrafficData] = useState<TrafficStats[]>([]);
  const [selectedRange, setSelectedRange] = useState(24); // 默认24小时
  const [summaryHours, setSummaryHours] = useState(24);
  const [trafficSummary, setTrafficSummary] = useState<ProxyTrafficSummary | null>(null);

  // 获取流量历史数据
  useEffect(() => {
    if (visible && proxyId && activeTab === 'history') {
      fetchTrafficHistory();
    }
  }, [visible, proxyId, selectedRange, activeTab]);

  // 获取流量汇总数据
  useEffect(() => {
    if (visible && proxyId && activeTab === 'summary') {
      fetchTrafficSummary();
    }
  }, [visible, proxyId, summaryHours, activeTab]);

  const fetchTrafficHistory = async () => {
    if (!proxyId) return;
    
    setLoading(true);
    try {
      const end = new Date();
      const start = new Date(end.getTime() - selectedRange * 60 * 60 * 1000);
      const data = await trafficApi.getHistory(proxyId, formatDateToISO(start), formatDateToISO(end));
      setTrafficData(data || []);
    } catch {
      toast.error('获取流量历史失败');
      setTrafficData([]);
    } finally {
      setLoading(false);
    }
  };

  const fetchTrafficSummary = async () => {
    if (!proxyId) return;
    setLoading(true);
    try {
      const data = await trafficApi.getProxiesTrafficSummary(summaryHours);
      setTrafficSummary(data[String(proxyId)] || null);
    } catch {
      toast.error('获取流量汇总失败');
      setTrafficSummary(null);
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setTrafficData([]);
    setTrafficSummary(null);
    onClose();
  };

  // 转换数据格式
  const chartData = trafficData.map(item => ({
    time: formatDateDisplay(item.record_time),
    inRate: item.current_rate_in,
    outRate: item.current_rate_out,
    bytesIn: item.bytes_in,
    bytesOut: item.bytes_out,
  }));

  return (
    <Modal
      open={visible}
      onClose={handleClose}
      title={`流量管理 - ${proxyName}`}
      size="lg"
    >
      <div className="space-y-4">
        {/* Tab 切换 */}
        <Tabs
          items={[
            { key: 'history', label: '流量历史', icon: <LineChart className="h-4 w-4" />, children: null },
            { key: 'summary', label: '流量汇总', icon: <BarChart3 className="h-4 w-4" />, children: null },
          ]}
          activeKey={activeTab}
          onChange={(key) => setActiveTab(key as TabType)}
          variant="line"
        />

        {/* 流量历史 Tab */}
        {activeTab === 'history' && (
          <>
            {/* 时间范围选择 */}
            <div className="flex flex-wrap gap-2">
              {rangePresets.map(preset => (
                <Button
                  key={preset.hours}
                  variant={selectedRange === preset.hours ? 'primary' : 'outline'}
                  size="sm"
                  onClick={() => setSelectedRange(preset.hours)}
                >
                  {preset.label}
                </Button>
              ))}
            </div>
            
            {/* 图表内容 */}
        {loading ? (
          <div className="flex h-96 items-center justify-center">
            <Spinner size="lg" />
          </div>
        ) : trafficData.length > 0 ? (
          <div className="h-96">
            <ResponsiveContainer width="100%" height="100%">
              <ComposedChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" stroke={chartTheme.grid} />
                <XAxis
                  dataKey="time"
                  stroke={chartTheme.axis}
                  tick={{ fill: chartTheme.tick, fontSize: 10 }}
                  tickLine={{ stroke: chartTheme.tickLine }}
                  angle={-45}
                  textAnchor="end"
                  height={60}
                />
                <YAxis
                  yAxisId="rate"
                  stroke={chartTheme.axis}
                  tick={{ fill: chartTheme.tick, fontSize: 10 }}
                  tickLine={{ stroke: chartTheme.tickLine }}
                  tickFormatter={(v) => `${formatBytes(v)}/s`}
                  orientation="left"
                />
                <YAxis
                  yAxisId="total"
                  stroke={chartTheme.axis}
                  tick={{ fill: chartTheme.tick, fontSize: 10 }}
                  tickLine={{ stroke: chartTheme.tickLine }}
                  tickFormatter={(v) => formatBytes(v)}
                  orientation="right"
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: chartTheme.tooltipBg,
                    border: `1px solid ${chartTheme.tooltipBorder}`,
                    borderRadius: '8px',
                  }}
                  labelStyle={{ color: chartTheme.tooltipText }}
                  formatter={(value: number, name: string) => {
                    const isRate = name === 'inRate' || name === 'outRate';
                    const label = {
                      inRate: '上传速率',
                      outRate: '下载速率',
                      bytesIn: '累计上传',
                      bytesOut: '累计下载',
                    }[name] || name;
                    return [isRate ? `${formatBytes(value)}/s` : formatBytes(value), label];
                  }}
                />
                <Legend
                  formatter={(value) => {
                    const labels: Record<string, string> = {
                      inRate: '上传速率',
                      outRate: '下载速率',
                      bytesIn: '累计上传',
                      bytesOut: '累计下载',
                    };
                    return <span className="text-foreground-secondary">{labels[value] || value}</span>;
                  }}
                />
                <Area
                  yAxisId="rate"
                  type="monotone"
                  dataKey="inRate"
                  stroke={chartTheme.success}
                  fill={chartTheme.success}
                  fillOpacity={0.2}
                  strokeWidth={2}
                />
                <Area
                  yAxisId="rate"
                  type="monotone"
                  dataKey="outRate"
                  stroke={chartTheme.info}
                  fill={chartTheme.info}
                  fillOpacity={0.2}
                  strokeWidth={2}
                />
                <Line
                  yAxisId="total"
                  type="monotone"
                  dataKey="bytesIn"
                  stroke={chartTheme.warning}
                  strokeWidth={2}
                  strokeDasharray="5 5"
                  dot={false}
                />
                <Line
                  yAxisId="total"
                  type="monotone"
                  dataKey="bytesOut"
                  stroke={chartTheme.purple}
                  strokeWidth={2}
                  strokeDasharray="5 5"
                  dot={false}
                />
              </ComposedChart>
            </ResponsiveContainer>
          </div>
            ) : (
              <div className="flex h-96 flex-col items-center justify-center text-foreground-muted">
                <svg className="mb-4 h-16 w-16" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
                <span>暂无流量数据</span>
              </div>
            )}
          </>
        )}

        {/* 流量汇总 Tab */}
        {activeTab === 'summary' && (
          <>
            {/* 时间范围选择 */}
            <div className="flex flex-wrap gap-2">
              {summaryPresets.map(preset => (
                <Button
                  key={preset.hours}
                  variant={summaryHours === preset.hours ? 'primary' : 'outline'}
                  size="sm"
                  onClick={() => setSummaryHours(preset.hours)}
                >
                  {preset.label}
                </Button>
              ))}
            </div>

            {/* 汇总内容 */}
            {loading ? (
              <div className="flex h-48 items-center justify-center">
                <Spinner size="lg" />
              </div>
            ) : trafficSummary ? (
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="rounded-lg p-6 border bg-surface border-border">
                    <div className="flex items-center gap-2 text-emerald-400 mb-2">
                      <ArrowUp className="h-5 w-5" />
                      <span className="text-sm font-medium">上传流量</span>
                    </div>
                    <div className="text-2xl font-bold text-foreground">
                      {formatBytes(trafficSummary.total_out)}
                    </div>
                  </div>
                  <div className="rounded-lg p-6 border bg-surface border-border">
                    <div className="flex items-center gap-2 text-blue-400 mb-2">
                      <ArrowDown className="h-5 w-5" />
                      <span className="text-sm font-medium">下载流量</span>
                    </div>
                    <div className="text-2xl font-bold text-foreground">
                      {formatBytes(trafficSummary.total_in)}
                    </div>
                  </div>
                </div>
                
                {/* 环形图 */}
                {(trafficSummary.total_in > 0 || trafficSummary.total_out > 0) && (
                  <div className="rounded-lg p-4 border bg-surface border-border">
                    <div className="h-48">
                      <ResponsiveContainer width="100%" height="100%">
                        <PieChart>
                          <Pie
                            data={[
                              { name: '上传', value: trafficSummary.total_out, color: chartTheme.success },
                              { name: '下载', value: trafficSummary.total_in, color: chartTheme.info },
                            ]}
                            cx="50%"
                            cy="50%"
                            innerRadius={50}
                            outerRadius={70}
                            dataKey="value"
                            label={({ cx, cy, midAngle, outerRadius, name, percent }) => {
                              const RADIAN = Math.PI / 180;
                              const radius = outerRadius + 25;
                              const x = cx + radius * Math.cos(-midAngle * RADIAN);
                              const y = cy + radius * Math.sin(-midAngle * RADIAN);
                              return (
                                <text x={x} y={y} fill={chartTheme.tooltipText} textAnchor={x > cx ? 'start' : 'end'} dominantBaseline="central" fontSize={12}>
                                  {`${name} ${(percent * 100).toFixed(1)}%`}
                                </text>
                              );
                            }}
                            labelLine={{ stroke: chartTheme.tick }}
                          >
                            <Cell fill={chartTheme.success} />
                            <Cell fill={chartTheme.info} />
                          </Pie>
                          <Tooltip
                            contentStyle={{
                              backgroundColor: chartTheme.tooltipBg,
                              border: `1px solid ${chartTheme.tooltipBorder}`,
                              borderRadius: '8px',
                            }}
                            itemStyle={{ color: chartTheme.tooltipText }}
                            formatter={(value: number) => formatBytes(value)}
                          />
                        </PieChart>
                      </ResponsiveContainer>
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <div className="flex h-48 flex-col items-center justify-center text-foreground-muted">
                <BarChart3 className="mb-4 h-16 w-16" />
                <span>暂无流量数据</span>
              </div>
            )}
          </>
        )}
      </div>
    </Modal>
  );
}