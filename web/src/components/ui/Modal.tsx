import { Fragment, type ReactNode } from 'react';
import { Dialog, DialogPanel, DialogTitle, Transition, TransitionChild } from '@headlessui/react';
import { X } from 'lucide-react';
import { cn } from '../../utils/cn';

export interface ModalProps {
  open: boolean;
  onClose: () => void;
  title?: ReactNode;
  description?: string;
  children: ReactNode;
  footer?: ReactNode;
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full' | '2xl' | '3xl' | '4xl' | '5xl';
  className?: string;
  showCloseButton?: boolean;
  /** 是否在移动端全屏显示 */
  fullscreenOnMobile?: boolean;
  /** 内容区域的自定义类名 */
  contentClassName?: string;
  /** 是否显示页脚分隔线 */
  showFooterBorder?: boolean;
}

const sizeClasses = {
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
  xl: 'max-w-xl',
  '2xl': 'max-w-2xl',
  '3xl': 'max-w-3xl',
  '4xl': 'max-w-4xl',
  '5xl': 'max-w-5xl',
  full: 'max-w-4xl',
};

export function Modal({
  open,
  onClose,
  title,
  description,
  children,
  footer,
  size = 'md',
  className,
  showCloseButton = true,
  fullscreenOnMobile = false,
  contentClassName,
  showFooterBorder = true,
}: ModalProps) {
  return (
    <Transition appear show={open} as={Fragment}>
      <Dialog as="div" className="relative z-50" onClose={onClose}>
        {/* Backdrop */}
        <TransitionChild
          as={Fragment}
          enter="ease-out duration-300"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-in duration-200"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-black/60 backdrop-blur-sm" />
        </TransitionChild>

        <div className="fixed inset-0 overflow-y-auto">
          <div className={cn(
            'flex min-h-full items-center justify-center',
            fullscreenOnMobile ? 'p-0 sm:p-4' : 'p-4'
          )}>
            <TransitionChild
              as={Fragment}
              enter="ease-out duration-300"
              enterFrom="opacity-0 scale-95"
              enterTo="opacity-100 scale-100"
              leave="ease-in duration-200"
              leaveFrom="opacity-100 scale-100"
              leaveTo="opacity-0 scale-95"
            >
              <DialogPanel
                className={cn(
                  'w-full transform overflow-hidden',
                  'bg-surface border border-border shadow-2xl',
                  'transition-all',
                  fullscreenOnMobile
                    ? 'min-h-screen sm:min-h-0 rounded-none sm:rounded-xl'
                    : 'rounded-xl',
                  fullscreenOnMobile
                    ? `max-w-none sm:${sizeClasses[size]}`
                    : sizeClasses[size],
                  className
                )}
              >
                {/* Header */}
                {(title || showCloseButton) && (
                  <div className="flex items-center justify-between border-b border-border-subtle px-4 py-3 sm:px-6 sm:py-4">
                    <div className="flex-1 min-w-0">
                      {title && (
                        <DialogTitle className="text-base sm:text-lg font-semibold truncate text-foreground">
                          {title}
                        </DialogTitle>
                      )}
                      {description && (
                        <p className="mt-1 text-xs sm:text-sm truncate text-foreground-muted">{description}</p>
                      )}
                    </div>
                    {showCloseButton && (
                      <button
                        onClick={onClose}
                        className="p-2 sm:p-1 rounded-lg text-foreground-subtle hover:text-foreground hover:bg-surface-hover transition-colors touch-manipulation ml-2 flex-shrink-0"
                      >
                        <X className="h-5 w-5" />
                      </button>
                    )}
                  </div>
                )}

                {/* Content */}
                <div className={cn(
                  'px-4 py-3 sm:px-6 sm:py-4',
                  fullscreenOnMobile && 'flex-1 overflow-y-auto',
                  contentClassName
                )}>
                  {children}
                </div>

                {/* Footer */}
                {footer && (
                  <div className={cn(
                    'flex items-center gap-2 sm:gap-3',
                    'px-4 py-3 sm:px-6 sm:py-4',
                    'bg-surface-hover',
                    showFooterBorder && 'border-t border-border-subtle',
                    fullscreenOnMobile && 'sticky bottom-0',
                    fullscreenOnMobile
                      ? 'flex-col sm:flex-row sm:justify-end [&>button]:w-full [&>button]:sm:w-auto'
                      : 'justify-end'
                  )}>
                    {footer}
                  </div>
                )}
              </DialogPanel>
            </TransitionChild>
          </div>
        </div>
      </Dialog>
    </Transition>
  );
}