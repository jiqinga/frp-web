import { AxiosError } from 'axios';

export const ErrorType = {
  NETWORK_ERROR: 'NETWORK_ERROR',
  TIMEOUT_ERROR: 'TIMEOUT_ERROR',
  SERVER_ERROR: 'SERVER_ERROR',
  UNAUTHORIZED: 'UNAUTHORIZED',
  BUSINESS_ERROR: 'BUSINESS_ERROR',
  UNKNOWN_ERROR: 'UNKNOWN_ERROR',
} as const;

export type ErrorType = typeof ErrorType[keyof typeof ErrorType];

export interface ErrorInfo {
  type: ErrorType;
  message: string;
  originalError?: unknown;
}

export function handleError(error: unknown): ErrorInfo {
  const axiosError = error as AxiosError;

  // 网络错误 - 后端未启动或无法连接
  if (axiosError.code === 'ERR_NETWORK' || !axiosError.response) {
    return {
      type: ErrorType.NETWORK_ERROR,
      message: '无法连接到后端服务，请确认后端是否已启动 (http://localhost:8080)',
      originalError: error,
    };
  }

  // 超时错误
  if (axiosError.code === 'ECONNABORTED' || axiosError.message?.includes('timeout')) {
    return {
      type: ErrorType.TIMEOUT_ERROR,
      message: '请求超时，请检查网络连接后重试',
      originalError: error,
    };
  }

  const status = axiosError.response?.status;
  const responseData = axiosError.response?.data as { message?: string };

  // 401 未授权
  if (status === 401) {
    return {
      type: ErrorType.UNAUTHORIZED,
      message: '登录已过期，请重新登录',
      originalError: error,
    };
  }

  // 500 服务器错误
  if (status === 500) {
    return {
      type: ErrorType.SERVER_ERROR,
      message: responseData?.message || '服务器内部错误，请稍后重试或联系管理员',
      originalError: error,
    };
  }

  // 业务错误 - 有具体错误信息
  if (responseData?.message) {
    return {
      type: ErrorType.BUSINESS_ERROR,
      message: responseData.message,
      originalError: error,
    };
  }

  // 未知错误
  return {
    type: ErrorType.UNKNOWN_ERROR,
    message: axiosError.message || '操作失败，请稍后重试',
    originalError: error,
  };
}