<script setup lang="ts">
import { computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useFavoriteStore } from '@/store/modules/favorite'
import { useUserStore } from '@/store/modules/user'

const props = defineProps<{
  targetType: number
  targetId: number
  size?: 'small' | 'default' | 'large'
  folderId?: number
}>()

const favStore = useFavoriteStore()
const userStore = useUserStore()

const favorited = computed(() =>
  userStore.isLogin && favStore.getOne(props.targetType, props.targetId)
)

async function onClick(e: Event) {
  e.stopPropagation()
  if (!userStore.isLogin) {
    ElMessage.warning('请先登录')
    return
  }
  try {
    await favStore.toggle(props.targetType, props.targetId, props.folderId)
  } catch (err: any) {
    ElMessage.error(err?.message || '操作失败')
  }
}
</script>

<template>
  <el-button :size="size" link @click="onClick">
    <el-icon :color="favorited ? '#e6a23c' : '#999'">
      <StarFilled v-if="favorited" />
      <Star v-else />
    </el-icon>
  </el-button>
</template>
