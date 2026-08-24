<script setup lang="ts">
// 默认布局（占位，转发到 dashboard）
import { useRouter } from 'vue-router'
const router = useRouter()
router.replace('/admin/dashboard')
</script>

<template>
  <div></div>
</template>
