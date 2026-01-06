import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from '@headlessui/react';
import { Check, ChevronDown } from 'lucide-react';
import { cn } from '../../utils/cn';

export interface SelectOption {
  value: string | number;
  label: string;
  disabled?: boolean;
}

export interface SelectProps {
  value?: string | number;
  onChange?: (value: string | number) => void;
  options: SelectOption[];
  placeholder?: string;
  label?: string;
  error?: string;
  disabled?: boolean;
  className?: string;
  size?: 'sm' | 'md' | 'lg';
}

export function Select({
  value,
  onChange,
  options,
  placeholder = '请选择',
  label,
  error,
  disabled,
  className,
  size = 'md',
}: SelectProps) {
  // 确保 value 始终是定义的值，避免 uncontrolled to controlled 警告
  const controlledValue = value ?? '';
  const selectedOption = options.find(opt => opt.value === controlledValue);

  const sizes = {
    sm: 'h-8 text-xs px-3',
    md: 'h-10 text-sm px-4',
    lg: 'h-12 text-base px-4',
  };

  return (
    <div className={cn('w-full', className)}>
      {label && (
        <label className="block text-sm font-medium mb-1.5 text-foreground-secondary">
          {label}
        </label>
      )}
      <Listbox value={controlledValue} onChange={onChange} disabled={disabled}>
        <ListboxButton
          className={cn(
            'relative w-full rounded-lg text-left cursor-pointer',
            'bg-input-bg border border-input-border',
            'focus:outline-none focus:ring-2 focus:ring-indigo-500/50 focus:border-indigo-500',
            'disabled:opacity-50 disabled:cursor-not-allowed',
            'transition-all duration-200',
            error && 'border-red-500 focus:ring-red-500/50 focus:border-red-500',
            sizes[size]
          )}
        >
          <span className={cn('block truncate', !selectedOption && 'text-input-placeholder')}>
            {selectedOption?.label || placeholder}
          </span>
          <span className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
            <ChevronDown className="h-4 w-4 text-foreground-subtle" aria-hidden="true" />
          </span>
        </ListboxButton>
        <ListboxOptions
          anchor="bottom start"
          className="z-50 mt-1 max-h-60 w-[var(--button-width)] overflow-auto rounded-lg py-1 shadow-lg focus:outline-none bg-surface border border-border transition duration-100 ease-out data-[closed]:opacity-0"
        >
              {options.map((option) => (
                <ListboxOption
                  key={option.value}
                  value={option.value}
                  disabled={option.disabled}
                  className={({ active, selected }) =>
                    cn(
                      'relative cursor-pointer select-none py-2 pl-10 pr-4 text-sm',
                      active ? 'bg-indigo-600/20 text-indigo-500' : 'text-foreground-secondary',
                      selected && 'bg-indigo-600/10',
                      option.disabled && 'opacity-50 cursor-not-allowed'
                    )
                  }
                >
                  {({ selected }) => (
                    <>
                      <span className={cn('block truncate', selected && 'font-medium text-indigo-400')}>
                        {option.label}
                      </span>
                      {selected && (
                        <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-indigo-400">
                          <Check className="h-4 w-4" aria-hidden="true" />
                        </span>
                      )}
                    </>
              )}
            </ListboxOption>
          ))}
        </ListboxOptions>
      </Listbox>
      {error && (
        <p className="mt-1.5 text-xs text-red-400">{error}</p>
      )}
    </div>
  );
}