import { Fragment, type ReactNode } from 'react';
import { Dialog, DialogPanel, DialogTitle, Transition, TransitionChild } from '@headlessui/react';
import { AlertTriangle, Info, CheckCircle, XCircle, X } from 'lucide-react';
import { cn } from '../../utils/cn';
import { Button } from './Button';

export type ConfirmModalType = 'confirm' | 'info' | 'success' | 'warning' | 'error';

export interface ConfirmModalProps {
  open: boolean;
  onClose: () => void;
  onConfirm?: () => void | Promise<void>;
  title: string;
  content: ReactNode;
  type?: ConfirmModalType;
  confirmText?: string;
  cancelText?: string;
  showCancel?: boolean;
  loading?: boolean;
  className?: string;
  fullWidthContent?: boolean;
}

const typeConfig = {
  confirm: {
    icon: AlertTriangle,
    iconColor: 'text-yellow-400',
    iconBg: 'bg-yellow-500/20',
  },
  info: {
    icon: Info,
    iconColor: 'text-blue-400',
    iconBg: 'bg-blue-500/20',
  },
  success: {
    icon: CheckCircle,
    iconColor: 'text-green-400',
    iconBg: 'bg-green-500/20',
  },
  warning: {
    icon: AlertTriangle,
    iconColor: 'text-yellow-400',
    iconBg: 'bg-yellow-500/20',
  },
  error: {
    icon: XCircle,
    iconColor: 'text-red-400',
    iconBg: 'bg-red-500/20',
  },
};

export function ConfirmModal({
  open,
  onClose,
  onConfirm,
  title,
  content,
  type = 'confirm',
  confirmText = '确定',
  cancelText = '取消',
  showCancel = true,
  loading = false,
  className,
  fullWidthContent = false,
}: ConfirmModalProps) {
  const config = typeConfig[type];
  const Icon = config.icon;

  const handleConfirm = async () => {
    if (onConfirm) {
      await onConfirm();
    }
    onClose();
  };

  return (
    <Transition appear show={open} as={Fragment}>
      <Dialog as="div" className="relative z-[60]" onClose={onClose}>
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
          <div className="flex min-h-full items-center justify-center p-4">
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
                  'w-full max-w-md transform overflow-hidden rounded-xl',
                  'bg-surface border border-border shadow-2xl',
                  'transition-all',
                  className
                )}
              >
                {/* Header */}
                <div className={cn("flex items-start gap-4 py-5", fullWidthContent ? "px-4" : "px-6")}>
                  <div className={cn('p-2 rounded-lg', config.iconBg)}>
                    <Icon className={cn('h-6 w-6', config.iconColor)} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <DialogTitle className="text-lg font-semibold text-foreground">
                      {title}
                    </DialogTitle>
                    {!fullWidthContent && (
                      <div className="mt-2 text-sm text-foreground-muted">
                        {content}
                      </div>
                    )}
                  </div>
                  <button
                    onClick={onClose}
                    className="p-1 rounded-lg transition-colors text-foreground-subtle hover:text-foreground hover:bg-surface-hover"
                  >
                    <X className="h-5 w-5" />
                  </button>
                </div>

                {/* Full Width Content */}
                {fullWidthContent && (
                  <div className="px-4 pb-4">
                    {content}
                  </div>
                )}

                {/* Footer */}
                <div className={cn(
                  'flex items-center justify-end gap-3 py-4 border-t border-border-subtle bg-surface-hover',
                  fullWidthContent ? 'px-4' : 'px-6'
                )}>
                  {showCancel && (
                    <Button variant="outline" onClick={onClose} disabled={loading}>
                      {cancelText}
                    </Button>
                  )}
                  <Button
                    variant={type === 'error' ? 'danger' : 'primary'}
                    onClick={handleConfirm}
                    disabled={loading}
                  >
                    {loading ? '处理中...' : confirmText}
                  </Button>
                </div>
              </DialogPanel>
            </TransitionChild>
          </div>
        </div>
      </Dialog>
    </Transition>
  );
}