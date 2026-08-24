<template>
  <div class="paper-page">
    <el-card>
      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-input
          v-model="query.keyword"
          clearable
          placeholder="搜索试卷名称"
          style="width: 220px;"
          @keyup.enter="fetchList"
        />
        <el-select v-model="query.strategy" placeholder="组卷策略" clearable style="width: 140px;" @change="fetchList">
          <el-option label="全部" :value="0" />
          <el-option label="固定" :value="1" />
          <el-option label="随机" :value="2" />
          <el-option label="遗传算法" :value="3" />
        </el-select>
        <el-button type="primary" @click="fetchList">
          <el-icon><Search /></el-icon> 搜索
        </el-button>
        <el-button type="success" @click="onCreate">
          <el-icon><Plus /></el-icon> 新建试卷
        </el-button>
        <el-button @click="fetchList">
          <el-icon><Refresh /></el-icon> 刷新
        </el-button>
      </div>

      <!-- 列表 -->
      <el-table :data="list" v-loading="loading" border stripe style="margin-top: 16px;">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="标题" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">
            <strong>{{ row.title }}</strong>
          </template>
        </el-table-column>
        <el-table-column label="组卷策略" width="120">
          <template #default="{ row }">
            <el-tag :type="strategyTagType(row.strategy)">
              {{ strategyName(row.strategy) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="total_score" label="总分" width="90" align="center" />
        <el-table-column prop="duration" label="时长(分)" width="100" align="center" />
        <el-table-column prop="pass_score" label="及格分" width="90" align="center" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" effect="plain">
              {{ statusName(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="题目数" width="90" align="center">
          <template #default="{ row }">
            {{ questionCountOf(row) }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="onPreview(row)">
              <el-icon><View /></el-icon>预览
            </el-button>
            <el-button type="warning" link @click="onEdit(row)">
              <el-icon><Edit /></el-icon>编辑
            </el-button>
            <el-button type="danger" link @click="onDelete(row)">
              <el-icon><Delete /></el-icon>删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.page_size"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        :page-sizes="[10, 20, 50]"
        @current-change="fetchList"
        @size-change="fetchList"
        style="margin-top: 16px; justify-content: flex-end;"
      />
    </el-card>

    <!-- 新建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑试卷' : '新建试卷'"
      width="900px"
      :close-on-click-modal="false"
      @close="onDialogClose"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
        label-position="right"
        v-loading="formLoading"
      >
        <el-form-item label="试卷标题" prop="title">
          <el-input v-model="form.title" maxlength="128" show-word-limit placeholder="请输入试卷标题" />
        </el-form-item>

        <el-form-item label="试卷描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="2"
            maxlength="500"
            show-word-limit
            placeholder="可选：试卷简介、考察范围等"
          />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="组卷策略" prop="strategy">
              <el-radio-group v-model="form.strategy" @change="onStrategyChange">
                <el-radio :value="1">固定</el-radio>
                <el-radio :value="2">随机</el-radio>
                <el-radio :value="3">遗传算法</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="考试时长" prop="duration">
              <el-input-number
                v-model="form.duration"
                :min="5"
                :max="600"
                :step="5"
                style="width: 100%;"
              />
              <span class="unit">分钟</span>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="总分" prop="total_score">
              <el-input-number
                v-model="form.total_score"
                :min="1"
                :max="500"
                :step="5"
                :precision="1"
                style="width: 100%;"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="及格分" prop="pass_score">
              <el-input-number
                v-model="form.pass_score"
                :min="0"
                :max="form.total_score"
                :step="1"
                :precision="1"
                style="width: 100%;"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 固定策略：题目选择 -->
        <template v-if="form.strategy === 1">
          <el-divider content-position="left">题目配置</el-divider>
          <el-form-item label="选择题目" prop="question_ids">
            <div class="question-picker">
              <div class="picker-toolbar">
                <el-input
                  v-model="questionQuery.keyword"
                  placeholder="搜索题目"
                  clearable
                  style="width: 200px;"
                  @input="onQuestionSearchDebounced"
                />
                <el-select v-model="questionQuery.type" placeholder="题型" clearable style="width: 130px;" @change="onQuestionSearch">
                  <el-option label="单选" :value="1" />
                  <el-option label="多选" :value="2" />
                  <el-option label="判断" :value="3" />
                  <el-option label="填空" :value="4" />
                </el-select>
                <el-select v-model="questionQuery.difficulty" placeholder="难度" clearable style="width: 130px;" @change="onQuestionSearch">
                  <el-option label="简单" :value="1" />
                  <el-option label="中等" :value="2" />
                  <el-option label="困难" :value="3" />
                </el-select>
                <el-button @click="onQuestionSearch">搜索</el-button>
              </div>
              <el-table
                ref="questionTableRef"
                :data="questions"
                v-loading="questionsLoading"
                height="380"
                @selection-change="onQuestionSelectionChange"
                border
                stripe
              >
                <el-table-column type="selection" width="50" />
                <el-table-column prop="id" label="ID" width="60" />
                <el-table-column prop="type" label="题型" width="80">
                  <template #default="{ row }">
                    <el-tag size="small">{{ typeName(row.type) }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="difficulty" label="难度" width="80">
                  <template #default="{ row }">
                    <el-rate v-model="row.difficulty" disabled :max="3" />
                  </template>
                </el-table-column>
                <el-table-column prop="title" label="题目" min-width="280" show-overflow-tooltip />
                <el-table-column prop="score" label="分值" width="70" align="center" />
              </el-table>

              <div class="picker-summary">
                <span>已选 <strong>{{ form.question_ids.length }}</strong> 题</span>
                <span>合计 <strong>{{ totalSelectedScore }}</strong> 分</span>
                <el-button v-if="form.question_ids.length > 0" type="danger" link size="small" @click="clearSelected">
                  清空选择
                </el-button>
              </div>
            </div>
          </el-form-item>
        </template>

        <!-- 随机策略：配置规则 -->
        <template v-else-if="form.strategy === 2 || form.strategy === 3">
          <el-divider content-position="left">随机组卷配置</el-divider>
          <el-form-item label="题型规则">
            <div class="random-rules">
              <el-table :data="form.config_rule.rules" border>
                <el-table-column label="题型" width="140">
                  <template #default="{ row }">
                    <el-select v-model="row.type" placeholder="选择题型">
                      <el-option label="单选" :value="1" />
                      <el-option label="多选" :value="2" />
                      <el-option label="判断" :value="3" />
                      <el-option label="填空" :value="4" />
                    </el-select>
                  </template>
                </el-table-column>
                <el-table-column label="难度" width="140">
                  <template #default="{ row }">
                    <el-select v-model="row.difficulty" placeholder="选择难度">
                      <el-option label="简单" :value="1" />
                      <el-option label="中等" :value="2" />
                      <el-option label="困难" :value="3" />
                    </el-select>
                  </template>
                </el-table-column>
                <el-table-column label="数量" width="140">
                  <template #default="{ row }">
                    <el-input-number v-model="row.count" :min="1" :max="100" />
                  </template>
                </el-table-column>
                <el-table-column label="每题分值" width="140">
                  <template #default="{ row }">
                    <el-input-number v-model="row.score" :min="0.5" :max="100" :step="0.5" :precision="1" />
                  </template>
                </el-table-column>
                <el-table-column label="小计" width="100">
                  <template #default="{ row }">{{ (row.count * row.score).toFixed(1) }}</template>
                </el-table-column>
                <el-table-column label="操作" width="80">
                  <template #default="{ $index }">
                    <el-button type="danger" link @click="form.config_rule.rules.splice($index, 1)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
              <el-button type="primary" plain style="margin-top: 8px;" @click="addRandomRule">
                <el-icon><Plus /></el-icon> 添加规则
              </el-button>
              <div class="picker-summary" style="margin-top: 12px;">
                <span>合计题目：<strong>{{ totalRandomCount }}</strong> 题</span>
                <span>合计分值：<strong>{{ totalRandomScore }}</strong> 分</span>
              </div>
            </div>
          </el-form-item>
        </template>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">
          {{ isEdit ? '保存修改' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 预览抽屉 -->
    <el-drawer v-model="previewVisible" title="试卷预览" size="780px" direction="rtl">
      <div v-if="previewData" class="preview-content">
        <h2 class="preview-title"> {{ previewData.paper.title }}</h2>
        <el-descriptions :column="2" border size="small" style="margin-bottom: 16px;">
          <el-descriptions-item label="组卷策略">
            <el-tag :type="strategyTagType(previewData.paper.strategy)" size="small">
              {{ strategyName(previewData.paper.strategy) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="总分">
            <strong style="color: #f56c6c;">{{ previewData.paper.total_score }} 分</strong>
          </el-descriptions-item>
          <el-descriptions-item label="时长">{{ previewData.paper.duration }} 分钟</el-descriptions-item>
          <el-descriptions-item label="及格分">{{ previewData.paper.pass_score }} 分</el-descriptions-item>
          <el-descriptions-item label="创建时间" :span="2">{{ formatTime(previewData.paper.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="试卷描述" :span="2">
            <span style="color: #909399;">{{ previewData.paper.description || '无' }}</span>
          </el-descriptions-item>
        </el-descriptions>

        <el-divider>题目列表 ({{ (previewData.questions || []).length }})</el-divider>
        <div v-for="(q, idx) in (previewData.questions || [])" :key="q.id" class="preview-question">
          <div class="q-header">
            <span class="q-num">{{ idx + 1 }}.</span>
            <el-tag size="small" effect="dark" :type="typeTagType(q.type)">{{ typeName(q.type) }}</el-tag>
            <el-tag size="small" effect="plain">{{ difficultyName(q.difficulty) }}</el-tag>
            <el-tag size="small" type="warning" effect="plain">{{ q.score || 0 }} 分</el-tag>
          </div>
          <div class="q-title">{{ q.title }}</div>

          <!-- 选项：单选/多选/不定项 -->
          <div v-if="[1, 2, 5].includes(q.type) && Array.isArray(q.options) && q.options.length > 0" class="q-options">
            <div v-for="opt in q.options" :key="opt.key || opt.label" class="q-option">
              <span class="opt-key">{{ opt.key || opt.label }}.</span>
              <span class="opt-text">{{ opt.text || opt.value || opt.label }}</span>
            </div>
          </div>

          <!-- 判断题 -->
          <div v-else-if="q.type === 3" class="q-options">
            <div class="q-option"><span class="opt-key">A.</span> 正确（）</div>
            <div class="q-option"><span class="opt-key">B.</span> 错误（）</div>
          </div>

          <!-- 填空题 -->
          <div v-else-if="q.type === 4" class="q-options">
            <div class="q-option blank-line">考生在此填写：______________________</div>
          </div>

          <!-- 编程题 -->
          <div v-else-if="q.type === 6" class="q-options">
            <div class="q-option code-mode"> 请在 IDE 中实现并提交代码评测</div>
          </div>

          <!-- 答案 -->
          <div v-if="q.answer" class="q-answer">
            <span class="answer-icon"></span>
            <span class="answer-label">参考答案：</span>
            <span class="answer-content">{{ q.answer }}</span>
          </div>

          <!-- 解析 -->
          <div v-if="q.analysis" class="q-analysis">
            <span class="analysis-icon"></span>
            <span class="analysis-label">解析：</span>
            <span class="analysis-content">{{ q.analysis }}</span>
          </div>
        </div>

        <el-empty v-if="(previewData.questions || []).length === 0" description="暂无题目（随机组卷需在考试时生成）" />
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus, Edit, Delete, View, Search, Refresh
} from '@element-plus/icons-vue'
import { paperApi } from '@/api/modules/paper'
import { questionApi } from '@/api/modules/question'
import type { PaperDetail } from '@/api/modules/paper'

const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const query = ref({
  page: 1,
  page_size: 10,
  keyword: '',
  strategy: 0,
})

const dialogVisible = ref(false)
const isEdit = ref(false)
const editingId = ref<number | null>(null)
const formLoading = ref(false)
const submitting = ref(false)
const formRef = ref()
const questionTableRef = ref()

function defaultForm() {
  return {
    title: '',
    description: '',
    strategy: 1 as number,
    duration: 60,
    total_score: 100,
    pass_score: 60,
    question_ids: [] as number[],
    config_rule: {
      rules: [] as Array<{ type: number; difficulty: number; count: number; score: number }>,
      total_score: 100,
    },
  }
}

const form = ref(defaultForm())

const rules = {
  title: [
    { required: true, message: '请输入试卷标题', trigger: 'blur' },
    { min: 2, max: 128, message: '标题长度 2-128 字', trigger: 'blur' },
  ],
  duration: [
    { required: true, message: '请设置考试时长', trigger: 'blur' },
    { type: 'number', min: 10, max: 600, message: '时长 10-600 分钟', trigger: 'blur' },
  ],
  total_score: [
    { required: true, message: '请设置总分', trigger: 'blur' },
    { type: 'number', min: 1, max: 1000, message: '总分 1-1000', trigger: 'blur' },
  ],
  pass_score: [
    { required: true, message: '请设置及格分', trigger: 'blur' },
    { type: 'number', min: 0, max: 1000, message: '及格分 0-1000', trigger: 'blur' },
    {
      validator: (_: any, value: number, cb: any) => {
        if (value > (form.value?.total_score || 0)) {
          return cb(new Error('及格分不能超过总分'))
        }
        cb()
      },
      trigger: 'blur',
    },
  ],
  strategy: [
    { required: true, message: '请选择组卷策略', trigger: 'change' },
  ],
}

// 题目选择
const questionsLoading = ref(false)
const questions = ref<any[]>([])
const questionQuery = ref({
  page: 1,
  page_size: 100,
  keyword: '',
  type: undefined as number | undefined,
  difficulty: undefined as number | undefined,
})

let searchTimer: any = null
function onQuestionSearchDebounced() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(onQuestionSearch, 300)
}

// 预览
const previewVisible = ref(false)
const previewData = ref<PaperDetail | null>(null)

// ============ 计算属性 ============
const totalSelectedScore = computed(() => {
  return form.value.question_ids.reduce((sum, qid) => {
    const q = questions.value.find((x: any) => x.id === qid)
    return sum + (q?.score || 1)
  }, 0)
})

const totalRandomCount = computed(() =>
  form.value.config_rule.rules.reduce((sum, r) => sum + (r.count || 0), 0)
)

const totalRandomScore = computed(() =>
  form.value.config_rule.rules.reduce((sum, r) => sum + (r.count * r.score || 0), 0)
)

// ============ 工具函数 ============
function strategyName(s: number) {
  return ['', '固定', '随机', '遗传算法'][s] || '未知'
}
function strategyTagType(s: number) {
  return ['', '', 'success', 'warning'][s] || 'info'
}
function statusName(s: number) {
  return ['', '已发布', '草稿', '已归档'][s] || '未知'
}
function statusTagType(s: number) {
  return ['', 'success', 'info', 'danger'][s] || 'info'
}
function typeName(t: number) {
  return ['', '单选', '多选', '判断', '填空', '不定项', '编程'][t] || '其他'
}
function typeTagType(t: number) {
  return ({ 1: 'primary', 2: 'warning', 3: 'success', 4: 'info', 5: 'warning', 6: 'danger' } as Record<number, string>)[t] || ''
}
function difficultyName(d?: number) {
  return ['', '简单', '中等', '困难'][d || 0] || '未知'
}
function formatTime(t?: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}
function questionCountOf(row: any): number {
  if (!row.question_ids) return 0
  try {
    const ids = typeof row.question_ids === 'string' ? JSON.parse(row.question_ids) : row.question_ids
    return Array.isArray(ids) ? ids.length : 0
  } catch {
    return 0
  }
}

// ============ 数据加载 ============
async function fetchList() {
  loading.value = true
  try {
    const data: any = await paperApi.list(query.value)
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

async function onQuestionSearch() {
  questionsLoading.value = true
  try {
    const params: any = {
      page: 1,
      page_size: 100,
    }
    if (questionQuery.value.keyword) params.keyword = questionQuery.value.keyword
    if (questionQuery.value.type !== undefined) params.type = questionQuery.value.type
    if (questionQuery.value.difficulty !== undefined) params.difficulty = questionQuery.value.difficulty
    const data: any = await questionApi.list(params)
    questions.value = data.list || []
    // 默认勾选已选的
    await nextTick()
    if (questionTableRef.value && form.value.question_ids.length > 0) {
      form.value.question_ids.forEach((qid: number) => {
        const row = questions.value.find((x: any) => x.id === qid)
        if (row) questionTableRef.value.toggleRowSelection(row, true)
      })
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '加载题目失败')
  } finally {
    questionsLoading.value = false
  }
}

function onQuestionSelectionChange(rows: any[]) {
  form.value.question_ids = rows.map((r: any) => r.id)
}

// ============ 操作 ============
function onCreate() {
  isEdit.value = false
  editingId.value = null
  form.value = defaultForm()
  dialogVisible.value = true
  loadQuestionSearch()
}

async function onEdit(row: any) {
  isEdit.value = true
  editingId.value = row.id
  formLoading.value = true
  dialogVisible.value = true
  try {
    const data: PaperDetail = await paperApi.detail(row.id)
    form.value = {
      title: data.paper.title,
      description: data.paper.description || '',
      strategy: data.paper.strategy,
      duration: data.paper.duration,
      total_score: data.paper.total_score,
      pass_score: data.paper.pass_score,
      question_ids: parseQuestionIds(data.paper.question_ids),
      config_rule: parseConfigRule(data.paper.config_rule),
    }
    await nextTick()
    loadQuestionSearch()
  } catch (e: any) {
    ElMessage.error(e?.message || '加载试卷详情失败')
    dialogVisible.value = false
  } finally {
    formLoading.value = false
  }
}

async function onPreview(row: any) {
  try {
    const data: any = await paperApi.detail(row.id)
    if (!data) {
      previewData.value = null
    } else {
      // 解析每道题的 options/answer 字符串
      const rawQuestions = data.questions || []
      previewData.value = {
        paper: data.paper || data,
        questions: rawQuestions.map((q: any) => ({
          ...q,
          options: parseOptionsField(q.options),
          answer: parseAnswerField(q.answer),
        })),
      }
    }
    previewVisible.value = true
  } catch (e: any) {
    ElMessage.error(e?.message || '加载预览失败')
  }
}

// 解析 options（可能是 JSON 字符串或数组）
function parseOptionsField(v: any): any[] {
  if (!v) return []
  if (Array.isArray(v)) return v
  try { return JSON.parse(v || '[]') } catch { return [] }
}

// 解析 answer（可能是字符串/数字/JSON 数组/布尔）
function parseAnswerField(v: any): string {
  if (v === null || v === undefined || v === '') return ''
  if (typeof v === 'boolean') return v ? '正确' : '错误'
  if (typeof v === 'number') return String(v)
  if (Array.isArray(v)) return v.join(', ')
  if (typeof v === 'string') {
    const s = v.trim()
    // JSON 数组
    if (s.startsWith('[')) {
      try {
        const arr = JSON.parse(s)
        return Array.isArray(arr) ? arr.join(', ') : s
      } catch { return s }
    }
    // JSON 字符串
    if (s.startsWith('"') && s.endsWith('"')) {
      try { return JSON.parse(s) } catch { return s }
    }
    return s
  }
  return String(v)
}

async function onDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      `确认删除试卷"${row.title}"？此操作不可恢复`,
      '删除确认',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消', confirmButtonClass: 'el-button--danger' }
    )
    await paperApi.remove(row.id)
    ElMessage.success('删除成功')
    await fetchList()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
  }
}

function onStrategyChange() {
  // 切换策略时重置相关字段
  if (form.value.strategy === 1) {
    form.value.config_rule.rules = []
  } else {
    form.value.question_ids = []
    if (form.value.config_rule.rules.length === 0) {
      addRandomRule()
    }
  }
}

function addRandomRule() {
  form.value.config_rule.rules.push({
    type: 1,
    difficulty: 1,
    count: 5,
    score: 2,
  })
}

function clearSelected() {
  form.value.question_ids = []
  if (questionTableRef.value) {
    questionTableRef.value.clearSelection()
  }
}

function onDialogClose() {
  formRef.value?.resetFields()
  form.value = defaultForm()
  editingId.value = null
  isEdit.value = false
}

async function loadQuestionSearch() {
  return onQuestionSearch()
}

function parseQuestionIds(s: string | number[]): number[] {
  if (Array.isArray(s)) return s
  if (!s) return []
  try {
    const arr = JSON.parse(s)
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}

function parseConfigRule(s: string): any {
  if (!s) return { rules: [], total_score: 100 }
  try {
    const cfg = typeof s === 'string' ? JSON.parse(s) : s
    return {
      rules: cfg.rules || [],
      total_score: cfg.total_score || 100,
    }
  } catch {
    return { rules: [], total_score: 100 }
  }
}

async function onSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    const payload: any = {
      title: form.value.title,
      description: form.value.description,
      strategy: form.value.strategy,
      total_score: form.value.total_score,
      duration: form.value.duration,
      pass_score: form.value.pass_score,
    }

    if (form.value.strategy === 1) {
      if (form.value.question_ids.length === 0) {
        ElMessage.error('请至少选择一道题目')
        submitting.value = false
        return
      }
      if (form.value.question_ids.length > 200) {
        ElMessage.error('固定组卷最多 200 道题')
        submitting.value = false
        return
      }
      payload.question_ids = form.value.question_ids
    } else {
      if (form.value.config_rule.rules.length === 0) {
        ElMessage.error('请至少添加一条随机规则')
        submitting.value = false
        return
      }
      // 检查每条规则的合法性
      for (const r of form.value.config_rule.rules) {
        if (!r.type) { ElMessage.error('请设置规则题型'); submitting.value = false; return }
        if (!r.count || r.count < 1) { ElMessage.error('每条规则抽取数量至少 1'); submitting.value = false; return }
        if (!r.score || r.score < 1) { ElMessage.error('每条规则分值至少 1'); submitting.value = false; return }
      }
      payload.config_rule = form.value.config_rule
    }

    if (isEdit.value && editingId.value !== null) {
      await paperApi.update(editingId.value, payload)
      ElMessage.success('更新成功')
    } else {
      await paperApi.create(payload)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    await fetchList()
  } catch (e: any) {
    ElMessage.error(e?.message || (isEdit.value ? '更新失败' : '创建失败'))
  } finally {
    submitting.value = false
  }
}

onMounted(fetchList)
</script>

<style scoped lang="scss">
.paper-page { padding: 16px; }
.filter-bar {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}
.unit {
  margin-left: 8px;
  color: var(--el-text-color-secondary);
}
.question-picker {
  width: 100%;
}
.picker-toolbar {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}
.picker-summary {
  display: flex;
  gap: 24px;
  align-items: center;
  margin-top: 12px;
  padding: 8px 12px;
  background: var(--el-color-info-light-9);
  border-radius: 4px;
  font-size: 14px;
  color: var(--el-text-color-regular);
}
.picker-summary strong {
  color: var(--el-color-primary);
  font-size: 16px;
  margin: 0 4px;
}
.random-rules {
  width: 100%;
}
.preview-content {
  padding: 0 16px;
}
.preview-content h2 {
  margin: 0 0 16px;
  font-size: 20px;
}
.preview-question {
  padding: 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.q-title {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
  margin-bottom: 8px;
}
.q-num {
  color: var(--el-color-primary);
  font-weight: 600;
  margin-right: 4px;
}
.q-options {
  padding-left: 24px;
  margin-bottom: 8px;
}
.q-option {
  padding: 4px 0;
}
.q-analysis {
  padding: 8px;
  background: var(--el-color-warning-light-9);
  border-radius: 4px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}
.preview-content { padding: 0 20px 20px; }
.preview-title {
  margin: 8px 0 16px;
  font-size: 22px;
  font-weight: 600;
  color: #303133;
}
.preview-question {
  padding: 16px;
  margin-bottom: 16px;
  background: #fafbfc;
  border-radius: 8px;
  border-left: 4px solid #409eff;
  transition: all 0.2s;
}
.preview-question:hover {
  background: #f0f7ff;
}
.q-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}
.q-num {
  color: #409eff;
  font-weight: 600;
  margin-right: 4px;
  font-size: 15px;
}
.q-title {
  font-size: 15px;
  color: #303133;
  line-height: 1.7;
  margin-bottom: 12px;
  padding: 4px 0;
}
.q-options {
  padding: 8px 0 8px 20px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 8px;
}
.q-option {
  display: flex;
  gap: 10px;
  font-size: 14px;
  color: #606266;
  line-height: 1.6;
  padding: 4px 8px;
  border-radius: 4px;
}
.q-option.blank-line {
  color: #909399;
  font-style: italic;
}
.q-option.code-mode {
  background: #fef0f0;
  color: #f56c6c;
  padding: 10px 12px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
}
.opt-key {
  font-weight: 600;
  color: #409eff;
  min-width: 24px;
}
.opt-text {
  flex: 1;
}
.q-answer {
  padding: 10px 12px;
  background: #f0f9ff;
  border-left: 3px solid #67c23a;
  border-radius: 4px;
  font-size: 13px;
  margin: 8px 0;
  display: flex;
  align-items: flex-start;
  gap: 6px;
}
.answer-icon { color: #67c23a; font-weight: bold; }
.answer-label { color: #67c23a; font-weight: 600; flex-shrink: 0; }
.answer-content { color: #303133; word-break: break-word; }
.q-analysis {
  padding: 10px 12px;
  background: #fdf6ec;
  border-left: 3px solid #e6a23c;
  border-radius: 4px;
  font-size: 13px;
  color: #606266;
  line-height: 1.7;
  display: flex;
  gap: 6px;
}
.analysis-icon { flex-shrink: 0; }
.analysis-label { color: #e6a23c; font-weight: 600; flex-shrink: 0; }
.analysis-content { word-break: break-word; }
</style>
