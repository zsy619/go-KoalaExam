export type UserRole = 1 | 2 | 3  // 1超管 2教师 3学生

export interface User {
  id: number
  username: string
  nickname?: string
  email?: string
  avatar?: string
  role: UserRole
  status: number
  class_id?: number
  department_id?: number
  created_at?: string
}

export type QuestionType = 1 | 2 | 3 | 4 | 5 | 6

export interface QuestionOption { key: string; text: string }

export interface Question {
  id: number
  category_id: number
  type: QuestionType
  difficulty: 1 | 2 | 3
  title: string
  options?: QuestionOption[]
  score: number
  answer?: string[] | string | boolean
  analysis?: string
  tags?: string
}

export interface Exam {
  id: number
  title: string
  paper_id: number
  start_time: string
  end_time: string
  duration: number
  shuffle_q: boolean
  shuffle_opt: boolean
  anti_cheat: boolean
  status: number
}

export interface ExamRecord {
  id: number
  exam_id: number
  user_id: number
  status: number
  start_time: string
  submit_time?: string
  total_score: number
  objective_score: number
  subjective_score: number
  passed: boolean
}

export interface FavoriteFolder {
  id: number
  name: string
  color?: string
  icon?: string
  is_system: boolean
  question_cnt: number
}

export interface WrongQuestionItem {
  log_id: number
  question: Question
  wrong_count: number
  last_wrong_at: string
  is_reviewed: boolean
  mastery_level: 1 | 2 | 3 | 4 | 5
  user_answer?: unknown
  correct_answer?: unknown
}
