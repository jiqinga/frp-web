// 图表主题配置 - 集中管理 Recharts 图表颜色

export interface ChartTheme {
  // 网格和坐标轴
  grid: string;
  axis: string;
  tick: string;
  tickLine: string;
  
  // Tooltip
  tooltipBg: string;
  tooltipBorder: string;
  tooltipText: string;
  
  // 数据系列颜色
  primary: string;
  secondary: string;
  success: string;
  warning: string;
  error: string;
  info: string;
  purple: string;
  
  // 渐变透明度
  gradientOpacityStart: number;
  gradientOpacityEnd: number;
}

const darkTheme: ChartTheme = {
  grid: '#334155',
  axis: '#64748b',
  tick: '#94a3b8',
  tickLine: '#475569',
  tooltipBg: '#1e293b',
  tooltipBorder: '#334155',
  tooltipText: '#e2e8f0',
  primary: '#6366f1',
  secondary: '#22d3ee',
  success: '#22c55e',
  warning: '#f59e0b',
  error: '#ef4444',
  info: '#3b82f6',
  purple: '#a855f7',
  gradientOpacityStart: 0.3,
  gradientOpacityEnd: 0,
};

const lightTheme: ChartTheme = {
  grid: '#cbd5e1',
  axis: '#475569',
  tick: '#475569',
  tickLine: '#94a3b8',
  tooltipBg: '#ffffff',
  tooltipBorder: '#e2e8f0',
  tooltipText: '#0f172a',
  primary: '#6366f1',
  secondary: '#22d3ee',
  success: '#22c55e',
  warning: '#f59e0b',
  error: '#ef4444',
  info: '#3b82f6',
  purple: '#a855f7',
  gradientOpacityStart: 0.3,
  gradientOpacityEnd: 0,
};

export const getChartTheme = (isLight: boolean): ChartTheme => 
  isLight ? lightTheme : darkTheme;

// 常用颜色数组
export const chartColors = ['#6366f1', '#22d3ee', '#10b981', '#f59e0b', '#ef4444'];

// 通用 Tooltip 样式
export const getTooltipStyle = (theme: ChartTheme) => ({
  contentStyle: {
    backgroundColor: theme.tooltipBg,
    border: `1px solid ${theme.tooltipBorder}`,
    borderRadius: '8px',
  },
  labelStyle: { color: theme.tooltipText },
  itemStyle: { color: theme.tooltipText },
});

// 通用坐标轴样式
export const getAxisStyle = (theme: ChartTheme) => ({
  stroke: theme.axis,
  tick: { fill: theme.tick, fontSize: 10 },
  tickLine: { stroke: theme.tickLine },
});