import request from '@/api'

// 统计接口
export interface DashboardSummary {
  user_total: number
  student_total: number
  teacher_total: number
  question_total: number
  paper_total: number
  exam_total: number
  record_total: number
  wrong_total: number
  today_records: number
}

export interface ExamOverview {
  exam_id: number
  avg_score: number
  max_score: number
  min_score: number
  total_count: number
  passed_count: number
  pass_rate: number
  score_range: Record<string, number>
}

export interface UserLearningStats {
  user_id: number
  total_favorites: number
  wrong_total: number
  mastery_dist: Record<string, number>
  exams_completed: number
  avg_score: number
  reviewed_count: number
}

export const statsApi = {
  // 超管首页
  dashboard: () => request<DashboardSummary>('/admin/stats/dashboard'),
  // 考试概览
  examOverview: (id: number) => request<ExamOverview>(`/stats/exam/${id}`),
  // 我的学习
  myLearning: () => request<UserLearningStats>('/stats/me'),
}
