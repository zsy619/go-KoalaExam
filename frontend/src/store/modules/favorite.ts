import { defineStore } from 'pinia'
import { favoriteApi } from '@/api/modules/favorite'
import type { FavoriteFolder } from '@/types/entity'

// 收藏态全局存储（避免重复请求后端查询收藏）
// key 格式: `${targetType}_${targetId}`
export const useFavoriteStore = defineStore('favorite', {
  state: () => ({
    favMap: new Map<string, boolean>(),
    folders: [] as FavoriteFolder[],
  }),
  actions: {
    // 初始化列表页收藏状态
    initBatch(items: { target_type: number; target_id: number; favorited?: boolean }[]) {
      items.forEach((it) => {
        if (typeof it.favorited === 'boolean') {
          this.favMap.set(`${it.target_type}_${it.target_id}`, it.favorited)
        }
      })
    },
    setOne(target_type: number, target_id: number, favorited: boolean) {
      this.favMap.set(`${target_type}_${target_id}`, favorited)
    },
    getOne(target_type: number, target_id: number): boolean {
      return this.favMap.get(`${target_type}_${target_id}`) || false
    },
    // 乐观更新：先改 UI，失败回滚
    async toggle(target_type: number, target_id: number, folder_id?: number) {
      const key = `${target_type}_${target_id}`
      const oldVal = this.getOne(target_type, target_id)
      this.favMap.set(key, !oldVal)
      try {
        await favoriteApi.toggle({ target_type, target_id, folder_id })
      } catch (e) {
        this.favMap.set(key, oldVal) // 回滚
        throw e
      }
    },
    async fetchFolders() {
      const { data } = await favoriteApi.listFolders()
      this.folders = data || []
    },
    async checkFromServer(target_type: number, target_id: number): Promise<boolean> {
      const { data } = await favoriteApi.check({ target_type, target_id })
      this.favMap.set(`${target_type}_${target_id}`, data!.favorited)
      return data!.favorited
    },
  },
})
