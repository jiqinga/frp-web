import { Spinner } from './ui';

export const LoadingFallback = () => {
  return (
    <div className="flex items-center justify-center h-screen bg-background">
      <div className="flex flex-col items-center gap-4">
        <Spinner size="lg" />
        <span className="text-sm text-foreground-muted">加载中...</span>
      </div>
    </div>
  );
};