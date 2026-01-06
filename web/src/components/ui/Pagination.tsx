import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react';
import { cn } from '../../utils/cn';
import { Button } from './Button';

export interface PaginationProps {
  current: number;
  total: number;
  pageSize?: number;
  onChange: (page: number) => void;
  showTotal?: boolean;
  showQuickJumper?: boolean;
  className?: string;
}

export function Pagination({
  current,
  total,
  pageSize = 10,
  onChange,
  showTotal = true,
  className,
}: PaginationProps) {
  const totalPages = Math.ceil(total / pageSize);

  const getPageNumbers = () => {
    const pages: (number | string)[] = [];
    const showPages = 5;
    
    if (totalPages <= showPages + 2) {
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      pages.push(1);
      
      if (current > 3) {
        pages.push('...');
      }
      
      const start = Math.max(2, current - 1);
      const end = Math.min(totalPages - 1, current + 1);
      
      for (let i = start; i <= end; i++) {
        pages.push(i);
      }
      
      if (current < totalPages - 2) {
        pages.push('...');
      }
      
      pages.push(totalPages);
    }
    
    return pages;
  };

  if (totalPages <= 1) return null;

  return (
    <div className={cn('flex items-center justify-between gap-4', className)}>
      {showTotal && (
        <span className="text-sm text-foreground-muted">
          共 <span className="font-medium text-foreground-secondary">{total}</span> 条
        </span>
      )}
      
      <div className="flex items-center gap-1">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onChange(1)}
          disabled={current === 1}
          icon={<ChevronsLeft className="h-4 w-4" />}
        />
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onChange(current - 1)}
          disabled={current === 1}
          icon={<ChevronLeft className="h-4 w-4" />}
        />
        
        {getPageNumbers().map((page, index) => (
          typeof page === 'number' ? (
            <button
              key={index}
              onClick={() => onChange(page)}
              className={cn(
                'min-w-[32px] h-8 px-2 text-sm font-medium rounded-md transition-colors',
                page === current
                  ? 'bg-indigo-600 text-white'
                  : 'text-foreground-muted hover:text-foreground hover:bg-surface-hover'
              )}
            >
              {page}
            </button>
          ) : (
            <span key={index} className="px-2 text-foreground-subtle">
              {page}
            </span>
          )
        ))}
        
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onChange(current + 1)}
          disabled={current === totalPages}
          icon={<ChevronRight className="h-4 w-4" />}
        />
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onChange(totalPages)}
          disabled={current === totalPages}
          icon={<ChevronsRight className="h-4 w-4" />}
        />
      </div>
    </div>
  );
}