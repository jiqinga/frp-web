import { Checkbox as HeadlessCheckbox } from '@headlessui/react';
import { Check, Minus } from 'lucide-react';
import { cn } from '../../utils/cn';

export interface CheckboxProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  indeterminate?: boolean;
  disabled?: boolean;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export function Checkbox({
  checked,
  onChange,
  indeterminate = false,
  disabled = false,
  size = 'md',
  className,
}: CheckboxProps) {
  const sizes = {
    sm: { box: 'h-4 w-4', icon: 'h-3 w-3' },
    md: { box: 'h-5 w-5', icon: 'h-3.5 w-3.5' },
    lg: { box: 'h-6 w-6', icon: 'h-4 w-4' },
  };

  const sizeConfig = sizes[size];
  const isActive = checked || indeterminate;

  return (
    <HeadlessCheckbox
      checked={checked}
      onChange={onChange}
      disabled={disabled}
      className={cn(
        'group relative inline-flex items-center justify-center rounded-md cursor-pointer',
        'transition-all duration-200 ease-in-out',
        'focus:outline-none focus:ring-2 focus:ring-indigo-500/50 focus:ring-offset-2',
        'focus:ring-offset-surface-elevated',
        'disabled:opacity-50 disabled:cursor-not-allowed',
        sizeConfig.box,
        className
      )}
      onClick={(e: React.MouseEvent) => e.stopPropagation()}
    >
      {/* 发光边框效果 */}
      <span
        className={cn(
          'absolute inset-0 rounded-md transition-all duration-200',
          isActive
            ? 'bg-gradient-to-r from-indigo-500/40 via-purple-500/40 to-indigo-500/40 blur-sm opacity-100'
            : 'opacity-0'
        )}
      />
      
      {/* 主体框 */}
      <span
        className={cn(
          'relative flex items-center justify-center rounded-md border-2 transition-all duration-200',
          sizeConfig.box,
          isActive
            ? 'bg-indigo-600 border-indigo-500 shadow-lg shadow-indigo-500/30'
            : 'bg-surface border-border group-hover:border-border-light'
        )}
      >
        {/* 勾选图标 */}
        <span
          className={cn(
            'transition-all duration-200',
            isActive ? 'scale-100 opacity-100' : 'scale-0 opacity-0'
          )}
        >
          {indeterminate ? (
            <Minus className={cn(sizeConfig.icon, 'text-white')} strokeWidth={3} />
          ) : (
            <Check className={cn(sizeConfig.icon, 'text-white')} strokeWidth={3} />
          )}
        </span>
      </span>
    </HeadlessCheckbox>
  );
}