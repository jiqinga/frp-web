import { type ReactNode, useState } from 'react';
import { cn } from '../../utils/cn';
import { Loader2, ChevronDown, ChevronUp } from 'lucide-react';

export interface Column<T> {
  key: string;
  title: string;
  dataIndex?: keyof T;
  width?: number | string;
  align?: 'left' | 'center' | 'right';
  render?: (value: unknown, record: T, index: number) => ReactNode;
  /** 是否在移动端隐藏此列 */
  hideOnMobile?: boolean;
  /** 是否在平板端隐藏此列 */
  hideOnTablet?: boolean;
  /** 是否为主要列（移动端卡片视图中显示） */
  primary?: boolean;
}

export interface TableProps<T> {
  columns: Column<T>[];
  data: T[];
  rowKey: keyof T | ((record: T) => string | number);
  loading?: boolean;
  emptyText?: string;
  className?: string;
  onRowClick?: (record: T) => void;
  rowClassName?: (record: T, index: number) => string;
  size?: 'sm' | 'md' | 'lg';
  /** 是否启用移动端卡片视图 */
  mobileCardView?: boolean;
  /** 移动端卡片视图的标题列 key */
  cardTitleKey?: string;
  /** 移动端卡片视图的副标题列 key */
  cardSubtitleKey?: string;
}

