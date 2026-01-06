import { Modal } from '../../../../components/ui/Modal';
import { Button } from '../../../../components/ui/Button';
import type { Proxy, Client } from '../../../../types';
import { useProxyForm } from './useProxyForm';
import { BasicInfoSection } from './BasicInfoSection';
import { DomainConfigSection } from './DomainConfigSection';
import { AdvancedSection } from './AdvancedSection';

interface ProxyFormModalProps {
  visible: boolean;
  editingProxy: Proxy | null;
  clients: Client[];
  selectedClient: number | undefined;
  onlineClientIds: Set<number>;
  onCancel: () => void;
  onSubmit: (values: Partial<Proxy>) => Promise<void>;
}

export function ProxyFormModal({
  visible,
  editingProxy,
  clients,
  selectedClient,
  onlineClientIds,
  onCancel,
  onSubmit,
}: ProxyFormModalProps) {
  const form = useProxyForm({ visible, editingProxy, selectedClient, clients });

  const handleSubmit = async () => {
    if (!form.validateForm()) return;
    form.setSubmitting(true);
    try {
      await onSubmit(form.buildSubmitValues());
    } finally {
      form.setSubmitting(false);
    }
  };

  return (
    <Modal
      open={visible}
      onClose={onCancel}
      title={editingProxy ? '编辑代理' : '新增代理'}
      size="lg"
    >
      <div className="space-y-4 max-h-[60vh] overflow-y-auto pr-2">
        <BasicInfoSection
          formData={form.formData}
          editingProxy={!!editingProxy}
          clients={clients}
          onlineClientIds={onlineClientIds}
          pluginEnabled={form.pluginEnabled}
          currentPluginType={form.currentPluginType}
          pluginConfig={form.pluginConfig}
          errors={form.errors}
          updateField={form.updateField}
          handleTypeChange={form.handleTypeChange}
          handlePluginEnabledChange={form.handlePluginEnabledChange}
          handlePluginTypeChange={form.handlePluginTypeChange}
          handlePluginConfigChange={form.handlePluginConfigChange}
        />

        <DomainConfigSection
          formData={form.formData}
          currentProxyType={form.currentProxyType}
          certificates={form.certificates}
          loadingCertificates={form.loadingCertificates}
          dnsProviders={form.dnsProviders}
          providerDomains={form.providerDomains}
          loadingDomains={form.loadingDomains}
          errors={form.errors}
          shouldShowField={form.shouldShowField}
          getMatchingCertificates={form.getMatchingCertificates}
          updateField={form.updateField}
          handleCertificateChange={form.handleCertificateChange}
          handleDnsSyncChange={form.handleDnsSyncChange}
          handleDnsProviderChange={form.handleDnsProviderChange}
          handleRootDomainChange={form.handleRootDomainChange}
        />

        <AdvancedSection
          formData={form.formData}
          bandwidthEnabled={form.bandwidthEnabled}
          bandwidthValue={form.bandwidthValue}
          bandwidthUnit={form.bandwidthUnit}
          errors={form.errors}
          shouldShowField={form.shouldShowField}
          updateField={form.updateField}
          setBandwidthEnabled={form.setBandwidthEnabled}
          setBandwidthValue={form.setBandwidthValue}
          setBandwidthUnit={form.setBandwidthUnit}
        />
      </div>

      <div className="flex justify-end gap-3 mt-6 pt-4 border-t border-border">
        <Button variant="secondary" onClick={onCancel}>
          取消
        </Button>
        <Button onClick={handleSubmit} loading={form.submitting}>
          {editingProxy ? '保存' : '创建'}
        </Button>
      </div>
    </Modal>
  );
}