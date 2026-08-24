import request from '@/api'
import type { Exam, ExamRecord, Question } from '@/types/entity'

export interface StartExamResp {
  exam_id: number
  record_id: number
  title: string
  duration: number
  questions: Question[]
  start_time: string
  end_time: string
  shuffle_q: boolean
  shuffle_opt: boolean
}

export const examApi = {
  create: (data: any) => request({ url: '/exams', method: 'POST', data }),
  list: (params: any) => request({ url: '/exams', method: 'GET', params }),
  available: () => request<Exam[]>({ url: '/exams/available', method: 'GET' }),
  start: (id: number) => request<StartExamResp>({ url: `/exams/${id}/start`, method: 'POST' }),
  saveAnswer: (data: { record_id: number; question_id: number; answer: unknown; elapsed?: number }) =>
    request({ url: '/exams/answer', method: 'POST', data }),
  audit: (data: { record_id: number; events: any }) =>
    request({ url: '/exams/audit', method: 'POST', data }),
  submit: (record_id: number) => request<ExamRecord>({ url: '/exams/submit', method: 'POST', data: { record_id } }),
  records: (id: number, params: any) => request({ url: `/exams/${id}/records`, method: 'GET', params }),
  myRecords: (params: any) => request({ url: '/exam-records/mine', method: 'GET', params }),
  gradeSubjective: (data: any) => request({ url: '/grading/subjective', method: 'POST', data }),
}
