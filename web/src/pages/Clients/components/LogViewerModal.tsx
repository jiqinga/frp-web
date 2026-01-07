import { useState, useEffect, useRef, useMemo } from 'react';
import { Modal } from '../../../components/ui/Modal';
import { Button } from '../../../components/ui/Button';
import { Play, Square, Trash2, Search, Download } from 'lucide-react';
import type { LogType } from '../hooks/useLogStream';
import { cn } from '../../../utils/cn';

interface LogViewerModalProps {
  open: boolean;
  onClose: () => void;
  clientName: string;
  logs: string[];
  isStreaming: boolean;
  logType: LogType | null;
  onStartStream: (logType: LogType, lines: number) => void;
  onStopStream: () => void;
  onClearLogs: () => void;
}

const LINE_OPTIONS = [100, 200, 300, 500];

export function LogViewerModal({
  open,
  onClose,
  clientName,
  logs,
  isStreaming,
  logType,
  onStartStream,
  onStopStream,
  onClearLogs,
}: LogViewerModalProps) {
  const [selectedLogType, setSelectedLogType] = useState<LogType>('frpc');
  const [selectedLines, setSelectedLines] = useState(100);
  const [searchKeyword, setSearchKeyword] = useState('');
  const logContainerRef = useRef<HTMLDivElement>(null);
  const [autoScroll, setAutoScroll] = useState(true);

  // 弹窗打开时重置搜索状态
  useEffect(() => {
    if (open) {
      setSearchKeyword('');
    }
  }, [open]);

  // 自动滚动到底部
  useEffect(() => {
    if (autoScroll && logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  // 关闭时停止流并清空日志
  const handleClose = () => {
    if (isStreaming) {
      onStopStream();
    }
    onClearLogs();
    onClose();
  };

  // 高亮搜索关键词
  const highlightedLogs = useMemo(() => {
    if (!searchKeyword.trim()) return logs;
    const regex = new RegExp(`(${searchKeyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
    return logs.map(line => line.replace(regex, '<mark class="bg-yellow-400/60 text-yellow-900 dark:bg-yellow-500/50 dark:text-yellow-200 rounded px-0.5">$1</mark>'));
  }, [logs, searchKeyword]);

  // 导出日志
  const handleExport = () => {
    const content = logs.join('\n');
    const blob = new Blob([content], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${clientName}-${selectedLogType}-${new Date().toISOString().slice(0, 10)}.log`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <Modal
      open={open}
      onClose={handleClose}
      title={`日志查看 - ${clientName}`}
      size="4xl"
      contentClassName="p-0"
    >
      <div className="flex flex-col h-[70vh]">
        {/* 工具栏 */}
        <div className="flex flex-wrap items-center gap-4 p-4 border-b border-border bg-surface-hover">
          {/* 日志类型选择 */}
          <div className="flex items-center gap-2">
            <span className="text-sm text-foreground-muted">日志类型:</span>
            <select
              value={selectedLogType}
              onChange={(e) => setSelectedLogType(e.target.value as LogType)}
              disabled={isStreaming}
              className="px-3 py-1.5 text-sm bg-surface border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
            >
              <option value="frpc">frpc日志</option>
              <option value="daemon">daemon日志</option>
            </select>
          </div>

          {/* 行数选择 */}
          <div className="flex items-center gap-2">
            <span className="text-sm text-foreground-muted">显示行数:</span>
            <select
              value={selectedLines}
              onChange={(e) => setSelectedLines(Number(e.target.value))}
              disabled={isStreaming}
              className="px-3 py-1.5 text-sm bg-surface border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
            >
              {LINE_OPTIONS.map(n => (
                <option key={n} value={n}>{n}行</option>
              ))}
            </select>
          </div>

          {/* 搜索框 */}
          <div className="flex items-center gap-2 shrink-0">
            <Search className="h-4 w-4 text-foreground-muted shrink-0" />
            <input
              type="text"
              value={searchKeyword}
              onChange={(e) => setSearchKeyword(e.target.value)}
              placeholder="搜索关键词..."
              className="w-40 px-3 py-1.5 text-sm bg-surface border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
            />
          </div>

          {/* 操作按钮 */}
          <div className="flex items-center gap-2 shrink-0">
            {isStreaming ? (
              <Button
                size="sm"
                variant="danger"
                icon={<Square className="h-4 w-4" />}
                onClick={onStopStream}
              >
                停止
              </Button>
            ) : (
              <Button
                size="sm"
                variant="primary"
                icon={<Play className="h-4 w-4" />}
                onClick={() => onStartStream(selectedLogType, selectedLines)}
              >
                开始
              </Button>
            )}
            <Button
              size="sm"
              variant="ghost"
              icon={<Trash2 className="h-4 w-4" />}
              onClick={onClearLogs}
              disabled={logs.length === 0}
            >
              清空
            </Button>
            <Button
              size="sm"
              variant="ghost"
              icon={<Download className="h-4 w-4" />}
              onClick={handleExport}
              disabled={logs.length === 0}
            >
              导出
            </Button>
          </div>
        </div>

        {/* 状态栏 */}
        <div className="flex items-center justify-between px-4 py-2 border-b border-border text-xs text-foreground-muted">
          <div className="flex items-center gap-4">
            <span>共 {logs.length} 行</span>
            {isStreaming && (
              <span className="flex items-center gap-1">
                <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
                正在接收 {logType === 'frpc' ? 'frpc' : 'daemon'} 日志...
              </span>
            )}
          </div>
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={autoScroll}
              onChange={(e) => setAutoScroll(e.target.checked)}
              className="rounded"
            />
            <span>自动滚动</span>
          </label>
        </div>

        {/* 日志内容 */}
        <div
          ref={logContainerRef}
          className={cn(
            "flex-1 overflow-auto p-4 font-mono text-xs leading-relaxed",
            "bg-surface-elevated text-foreground-secondary"
          )}
        >
          {logs.length === 0 ? (
            <div className="flex items-center justify-center h-full text-foreground-muted">
              {isStreaming ? '等待日志数据...' : '点击"开始"按钮开始接收日志'}
            </div>
          ) : (
            <div className="space-y-0.5">
              {highlightedLogs.map((line, index) => (
                <div
                  key={index}
                  className="hover:bg-hover-overlay px-2 py-0.5 rounded"
                  dangerouslySetInnerHTML={{ __html: line }}
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </Modal>
  );
}