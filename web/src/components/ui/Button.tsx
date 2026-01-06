import { forwardRef, type ButtonHTMLAttributes, type ReactNode } from 'react';
import { Loader2 } from 'lucide-react';
import { cn } from '../../utils/cn';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost' | 'outline';
  size?: 'sm' | 'md' | 'lg';
  loading?: boolean;
  icon?: ReactNode;
  children?: ReactNode;
  /** 是否全宽显示 */
  fullWidth?: boolean;
  /** 是否只显示图标（移动端优化） */
  iconOnly?: boolean;
  /** 移动端是否只显示图标 */
  iconOnlyOnMobile?: boolean;
}

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({
    className,
    variant = 'primary',
    size = 'md',
    loading,
    disabled,
    icon,
    children,
    fullWidth = false,
    iconOnly = false,
    iconOnlyOnMobile = false,
    ...props
  }, ref) => {
    const baseStyles = cn(
      'group inline-flex items-center justify-center font-medium',
      'transition-all duration-300 ease-out',
      'focus:outline-none focus:ring-2 focus:ring-offset-2',
      'focus:ring-offset-surface-elevated',
      'disabled:opacity-50 disabled:cursor-not-allowed',
      'touch-manipulation active:scale-[0.98]',
      'relative overflow-hidden'
    );
    
    const variants = {
      primary: cn(
        'bg-gradient-to-r from-indigo-600 to-indigo-500',
        'hover:from-indigo-500 hover:to-indigo-400',
        'text-white shadow-lg shadow-indigo-500/25',
        'hover:shadow-indigo-500/50 hover:shadow-xl hover:scale-105',
        'focus:ring-indigo-500',
        'before:absolute before:inset-0',
        'before:bg-gradient-to-r before:from-transparent before:via-white/20 before:to-transparent',
        'before:translate-x-[-200%] hover:before:translate-x-[200%]',
        'before:transition-transform before:duration-700'
      ),
      secondary: 'bg-surface hover:bg-surface-hover text-foreground border border-border hover:border-border-light focus:ring-border-light',
      danger: cn(
        'bg-gradient-to-r from-red-600 to-red-500',
        'hover:from-red-500 hover:to-red-400',
        'text-white shadow-lg shadow-red-500/25',
        'hover:shadow-red-500/50 hover:shadow-xl hover:scale-105',
        'focus:ring-red-500',
        'before:absolute before:inset-0',
        'before:bg-gradient-to-r before:from-transparent before:via-white/20 before:to-transparent',
        'before:translate-x-[-200%] hover:before:translate-x-[200%]',
        'before:transition-transform before:duration-700'
      ),
      ghost: 'bg-transparent hover:bg-surface-hover text-foreground-muted hover:text-foreground focus:ring-border-light',
      outline: 'bg-transparent border border-border hover:border-indigo-500 text-foreground-muted hover:text-indigo-500 focus:ring-indigo-500',
    };
    
    const sizes = {
      sm: iconOnly ? 'h-8 w-8 p-0 text-xs rounded-md' : 'h-8 px-3 text-xs rounded-md gap-1.5',
      md: iconOnly ? 'h-10 w-10 p-0 text-sm rounded-lg' : 'h-10 px-4 text-sm rounded-lg gap-2',
      lg: iconOnly ? 'h-12 w-12 p-0 text-base rounded-lg' : 'h-12 px-6 text-base rounded-lg gap-2.5',
    };

    // 移动端只显示图标的样式
    const mobileIconOnlyStyles = iconOnlyOnMobile
      ? 'sm:gap-2 [&>span:last-child]:hidden sm:[&>span:last-child]:inline'
      : '';

    return (
      <button
        ref={ref}
        className={cn(
          baseStyles,
          variants[variant],
          sizes[size],
          fullWidth && 'w-full',
          mobileIconOnlyStyles,
          className
        )}
        disabled={disabled || loading}
        {...props}
      >
        {loading ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : icon ? (
          <span className="flex-shrink-0 transition-transform duration-300 group-hover:rotate-90">{icon}</span>
        ) : null}
        {children && !iconOnly && <span className="relative z-10">{children}</span>}
      </button>
    );
  }
);

Button.displayName = 'Button';

export { Button };