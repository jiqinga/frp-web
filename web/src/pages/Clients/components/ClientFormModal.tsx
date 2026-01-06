import { useState, useEffect } from 'react';
import { ChevronDown, ChevronRight, FileText } from 'lucide-react';
import { Modal } from '../../../components/ui/Modal';
import { Input } from '../../../components/ui/Input';
import { Textarea } from '../../../components/ui/Textarea';
import { Button } from '../../../components/ui/Button';
import type { Client } from '../../../types';

interface ClientFormModalProps {
  visible: boolean;
  editingClient: Client | null;
  onCancel: () => void;
  onSubmit: (values: Partial<Client>) => Promise<void>;
  onParseConfig: (configContent: string) => Promise<{
    server_addr?: string;
    server_port?: number;
    token?: string;
    frpc_admin_host?: string;
    frpc_admin_port?: number;
    frpc_admin_user?: string;
    frpc_admin_pwd?: string;
  } | null>;
}

interface FormData {
  name: string;
  server_addr: string;
  server_port: string;
  token: string;
  frpc_admin_host: string;
  frpc_admin_port: string;
  frpc_admin_user: string;
  frpc_admin_pwd: string;
  remark: string;
}

const initialFormData: FormData = {
  name: '',
  server_addr: '',
  server_port: '',
  token: '',
  frpc_admin_host: '',
  frpc_admin_port: '',
  frpc_admin_user: '',
  frpc_admin_pwd: '',
  remark: '',
};

