import { defineStore } from 'pinia'
import { userApi, type LoginReq } from '@/api/modules/user'
import type { User } from '@/types/entity'

export const useUserStore = defineStore('user', {
  state: () => ({
    accessToken: localStorage.getItem('koala_token') || '',
    refreshToken: localStorage.getItem('koala_refresh') || '',
    profile: null as User | null,
  }),
  getters: {
    isLogin: (s) => !!s.accessToken,
    role: (s) => s.profile?.role || 3,
    isAdmin: (s) => s.profile?.role === 1,
    isTeacher: (s) => s.profile?.role === 2 || s.profile?.role === 1,
    isStudent: (s) => s.profile?.role === 3,
  },
  actions: {
    async login(payload: LoginReq) {
      const { data } = await userApi.login(payload)
      this.accessToken = data!.access_token
      this.refreshToken = data!.refresh_token
      this.profile = data!.user
      localStorage.setItem('koala_token', this.accessToken)
      localStorage.setItem('koala_refresh', this.refreshToken)
      localStorage.setItem('koala_user', JSON.stringify(this.profile))
    },
    async fetchProfile() {
      const { data } = await userApi.profile()
      this.profile = data!
      localStorage.setItem('koala_user', JSON.stringify(this.profile))
    },
    logout() {
      this.accessToken = ''
      this.refreshToken = ''
      this.profile = null
      localStorage.removeItem('koala_token')
      localStorage.removeItem('koala_refresh')
      localStorage.removeItem('koala_user')
    },
  },
})
