export interface Category {
  id: number
  parent_id: number
  name: string
  code?: string
  sort: number
  creator_id?: number
  created_at?: string
  updated_at?: string
}

import request from '@/api'
import type { Question } from '@/types/entity'

export const questionApi = {
  // 题目管理
  list: (params: any) => request({ url: '/questions', method: 'GET', params }),
  detail: (id: number) => request<Question>({ url: `/questions/${id}`, method: 'GET' }),
  create: (data: any) => request({ url: '/questions', method: 'POST', data }),
  update: (id: number, data: any) => request({ url: `/questions/${id}`, method: 'PUT', data }),
  remove: (id: number) => request({ url: `/questions/${id}`, method: 'DELETE' }),

  // 分类管理（CRUD）
  listCategories: () => request<Category[]>({ url: '/question-categories', method: 'GET' }),
  categories: () => request<Category[]>({ url: '/question-categories', method: 'GET' }),
  createCategory: (data: any) => request({ url: '/question-categories', method: 'POST', data }),
  updateCategory: (id: number, data: any) => request({ url: `/question-categories/${id}`, method: 'PUT', data }),
  deleteCategory: (id: number) => request({ url: `/question-categories/${id}`, method: 'DELETE' }),

  // 批量导入
  batchImport: (data: any) => request<{ imported: number }>({ url: '/questions/import', method: 'POST', data }),
}
