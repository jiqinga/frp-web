import { useState, useEffect, useCallback } from 'react';
import { Plus, FileCode, Search, AlertTriangle } from 'lucide-react';
import { clientApi, type UpdateClientRequest } from '../../api/client';
import type { Client } from '../../types';
import { Button } from '../../components/ui/Button';
import { Input } from '../../components/ui/Input';
import { Modal } from '../../components/ui/Modal';
import { toast } from '../../components/ui/Toast';

// 导入拆分的组件
import { ClientTable } from './components/ClientTable';
import { ClientFormModal } from './components/ClientFormModal';
import { ScriptGeneratorModal } from './components/ScriptGeneratorModal';
import { UpdateModal } from './components/UpdateModal';
import { ConfigViewModal } from './components/ConfigViewModal';
import { LogViewerModal } from './components/LogViewerModal';

// 导入 hooks
import { useClients } from './hooks/useClients';
import { useUpdateWebSocket } from './hooks/useUpdateWebSocket';
import { useLogStream, type LogType } from './hooks/useLogStream';
import { useClientModals } from './hooks/useClientModals';
import { useFrpcControl } from './hooks/useFrpcControl';

export function Component() {
  // 使用自定义 hooks
  const {
    clients,
    total,
    loading,
    page,
    keyword,
    frpServers,
    githubMirrors,
    setPage,
    setKeyword,
    fetchClients,
    loadFrpServers,
    loadGithubMirrors,
    createClient,
    updateClient,
    deleteClient,
    getConfig,
    parseConfig,
  } = useClients();

  const {
    updateProgress,
    updateResult,
    connectWebSocket,
    resetProgress,
  } = useUpdateWebSocket();

  const {
    logs,
    isStreaming,
    logType,
    startLogStream,
    stopLogStream,
    clearLogs,
  } = useLogStream();

  const {
    controlFrpc,
    loadingMap: frpcLoadingMap,
  } = useFrpcControl();

  // 使用模态框状态管理 hook
  const modals = useClientModals();
  
  // 搜索状态
  const [searchValue, setSearchValue] = useState('');

  // 加载数据
  useEffect(() => {
    fetchClients();
  }, [fetchClients, page, keyword]);

  // WebSocket 连接
  useEffect(() => {
    if (modals.updateModalVisible && modals.updatingClient) {
      const cleanup = connectWebSocket(modals.updatingClient, fetchClients);
      return cleanup;
    }
  }, [modals.updateModalVisible, modals.updatingClient, connectWebSocket, fetchClients]);

  // 处理删除确认
  const handleDelete = (id: number) => {
    const client = clients.find(c => c.id === id);
    if (client) {
      modals.openDeleteConfirm(client);
    }
  };

  // 确认删除
  const confirmDelete = async () => {
    if (modals.deletingClient) {
      const success = await deleteClient(modals.deletingClient.id);
      if (success) {
        fetchClients();
      }
    }
    modals.closeDeleteConfirm();
  };

  // 处理表单提交
  const handleSubmit = async (values: Partial<Client>) => {
    let success: boolean;
    if (modals.editingClient) {
      success = await updateClient(modals.editingClient.id, values);
    } else {
      success = await createClient(values);
    }
    if (success) {
      modals.closeFormModal();
      fetchClients();
    }
  };

  // 处理查看配置
  const handleViewConfig = async (client: Client) => {
    modals.openConfigView(client);
    const config = await getConfig(client.id);
    modals.setConfigContent(config || '');
    modals.setConfigLoading(false);
  };

  // 处理日志流开始
  const handleStartLogStream = async (type: LogType, lines: number) => {
    if (modals.logViewClient) {
      await startLogStream(modals.logViewClient.id, type, lines);
    }
  };

  // 执行frpc控制
  const handleFrpcStart = async (client: Client) => {
    if (client.ws_connected) {
      const success = await controlFrpc(client.id, 'start');
      if (success) fetchClients();
    }
  };

  const handleFrpcStop = async (client: Client) => {
    if (client.ws_connected) {
      const success = await controlFrpc(client.id, 'stop');
      if (success) fetchClients();
    }
  };

  const handleFrpcRestart = async (client: Client) => {
    if (client.ws_connected) {
      const success = await controlFrpc(client.id, 'restart');
      if (success) fetchClients();
    }
  };

  // 加载脚本生成器数据
  const handleLoadScriptData = useCallback(async () => {
    await Promise.all([loadFrpServers(), loadGithubMirrors()]);
  }, [loadFrpServers, loadGithubMirrors]);

  // 打开更新对话框
  const handleOpenUpdateModal = async (client: Client) => {
    if (!client.ws_connected) {
      toast.warning('只能更新WS连接的客户端');
      return;
    }
    resetProgress();
    modals.openUpdateModal(client);
    await loadGithubMirrors();
  };

  // 执行更新
  const handleUpdate = async () => {
    if (!modals.updatingClient) return;
    try {
      const req: UpdateClientRequest = {
        update_type: modals.updateType,
        mirror_id: modals.updateMirrorId,
      };
      await clientApi.updateClientSoftware(modals.updatingClient.id, req);
    } catch {
      toast.error('发送更新命令失败');
    }
  };

  // 关闭更新对话框
  const handleCloseUpdateModal = () => {
    modals.closeUpdateModal();
    resetProgress();
  };

  // 处理搜索
  const handleSearch = () => {
    setKeyword(searchValue);
  };

  const handleSearchKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch();
    }
  };

  return (
    <div className="p-6 space-y-6">
      {/* 页面标题和操作栏 */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-foreground">客户端管理</h1>
          <p className="text-sm mt-1 text-foreground-muted">管理 FRP 客户端连接和配置</p>
        </div>
        
        <div className="flex flex-wrap items-center gap-3">
          {/* 搜索框 */}
          <div className="relative">
            <Input
              placeholder="搜索客户端..."
              value={searchValue}
              onChange={(e) => setSearchValue(e.target.value)}
              onKeyDown={handleSearchKeyDown}
              className="w-48 pr-10"
            />
            <Button
              variant="ghost"
              size="sm"
              icon={<Search className="h-4 w-4" />}
              onClick={handleSearch}
              className="absolute right-2 top-1/2 -translate-y-1/2 !p-1"
            />
          </div>
          
          {/* 操作按钮 */}
          <Button onClick={modals.openAddModal} icon={<Plus className="h-4 w-4" />}>
            新增客户端
          </Button>
          <Button
            variant="secondary"
            onClick={modals.openScriptModal}
            icon={<FileCode className="h-4 w-4" />}
          >
            生成注册脚本
          </Button>
        </div>
      </div>

      {/* 客户端表格 */}
      <ClientTable
        clients={clients}
        loading={loading}
        page={page}
        total={total}
        onPageChange={setPage}
        onEdit={modals.openEditModal}
        onDelete={handleDelete}
        onViewConfig={handleViewConfig}
        onUpdate={handleOpenUpdateModal}
        onViewLogs={modals.openLogView}
        onFrpcStart={handleFrpcStart}
        onFrpcStop={handleFrpcStop}
        onFrpcRestart={handleFrpcRestart}
        frpcLoadingMap={frpcLoadingMap}
      />

      {/* 新增/编辑客户端模态框 */}
      <ClientFormModal
        visible={modals.modalVisible}
        editingClient={modals.editingClient}
        onCancel={modals.closeFormModal}
        onSubmit={handleSubmit}
        onParseConfig={parseConfig}
      />

      {/* 脚本生成器模态框 */}
      <ScriptGeneratorModal
        visible={modals.scriptModalVisible}
        frpServers={frpServers}
        githubMirrors={githubMirrors}
        onCancel={modals.closeScriptModal}
        onLoadData={handleLoadScriptData}
      />

      {/* 更新模态框 */}
      <UpdateModal
        visible={modals.updateModalVisible}
        client={modals.updatingClient}
        githubMirrors={githubMirrors}
        updateType={modals.updateType}
        updateMirrorId={modals.updateMirrorId}
        updateProgress={updateProgress}
        updateResult={updateResult}
        onUpdateTypeChange={modals.setUpdateType}
        onMirrorIdChange={modals.setUpdateMirrorId}
        onUpdate={handleUpdate}
        onClose={handleCloseUpdateModal}
      />

      {/* 配置查看模态框 */}
      <ConfigViewModal
        visible={modals.configViewVisible}
        clientName={modals.viewingClient?.name || ''}
        config={modals.configContent}
        loading={modals.configLoading}
        onClose={modals.closeConfigView}
      />

      {/* 日志查看模态框 */}
      <LogViewerModal
        open={modals.logViewVisible}
        onClose={modals.closeLogView}
        clientName={modals.logViewClient?.name || ''}
        logs={logs}
        isStreaming={isStreaming}
        logType={logType}
        onStartStream={handleStartLogStream}
        onStopStream={stopLogStream}
        onClearLogs={clearLogs}
      />

      {/* 删除确认模态框 */}
      <Modal
        open={modals.deleteConfirmVisible}
        onClose={modals.closeDeleteConfirm}
        title="确认删除客户端"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={modals.closeDeleteConfirm}>
              取消
            </Button>
            <Button variant="danger" onClick={confirmDelete}>
              确定删除
            </Button>
          </>
        }
      >
        <div className="flex items-start gap-3">
          <div className="flex-shrink-0 w-10 h-10 rounded-full bg-red-500/20 flex items-center justify-center">
            <AlertTriangle className="h-5 w-5 text-red-400" />
          </div>
          <div>
            <p className="text-foreground-secondary">
              删除客户端 "<span className="font-medium text-foreground">{modals.deletingClient?.name}</span>" 将同时删除所有关联的代理配置。
            </p>
            <p className="text-sm mt-1 text-foreground-muted">此操作不可恢复，确定继续吗？</p>
          </div>
        </div>
      </Modal>
    </div>
  );
}