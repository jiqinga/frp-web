import { useState, useEffect, useCallback, useRef } from 'react';
import { settingApi, type Setting } from '../../../api/setting';
import { githubMirrorApi, type GithubMirror } from '../../../api/githubMirror';
import { toast } from '../../../components/ui/Toast';

// 简单对象浅比较
const shallowEqual = (a: Record<string, string>, b: Record<string, string>): boolean => {
  const keysA = Object.keys(a);
  const keysB = Object.keys(b);
  if (keysA.length !== keysB.length) return false;
  return keysA.every(key => a[key] === b[key]);
};

// 采集间隔相关的设置项 key
export const INTERVAL_SETTINGS = [
  'traffic_interval',
  'server_info_interval',
  'proxy_status_interval',
  'client_check_interval',
  'server_status_check_interval'
];

// 邮件配置相关的设置项 key
export const EMAIL_SETTINGS = [
  'smtp_host',
  'smtp_port',
  'smtp_username',
  'smtp_password',
  'smtp_from',
  'smtp_ssl'
];

export const EMAIL_LABELS: Record<string, string> = {
  smtp_host: 'SMTP服务器地址',
  smtp_port: 'SMTP端口',
  smtp_username: 'SMTP用户名',
  smtp_password: 'SMTP密码',
  smtp_from: '发件人地址',
  smtp_ssl: '启用SSL/TLS'
};

export function useSettings() {
  const [settings, setSettings] = useState<Setting[]>([]);
  const [formValues, setFormValues] = useState<Record<string, string>>({});
  const [intervalValues, setIntervalValues] = useState<Record<string, string>>({});
  const [emailValues, setEmailValues] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const [intervalLoading, setIntervalLoading] = useState(false);
  const [emailLoading, setEmailLoading] = useState(false);
  
  // 原始值存储，用于脏检查
  const originalFormValues = useRef<Record<string, string>>({});
  const originalIntervalValues = useRef<Record<string, string>>({});
  const originalEmailValues = useRef<Record<string, string>>({});

  const fetchSettings = useCallback(async () => {
    try {
      const data = await settingApi.getSettings();
      setSettings(data);
      
      const normalValues: Record<string, string> = {};
      const intervalVals: Record<string, string> = {};
      const emailVals: Record<string, string> = {};
      data.forEach(s => {
        if (INTERVAL_SETTINGS.includes(s.key)) {
          intervalVals[s.key] = s.value;
        } else if (EMAIL_SETTINGS.includes(s.key)) {
          emailVals[s.key] = s.value;
        } else {
          normalValues[s.key] = s.value;
        }
      });
      setFormValues(normalValues);
      setIntervalValues(intervalVals);
      setEmailValues(emailVals);
      // 同步更新原始值
      originalFormValues.current = { ...normalValues };
      originalIntervalValues.current = { ...intervalVals };
      originalEmailValues.current = { ...emailVals };
    } catch {
      // ignore
    }
  }, []);

  useEffect(() => {
    fetchSettings();
  }, [fetchSettings]);

  const handleSubmit = async () => {
    if (shallowEqual(formValues, originalFormValues.current)) {
      toast.info('设置未变化，无需保存');
      return;
    }
    setLoading(true);
    try {
      await Promise.all(
        Object.entries(formValues).map(([key, value]) =>
          settingApi.updateSetting({ key, value })
        )
      );
      originalFormValues.current = { ...formValues };
      toast.success('保存成功');
    } catch {
      toast.error('保存失败');
    } finally {
      setLoading(false);
    }
  };

  const handleIntervalSubmit = async () => {
    if (shallowEqual(intervalValues, originalIntervalValues.current)) {
      toast.info('设置未变化，无需保存');
      return;
    }
    setIntervalLoading(true);
    try {
      await Promise.all(
        Object.entries(intervalValues).map(([key, value]) =>
          settingApi.updateSetting({ key, value })
        )
      );
      originalIntervalValues.current = { ...intervalValues };
      toast.success('采集间隔设置保存成功');
    } catch {
      toast.error('保存失败');
    } finally {
      setIntervalLoading(false);
    }
  };

  const handleEmailSubmit = async () => {
    if (shallowEqual(emailValues, originalEmailValues.current)) {
      toast.info('设置未变化，无需保存');
      return;
    }
    setEmailLoading(true);
    try {
      await Promise.all(
        Object.entries(emailValues).map(([key, value]) =>
          settingApi.updateSetting({ key, value })
        )
      );
      originalEmailValues.current = { ...emailValues };
      toast.success('邮件设置保存成功');
    } catch {
      toast.error('保存失败');
    } finally {
      setEmailLoading(false);
    }
  };

  const normalSettings = settings.filter(s => !INTERVAL_SETTINGS.includes(s.key) && !EMAIL_SETTINGS.includes(s.key));
  const intervalSettingsList = settings.filter(s => INTERVAL_SETTINGS.includes(s.key));
  const emailSettingsList = settings.filter(s => EMAIL_SETTINGS.includes(s.key));

  return {
    normalSettings,
    intervalSettings: intervalSettingsList,
    emailSettings: emailSettingsList,
    formValues,
    setFormValues,
    intervalValues,
    setIntervalValues,
    emailValues,
    setEmailValues,
    loading,
    intervalLoading,
    emailLoading,
    handleSubmit,
    handleIntervalSubmit,
    handleEmailSubmit,
  };
}

