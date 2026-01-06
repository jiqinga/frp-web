import request from './request';

export interface AlertRecipient {
  id?: number;
  name: string;
  email: string;
  enabled: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface AlertRecipientGroup {
  id?: number;
  name: string;
  description: string;
  enabled: boolean;
  recipients?: AlertRecipient[];
  created_at?: string;
  updated_at?: string;
}

export const alertRecipientApi = {
  // 接收人
  getRecipients: () => request.get<AlertRecipient[]>('/alerts/recipients'),
  createRecipient: (data: AlertRecipient) => request.post('/alerts/recipients', data),
  updateRecipient: (id: number, data: AlertRecipient) => request.put(`/alerts/recipients/${id}`, data),
  deleteRecipient: (id: number) => request.delete(`/alerts/recipients/${id}`),
  
  // 分组
  getGroups: () => request.get<AlertRecipientGroup[]>('/alerts/groups'),
  createGroup: (data: AlertRecipientGroup) => request.post('/alerts/groups', data),
  updateGroup: (id: number, data: AlertRecipientGroup) => request.put(`/alerts/groups/${id}`, data),
  deleteGroup: (id: number) => request.delete(`/alerts/groups/${id}`),
  setGroupRecipients: (id: number, recipientIds: number[]) => 
    request.put(`/alerts/groups/${id}/recipients`, { recipient_ids: recipientIds }),
};