<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { examApi } from '@/api/modules/exam'
import { statsApi, type UserLearningStats } from '@/api/modules/stats'
import dayjs from 'dayjs'

const list = ref<any[]>([])
const total = ref(0)
const loading = ref(false)
const query = ref({ page: 1, page_size: 20 })
const stats = ref<UserLearningStats | null>(null)

async function fetchList() {
  loading.value = true
  try {
    const data: any = await examApi.myRecords(query.value)
    list.value = data.list || []
    total.value = data.total
  } finally { loading.value = false }
}

async function fetchStats() {
  try {
    const data: any = await statsApi.myLearning()
    stats.value = data || null
  } catch { /* ignore */ }
}

onMounted(() => { fetchList(); fetchStats() })
</script>

<template>
  <div class="koala-page">
    <el-row :gutter="16" v-if="stats" class="kpi-row">
      <el-col :span="6"><el-card><div class="lbl">⭐ 我的收藏</div><div class="num">{{ stats.total_favorites }}</div></el-card></el-col>
      <el-col :span="6"><el-card><div class="lbl">🐛 错题总数</div><div class="num">{{ stats.wrong_total }}</div></el-card></el-col>
      <el-col :span="6"><el-card><div class="lbl">✅ 已通过</div><div class="num">{{ stats.exams_completed }}</div></el-card></el-col>
      <el-col :span="6"><el-card><div class="lbl">📊 平均分</div><div class="num">{{ stats.avg_score.toFixed(1) }}</div></el-card></el-col>
    </el-row>
    <el-card style="margin-top:16px">
      <template #header>考试记录（{{ total }}）</template>
      <el-table :data="list" v-loading="loading">
        <el-table-column prop="exam_id" label="考试 ID" width="100" />
        <el-table-column label="提交时间" width="200">
          <template #default="{ row }">{{ row.submit_time ? dayjs(row.submit_time).format('YYYY-MM-DD HH:mm') : '-' }}</template>
        </el-table-column>
        <el-table-column prop="duration" label="用时" width="100">
          <template #default="{ row }">{{ Math.round(row.duration / 60) }} 分</template>
        </el-table-column>
        <el-table-column prop="total_score" label="总分" width="100" />
        <el-table-column prop="passed" label="是否通过" width="100">
          <template #default="{ row }"><el-tag :type="row.passed ? 'success' : 'danger'">{{ row.passed ? '是' : '否' }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="status" label="状态">
          <template #default="{ row }">
            <el-tag :type="row.status === 2 ? 'success' : (row.status === 1 ? 'warning' : 'info')">
              {{ { 0: '进行中', 1: '已交卷', 2: '已批改', 3: '异常' }[row.status] || '-' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.page_size"
        :total="total"
        @current-change="fetchList"
        layout="prev, pager, next"
        style="margin-top:16px;justify-content:flex-end"
      />
    </el-card>
  </div>
</template>

<style scoped lang="scss">
.lbl { color: #666; padding: 8px; }
.num { font-size: 28px; font-weight: 600; color: #409eff; text-align: center; padding: 8px; }
</style>
