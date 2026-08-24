import request from '@/api'
import type { User } from '@/types/entity'

export interface LoginReq { username: string; password: string 
  profileUpdate: (data: Partial<User>) => request({ url: '/user/profile', method: 'PUT', data }),
  changePassword: (data: { old_password: string; new_password: string }) => request({ url: '/user/password', method: 'PUT', data }),
  getByID: (id: number) => request<User>({ url: `/admin/users/${id}`, method: 'GET' }),
  toggleStatus: (id: number, status: number) => request({ url: `/admin/users/${id}/status`, method: 'PUT', data: { status } }),
  remove: (id: number) => request({ url: `/admin/users/${id}`, method: 'DELETE' }),
  examDetail: (id: number) => request({ url: `/exams/${id}`, method: 'GET' }),
  examUpdate: (id: number, data: any) => request({ url: `/exams/${id}`, method: 'PUT', data }),
  examRemove: (id: number) => request({ url: `/exams/${id}`, method: 'DELETE' }),
  recordDetail: (id: number) => request({ url: `/exam-records/${id}`, method: 'GET' }),
}
export interface LoginResp {
  user: User
  access_token: string
  refresh_token: string
  expires_in: number
}

export const userApi = {
  login: (data: LoginReq) => request<LoginResp>({ url: '/auth/login', method: 'POST', data }),
  refresh: (refresh_token: string) => request<{ access_token: string; expires_in: number }>({ url: '/auth/refresh', method: 'POST', data: { refresh_token } }),
  profile: () => request<User>({ url: '/user/profile', method: 'GET' }),
  list: (params: any) => request({ url: '/admin/users', method: 'GET', params }),
  create: (data: Partial<User>) => request({ url: '/admin/users', method: 'POST', data }),
  resetPwd: (id: number) => request<{ new_password: string }>({ url: `/admin/users/${id}/reset-password`, method: 'POST' }),
}
