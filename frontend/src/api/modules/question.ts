import request from '@/api'
import type { Question } from '@/types/entity'

export const questionApi = {
  list: (params: any) => request({ url: '/questions', method: 'GET', params }),
  detail: (id: number) => request<Question>({ url: `/questions/${id}`, method: 'GET' }),
  create: (data: any) => request({ url: '/questions', method: 'POST', data }),
  update: (id: number, data: any) => request({ url: `/questions/${id}`, method: 'PUT', data }),
  remove: (id: number) => request({ url: `/questions/${id}`, method: 'DELETE' }),
  categories: () => request({ url: '/question-categories', method: 'GET' }),
  createCategory: (data: any) => request({ url: '/question-categories', method: 'POST', data }),
  batchImport: (data: any) => request<{ imported: number }>({ url: '/questions/import', method: 'POST', data }),
}
