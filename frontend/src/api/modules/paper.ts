import request from '@/api'

export interface PaperConfigRule {
  rules: Array<{
    type: number       // 题型
    difficulty: number // 难度
    count: number      // 抽取数量
    score: number      // 每题分值
  }>
  total_score: number
}

export interface PaperDetail {
  paper: {
    id: number
    title: string
    description: string
    strategy: number  // 1固定 2随机 3GA
    total_score: number
    duration: number
    pass_score: number
    status: number
    creator_id: number
    config_rule: string  // JSON string
    question_ids: string // JSON string
    created_at: string
    updated_at: string
  }
  questions: Array<{
    id: number
    type: number
    title: string
    options: any[]
    answer: any
    analysis?: string
    score: number
    difficulty?: number
    category_id?: number
  }>
}

export const paperApi = {
  list: (params: any) => request({ url: '/papers', method: 'GET', params }),
  detail: (id: number) => request<PaperDetail>({ url: `/papers/${id}`, method: 'GET' }),
  create: (data: any) => request({ url: '/papers', method: 'POST', data }),
  update: (id: number, data: any) => request({ url: `/papers/${id}`, method: 'PUT', data }),
  remove: (id: number) => request({ url: `/papers/${id}`, method: 'DELETE' }),
}
