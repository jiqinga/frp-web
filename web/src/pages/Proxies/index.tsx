import { useState, useCallback } from 'react';
import { Plus, Filter, X } from 'lucide-react';
import { Button } from '../../components/ui/Button';
import { Select, type SelectOption } from '../../components/ui/Select';
import type { Proxy } from '../../types';
import { ProxyTable, ProxyFormModal } from './components';
import { useProxies } from './hooks';
import { PluginUsageModal } from '../../components/PluginUsageModal';
import { TrafficHistoryModal } from '../../components/TrafficHistoryModal';

/**
 * Proxies 代理管理页面主组件
 * 
 * 功能：
 * - 代理列表展示（支持按客户端筛选）
 * - 新增/编辑/删除代理
 * - 启用/禁用代理
 * - 查看插件使用说明
 * - 查看流量历史
 */
export function Component() {
  // ==================== 数据管理 ====================
  const {
    proxies,
    clients,
    loading,
    selectedClient,
    onlineClientIds,
    setSelectedClient,
    createProxy,
    updateProxy,
    deleteProxy,
    toggleProxy,
    getClientById,
    getAccessUrl,
    getServerAddrForProxy,
    isClientOnline,
  } = useProxies();

  // ==================== 表单模态框状态 ====================
  const [modalVisible, setModalVisible] = useState(false);
  const [editingProxy, setEditingProxy] = useState<Proxy | null>(null);

  // ==================== 插件使用说明模态框状态 ====================
  const [pluginUsageModalVisible, setPluginUsageModalVisible] = useState(false);
  const [selectedProxyForUsage, setSelectedProxyForUsage] = useState<Proxy | null>(null);

  // ==================== 流量历史模态框状态 ====================
  const [trafficHistoryModalVisible, setTrafficHistoryModalVisible] = useState(false);
  const [selectedProxyForTraffic, setSelectedProxyForTraffic] = useState<Proxy | null>(null);

  // ==================== Switch loading 状态 ====================
  const [togglingIds, setTogglingIds] = useState<Set<number>>(new Set());

  // ==================== 事件处理函数 ====================
  
  // 判断当前筛选的客户端是否在线（用于控制新增按钮）
  const isSelectedClientOnline = selectedClient ? isClientOnline(selectedClient) : true;
  // 检查是否有任何在线客户端
  const hasOnlineClients = clients.some(c => onlineClientIds.has(c.id!));

  /**
   * 打开新增代理表单
   */
  const handleAdd = () => {
    setEditingProxy(null);
    setModalVisible(true);
  };

  /**
   * 打开编辑代理表单
   */
  const handleEdit = (proxy: Proxy) => {
    setEditingProxy(proxy);
    setModalVisible(true);
  };

  /**
   * 删除代理
   * @param id 代理ID
   * @param deleteDNS 是否同时删除DNS记录，默认为true
   */
  const handleDelete = async (id: number, deleteDNS: boolean = true) => {
    await deleteProxy(id, deleteDNS);
  };

  /**
   * 切换代理启用/禁用状态
   */
  const handleToggle = async (id: number) => {
    setTogglingIds(prev => new Set(prev).add(id));
    try {
      await toggleProxy(id);
    } finally {
      setTogglingIds(prev => {
        const next = new Set(prev);
        next.delete(id);
        return next;
      });
    }
  };

  /**
   * 检查代理是否正在切换状态
   */
  const isToggling = useCallback((id: number) => togglingIds.has(id), [togglingIds]);

  /**
   * 提交表单（新增或编辑）
   */
  const handleSubmit = async (values: Partial<Proxy>) => {
    if (editingProxy) {
      await updateProxy(editingProxy.id, values);
    } else {
      await createProxy(values);
    }
    setModalVisible(false);
  };

  /**
   * 显示插件使用说明
   */
  const handleShowPluginUsage = (proxy: Proxy) => {
    setSelectedProxyForUsage(proxy);
    setPluginUsageModalVisible(true);
  };

  /**
   * 显示流量历史
   */
  const handleShowTrafficHistory = (proxy: Proxy) => {
    setSelectedProxyForTraffic(proxy);
    setTrafficHistoryModalVisible(true);
  };

  /**
   * 清除客户端筛选，显示所有代理
   */
  const handleClearFilter = () => {
    setSelectedClient(undefined);
  };

  // 客户端选项
  const clientOptions: SelectOption[] = clients.map(c => ({
    value: c.id!,
    label: c.name,
  }));

  // ==================== 渲染 ====================
  return (
    <div className="p-6 space-y-6">
      {/* 页面标题和工具栏 */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-foreground">代理管理</h1>
          <p className="mt-1 text-foreground-muted">管理 FRP 代理配置，支持 TCP、UDP、HTTP、HTTPS、STCP 等类型</p>
        </div>
        
        <Button
          icon={<Plus className="h-4 w-4" />}
          onClick={handleAdd}
          disabled={selectedClient ? !isSelectedClientOnline : !hasOnlineClients}
          title={selectedClient && !isSelectedClientOnline ? '当前筛选的客户端离线，无法新增代理' : (!hasOnlineClients ? '没有在线的客户端' : undefined)}
        >
          新增代理
        </Button>
      </div>

      {/* 筛选工具栏 */}
      <div className="flex flex-wrap items-center gap-3">
        <div className="flex items-center gap-2">
          <Filter className="h-4 w-4 text-foreground-muted" />
          <span className="text-sm text-foreground-muted">筛选:</span>
        </div>
        
        <Select
          value={selectedClient}
          onChange={(value) => setSelectedClient(value as number)}
          options={clientOptions}
          placeholder="全部客户端"
          className="w-48"
        />
        
        {selectedClient && (
          <Button
            variant="ghost"
            size="sm"
            icon={<X className="h-4 w-4" />}
            onClick={handleClearFilter}
          >
            清除筛选
          </Button>
        )}
        
        {/* 统计信息 */}
        <div className="ml-auto text-sm text-foreground-muted">
          共 <span className="text-indigo-400 font-medium">{proxies.length}</span> 个代理
          {selectedClient && (
            <span className="ml-2">
              (已筛选客户端: <span className="text-indigo-400">{clients.find(c => c.id === selectedClient)?.name}</span>)
            </span>
          )}
        </div>
      </div>

      {/* 代理列表表格 */}
      <ProxyTable
        proxies={proxies}
        loading={loading}
        getClientById={getClientById}
        getAccessUrl={getAccessUrl}
        isClientOnline={isClientOnline}
        onEdit={handleEdit}
        onDelete={handleDelete}
        onToggle={handleToggle}
        isToggling={isToggling}
        onShowPluginUsage={handleShowPluginUsage}
        onShowTrafficHistory={handleShowTrafficHistory}
      />

      {/* 新增/编辑代理表单模态框 */}
      <ProxyFormModal
        visible={modalVisible}
        editingProxy={editingProxy}
        clients={clients}
        selectedClient={selectedClient}
        onlineClientIds={onlineClientIds}
        onSubmit={handleSubmit}
        onCancel={() => setModalVisible(false)}
      />

      {/* 插件使用说明模态框 */}
      <PluginUsageModal
        visible={pluginUsageModalVisible}
        onClose={() => {
          setPluginUsageModalVisible(false);
          setSelectedProxyForUsage(null);
        }}
        proxy={selectedProxyForUsage}
        serverAddr={selectedProxyForUsage ? getServerAddrForProxy(selectedProxyForUsage) : ''}
      />

      {/* 流量历史模态框 */}
      <TrafficHistoryModal
        visible={trafficHistoryModalVisible}
        onClose={() => {
          setTrafficHistoryModalVisible(false);
          setSelectedProxyForTraffic(null);
        }}
        proxyId={selectedProxyForTraffic?.id || null}
        proxyName={selectedProxyForTraffic?.name || ''}
      />
    </div>
  );
}