import request from './request';

export interface Setting {
  id: number;
  key: string;
  value: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export const settingApi = {
  getSettings: () => request.get<Setting[]>('/settings'),
  updateSetting: (data: { key: string; value: string }) => request.put('/settings', data),
  testEmail: (to: string) => request.post('/settings/test-email', { to })
};