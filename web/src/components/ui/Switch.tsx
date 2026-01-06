import { Switch as HeadlessSwitch } from '@headlessui/react';
import { Loader2 } from 'lucide-react';
import { cn } from '../../utils/cn';

export interface SwitchProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  label?: string;
  description?: string;
  disabled?: boolean;
  loading?: boolean;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export function Switch({
  checked,
  onChange,
  label,
  description,
  disabled = false,
  loading = false,
  size = 'md',
  className,
}: SwitchProps) {
  const sizes = {
    sm: { switch: 'h-5 w-9', thumb: 'h-4 w-4', translate: 'translate-x-4', loader: 'h-3 w-3' },
    md: { switch: 'h-6 w-11', thumb: 'h-5 w-5', translate: 'translate-x-5', loader: 'h-3.5 w-3.5' },
    lg: { switch: 'h-7 w-14', thumb: 'h-6 w-6', translate: 'translate-x-7', loader: 'h-4 w-4' },
  };

  const sizeConfig = sizes[size];
  const isDisabled = disabled || loading;

  return (
    <HeadlessSwitch.Group>
      <div className={cn('flex items-center gap-3', className)}>
        <HeadlessSwitch
          checked={checked}
          onChange={onChange}
          disabled={isDisabled}
          className={cn(
            'relative inline-flex shrink-0 cursor-pointer rounded-full border-2 border-transparent',
            'transition-colors duration-200 ease-in-out',
            'focus:outline-none focus:ring-2 focus:ring-indigo-500/50 focus:ring-offset-2',
            'focus:ring-offset-surface-elevated',
            'disabled:opacity-50 disabled:cursor-not-allowed',
            checked ? 'bg-indigo-600' : 'bg-border',
            sizeConfig.switch
          )}
        >
          <span
            aria-hidden="true"
            className={cn(
              'pointer-events-none inline-flex items-center justify-center rounded-full bg-white shadow-lg ring-0',
              'transform transition duration-200 ease-in-out',
              checked ? sizeConfig.translate : 'translate-x-0',
              sizeConfig.thumb
            )}
          >
            {loading && <Loader2 className={cn('animate-spin text-indigo-600', sizeConfig.loader)} />}
          </span>
        </HeadlessSwitch>
        {(label || description) && (
          <div className="flex flex-col">
            {label && (
              <HeadlessSwitch.Label className="text-sm font-medium cursor-pointer text-foreground">
                {label}
              </HeadlessSwitch.Label>
            )}
            {description && (
              <HeadlessSwitch.Description className="text-xs text-foreground-muted">
                {description}
              </HeadlessSwitch.Description>
            )}
          </div>
        )}
      </div>
    </HeadlessSwitch.Group>
  );
}