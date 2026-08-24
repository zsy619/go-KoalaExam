<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { examApi } from '@/api/modules/exam'
import { questionApi } from '@/api/modules/question'

interface Record {
  id: number
  exam_id: number
  exam_title: string
  user_id: number
  user_name: string
  user_account: string
  status: number
  objective_score: number
  subjective_score: number
  total_score: number
  submit_time?: string
  answers?: string
  duration?: number
  tab_switch_cnt?: number
}

interface Question {
  id: number
  title: string
  type: number
  score: number
  options?: any[]
  answer?: string | string[]
  analysis?: string
}

interface GradedDetail {
  question_id: number
  score: number
  comment: string
  grader_id?: number
  graded_at?: string
}

// ===== 列表 =====
const loading = ref(false)
const records = ref<Record[]>([])
const exams = ref<any[]>([])
const filterExam = ref<number | null>(null)
const filterStatus = ref<number | null>(null)
const filterKeyword = ref('')
const page = ref(1)
const pageSize = ref(15)
const total = ref(0)

function getStatusType(s: number) {
  return ['', 'warning', 'success', 'danger'][s] || 'info'
}
function getStatusText(s: number) {
  return ['进行中', '已交卷', '已批改', '异常'][s] || '未知'
}
function formatTime(t?: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}
function typeName(t: number) {
  return ['', '单选', '多选', '判断', '填空', '不定项', '编程'][t] || '其他'
}
function typeTagType(t: number) {
  return ({ 1: 'primary', 2: 'warning', 3: 'success', 4: 'info', 5: 'warning', 6: 'danger' } as Record<number, string>)[t] || ''
}

function parseOptions(opts: any): any[] {
  if (!opts) return []
  if (Array.isArray(opts)) return opts
  try { return JSON.parse(opts) } catch { return [] }
}

function parseAnswer(ans: any): string {
  if (ans === null || ans === undefined || ans === '') return ''
  if (typeof ans === 'boolean') return ans ? '正确' : '错误'
  if (Array.isArray(ans)) return ans.join(', ')
  if (typeof ans === 'string') {
    if (ans.startsWith('[')) {
      try { return JSON.parse(ans).join(', ') } catch { return ans }
    }
    if (ans.startsWith('"')) {
      try { return JSON.parse(ans) } catch { return ans }
    }
    return ans
  }
  return String(ans)
}

