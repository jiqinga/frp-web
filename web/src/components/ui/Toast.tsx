import { useState, useEffect, createContext, useContext, useCallback, useRef } from 'react';
import { CheckCircle, XCircle, AlertCircle, Info, X } from 'lucide-react';
import { cn } from '../../utils/cn';

type ToastType = 'success' | 'error' | 'warning' | 'info';

interface Toast {
  id: string;
  type: ToastType;
  message: string;
}

interface ToastContextType {
  success: (message: string) => void;
  error: (message: string) => void;
  warning: (message: string) => void;
  info: (message: string) => void;
}

const ToastContext = createContext<ToastContextType | null>(null);

export const useToast = () => {
  const context = useContext(ToastContext);
  if (!context) {
    // 返回一个默认实现，使用 console
    return {
      success: (msg: string) => console.log('✅', msg),
      error: (msg: string) => console.error('❌', msg),
      warning: (msg: string) => console.warn('⚠️', msg),
      info: (msg: string) => console.info('ℹ️', msg),
    };
  }
  return context;
};

const icons = {
  success: CheckCircle,
  error: XCircle,
  warning: AlertCircle,
  info: Info,
};

const colors = {
  success: 'border-green-500/50 bg-green-500/10 text-green-400',
  error: 'border-red-500/50 bg-red-500/10 text-red-400',
  warning: 'border-yellow-500/50 bg-yellow-500/10 text-yellow-400',
  info: 'border-blue-500/50 bg-blue-500/10 text-blue-400',
};

interface ToastItemProps {
  toast: Toast;
  onRemove: (id: string) => void;
}

const ToastItem = ({ toast, onRemove }: ToastItemProps) => {
  const Icon = icons[toast.type];

  useEffect(() => {
    const timer = setTimeout(() => {
      onRemove(toast.id);
    }, 3000);
    return () => clearTimeout(timer);
  }, [toast.id, onRemove]);

  return (
    <div
      className={cn(
        'flex items-center gap-3 rounded-lg border px-4 py-3 shadow-lg backdrop-blur-sm',
        'animate-in slide-in-from-right-full duration-300',
        colors[toast.type]
      )}
    >
      <Icon className="h-5 w-5 flex-shrink-0" />
      <span className="text-sm">{toast.message}</span>
      <button
        onClick={() => onRemove(toast.id)}
        className="ml-2 rounded p-1 hover:bg-white/10"
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  );
};

export const ToastProvider = ({ children }: { children: React.ReactNode }) => {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const addToast = useCallback((type: ToastType, message: string) => {
    const id = Math.random().toString(36).substring(2, 9);
    setToasts((prev) => [...prev, { id, type, message }]);
  }, []);

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  // 使用 useRef 存储 toast 方法，避免每次渲染创建新对象
  const toastMethodsRef = useRef<ToastContextType>({
    success: (message) => addToast('success', message),
    error: (message) => addToast('error', message),
    warning: (message) => addToast('warning', message),
    info: (message) => addToast('info', message),
  });

  // 更新 ref 中的方法（addToast 是稳定的，所以这里只需要在挂载时设置一次）
  toastMethodsRef.current = {
    success: (message) => addToast('success', message),
    error: (message) => addToast('error', message),
    warning: (message) => addToast('warning', message),
    info: (message) => addToast('info', message),
  };

  // 自动初始化全局 toast 实例（只在挂载时执行一次）
  useEffect(() => {
    setGlobalToast(toastMethodsRef.current);
    return () => {
      globalToast = null;
    };
  }, []);

  return (
    <ToastContext.Provider value={toastMethodsRef.current}>
      {children}
      <div className="fixed right-4 top-4 z-[100] flex flex-col gap-2">
        {toasts.map((toast) => (
          <ToastItem key={toast.id} toast={toast} onRemove={removeToast} />
        ))}
      </div>
    </ToastContext.Provider>
  );
};

// 全局 toast 实例（用于非组件环境）
let globalToast: ToastContextType | null = null;

export const setGlobalToast = (toast: ToastContextType) => {
  globalToast = toast;
};

export const toast = {
  success: (message: string) => globalToast?.success(message) || console.log('✅', message),
  error: (message: string) => globalToast?.error(message) || console.error('❌', message),
  warning: (message: string) => globalToast?.warning(message) || console.warn('⚠️', message),
  info: (message: string) => globalToast?.info(message) || console.info('ℹ️', message),
};