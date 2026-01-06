import { useState, useCallback } from 'react';
import type { Client } from '../../../types';

export interface ClientModalsState {
  // 表单模态框
  modalVisible: boolean;
  editingClient: Client | null;
  // 脚本生成器
  scriptModalVisible: boolean;
  // 更新模态框
  updateModalVisible: boolean;
  updatingClient: Client | null;
  // 批量更新
  batchUpdateModalVisible: boolean;
  // 删除确认
  deleteConfirmVisible: boolean;
  deletingClient: Client | null;
  // 配置查看
  configViewVisible: boolean;
  viewingClient: Client | null;
  configContent: string;
  configLoading: boolean;
  // 日志查看
  logViewVisible: boolean;
  logViewClient: Client | null;
  // 更新相关
  updateType: 'frpc' | 'daemon';
  updateMirrorId: number | undefined;
}

export interface ClientModalsActions {
  // 表单模态框
  openAddModal: () => void;
  openEditModal: (client: Client) => void;
  closeFormModal: () => void;
  // 脚本生成器
  openScriptModal: () => void;
  closeScriptModal: () => void;
  // 更新模态框
  openUpdateModal: (client: Client) => void;
  closeUpdateModal: () => void;
  // 批量更新
  openBatchUpdateModal: () => void;
  closeBatchUpdateModal: () => void;
  // 删除确认
  openDeleteConfirm: (client: Client) => void;
  closeDeleteConfirm: () => void;
  // 配置查看
  openConfigView: (client: Client) => void;
  closeConfigView: () => void;
  setConfigContent: (content: string) => void;
  setConfigLoading: (loading: boolean) => void;
  // 日志查看
  openLogView: (client: Client) => void;
  closeLogView: () => void;
  // 更新相关
  setUpdateType: (type: 'frpc' | 'daemon') => void;
  setUpdateMirrorId: (id: number | undefined) => void;
}

export function useClientModals(): ClientModalsState & ClientModalsActions {
  // 表单模态框状态
  const [modalVisible, setModalVisible] = useState(false);
  const [editingClient, setEditingClient] = useState<Client | null>(null);
  
  // 脚本生成器状态
  const [scriptModalVisible, setScriptModalVisible] = useState(false);
  
  // 更新模态框状态
  const [updateModalVisible, setUpdateModalVisible] = useState(false);
  const [updatingClient, setUpdatingClient] = useState<Client | null>(null);
  
  // 批量更新状态
  const [batchUpdateModalVisible, setBatchUpdateModalVisible] = useState(false);
  
  // 删除确认状态
  const [deleteConfirmVisible, setDeleteConfirmVisible] = useState(false);
  const [deletingClient, setDeletingClient] = useState<Client | null>(null);
  
  // 配置查看状态
  const [configViewVisible, setConfigViewVisible] = useState(false);
  const [viewingClient, setViewingClient] = useState<Client | null>(null);
  const [configContent, setConfigContent] = useState('');
  const [configLoading, setConfigLoading] = useState(false);
  
  // 日志查看状态
  const [logViewVisible, setLogViewVisible] = useState(false);
  const [logViewClient, setLogViewClient] = useState<Client | null>(null);
  
  // 更新相关状态
  const [updateType, setUpdateType] = useState<'frpc' | 'daemon'>('frpc');
  const [updateMirrorId, setUpdateMirrorId] = useState<number | undefined>(undefined);

  // 表单模态框操作
  const openAddModal = useCallback(() => {
    setEditingClient(null);
    setModalVisible(true);
  }, []);

  const openEditModal = useCallback((client: Client) => {
    setEditingClient(client);
    setModalVisible(true);
  }, []);

  const closeFormModal = useCallback(() => {
    setModalVisible(false);
  }, []);

  // 脚本生成器操作
  const openScriptModal = useCallback(() => {
    setScriptModalVisible(true);
  }, []);

  const closeScriptModal = useCallback(() => {
    setScriptModalVisible(false);
  }, []);

  // 更新模态框操作
  const openUpdateModal = useCallback((client: Client) => {
    setUpdatingClient(client);
    setUpdateType('frpc');
    setUpdateMirrorId(undefined);
    setUpdateModalVisible(true);
  }, []);

  const closeUpdateModal = useCallback(() => {
    setUpdateModalVisible(false);
    setUpdatingClient(null);
  }, []);

  // 批量更新操作
  const openBatchUpdateModal = useCallback(() => {
    setBatchUpdateModalVisible(true);
    setUpdateType('frpc');
    setUpdateMirrorId(undefined);
  }, []);

  const closeBatchUpdateModal = useCallback(() => {
    setBatchUpdateModalVisible(false);
  }, []);

  // 删除确认操作
  const openDeleteConfirm = useCallback((client: Client) => {
    setDeletingClient(client);
    setDeleteConfirmVisible(true);
  }, []);

  const closeDeleteConfirm = useCallback(() => {
    setDeleteConfirmVisible(false);
    setDeletingClient(null);
  }, []);

  // 配置查看操作
  const openConfigView = useCallback((client: Client) => {
    setViewingClient(client);
    setConfigContent('');
    setConfigLoading(true);
    setConfigViewVisible(true);
  }, []);

  const closeConfigView = useCallback(() => {
    setConfigViewVisible(false);
    setViewingClient(null);
    setConfigContent('');
  }, []);

  // 日志查看操作
  const openLogView = useCallback((client: Client) => {
    setLogViewClient(client);
    setLogViewVisible(true);
  }, []);

  const closeLogView = useCallback(() => {
    setLogViewVisible(false);
    setLogViewClient(null);
  }, []);

  return {
    // 状态
    modalVisible,
    editingClient,
    scriptModalVisible,
    updateModalVisible,
    updatingClient,
    batchUpdateModalVisible,
    deleteConfirmVisible,
    deletingClient,
    configViewVisible,
    viewingClient,
    configContent,
    configLoading,
    logViewVisible,
    logViewClient,
    updateType,
    updateMirrorId,
    // 操作
    openAddModal,
    openEditModal,
    closeFormModal,
    openScriptModal,
    closeScriptModal,
    openUpdateModal,
    closeUpdateModal,
    openBatchUpdateModal,
    closeBatchUpdateModal,
    openDeleteConfirm,
    closeDeleteConfirm,
    openConfigView,
    closeConfigView,
    setConfigContent,
    setConfigLoading,
    openLogView,
    closeLogView,
    setUpdateType,
    setUpdateMirrorId,
  };
}