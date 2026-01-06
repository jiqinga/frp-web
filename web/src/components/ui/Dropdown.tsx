import { Fragment, type ReactNode } from 'react';
import { Menu, MenuButton, MenuItem, MenuItems, Transition } from '@headlessui/react';
import { ChevronDown } from 'lucide-react';
import { cn } from '../../utils/cn';

export interface DropdownItem {
  key: string;
  label: ReactNode;
  icon?: ReactNode;
  disabled?: boolean;
  danger?: boolean;
  onClick?: () => void;
  divider?: boolean;
}

export interface DropdownProps {
  trigger: ReactNode;
  items: DropdownItem[];
  align?: 'left' | 'right';
  className?: string;
  showArrow?: boolean;
}

export function Dropdown({
  trigger,
  items,
  align = 'right',
  className,
  showArrow = false,
}: DropdownProps) {
  return (
    <Menu as="div" className={cn('relative inline-block text-left', className)}>
      <MenuButton className="inline-flex items-center gap-1">
        {trigger}
        {showArrow && <ChevronDown className="h-4 w-4 text-foreground-subtle" />}
      </MenuButton>

      <Transition
        as={Fragment}
        enter="transition ease-out duration-100"
        enterFrom="transform opacity-0 scale-95"
        enterTo="transform opacity-100 scale-100"
        leave="transition ease-in duration-75"
        leaveFrom="transform opacity-100 scale-100"
        leaveTo="transform opacity-0 scale-95"
      >
        <MenuItems
          className={cn(
            'absolute z-50 mt-2 w-56 origin-top-right rounded-lg',
            'bg-surface border border-border shadow-lg',
            'ring-1 ring-black ring-opacity-5 focus:outline-none',
            'divide-y divide-border-subtle',
            align === 'left' ? 'left-0' : 'right-0'
          )}
        >
          <div className="py-1">
            {items.map((item) =>
              item.divider ? (
                <div key={item.key} className="my-1 border-t border-border" />
              ) : (
                <MenuItem key={item.key} disabled={item.disabled}>
                  {({ active }) => (
                    <button
                      onClick={item.onClick}
                      disabled={item.disabled}
                      className={cn(
                        'flex w-full items-center gap-2 px-4 py-2 text-sm',
                        'transition-colors duration-150',
                        active && !item.disabled && 'bg-surface-hover',
                        item.danger
                          ? 'text-red-500 hover:text-red-600'
                          : 'text-foreground-secondary hover:text-foreground',
                        item.disabled && 'opacity-50 cursor-not-allowed'
                      )}
                    >
                      {item.icon && <span className="flex-shrink-0">{item.icon}</span>}
                      {item.label}
                    </button>
                  )}
                </MenuItem>
              )
            )}
          </div>
        </MenuItems>
      </Transition>
    </Menu>
  );
}