export function ClientFormModal({
  visible,
  editingClient,
  onCancel,
  onSubmit,
  onParseConfig,
}: ClientFormModalProps) {
  const [formData, setFormData] = useState<FormData>(initialFormData);
  const [configContent, setConfigContent] = useState('');
  const [importLoading, setImportLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [configExpanded, setConfigExpanded] = useState(false);
  const [errors, setErrors] = useState<Partial<Record<keyof FormData, string>>>({});

  // 当编辑客户端变化时，重置表单
  useEffect(() => {
    if (visible) {
      if (editingClient) {
        setFormData({
          name: editingClient.name || '',
          server_addr: editingClient.server_addr || '',
          server_port: editingClient.server_port?.toString() || '',
          token: editingClient.token || '',
          frpc_admin_host: editingClient.frpc_admin_host || '',
          frpc_admin_port: editingClient.frpc_admin_port?.toString() || '',
          frpc_admin_user: editingClient.frpc_admin_user || '',
          frpc_admin_pwd: editingClient.frpc_admin_pwd || '',
          remark: editingClient.remark || '',
        });
      } else {
        setFormData(initialFormData);
      }
      setConfigContent('');
      setErrors({});
      setConfigExpanded(false);
    }
  }, [visible, editingClient]);

  const handleChange = (field: keyof FormData, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  const handleImportConfig = async () => {
    if (!configContent.trim()) return;
    
    setImportLoading(true);
    try {
      const res = await onParseConfig(configContent);
      if (res) {
        setFormData(prev => ({
          ...prev,
          server_addr: res.server_addr || prev.server_addr,
          server_port: res.server_port?.toString() || prev.server_port,
          token: res.token || prev.token,
          frpc_admin_host: res.frpc_admin_host || prev.frpc_admin_host,
          frpc_admin_port: res.frpc_admin_port?.toString() || prev.frpc_admin_port,
          frpc_admin_user: res.frpc_admin_user || prev.frpc_admin_user,
          frpc_admin_pwd: res.frpc_admin_pwd || prev.frpc_admin_pwd,
        }));
        setConfigContent('');
        setConfigExpanded(false);
      }
    } finally {
      setImportLoading(false);
    }
  };

  const validate = (): boolean => {
    const newErrors: Partial<Record<keyof FormData, string>> = {};
    if (!formData.name.trim()) newErrors.name = '请输入客户端名称';
    if (!formData.server_addr.trim()) newErrors.server_addr = '请输入服务器地址';
    if (!formData.server_port.trim()) newErrors.server_port = '请输入端口';
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async () => {
    if (!validate()) return;
    
    setSubmitting(true);
    try {
      await onSubmit({
        name: formData.name,
        server_addr: formData.server_addr,
        server_port: formData.server_port ? parseInt(formData.server_port) : undefined,
        token: formData.token || undefined,
        frpc_admin_host: formData.frpc_admin_host || undefined,
        frpc_admin_port: formData.frpc_admin_port ? parseInt(formData.frpc_admin_port) : undefined,
        frpc_admin_user: formData.frpc_admin_user || undefined,
        frpc_admin_pwd: formData.frpc_admin_pwd || undefined,
        remark: formData.remark || undefined,
      });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      open={visible}
      onClose={onCancel}
      title={editingClient ? '编辑客户端' : '新增客户端'}
      size="lg"
      footer={
        <>
          <Button variant="secondary" onClick={onCancel}>
            取消
          </Button>
          <Button onClick={handleSubmit} loading={submitting}>
            确定
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        <Input
          label="名称"
          placeholder="客户端名称"
          value={formData.name}
          onChange={(e) => handleChange('name', e.target.value)}
          error={errors.name}
          required
        />

        {/* 配置导入折叠面板 */}
        <div className="border border-border rounded-lg overflow-hidden">
          <Button
            variant="ghost"
            onClick={() => setConfigExpanded(!configExpanded)}
            className="w-full px-4 py-3 rounded-none text-left bg-surface-hover hover:bg-surface-active"
          >
            <span className="flex items-center gap-2 whitespace-nowrap">
              {configExpanded ? (
                <ChevronDown className="h-4 w-4 text-foreground-muted flex-shrink-0" />
              ) : (
                <ChevronRight className="h-4 w-4 text-foreground-muted flex-shrink-0" />
              )}
              <FileText className="h-4 w-4 text-indigo-400 flex-shrink-0" />
              <span className="text-sm text-foreground-secondary">从配置文件导入</span>
            </span>
          </Button>
          
          {configExpanded && (
            <div className="p-4 border-t border-border space-y-3">
              <Textarea
                rows={6}
                placeholder={'粘贴frpc配置文件内容(INI格式)\n例如:\n[common]\nserver_addr = 192.168.1.100\nserver_port = 7000\ntoken = your_token'}
                value={configContent}
                onChange={(e) => setConfigContent(e.target.value)}
              />
              <div className="flex gap-2">
                <Button
                  size="sm"
                  onClick={handleImportConfig}
                  loading={importLoading}
                >
                  解析并填充表单
                </Button>
                <Button
                  size="sm"
                  variant="secondary"
                  onClick={() => setConfigContent('')}
                >
                  清空
                </Button>
              </div>
            </div>
          )}
        </div>

        <div className="h-px bg-border" />

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Input
            label="服务器地址"
            placeholder="FRP服务器地址"
            value={formData.server_addr}
            onChange={(e) => handleChange('server_addr', e.target.value)}
            error={errors.server_addr}
            required
          />
          <Input
            label="端口"
            type="number"
            placeholder="FRP服务器端口"
            value={formData.server_port}
            onChange={(e) => handleChange('server_port', e.target.value)}
            error={errors.server_port}
            required
          />
        </div>

        <Input
          label="令牌"
          placeholder="FRP连接Token（可选）"
          value={formData.token}
          onChange={(e) => handleChange('token', e.target.value)}
        />

        {/* Admin API 配置区域 */}
        <div className="pt-2">
          <div className="text-sm font-medium text-indigo-400 mb-3">
            Admin API 配置（用于在线状态检测）
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input
              label="管理接口地址"
              placeholder="例如: 127.0.0.1 或客户端IP地址"
              value={formData.frpc_admin_host}
              onChange={(e) => handleChange('frpc_admin_host', e.target.value)}
              hint="frpc客户端的Admin API地址"
            />
            <Input
              label="管理接口端口"
              type="number"
              placeholder="例如: 7400"
              value={formData.frpc_admin_port}
              onChange={(e) => handleChange('frpc_admin_port', e.target.value)}
              hint="默认为7400"
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
            <Input
              label="管理接口用户名"
              placeholder="认证用户名（如果启用了认证）"
              value={formData.frpc_admin_user}
              onChange={(e) => handleChange('frpc_admin_user', e.target.value)}
            />
            <Input
              label="管理接口密码"
              type="password"
              placeholder="认证密码（如果启用了认证）"
              value={formData.frpc_admin_pwd}
              onChange={(e) => handleChange('frpc_admin_pwd', e.target.value)}
            />
          </div>
        </div>

        <Textarea
          label="备注"
          rows={3}
          placeholder="客户端备注信息"
          value={formData.remark}
          onChange={(e) => handleChange('remark', e.target.value)}
        />
      </div>
    </Modal>
  );
}