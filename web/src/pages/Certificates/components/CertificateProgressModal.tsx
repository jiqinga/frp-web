import { CheckCircle, Circle, Loader2, XCircle } from 'lucide-react';
import { Modal, Button } from '../../../components/ui';

export type CertProgressStep = 'validating' | 'adding_dns' | 'waiting_dns' | 'requesting' | 'saving' | 'completed' | 'failed';

export interface CertProgressState {
  taskId: string;
  domain: string;
  step: CertProgressStep;
  message?: string;
  error?: string;
}

interface CertificateProgressModalProps {
  visible: boolean;
  progress: CertProgressState | null;
  onClose: () => void;
}

const STEPS: { key: CertProgressStep; label: string }[] = [
  { key: 'validating', label: '验证DNS提供商配置' },
  { key: 'adding_dns', label: '添加DNS TXT验证记录' },
  { key: 'waiting_dns', label: '等待DNS记录生效' },
  { key: 'requesting', label: '向Let\'s Encrypt申请证书' },
  { key: 'saving', label: '保存证书' },
];

function getStepIndex(step: CertProgressStep): number {
  if (step === 'completed') return STEPS.length;
  if (step === 'failed') return -1;
  return STEPS.findIndex(s => s.key === step);
}

export function CertificateProgressModal({ visible, progress, onClose }: CertificateProgressModalProps) {
  const currentIndex = progress ? getStepIndex(progress.step) : -1;
  const isFailed = progress?.step === 'failed';
  const isCompleted = progress?.step === 'completed';
  const canClose = isFailed || isCompleted;

  return (
    <Modal open={visible} onClose={canClose ? onClose : () => {}} title="申请SSL证书">
      <div className="space-y-4">
        {progress && (
          <div className="p-3 bg-background-secondary rounded-md">
            <p className="text-xs text-foreground-secondary mb-1">证书域名：</p>
            <p className="font-mono text-sm text-foreground">{progress.domain}</p>
          </div>
        )}

        <div className="space-y-3">
          {STEPS.map((step, index) => {
            const isActive = index === currentIndex;
            const isDone = currentIndex > index || isCompleted;
            const isPending = currentIndex < index && !isFailed;

            return (
              <div key={step.key} className="flex items-center gap-3">
                {isDone && <CheckCircle className="h-5 w-5 text-green-500 flex-shrink-0" />}
                {isActive && !isFailed && <Loader2 className="h-5 w-5 text-blue-500 animate-spin flex-shrink-0" />}
                {isActive && isFailed && <XCircle className="h-5 w-5 text-red-500 flex-shrink-0" />}
                {isPending && <Circle className="h-5 w-5 text-foreground-muted flex-shrink-0" />}
                <span className={`text-sm ${isActive ? 'text-foreground font-medium' : isDone ? 'text-foreground-secondary' : 'text-foreground-muted'}`}>
                  {step.label}
                </span>
              </div>
            );
          })}
        </div>

        {progress?.message && !isFailed && (
          <p className="text-xs text-foreground-secondary">{progress.message}</p>
        )}

        {isFailed && progress?.error && (
          <div className="p-3 bg-red-500/10 border border-red-500/30 rounded-md">
            <p className="text-sm text-red-500">{progress.error}</p>
          </div>
        )}

        {isCompleted && (
          <div className="p-3 bg-green-500/10 border border-green-500/30 rounded-md">
            <p className="text-sm text-green-500">证书申请成功！</p>
          </div>
        )}

        <div className="flex justify-end pt-2">
          <Button variant={canClose ? 'primary' : 'secondary'} onClick={onClose} disabled={!canClose}>
            {canClose ? '关闭' : '申请中...'}
          </Button>
        </div>
      </div>
    </Modal>
  );
}