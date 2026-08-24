<template>
  <div class="grading-page">
    <el-card>
      <template #header>
        <div class="header">
          <span>人工批改</span>
          <div>
            <el-select v-model="filterExam" placeholder="按考试筛选" clearable @change="loadList" style="width: 200px;">
              <el-option v-for="e in exams" :key="e.id" :label="e.title" :value="e.id" />
            </el-select>
            <el-button @click="loadList" style="margin-left:8px;">刷新</el-button>
          </div>
        </div>
      </template>

      <el-table :data="records" v-loading="loading" @row-click="onSelect">
        <el-table-column prop="id" label="记录ID" width="80" />
        <el-table-column prop="user_name" label="考生" width="120" />
        <el-table-column prop="exam_title" label="考试" min-width="180" />
        <el-table-column label="总分" width="100">
          <template #default="{ row }">
            <el-tag>{{ row.objective_score || 0 }}</el-tag> / {{ row.total_score || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ getStatusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="submit_time" label="提交时间" width="180">
          <template #default="{ row }">{{ formatTime(row.submit_time) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click.stop="onSelect(row)">批改</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        layout="total, sizes, prev, pager, next"
        :page-sizes="[10, 20, 50]"
        @current-change="loadList"
        @size-change="loadList"
        style="margin-top: 16px; justify-content: flex-end;"
      />
    </el-card>

    <!-- 批改抽屉 -->
    <el-drawer v-model="drawerVisible" title="主观题批改" size="700px" direction="rtl">
      <div v-if="currentRecord" class="grading-detail">
        <div class="record-info">
          <h3>{{ currentRecord.exam_title }}</h3>
          <p>考生：{{ currentRecord.user_name }} | 总分：{{ currentRecord.objective_score || 0 }} / {{ currentRecord.total_score || 0 }}</p>
        </div>

        <el-divider />

        <div v-for="(q, idx) in subjectiveQuestions" :key="q.id" class="question-block">
          <div class="question-title">
            <span class="q-num">{{ idx + 1 }}.</span>
            <span>{{ q.title }}</span>
            <el-tag size="small">分值：{{ q.score }}</el-tag>
          </div>
          <div class="user-answer">
            <strong>考生答案：</strong>
            <div class="answer-content">{{ getUserAnswer(q.id) || '（未作答）' }}</div>
          </div>
          <div class="grading-input">
            <el-input-number
              v-model="grades[q.id]"
              :min="0"
              :max="q.score"
              :step="0.5"
              size="small"
              style="width: 140px;"
            />
            <span class="score-tip">/ {{ q.score }} 分</span>
            <el-button size="small" @click="grades[q.id] = q.score" style="margin-left: 8px;">满分</el-button>
            <el-button size="small" @click="grades[q.id] = 0" type="danger" plain>零分</el-button>
          </div>
          <el-divider />
        </div>

        <div class="drawer-footer">
          <el-button @click="drawerVisible = false">关闭</el-button>
          <el-button type="primary" @click="onSubmitGrades" :loading="submitting">提交批改</el-button>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { examApi } from '@/api/modules/exam'

interface Record {
  id: number
  exam_id: number
  exam_title: string
  user_id: number
  user_name: string
  status: number
  objective_score: number
  subjective_score: number
  total_score: number
  submit_time?: string
  subjective_needs_grading?: number
}

interface Question {
  id: number
  title: string
  type: number
  score: number
}

const loading = ref(false)
const submitting = ref(false)
const records = ref<Record[]>([])
const exams = ref<any[]>([])
const filterExam = ref<number | null>(null)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const drawerVisible = ref(false)
const currentRecord = ref<Record | null>(null)
const subjectiveQuestions = ref<Question[]>([])
const grades = ref<Record<number, number>>({})

function getStatusType(s: number) {
  return ['', 'warning', 'success', 'danger'][s] || 'info'
}

function getStatusText(s: number) {
  return ['', '已交卷', '已批改', '异常'][s] || '未知'
}

function formatTime(t?: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

function getUserAnswer(qid: number): string {
  if (!currentRecord.value?.answers) return ''
  try {
    const answers = typeof currentRecord.value.answers === 'string'
      ? JSON.parse(currentRecord.value.answers)
      : currentRecord.value.answers
    return String(answers[qid] ?? answers[String(qid)] ?? '')
  } catch {
    return ''
  }
}

async function loadList() {
  loading.value = true
  try {
    const data: any = await examApi.listRecords({ page: page.value, page_size: pageSize.value })
    records.value = data.list || []
    total.value = data.total || 0
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

async function loadExams() {
  try {
    const data: any = await examApi.list()
    exams.value = data.list || []
  } catch (e) {}
}

async function onSelect(row: Record) {
  currentRecord.value = row
  drawerVisible.value = true

  // 加载主观题
  try {
    const detail: any = await examApi.getRecord(row.id)
    subjectiveQuestions.value = (detail.questions || []).filter((q: Question) => q.type >= 3)
    // 初始化分数
    const def: Record<number, number> = {}
    subjectiveQuestions.value.forEach((q) => def[q.id] = 0)
    grades.value = def
  } catch (e: any) {
    ElMessage.error(e?.message || '加载试题失败')
  }
}

async function onSubmitGrades() {
  if (!currentRecord.value) return
  submitting.value = true
  try {
    const items = subjectiveQuestions.value.map((q) => ({
      question_id: q.id,
      score: grades.value[q.id] || 0
    }))
    await examApi.gradeSubjective({ record_id: currentRecord.value.id, items })
    ElMessage.success('批改完成')
    drawerVisible.value = false
    await loadList()
  } catch (e: any) {
    ElMessage.error(e?.message || '批改失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadExams()
  loadList()
})
</script>

<style scoped lang="scss">
.grading-page { padding: 16px; }
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.grading-detail { padding: 0 16px; }
.record-info h3 { margin: 0 0 8px; }
.record-info p { color: var(--el-text-color-secondary); margin: 0; }
.question-block { margin-bottom: 16px; }
.question-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  margin-bottom: 12px;
}
.q-num { color: var(--el-color-primary); }
.user-answer {
  padding: 12px;
  background: var(--el-color-info-light-9);
  border-radius: 4px;
  margin-bottom: 12px;
  line-height: 1.6;
}
.answer-content { white-space: pre-wrap; margin-top: 4px; }
.grading-input {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-left: 12px;
}
.score-tip { color: var(--el-text-color-secondary); }
.drawer-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 16px;
  border-top: 1px solid var(--el-border-color-lighter);
}
</style>
