import { cn } from '../../utils/cn';
import type { ReactNode } from 'react';

export interface RadioOption<T extends string | number = string> {
  value: T;
  label: ReactNode;
  icon?: ReactNode;
  disabled?: boolean;
}

interface RadioGroupProps<T extends string | number = string> {
  value: T;
  onChange: (value: T) => void;
  options: RadioOption<T>[];
  name?: string;
  className?: string;
  direction?: 'horizontal' | 'vertical';
}

export function RadioGroup<T extends string | number = string>({
  value,
  onChange,
  options,
  name,
  className,
  direction = 'horizontal',
}: RadioGroupProps<T>) {
  return (
    <div className={cn(
      'flex gap-4',
      direction === 'vertical' && 'flex-col',
      className
    )}>
      {options.map((option) => (
        <label
          key={String(option.value)}
          className={cn(
            'flex items-center gap-2 cursor-pointer',
            option.disabled && 'opacity-50 cursor-not-allowed'
          )}
        >
          <input
            type="radio"
            name={name}
            value={String(option.value)}
            checked={value === option.value}
            onChange={() => !option.disabled && onChange(option.value)}
            disabled={option.disabled}
            className="w-4 h-4 text-indigo-500 focus:ring-indigo-500 bg-input-bg border-input-border"
          />
          {option.icon}
          <span className="text-foreground-secondary">{option.label}</span>
        </label>
      ))}
    </div>
  );
}