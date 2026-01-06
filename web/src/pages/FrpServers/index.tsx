import { useCallback, useState } from 'react';
import { Plus, Server, Copy, Check } from 'lucide-react';
import { Button } from '../../components/ui/Button';
import { ConfirmModal } from '../../components/ui/ConfirmModal';
import { toast } from '../../components/ui/Toast';
import { ServerTable, ServerFormModal, InstallLogModal, RemoteOperationModals, MetricsModal, StatsCards } from './components';
import { useFrpServers, useRemoteOperations, useModalState, useOperationLoading } from './hooks';
import type { FrpServer } from '../../api/frpServer';
import { cn } from '../../utils/cn';

function FrpServers() {
  const { servers, loading, mirrors, loadServers, createServer, updateServer, deleteServer, testConnection, testSSH, refreshLocalVersion, refreshRemoteVersion, generateToken, parseConfig } = useFrpServers();
  const { installLogs, downloadProgress, remoteInstall, remoteReinstall, remoteUpgrade, remoteStart, remoteStop, remoteRestart, getRemoteLogs, clearLogs } = useRemoteOperations();
  const modal = useModalState();
  const { isLoading, withLoading } = useOperationLoading();

  const confirmDelete = useCallback(async () => {
    if (!modal.serverToDelete?.id) return;
    modal.setDeleteLoading(true);
    try {
      await deleteServer(modal.serverToDelete.id);
      modal.setDeleteConfirmVisible(false);
      modal.setServerToDelete(null);
    } catch { toast.error('删除失败'); }
    finally { modal.setDeleteLoading(false); }
  }, [modal.serverToDelete, deleteServer]);

  const handleSubmit = useCallback(async (values: Partial<FrpServer>) => {
    try {
      if (modal.editingServer?.id) await updateServer(modal.editingServer.id, values);
      else await createServer(values);
      modal.setModalVisible(false);
    } catch { toast.error('操作失败'); }
  }, [modal.editingServer, createServer, updateServer]);

  const handleTest = useCallback(async (server: FrpServer) => {
    await withLoading(server.id!, 'testConnection', async () => {
      try { await testConnection(server); } catch { toast.error('连接测试失败'); }
    });
  }, [testConnection, withLoading]);

  const handleTestSSH = useCallback(async (id: number) => {
    await withLoading(id, 'testSSH', async () => {
      try { await testSSH(id); toast.success('SSH 连接测试成功'); } catch { toast.error('SSH 连接测试失败'); }
    });
  }, [testSSH, withLoading]);

  const handleRefreshVersion = useCallback(async (id: number) => {
    await withLoading(id, 'refreshVersion', async () => {
      try { await refreshRemoteVersion(id); } catch { toast.error('刷新版本失败'); }
    });
  }, [refreshRemoteVersion, withLoading]);

  const handleLocalRefreshVersion = useCallback(async (id: number) => {
    await withLoading(id, 'refreshLocalVersion', async () => {
      try { await refreshLocalVersion(id); } catch { toast.error('刷新版本失败'); }
    });
  }, [refreshLocalVersion, withLoading]);

  const confirmRemoteInstall = useCallback(async () => {
    if (!modal.installingServerId) return;
    modal.setInstallVisible(false);
    modal.setInstallLogVisible(true);
    clearLogs();
    await remoteInstall(modal.installingServerId, modal.installMirrorId, loadServers);
  }, [modal.installingServerId, modal.installMirrorId, remoteInstall, clearLogs, loadServers]);

  const confirmRemoteReinstall = useCallback(async () => {
    if (!modal.reinstallingServerId) return;
    modal.setReinstallVisible(false);
    modal.setInstallLogVisible(true);
    clearLogs();
    await remoteReinstall(modal.reinstallingServerId, modal.reinstallRegenerateAuth, modal.reinstallMirrorId, loadServers);
  }, [modal.reinstallingServerId, modal.reinstallRegenerateAuth, modal.reinstallMirrorId, remoteReinstall, clearLogs, loadServers]);

  const handleRemoteUpgrade = useCallback(async () => {
    if (!modal.upgradingServerId) { toast.error('请选择服务器'); return; }
    modal.setUpgradeVisible(false);
    modal.setInstallLogVisible(true);
    clearLogs();
    await remoteUpgrade(modal.upgradingServerId, modal.upgradeVersion || '0.65.0', modal.upgradeMirrorId, loadServers);
  }, [modal.upgradingServerId, modal.upgradeVersion, modal.upgradeMirrorId, remoteUpgrade, clearLogs, loadServers]);

  const handleRemoteStart = useCallback(async (id: number) => {
    await withLoading(id, 'start', async () => {
      try { await remoteStart(id); await loadServers(); } catch { toast.error('启动失败'); }
    });
  }, [remoteStart, loadServers, withLoading]);

  const handleRemoteStop = useCallback(async (id: number) => {
    await withLoading(id, 'stop', async () => {
      try { await remoteStop(id); await loadServers(); } catch { toast.error('停止失败'); }
    });
  }, [remoteStop, loadServers, withLoading]);

  const handleRemoteRestart = useCallback(async (id: number) => {
    await withLoading(id, 'restart', async () => {
      try { await remoteRestart(id); await loadServers(); } catch { toast.error('重启失败'); }
    });
  }, [remoteRestart, loadServers, withLoading]);

  const handleViewLogs = useCallback(async (id: number) => {
    await withLoading(id, 'viewLogs', async () => {
      try { const logs = await getRemoteLogs(id); modal.setLogsContent(logs.logs || '暂无日志'); modal.setLogsVisible(true); }
      catch { toast.error('获取日志失败'); }
    });
  }, [getRemoteLogs, withLoading]);

  const handleCloseInstallLog = useCallback(() => { modal.setInstallLogVisible(false); clearLogs(); }, [clearLogs]);

  const handleImportConfig = useCallback(async () => {
    if (!modal.configContent.trim()) { toast.error('请输入配置内容'); return; }
    modal.setImportLoading(true);
    try { await parseConfig(modal.configContent); toast.success('配置解析成功'); }
    catch { toast.error('配置解析失败'); }
    finally { modal.setImportLoading(false); }
  }, [modal.configContent, parseConfig]);

  return (
    <div className="p-6 space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-lg bg-indigo-500/20 border border-indigo-500/30"><Server className="w-6 h-6 text-indigo-400" /></div>
          <div><h1 className="text-xl font-bold text-foreground">FRP 服务器管理</h1><p className="text-sm text-foreground-muted">管理本地和远程 FRP 服务器</p></div>
        </div>
        <Button variant="primary" onClick={modal.handleAdd} icon={<Plus className="w-4 h-4" />}>添加服务器</Button>
      </div>

      <StatsCards servers={servers} />

      <div className="relative">
        <div className="absolute -top-px left-0 right-0 h-px bg-gradient-to-r from-transparent via-indigo-500/50 to-transparent" />
        <ServerTable servers={servers} loading={loading} onEdit={modal.handleEdit} onDelete={modal.handleDelete} onTestConnection={handleTest}
          onTestSSH={handleTestSSH} onRemoteInstall={modal.handleRemoteInstall} onRemoteStart={handleRemoteStart} onRemoteStop={handleRemoteStop}
          onRemoteRestart={handleRemoteRestart} onViewAuth={modal.handleViewAuth} onRefreshRemoteVersion={handleRefreshVersion}
          onRefreshLocalVersion={handleLocalRefreshVersion} onViewMetrics={modal.handleViewMetrics} onRemoteReinstall={modal.handleRemoteReinstall}
          onRemoteUpgrade={modal.showUpgradeModal} onViewLogs={handleViewLogs} isLoading={isLoading} />
      </div>

      <ServerFormModal visible={modal.modalVisible} editingServer={modal.editingServer} serverType={modal.serverType} mirrors={mirrors}
        configContent={modal.configContent} importLoading={modal.importLoading} onServerTypeChange={modal.setServerType}
        onConfigContentChange={modal.setConfigContent} onImportConfig={handleImportConfig} onGenerateToken={generateToken}
        onCancel={() => modal.setModalVisible(false)} onSubmit={handleSubmit} />

      <InstallLogModal visible={modal.installLogVisible} logs={installLogs} downloadProgress={downloadProgress} onClose={handleCloseInstallLog} />

      <RemoteOperationModals mirrors={mirrors} upgradeVisible={modal.upgradeVisible} upgradeVersion={modal.upgradeVersion}
        upgradeMirrorId={modal.upgradeMirrorId} installVisible={modal.installVisible} installMirrorId={modal.installMirrorId}
        reinstallVisible={modal.reinstallVisible} reinstallMirrorId={modal.reinstallMirrorId} reinstallRegenerateAuth={modal.reinstallRegenerateAuth}
        onUpgradeCancel={() => modal.setUpgradeVisible(false)} onUpgradeConfirm={handleRemoteUpgrade} onUpgradeVersionChange={modal.setUpgradeVersion}
        onUpgradeMirrorChange={modal.setUpgradeMirrorId} onInstallCancel={() => modal.setInstallVisible(false)} onInstallConfirm={confirmRemoteInstall}
        onInstallMirrorChange={modal.setInstallMirrorId} onReinstallCancel={() => modal.setReinstallVisible(false)} onReinstallConfirm={confirmRemoteReinstall}
        onReinstallMirrorChange={modal.setReinstallMirrorId} onReinstallRegenerateAuthChange={modal.setReinstallRegenerateAuth} />

      <MetricsModal visible={modal.metricsModalVisible} loading={modal.metricsLoading} metrics={modal.currentMetrics} serverId={modal.metricsServerId}
        onClose={() => { modal.setMetricsModalVisible(false); modal.setMetricsServerId(null); modal.setCurrentMetrics(null); }} />

      <ConfirmModal open={modal.deleteConfirmVisible} onClose={() => { modal.setDeleteConfirmVisible(false); modal.setServerToDelete(null); }}
        onConfirm={confirmDelete} title="确认删除" content={`确定要删除服务器 "${modal.serverToDelete?.name}" 吗？此操作不可恢复。`}
        type="warning" confirmText="删除" cancelText="取消" loading={modal.deleteLoading} />

      <ConfirmModal open={modal.authInfoVisible} onClose={() => { modal.setAuthInfoVisible(false); modal.setAuthInfoServer(null); }} title="认证信息"
        content={<AuthInfoContent server={modal.authInfoServer} />}
        type="info" showCancel={false} confirmText="关闭" />

      <ConfirmModal open={modal.logsVisible} onClose={() => { modal.setLogsVisible(false); modal.setLogsContent(''); }} title="服务器日志"
        content={<pre className="max-h-96 overflow-auto p-3 rounded-lg text-xs font-mono whitespace-pre-wrap break-all bg-surface-elevated text-foreground border border-border">{modal.logsContent}</pre>}
        type="info" showCancel={false} confirmText="关闭" className="max-w-4xl" fullWidthContent />
    </div>
  );
}

