// useFavorite 收藏管理组合式 API
//
// 功能：
//   - 切换/批量收藏
//   - 错题本智能筛选
//   - 收藏统计与分组
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as api from '@/api/modules/favorite'

export interface FavoriteStats {
  total: number
  byType: Record<number, number>
  byFolder: Record<number, number>
  recentWeek: number
}

export interface WrongBookItem {
  log_id: number
  question_id: number
  title: string
  wrong_count: number
  mastery_level: number
  last_wrong_at: string
  is_reviewed: boolean
  user_answer?: any
  correct_answer?: any
}

export function useFavorite() {
  // 状态
  const loading = ref(false)
  const favorites = ref<any[]>([])
  const folders = ref<any[]>([])
  const wrongBook = ref<WrongBookItem[]>([])
  const stats = ref<FavoriteStats>({
    total: 0,
    byType: {},
    byFolder: {},
    recentWeek: 0
  })

  // 切换收藏
  const toggle = async (targetType: number, targetID: number, folderID?: number) => {
    loading.value = true
    try {
      const res: any = await api.toggleFavorite({ target_type: targetType, target_id: targetID, folder_id: folderID })
      ElMessage.success(res.favorited ? '已收藏' : '已取消收藏')
      await loadStats()
      return res.favorited
    } catch (e: any) {
      ElMessage.error(e?.message || '操作失败')
      throw e
    } finally {
      loading.value = false
    }
  }

  // 批量收藏（错题自动入库）
  const batchAdd = async (questionIDs: number[], folderID?: number) => {
    if (questionIDs.length === 0) return { added_count: 0, skipped_count: 0 }
    loading.value = true
    try {
      const res: any = await api.batchAddFavorites({
        target_type: 1, // 题目
        target_ids: questionIDs,
        folder_id: folderID
      })
      ElMessage.success(`已添加 ${res.added_count} 题，跳过 ${res.skipped_count} 题`)
      await loadStats()
      return res
    } catch (e: any) {
      ElMessage.error(e?.message || '批量收藏失败')
      throw e
    } finally {
      loading.value = false
    }
  }

  // 加载收藏列表
  const loadFavorites = async (targetType?: number, folderID?: number) => {
    loading.value = true
    try {
      const data: any = await api.listFavorites(targetType, folderID)
      favorites.value = Array.isArray(data) ? data : []
    } catch (e: any) {
      ElMessage.error(e?.message || '加载失败')
    } finally {
      loading.value = false
    }
  }

  // 加载收藏夹
  const loadFolders = async () => {
    try {
      const res: any = await api.listFolders()
      folders.value = res || []
    } catch (e: any) {
      ElMessage.error(e?.message || '加载收藏夹失败')
    }
  }

  // 创建收藏夹
  const createFolder = async (name: string, color?: string, icon?: string) => {
    try {
      await api.createFolder({ name, color, icon })
      ElMessage.success('收藏夹已创建')
      await loadFolders()
    } catch (e: any) {
      ElMessage.error(e?.message || '创建失败')
    }
  }

  // 删除收藏夹（含确认）
  const deleteFolder = async (id: number) => {
    try {
      await ElMessageBox.confirm('确定删除该收藏夹？收藏项不会删除', '提示', {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消'
      })
      await api.deleteFolder(id)
      ElMessage.success('已删除')
      await loadFolders()
    } catch (e: any) {
      if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
    }
  }

  // 加载错题本
  const loadWrongBook = async (opts: { masteryLevel?: number; isReviewed?: boolean; page?: number; pageSize?: number } = {}) => {
    loading.value = true
    try {
      const data: any = await api.getWrongBook(opts)
      wrongBook.value = Array.isArray(data) ? data : (data.list || [])
      return { list: wrongBook.value, total: data.total || wrongBook.value.length }
    } catch (e: any) {
      ElMessage.error(e?.message || '加载错题本失败')
      return { list: [], total: 0 }
    } finally {
      loading.value = false
    }
  }

  // 标记已复习
  const markReviewed = async (logID: number, masteryLevel = 3) => {
    try {
      await api.markReviewed(logID, masteryLevel)
      ElMessage.success('已标记为已复习')
      // 局部更新
      const item = wrongBook.value.find((w) => w.log_id === logID)
      if (item) {
        item.is_reviewed = true
        item.mastery_level = masteryLevel
      }
    } catch (e: any) {
      ElMessage.error(e?.message || '标记失败')
    }
  }

  // 加载统计（从后端获取）
  const loadStats = async () => {
    try {
      const data: any = await api.getFavoriteStats()
      stats.value.total = data.total || 0
      stats.value.byType = data.by_type || {}
      stats.value.byFolder = data.by_folder || {}
      stats.value.recentWeek = data.recent_week_added || 0
    } catch (e) {
      // 静默失败（保持旧值）
    }
  }

  // 计算属性 - 按掌握度分组
  const wrongByMastery = computed(() => {
    const groups: Record<number, WrongBookItem[]> = {}
    wrongBook.value.forEach((w) => {
      const lvl = w.mastery_level || 1
      if (!groups[lvl]) groups[lvl] = []
      groups[lvl].push(w)
    })
    return groups
  })

  return {
    // 状态
    loading, favorites, folders, wrongBook, stats,
    // 方法
    toggle, batchAdd, loadFavorites, loadFolders, createFolder, deleteFolder,
    loadWrongBook, markReviewed, loadStats,
    // 计算
    wrongByMastery
  }
}
