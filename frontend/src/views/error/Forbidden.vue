<template>
  <el-result icon="warning" title="403" sub-title="抱歉，您无权访问该页面">
    <template #extra>
      <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
    </template>
  </el-result>
</template>

<script setup lang="ts"></script>

<style scoped>
.el-result { min-height: 80vh; display: flex; align-items: center; }
</style>
