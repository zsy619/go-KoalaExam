// 用户认证 Pinia Store
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as loginApi, getProfile, logout as logoutApi } from '@/api/modules/user'
import type { User } from '@/types/entity'

const TOKEN_KEY = 'koala_token'
const REFRESH_KEY = 'koala_refresh_token'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem(TOKEN_KEY) || '')
  const refreshToken = ref<string>(localStorage.getItem(REFRESH_KEY) || '')
  const user = ref<User | null>(null)
  const loginAttempts = ref(0)
  const lockUntil = ref(0)

  const isLoggedIn = computed(() => !!token.value)
  const isRole = computed(() => (role: number) => user.value?.role === role)
  const isLocked = computed(() => Date.now() < lockUntil.value)
  // 别名（兼容旧代码）
  const profile = computed(() => user.value)
  const role = computed(() => user.value?.role || 3)
  const isAdmin = computed(() => user.value?.role === 1)
  const isTeacher = computed(() => user.value?.role === 2 || user.value?.role === 1)
  const isStudent = computed(() => user.value?.role === 3)

  async function loadProfile() {
    if (!token.value) return null
    try {
      const res: any = await getProfile()
      user.value = res || res.data
      return user.value
    } catch (e: any) {
      console.warn('Load profile failed:', e?.message)
      return null
    }
  }

  async function login(username: string, password: string, ip?: string): Promise<boolean> {
    if (isLocked.value) {
      const remaining = Math.ceil((lockUntil.value - Date.now()) / 1000)
      throw new Error(`账号已锁定，请 ${remaining} 秒后重试`)
    }

    try {
      const data: any = await loginApi({ username, password })
      // loginApi 已经解包 .data，这里 data 直接是 { user, access_token, ... }
      token.value = data?.access_token || data?.token || ''
      if (data?.refresh_token) refreshToken.value = data.refresh_token
      user.value = data?.user || null
      loginAttempts.value = 0
      localStorage.setItem(TOKEN_KEY, token.value)
      if (refreshToken.value) {
        localStorage.setItem(REFRESH_KEY, refreshToken.value)
      }
      return true
    } catch (e: any) {
      loginAttempts.value++
      if (loginAttempts.value >= 5) {
        lockUntil.value = Date.now() + 5 * 60 * 1000
      }
      throw e
    }
  }

  async function logout() {
    try {
      await logoutApi()
    } catch (e) {}
    token.value = ''
    refreshToken.value = ''
    user.value = null
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_KEY)
  }

  function updateProfile(p: Partial<User>) {
    if (user.value) {
      user.value = { ...user.value, ...p }
    }
  }

  function isTokenExpiringSoon(): boolean {
    if (!token.value) return true
    try {
      const payload = JSON.parse(atob(token.value.split('.')[1]))
      const exp = payload.exp * 1000
      return Date.now() > exp - 5 * 60 * 1000
    } catch {
      return true
    }
  }

  return {
    token, refreshToken, user, profile,
    isLoggedIn, isLogin: isLoggedIn, isRole, isLocked,
    isAdmin, isTeacher, isStudent, role,
    loginAttempts, lockUntil,
    login, logout, loadProfile, updateProfile, isTokenExpiringSoon
  }
})
