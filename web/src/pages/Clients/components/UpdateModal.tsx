import { useEffect, useMemo } from 'react';
import { Download, CheckCircle, XCircle, Loader2, Monitor, Cpu } from 'lucide-react';
import { Modal } from '../../../components/ui/Modal';
import { Select } from '../../../components/ui/Select';
import { Button } from '../../../components/ui/Button';
import { Badge } from '../../../components/ui/Badge';
import { CardRadioGroup } from '../../../components/ui/CardRadioGroup';
import type { Client } from '../../../types';
import type { GithubMirror } from '../../../api/githubMirror';
import type { UpdateProgress, UpdateResult } from '../hooks/useUpdateWebSocket';
import { STAGE_LABELS } from '../hooks/useUpdateWebSocket';

interface UpdateModalProps {
  visible: boolean;
  client: Client | null;
  githubMirrors: GithubMirror[];
  updateType: 'frpc' | 'daemon';
  updateMirrorId: number | undefined;
  updateProgress: UpdateProgress | null;
  updateResult: UpdateResult | null;
  onUpdateTypeChange: (type: 'frpc' | 'daemon') => void;
  onMirrorIdChange: (id: number | undefined) => void;
  onUpdate: () => void;
  onClose: () => void;
}

export function UpdateModal({
  visible,
  client,
  githubMirrors,
  updateType,
  updateMirrorId,
  updateProgress,
  updateResult,
  onUpdateTypeChange,
  onMirrorIdChange,
  onUpdate,
  onClose,
}: UpdateModalProps) {
  useEffect(() => {
    if (visible && !updateMirrorId) {
      const defaultMirror = githubMirrors.find(m => m.is_default && m.enabled);
      if (defaultMirror) {
        onMirrorIdChange(defaultMirror.id);
      }
    }
  }, [visible, githubMirrors, updateMirrorId, onMirrorIdChange]);

  const isUpdating = !!updateProgress && !updateResult;

  const mirrorOptions = githubMirrors.map(mirror => ({
    value: mirror.id as number,
    label: `${mirror.name}${mirror.is_default ? ' (默认)' : ''}`,
  }));

  const progressPercent = updateProgress?.progress || 0;
  const progressStatus = updateProgress?.stage === 'failed' ? 'error' :
                         updateProgress?.stage === 'completed' ? 'success' : 'active';

  const updateTypeOptions = useMemo(() => [
    { value: 'frpc' as const, label: '更新 frpc (与服务端版本同步)', icon: Monitor },
    { value: 'daemon' as const, label: '更新 ws-daemon', icon: Cpu },
  ], []);

  return (
    <Modal
      open={visible}
      onClose={onClose}
      title={`更新客户端: ${client?.name || ''}`}
      size="md"
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>
            关闭
          </Button>
          <Button
            onClick={onUpdate}
            disabled={isUpdating}
            icon={isUpdating ? <Loader2 className="h-4 w-4 animate-spin" /> : <Download className="h-4 w-4" />}
          >
            开始更新
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        {/* 更新类型选择 */}
        <div>
          <label className="block text-sm font-medium mb-2 text-foreground-secondary">选择更新类型:</label>
          <CardRadioGroup
            name="update_type"
            value={updateType}
            onChange={onUpdateTypeChange}
            options={updateTypeOptions}
            disabled={isUpdating}
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
            disabled={isUpdating}
          />
        )}

        {/* 当前版本信息 */}
        {client && (
          <div className="p-3 rounded-lg border bg-card-bg border-card-border">
            <div className="text-sm font-medium mb-2 text-foreground-secondary">当前版本信息:</div>
            <div className="text-xs space-y-1 text-foreground-muted">
              <div>frpc: <span className="text-foreground-secondary">{client.frpc_version || '未知'}</span></div>
              <div>daemon: <span className="text-foreground-secondary">{client.daemon_version || '未知'}</span></div>
              <div>系统: <span className="text-foreground-secondary">{client.os || '未知'}/{client.arch || '未知'}</span></div>
            </div>
          </div>
        )}

        {/* 更新进度 */}
        {updateProgress && (
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Badge variant={updateProgress.stage === 'failed' ? 'danger' : 'info'}>
                {STAGE_LABELS[updateProgress.stage] || updateProgress.stage}
              </Badge>
              <span className="text-sm text-foreground-secondary">{updateProgress.message}</span>
            </div>
            
            {/* 自定义进度条 */}
            <div className="relative h-2 rounded-full overflow-hidden bg-progress-bg">
              <div
                className={`absolute inset-y-0 left-0 transition-all duration-300 rounded-full ${
                  progressStatus === 'error' ? 'bg-red-500' :
                  progressStatus === 'success' ? 'bg-green-500' :
                  'bg-indigo-500'
                }`}
                style={{ width: `${progressPercent}%` }}
              />
              {progressStatus === 'active' && (
                <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent animate-pulse" />
              )}
            </div>
            
            {updateProgress.totalBytes > 0 && (
              <div className="text-xs text-foreground-muted">
                已下载: {(updateProgress.downloadedBytes / 1024 / 1024).toFixed(2)} MB / {(updateProgress.totalBytes / 1024 / 1024).toFixed(2)} MB
              </div>
            )}
          </div>
        )}

        {/* 更新结果 */}
        {updateResult && (
          <div className={`flex items-start gap-2 p-3 rounded-lg border ${
            updateResult.success 
              ? 'bg-green-500/10 border-green-500/30' 
              : 'bg-red-500/10 border-red-500/30'
          }`}>
            {updateResult.success ? (
              <CheckCircle className="h-5 w-5 text-green-400 flex-shrink-0 mt-0.5" />
            ) : (
              <XCircle className="h-5 w-5 text-red-400 flex-shrink-0 mt-0.5" />
            )}
            <div>
              <div className={`text-sm font-medium ${updateResult.success ? 'text-green-300' : 'text-red-300'}`}>
                {updateResult.success ? '更新成功' : '更新失败'}
              </div>
              <div className="text-sm text-foreground-muted">{updateResult.message}</div>
              {updateResult.success && updateResult.version && (
                <div className="text-sm text-green-400 mt-1">
                  新版本: {updateResult.version}
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </Modal>
  );
}