export function Table<T>({
  columns,
  data,
  rowKey,
  loading = false,
  emptyText = '暂无数据',
  className,
  onRowClick,
  rowClassName,
  size = 'md',
  mobileCardView = false,
  cardTitleKey,
  cardSubtitleKey,
}: TableProps<T>) {
  const [expandedCards, setExpandedCards] = useState<Set<string | number>>(new Set());

  const getRowKey = (record: T): string | number => {
    if (typeof rowKey === 'function') {
      return rowKey(record);
    }
    return record[rowKey] as string | number;
  };

  const getCellValue = (record: T, column: Column<T>, index: number): ReactNode => {
    if (column.render) {
      const value = column.dataIndex ? record[column.dataIndex] : undefined;
      return column.render(value, record, index);
    }
    if (column.dataIndex) {
      return record[column.dataIndex] as ReactNode;
    }
    return null;
  };

  const toggleCardExpand = (key: string | number) => {
    setExpandedCards(prev => {
      const next = new Set(prev);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
      }
      return next;
    });
  };

  const sizeClasses = {
    sm: 'text-xs',
    md: 'text-sm',
    lg: 'text-base',
  };

  const paddingClasses = {
    sm: 'px-2 py-1.5 sm:px-3 sm:py-2',
    md: 'px-3 py-2 sm:px-4 sm:py-3',
    lg: 'px-4 py-3 sm:px-6 sm:py-4',
  };

  // 获取可见列（根据响应式设置）
  const getVisibleColumns = () => {
    return columns.map(col => ({
      ...col,
      className: cn(
        col.hideOnMobile && 'hidden sm:table-cell',
        col.hideOnTablet && 'hidden md:table-cell'
      )
    }));
  };

  const visibleColumns = getVisibleColumns();

  // 移动端卡片视图渲染
  const renderMobileCard = (record: T, index: number) => {
    const key = getRowKey(record);
    const isExpanded = expandedCards.has(key);
    const titleColumn = columns.find(c => c.key === cardTitleKey);
    const subtitleColumn = columns.find(c => c.key === cardSubtitleKey);
    const primaryColumns = columns.filter(c => c.primary);
    const secondaryColumns = columns.filter(c => !c.primary && c.key !== cardTitleKey && c.key !== cardSubtitleKey);

    return (
      <div
        key={key}
        className={cn(
          'rounded-lg border p-4 space-y-3 transition-colors',
          'bg-surface/80 border-border-subtle hover:bg-surface-hover',
          onRowClick && 'cursor-pointer',
          rowClassName?.(record, index)
        )}
        onClick={() => onRowClick?.(record)}
      >
        {/* 卡片头部 */}
        <div className="flex items-start justify-between gap-3">
          <div className="flex-1 min-w-0">
            {titleColumn && (
              <div className="font-medium truncate text-foreground">
                {getCellValue(record, titleColumn, index)}
              </div>
            )}
            {subtitleColumn && (
              <div className="text-sm mt-0.5 text-foreground-muted">
                {getCellValue(record, subtitleColumn, index)}
              </div>
            )}
          </div>
          {secondaryColumns.length > 0 && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                toggleCardExpand(key);
              }}
              className="p-1 transition-colors text-foreground-subtle hover:text-foreground"
            >
              {isExpanded ? (
                <ChevronUp className="h-5 w-5" />
              ) : (
                <ChevronDown className="h-5 w-5" />
              )}
            </button>
          )}
        </div>

        {/* 主要信息 */}
        {primaryColumns.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {primaryColumns.map(col => (
              <div key={col.key} className="text-sm">
                {getCellValue(record, col, index)}
              </div>
            ))}
          </div>
        )}

        {/* 展开的详细信息 */}
        {isExpanded && secondaryColumns.length > 0 && (
          <div className="pt-3 border-t space-y-2 border-border-subtle">
            {secondaryColumns.map(col => (
              <div key={col.key} className="flex items-start gap-2 text-sm">
                <span className="shrink-0 text-foreground-subtle">{col.title}:</span>
                <span className="text-foreground-secondary">{getCellValue(record, col, index)}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    );
  };

  // 加载状态
  if (loading) {
    return (
      <div className={cn('w-full overflow-hidden rounded-lg border border-border', className)}>
        <div className="flex items-center justify-center gap-2 py-12 text-foreground-muted">
          <Loader2 className="h-5 w-5 animate-spin" />
          <span>加载中...</span>
        </div>
      </div>
    );
  }

  // 空状态
  if (data.length === 0) {
    return (
      <div className={cn('w-full overflow-hidden rounded-lg border border-border', className)}>
        <div className="py-12 text-center text-foreground-muted">{emptyText}</div>
      </div>
    );
  }

  return (
    <div className={cn('w-full', className)}>
      {/* 移动端卡片视图 */}
      {mobileCardView && (
        <div className="sm:hidden space-y-3">
          {data.map((record, index) => renderMobileCard(record, index))}
        </div>
      )}

      {/* 桌面端表格视图 */}
      <div className={cn(
        'overflow-hidden rounded-lg border border-border',
        mobileCardView && 'hidden sm:block'
      )}>
        <div className="overflow-x-auto scrollbar-custom">
          <table className={cn('w-full min-w-[640px]', sizeClasses[size])}>
            <thead>
              <tr className="border-b bg-surface-hover border-border">
                {visibleColumns.map((column) => (
                  <th
                    key={column.key}
                    className={cn(
                      paddingClasses[size],
                      'font-semibold whitespace-nowrap text-center text-foreground-secondary',
                      column.align === 'left' && 'text-left',
                      column.align === 'right' && 'text-right',
                      column.className
                    )}
                    style={{ width: column.width }}
                  >
                    {column.title}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-border-subtle">
              {data.map((record, index) => (
                <tr
                  key={getRowKey(record)}
                  className={cn(
                    'transition-colors bg-surface/30 hover:bg-surface-hover',
                    onRowClick && 'cursor-pointer',
                    rowClassName?.(record, index)
                  )}
                  onClick={() => onRowClick?.(record)}
                >
                  {visibleColumns.map((column) => (
                    <td
                      key={column.key}
                      className={cn(
                        paddingClasses[size],
                        'text-center text-foreground-secondary',
                        column.align === 'left' && 'text-left',
                        column.align === 'right' && 'text-right',
                        column.className
                      )}
                    >
                      {getCellValue(record, column, index)}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}