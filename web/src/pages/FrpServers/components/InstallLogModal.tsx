import { useRef, useEffect } from 'react';
import { Terminal, X } from 'lucide-react';
import { Modal } from '../../../components/ui/Modal';
import { Button } from '../../../components/ui/Button';
import { useThemeStore } from '../../../store/theme';
import type { EnhancedLogEntry } from '../../../types';

interface InstallLogModalProps {
  visible: boolean;
  logs: EnhancedLogEntry[];
  downloadProgress: number;
  onClose: () => void;
}

/**
 * 安装日志弹窗组件
 * 显示远程操作（安装/重装/升级）的实时日志
 * 使用 Tailwind CSS + Lucide Icons 重构
 */
export function InstallLogModal({
  visible,
  logs,
  downloadProgress,
  onClose,
}: InstallLogModalProps) {
  const { theme } = useThemeStore();
  const isLight = theme === 'light';

  // 日志容器引用，用于自动滚动
  const logContainerRef = useRef<HTMLDivElement>(null);
  const logEndRef = useRef<HTMLDivElement>(null);

  // 自动滚动到底部
  useEffect(() => {
    if (logEndRef.current) {
      logEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [logs]);

  // 获取日志颜色
  const getLogColor = (type: string): string => {
    switch (type) {
      case 'error':
        return 'text-red-500 dark:text-red-400';
      case 'success':
        return 'text-green-500 dark:text-green-400';
      case 'warning':
        return 'text-yellow-500 dark:text-yellow-400';
      case 'info':
        return 'text-blue-500 dark:text-blue-400';
      case 'progress':
        return 'text-cyan-500 dark:text-cyan-400';
      default:
        return 'text-foreground-secondary';
    }
  };

  // 获取日志图标
  const getLogIcon = (type: string): string => {
    switch (type) {
      case 'error':
        return '✗';
      case 'success':
        return '✓';
      case 'warning':
        return '⚠';
      case 'info':
        return 'ℹ';
      case 'progress':
        return '→';
      default:
        return '•';
    }
  };

  return (
    <Modal
      open={visible}
      onClose={onClose}
      title={
        <div className="flex items-center gap-2">
          <Terminal className="h-5 w-5 text-indigo-400" />
          <span>远程操作日志</span>
        </div>
      }
      className="w-[70vw] max-w-[70vw]"
      footer={
        <div className="flex justify-end">
          <Button variant="ghost" onClick={onClose} icon={<X className="h-4 w-4" />}>
            关闭
          </Button>
        </div>
      }
    >
      <div className="space-y-4">
        {/* 下载进度条 */}
        {downloadProgress > 0 && (
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="text-foreground-muted">下载进度</span>
              <span className="text-indigo-500 font-medium">{downloadProgress}%</span>
            </div>
            <div className="h-2 rounded-full overflow-hidden bg-surface-active">
              <div
                className="h-full bg-gradient-to-r from-indigo-500 via-purple-500 to-cyan-500 transition-all duration-300 ease-out"
                style={{ width: `${downloadProgress}%` }}
              />
            </div>
          </div>
        )}

        {/* 日志终端 */}
        <div className="relative">
          {/* 终端头部 */}
          <div className="flex items-center gap-2 px-4 py-2 border border-b-0 rounded-t-lg bg-surface-elevated border-border">
            <div className="flex gap-1.5">
              <div className="w-3 h-3 rounded-full bg-red-500/80" />
              <div className="w-3 h-3 rounded-full bg-yellow-500/80" />
              <div className="w-3 h-3 rounded-full bg-green-500/80" />
            </div>
            <span className="text-xs ml-2 text-foreground-muted">终端</span>
          </div>

          {/* 日志内容 */}
          <div
            ref={logContainerRef}
            className="border border-t-0 rounded-b-lg p-4 max-h-[400px] overflow-y-auto font-mono text-sm bg-surface border-border"
            style={{
              backgroundImage: isLight
                ? 'linear-gradient(rgba(99, 102, 241, 0.03) 1px, transparent 1px)'
                : 'linear-gradient(rgba(99, 102, 241, 0.02) 1px, transparent 1px)',
              backgroundSize: '100% 24px',
            }}
          >
            {logs.length === 0 ? (
              <div className="text-center py-8 text-foreground-muted">
                <Terminal className="h-8 w-8 mx-auto mb-2 opacity-50" />
                <p>等待日志输出...</p>
              </div>
            ) : (
              logs.map((entry, index) => (
                <div
                  key={index}
                  className={`flex items-start gap-2 mb-1 ${getLogColor(entry.type)}`}
                >
                  {/* 时间戳 */}
                  <span className="shrink-0 text-cyan-500/70 dark:text-cyan-500/70">
                    [{entry.timestamp}]
                  </span>
                  {/* 图标 */}
                  <span className="shrink-0 w-4 text-center">
                    {getLogIcon(entry.type)}
                  </span>
                  {/* 消息 */}
                  <span className="break-all">{entry.message}</span>
                </div>
              ))
            )}
            <div ref={logEndRef} />
          </div>

          {/* 扫描线效果 */}
          <div className="absolute inset-0 pointer-events-none overflow-hidden rounded-lg">
            <div className="absolute inset-0 bg-gradient-to-b from-transparent via-indigo-500/5 to-transparent h-[200%] animate-scan" />
          </div>
        </div>

        {/* 日志统计 */}
        <div className="flex items-center justify-between text-xs text-foreground-muted">
          <span>共 {logs.length} 条日志</span>
          <div className="flex items-center gap-4">
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-green-400" />
              成功: {logs.filter(l => l.type === 'success').length}
            </span>
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-red-400" />
              错误: {logs.filter(l => l.type === 'error').length}
            </span>
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-blue-400" />
              信息: {logs.filter(l => l.type === 'info').length}
            </span>
          </div>
        </div>
      </div>
    </Modal>
  );
}