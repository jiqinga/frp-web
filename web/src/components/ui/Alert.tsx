import { AlertCircle, CheckCircle, Info, AlertTriangle, X } from 'lucide-react';
import { cn } from '../../utils/cn';

const variants = {
  error: { bg: 'bg-red-500/10', border: 'border-red-500/30', text: 'text-red-400', Icon: AlertCircle },
  success: { bg: 'bg-green-500/10', border: 'border-green-500/30', text: 'text-green-400', Icon: CheckCircle },
  warning: { bg: 'bg-yellow-500/10', border: 'border-yellow-500/30', text: 'text-yellow-400', Icon: AlertTriangle },
  info: { bg: 'bg-blue-500/10', border: 'border-blue-500/30', text: 'text-blue-400', Icon: Info },
} as const;

export interface AlertProps {
  type?: keyof typeof variants;
  message: string;
  title?: string;
  closable?: boolean;
  onClose?: () => void;
  className?: string;
}

export function Alert({ type = 'error', message, title, closable, onClose, className }: AlertProps) {
  const { bg, border, text, Icon } = variants[type];

  return (
    <div className={cn("flex items-start gap-3 rounded-lg border p-4", bg, border, text, className)}>
      <Icon className="h-5 w-5 flex-shrink-0 mt-0.5" />
      <div className="flex-1 min-w-0">
        {title && <p className="font-medium">{title}</p>}
        <p className="text-sm">{message}</p>
      </div>
      {closable && onClose && (
        <button onClick={onClose} className="flex-shrink-0 hover:opacity-70">
          <X className="h-4 w-4" />
        </button>
      )}
    </div>
  );
}