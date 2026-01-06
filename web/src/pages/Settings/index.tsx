import { Settings as SettingsIcon, Clock, Globe, Mail, Users, Cloud } from 'lucide-react';
import { Tabs } from '../../components/ui';
import { BasicSettings, IntervalSettings, MirrorSettings, EmailSettings, RecipientSettings, DNSSettings } from './components';
import { useSettings, useMirrors } from './hooks';

export function Component() {
  const {
    normalSettings,
    intervalSettings,
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
  } = useSettings();

  const {
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
  } = useMirrors();

  const tabItems = [
    {
      key: 'basic',
      label: '基础设置',
      icon: <SettingsIcon className="h-4 w-4" />,
      children: (
        <BasicSettings
          settings={normalSettings}
          formValues={formValues}
          setFormValues={setFormValues}
          loading={loading}
          onSubmit={handleSubmit}
        />
      ),
    },
    {
      key: 'interval',
      label: '采集间隔',
      icon: <Clock className="h-4 w-4" />,
      children: (
        <IntervalSettings
          settings={intervalSettings}
          intervalValues={intervalValues}
          setIntervalValues={setIntervalValues}
          loading={intervalLoading}
          onSubmit={handleIntervalSubmit}
        />
      ),
    },
    {
      key: 'email',
      label: '邮件设置',
      icon: <Mail className="h-4 w-4" />,
      children: (
        <EmailSettings
          emailValues={emailValues}
          setEmailValues={setEmailValues}
          loading={emailLoading}
          onSubmit={handleEmailSubmit}
        />
      ),
    },
    {
      key: 'mirror',
      label: '加速源管理',
      icon: <Globe className="h-4 w-4" />,
      children: (
        <MirrorSettings
          mirrors={mirrors}
          mirrorModalVisible={mirrorModalVisible}
          setMirrorModalVisible={setMirrorModalVisible}
          editingMirror={editingMirror}
          mirrorForm={mirrorForm}
          setMirrorForm={setMirrorForm}
          onAdd={handleAddMirror}
          onEdit={handleEditMirror}
          onDelete={handleDeleteMirror}
          onSetDefault={handleSetDefault}
          onSubmit={handleMirrorSubmit}
        />
      ),
    },
    {
      key: 'recipient',
      label: '告警接收人',
      icon: <Users className="h-4 w-4" />,
      children: <RecipientSettings />,
    },
    {
      key: 'dns',
      label: 'DNS 管理',
      icon: <Cloud className="h-4 w-4" />,
      children: <DNSSettings />,
    },
  ];

  return (
    <div className="space-y-6 p-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground">系统设置</h1>
        <p className="mt-1 text-foreground-muted">配置系统参数和加速源</p>
      </div>

      <Tabs items={tabItems} defaultActiveKey="basic" />
    </div>
  );
}