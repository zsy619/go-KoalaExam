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
  detail: (id: number) => request<Exam>({ url: `/exams/${id}`, method: 'GET' }),
  update: (id: number, data: any) => request({ url: `/exams/${id}`, method: 'PUT', data }),
  remove: (id: number) => request({ url: `/exams/${id}`, method: 'DELETE' }),
  available: () => request<Exam[]>({ url: '/exams/available', method: 'GET' }),
  start: (id: number) => request<StartExamResp>({ url: `/exams/${id}/start`, method: 'POST' }),
  saveAnswer: (data: { record_id: number; question_id: number; answer: unknown; elapsed?: number }) =>
    request({ url: '/exams/answer', method: 'POST', data }),
  audit: (data: { record_id: number; events: any }) =>
    request({ url: '/exams/audit', method: 'POST', data }),
  submit: (record_id: number) => request<ExamRecord>({ url: '/exams/submit', method: 'POST', data: { record_id } }),
  records: (id: number, params: any) => request({ url: `/exams/${id}/records`, method: 'GET', params }),
  myRecords: (params: any) => request({ url: '/exam-records/mine', method: 'GET', params }),
  // 考试记录管理
  listRecordsByExam: (examId: number, page: number, pageSize: number) =>
    request({ url: `/exams/${examId}/records`, method: 'GET', params: { page, page_size: pageSize } }),
  listRecords: (params: any) => request({ url: '/exam-records/mine', method: 'GET', params }),
  getRecord: (id: number) => request({ url: `/exam-records/${id}`, method: 'GET' }),

  // 主观题批改
  gradeSubjective: (data: any) => request({ url: '/grading/subjective', method: 'POST', data }),
}
