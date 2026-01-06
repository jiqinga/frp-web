import { Loader2 } from 'lucide-react';
import { cn } from '../../utils/cn';

export interface SpinnerProps {
  size?: 'sm' | 'md' | 'lg' | 'xl';
  className?: string;
  label?: string;
}

export function Spinner({ size = 'md', className, label }: SpinnerProps) {
  const sizes = {
    sm: 'h-4 w-4',
    md: 'h-6 w-6',
    lg: 'h-8 w-8',
    xl: 'h-12 w-12',
  };

  return (
    <div className={cn('flex items-center justify-center gap-2', className)}>
      <Loader2 className={cn('animate-spin text-indigo-500', sizes[size])} />
      {label && <span className="text-sm text-foreground-muted">{label}</span>}
    </div>
  );
}

// 全屏加载组件
export function FullPageSpinner({ label = '加载中...' }: { label?: string }) {
  return (
    <div className="fixed inset-0 flex items-center justify-center backdrop-blur-sm z-50 bg-bg-primary/80">
      <div className="flex flex-col items-center gap-4">
        <div className="relative">
          <div className="h-16 w-16 rounded-full border-4 border-border" />
          <div className="absolute inset-0 h-16 w-16 rounded-full border-4 border-transparent border-t-indigo-500 animate-spin" />
        </div>
        <span className="font-medium text-foreground-secondary">{label}</span>
      </div>
    </div>
  );
}