async function loadList() {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (filterExam.value) params.exam_id = filterExam.value
    const data: any = await examApi.listRecords(params)
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

watch([filterExam, filterStatus, filterKeyword], () => {
  page.value = 1
  loadList()
})

// ===== 抽屉（批改详情）=====
const drawerVisible = ref(false)
const currentRecord = ref<Record | null>(null)
const detailLoading = ref(false)
const questions = ref<Question[]>([])
const answersMap = ref<Record<number, any>>({})
const grades = reactive<Record<number, number>>({})
const comments = reactive<Record<number, string>>({})
const gradedDetails = ref<GradedDetail[]>([])
const submitting = ref(false)

// 进度
const gradedCount = computed(() =>
  questions.value.filter(q => grades[q.id] !== undefined && grades[q.id] !== null).length
)
const totalScore = computed(() =>
  questions.value.reduce((s, q) => s + (grades[q.id] || 0), 0)
)
const fullScore = computed(() =>
  questions.value.reduce((s, q) => s + (q.score || 0), 0)
)
const progress = computed(() =>
  questions.value.length ? Math.round((gradedCount.value / questions.value.length) * 100) : 0
)

async function onSelect(row: Record) {
  currentRecord.value = row
  drawerVisible.value = true
  detailLoading.value = true
  // 重置
  questions.value = []
  answersMap.value = {}
  gradedDetails.value = []
  Object.keys(grades).forEach(k => delete grades[+k])
  Object.keys(comments).forEach(k => delete comments[+k])

  try {
    // 1. 拿考试详情（含试卷信息）
    const examDetail: any = await examApi.detail(row.exam_id)
    const paperId = examDetail?.paper_id

    // 2. 拿试卷详情（含题目列表）
    let paperQuestions: Question[] = []
    if (paperId) {
      const paperDetail: any = await examApi.records(paperId, { page: 1, page_size: 100 })
      // 试卷题目列表
      paperQuestions = (paperDetail.list || []).map((q: any) => ({ ...q }))
    }

    // 3. 拿考试记录详情（含考生答案 + paper_snapshot）
    let recordDetail: any = null
    try {
      recordDetail = await examApi.getRecord(row.id)
    } catch {}

    // 4. 优先从 paper_snapshot 取题目（快照保留了完整题目数据）
    if (recordDetail?.paper_snapshot) {
      try {
        const snap = typeof recordDetail.paper_snapshot === 'string'
          ? JSON.parse(recordDetail.paper_snapshot)
          : recordDetail.paper_snapshot
        if (Array.isArray(snap) && snap.length > 0) {
          paperQuestions = snap.filter((q: any) => q.type >= 3) // 只显示主观题
        }
      } catch {}
    }

    questions.value = paperQuestions

    // 5. 解析考生答案
    const answersStr = recordDetail?.answers || row.answers || '{}'
    try {
      const userAns = JSON.parse(answersStr)
      answersMap.value = userAns
    } catch {
      answersMap.value = {}
    }

    // 6. 解析已有批改详情
    if (recordDetail?.subjective_detail) {
      try {
        gradedDetails.value = typeof recordDetail.subjective_detail === 'string'
          ? JSON.parse(recordDetail.subjective_detail)
          : recordDetail.subjective_detail
        // 回填到 grades 和 comments
        gradedDetails.value.forEach(d => {
          grades[d.question_id] = d.score
          comments[d.question_id] = d.comment || ''
        })
      } catch {}
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '加载详情失败')
  } finally {
    detailLoading.value = false
  }
}

function getUserAnswer(qid: number): string {
  const v = answersMap.value[qid] ?? answersMap.value[String(qid)]
  return v === undefined || v === null || v === '' ? '（未作答）' : parseAnswer(v)
}

function setScore(qid: number, score: number) {
  grades[qid] = score
}

function setFullScore(q: Question) {
  grades[q.id] = q.score
}
function setZero(qid: number) {
  grades[qid] = 0
}
function setHalf(q: Question) {
  grades[q.id] = Math.round(q.score * 0.5 * 10) / 10
}

async function onSubmitGrades() {
  if (!currentRecord.value) return
  // 校验所有主观题是否都已批改
  const ungraded = questions.value.filter(q => grades[q.id] === undefined || grades[q.id] === null)
  if (ungraded.length > 0) {
    try {
      await ElMessageBox.confirm(
        `还有 ${ungraded.length} 道题未批改，确认提交吗？`,
        '提示',
        { type: 'warning' }
      )
    } catch {
      return
    }
  }
  submitting.value = true
  try {
    const items = questions.value.map(q => ({
      question_id: q.id,
      score: grades[q.id] || 0,
      comment: comments[q.id] || '',
    }))
    await examApi.gradeSubjective({
      record_id: currentRecord.value.id,
      items,
    })
    ElMessage.success(`批改完成！主观题得分： ${totalScore.value} / ${fullScore.value}`)
    drawerVisible.value = false
    await loadList()
  } catch (e: any) {
    ElMessage.error(e?.message || '批改失败')
  } finally {
    submitting.value = false
  }
}

async function onResetGrades() {
  try {
    await ElMessageBox.confirm('重置当前所有评分？', '提示', { type: 'warning' })
    Object.keys(grades).forEach(k => delete grades[+k])
    Object.keys(comments).forEach(k => delete comments[+k])
  } catch {}
}

function handlePageChange(p: number) { page.value = p; loadList() }
function handleSizeChange(s: number) { pageSize.value = s; page.value = 1; loadList() }

onMounted(() => {
  loadExams()
  loadList()
})
</script>

<template>
  <div class="grading-page">
    <el-card>
      <!-- 搜索栏 -->
      <div class="filter">
        <el-input v-model="filterKeyword" placeholder="搜索考生账号/昵称" clearable style="width: 220px;" @keyup.enter="loadList" />
        <el-select v-model="filterExam" placeholder="按考试筛选" clearable style="width: 220px;">
          <el-option v-for="e in exams" :key="e.id" :label="e.title" :value="e.id" />
        </el-select>
        <el-button type="primary" @click="loadList">
          <el-icon><Search /></el-icon> 搜索
        </el-button>
      </div>

      <!-- 列表 -->
      <el-table :data="records" v-loading="loading" stripe border style="margin-top: 16px;" @row-click="onSelect">
        <el-table-column prop="id" label="记录ID" width="80" />
        <el-table-column label="考生" width="160">
          <template #default="{ row }">
            <div>
              <strong>{{ row.user_name || row.user_account }}</strong>
            </div>
            <div style="font-size:12px;color:#909399;">@{{ row.user_account }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="exam_title" label="考试" min-width="180" show-overflow-tooltip />
        <el-table-column label="客观分" width="100" align="center">
          <template #default="{ row }">
            <el-tag effect="plain">{{ row.objective_score || 0 }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="主观分" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.subjective_score > 0 ? 'success' : 'danger'"' effect="plain">
              {{ row.subjective_score || 0 }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="总分" width="100" align="center">
          <template #default="{ row }">
            <strong :style="{ color: row.passed ? '#67c23a' : '#f56c6c' }">
              {{ row.total_score || 0 }}
            </strong>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ getStatusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="提交时间" width="170">
          <template #default="{ row }">{{ formatTime(row.submit_time) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click.stop="onSelect(row)">批改</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          :page-sizes="[15, 30, 50, 100]"
          background
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>

    <!-- 批改抽屉 -->
    <el-drawer v-model="drawerVisible" :title="'批改详情：' + (currentRecord?.exam_title || '')" size="900px" direction="rtl">
      <div v-if="currentRecord" v-loading="detailLoading" class="grading-detail">
        <!-- 考生信息卡片 -->
        <el-card shadow="never" class="info-card">
          <div class="info-grid">
            <div class="info-item">
              <span class="label">考生：</span>
              <strong>{{ currentRecord.user_name }}</strong>
              <span class="account">@{{ currentRecord.user_account }}</span>
            </div>
            <div class="info-item">
              <span class="label">考试：</span>
              <strong>{{ currentRecord.exam_title }}</strong>
            </div>
            <div class="info-item">
              <span class="label">提交时间：</span>
              <span>{{ formatTime(currentRecord.submit_time) }}</span>
            </div>
            <div class="info-item">
              <span class="label">用时：</span>
              <span>{{ Math.floor((currentRecord.duration || 0) / 60) }} 分 {{ (currentRecord.duration || 0) % 60 }} 秒</span>
            </div>
            <div class="info-item">
              <span class="label">切屏：</span>
              <el-tag :type="(currentRecord.tab_switch_cnt || 0) > 0 ? 'warning' : 'success'" size="small">
                {{ currentRecord.tab_switch_cnt || 0 }} 次
              </el-tag>
            </div>
            <div class="info-item">
              <span class="label">状态：</span>
              <el-tag :type="getStatusType(currentRecord.status)">{{ getStatusText(currentRecord.status) }}</el-tag>
            </div>
          </div>

          <!-- 成绩摘要 -->
          <el-divider />
          <div class="score-summary">
            <div class="score-item">
              <div class="score-label">客观题</div>
              <div class="score-value">{{ currentRecord.objective_score || 0 }}</div>
            </div>
            <div class="score-plus">+</div>
            <div class="score-item">
              <div class="score-label">主观题 (待批改)</div>
              <div class="score-value">{{ totalScore }} / {{ fullScore }}</div>
            </div>
            <div class="score-plus">=</div>
            <div class="score-item total">
              <div class="score-label">总分</div>
              <div class="score-value">{{ (currentRecord.objective_score || 0) + totalScore }}</div>
            </div>
          </div>

          <!-- 批改进度 -->
          <el-progress
            :percentage="progress"
            :status="progress === 100 ? 'success' : ''"
            :stroke-width="10"
            style="margin-top: 12px;"
          />
          <div class="progress-text">
            已批改 {{ gradedCount }} / {{ questions.length }} 题
          </div>
        </el-card>

        <!-- 主观题列表 -->
        <div class="questions-section">
          <h3 class="section-title">
            <el-icon><Document /></el-icon>
            主观题 ({{ questions.length }} 题)
          </h3>

          <el-empty v-if="!detailLoading && questions.length === 0" description="该试卷无主观题" />

          <div v-for="(q, idx) in questions" :key="q.id" class="question-card">
            <div class="question-header">
              <span class="q-num">{{ idx + 1 }}.</span>
              <el-tag size="small" effect="dark" :type="typeTagType(q.type)">{{ typeName(q.type) }}</el-tag>
              <span class="q-title">{{ q.title }}</span>
              <el-tag size="small" type="warning" effect="plain">{{ q.score }} 分</el-tag>
              <el-tag v-if="grades[q.id] !== undefined" size="small" type="success" effect="plain">
                已批：{{ grades[q.id] }} 分
              </el-tag>
            </div>

            <!-- 考生答案 -->
            <div class="answer-box user-answer">
              <div class="box-label">考生答案</div>
              <div class="box-content">{{ getUserAnswer(q.id) }}</div>
            </div>

            <!-- 参考答案 -->
            <div class="answer-box ref-answer" v-if="q.answer">
              <div class="box-label">参考答案</div>
              <div class="box-content">{{ parseAnswer(q.answer) }}</div>
            </div>

            <!-- 题目解析 -->
            <div class="answer-box analysis" v-if="q.analysis">
              <div class="box-label">题目解析</div>
              <div class="box-content">{{ q.analysis }}</div>
            </div>

            <!-- 批改输入区 -->
            <div class="grading-area">
              <div class="grading-row">
                <span class="grading-label">评分：</span>
                <el-input-number
                  v-model="grades[q.id]"
                  :min="0"
                  :max="q.score"
                  :step="0.5"
                  :precision="1"
                  size="default"
                  style="width: 140px;"
                  placeholder="未批"
                />
                <span class="score-tip">/ {{ q.score }} 分</span>
                <el-button-group style="margin-left: 8px;">
                  <el-button size="small" @click="setFullScore(q)">满分</el-button>
                  <el-button size="small" @click="setHalf(q)">一半</el-button>
                  <el-button size="small" type="danger" plain @click="setZero(q.id)">零分</el-button>
                </el-button-group>
              </div>
              <el-input
                v-model="comments[q.id]"
                type="textarea"
                :rows="2"
                placeholder="评语（可选）：对考生答案的点评..."
                maxlength="500"
                show-word-limit
                style="margin-top: 8px;"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- 抽屉底部操作 -->
      <template #footer>
        <div class="drawer-footer">
          <div class="footer-info">
            <el-tag>已批 {{ gradedCount }} 题</el-tag>
            <el-tag type="success">得分 {{ totalScore }} / {{ fullScore }}</el-tag>
          </div>
          <div>
            <el-button @click="onResetGrades" :disabled="gradedCount === 0">重置</el-button>
            <el-button @click="drawerVisible = false">关闭</el-button>
            <el-button type="primary" :loading="submitting" @click="onSubmitGrades">
              提交批改
            </el-button>
          </div>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<style scoped lang="scss">
.grading-page { padding: 16px; }
.filter { display: flex; gap: 12px; margin-bottom: 8px; flex-wrap: wrap; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; }

.grading-detail { padding: 0 16px 16px; }
.info-card { margin-bottom: 16px; }

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  font-size: 14px;
}
.info-item .label { color: #909399; margin-right: 4px; }
.info-item .account { color: #909399; font-size: 12px; margin-left: 8px; }

.score-summary {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 12px 0;
}
.score-item {
  text-align: center;
  padding: 8px 16px;
  background: #f5f7fa;
  border-radius: 8px;
  min-width: 100px;
}
.score-item.total {
  background: #ecf5ff;
  border: 1px solid #409eff;
}
.score-label { font-size: 12px; color: #909399; margin-bottom: 4px; }
.score-value { font-size: 22px; font-weight: 600; color: #303133; }
.score-plus { font-size: 20px; color: #c0c4cc; }

.progress-text { text-align: center; color: #909399; font-size: 12px; margin-top: 4px; }

.section-title {
  margin: 16px 0 12px;
  font-size: 16px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 6px;
  color: #303133;
}

.question-card {
  padding: 16px;
  margin-bottom: 16px;
  background: #fafbfc;
  border-radius: 8px;
  border-left: 4px solid #409eff;
}

.question-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.q-num { color: #409eff; font-weight: 600; font-size: 15px; }
.q-title {
  flex: 1;
  color: #303133;
  font-size: 14px;
  line-height: 1.6;
}

.answer-box {
  padding: 10px 12px;
  border-radius: 6px;
  margin-bottom: 10px;
  font-size: 13px;
}
.user-answer { background: #fdf6ec; border-left: 3px solid #e6a23c; }
.ref-answer { background: #f0f9ff; border-left: 3px solid #67c23a; }
.analysis { background: #f4f4f5; border-left: 3px solid #909399; color: #606266; }
.box-label {
  font-size: 12px;
  font-weight: 600;
  color: #606266;
  margin-bottom: 4px;
}
.box-content {
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.7;
  color: #303133;
}

.grading-area {
  padding: 12px;
  background: #ecf5ff;
  border-radius: 6px;
}
.grading-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.grading-label { font-weight: 600; color: #606266; }
.score-tip { color: #909399; font-size: 13px; }

.drawer-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
}
.footer-info {
  display: flex;
  gap: 8px;
}
</style>
