import request from '@/api'

export const favoriteApi = {
  // 切换收藏
  toggle: (data: { target_type: number; target_id: number; folder_id?: number; note?: string }) =>
    request({ url: '/favorites/toggle', method: 'POST', data }),
  // 批量收藏
  batchAdd: (data: { target_type: number; target_ids: number[]; folder_id?: number }) =>
    request({ url: '/favorites/batch', method: 'POST', data }),
  // 检查
  check: (target_type: number, target_id: number) =>
    request({ url: '/favorites/check', method: 'GET', params: { target_type, target_id } }),
  // 列出
  list: (target_type?: number, folder_id?: number) =>
    request({ url: '/favorites', method: 'GET', params: { target_type, folder_id } }),
  // 文件夹
  listFolders: () => request({ url: '/favorite-folders', method: 'GET' }),
  createFolder: (data: { name: string; color?: string; icon?: string }) =>
    request({ url: '/favorite-folders', method: 'POST', data }),
  deleteFolder: (id: number) => request({ url: `/favorite-folders/${id}`, method: 'DELETE' }),
  // 错题本
  getWrongBook: (opts: { mastery_level?: number; is_reviewed?: boolean; page?: number; page_size?: number } = {}) =>
    request({ url: '/wrong-book', method: 'GET', params: opts }),
  markReviewed: (logID: number, mastery_level: number) =>
    request({ url: `/wrong-log/${logID}/reviewed`, method: 'POST', data: null, params: { mastery_level } }),
  // 统计
  stats: () => request({ url: '/favorites/stats', method: 'GET' }),
  masteryDistribution: () => request({ url: '/wrong-book/distribution', method: 'GET' }),
}

// 同时导出独立函数（兼容 composables 里的旧 import）
export function toggleFavorite(data: any) {
  return favoriteApi.toggle(data)
}
export function batchAddFavorites(data: any) {
  return favoriteApi.batchAdd(data)
}
export function checkFavorite(targetType: number, targetID: number) {
  return favoriteApi.check(targetType, targetID)
}
export function listFavorites(targetType?: number, folderID?: number) {
  return favoriteApi.list(targetType, folderID)
}
export function listFolders() {
  return favoriteApi.listFolders()
}
export function createFolder(data: any) {
  return favoriteApi.createFolder(data)
}
export function deleteFolder(id: number) {
  return favoriteApi.deleteFolder(id)
}
export function getWrongBook(opts: any = {}) {
  return favoriteApi.getWrongBook(opts)
}
export function markReviewed(logID: number, masteryLevel: number) {
  return favoriteApi.markReviewed(logID, masteryLevel)
}
export function getFavoriteStats() {
  return favoriteApi.stats()
}
export function getMasteryDistribution() {
  return favoriteApi.masteryDistribution()
}
