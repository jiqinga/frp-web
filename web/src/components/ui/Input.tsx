import { forwardRef, type InputHTMLAttributes, type ReactNode } from 'react';
import { cn } from '../../utils/cn';

export interface InputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'size' | 'prefix'> {
  label?: string;
  error?: string;
  hint?: string;
  prefix?: ReactNode;
  suffix?: ReactNode;
  size?: 'sm' | 'md' | 'lg';
}

const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, error, hint, prefix, suffix, size = 'md', type = 'text', id, ...props }, ref) => {
    const inputId = id || label?.toLowerCase().replace(/\s+/g, '-');
    
    const sizes = {
      sm: 'h-8 text-xs px-3',
      md: 'h-10 text-sm px-4',
      lg: 'h-12 text-base px-4',
    };

    const inputStyles = cn(
      'w-full rounded-lg transition-all duration-200',
      'bg-input-bg border border-input-border text-foreground placeholder:text-input-placeholder',
      'focus:outline-none focus:ring-2 focus:ring-indigo-500/50 focus:border-indigo-500',
      'disabled:opacity-50 disabled:cursor-not-allowed',
      error && 'border-red-500 focus:ring-red-500/50 focus:border-red-500',
      sizes[size],
      prefix && 'pl-10',
      suffix && 'pr-10',
      className
    );

    return (
      <div className="w-full">
        {label && (
          <label htmlFor={inputId} className="block text-sm font-medium mb-1.5 text-foreground-secondary">
            {label}
          </label>
        )}
        <div className="relative">
          {prefix && (
            <div className="absolute left-3 top-1/2 -translate-y-1/2 text-foreground-subtle">
              {prefix}
            </div>
          )}
          <input
            ref={ref}
            id={inputId}
            type={type}
            className={inputStyles}
            {...props}
          />
          {suffix && (
            <div className="absolute right-3 top-1/2 -translate-y-1/2 text-foreground-subtle">
              {suffix}
            </div>
          )}
        </div>
        {error && (
          <p className="mt-1.5 text-xs text-red-500">{error}</p>
        )}
        {hint && !error && (
          <p className="mt-1.5 text-xs text-foreground-subtle">{hint}</p>
        )}
      </div>
    );
  }
);

Input.displayName = 'Input';

export { Input };