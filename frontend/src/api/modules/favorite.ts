import request from '@/api'
import type { FavoriteFolder, WrongQuestionItem } from '@/types/entity'

export const favoriteApi = {
  toggle: (data: { target_type: number; target_id: number; folder_id?: number; note?: string }) =>
    request<{ favorited: boolean }>({ url: '/favorites/toggle', method: 'POST', data }),
  batch: (data: { target_type: number; target_ids: number[]; folder_id?: number; source_type?: number }) =>
    request<{ count: number }>({ url: '/favorites/batch', method: 'POST', data }),
  check: (params: { target_type: number; target_id: number }) =>
    request<{ favorited: boolean }>({ url: '/favorites/check', method: 'GET', params }),
  list: (params: any) => request({ url: '/favorites', method: 'GET', params }),
  listFolders: () => request<FavoriteFolder[]>({ url: '/favorite-folders', method: 'GET' }),
  createFolder: (data: { name: string; color?: string; icon?: string }) =>
    request<{ id: number }>({ url: '/favorite-folders', method: 'POST', data }),
  deleteFolder: (id: number) => request({ url: `/favorite-folders/${id}`, method: 'DELETE' }),
  wrongBook: (params: { mastery_level?: number; page?: number; page_size?: number }) =>
    request<{ list: WrongQuestionItem[]; total: number }>({ url: '/wrong-book', method: 'GET', params }),
  markReviewed: (id: number, mastery_level: number) =>
    request({ url: `/wrong-log/${id}/reviewed`, method: 'POST', params: { mastery_level } }),
}