export function Component() { return <FrpServers />; }

function AuthInfoContent({ server }: { server: FrpServer | null }) {
  const [copiedField, setCopiedField] = useState<string | null>(null);
  
  const copyToClipboard = async (text: string, field: string) => {
    await navigator.clipboard.writeText(text);
    setCopiedField(field);
    setTimeout(() => setCopiedField(null), 2000);
  };
  
  const getServerAddress = () => {
    if (!server) return '';
    // 远程服务器使用 ssh_host，本地服务器使用 host（如果是 0.0.0.0 则显示 localhost）
    if (server.server_type === 'remote' && server.ssh_host) return server.ssh_host;
    return server.host === '0.0.0.0' ? 'localhost' : server.host;
  };
  
  const token = server?.token || '未设置';
  const address = `${getServerAddress()}:${server?.bind_port}`;
  
  return (
    <div className="space-y-3">
      <div className="flex items-start gap-2">
        <span className="w-16 shrink-0 text-foreground-muted">Token:</span>
        <code className={cn("flex-1 px-2 py-1 rounded text-sm font-mono break-all bg-surface-elevated text-indigo-500 dark:text-indigo-300")}>{token}</code>
        <Button
          variant="ghost"
          size="sm"
          icon={copiedField === 'token' ? <Check className="w-4 h-4 text-green-500" /> : <Copy className="w-4 h-4" />}
          onClick={() => copyToClipboard(token, 'token')}
          className="!p-1"
        />
      </div>
      <div className="flex items-center gap-2">
        <span className="w-16 shrink-0 text-foreground-muted">地址:</span>
        <code className={cn("flex-1 px-2 py-1 rounded text-sm font-mono bg-surface-elevated text-emerald-500 dark:text-emerald-300")}>{address}</code>
        <Button
          variant="ghost"
          size="sm"
          icon={copiedField === 'address' ? <Check className="w-4 h-4 text-green-500" /> : <Copy className="w-4 h-4" />}
          onClick={() => copyToClipboard(address, 'address')}
          className="!p-1"
        />
      </div>
    </div>
  );
}