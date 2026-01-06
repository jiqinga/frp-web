import { useState, useCallback } from 'react';

export type OperationType = 
  | 'testSSH' | 'start' | 'stop' | 'restart' 
  | 'refreshVersion' | 'viewLogs' | 'viewMetrics'
  | 'testConnection' | 'refreshLocalVersion';

export type LoadingOperations = Map<number, Set<OperationType>>;

/**
 * 操作 loading 状态管理 Hook
 */
export function useOperationLoading() {
  const [loadingOperations, setLoadingOperations] = useState<LoadingOperations>(new Map());

  const startLoading = useCallback((serverId: number, operation: OperationType) => {
    setLoadingOperations(prev => {
      const next = new Map(prev);
      const ops = next.get(serverId) || new Set();
      ops.add(operation);
      next.set(serverId, ops);
      return next;
    });
  }, []);

  const stopLoading = useCallback((serverId: number, operation: OperationType) => {
    setLoadingOperations(prev => {
      const next = new Map(prev);
      const ops = next.get(serverId);
      if (ops) {
        ops.delete(operation);
        if (ops.size === 0) next.delete(serverId);
        else next.set(serverId, ops);
      }
      return next;
    });
  }, []);

  const isLoading = useCallback((serverId: number, operation: OperationType) => {
    return loadingOperations.get(serverId)?.has(operation) ?? false;
  }, [loadingOperations]);

  const withLoading = useCallback(<T,>(
    serverId: number,
    operation: OperationType,
    fn: () => Promise<T>
  ): Promise<T> => {
    startLoading(serverId, operation);
    return fn().finally(() => stopLoading(serverId, operation));
  }, [startLoading, stopLoading]);

  return { loadingOperations, isLoading, withLoading };
}