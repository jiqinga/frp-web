import { useState, useEffect } from 'react';
import { FileText, User, Clock, MapPin, Filter } from 'lucide-react';
import { logApi, type OperationLog } from '../../api/log';
import { Table, Card, CardHeader, CardContent, Badge, Select, Pagination } from '../../components/ui';

// 格式化日期
const formatDate = (dateStr: string) => {
  const date = new Date(dateStr);
  const pad = (n: number) => n.toString().padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
};

// 操作类型选项
const operationOptions = [
  { value: '', label: '全部操作' },
  { value: 'login', label: '登录' },
  { value: 'login_failed', label: '登录失败' },
  { value: 'create', label: '创建' },
  { value: 'update', label: '更新' },
  { value: 'delete', label: '删除' },
  { value: 'start', label: '启动' },
  { value: 'stop', label: '停止' },
  { value: 'restart', label: '重启' },
  { value: 'download', label: '下载' },
  { value: 'remote_install', label: '远程安装' },
  { value: 'remote_start', label: '远程启动' },
  { value: 'remote_stop', label: '远程停止' },
  { value: 'remote_restart', label: '远程重启' },
  { value: 'remote_uninstall', label: '远程卸载' },
  { value: 'remote_reinstall', label: '远程重装' },
  { value: 'remote_upgrade', label: '远程升级' },
  { value: 'generate_token', label: '生成令牌' },
  { value: 'update_software', label: '更新软件' },
  { value: 'batch_update_software', label: '批量更新' },
  { value: 'change_password', label: '修改密码' },
  { value: 'set_default', label: '设为默认' },
];

// 资源类型选项
const resourceOptions = [
  { value: '', label: '全部资源' },
  { value: 'user', label: '用户' },
  { value: 'client', label: '客户端' },
  { value: 'proxy', label: '代理' },
  { value: 'frps', label: 'FRP服务器' },
  { value: 'client_register_token', label: '注册令牌' },
  { value: 'setting', label: '系统设置' },
  { value: 'alert_rule', label: '告警规则' },
  { value: 'github_mirror', label: 'GitHub镜像' },
];

// 每页条数选项
const pageSizeOptions = [
  { value: 10, label: '10 条/页' },
  { value: 20, label: '20 条/页' },
  { value: 50, label: '50 条/页' },
  { value: 100, label: '100 条/页' },
];

