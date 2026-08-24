import request from '@/api'
import type { User } from '@/types/entity'

export interface LoginReq {
  username: string
  password: string
}

export interface LoginResp {
  user: User
  access_token: string
  refresh_token: string
  expires_in: number
}

// 同时导出独立函数（供 store 解构使用）
export const login = (data: LoginReq) =>
  request<LoginResp>({ url: '/auth/login', method: 'POST', data })

export const getProfile = () =>
  request<User>({ url: '/user/profile', method: 'GET' })

export const logout = () =>
  request({ url: '/auth/logout', method: 'POST' })

// 主对象（供 userApi.* 使用）
export const userApi = {
  // Auth
  login: (data: LoginReq) => request<LoginResp>({ url: '/auth/login', method: 'POST', data }),
  // 注册接口已移除，请联系管理员创建账号
  refresh: (refresh_token: string) => request<{ access_token: string; expires_in: number }>({ url: '/auth/refresh', method: 'POST', data: { refresh_token } }),
  logout: () => request({ url: '/auth/logout', method: 'POST' }),

  // Profile
  profile: () => request<User>({ url: '/user/profile', method: 'GET' }),
  profileUpdate: (data: Partial<User>) => request({ url: '/user/profile', method: 'PUT', data }),
  changePassword: (data: { old_password: string; new_password: string }) => request({ url: '/user/password', method: 'PUT', data }),

  // Admin: users
  list: (params: any) => request({ url: '/admin/users', method: 'GET', params }),
  create: (data: Partial<User> & { password: string }) => request({ url: '/admin/users', method: 'POST', data }),
  detail: (id: number) => request<User>({ url: `/admin/users/${id}`, method: 'GET' }),
  update: (id: number, data: Partial<User>) => request({ url: `/admin/users/${id}`, method: 'PUT', data }),
  resetPwd: (id: number) => request<{ new_password: string }>({ url: `/admin/users/${id}/reset-password`, method: 'POST' }),
  // 管理员更新用户（昵称/邮箱/手机/角色等）
  adminUpdate: (id: number, data: Partial<User>) => request({ url: `/admin/users/${id}`, method: 'PUT', data }),
  toggleStatus: (id: number, status: number) => request({ url: `/admin/users/${id}/status`, method: 'PUT', data: { status } }),
  remove: (id: number) => request({ url: `/admin/users/${id}`, method: 'DELETE' }),

  // 其他保留（之前 dashboard 调用）
  getByID: (id: number) => request<User>({ url: `/admin/users/${id}`, method: 'GET' }),
  examDetail: (id: number) => request({ url: `/exams/${id}`, method: 'GET' }),
  examUpdate: (id: number, data: any) => request({ url: `/exams/${id}`, method: 'PUT', data }),
  examRemove: (id: number) => request({ url: `/exams/${id}`, method: 'DELETE' }),
  recordDetail: (id: number) => request({ url: `/exam-records/${id}`, method: 'GET' }),
}
