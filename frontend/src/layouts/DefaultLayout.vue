<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter, RouterView } from 'vue-router'

const router = useRouter()
onMounted(() => {
  // 跳转到 dashboard
  router.replace('/dashboard')
})
</script>

<template>
  <RouterView />
</template>

<style scoped>
:deep(.el-result) { min-height: 80vh; display: flex; align-items: center; }
</style>
