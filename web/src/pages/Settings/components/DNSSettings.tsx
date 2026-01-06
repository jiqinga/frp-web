
import { useState, useEffect, useCallback, useMemo } from 'react';
import { Cloud, Plus, Edit, Trash2, RefreshCw, CheckCircle, Eye, EyeOff, Globe, Mail, Save } from 'lucide-react';
import { Button, Input, Modal, Table, Card, CardHeader, CardContent, Badge, Switch, Select, Tooltip } from '../../../components/ui';
import { toast } from '../../../components/ui';
import { dnsApi } from '../../../api/dns';
import { settingApi } from '../../../api/setting';
import type { DNSProvider, DNSProviderType } from '../../../types';
import { DNS_PROVIDER_TYPE_LABELS, DNS_PROVIDER_AUTH_FIELDS } from '../../../types';

interface DNSProviderForm {
  name: string;
  type: DNSProviderType;
  access_key: string;
  secret_key: string;
  enabled: boolean;
}

const initialForm: DNSProviderForm = {
  name: '',
  type: 'aliyun',
  access_key: '',
  secret_key: '',
  enabled: true,
};

const providerTypeOptions = [
  { value: 'aliyun', label: '阿里云 DNS' },
  { value: 'cloudflare', label: 'Cloudflare' },
  { value: 'tencent', label: '腾讯云 DNS' },
];

