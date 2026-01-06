import { useState } from 'react';
import { Mail, Send, Save } from 'lucide-react';
import { Button, Input, Card, CardHeader, CardContent, Switch } from '../../../components/ui';
import { toast } from '../../../components/ui/Toast';
import { settingApi } from '../../../api/setting';
import { EMAIL_LABELS } from '../hooks/useSettings';

// SMTP 相关设置项
const SMTP_SETTINGS = ['smtp_host', 'smtp_port', 'smtp_username', 'smtp_password', 'smtp_from', 'smtp_ssl'];

interface EmailSettingsProps {
  emailValues: Record<string, string>;
  setEmailValues: React.Dispatch<React.SetStateAction<Record<string, string>>>;
  loading: boolean;
  onSubmit: () => void;
}

export function EmailSettings({
  emailValues,
  setEmailValues,
  loading,
  onSubmit,
}: EmailSettingsProps) {
  const [testEmail, setTestEmail] = useState('');
  const [testing, setTesting] = useState(false);

  const handleTest = async () => {
    if (!testEmail) return;
    setTesting(true);
    try {
      await settingApi.testEmail(testEmail);
      toast.success('测试邮件已发送，请检查收件箱');
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } };
      toast.error(error.response?.data?.message || '发送失败');
    } finally {
      setTesting(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Mail className="h-5 w-5 text-indigo-400" />
          <span>邮件服务器设置</span>
        </div>
      </CardHeader>
      <CardContent>
        <div className="max-w-xl space-y-4">
          {SMTP_SETTINGS.map(key => (
            <div key={key}>
              {key === 'smtp_ssl' ? (
                <Switch
                  checked={emailValues[key] === 'true'}
                  onChange={(checked) => setEmailValues(prev => ({ ...prev, [key]: checked ? 'true' : 'false' }))}
                  label={EMAIL_LABELS[key]}
                />
              ) : (
                <Input
                  label={EMAIL_LABELS[key]}
                  type={key === 'smtp_password' ? 'password' : 'text'}
                  value={emailValues[key] || ''}
                  onChange={(e) => setEmailValues(prev => ({ ...prev, [key]: e.target.value }))}
                  placeholder={key === 'smtp_port' ? '465' : ''}
                />
              )}
            </div>
          ))}
          
          <Button onClick={onSubmit} loading={loading} icon={<Save />}>
            保存设置
          </Button>

          <div className="mt-6 border-t pt-4 border-border">
            <h4 className="text-sm font-medium mb-3 text-foreground-secondary">测试邮件配置</h4>
            <div className="flex gap-2">
              <Input
                placeholder="输入测试邮箱地址"
                value={testEmail}
                onChange={(e) => setTestEmail(e.target.value)}
                className="flex-1"
              />
              <Button onClick={handleTest} loading={testing} icon={<Send />} variant="secondary">
                发送测试
              </Button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}