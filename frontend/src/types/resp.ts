// 统一后端响应
export interface BaseResp<T = unknown> {
  code: number
  message: string
  data?: T
  trace_id?: string
}

export interface PageData<T = unknown> {
  list: T[]
  total: number
  page: number
  page_size: number
}
