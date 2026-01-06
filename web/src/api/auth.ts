import request from './request';
import type { User } from '../types';

export const authApi = {
  login: (data: { username: string; password: string }) =>
    request.post<{ token: string; user: User }>('/auth/login', data),

  getProfile: () => request.get<User>('/auth/profile'),

  changePassword: (data: { old_password: string; new_password: string }) =>
    request.put<{ message: string }>('/auth/password', data),
};