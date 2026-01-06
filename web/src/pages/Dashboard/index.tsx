import { useState, useEffect } from 'react';
import { Monitor, Network, Activity, TrendingUp, Server, Wifi, BarChart3, AlertCircle } from 'lucide-react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { clientApi } from '../../api/client';
import { trafficApi } from '../../api/traffic';
import { proxyApi } from '../../api/proxy';
import { frpServerApi } from '../../api/frpServer';
import { TrafficMonitor } from '../../components/TrafficMonitor';
import { Card, CardHeader, CardContent, Badge, Table, Spinner, StatCard, ProgressRing } from '../../components/ui';
import { useThemeStore } from '../../store/theme';
import { getChartTheme } from '../../utils/chartTheme';
import type { Proxy, TrafficTrendPoint } from '../../types';

export function Component() {
  const [stats, setStats] = useState({ clients: 0, proxies: 0, online: 0, servers: 0 });
  const [topProxies, setTopProxies] = useState<Proxy[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [trafficHistory, setTrafficHistory] = useState<TrafficTrendPoint[]>([]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [clientRes, trafficRes, trendRes, proxiesRes, serversRes] = await Promise.all([
          clientApi.getClients({ page: 1, page_size: 1 }),
          trafficApi.getSummary(),
          trafficApi.getTrend(24),
          proxyApi.getAllProxies(),
          frpServerApi.getAll()
        ]);
        setStats({
          clients: clientRes.total,
          proxies: trafficRes.total_proxies,
          online: trafficRes.active_proxies,
          servers: serversRes.length
        });
        // 转换流量数据为MB单位
        const trendData = trendRes.map(item => ({
          ...item,
          inbound: Math.round(item.inbound / 1024 / 1024 * 100) / 100,
          outbound: Math.round(item.outbound / 1024 / 1024 * 100) / 100
        }));
        setTrafficHistory(trendData);
        // 按总流量排序取前10
        const sorted = [...proxiesRes].sort((a, b) =>
          (b.total_bytes_in + b.total_bytes_out) - (a.total_bytes_in + a.total_bytes_out)
        ).slice(0, 10);
        setTopProxies(sorted);
      } catch (err) {
        console.error('获取仪表盘数据失败:', err);
        setError('获取数据失败，请稍后重试');
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  const onlineRate = stats.proxies ? (stats.online / stats.proxies) * 100 : 0;

  const { theme } = useThemeStore();
  const chartTheme = getChartTheme(theme === 'light');

  const columns = [
    {
      key: 'name',
      title: '代理名称',
      render: (_: unknown, record: Proxy) => (
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-500/20">
            <Network className="h-4 w-4 text-indigo-400" />
          </div>
          <span className="font-medium text-foreground">{record.name}</span>
        </div>
      )
    },
    {
      key: 'type',
      title: '类型',
      render: (_: unknown, record: Proxy) => (
        <Badge variant={record.type === 'tcp' ? 'primary' : record.type === 'http' ? 'success' : 'warning'}>
          {record.type.toUpperCase()}
        </Badge>
      )
    },
    {
      key: 'total_bytes_out',
      title: '上传流量',
      render: (_: unknown, record: Proxy) => (
        <span className="text-foreground-secondary">
          {((record.total_bytes_out || 0) / 1024 / 1024).toFixed(2)} MB
        </span>
      )
    },
    {
      key: 'total_bytes_in',
      title: '下载流量',
      render: (_: unknown, record: Proxy) => (
        <span className="text-foreground-secondary">
          {((record.total_bytes_in || 0) / 1024 / 1024).toFixed(2)} MB
        </span>
      )
    }
  ];

  // 自定义 Tooltip
  const CustomTooltip = ({ active, payload, label }: { active?: boolean; payload?: Array<{ value: number; name: string; color: string }>; label?: string }) => {
    if (active && payload && payload.length) {
      return (
        <div className="rounded-lg border border-border bg-surface p-3 shadow-xl">
          <p className="mb-2 text-sm text-foreground-muted">{label}</p>
          {payload.map((entry, index) => (
            <p key={index} className="text-sm" style={{ color: entry.color }}>
              {entry.name === 'inbound' ? '入站' : '出站'}: {entry.value} MB
            </p>
          ))}
        </div>
      );
    }
    return null;
  };

  return (
    <div className="space-y-6 p-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">仪表盘</h1>
          <p className="mt-1 text-foreground-muted">系统运行状态概览</p>
        </div>
        <div className="flex items-center gap-2 rounded-lg border border-border bg-surface/80 px-4 py-2">
          <div className="h-2 w-2 animate-pulse rounded-full bg-green-500" />
          <span className="text-sm text-foreground-secondary">系统运行正常</span>
        </div>
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="flex items-center gap-3 rounded-lg border border-red-500/30 bg-red-500/10 p-4">
          <AlertCircle className="h-5 w-5 text-red-500" />
          <span className="text-sm text-red-500">{error}</span>
        </div>
      )}

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard title="客户端总数" value={stats.clients} icon={<Monitor className="h-6 w-6 text-indigo-400" />} color="indigo" loading={loading} />
        <StatCard title="代理总数" value={stats.proxies} icon={<Network className="h-6 w-6 text-cyan-400" />} color="cyan" loading={loading} />
        <StatCard title="在线代理" value={stats.online} icon={<Wifi className="h-6 w-6 text-green-400" />} color="green" loading={loading} />
        <StatCard title="服务器数量" value={stats.servers} icon={<Server className="h-6 w-6 text-purple-400" />} color="purple" loading={loading} />
      </div>

      {/* 流量监控和在线率 */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3 items-stretch">
        {/* 实时流量 */}
        <div className="lg:col-span-2 flex">
          <Card className="h-full flex flex-col flex-1">
            <CardHeader>
              <div className="flex items-center gap-2">
                <Activity className="h-5 w-5 text-indigo-400" />
                <span>实时流量监控</span>
              </div>
            </CardHeader>
            <CardContent className="flex-1 flex flex-col justify-center">
              <TrafficMonitor />
            </CardContent>
          </Card>
        </div>

        {/* 在线率 */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <TrendingUp className="h-5 w-5 text-green-400" />
              <span>代理在线率</span>
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col items-center justify-center py-4">
              {loading ? (
                <Spinner size="lg" />
              ) : (
                <>
                  <ProgressRing value={onlineRate} label="在线率" />
                  <div className="mt-4 grid w-full grid-cols-2 gap-4 text-center">
                    <div>
                      <p className="text-2xl font-bold text-green-500">{stats.online}</p>
                      <p className="text-xs text-foreground-muted">在线</p>
                    </div>
                    <div>
                      <p className="text-2xl font-bold text-foreground-muted">{stats.proxies - stats.online}</p>
                      <p className="text-xs text-foreground-muted">离线</p>
                    </div>
                  </div>
                </>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 流量趋势图 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <TrendingUp className="h-5 w-5 text-indigo-400" />
              <span>24小时流量趋势</span>
            </div>
            <div className="flex items-center gap-4 text-sm">
              <div className="flex items-center gap-2">
                <div className="h-3 w-3 rounded-full bg-indigo-500" />
                <span className="text-foreground-muted">入站流量</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="h-3 w-3 rounded-full bg-cyan-500" />
                <span className="text-foreground-muted">出站流量</span>
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="h-80">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={trafficHistory} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="inboundGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor={chartTheme.primary} stopOpacity={chartTheme.gradientOpacityStart} />
                    <stop offset="95%" stopColor={chartTheme.primary} stopOpacity={chartTheme.gradientOpacityEnd} />
                  </linearGradient>
                  <linearGradient id="outboundGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor={chartTheme.secondary} stopOpacity={chartTheme.gradientOpacityStart} />
                    <stop offset="95%" stopColor={chartTheme.secondary} stopOpacity={chartTheme.gradientOpacityEnd} />
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
                  tickFormatter={(value) => `${value} MB`}
                />
                <Tooltip content={<CustomTooltip />} />
                <Area
                  type="monotone"
                  dataKey="inbound"
                  name="inbound"
                  stroke={chartTheme.primary}
                  strokeWidth={2}
                  fill="url(#inboundGradient)"
                />
                <Area
                  type="monotone"
                  dataKey="outbound"
                  name="outbound"
                  stroke={chartTheme.secondary}
                  strokeWidth={2}
                  fill="url(#outboundGradient)"
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>

      {/* 流量排行 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5 text-purple-400" />
              <span>流量排行</span>
            </div>
            <Badge variant="default">{topProxies.length} 个代理</Badge>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          <Table
            columns={columns}
            data={topProxies}
            rowKey="id"
            emptyText="暂无流量数据"
          />
        </CardContent>
      </Card>
    </div>
  );
}