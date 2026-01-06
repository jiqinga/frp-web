import { useState, useEffect } from 'react';
import { RefreshCw, ChevronDown, ChevronRight, Server, Cloud, Info } from 'lucide-react';
import { Modal } from '../../../components/ui/Modal';
import { Button } from '../../../components/ui/Button';
import { Input } from '../../../components/ui/Input';
import { Select, type SelectOption } from '../../../components/ui/Select';
import { Textarea } from '../../../components/ui/Textarea';
import { Tooltip } from '../../../components/ui/Tooltip';
import { Switch } from '../../../components/ui/Switch';
import { RadioGroup, type RadioOption } from '../../../components/ui/RadioGroup';
import type { FrpServer } from '../../../api/frpServer';
import type { GithubMirror } from '../../../api/githubMirror';

interface ServerFormModalProps {
  visible: boolean;
  editingServer: FrpServer | null;
  serverType: 'local' | 'remote';
  mirrors: GithubMirror[];
  configContent: string;
  importLoading: boolean;
  onCancel: () => void;
  onSubmit: (values: Partial<FrpServer>) => void;
  onServerTypeChange: (type: 'local' | 'remote') => void;
  onConfigContentChange: (content: string) => void;
  onImportConfig: () => void;
  onGenerateToken: () => string;
}

/**
 * 服务器表单弹窗组件
 * 支持新增/编辑本地和远程服务器
 * 使用 Tailwind CSS + Lucide Icons 重构
 */
