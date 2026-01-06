import { useState, useEffect, useMemo } from 'react';
import { Copy, Terminal, Check } from 'lucide-react';
import { Modal } from '../../../components/ui/Modal';
import { Input } from '../../../components/ui/Input';
import { Select } from '../../../components/ui/Select';
import { Textarea } from '../../../components/ui/Textarea';
import { Button } from '../../../components/ui/Button';
import { CardRadioGroup } from '../../../components/ui/CardRadioGroup';
import { toast } from '../../../components/ui/Toast';
import { cn } from '../../../utils/cn';
import { clientApi } from '../../../api/client';
import type { FrpServer } from '../../../api/frpServer';
import type { GithubMirror } from '../../../api/githubMirror';

interface ScriptGeneratorModalProps {
  visible: boolean;
  frpServers: FrpServer[];
  githubMirrors: GithubMirror[];
  onCancel: () => void;
  onLoadData: () => Promise<void>;
}

interface FormData {
  client_name: string;
  frp_server_id: number | undefined;
  server_addr: string;
  server_port: string;
  token_str: string;
  protocol: string;
  script_type: string;
  install_path: string;
  mirror_id: number | undefined;
  remark: string;
}

const initialFormData: FormData = {
  client_name: '',
  frp_server_id: undefined,
  server_addr: '',
  server_port: '',
  token_str: '',
  protocol: 'tcp',
  script_type: 'bash',
  install_path: '/opt/frpc',
  mirror_id: undefined,
  remark: '',
};

