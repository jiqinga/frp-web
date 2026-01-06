import { Minus, Plus } from 'lucide-react';
import { cn } from '../../utils/cn';

export interface NumberStepperProps {
  value: string | number;
  onChange: (value: string) => void;
  min?: number;
  max?: number;
  step?: number;
  className?: string;
  disabled?: boolean;
}

export function NumberStepper({
  value,
  onChange,
  min = 0,
  max = Infinity,
  step = 1,
  className,
  disabled = false,
}: NumberStepperProps) {
  const numValue = typeof value === 'string' ? parseFloat(value) || 0 : value;

  const handleDecrement = () => {
    const newValue = Math.max(min, numValue - step);
    onChange(String(newValue));
  };

  const handleIncrement = () => {
    const newValue = Math.min(max, numValue + step);
    onChange(String(newValue));
  };

  const buttonClass = cn(
    'flex items-center justify-center w-8 h-8 rounded-lg',
    'transition-colors duration-200',
    'disabled:opacity-50 disabled:cursor-not-allowed',
    'bg-surface-hover hover:bg-surface-active text-foreground-secondary'
  );

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <button
        type="button"
        onClick={handleDecrement}
        disabled={disabled || numValue <= min}
        className={buttonClass}
      >
        <Minus className="h-4 w-4" />
      </button>
      <input
        type="number"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        min={min}
        max={max}
        disabled={disabled}
        className={cn(
          'w-20 h-8 text-center rounded-lg',
          'focus:outline-none focus:ring-2 focus:ring-indigo-500/50 focus:border-indigo-500',
          'disabled:opacity-50 disabled:cursor-not-allowed',
          '[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none',
          'bg-input-bg border border-input-border text-foreground'
        )}
      />
      <button
        type="button"
        onClick={handleIncrement}
        disabled={disabled || numValue >= max}
        className={buttonClass}
      >
        <Plus className="h-4 w-4" />
      </button>
    </div>
  );
}