export function ServerFormModal({
  visible,
  editingServer,
  serverType,
  mirrors,
  configContent,
  importLoading,
  onCancel,
  onSubmit,
  onServerTypeChange,
  onConfigContentChange,
  onImportConfig,
  onGenerateToken,
}: ServerFormModalProps) {
  // 表单数据状态
  const [formData, setFormData] = useState<Partial<FrpServer>>({
    name: '',
    server_type: 'local',
    host: '',
    dashboard_port: 7500,
    dashboard_user: '',
    dashboard_pwd: '',
    bind_port: 7000,
    token: '',
    ssh_host: '',
    ssh_port: 22,
    ssh_user: 'root',
    ssh_password: '',
    install_path: '/opt/frps',
    mirror_id: undefined,
    enabled: true,
  });

  // 折叠面板状态
  const [configImportExpanded, setConfigImportExpanded] = useState(false);
  const [advancedConfigExpanded, setAdvancedConfigExpanded] = useState(false);

  // 表单验证错误
  const [errors, setErrors] = useState<Record<string, string>>({});

  // 初始化表单数据
  useEffect(() => {
    if (visible) {
      if (editingServer) {
        setFormData({
          ...editingServer,
        });
        onServerTypeChange(editingServer.server_type as 'local' | 'remote');
      } else {
        setFormData({
          name: '',
          server_type: serverType,
          host: serverType === 'local' ? '127.0.0.1' : '0.0.0.0',
          dashboard_port: 7500,
          dashboard_user: '',
          dashboard_pwd: '',
          bind_port: 7000,
          token: '',
          ssh_host: '',
          ssh_port: 22,
          ssh_user: 'root',
          ssh_password: '',
          install_path: '/opt/frps',
          mirror_id: undefined,
          enabled: true,
        });
      }
      setErrors({});
      setConfigImportExpanded(false);
      setAdvancedConfigExpanded(false);
    }
  }, [visible, editingServer, serverType, onServerTypeChange]);

  // 更新表单字段
  const updateField = (field: string, value: unknown) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    // 清除该字段的错误
    if (errors[field]) {
      setErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[field];
        return newErrors;
      });
    }
  };

  // 处理服务器类型切换
  const handleServerTypeChange = (type: 'local' | 'remote') => {
    onServerTypeChange(type);
    updateField('server_type', type);
    if (type === 'local') {
      updateField('host', '127.0.0.1');
    } else {
      updateField('host', '0.0.0.0');
    }
  };

  // 生成随机 Token
  const handleGenerateToken = () => {
    const token = onGenerateToken();
    updateField('token', token);
  };

  // 表单验证
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name?.trim()) {
      newErrors.name = '请输入服务器名称';
    }

    if (serverType === 'remote') {
      if (!formData.ssh_host?.trim()) {
        newErrors.ssh_host = '请输入SSH主机地址';
      }
      if (!formData.ssh_user?.trim()) {
        newErrors.ssh_user = '请输入SSH用户名';
      }
      if (!formData.ssh_password?.trim()) {
        newErrors.ssh_password = '请输入SSH密码';
      }
    } else {
      if (!formData.host?.trim()) {
        newErrors.host = '请输入主机地址';
      }
      if (!formData.dashboard_port) {
        newErrors.dashboard_port = '请输入Dashboard端口';
      }
      if (!formData.bind_port) {
        newErrors.bind_port = '请输入绑定端口';
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // 提交表单
  const handleSubmit = () => {
    if (validateForm()) {
      onSubmit(formData);
    }
  };

  // 镜像选项
  const mirrorOptions: SelectOption[] = mirrors.map(m => ({
    value: m.id!,
    label: `${m.name}${m.is_default ? ' (默认)' : ''}`,
  }));

  return (
    <Modal
      open={visible}
      onClose={onCancel}
      title={editingServer ? '编辑服务器' : '添加服务器'}
      size="lg"
      footer={
        <div className="flex justify-end gap-3">
          <Button variant="ghost" onClick={onCancel}>
            取消
          </Button>
          <Button onClick={handleSubmit}>
            {editingServer ? '保存' : '添加'}
          </Button>
        </div>
      }
    >
      <div className="space-y-6 max-h-[70vh] overflow-y-auto pr-2">
        {/* 基本信息 */}
        <div className="space-y-4">
          {/* 名称 */}
          <div>
            <label className="block text-sm font-medium mb-1.5 text-foreground-secondary">
              名称 <span className="text-red-400">*</span>
            </label>
            <Input
              value={formData.name || ''}
              onChange={(e) => updateField('name', e.target.value)}
              placeholder="输入服务器名称"
              error={errors.name}
            />
          </div>

          {/* 服务器类型 */}
          <div>
            <label className="block text-sm font-medium mb-1.5 text-foreground-secondary">
              服务器类型 <span className="text-red-400">*</span>
            </label>
            <RadioGroup
              name="server_type"
              value={serverType}
              onChange={handleServerTypeChange}
              options={[
                { value: 'local', label: '本地服务器', icon: <Server className="h-4 w-4 text-green-400" /> },
                { value: 'remote', label: '远程服务器', icon: <Cloud className="h-4 w-4 text-blue-400" /> },
              ] as RadioOption<'local' | 'remote'>[]}
            />
          </div>
        </div>

        {/* 配置导入折叠面板 */}
        <div className="border rounded-lg overflow-hidden border-border">
          <Button
            variant="ghost"
            className="w-full px-4 py-3 flex items-center justify-between rounded-none bg-surface-elevated hover:bg-surface-hover"
            onClick={() => setConfigImportExpanded(!configImportExpanded)}
          >
            <span className="text-sm font-medium text-foreground-secondary">从配置文件导入</span>
            {configImportExpanded ? (
              <ChevronDown className="h-4 w-4 text-foreground-muted" />
            ) : (
              <ChevronRight className="h-4 w-4 text-foreground-muted" />
            )}
          </Button>
          {configImportExpanded && (
            <div className="p-4 border-t space-y-3 border-border">
              <Textarea
                rows={8}
                placeholder={`粘贴frps配置文件内容(YAML格式)
例如:
bindPort: 7000
auth:
  method: token
  token: your_token
webServer:
  addr: 0.0.0.0
  port: 7500
  user: admin
  password: your_password`}
                value={configContent}
                onChange={(e) => onConfigContentChange(e.target.value)}
              />
              <div className="flex gap-2">
                <Button
                  size="sm"
                  onClick={onImportConfig}
                  loading={importLoading}
                >
                  解析并填充表单
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onConfigContentChange('')}
                >
                  清空
                </Button>
              </div>
            </div>
          )}
        </div>

        {/* 分隔线 */}
        <div className="border-t border-border" />

        {/* 远程服务器 SSH 配置 */}
        {serverType === 'remote' && (
          <div className="space-y-4">
            <h3 className="text-sm font-medium flex items-center gap-2 text-foreground-secondary">
              <Cloud className="h-4 w-4 text-blue-400" />
              SSH 连接配置（必填）
            </h3>
            
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1.5 text-foreground-secondary">
                  SSH主机 <span className="text-red-400">*</span>
                </label>
                <Input
                  value={formData.ssh_host || ''}
                  onChange={(e) => updateField('ssh_host', e.target.value)}
                  placeholder="例如: 192.168.1.100"
                  error={errors.ssh_host}
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1.5 text-foreground-secondary">
                  SSH端口 <span className="text-red-400">*</span>
                </label>
                <Input
                  type="number"
                  value={formData.ssh_port || 22}
                  onChange={(e) => updateField('ssh_port', parseInt(e.target.value) || 22)}
                  min={1}
                  max={65535}
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1.5 text-foreground-secondary">
                  SSH用户名 <span className="text-red-400">*</span>
                </label>
                <Input
                  value={formData.ssh_user || ''}
                  onChange={(e) => updateField('ssh_user', e.target.value)}
                  placeholder="例如: root"
                  error={errors.ssh_user}
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1.5 text-foreground-secondary">
                  SSH密码 <span className="text-red-400">*</span>
                </label>
                <Input
                  type="password"
                  value={formData.ssh_password || ''}
                  onChange={(e) => updateField('ssh_password', e.target.value)}
                  error={errors.ssh_password}
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium mb-1.5 text-foreground-secondary">
                安装路径
              </label>
              <Input
                value={formData.install_path || ''}
                onChange={(e) => updateField('install_path', e.target.value)}
                placeholder="默认: /opt/frps"
              />
            </div>

            {/* 远程服务器的 FRP 配置（可选） */}
            <div className="border rounded-lg overflow-hidden mt-4 border-border">
              <Button
                variant="ghost"
                className="w-full px-4 py-3 flex items-center justify-between rounded-none bg-surface-elevated hover:bg-surface-hover"
                onClick={() => setAdvancedConfigExpanded(!advancedConfigExpanded)}
              >
                <span className="text-sm font-medium text-foreground-secondary">FRP服务配置（可选，不填使用默认值）</span>
                {advancedConfigExpanded ? (
                  <ChevronDown className="h-4 w-4 text-foreground-muted" />
                ) : (
                  <ChevronRight className="h-4 w-4 text-foreground-muted" />
                )}
              </Button>
              {advancedConfigExpanded && (
                <div className="p-4 border-t space-y-4 border-border">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium mb-1.5 flex items-center gap-1 text-foreground-secondary">
                        主机地址
                        <Tooltip content="FRP服务监听地址，默认0.0.0.0">
                          <Info className="h-3.5 w-3.5 text-foreground-muted" />
                        </Tooltip>
                      </label>
                      <Input
                        value={formData.host || ''}
                        onChange={(e) => updateField('host', e.target.value)}
                        placeholder="默认: 0.0.0.0"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium mb-1.5 flex items-center gap-1 text-foreground-secondary">
                        Dashboard端口
                        <Tooltip content="默认7500">
                          <Info className="h-3.5 w-3.5 text-foreground-muted" />
                        </Tooltip>
                      </label>
                      <Input
                        type="number"
                        value={formData.dashboard_port || ''}
                        onChange={(e) => updateField('dashboard_port', parseInt(e.target.value) || undefined)}
                        placeholder="7500"
                        min={1}
                        max={65535}
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium mb-1.5 text-foreground-secondary">
                        Dashboard用户名
                      </label>
                      <Input
                        value={formData.dashboard_user || ''}
                        onChange={(e) => updateField('dashboard_user', e.target.value)}
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium mb-1.5 text-foreground-secondary">
                        Dashboard密码
                      </label>
                      <Input
                        type="password"
                        value={formData.dashboard_pwd || ''}
                        onChange={(e) => updateField('dashboard_pwd', e.target.value)}
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium mb-1.5 flex items-center gap-1 text-foreground-secondary">
                        绑定端口
                        <Tooltip content="默认7000">
                          <Info className="h-3.5 w-3.5 text-foreground-muted" />
                        </Tooltip>
                      </label>
                      <Input
                        type="number"
                        value={formData.bind_port || ''}
                        onChange={(e) => updateField('bind_port', parseInt(e.target.value) || undefined)}
                        placeholder="7000"
                        min={1}
                        max={65535}
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium mb-1.5 flex items-center gap-1 text-foreground-secondary">
                        认证Token
                        <Tooltip content="用于客户端连接认证，留空自动生成">
                          <Info className="h-3.5 w-3.5 text-foreground-muted" />
                        </Tooltip>
                      </label>
                      <div className="flex gap-2">
                        <Input
                          type="password"
                          value={formData.token || ''}
                          onChange={(e) => updateField('token', e.target.value)}
                          placeholder="留空自动生成48位随机Token"
                          className="flex-1"
                        />
                        <Tooltip content="生成随机Token">
                          <Button
                            variant="ghost"
                            size="sm"
                            icon={<RefreshCw className="h-4 w-4" />}
                            onClick={handleGenerateToken}
                          />
                        </Tooltip>
                      </div>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium mb-1.5 flex items-center gap-1 text-foreground-secondary">
                        GitHub加速源
                        <Tooltip content="选择用于下载FRP的加速源">
                          <Info className="h-3.5 w-3.5 text-foreground-muted" />
                        </Tooltip>
                      </label>
                      <Select
                        value={formData.mirror_id}
                        onChange={(value) => updateField('mirror_id', value)}
                        options={mirrorOptions}
                        placeholder="使用默认加速源"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium mb-1.5 text-foreground-secondary">
                        启用
                      </label>
                      <Switch
                        checked={formData.enabled ?? true}
                        onChange={(checked) => updateField('enabled', checked)}
                      />
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        {/* 本地服务器配置 */}
        {serverType === 'local' && (
          <div className="space-y-4">
            <h3 className="text-sm font-medium flex items-center gap-2 text-foreground-secondary">
              <Server className="h-4 w-4 text-green-400" />
              本地服务器配置
            </h3>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1.5 text-foreground-secondary">
                  主机地址 <span className="text-red-400">*</span>
                </label>
                <Input
                  value={formData.host || ''}
                  onChange={(e) => updateField('host', e.target.value)}
                  placeholder="例如: 127.0.0.1"
                  error={errors.host}
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1.5 text-foreground-secondary">
                  Dashboard端口 <span className="text-red-400">*</span>
                </label>
                <Input
                  type="number"
                  value={formData.dashboard_port || ''}
                  onChange={(e) => updateField('dashboard_port', parseInt(e.target.value) || undefined)}
                  min={1}
                  max={65535}
                  error={errors.dashboard_port}
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1.5 text-foreground-secondary">
                  Dashboard用户名
                </label>
                <Input
                  value={formData.dashboard_user || ''}
                  onChange={(e) => updateField('dashboard_user', e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1.5 text-foreground-secondary">
                  Dashboard密码
                </label>
                <Input
                  type="password"
                  value={formData.dashboard_pwd || ''}
                  onChange={(e) => updateField('dashboard_pwd', e.target.value)}
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1.5 text-foreground-secondary">
                  绑定端口 <span className="text-red-400">*</span>
                </label>
                <Input
                  type="number"
                  value={formData.bind_port || ''}
                  onChange={(e) => updateField('bind_port', parseInt(e.target.value) || undefined)}
                  min={1}
                  max={65535}
                  error={errors.bind_port}
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1.5 flex items-center gap-1 text-foreground-secondary">
                  认证Token
                  <Tooltip content="用于客户端连接认证，留空自动生成">
                    <Info className="h-3.5 w-3.5 text-foreground-muted" />
                  </Tooltip>
                </label>
                <div className="flex gap-2">
                  <Input
                    type="password"
                    value={formData.token || ''}
                    onChange={(e) => updateField('token', e.target.value)}
                    placeholder="留空自动生成48位随机Token"
                    className="flex-1"
                  />
                  <Tooltip content="生成随机Token">
                    <Button
                      variant="ghost"
                      size="sm"
                      icon={<RefreshCw className="h-4 w-4" />}
                      onClick={handleGenerateToken}
                    />
                  </Tooltip>
                </div>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1.5 flex items-center gap-1 text-foreground-secondary">
                  GitHub加速源
                  <Tooltip content="选择用于下载FRP的加速源">
                    <Info className="h-3.5 w-3.5 text-foreground-muted" />
                  </Tooltip>
                </label>
                <Select
                  value={formData.mirror_id}
                  onChange={(value) => updateField('mirror_id', value)}
                  options={mirrorOptions}
                  placeholder="使用默认加速源"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1.5 text-foreground-secondary">
                  启用
                </label>
                <Switch
                  checked={formData.enabled ?? true}
                  onChange={(checked) => updateField('enabled', checked)}
                />
              </div>
            </div>
          </div>
        )}
      </div>
    </Modal>
  );
}