import { useState, type ReactNode } from 'react';
import { cn } from '../../utils/cn';

export interface TabItem {
  key: string;
  label: ReactNode;
  children: ReactNode;
  disabled?: boolean;
  icon?: ReactNode;
}

export interface TabsProps {
  items: TabItem[];
  defaultActiveKey?: string;
  activeKey?: string;
  onChange?: (key: string) => void;
  className?: string;
  variant?: 'line' | 'pills';
}

export function Tabs({
  items,
  defaultActiveKey,
  activeKey: controlledActiveKey,
  onChange,
  className,
  variant = 'line',
}: TabsProps) {
  const [internalActiveKey, setInternalActiveKey] = useState(
    defaultActiveKey || items[0]?.key || ''
  );

  const activeKey = controlledActiveKey ?? internalActiveKey;

  const handleTabClick = (key: string) => {
    if (!controlledActiveKey) {
      setInternalActiveKey(key);
    }
    onChange?.(key);
  };

  const activeItem = items.find((item) => item.key === activeKey);

  const lineStyles = {
    tab: 'px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors',
    active: 'border-indigo-500 text-indigo-400',
    inactive: 'border-transparent text-foreground-muted hover:text-foreground hover:border-border',
    container: 'border-b border-border',
  };

  const pillStyles = {
    tab: 'px-4 py-2 text-sm font-medium rounded-lg transition-colors',
    active: 'bg-indigo-600 text-white',
    inactive: 'text-foreground-muted hover:text-foreground hover:bg-surface-hover',
    container: 'bg-surface-hover p-1 rounded-lg',
  };

  const styles = variant === 'line' ? lineStyles : pillStyles;

  return (
    <div className={className}>
      <div className={cn('flex gap-1', styles.container)}>
        {items.map((item) => (
          <button
            key={item.key}
            onClick={() => !item.disabled && handleTabClick(item.key)}
            disabled={item.disabled}
            className={cn(
              styles.tab,
              item.key === activeKey ? styles.active : styles.inactive,
              item.disabled && 'opacity-50 cursor-not-allowed',
              'flex items-center gap-2'
            )}
          >
            {item.icon}
            {item.label}
          </button>
        ))}
      </div>
      {activeItem?.children && <div className="mt-4">{activeItem.children}</div>}
    </div>
  );
}