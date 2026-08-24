import request from '@/api'

export const paperApi = {
  list: (params: any) => request({ url: '/papers', method: 'GET', params }),
  detail: (id: number) => request({ url: `/papers/${id}`, method: 'GET' }),
  create: (data: any) => request({ url: '/papers', method: 'POST', data }),
  remove: (id: number) => request({ url: `/papers/${id}`, method: 'DELETE' }),
}
