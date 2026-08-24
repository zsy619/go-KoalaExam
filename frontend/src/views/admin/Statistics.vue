<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { statsApi, type ExamOverview } from '@/api/modules/stats'

const overview = ref<ExamOverview | null>(null)
const scoreRangeData = computed(() => {
  if (!overview.value) return []
  return Object.entries(overview.value.score_range).map(([range, count]) => ({ range, count }))
})
const loading = ref(false)
const examId = ref<number>(1)

async function load() {
  loading.value = true
  try {
    const { data } = await statsApi.examOverview(examId.value)
    overview.value = data!
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally { loading.value = false }
}
onMounted(load)
</script>

<template>
  <div class="koala-page">
    <el-card>
      <template #header>
        <div style="display:flex;gap:16px;align-items:center">
          <span>考试 ID：</span>
          <el-input-number v-model="examId" :min="1" />
          <el-button type="primary" @click="load" :loading="loading">刷新</el-button>
        </div>
      </template>
      <el-row :gutter="16" v-if="overview">
        <el-col :span="6"><el-card><div class="lbl">参考人数</div><div class="num">{{ overview.total_count }}</div></el-card></el-col>
        <el-col :span="6"><el-card><div class="lbl">平均分</div><div class="num">{{ overview.avg_score.toFixed(1) }}</div></el-card></el-col>
        <el-col :span="6"><el-card><div class="lbl">及格率</div><div class="num">{{ overview.pass_rate.toFixed(1) }}%</div></el-card></el-col>
        <el-col :span="6"><el-card><div class="lbl">最高 / 最低</div><div class="num">{{ overview.max_score.toFixed(0) }} / {{ overview.min_score.toFixed(0) }}</div></el-card></el-col>
      </el-row>
      <el-table v-if="overview" :data="scoreRangeData" style="margin-top:16px">
        <el-table-column prop="range" label="分数段" />
        <el-table-column prop="count" label="人数" />
        <el-table-column label="占比">
          <template #default="{ row }">{{ overview.total_count > 0 ? ((row.count / overview.total_count) * 100).toFixed(1) : 0 }}%</template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>


<style scoped lang="scss">
.lbl { color: #666; padding: 8px; }
.num { font-size: 28px; font-weight: 600; color: #409eff; text-align: center; padding: 8px; }
</style>