export function ScriptGeneratorModal({
  visible,
  frpServers,
  githubMirrors,
  onCancel,
  onLoadData,
}: ScriptGeneratorModalProps) {
  const [formData, setFormData] = useState<FormData>(initialFormData);
  const [generatedScript, setGeneratedScript] = useState('');
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);
  const [errors, setErrors] = useState<Partial<Record<keyof FormData, string>>>({});

  const scriptTypeOptions = useMemo(() => [
    { value: 'bash', label: 'Linux (Bash)', icon: <span>🐧</span> },
    { value: 'powershell', label: 'Windows (PowerShell)', icon: <span>🪟</span> },
  ], []);

  useEffect(() => {
    if (visible) {
      setFormData(initialFormData);
      setGeneratedScript('');
      setErrors({});
      onLoadData();
    }
  }, [visible, onLoadData]);

  const handleChange = (field: keyof FormData, value: string | number | undefined) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  const handleServerChange = (serverId: number | string) => {
    const server = frpServers.find(s => s.id === serverId);
    if (server) {
      const serverAddr = server.server_type === 'remote' ? server.ssh_host : server.host;
      setFormData(prev => ({
        ...prev,
        frp_server_id: server.id,
        server_addr: serverAddr || '',
        server_port: server.bind_port?.toString() || '',
        token_str: server.token || '',
      }));
    }
  };

  const handleScriptTypeChange = (type: string) => {
    const defaultPath = type === 'bash' ? '/opt/frpc' : 'C:\\frpc';
    setFormData(prev => ({
      ...prev,
      script_type: type,
      install_path: defaultPath,
    }));
  };

  const validate = (): boolean => {
    const newErrors: Partial<Record<keyof FormData, string>> = {};
    if (!formData.client_name.trim()) newErrors.client_name = '请输入客户端名称';
    if (!formData.frp_server_id) newErrors.frp_server_id = '请选择FRP服务器';
    if (!formData.server_addr.trim()) newErrors.server_addr = '请输入服务器地址';
    if (!formData.server_port.trim()) newErrors.server_port = '请输入端口';
    if (!formData.install_path.trim()) newErrors.install_path = '请输入安装路径';
    if (!formData.mirror_id) newErrors.mirror_id = '请选择下载加速源';
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleScriptGenerate = async () => {
    if (!validate()) return;

    setLoading(true);
    try {
      const tokenRes = await clientApi.generateRegisterToken({
        client_name: formData.client_name,
        frp_server_id: formData.frp_server_id!,
        server_addr: formData.server_addr,
        server_port: parseInt(formData.server_port),
        token_str: formData.token_str || undefined,
        protocol: formData.protocol,
        remark: formData.remark || undefined,
      }) as { token: string };

      const script = await clientApi.generateRegisterScript({
        token: tokenRes.token,
        type: formData.script_type,
        mirror: formData.mirror_id!.toString(),
      }) as string;

      setGeneratedScript(script);
    } catch (error: unknown) {
      const err = error as Error;
      toast.error(`生成脚本失败: ${err?.message || '未知错误'}`);
    } finally {
      setLoading(false);
    }
  };

  const handleCopyScript = async () => {
    await navigator.clipboard.writeText(generatedScript);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleClose = () => {
    setGeneratedScript('');
    setCopied(false);
    onCancel();
  };

  const serverOptions = frpServers
    .filter(server => server.id !== undefined)
    .map(server => {
      const displayAddr = server.server_type === 'remote' ? server.ssh_host : server.host;
      return {
        value: server.id as number,
        label: `${server.name} (${displayAddr}:${server.bind_port})`,
      };
    });

  const mirrorOptions = githubMirrors.map(mirror => ({
    value: mirror.id,
    label: `${mirror.name}${mirror.is_default ? ' (默认)' : ''}`,
  }));

  const protocolOptions = [
    { value: 'tcp', label: 'TCP' },
    { value: 'kcp', label: 'KCP' },
    { value: 'websocket', label: 'WebSocket' },
  ];

  return (
    <Modal
      open={visible}
      onClose={handleClose}
      title="生成客户端注册脚本"
      size="xl"
      footer={
        generatedScript ? (
          <>
            <Button
              variant="primary"
              icon={copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
              onClick={handleCopyScript}
            >
              {copied ? '已复制' : '复制脚本'}
            </Button>
            <Button variant="secondary" onClick={handleClose}>
              关闭
            </Button>
          </>
        ) : undefined
      }
    >
      {!generatedScript ? (
        <div className="space-y-4">
          <Input
            label="客户端名称"
            placeholder="输入客户端名称"
            value={formData.client_name}
            onChange={(e) => handleChange('client_name', e.target.value)}
            error={errors.client_name}
            required
          />

          <Select
            label="FRP服务器"
            placeholder="选择FRP服务器"
            options={serverOptions}
            value={formData.frp_server_id}
            onChange={handleServerChange}
            error={errors.frp_server_id}
          />

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input
              label="服务器地址"
              value={formData.server_addr}
              onChange={(e) => handleChange('server_addr', e.target.value)}
              error={errors.server_addr}
              disabled
              required
            />
            <Input
              label="端口"
              type="number"
              value={formData.server_port}
              onChange={(e) => handleChange('server_port', e.target.value)}
              error={errors.server_port}
              disabled
              required
            />
          </div>

          <Input
            label="令牌"
            value={formData.token_str}
            onChange={(e) => handleChange('token_str', e.target.value)}
            disabled
            placeholder="自动从选择的FRP服务器获取"
          />

          <Select
            label="协议"
            options={protocolOptions}
            value={formData.protocol}
            onChange={(v) => handleChange('protocol', v as string)}
          />

          {/* 脚本类型选择 */}
          <div>
            <label className="block text-sm font-medium mb-2 text-foreground-secondary">
              脚本类型 <span className="text-red-400">*</span>
            </label>
            <CardRadioGroup
              name="script_type"
              value={formData.script_type}
              onChange={handleScriptTypeChange}
              options={scriptTypeOptions}
              equalWidth={false}
            />
          </div>

          <Input
            label="安装路径"
            placeholder="客户端安装路径"
            value={formData.install_path}
            onChange={(e) => handleChange('install_path', e.target.value)}
            error={errors.install_path}
            required
          />

          <Select
            label="下载加速源"
            placeholder="选择下载加速源"
            options={mirrorOptions}
            value={formData.mirror_id}
            onChange={(v) => handleChange('mirror_id', v as number)}
            error={errors.mirror_id}
          />

          <Textarea
            label="备注"
            placeholder="可选备注信息"
            value={formData.remark}
            onChange={(e) => handleChange('remark', e.target.value)}
            rows={2}
          />

          <Button
            className="w-full"
            onClick={handleScriptGenerate}
            loading={loading}
            icon={<Terminal className="h-4 w-4" />}
          >
            生成脚本
          </Button>
        </div>
      ) : (
        <div className="space-y-4">
          <p className="text-sm text-foreground-secondary">请在客户端机器上执行以下命令:</p>
          
          {/* 脚本显示区域 - 科技风样式 */}
          <div className="relative">
            <div className="absolute -inset-0.5 bg-gradient-to-r from-indigo-500/20 via-purple-500/20 to-indigo-500/20 rounded-lg blur opacity-50" />
            <div className={cn("relative rounded-lg p-4 font-mono text-sm border bg-card-bg border-card-border")}>
              <pre className="text-green-600 dark:text-green-400 whitespace-pre-wrap break-all overflow-x-auto">
                {generatedScript}
              </pre>
            </div>
          </div>

          <p className="text-slate-500 text-xs">
            提示: 该命令会自动下载并执行安装脚本,完成frpc客户端的安装和注册
          </p>
        </div>
      )}
    </Modal>
  );
}