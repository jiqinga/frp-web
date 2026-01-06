import { useState, useMemo } from 'react';
import { Search, ChevronRight, ChevronLeft } from 'lucide-react';
import { cn } from '../../utils/cn';
import { Checkbox } from './Checkbox';

export interface TransferItem {
  key: string | number;
  title: string;
  description?: string;
  disabled?: boolean;
}

export interface TransferProps {
  dataSource: TransferItem[];
  targetKeys: (string | number)[];
  onChange: (targetKeys: (string | number)[]) => void;
  titles?: [string, string];
  searchPlaceholder?: string;
  height?: number;
  className?: string;
}

interface PanelProps {
  items: TransferItem[];
  selectedKeys: Set<string | number>;
  onSelectChange: (keys: Set<string | number>) => void;
  title: string;
  searchPlaceholder: string;
  height: number;
}

function Panel({ items, selectedKeys, onSelectChange, title, searchPlaceholder, height }: PanelProps) {
  const [search, setSearch] = useState('');
  
  const filtered = useMemo(() => {
    if (!search) return items;
    const lower = search.toLowerCase();
    return items.filter(i => i.title.toLowerCase().includes(lower) || i.description?.toLowerCase().includes(lower));
  }, [items, search]);

  const enabledItems = filtered.filter(i => !i.disabled);
  const allSelected = enabledItems.length > 0 && enabledItems.every(i => selectedKeys.has(i.key));
  const someSelected = enabledItems.some(i => selectedKeys.has(i.key));

  const toggleAll = () => {
    if (allSelected) {
      const newKeys = new Set(selectedKeys);
      enabledItems.forEach(i => newKeys.delete(i.key));
      onSelectChange(newKeys);
    } else {
      const newKeys = new Set(selectedKeys);
      enabledItems.forEach(i => newKeys.add(i.key));
      onSelectChange(newKeys);
    }
  };

  const toggle = (key: string | number) => {
    const newKeys = new Set(selectedKeys);
    if (newKeys.has(key)) newKeys.delete(key);
    else newKeys.add(key);
    onSelectChange(newKeys);
  };

  return (
    <div className="flex-1 flex flex-col rounded-lg overflow-hidden bg-surface border border-border">
      <div className="px-3 py-2 border-b border-border-subtle flex items-center justify-between">
        <div className="flex items-center gap-2 cursor-pointer" onClick={toggleAll}>
          <Checkbox
            checked={allSelected}
            indeterminate={someSelected && !allSelected}
            onChange={toggleAll}
            size="sm"
          />
          <span className="text-sm text-foreground-secondary">{title}</span>
        </div>
        <span className="text-xs text-foreground-muted">{selectedKeys.size}/{items.length}</span>
      </div>
      <div className="px-2 py-2 border-b border-border-subtle">
        <div className="relative">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-4 w-4 text-foreground-subtle" />
          <input
            type="text"
            value={search}
            onChange={e => setSearch(e.target.value)}
            placeholder={searchPlaceholder}
            className="w-full pl-8 pr-3 py-1.5 text-sm rounded focus:outline-none focus:border-indigo-500 bg-input-bg border border-input-border text-foreground placeholder-input-placeholder"
          />
        </div>
      </div>
      <div className="flex-1 overflow-y-auto" style={{ maxHeight: height }}>
        {filtered.length === 0 ? (
          <div className="p-4 text-center text-sm text-foreground-muted">无数据</div>
        ) : (
          filtered.map(item => (
            <div
              key={item.key}
              className={cn(
                'flex items-center gap-2 px-3 py-2 cursor-pointer hover:bg-surface-hover',
                item.disabled && 'opacity-50 cursor-not-allowed'
              )}
              onClick={() => !item.disabled && toggle(item.key)}
            >
              <Checkbox
                checked={selectedKeys.has(item.key)}
                onChange={() => !item.disabled && toggle(item.key)}
                disabled={item.disabled}
                size="sm"
              />
              <div className="flex-1 min-w-0">
                <div className="text-sm truncate text-foreground">{item.title}</div>
                {item.description && <div className="text-xs truncate text-foreground-muted">{item.description}</div>}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}

export function Transfer({
  dataSource,
  targetKeys,
  onChange,
  titles = ['可选', '已选'],
  searchPlaceholder = '搜索...',
  height = 200,
  className
}: TransferProps) {
  const [leftSelected, setLeftSelected] = useState<Set<string | number>>(new Set());
  const [rightSelected, setRightSelected] = useState<Set<string | number>>(new Set());

  const targetSet = useMemo(() => new Set(targetKeys), [targetKeys]);
  const leftItems = useMemo(() => dataSource.filter(i => !targetSet.has(i.key)), [dataSource, targetSet]);
  const rightItems = useMemo(() => dataSource.filter(i => targetSet.has(i.key)), [dataSource, targetSet]);

  const moveRight = () => {
    const newTargets = [...targetKeys, ...leftSelected];
    onChange(newTargets);
    setLeftSelected(new Set());
  };

  const moveLeft = () => {
    const newTargets = targetKeys.filter(k => !rightSelected.has(k));
    onChange(newTargets);
    setRightSelected(new Set());
  };

  return (
    <div className={cn('flex gap-2', className)}>
      <Panel
        items={leftItems}
        selectedKeys={leftSelected}
        onSelectChange={setLeftSelected}
        title={titles[0]}
        searchPlaceholder={searchPlaceholder}
        height={height}
      />
      <div className="flex flex-col justify-center gap-2">
        <button
          onClick={moveRight}
          disabled={leftSelected.size === 0}
          className="p-2 disabled:opacity-50 disabled:cursor-not-allowed rounded transition-colors bg-surface-hover hover:bg-surface-active"
        >
          <ChevronRight className="h-4 w-4 text-foreground" />
        </button>
        <button
          onClick={moveLeft}
          disabled={rightSelected.size === 0}
          className="p-2 disabled:opacity-50 disabled:cursor-not-allowed rounded transition-colors bg-surface-hover hover:bg-surface-active"
        >
          <ChevronLeft className="h-4 w-4 text-foreground" />
        </button>
      </div>
      <Panel
        items={rightItems}
        selectedKeys={rightSelected}
        onSelectChange={setRightSelected}
        title={titles[1]}
        searchPlaceholder={searchPlaceholder}
        height={height}
      />
    </div>
  );
}