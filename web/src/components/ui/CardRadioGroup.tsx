import { isValidElement } from 'react';
import { cn } from '../../utils/cn';
import { useThemeStore } from '../../store/theme';
import type { ReactNode, ComponentType } from 'react';

export interface CardRadioOption<T extends string | number = string> {
  value: T;
  label: string;
  icon?: ReactNode | ComponentType<{ className?: string }>;
  disabled?: boolean;
}

export interface CardRadioGroupProps<T extends string | number = string> {
  value: T;
  onChange: (value: T) => void;
  options: CardRadioOption<T>[];
  name?: string;
  className?: string;
  disabled?: boolean;
  /** 是否让每个选项占据相等的空间 */
  equalWidth?: boolean;
}

export function CardRadioGroup<T extends string | number = string>({
  value,
  onChange,
  options,
  name,
  className,
  disabled = false,
  equalWidth = true,
}: CardRadioGroupProps<T>) {
  const { theme } = useThemeStore();
  const isLight = theme === 'light';

  return (
    <div className={cn('flex gap-3', className)}>
      {options.map((option) => {
        const isSelected = value === option.value;
        const isDisabled = disabled || option.disabled;

        // 处理 icon：可能是 ReactNode 或 ComponentType
        const renderIcon = () => {
          if (!option.icon) return null;
          
          // 如果已经是 React 元素，直接返回
          if (isValidElement(option.icon)) {
            return option.icon;
          }
          
          // 否则当作组件类型渲染（包括普通函数组件和 forwardRef 组件）
          const IconComponent = option.icon as ComponentType<{ className?: string }>;
          return <IconComponent className="h-4 w-4" />;
        };

        return (
          <label
            key={String(option.value)}
            className={cn(
              'flex items-center gap-2 px-4 py-3 rounded-lg border cursor-pointer transition-all',
              equalWidth && 'flex-1',
              isDisabled && 'opacity-50 cursor-not-allowed',
              isSelected
                ? 'border-indigo-500 bg-indigo-500/10 text-indigo-600 dark:text-indigo-400'
                : isLight
                  ? 'border-slate-300 bg-slate-100 text-slate-700 hover:border-slate-400'
                  : 'border-slate-600 bg-slate-800/50 text-slate-300 hover:border-slate-500'
            )}
          >
            <input
              type="radio"
              name={name}
              value={String(option.value)}
              checked={isSelected}
              onChange={() => !isDisabled && onChange(option.value)}
              disabled={isDisabled}
              className="sr-only"
            />
            {renderIcon()}
            <span className="text-sm">{option.label}</span>
          </label>
        );
      })}
    </div>
  );
}