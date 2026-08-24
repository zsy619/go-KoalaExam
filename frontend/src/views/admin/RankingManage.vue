<template>
  <div class="ranking-page">
    <el-card class="filter-card">
      <el-form :inline="true" :model="filter" @submit.prevent="loadRanking">
        <el-form-item label="选择考试">
          <el-select v-model="filter.exam_id" placeholder="选择考试查看排行" style="width: 240px;" @change="loadRanking">
            <el-option v-for="e in exams" :key="e.id" :label="e.title" :value="e.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序方式">
          <el-radio-group v-model="filter.order_by" @change="loadRanking">
            <el-radio-button value="score">分数</el-radio-button>
            <el-radio-button value="duration">速度</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="范围">
          <el-radio-group v-model="filter.scope" @change="loadRanking">
            <el-radio-button value="all">全部</el-radio-button>
            <el-radio-button value="passed">及格</el-radio-button>
          </el-radio-group>
        </el-form-item>
      </el-form>
    </el-card>

    <el-row :gutter="16" v-if="filter.exam_id">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-content">
            <el-icon class="stat-icon" color="#409eff"><User /></el-icon>
            <div>
              <div class="stat-value">{{ ranking.length }}</div>
              <div class="stat-label">参与人数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-content">
            <el-icon class="stat-icon" color="#67c23a"><Trophy /></el-icon>
            <div>
              <div class="stat-value">{{ topScore }}</div>
              <div class="stat-label">最高分</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-content">
            <el-icon class="stat-icon" color="#e6a23c"><DataLine /></el-icon>
            <div>
              <div class="stat-value">{{ avgScore }}</div>
              <div class="stat-label">平均分</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-content">
            <el-icon class="stat-icon" color="#f56c6c"><Warning /></el-icon>
            <div>
              <div class="stat-value">{{ passRate }}%</div>
              <div class="stat-label">及格率</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card style="margin-top:16px;">
      <el-empty v-if="!filter.exam_id" description="请先选择一场考试" />
      <el-empty v-else-if="ranking.length === 0 && !loading" description="暂无数据" />
      <el-table v-else :data="ranking" v-loading="loading">
        <el-table-column label="排名" width="80">
          <template #default="{ $index }">
            <span v-if="$index < 3" class="medal" :class="'medal-' + ($index + 1)">{{ $index + 1 }}</span>
            <span v-else>{{ $index + 1 }}</span>
          </template>
        </el-table-column>
        <el-table-column label="考生" min-width="180">
          <template #default="{ row }">
            <div class="user-cell">
              <el-avatar :size="32">{{ row.user_name?.[0] || row.user_id }}</el-avatar>
              <div>
                <div class="user-name">{{ row.user_name }}</div>
                <div class="user-id">ID: {{ row.user_id }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="total_score" label="总分" width="120" sortable>
          <template #default="{ row }">
            <span class="score-val">{{ row.total_score || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="objective_score" label="客观" width="100" />
        <el-table-column prop="subjective_score" label="主观" width="100" />
        <el-table-column label="用时" width="120">
          <template #default="{ row }">{{ formatDuration(row.duration) }}</template>
        </el-table-column>
        <el-table-column label="结果" width="100">
          <template #default="{ row }">
            <el-tag :type="row.passed ? 'success' : 'danger'">
              {{ row.passed ? '及格' : '未及格' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="submit_time" label="提交时间" width="180">
          <template #default="{ row }">{{ formatTime(row.submit_time) }}</template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { User, Trophy, DataLine, Warning } from '@element-plus/icons-vue'
import { examApi } from '@/api/modules/exam'

interface RankItem {
  user_id: number
  user_name: string
  total_score: number
  objective_score: number
  subjective_score: number
  duration: number
  passed: boolean
  submit_time: string
}

const loading = ref(false)
const exams = ref<any[]>([])
const ranking = ref<RankItem[]>([])

const filter = ref({
  exam_id: 0,
  order_by: 'score',
  scope: 'all'
})

const topScore = computed(() => {
  if (ranking.value.length === 0) return 0
  return Math.max(...ranking.value.map((r) => r.total_score || 0))
})

const avgScore = computed(() => {
  if (ranking.value.length === 0) return 0
  const sum = ranking.value.reduce((acc, r) => acc + (r.total_score || 0), 0)
  return (sum / ranking.value.length).toFixed(1)
})

const passRate = computed(() => {
  if (ranking.value.length === 0) return 0
  const passed = ranking.value.filter((r) => r.passed).length
  return ((passed / ranking.value.length) * 100).toFixed(0)
})

function formatDuration(s?: number) {
  if (!s) return '-'
  const m = Math.floor(s / 60)
  const sec = s % 60
  return m > 0 ? `${m}分${sec}秒` : `${sec}秒`
}

function formatTime(t?: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

async function loadExams() {
  try {
    const data: any = await examApi.list()
    exams.value = data.list || []
    if (exams.value.length > 0) {
      filter.value.exam_id = exams.value[0].id
      loadRanking()
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '加载考试列表失败')
  }
}

async function loadRanking() {
  if (!filter.value.exam_id) return
  loading.value = true
  try {
    // TODO: 后端实现 /stats/exam/:id/ranking 后调用
    // 这里使用 records 接口近似
    const res: any = await examApi.listRecordsByExam(filter.value.exam_id, 1, 50)
    let list: RankItem[] = (res.list || []).map((r: any) => ({
      user_id: r.user_id,
      user_name: r.user_name || `用户${r.user_id}`,
      total_score: r.total_score || 0,
      objective_score: r.objective_score || 0,
      subjective_score: r.subjective_score || 0,
      duration: r.duration || 0,
      passed: r.passed,
      submit_time: r.submit_time
    }))
    if (filter.value.scope === 'passed') {
      list = list.filter((r) => r.passed)
    }
    if (filter.value.order_by === 'score') {
      list.sort((a, b) => b.total_score - a.total_score)
    } else {
      list.sort((a, b) => a.duration - b.duration)
    }
    ranking.value = list
  } catch (e: any) {
    ElMessage.error(e?.message || '加载排行失败')
    ranking.value = []
  } finally {
    loading.value = false
  }
}

onMounted(loadExams)
</script>

<style scoped lang="scss">
.ranking-page { padding: 16px; }
.filter-card { margin-bottom: 16px; }
.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
}
.stat-icon { font-size: 36px; }
.stat-value { font-size: 24px; font-weight: 600; }
.stat-label { font-size: 13px; color: var(--el-text-color-secondary); margin-top: 4px; }
.medal {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  color: #fff;
  font-weight: 600;
  font-size: 13px;
}
.medal-1 { background: linear-gradient(135deg, #FFD700, #FFA500); }
.medal-2 { background: linear-gradient(135deg, #C0C0C0, #909399); }
.medal-3 { background: linear-gradient(135deg, #CD7F32, #8B4513); }
.user-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}
.user-name { font-weight: 500; }
.user-id { font-size: 12px; color: var(--el-text-color-secondary); }
.score-val { font-weight: 600; color: var(--el-color-primary); }
</style>
