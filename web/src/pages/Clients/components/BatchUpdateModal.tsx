import { useEffect, useState, useMemo } from 'react';
import { Monitor, Cpu, Users, AlertCircle } from 'lucide-react';
import { Modal } from '../../../components/ui/Modal';
import { Select } from '../../../components/ui/Select';
import { Button } from '../../../components/ui/Button';
import { Badge } from '../../../components/ui/Badge';
import { CardRadioGroup } from '../../../components/ui/CardRadioGroup';
import { clientApi } from '../../../api/client';
import type { Client } from '../../../types';
import type { GithubMirror } from '../../../api/githubMirror';

interface BatchUpdateModalProps {
  visible: boolean;
  clients: Client[];
  selectedRowKeys: number[];
  githubMirrors: GithubMirror[];
  updateType: 'frpc' | 'daemon';
  updateMirrorId: number | undefined;
  onUpdateTypeChange: (type: 'frpc' | 'daemon') => void;
  onMirrorIdChange: (id: number | undefined) => void;
  onClose: () => void;
  onSuccess: () => void;
}

export function BatchUpdateModal({
  visible,
  clients,
  selectedRowKeys,
  githubMirrors,
  updateType,
  updateMirrorId,
  onUpdateTypeChange,
  onMirrorIdChange,
  onClose,
  onSuccess,
}: BatchUpdateModalProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 设置默认镜像源
  useEffect(() => {
    if (visible && !updateMirrorId) {
      const defaultMirror = githubMirrors.find(m => m.is_default && m.enabled);
      if (defaultMirror) {
        onMirrorIdChange(defaultMirror.id);
      }
    }
    if (visible) {
      setError(null);
    }
  }, [visible, githubMirrors, updateMirrorId, onMirrorIdChange]);

  const selectedClients = clients.filter(c => selectedRowKeys.includes(c.id));

  const updateTypeOptions = useMemo(() => [
    { value: 'frpc' as const, label: '更新 frpc (与服务端版本同步)', icon: Monitor },
    { value: 'daemon' as const, label: '更新 ws-daemon', icon: Cpu },
  ], []);

  const handleBatchUpdate = async () => {
    if (selectedRowKeys.length === 0) {
      setError('请选择要更新的客户端');
      return;
    }

    // 检查是否都WS连接
    const offlineClients = selectedClients.filter(c => !c.ws_connected);
    if (offlineClients.length > 0) {
      setError(`以下客户端WS未连接，无法更新: ${offlineClients.map(c => c.name).join(', ')}`);
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const result = await clientApi.batchUpdateClients({
        client_ids: selectedRowKeys,
        update_type: updateType,
        mirror_id: updateMirrorId,
      });
      
      if (result.failed_clients && result.failed_clients.length > 0) {
        setError(`部分客户端更新失败: ${result.failed_clients.join(', ')}`);
      }
      
      onClose();
      onSuccess();
    } catch {
      setError('批量更新失败');
    } finally {
      setLoading(false);
    }
  };

  const mirrorOptions = githubMirrors.map(mirror => ({
    value: mirror.id as number,
    label: `${mirror.name}${mirror.is_default ? ' (默认)' : ''}`,
  }));

  return (
    <Modal
      open={visible}
      onClose={onClose}
      title="批量更新客户端"
      size="md"
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>
            取消
          </Button>
          <Button onClick={handleBatchUpdate} loading={loading}>
            开始批量更新
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        {/* 已选择的客户端 */}
        <div>
          <div className="flex items-center gap-2 text-sm mb-2 text-foreground-secondary">
            <Users className="h-4 w-4" />
            <span>已选择 {selectedRowKeys.length} 个客户端</span>
          </div>
          <div className="flex flex-wrap gap-2 max-h-32 overflow-y-auto p-2 rounded-lg border bg-card-bg border-card-border">
            {selectedClients.map(c => (
              <Badge
                key={c.id}
                variant={c.ws_connected ? 'success' : 'default'}
                size="sm"
              >
                {c.name}
              </Badge>
            ))}
          </div>
        </div>

        {/* 更新类型选择 */}
        <div>
          <label className="block text-sm font-medium mb-2 text-foreground-secondary">选择更新类型:</label>
          <CardRadioGroup
            name="batch_update_type"
            value={updateType}
            onChange={onUpdateTypeChange}
            options={updateTypeOptions}
          />
        </div>

        {/* 镜像源选择 */}
        {updateType === 'frpc' && (
          <Select
            label="下载加速源"
            placeholder="选择下载加速源"
            options={mirrorOptions}
            value={updateMirrorId}
            onChange={(v) => onMirrorIdChange(v as number)}
          />
        )}

        {/* 错误提示 */}
        {error && (
          <div className="flex items-start gap-2 p-3 bg-red-500/10 border border-red-500/30 rounded-lg">
            <AlertCircle className="h-5 w-5 text-red-400 flex-shrink-0 mt-0.5" />
            <span className="text-sm text-red-300">{error}</span>
          </div>
        )}
      </div>
    </Modal>
  );
}