export function Component() {
  const [logs, setLogs] = useState<OperationLog[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [filters, setFilters] = useState({ page: 1, page_size: 10, operation_type: '', resource_type: '' });

  useEffect(() => {
    fetchLogs();
  }, [filters]);

  const fetchLogs = async () => {
    setLoading(true);
    try {
      const res = await logApi.getLogs(filters);
      setLogs(res.list);
      setTotal(res.total);
    } catch {
      // 静默处理，日志加载失败时保持当前状态
    } finally {
      setLoading(false);
    }
  };

  const getOperationVariant = (type: string): 'success' | 'primary' | 'danger' | 'warning' | 'info' | 'default' => {
    const variantMap: Record<string, 'success' | 'primary' | 'danger' | 'warning' | 'info' | 'default'> = {
      create: 'success',
      update: 'primary',
      delete: 'danger',
      start: 'info',
      stop: 'warning',
      restart: 'primary',
      login: 'info',
      login_failed: 'danger',
      download: 'info',
      remote_install: 'success',
      remote_start: 'info',
      remote_stop: 'warning',
      remote_restart: 'primary',
      remote_uninstall: 'danger',
      remote_reinstall: 'primary',
      remote_upgrade: 'success',
      generate_token: 'success',
      update_software: 'primary',
      batch_update_software: 'primary',
      change_password: 'warning',
      set_default: 'info'
    };
    return variantMap[type] || 'default';
  };

  const getOperationLabel = (type: string) => {
    const labelMap: Record<string, string> = {
      create: '创建',
      update: '更新',
      delete: '删除',
      start: '启动',
      stop: '停止',
      restart: '重启',
      login: '登录',
      login_failed: '登录失败',
      download: '下载',
      remote_install: '远程安装',
      remote_start: '远程启动',
      remote_stop: '远程停止',
      remote_restart: '远程重启',
      remote_uninstall: '远程卸载',
      remote_reinstall: '远程重装',
      remote_upgrade: '远程升级',
      generate_token: '生成令牌',
      update_software: '更新软件',
      batch_update_software: '批量更新',
      change_password: '修改密码',
      set_default: '设为默认'
    };
    return labelMap[type] || type;
  };

  const getResourceLabel = (type: string) => {
    const labelMap: Record<string, string> = {
      client: '客户端',
      proxy: '代理',
      frps: 'FRP服务器',
      user: '用户',
      client_register_token: '注册令牌',
      setting: '系统设置',
      alert_rule: '告警规则',
      github_mirror: 'GitHub镜像'
    };
    return labelMap[type] || type;
  };

  const columns = [
    {
      key: 'id',
      title: '编号',
      render: (_: unknown, record: OperationLog) => (
        <span className="font-mono text-sm text-foreground-muted">{record.id}</span>
      )
    },
    { 
      key: 'username', 
      title: '操作用户',
      render: (_: unknown, record: OperationLog) => (
        <div className="flex items-center gap-2">
          <div className="flex h-7 w-7 items-center justify-center rounded-full bg-indigo-500/20">
            <User className="h-3.5 w-3.5 text-indigo-400" />
          </div>
          <span className="font-medium text-foreground">{record.username || '-'}</span>
        </div>
      )
    },
    { 
      key: 'operation_type', 
      title: '操作类型',
      render: (_: unknown, record: OperationLog) => (
        <Badge variant={getOperationVariant(record.operation_type)}>
          {getOperationLabel(record.operation_type)}
        </Badge>
      )
    },
    { 
      key: 'resource_type', 
      title: '资源类型',
      render: (_: unknown, record: OperationLog) => (
        <Badge variant="info">{getResourceLabel(record.resource_type)}</Badge>
      )
    },
    { 
      key: 'description', 
      title: '描述',
      render: (_: unknown, record: OperationLog) => (
        <span className="line-clamp-1 text-foreground-secondary">{record.description}</span>
      )
    },
    {
      key: 'ip_address',
      title: 'IP地址',
      render: (_: unknown, record: OperationLog) => (
        <span className="font-mono text-sm text-foreground-secondary">{record.ip_address}</span>
      )
    },
    {
      key: 'ip_location',
      title: 'IP归属地',
      render: (_: unknown, record: OperationLog) => (
        <div className="flex items-center gap-1 text-foreground-muted">
          <MapPin className="h-3.5 w-3.5" />
          <span className="text-sm">{record.ip_location || '-'}</span>
        </div>
      )
    },
    { 
      key: 'created_at', 
      title: '时间',
      render: (_: unknown, record: OperationLog) => (
        <div className="flex items-center gap-1 text-cyan-400">
            <Clock className="h-3.5 w-3.5" />
            <span className="text-sm">{formatDate(record.created_at)}</span>
          </div>
      )
    }
  ];

  return (
    <div className="space-y-6 p-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">操作日志</h1>
          <p className="mt-1 text-foreground-muted">查看系统操作记录</p>
        </div>
      </div>

      {/* 日志卡片 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <FileText className="h-5 w-5 text-indigo-400" />
              <span>操作日志</span>
              <Badge variant="default">{total} 条记录</Badge>
            </div>
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-2">
                <Filter className="h-4 w-4 text-foreground-muted" />
                <Select
                  value={filters.operation_type}
                  onChange={(value) => setFilters({ ...filters, page: 1, operation_type: String(value) })}
                  options={operationOptions}
                  className="w-32"
                />
                <Select
                  value={filters.resource_type}
                  onChange={(value) => setFilters({ ...filters, page: 1, resource_type: String(value) })}
                  options={resourceOptions}
                  className="w-32"
                />
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          <Table
            columns={columns}
            data={logs}
            rowKey="id"
            loading={loading}
            emptyText="暂无日志记录"
          />
        </CardContent>
      </Card>

      {/* 分页 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Select
            value={filters.page_size}
            onChange={(value) => setFilters({ ...filters, page: 1, page_size: Number(value) })}
            options={pageSizeOptions}
            className="w-28"
          />
        </div>
        <Pagination
          current={filters.page}
          total={total}
          pageSize={filters.page_size}
          onChange={(page) => setFilters({ ...filters, page })}
          showTotal
        />
      </div>
    </div>
  );
}