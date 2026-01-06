import { useState, useCallback } from 'react';
import { toast } from '../../../components/ui/Toast';
import { clientApi } from '../../../api/client';

export type FrpcAction = 'start' | 'stop' | 'restart';

export function useFrpcControl() {
  // 每个客户端的 loading 状态
  const [loadingMap, setLoadingMap] = useState<Record<number, FrpcAction | null>>({});

  // 执行 frpc 控制（HTTP 同步返回结果）
  const controlFrpc = useCallback(async (clientId: number, action: FrpcAction): Promise<boolean> => {
    // 如果已经在 loading，不重复发送
    if (loadingMap[clientId]) {
      return false;
    }

    // 设置 loading 状态
    setLoadingMap(prev => ({ ...prev, [clientId]: action }));

    const actionText = action === 'start' ? '启动' : action === 'stop' ? '停止' : '重启';

    try {
      const response = await clientApi.controlFrpc(clientId, { action });
      // 清除 loading
      setLoadingMap(prev => {
        const next = { ...prev };
        delete next[clientId];
        return next;
      });
      
      if (response.success) {
        toast.success(`frpc ${actionText}成功`);
        return true;
      } else {
        toast.error(response.message || `frpc ${actionText}失败`);
        return false;
      }
    } catch (error: unknown) {
      // 清除 loading
      setLoadingMap(prev => {
        const next = { ...prev };
        delete next[clientId];
        return next;
      });
      
      const errorMessage = error instanceof Error ? error.message : `frpc ${actionText}失败`;
      toast.error(errorMessage);
      return false;
    }
  }, [loadingMap]);

  return {
    controlFrpc,
    loadingMap,
  };
}