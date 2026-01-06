import request from './request';

export interface GithubMirror {
  id: number;
  name: string;
  base_url: string;
  is_default: boolean;
  enabled: boolean;
  description: string;
  created_at: string;
  updated_at: string;
}

export const githubMirrorApi = {
  getAll: () => request.get<GithubMirror[]>('/github-mirrors'),
  getById: (id: number) => request.get<GithubMirror>(`/github-mirrors/${id}`),
  create: (data: Partial<GithubMirror>) => request.post<GithubMirror>('/github-mirrors', data),
  update: (id: number, data: Partial<GithubMirror>) => request.put<GithubMirror>(`/github-mirrors/${id}`, data),
  delete: (id: number) => request.delete(`/github-mirrors/${id}`),
  setDefault: (id: number) => request.post(`/github-mirrors/${id}/set-default`),
};