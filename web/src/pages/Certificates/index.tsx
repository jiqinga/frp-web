import { useState } from 'react';
import { Shield, Plus, AlertTriangle } from 'lucide-react';
import { Button, Card, CardHeader, CardContent } from '../../components/ui';
import { ConfirmModal } from '../../components/ui/ConfirmModal';
import { useCertificates } from './hooks';
import { CertificateStatsCard, CertificateTable, CertificateFormModal, CertificateProgressModal } from './components';

function Certificates() {
  const {
    certificates, providers, loading, stats,
    isAcmeConfigured, acmeEmailLoading, certProgress,
    requestCertificate, renewCertificate, reapplyCertificate, toggleAutoRenew, deleteCertificate, downloadCertificate, clearProgress,
  } = useCertificates();

  const [modalVisible, setModalVisible] = useState(false);
  const [deleteConfirmVisible, setDeleteConfirmVisible] = useState(false);
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [renewingId, setRenewingId] = useState<number | null>(null);
  const [reapplyingId, setReapplyingId] = useState<number | null>(null);
  const [togglingAutoRenewId, setTogglingAutoRenewId] = useState<number | null>(null);

  const handleAdd = () => setModalVisible(true);

  const handleRenew = async (id: number) => {
    setRenewingId(id);
    await renewCertificate(id);
    setRenewingId(null);
  };

  const handleReapply = async (id: number) => {
    setReapplyingId(id);
    await reapplyCertificate(id);
    setReapplyingId(null);
  };

  const handleToggleAutoRenew = async (id: number, autoRenew: boolean) => {
    setTogglingAutoRenewId(id);
    await toggleAutoRenew(id, autoRenew);
    setTogglingAutoRenewId(null);
  };

  const handleDelete = (id: number) => {
    setDeletingId(id);
    setDeleteConfirmVisible(true);
  };

  const confirmDelete = async () => {
    if (deletingId !== null) {
      await deleteCertificate(deletingId);
    }
    setDeleteConfirmVisible(false);
    setDeletingId(null);
  };

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">证书管理</h1>
          <p className="mt-1 text-foreground-muted">管理SSL/TLS证书，支持自动申请和续期</p>
        </div>
        <Button
          onClick={handleAdd}
          icon={<Plus className="h-4 w-4" />}
          disabled={!isAcmeConfigured || acmeEmailLoading}
          title={!isAcmeConfigured ? '请先在系统设置中配置ACME证书申请邮箱' : undefined}
        >
          申请证书
        </Button>
      </div>

      {!isAcmeConfigured && !acmeEmailLoading && (
        <div className="flex items-start gap-3 rounded-lg border p-4 border-yellow-500/30 bg-yellow-500/10">
          <AlertTriangle className="h-5 w-5 flex-shrink-0 text-yellow-500" />
          <div className="text-sm">
            <p className="font-medium text-yellow-700 dark:text-yellow-200">未配置ACME邮箱</p>
            <p className="mt-1 text-yellow-600 dark:text-yellow-200/80">
              请先在「系统设置 → DNS管理」中配置证书申请邮箱，才能申请SSL证书。
            </p>
          </div>
        </div>
      )}

      <CertificateStatsCard stats={stats} />

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Shield className="h-5 w-5 text-green-400" />
            <span>证书列表</span>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          <CertificateTable
            certificates={certificates}
            providers={providers}
            loading={loading}
            renewingId={renewingId}
            reapplyingId={reapplyingId}
            togglingAutoRenewId={togglingAutoRenewId}
            onRenew={handleRenew}
            onReapply={handleReapply}
            onToggleAutoRenew={handleToggleAutoRenew}
            onDelete={handleDelete}
            onDownload={downloadCertificate}
          />
        </CardContent>
      </Card>

      <CertificateFormModal
        visible={modalVisible}
        providers={providers}
        onCancel={() => setModalVisible(false)}
        onSubmit={async (data) => {
          setModalVisible(false);
          return requestCertificate(data);
        }}
      />

      <CertificateProgressModal
        visible={!!certProgress}
        progress={certProgress}
        onClose={clearProgress}
      />

      <ConfirmModal
        open={deleteConfirmVisible}
        onClose={() => setDeleteConfirmVisible(false)}
        onConfirm={confirmDelete}
        title="删除证书"
        content="确定删除此证书吗？删除后无法恢复。"
        type="warning"
        confirmText="删除"
        cancelText="取消"
      />
    </div>
  );
}

export function Component() {
  return <Certificates />;
}

export default Certificates;