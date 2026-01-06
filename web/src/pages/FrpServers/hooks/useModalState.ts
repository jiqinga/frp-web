import { useState, useCallback } from 'react';
import type { FrpServer, FrpsMetrics } from '../../../api/frpServer';
import { frpServerApi } from '../../../api/frpServer';
import { toast } from '../../../components/ui/Toast';

export function useModalState() {
  const [modalVisible, setModalVisible] = useState(false);
  const [editingServer, setEditingServer] = useState<FrpServer | null>(null);
  const [serverType, setServerType] = useState<'local' | 'remote'>('local');
  const [configContent, setConfigContent] = useState('');
  const [importLoading, setImportLoading] = useState(false);
  const [installLogVisible, setInstallLogVisible] = useState(false);
  const [upgradeVisible, setUpgradeVisible] = useState(false);
  const [upgradingServerId, setUpgradingServerId] = useState<number | null>(null);
  const [upgradeVersion, setUpgradeVersion] = useState('');
  const [upgradeMirrorId, setUpgradeMirrorId] = useState<number | undefined>();
  const [installVisible, setInstallVisible] = useState(false);
  const [installingServerId, setInstallingServerId] = useState<number | null>(null);
  const [installMirrorId, setInstallMirrorId] = useState<number | undefined>();
  const [reinstallVisible, setReinstallVisible] = useState(false);
  const [reinstallingServerId, setReinstallingServerId] = useState<number | null>(null);
  const [reinstallMirrorId, setReinstallMirrorId] = useState<number | undefined>();
  const [reinstallRegenerateAuth, setReinstallRegenerateAuth] = useState(false);
  const [metricsModalVisible, setMetricsModalVisible] = useState(false);
  const [metricsLoading, setMetricsLoading] = useState(false);
  const [currentMetrics, setCurrentMetrics] = useState<FrpsMetrics | null>(null);
  const [metricsServerId, setMetricsServerId] = useState<number | null>(null);
  const [deleteConfirmVisible, setDeleteConfirmVisible] = useState(false);
  const [serverToDelete, setServerToDelete] = useState<FrpServer | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [authInfoVisible, setAuthInfoVisible] = useState(false);
  const [authInfoServer, setAuthInfoServer] = useState<FrpServer | null>(null);
  const [logsVisible, setLogsVisible] = useState(false);
  const [logsContent, setLogsContent] = useState('');

  const handleAdd = useCallback(() => {
    setEditingServer(null);
    setServerType('local');
    setModalVisible(true);
  }, []);

  const handleEdit = useCallback((server: FrpServer) => {
    setEditingServer(server);
    setServerType(server.server_type as 'local' | 'remote');
    setModalVisible(true);
  }, []);

  const handleDelete = useCallback((server: FrpServer) => {
    setServerToDelete(server);
    setDeleteConfirmVisible(true);
  }, []);

  const handleViewAuth = useCallback((server: FrpServer) => {
    setAuthInfoServer(server);
    setAuthInfoVisible(true);
  }, []);

  const handleViewMetrics = useCallback(async (id: number) => {
    setMetricsServerId(id);
    setCurrentMetrics(null);
    setMetricsLoading(true);
    setMetricsModalVisible(true);
    try {
      const metrics = await frpServerApi.getMetrics(id);
      setCurrentMetrics(metrics);
    } catch {
      toast.error('获取指标失败');
      setCurrentMetrics(null);
    } finally {
      setMetricsLoading(false);
    }
  }, []);

  const handleRemoteInstall = useCallback((id: number) => {
    setInstallingServerId(id);
    setInstallMirrorId(undefined);
    setInstallVisible(true);
  }, []);

  const handleRemoteReinstall = useCallback((id: number) => {
    setReinstallingServerId(id);
    setReinstallMirrorId(undefined);
    setReinstallRegenerateAuth(false);
    setReinstallVisible(true);
  }, []);

  const showUpgradeModal = useCallback((id: number) => {
    setUpgradingServerId(id);
    setUpgradeVersion('');
    setUpgradeMirrorId(undefined);
    setUpgradeVisible(true);
  }, []);

  return {
    modalVisible, setModalVisible, editingServer, setEditingServer,
    serverType, setServerType, configContent, setConfigContent,
    importLoading, setImportLoading, installLogVisible, setInstallLogVisible,
    upgradeVisible, setUpgradeVisible, upgradingServerId, upgradeVersion, setUpgradeVersion,
    upgradeMirrorId, setUpgradeMirrorId, installVisible, setInstallVisible,
    installingServerId, installMirrorId, setInstallMirrorId,
    reinstallVisible, setReinstallVisible, reinstallingServerId,
    reinstallMirrorId, setReinstallMirrorId, reinstallRegenerateAuth, setReinstallRegenerateAuth,
    metricsModalVisible, setMetricsModalVisible, metricsLoading, currentMetrics, setCurrentMetrics,
    metricsServerId, setMetricsServerId, deleteConfirmVisible, setDeleteConfirmVisible,
    serverToDelete, setServerToDelete, deleteLoading, setDeleteLoading,
    authInfoVisible, setAuthInfoVisible, authInfoServer, setAuthInfoServer,
    logsVisible, setLogsVisible, logsContent, setLogsContent,
    handleAdd, handleEdit, handleDelete, handleViewAuth, handleViewMetrics,
    handleRemoteInstall, handleRemoteReinstall, showUpgradeModal,
  };
}