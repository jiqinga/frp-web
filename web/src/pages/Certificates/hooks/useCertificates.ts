import { useState, useEffect, useCallback, useRef } from 'react';
import JSZip from 'jszip';
import { certificateApi, type Certificate } from '../../../api/certificate';
import { dnsApi } from '../../../api/dns';
import { settingApi } from '../../../api/setting';
import { toast } from '../../../components/ui/Toast';
import type { DNSProvider } from '../../../types';
import type { CertProgressState, CertProgressStep } from '../components/CertificateProgressModal';

export interface CertificateStats {
  total: number;
  active: number;
  expiring: number;
  expired: number;
  pending: number;
  failed: number;
}

export function useCertificates() {
  const [certificates, setCertificates] = useState<Certificate[]>([]);
  const [providers, setProviders] = useState<DNSProvider[]>([]);
  const [loading, setLoading] = useState(false);
  const [acmeEmail, setAcmeEmail] = useState('');
  const [acmeEmailLoading, setAcmeEmailLoading] = useState(true);
  const [certProgress, setCertProgress] = useState<CertProgressState | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  const fetchCertificates = useCallback(async () => {
    setLoading(true);
    try {
      const data = await certificateApi.list();
      setCertificates(data || []);
    } catch {
      toast.error('获取证书列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchProviders = useCallback(async () => {
    try {
      const data = await dnsApi.getProviders();
      setProviders(data.filter(p => p.enabled));
    } catch {
      // ignore
    }
  }, []);

  const fetchAcmeEmail = useCallback(async () => {
    setAcmeEmailLoading(true);
    try {
      const settings = await settingApi.getSettings();
      const acmeSetting = settings.find(s => s.key === 'acme_email');
      setAcmeEmail(acmeSetting?.value || '');
    } catch {
      // ignore
    } finally {
      setAcmeEmailLoading(false);
    }
  }, []);

  // WebSocket 连接
  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/ws/realtime?token=${token}`;
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type === 'cert_progress') {
          setCertProgress({
            taskId: data.task_id,
            domain: data.domain,
            step: data.step as CertProgressStep,
            message: data.message,
            error: data.error,
          });
          if (data.step === 'completed' || data.step === 'failed') {
            fetchCertificates();
          }
        }
      } catch {
        // ignore parse errors
      }
    };

    return () => {
      ws.close();
    };
  }, [fetchCertificates]);

  useEffect(() => {
    fetchCertificates();
    fetchProviders();
    fetchAcmeEmail();
  }, [fetchCertificates, fetchProviders, fetchAcmeEmail]);

  const requestCertificate = async (data: { proxy_id: number; domain: string; dns_provider_id: number }) => {
    try {
      // 初始化进度状态
      setCertProgress({
        taskId: '',
        domain: data.domain,
        step: 'validating',
        message: '正在验证配置...',
      });
      await certificateApi.request(data);
      return true;
    } catch {
      setCertProgress(prev => prev ? { ...prev, step: 'failed', error: '证书申请失败' } : null);
      return false;
    }
  };

  const clearProgress = useCallback(() => {
    setCertProgress(null);
  }, []);

  const renewCertificate = async (id: number) => {
    try {
      await certificateApi.renew(id);
      toast.success('证书续期成功');
      fetchCertificates();
      return true;
    } catch {
      return false;
    }
  };

  const reapplyCertificate = async (id: number) => {
    try {
      await certificateApi.reapply(id);
      toast.success('已开始重新申请证书');
      fetchCertificates();
      return true;
    } catch {
      toast.error('重新申请证书失败');
      return false;
    }
  };

  const toggleAutoRenew = async (id: number, autoRenew: boolean) => {
    try {
      await certificateApi.updateAutoRenew(id, autoRenew);
      toast.success(autoRenew ? '已开启自动续期' : '已关闭自动续期');
      fetchCertificates();
      return true;
    } catch {
      toast.error('更新自动续期状态失败');
      return false;
    }
  };

  const deleteCertificate = async (id: number) => {
    try {
      await certificateApi.delete(id);
      toast.success('证书删除成功');
      fetchCertificates();
      return true;
    } catch {
      return false;
    }
  };

  const downloadCertificate = async (id: number) => {
    try {
      const data = await certificateApi.download(id);
      const zip = new JSZip();
      
      // 添加证书文件到 zip
      zip.file(`${data.domain}.crt`, data.cert_pem);
      
      // 添加完整证书链到 zip
      if (data.full_chain_pem) {
        zip.file(`${data.domain}.fullchain.crt`, data.full_chain_pem);
      }
      
      // 生成并下载 zip 文件
      const blob = await zip.generateAsync({ type: 'blob' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${data.domain}-certificates.zip`;
      a.click();
      URL.revokeObjectURL(url);

      toast.success('证书下载成功');
      return true;
    } catch {
      toast.error('下载证书失败');
      return false;
    }
  };

  const stats: CertificateStats = {
    total: certificates.length,
    active: certificates.filter(c => c.status === 'active').length,
    expiring: certificates.filter(c => c.status === 'expiring').length,
    expired: certificates.filter(c => c.status === 'expired').length,
    pending: certificates.filter(c => c.status === 'pending').length,
    failed: certificates.filter(c => c.status === 'failed').length,
  };

  return {
    certificates,
    providers,
    loading,
    stats,
    acmeEmail,
    acmeEmailLoading,
    isAcmeConfigured: !!acmeEmail,
    certProgress,
    fetchCertificates,
    requestCertificate,
    renewCertificate,
    reapplyCertificate,
    toggleAutoRenew,
    deleteCertificate,
    downloadCertificate,
    clearProgress,
  };
}