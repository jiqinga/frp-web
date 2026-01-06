import { type ReactNode } from 'react';
import { cn } from '../../utils/cn';

export interface TechTableContainerProps {
  children: ReactNode;
  header?: ReactNode;
  className?: string;
  showScanLine?: boolean;
}

/**
 * 科技风表格容器组件
 * 
 * 提供统一的科技风格样式：
 * - 发光边框效果（indigo-purple-cyan 渐变）
 * - 可选的扫描线动画
 * - 背景模糊和半透明效果
 * - 可选的表头插槽
 */
export function TechTableContainer({
  children,
  header,
  className,
  showScanLine = true,
}: TechTableContainerProps) {
  return (
    <div className={cn('relative', className)}>
      {/* 发光边框效果 */}
      <div className="absolute -inset-0.5 bg-gradient-to-r from-indigo-500/20 via-purple-500/20 to-cyan-500/20 rounded-xl blur opacity-30" />
      
      {/* 主容器 */}
      <div className="relative backdrop-blur-sm rounded-xl overflow-hidden border border-border bg-surface/80">
        {/* 扫描线动画 */}
        {showScanLine && (
          <div className="absolute inset-0 overflow-hidden pointer-events-none">
            <div className="absolute inset-0 bg-gradient-to-b from-transparent via-indigo-500/5 to-transparent h-[200%] animate-scan" />
          </div>
        )}
        
        {/* 可选的表头 */}
        {header && (
          <div className="relative z-10">
            {header}
          </div>
        )}
        
        {/* 表格内容 */}
        <div className="relative z-10">
          {children}
        </div>
      </div>
    </div>
  );
}