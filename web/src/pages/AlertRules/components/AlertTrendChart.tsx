import { useMemo } from 'react';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import type { AlertLog } from '../../../api/alert';
import { Card, CardHeader, CardContent } from '../../../components/ui';
import { useThemeStore } from '../../../store/theme';
import { getChartTheme } from '../../../utils/chartTheme';

interface AlertTrendChartProps {
  logs: AlertLog[];
}

export function AlertTrendChart({ logs }: AlertTrendChartProps) {
  const { theme } = useThemeStore();
  const isLight = theme === 'light';
  const chartTheme = getChartTheme(isLight);
  
  const chartData = useMemo(() => {
    const last7Days = Array.from({ length: 7 }, (_, i) => {
      const date = new Date();
      date.setDate(date.getDate() - (6 - i));
      return {
        date: date.toLocaleDateString('zh-CN', { month: 'numeric', day: 'numeric' }),
        timestamp: new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime(),
        count: 0,
      };
    });

    logs.forEach(log => {
      const logDate = new Date(log.created_at);
      const logTimestamp = new Date(logDate.getFullYear(), logDate.getMonth(), logDate.getDate()).getTime();
      const day = last7Days.find(d => d.timestamp === logTimestamp);
      if (day) day.count++;
    });

    return last7Days;
  }, [logs]);

  return (
    <Card>
      <CardHeader>近7日告警趋势</CardHeader>
      <CardContent>
        <div className="h-48">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={chartData}>
              <XAxis dataKey="date" tick={{ fill: chartTheme.tick, fontSize: 12 }} axisLine={false} tickLine={false} />
              <YAxis tick={{ fill: chartTheme.tick, fontSize: 12 }} axisLine={false} tickLine={false} allowDecimals={false} />
              <Tooltip
                contentStyle={{
                  background: chartTheme.tooltipBg,
                  border: `1px solid ${chartTheme.tooltipBorder}`,
                  borderRadius: 8
                }}
                labelStyle={{ color: chartTheme.tooltipText }}
              />
              <Bar dataKey="count" name="告警数" fill={chartTheme.warning} radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}