export function DNSSettings() {
  const [providers, setProviders] = useState<DNSProvider[]>([]);
  const [loading, setLoading] = useState(false);
  const [acmeEmail, setAcmeEmail] = useState('');
  const [acmeEmailLoading, setAcmeEmailLoading] = useState(false);
  const [acmeEmailSaving, setAcmeEmailSaving] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingProvider, setEditingProvider] = useState<DNSProvider | null>(null);
  const [form, setForm] = useState<DNSProviderForm>(initialForm);
  const [testingId, setTestingId] = useState<number | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [testingConfig, setTestingConfig] = useState(false);
  const [showAccessKey, setShowAccessKey] = useState(false);
  const [showSecretKey, setShowSecretKey] = useState(false);
  const [loadingSecret, setLoadingSecret] = useState(false);
  const [secretLoaded, setSecretLoaded] = useState(false);
  const [providerDomains, setProviderDomains] = useState<Record<number, string[]>>({});
  const [domainsModalVisible, setDomainsModalVisible] = useState(false);
  const [viewingDomains, setViewingDomains] = useState<{ name: string; domains: string[] }>({ name: '', domains: [] });

  const fetchAcmeEmail = useCallback(async () => {
    setAcmeEmailLoading(true);
    try {
      const settings = await settingApi.getSettings();
      const acmeSetting = settings.find(s => s.key === 'acme_email');
      setAcmeEmail(acmeSetting?.value || '');
    } catch {
      // ignore
    } finally {
      setAcmeEmailLoading(false);
    }
  }, []);

  const handleSaveAcmeEmail = async () => {
    setAcmeEmailSaving(true);
    try {
      await settingApi.updateSetting({ key: 'acme_email', value: acmeEmail });
      toast.success('ACME邮箱保存成功');
    } catch {
      toast.error('保存失败');
    } finally {
      setAcmeEmailSaving(false);
    }
  };

  const fetchProviders = useCallback(async () => {
    setLoading(true);
    try {
      const data = await dnsApi.getProviders();
      setProviders(data);
      // 批量获取每个提供商的托管域名
      const domainsMap: Record<number, string[]> = {};
      await Promise.all(
        data.map(async (provider) => {
          try {
            const domains = await dnsApi.getProviderDomains(provider.id);
            domainsMap[provider.id] = domains;
          } catch {
            domainsMap[provider.id] = [];
          }
        })
      );
      setProviderDomains(domainsMap);
    } catch {
      toast.error('获取 DNS 提供商列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchProviders();
    fetchAcmeEmail();
  }, [fetchProviders, fetchAcmeEmail]);

  const handleAdd = () => {
    setEditingProvider(null);
    setForm(initialForm);
    setShowAccessKey(false);
    setShowSecretKey(false);
    setModalVisible(true);
  };

  const handleEdit = (provider: DNSProvider) => {
    setEditingProvider(provider);
    setForm({
      name: provider.name,
      type: provider.type,
      access_key: provider.access_key,
      secret_key: '',
      enabled: provider.enabled,
    });
    setShowAccessKey(false);
    setShowSecretKey(false);
    setSecretLoaded(false);
    setModalVisible(true);
  };

  // 点击眼睛图标时获取并显示密钥
  const handleToggleSecretKey = async () => {
    if (showSecretKey) {
      setShowSecretKey(false);
      return;
    }
    
    // 编辑模式且密钥未加载过，需要从后端获取
    if (editingProvider && !secretLoaded && !form.secret_key) {
      setLoadingSecret(true);
      try {
        const data = await dnsApi.getProviderSecret(editingProvider.id);
        setForm(prev => ({ ...prev, secret_key: data.secret_key }));
        setSecretLoaded(true);
        setShowSecretKey(true);
      } catch {
        toast.error('获取密钥失败');
      } finally {
        setLoadingSecret(false);
      }
    } else {
      setShowSecretKey(true);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('确定要删除此 DNS 提供商吗？')) return;
    try {
      await dnsApi.deleteProvider(id);
      toast.success('删除成功');
      fetchProviders();
    } catch {
      toast.error('删除失败');
    }
  };

  const handleTest = async (id: number) => {
    setTestingId(id);
    try {
      await dnsApi.testProviderById(id);
      toast.success('连接测试成功');
    } catch {
      toast.error('连接测试失败');
    } finally {
      setTestingId(null);
    }
  };

  // 获取当前提供商的认证字段配置
  const authFieldConfig = useMemo(() => {
    return DNS_PROVIDER_AUTH_FIELDS[form.type];
  }, [form.type]);

  const handleTestConfig = async () => {
    if (!form.access_key) {
      toast.error('请填写 Access Key');
      return;
    }
    if (authFieldConfig.secretKeyRequired && !form.secret_key && !editingProvider) {
      toast.error(`请填写 ${authFieldConfig.secretKeyLabel}`);
      return;
    }

    setTestingConfig(true);
    try {
      await dnsApi.testProvider({
        type: form.type,
        access_key: form.access_key,
        secret_key: form.secret_key || undefined,
      });
      toast.success('连接测试成功');
    } catch {
      toast.error('连接测试失败');
    } finally {
      setTestingConfig(false);
    }
  };

  const handleSubmit = async () => {
    if (!form.name || !form.access_key) {
      toast.error('请填写必填字段');
      return;
    }
    
    // 根据提供商类型检查是否需要 Secret Key
    if (!editingProvider && authFieldConfig.secretKeyRequired && !form.secret_key) {
      toast.error(`请填写 ${authFieldConfig.secretKeyLabel}`);
      return;
    }

    setSubmitting(true);
    try {
      if (editingProvider) {
        await dnsApi.updateProvider(editingProvider.id, {
          name: form.name,
          type: form.type,
          access_key: form.access_key,
          secret_key: form.secret_key || undefined,
          enabled: form.enabled,
        });
        toast.success('更新成功');
      } else {
        await dnsApi.createProvider({
          name: form.name,
          type: form.type,
          access_key: form.access_key,
          secret_key: form.secret_key,
          enabled: form.enabled,
        });
        toast.success('创建成功');
      }
      setModalVisible(false);
      fetchProviders();
    } catch {
      toast.error(editingProvider ? '更新失败' : '创建失败');
    } finally {
      setSubmitting(false);
    }
  };

  const columns = [
    {
      key: 'name',
      title: '名称',
      render: (_: unknown, record: DNSProvider) => (
        <div className="flex items-center justify-center gap-2">
          <Cloud className="h-4 w-4 text-blue-400" />
          <span className="font-medium text-foreground">{record.name}</span>
        </div>
      )
    },
    {
      key: 'type',
      title: '类型',
      render: (_: unknown, record: DNSProvider) => {
        const typeLabel = DNS_PROVIDER_TYPE_LABELS[record.type] || record.type;
        return <span className="text-foreground-secondary">{typeLabel}</span>;
      }
    },
    {
      key: 'domains',
      title: '托管域名',
      render: (_: unknown, record: DNSProvider) => {
        const domains = providerDomains[record.id] || [];
        if (domains.length === 0) {
          return <span className="text-foreground-subtle text-sm">暂无域名</span>;
        }
        const visibleDomains = domains.slice(0, 2);
        const remainingCount = domains.length - 2;
        return (
          <div className="flex flex-wrap gap-1 justify-center items-center">
            {visibleDomains.map((domain) => (
              <Badge key={domain} variant="info" className="text-xs">
                {domain}
              </Badge>
            ))}
            {remainingCount > 0 && (
              <Badge
                variant="default"
                className="text-xs cursor-pointer"
                onClick={() => {
                  setViewingDomains({ name: record.name, domains });
                  setDomainsModalVisible(true);
                }}
              >
                +{remainingCount} 更多
              </Badge>
            )}
          </div>
        );
      }
    },
    {
      key: 'status',
      title: '状态',
      render: (_: unknown, record: DNSProvider) => (
        <Badge variant={record.enabled ? 'success' : 'default'}>
          {record.enabled ? '启用' : '禁用'}
        </Badge>
      ),
    },
    {
      key: 'action',
      title: '操作',
      render: (_: unknown, record: DNSProvider) => (
        <div className="flex items-center justify-center gap-1">
          <Tooltip content={testingId === record.id ? '测试中...' : '测试连接'}>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => handleTest(record.id)}
              disabled={testingId === record.id}
            >
              {testingId === record.id ? (
                <RefreshCw className="h-4 w-4 animate-spin" />
              ) : (
                <CheckCircle className="h-4 w-4" />
              )}
            </Button>
          </Tooltip>
          <Tooltip content="编辑">
            <Button size="sm" variant="ghost" onClick={() => handleEdit(record)}>
              <Edit className="h-4 w-4" />
            </Button>
          </Tooltip>
          <Tooltip content="删除">
            <Button size="sm" variant="ghost" onClick={() => handleDelete(record.id)}>
              <Trash2 className="h-4 w-4 text-red-400" />
            </Button>
          </Tooltip>
        </div>
      ),
    },
  ];

  return (
    <>
      {/* ACME 证书设置 */}
      <Card className="mb-6">
        <CardHeader>
          <div className="flex items-center gap-2">
            <Mail className="h-5 w-5 text-green-400" />
            <span>ACME 证书设置</span>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-end gap-4">
            <div className="flex-1">
              <Input
                label="证书申请邮箱"
                value={acmeEmail}
                onChange={(e) => setAcmeEmail(e.target.value)}
                placeholder="用于Let's Encrypt证书申请的联系邮箱"
                disabled={acmeEmailLoading}
              />
            </div>
            <Button
              onClick={handleSaveAcmeEmail}
              disabled={acmeEmailSaving || acmeEmailLoading}
              icon={acmeEmailSaving ? <RefreshCw className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            >
              {acmeEmailSaving ? '保存中...' : '保存'}
            </Button>
          </div>
          <p className="mt-2 text-xs text-foreground-secondary">
            申请SSL证书时，Let's Encrypt会向此邮箱发送证书到期提醒等通知
          </p>
        </CardContent>
      </Card>

      {/* DNS 提供商管理 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Cloud className="h-5 w-5 text-blue-400" />
              <span>DNS 提供商管理</span>
            </div>
            <Button size="sm" onClick={handleAdd} icon={<Plus />}>
              添加提供商
            </Button>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          <Table
            columns={columns}
            data={providers}
            rowKey="id"
            loading={loading}
            emptyText="暂无 DNS 提供商"
          />
        </CardContent>
      </Card>

      <Modal
        open={modalVisible}
        onClose={() => setModalVisible(false)}
        title={editingProvider ? '编辑 DNS 提供商' : '添加 DNS 提供商'}
      >
        <div className="space-y-4">
          <Input
            label="名称"
            value={form.name}
            onChange={(e) => setForm(prev => ({ ...prev, name: e.target.value }))}
            placeholder="例如：阿里云主账号"
            required
          />
          <Select
            label="类型"
            value={form.type}
            onChange={(value) => setForm(prev => ({ ...prev, type: value as DNSProviderType }))}
            options={providerTypeOptions}
          />
          <Input
            label={authFieldConfig.accessKeyLabel}
            type={form.type === 'cloudflare' ? (showAccessKey ? 'text' : 'password') : 'text'}
            value={form.access_key}
            onChange={(e) => setForm(prev => ({ ...prev, access_key: e.target.value }))}
            placeholder={authFieldConfig.accessKeyPlaceholder}
            required
            suffix={form.type === 'cloudflare' && (
              <button
                type="button"
                onClick={() => setShowAccessKey(!showAccessKey)}
                className="text-foreground-subtle hover:text-foreground transition-colors"
              >
                {showAccessKey ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            )}
          />
          {/* Cloudflare 只需要 API Token，不需要 Secret Key */}
          {authFieldConfig.secretKeyRequired && (
            <Input
              label={authFieldConfig.secretKeyLabel}
              type={showSecretKey ? 'text' : 'password'}
              value={form.secret_key}
              onChange={(e) => setForm(prev => ({ ...prev, secret_key: e.target.value }))}
              placeholder={editingProvider && !secretLoaded ? '••••••••（点击眼睛查看）' : (editingProvider ? '留空则不修改' : authFieldConfig.secretKeyPlaceholder)}
              required={!editingProvider}
              suffix={
                <button
                  type="button"
                  onClick={handleToggleSecretKey}
                  disabled={loadingSecret}
                  className="text-foreground-subtle hover:text-foreground transition-colors disabled:opacity-50"
                >
                  {loadingSecret ? (
                    <RefreshCw className="h-4 w-4 animate-spin" />
                  ) : showSecretKey ? (
                    <EyeOff className="h-4 w-4" />
                  ) : (
                    <Eye className="h-4 w-4" />
                  )}
                </button>
              }
            />
          )}
          {/* Cloudflare 提示信息 */}
          {form.type === 'cloudflare' && (
            <div className="text-xs text-foreground-secondary bg-blue-500/10 p-3 rounded-md">
              <p className="font-medium mb-1">💡 Cloudflare API Token 获取方式：</p>
              <ol className="list-decimal list-inside space-y-1">
                <li>登录 Cloudflare Dashboard</li>
                <li>进入 个人资料 → API 令牌</li>
                <li>创建 Token，选择 "编辑区域 DNS" 模板</li>
                <li>选择需要管理的域名区域</li>
              </ol>
            </div>
          )}
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-foreground-secondary">启用</span>
            <Switch
              checked={form.enabled}
              onChange={(checked) => setForm(prev => ({ ...prev, enabled: checked }))}
            />
          </div>
          <div className="flex justify-end gap-3 pt-4">
            <Button variant="secondary" onClick={() => setModalVisible(false)}>
              取消
            </Button>
            <Button
              variant="secondary"
              onClick={handleTestConfig}
              disabled={testingConfig}
              icon={testingConfig ? <RefreshCw className="h-4 w-4 animate-spin" /> : <CheckCircle className="h-4 w-4" />}
            >
              {testingConfig ? '测试中...' : '测试连接'}
            </Button>
            <Button onClick={handleSubmit} disabled={submitting}>
              {submitting ? '提交中...' : (editingProvider ? '更新' : '创建')}
            </Button>
          </div>
        </div>
      </Modal>

      {/* 域名列表弹窗 */}
      <Modal
        open={domainsModalVisible}
        onClose={() => setDomainsModalVisible(false)}
        title={`${viewingDomains.name} - 托管域名`}
      >
        <div className="space-y-2 max-h-96 overflow-y-auto">
          {viewingDomains.domains.map((domain) => (
            <div key={domain} className="flex items-center gap-2 p-2 bg-background-secondary rounded">
              <Globe className="h-4 w-4 text-blue-400" />
              <span className="font-mono text-sm">{domain}</span>
            </div>
          ))}
        </div>
      </Modal>
    </>
  );
}