export function useMirrors() {
  const [mirrors, setMirrors] = useState<GithubMirror[]>([]);
  const [mirrorModalVisible, setMirrorModalVisible] = useState(false);
  const [editingMirror, setEditingMirror] = useState<GithubMirror | null>(null);
  const [mirrorForm, setMirrorForm] = useState({
    name: '',
    base_url: '',
    description: '',
    enabled: true
  });

  const fetchMirrors = useCallback(async () => {
    try {
      const data = await githubMirrorApi.getAll();
      setMirrors(data);
    } catch {
      toast.error('加载加速源失败');
    }
  }, []);

  useEffect(() => {
    fetchMirrors();
  }, [fetchMirrors]);

  const handleAddMirror = () => {
    setEditingMirror(null);
    setMirrorForm({ name: '', base_url: '', description: '', enabled: true });
    setMirrorModalVisible(true);
  };

  const handleEditMirror = (mirror: GithubMirror) => {
    setEditingMirror(mirror);
    setMirrorForm({
      name: mirror.name,
      base_url: mirror.base_url,
      description: mirror.description || '',
      enabled: mirror.enabled
    });
    setMirrorModalVisible(true);
  };

  const handleDeleteMirror = async (id: number) => {
    if (!confirm('确定删除此加速源吗？')) return;
    try {
      await githubMirrorApi.delete(id);
      toast.success('删除成功');
      fetchMirrors();
    } catch {
      toast.error('删除失败');
    }
  };

  const handleSetDefault = async (id: number) => {
    try {
      await githubMirrorApi.setDefault(id);
      toast.success('设置成功');
      fetchMirrors();
    } catch {
      toast.error('设置失败');
    }
  };

  const handleMirrorSubmit = async () => {
    if (!mirrorForm.name || !mirrorForm.base_url) {
      toast.error('请填写必填字段');
      return;
    }
    try {
      if (editingMirror) {
        await githubMirrorApi.update(editingMirror.id, mirrorForm);
        toast.success('更新成功');
      } else {
        await githubMirrorApi.create(mirrorForm);
        toast.success('创建成功');
      }
      setMirrorModalVisible(false);
      fetchMirrors();
    } catch {
      toast.error('操作失败');
    }
  };

  return {
    mirrors,
    mirrorModalVisible,
    setMirrorModalVisible,
    editingMirror,
    mirrorForm,
    setMirrorForm,
    handleAddMirror,
    handleEditMirror,
    handleDeleteMirror,
    handleSetDefault,
    handleMirrorSubmit,
  };
}