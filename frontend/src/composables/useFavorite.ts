import { useFavoriteStore } from '@/store/modules/favorite'

/**
 * 收藏 Composable
 * - 乐观更新 + 失败回滚
 * - 用于 FavoriteStar 组件
 */
export function useFavorite() {
  const store = useFavoriteStore()

  function isFav(target_type: number, target_id: number) {
    return store.getOne(target_type, target_id)
  }

  async function toggle(target_type: number, target_id: number, folder_id?: number) {
    await store.toggle(target_type, target_id, folder_id)
  }

  return { isFav, toggle }
}
