import { useState, useRef, useEffect, type ReactNode, type ReactElement, cloneElement } from 'react';
import { createPortal } from 'react-dom';
import { cn } from '../../utils/cn';

export interface TooltipProps {
  content: ReactNode;
  children: ReactElement;
  position?: 'top' | 'bottom' | 'left' | 'right';
  delay?: number;
  className?: string;
}

export function Tooltip({
  content,
  children,
  position = 'top',
  delay = 200,
  className,
}: TooltipProps) {
  const [isVisible, setIsVisible] = useState(false);
  const [tooltipPosition, setTooltipPosition] = useState({ top: 0, left: 0 });
  const [timeoutId, setTimeoutId] = useState<ReturnType<typeof setTimeout> | null>(null);
  const triggerRef = useRef<HTMLDivElement>(null);
  const tooltipRef = useRef<HTMLDivElement>(null);

  const showTooltip = () => {
    const id = setTimeout(() => setIsVisible(true), delay);
    setTimeoutId(id);
  };

  const hideTooltip = () => {
    if (timeoutId) {
      clearTimeout(timeoutId);
      setTimeoutId(null);
    }
    setIsVisible(false);
  };

  // 计算 tooltip 位置
  useEffect(() => {
    if (isVisible && triggerRef.current) {
      const triggerRect = triggerRef.current.getBoundingClientRect();
      const tooltipEl = tooltipRef.current;
      const gap = 8; // 间距

      let top = 0;
      let left = 0;

      // 先设置一个初始位置，等 tooltip 渲染后再调整
      const tooltipWidth = tooltipEl?.offsetWidth || 0;
      const tooltipHeight = tooltipEl?.offsetHeight || 0;

      switch (position) {
        case 'top':
          top = triggerRect.top - tooltipHeight - gap;
          left = triggerRect.left + triggerRect.width / 2 - tooltipWidth / 2;
          break;
        case 'bottom':
          top = triggerRect.bottom + gap;
          left = triggerRect.left + triggerRect.width / 2 - tooltipWidth / 2;
          break;
        case 'left':
          top = triggerRect.top + triggerRect.height / 2 - tooltipHeight / 2;
          left = triggerRect.left - tooltipWidth - gap;
          break;
        case 'right':
          top = triggerRect.top + triggerRect.height / 2 - tooltipHeight / 2;
          left = triggerRect.right + gap;
          break;
      }

      setTooltipPosition({ top, left });
    }
  }, [isVisible, position]);

  const arrows = {
    top: 'top-full left-1/2 -translate-x-1/2 border-t-surface border-x-transparent border-b-transparent',
    bottom: 'bottom-full left-1/2 -translate-x-1/2 border-b-surface border-x-transparent border-t-transparent',
    left: 'left-full top-1/2 -translate-y-1/2 border-l-surface border-y-transparent border-r-transparent',
    right: 'right-full top-1/2 -translate-y-1/2 border-r-surface border-y-transparent border-l-transparent',
  };

  const tooltipContent = isVisible && content && createPortal(
    <div
      ref={tooltipRef}
      style={{
        position: 'fixed',
        top: tooltipPosition.top,
        left: tooltipPosition.left,
        zIndex: 9999,
      }}
      className={cn(
        'px-2 py-1 text-xs font-medium rounded-md shadow-lg whitespace-nowrap',
        'animate-in fade-in-0 zoom-in-95 duration-150',
        'text-foreground-secondary bg-surface border border-border',
        className
      )}
      role="tooltip"
    >
      {content}
      <span
        className={cn(
          'absolute w-0 h-0 border-4',
          arrows[position]
        )}
      />
    </div>,
    document.body
  );

  return (
    <div ref={triggerRef} className="inline-flex">
      {cloneElement(children, {
        onMouseEnter: showTooltip,
        onMouseLeave: hideTooltip,
        onFocus: showTooltip,
        onBlur: hideTooltip,
      })}
      {tooltipContent}
    </div>
  );
}