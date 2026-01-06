import axios, { type AxiosRequestConfig } from 'axios';
import { handleError, ErrorType } from '../utils/errorHandler';
import { toast } from '../components/ui/Toast';

const instance = axios.create({
  baseURL: '/api',
  timeout: 10000,
});

instance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

instance.interceptors.response.use(
  (response) => {
    // 对于非JSON响应(如text、blob),直接返回原始数据
    if (response.config.responseType === 'text' || response.config.responseType === 'blob') {
      return Promise.resolve(response.data);
    }
    
    const { code, message: msg, data } = response.data;
    if (code !== 0) {
      toast.error(msg || '请求失败');
      return Promise.reject(new Error(msg));
    }
    return Promise.resolve(data);
  },
  (error) => {
    const errorInfo = handleError(error);
    
    if (errorInfo.type === ErrorType.UNAUTHORIZED) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    
    toast.error(errorInfo.message);
    return Promise.reject(error);
  }
);

const request = {
  get: <T = unknown>(url: string, config?: AxiosRequestConfig) =>
    instance.get(url, config).then(res => res as unknown as T),
  post: <T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
    instance.post(url, data, config).then(res => res as unknown as T),
  put: <T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
    instance.put(url, data, config).then(res => res as unknown as T),
  delete: <T = unknown>(url: string, config?: AxiosRequestConfig) =>
    instance.delete(url, config).then(res => res as unknown as T),
};

export default request;