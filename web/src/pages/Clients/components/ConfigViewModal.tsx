import { useState } from 'react';
import { Copy, Check } from 'lucide-react';
import { Modal } from '../../../components/ui/Modal';
import { Button } from '../../../components/ui/Button';
import { toast } from '../../../components/ui/Toast';
import { useThemeStore } from '../../../store/theme';

interface ConfigViewModalProps {
  visible: boolean;
  clientName: string;
  config: string;
  loading?: boolean;
  onClose: () => void;
}

export function ConfigViewModal({
  visible,
  clientName,
  config,
  loading = false,
  onClose,
}: ConfigViewModalProps) {
  const { theme } = useThemeStore();
  const isLight = theme === 'light';
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(config);
      setCopied(true);
      toast.success('配置已复制到剪贴板');
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error('复制失败');
    }
  };

  return (
    <Modal
      open={visible}
      onClose={onClose}
      title={`查看配置 - ${clientName}`}
      size="2xl"
      footer={
        <>
          <Button
            variant="secondary"
            icon={copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
            onClick={handleCopy}
            disabled={loading || !config}
          >
            {copied ? '已复制' : '复制配置'}
          </Button>
          <Button onClick={onClose}>关闭</Button>
        </>
      }
    >
      {loading ? (
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-500" />
          <span className={`ml-3 ${isLight ? 'text-slate-500' : 'text-slate-400'}`}>加载中...</span>
        </div>
      ) : config ? (
        <div className="relative">
          <pre className={`rounded-lg p-4 overflow-auto max-h-[60vh] text-sm font-mono whitespace-pre-wrap break-all border ${isLight ? 'bg-slate-100 border-slate-200 text-slate-700' : 'bg-slate-900/50 border-slate-700 text-slate-300'}`}>
            {config}
          </pre>
        </div>
      ) : (
        <div className={`text-center py-12 ${isLight ? 'text-slate-500' : 'text-slate-400'}`}>
          暂无配置内容
        </div>
      )}
    </Modal>
  );
}