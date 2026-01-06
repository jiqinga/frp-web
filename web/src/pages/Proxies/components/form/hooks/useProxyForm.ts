import { useState, useEffect, useCallback } from 'react';
import type { Proxy, Client, ProxyType, PluginType, DNSProvider } from '../../../../../types';
import type { Certificate } from '../../../../../api/certificate';
import { certificateApi } from '../../../../../api/certificate';
import { dnsApi } from '../../../../../api/dns';
import {
  parseBandwidthLimit,
  formatBandwidthLimit,
  PROXY_TYPE_FIELDS,
  PROXY_REQUIRED_FIELDS,
} from '../../../constants';
import type { FormData } from '../types';
import { initialFormData } from '../types';
import {
  extractPrefixesFromDomains,
  generateDomainsFromPrefixes,
  validateDomainPrefix,
  generateCertPaths,
  parseCertificateDomain,
} from '../utils';

interface UseProxyFormProps {
  visible: boolean;
  editingProxy: Proxy | null;
  clients: Client[];
  selectedClient: number | undefined;
  onSubmit: (values: Partial<Proxy>) => Promise<void>;
}

export function useProxyForm({
  visible,
  editingProxy,
  clients,
  selectedClient,
  onSubmit,
}: UseProxyFormProps) {
  // 表单数据状态
  const [formData, setFormData] = useState<FormData>(initialFormData);
  const [currentProxyType, setCurrentProxyType] = useState<ProxyType>('tcp');
  
  // 带宽限制状态
  const [bandwidthEnabled, setBandwidthEnabled] = useState(false);
  const [bandwidthValue, setBandwidthValue] = useState<number | undefined>(undefined);
  const [bandwidthUnit, setBandwidthUnit] = useState('MB');

  // 插件配置状态
  const [pluginEnabled, setPluginEnabled] = useState(false);
  const [currentPluginType, setCurrentPluginType] = useState<PluginType>('http_proxy');
  const [pluginConfig, setPluginConfig] = useState<Record<string, string>>({});

  // 表单验证错误
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  // DNS 提供商和域名状态
  const [dnsProviders, setDnsProviders] = useState<DNSProvider[]>([]);
  const [providerDomains, setProviderDomains] = useState<string[]>([]);
  const [loadingDomains, setLoadingDomains] = useState(false);

  // 证书状态
  const [certificates, setCertificates] = useState<Certificate[]>([]);
  const [loadingCertificates, setLoadingCertificates] = useState(false);

  // 判断字段是否应该显示
  const shouldShowField = useCallback((fieldName: string): boolean => {
    const fields = PROXY_TYPE_FIELDS[currentProxyType];
    return fields.includes(fieldName);
  }, [currentProxyType]);

  // 判断字段是否必填
  const isFieldRequired = useCallback((fieldName: string): boolean => {
    const requiredFields = PROXY_REQUIRED_FIELDS[currentProxyType];
    return requiredFields.includes(fieldName);
  }, [currentProxyType]);

  // 更新表单字段
  const updateField = useCallback((field: keyof FormData, value: string | number | boolean | undefined) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[field];
        return newErrors;
      });
    }
  }, [errors]);

  // 处理代理类型变化
  const handleTypeChange = useCallback((value: string | number) => {
    const newType = value as ProxyType;
    setCurrentProxyType(newType);
    
    if (newType === 'https') {
      setFormData(prev => ({ ...prev, type: newType }));
      setPluginEnabled(true);
      setCurrentPluginType('https2http');
      setPluginConfig({});
    } else {
      setFormData(prev => ({ ...prev, type: newType }));
      if (pluginEnabled && (currentPluginType === 'https2http' || currentPluginType === 'https2https')) {
        setPluginEnabled(false);
        setCurrentPluginType('http_proxy');
        setPluginConfig({});
      }
    }
  }, [pluginEnabled, currentPluginType]);

  // 处理插件类型变化
  const handlePluginTypeChange = useCallback((value: PluginType) => {
    setCurrentPluginType(value);
    setPluginConfig({});
  }, []);

  // 处理插件配置字段变化
  const handlePluginConfigChange = useCallback((field: string, value: string) => {
    setPluginConfig(prev => ({ ...prev, [field]: value }));
  }, []);

  // 处理启用插件开关变化
  const handlePluginEnabledChange = useCallback((checked: boolean) => {
    setPluginEnabled(checked);
    if (checked) {
      if (currentProxyType === 'https') {
        setCurrentPluginType('https2http');
      } else {
        setCurrentProxyType('tcp');
        setFormData(prev => ({ ...prev, type: 'tcp' }));
      }
    }
  }, [currentProxyType]);

  // 处理 DNS 提供商变化
  const handleDnsProviderChange = useCallback(async (providerId: number | undefined, preserveValues?: boolean) => {
    if (!preserveValues) {
      setFormData(prev => ({ ...prev, dns_provider_id: providerId, dns_root_domain: '', domain_prefixes: '' }));
      setProviderDomains([]);
    }
    
    if (providerId) {
      setLoadingDomains(true);
      try {
        const domains = await dnsApi.getProviderDomains(providerId);
        setProviderDomains(domains);
      } catch (error) {
        console.error('获取域名列表失败:', error);
      } finally {
        setLoadingDomains(false);
      }
    }
  }, []);

  // 处理根域名变化
  const handleRootDomainChange = useCallback((rootDomain: string) => {
    setFormData(prev => ({
      ...prev,
      dns_root_domain: rootDomain,
      domain_prefixes: prev.dns_root_domain !== rootDomain ? '' : prev.domain_prefixes,
    }));
  }, []);

  // 处理证书选择变化
  const handleCertificateChange = useCallback((certId: number | undefined) => {
    setFormData(prev => ({ ...prev, cert_id: certId }));
    
    if (certId) {
      const selectedCert = certificates.find(c => c.id === certId);
      if (selectedCert) {
        const { crtPath, keyPath } = generateCertPaths(selectedCert.domain);
        setPluginConfig(prev => ({ ...prev, crtPath, keyPath }));
      }
    }
  }, [certificates]);

  // 处理 DNS 同步开关变化
  const handleDnsSyncChange = useCallback(async (checked: boolean) => {
    if (checked && formData.cert_id) {
      const selectedCert = certificates.find(c => c.id === formData.cert_id);
      if (selectedCert) {
        const { rootDomain, prefix } = parseCertificateDomain(selectedCert.domain);
        const providerId = selectedCert.provider_id;
        
        setFormData(prev => ({
          ...prev,
          enable_dns_sync: true,
          dns_provider_id: providerId,
          dns_root_domain: rootDomain,
          domain_prefixes: prefix,
        }));
        
        if (providerId) {
          setLoadingDomains(true);
          try {
            const domains = await dnsApi.getProviderDomains(providerId);
            setProviderDomains(domains);
          } catch (error) {
            console.error('获取域名列表失败:', error);
          } finally {
            setLoadingDomains(false);
          }
        }
        return;
      }
    }
    
    setFormData(prev => ({
      ...prev,
      enable_dns_sync: checked,
      dns_provider_id: checked ? prev.dns_provider_id : undefined,
      dns_root_domain: checked ? prev.dns_root_domain : '',
      domain_prefixes: checked ? prev.domain_prefixes : '',
    }));
    if (!checked) {
      setProviderDomains([]);
    }
  }, [formData.cert_id, certificates]);

  // 加载 DNS 提供商列表
  useEffect(() => {
    const loadProviders = async () => {
      try {
        const providers = await dnsApi.getProviders();
        setDnsProviders(providers.filter(p => p.enabled));
      } catch (error) {
        console.error('获取DNS提供商列表失败:', error);
      }
    };
    if (visible) {
      loadProviders();
    }
  }, [visible]);

  // 加载证书列表
  useEffect(() => {
    const loadCertificates = async () => {
      if (currentProxyType !== 'https') {
        setCertificates([]);
        return;
      }
      setLoadingCertificates(true);
      try {
        const certs = await certificateApi.getActive();
        setCertificates(certs);
      } catch (error) {
        console.error('获取证书列表失败:', error);
      } finally {
        setLoadingCertificates(false);
      }
    };
    if (visible) {
      loadCertificates();
    }
  }, [visible, currentProxyType]);

  // 根据域名筛选匹配的证书
  const getMatchingCertificates = useCallback(() => {
    const domain = formData.enable_dns_sync && formData.dns_root_domain && formData.domain_prefixes
      ? generateDomainsFromPrefixes(formData.domain_prefixes, formData.dns_root_domain).split(',')[0]
      : formData.custom_domains?.split(',')[0]?.trim();
    
    if (!domain) return certificates;
    
    return certificates.filter(cert => {
      if (cert.domain === domain) return true;
      if (cert.domain.startsWith('*.')) {
        const wildcardBase = cert.domain.slice(2);
        const domainParts = domain.split('.');
        if (domainParts.length >= 2) {
          const domainBase = domainParts.slice(1).join('.');
          if (domainBase === wildcardBase) return true;
        }
      }
      return false;
    });
  }, [certificates, formData.custom_domains, formData.enable_dns_sync, formData.dns_root_domain, formData.domain_prefixes]);

  // 重置表单
  useEffect(() => {
    if (visible) {
      if (editingProxy) {
        setCurrentProxyType((editingProxy.type?.toLowerCase() || 'tcp') as ProxyType);
        
        const { value, unit } = parseBandwidthLimit(editingProxy.bandwidth_limit);
        setBandwidthEnabled(!!editingProxy.bandwidth_limit);
        setBandwidthValue(value);
        setBandwidthUnit(unit);
        
        if (editingProxy.plugin_type && editingProxy.plugin_config) {
          setPluginEnabled(true);
          setCurrentPluginType(editingProxy.plugin_type as PluginType);
          try {
            setPluginConfig(JSON.parse(editingProxy.plugin_config));
          } catch {
            setPluginConfig({});
          }
        } else {
          setPluginEnabled(false);
          setCurrentPluginType('http_proxy');
          setPluginConfig({});
        }
        
        const extractedPrefixes = extractPrefixesFromDomains(
          editingProxy.custom_domains || '',
          editingProxy.dns_root_domain || ''
        );
        
        setFormData({
          client_id: editingProxy.client_id,
          name: editingProxy.name || '',
          type: (editingProxy.type?.toLowerCase() || 'tcp') as ProxyType,
          local_ip: editingProxy.local_ip || '127.0.0.1',
          local_port: editingProxy.local_port,
          remote_port: editingProxy.remote_port,
          custom_domains: editingProxy.custom_domains || '',
          subdomain: editingProxy.subdomain || '',
          locations: editingProxy.locations || '',
          host_header_rewrite: editingProxy.host_header_rewrite || '',
          http_user: editingProxy.http_user || '',
          http_password: editingProxy.http_password || '',
          secret_key: editingProxy.secret_key || '',
          allow_users: editingProxy.allow_users || '',
          enable_dns_sync: editingProxy.enable_dns_sync || false,
          dns_provider_id: editingProxy.dns_provider_id,
          dns_root_domain: editingProxy.dns_root_domain || '',
          domain_prefixes: extractedPrefixes,
          cert_id: editingProxy.cert_id,
        });
        
        if (editingProxy.dns_provider_id) {
          handleDnsProviderChange(editingProxy.dns_provider_id, true);
        }
      } else {
        setCurrentProxyType('tcp');
        setBandwidthEnabled(false);
        setBandwidthValue(undefined);
        setBandwidthUnit('MB');
        setPluginEnabled(false);
        setCurrentPluginType('http_proxy');
        setPluginConfig({});
        
        const defaultClientId = selectedClient || (clients.length > 0 ? clients[0].id : undefined);
        setFormData({ ...initialFormData, client_id: defaultClientId });
      }
      setErrors({});
    }
  }, [visible, editingProxy, selectedClient, clients, handleDnsProviderChange]);

  // 验证表单
  const validateForm = useCallback((): boolean => {
    const newErrors: Record<string, string> = {};
    
    if (!formData.client_id) {
      newErrors.client_id = '请选择客户端';
    }
    
    if (!formData.name?.trim()) {
      newErrors.name = '请输入代理名称';
    }
    
    if (!pluginEnabled) {
      if (!formData.local_ip?.trim()) {
        newErrors.local_ip = '请输入本地IP';
      }
      if (!formData.local_port) {
        newErrors.local_port = '请输入本地端口';
      }
    }
    
    if (isFieldRequired('custom_domains')) {
      if (formData.enable_dns_sync && formData.dns_root_domain) {
        if (!formData.domain_prefixes?.trim()) {
          newErrors.domain_prefixes = '请输入域名前缀';
        } else {
          const prefixes = formData.domain_prefixes.split(',').map(p => p.trim()).filter(Boolean);
          const invalidPrefixes = prefixes.filter(p => !validateDomainPrefix(p));
          if (invalidPrefixes.length > 0) {
            newErrors.domain_prefixes = `无效的前缀: ${invalidPrefixes.join(', ')}`;
          }
        }
      } else if (!formData.custom_domains?.trim() && !formData.subdomain?.trim()) {
        newErrors.custom_domains = '请输入自定义域名或子域名';
      }
    }
    
    if (isFieldRequired('secret_key') && !formData.secret_key?.trim()) {
      newErrors.secret_key = '请输入密钥';
    }

    if (currentProxyType === 'https' && !formData.cert_id) {
      newErrors.cert_id = '请选择证书';
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }, [formData, pluginEnabled, isFieldRequired, currentProxyType]);

  // 提交表单
  const handleSubmit = useCallback(async () => {
    if (!validateForm()) return;
    
    setSubmitting(true);
    try {
      let customDomains = formData.custom_domains;
      if (formData.enable_dns_sync && formData.dns_root_domain && formData.domain_prefixes) {
        customDomains = generateDomainsFromPrefixes(formData.domain_prefixes, formData.dns_root_domain);
      }
      
      const submitValues: Partial<Proxy> = {
        ...formData,
        custom_domains: customDomains,
        type: formData.type,
        enable_dns_sync: formData.enable_dns_sync,
        dns_provider_id: formData.dns_provider_id,
        dns_root_domain: formData.dns_root_domain,
        cert_id: formData.cert_id,
      };
      
      if (bandwidthEnabled && bandwidthValue) {
        submitValues.bandwidth_limit = formatBandwidthLimit(bandwidthValue, bandwidthUnit);
      } else {
        submitValues.bandwidth_limit = undefined;
      }
      
      if (pluginEnabled) {
        submitValues.plugin_type = currentPluginType;
        submitValues.plugin_config = JSON.stringify(pluginConfig);
        if (!submitValues.local_ip) submitValues.local_ip = '127.0.0.1';
        if (!submitValues.local_port) submitValues.local_port = 0;
      } else {
        submitValues.plugin_type = '';
        submitValues.plugin_config = '';
      }
      
      await onSubmit(submitValues);
    } finally {
      setSubmitting(false);
    }
  }, [validateForm, formData, bandwidthEnabled, bandwidthValue, bandwidthUnit, pluginEnabled, currentPluginType, pluginConfig, onSubmit]);

  return {
    formData,
    currentProxyType,
    bandwidthEnabled,
    bandwidthValue,
    bandwidthUnit,
    pluginEnabled,
    currentPluginType,
    pluginConfig,
    errors,
    submitting,
    dnsProviders,
    providerDomains,
    loadingDomains,
    certificates,
    loadingCertificates,
    shouldShowField,
    isFieldRequired,
    updateField,
    handleTypeChange,
    handlePluginTypeChange,
    handlePluginConfigChange,
    handlePluginEnabledChange,
    handleDnsProviderChange,
    handleRootDomainChange,
    handleCertificateChange,
    handleDnsSyncChange,
    getMatchingCertificates,
    handleSubmit,
    setBandwidthEnabled,
    setBandwidthValue,
    setBandwidthUnit,
  };
}