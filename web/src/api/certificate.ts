import request from './request';

export interface Certificate {
  id: number;
  proxy_id: number;
  domain: string;
  provider_id: number;
  status: 'pending' | 'active' | 'expiring' | 'expired' | 'failed';
  cert_pem: string;
  issuer_cert_pem: string;
  not_before: string | null;
  not_after: string | null;
  last_error: string;
  auto_renew: boolean;
  acme_account_id: string;
  created_at: string;
  updated_at: string;
}

export interface RequestCertificateInput {
  proxy_id: number;
  domain: string;
  dns_provider_id: number;
  auto_renew?: boolean;
}

export interface CertificateDownload {
  domain: string;
  cert_pem: string;
  issuer_cert_pem: string;
  full_chain_pem: string;
}

export const certificateApi = {
  list: () => request.get<Certificate[]>('/certificates'),
  getById: (id: number) => request.get<Certificate>(`/certificates/${id}`),
  request: (data: RequestCertificateInput) => request.post<Certificate>('/certificates', data),
  renew: (id: number) => request.post(`/certificates/${id}/renew`),
  reapply: (id: number) => request.post(`/certificates/${id}/reapply`),
  updateAutoRenew: (id: number, autoRenew: boolean) => request.put(`/certificates/${id}/auto-renew`, { auto_renew: autoRenew }),
  download: (id: number) => request.get<CertificateDownload>(`/certificates/${id}/download`),
  delete: (id: number) => request.delete(`/certificates/${id}`),
  getByDomain: (domain: string) => request.get<Certificate[]>('/certificates/by-domain', { params: { domain } }),
  getExpiring: () => request.get<Certificate[]>('/certificates/expiring'),
  getActive: () => request.get<Certificate[]>('/certificates/active'),
  getMatching: (domain: string) => request.get<Certificate[]>('/certificates/match', { params: { domain } }),
};