import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import type { BaseResp } from '@/types/resp'
import { useUserStore } from '@/store/modules/user'

const instance: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: Number(import.meta.env.VITE_API_TIMEOUT) || 30000,
})

// 请求拦截器：注入 Token
instance.interceptors.request.use((config) => {
  const userStore = useUserStore()
  if (userStore.token) {
    config.headers.Authorization = `Bearer ${userStore.token}`
  }
  return config
})

// 响应拦截器：统一错误处理
instance.interceptors.response.use(
  (resp) => {
    const data = resp.data as BaseResp
    if (data.code === 0) {
      // 返回业务载荷 (data.data)，调用方无需再解一层
      return data.data !== undefined ? data.data : data
    }
    if (data.code === 200002 || data.code === 200005 || data.code === 100002 || data.code === 100005) {
      // 未登录/token 失效
      const userStore = useUserStore()
      userStore.logout()
      window.location.href = '/login'
      ElMessage.error('登录已过期，请重新登录')
      return Promise.reject(data)
    }
    ElMessage.error(data.message || '请求失败')
    return Promise.reject(data)
  },
  (err) => {
    ElMessage.error(err.message || '网络错误')
    return Promise.reject(err)
  }
)

export const request = <T = unknown>(config: AxiosRequestConfig): Promise<T> => {
  return instance.request<BaseResp<T>>(config).then((resp: any) => {
    // 拦截器已经解包 data.data，这里确保返回 T
    return resp?.data?.data !== undefined ? resp.data.data : (resp?.data ?? resp)
  }) as unknown as Promise<T>
}

export default request
