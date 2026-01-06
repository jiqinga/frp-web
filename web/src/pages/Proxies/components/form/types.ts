import type { Proxy, Client, ProxyType, PluginType, DNSProvider } from '../../../../types';
import type { Certificate } from '../../../../api/certificate';

export interface ProxyFormModalProps {
  visible: boolean;
  editingProxy: Proxy | null;
  clients: Client[];
  selectedClient: number | undefined;
  onlineClientIds: Set<number>;
  onCancel: () => void;
  onSubmit: (values: Partial<Proxy>) => Promise<void>;
}

export interface FormData {
  client_id?: number;
  name: string;
  type: ProxyType;
  local_ip: string;
  local_port: number | undefined;
  remote_port: number | undefined;
  custom_domains: string;
  subdomain: string;
  locations: string;
  host_header_rewrite: string;
  http_user: string;
  http_password: string;
  secret_key: string;
  allow_users: string;
  enable_dns_sync: boolean;
  dns_provider_id?: number;
  dns_root_domain: string;
  domain_prefixes: string;
  cert_id?: number;
}

export const initialFormData: FormData = {
  client_id: undefined,
  name: '',
  type: 'tcp',
  local_ip: '127.0.0.1',
  local_port: undefined,
  remote_port: undefined,
  custom_domains: '',
  subdomain: '',
  locations: '',
  host_header_rewrite: '',
  http_user: '',
  http_password: '',
  secret_key: '',
  allow_users: '',
  enable_dns_sync: false,
  dns_provider_id: undefined,
  dns_root_domain: '',
  domain_prefixes: '',
  cert_id: undefined,
};

export interface ProxyFormState {
  formData: FormData;
  currentProxyType: ProxyType;
  bandwidthEnabled: boolean;
  bandwidthValue: number | undefined;
  bandwidthUnit: string;
  pluginEnabled: boolean;
  currentPluginType: PluginType;
  pluginConfig: Record<string, string>;
  errors: Record<string, string>;
  submitting: boolean;
  dnsProviders: DNSProvider[];
  providerDomains: string[];
  loadingDomains: boolean;
  certificates: Certificate[];
  loadingCertificates: boolean;
}