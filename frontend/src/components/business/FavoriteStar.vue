<script setup lang="ts">
import { computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useFavorite } from '@/composables/useFavorite'

const props = defineProps<{
  targetType: number
  targetId: number
  size?: 'small' | 'default' | 'large'
  folderId?: number
}>()

const { isFav, toggle } = useFavorite()
const favorited = computed(() => isFav(props.targetType, props.targetId))

async function onClick(e: Event) {
  e.stopPropagation()
  try {
    await toggle(props.targetType, props.targetId, props.folderId)
  } catch {
    ElMessage.error('操作失败')
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
