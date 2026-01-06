export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data: T;
}

export interface PaginationResponse<T> {
  list: T[];
  total: number;
}