import { useMemo } from 'react';
import type { ProxyHistory } from '../../hooks/useRealtimeMonitor';

interface SparkLineProps {
  data: ProxyHistory[];
  width?: number;
  height?: number;
  color?: string;
}

export function SparkLine({ data, width = 80, height = 24, color = '#6366f1' }: SparkLineProps) {
  const path = useMemo(() => {
    if (data.length < 2) return '';
    
    const values = data.map(d => d.inRate + d.outRate);
    const max = Math.max(...values, 1);
    const min = Math.min(...values, 0);
    const range = max - min || 1;
    
    const points = values.map((v, i) => {
      const x = (i / (values.length - 1)) * width;
      const y = height - ((v - min) / range) * height;
      return `${x},${y}`;
    });
    
    return `M${points.join(' L')}`;
  }, [data, width, height]);

  // 创建渐变填充路径
  const areaPath = useMemo(() => {
    if (data.length < 2) return '';
    
    const values = data.map(d => d.inRate + d.outRate);
    const max = Math.max(...values, 1);
    const min = Math.min(...values, 0);
    const range = max - min || 1;
    
    const points = values.map((v, i) => {
      const x = (i / (values.length - 1)) * width;
      const y = height - ((v - min) / range) * height;
      return `${x},${y}`;
    });
    
    return `M0,${height} L${points.join(' L')} L${width},${height} Z`;
  }, [data, width, height]);

  if (data.length < 2) {
    return (
      <div
        className="rounded bg-surface-hover"
        style={{ width, height }}
      />
    );
  }

  return (
    <svg width={width} height={height} className="block">
      <defs>
        <linearGradient id={`sparkGradient-${color.replace('#', '')}`} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={color} stopOpacity={0.3} />
          <stop offset="100%" stopColor={color} stopOpacity={0} />
        </linearGradient>
      </defs>
      <path 
        d={areaPath} 
        fill={`url(#sparkGradient-${color.replace('#', '')})`}
      />
      <path 
        d={path} 
        fill="none" 
        stroke={color} 
        strokeWidth={1.5}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}