<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { examApi } from '@/api/modules/exam'
import dayjs from 'dayjs'

const router = useRouter()
const list = ref<any[]>([])
const loading = ref(false)

async function fetchList() {
  loading.value = true
  try {
    const data: any = await examApi.available()
    list.value = Array.isArray(data) ? data : (data.list || [])
  } finally { loading.value = false }
}

function goStart(examId: number) {
  router.push({ name: 'ExamRoom', params: { id: examId } })
}

function fmt(s: string) { return dayjs(s).format('MM-DD HH:mm') }
onMounted(fetchList)
</script>

<template>
  <div class="koala-page">
    <h2>📚 考试大厅</h2>
    <el-empty v-if="!loading && list.length === 0" description="暂无可参加的考试" />
    <el-row :gutter="20" v-loading="loading">
      <el-col :span="8" v-for="exam in list" :key="exam.id" style="margin-bottom:20px">
        <el-card class="exam-card">
          <h3>{{ exam.title }}</h3>
          <p class="desc">{{ exam.description || '暂无描述' }}</p>
          <div class="meta"><el-icon><Calendar /></el-icon> {{ fmt(exam.start_time) }} ~ {{ fmt(exam.end_time) }}</div>
          <div class="meta"><el-icon><Clock /></el-icon> 时长 {{ exam.duration }} 分钟</div>
          <el-button type="primary" style="margin-top:12px;width:100%" @click="goStart(exam.id)">开始考试</el-button>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.exam-card h3 { margin-top: 0; }
.desc { color: #999; min-height: 40px; }
.meta { display: flex; align-items: center; gap: 6px; color: #666; padding: 4px 0; font-size: 13px; }
</style>
