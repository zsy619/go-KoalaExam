<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ level: 1 | 2 | 3 | 4 | 5 }>()

const cfg = computed(() => {
  switch (props.level) {
    case 5: return { text: '熟练', color: '#67c23a' }
    case 4: return { text: '良好', color: '#409eff' }
    case 3: return { text: '一般', color: '#e6a23c' }
    case 2: return { text: '薄弱', color: '#f78989' }
    default: return { text: '未掌握', color: '#f56c6c' }
  }
})
</script>

<template>
  <el-tag :color="cfg.color" effect="dark" size="small">
    {{ cfg.text }}
  </el-tag>
</template>
