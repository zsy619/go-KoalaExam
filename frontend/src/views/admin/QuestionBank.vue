<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { questionApi } from '@/api/modules/question'
import FavoriteStar from '@/components/business/FavoriteStar.vue'

const list = ref<any[]>([])
const total = ref(0)
const loading = ref(false)
const categories = ref<any[]>([])
const query = ref({ page: 1, page_size: 15, category_id: 0, type: 0, difficulty: 0, keyword: '' })

const currentPage = ref(1)
const pageSize = ref(15)

function handlePageChange(p: number) {
  query.value.page = p
  fetchList()
}
function handleSizeChange(s: number) {
  pageSize.value = s
  query.value.page_size = s
  query.value.page = 1
  currentPage.value = 1
  fetchList()
}

const formRef = ref<FormInstance>()

const rules: FormRules = {
  category_id: [
    { required: true, message: '请选择分类', trigger: 'change' },
  ],
  type: [
    { required: true, message: '请选择题型', trigger: 'change' },
  ],
  title: [
    { required: true, message: '请输入题干', trigger: 'blur' },
    { min: 5, max: 1000, message: '题干长度 5-1000 字', trigger: 'blur' },
  ],
  options: [
    {
      validator: (_: any, value: any[], cb: any) => {
        const type = editing.value?.type
        if ([1, 2, 5].includes(type)) {
          if (!Array.isArray(value) || value.length < 2) {
            return cb(new Error('至少需要 2 个选项'))
          }
          if (value.some((o: any) => !o.text || !o.text.trim())) {
            return cb(new Error('每个选项内容不能为空'))
          }
        }
        cb()
      },
      trigger: 'change',
    },
  ],
  answer: [
    { required: true, message: '请输入答案', trigger: 'blur' },
  ],
  score: [
    { type: 'number', min: 1, max: 100, message: '分值 1-100', trigger: 'blur' },
  ],
}

async function fetchList() {
  loading.value = true
  try {
    const data: any = await questionApi.list(query.value)
    list.value = (data && data.list) || []
    total.value = (data && data.total) || 0
  } finally {
    loading.value = false
  }
}

async function fetchCategories() {
  const cats: any = await questionApi.categories()
  categories.value = Array.isArray(cats) ? cats : []
}

const dialogVisible = ref(false)
const editing = ref<any>({})

// 分类查找
function getCategoryName(id: number): string {
  const c = categories.value.find(x => x.id === id)
  return c ? c.name : '#' + id
}

function typeName(t: number): string {
  return ['', '单选', '多选', '判断', '填空', '不定项', '编程'][t] || '未知'
}

function typeTagType(t: number) {
  // Element Plus 内置颜色：success/warning/danger/info/primary
  return ({ 1: 'primary', 2: 'warning', 3: 'success', 4: 'info', 5: 'warning', 6: 'danger' } as Record<number, string>)[t] || ''
}

function defaultOptions() {
  return [
    { key: 'A', text: '' },
    { key: 'B', text: '' },
    { key: 'C', text: '' },
    { key: 'D', text: '' },
  ]
}

function normalizeQuestion(q: any) {
  let opts: any[] = []
  try {
    const raw = typeof q.options === 'string' ? JSON.parse(q.options || '[]') : (q.options || [])
    opts = Array.isArray(raw) ? raw : []
  } catch {
    opts = []
  }
  if (opts.length === 0) opts = defaultOptions()

  let ans: any = ''
  try {
    const raw = typeof q.answer === 'string' ? JSON.parse(q.answer || '""') : (q.answer || '')
    if (Array.isArray(raw)) ans = raw.join(',')
    else ans = String(raw)
  } catch {
    ans = ''
  }
  if (ans === null || ans === undefined) ans = ''

  return {
    id: q.id,
    category_id: q.category_id,
    type: q.type,
    difficulty: q.difficulty || 2,
    title: q.title || '',
    options: opts,
    answer: String(ans),
    analysis: q.analysis || '',
    tags: q.tags || '',
    score: q.score || 2,
  }
}

function openCreate() {
  editing.value = {
    type: 1,
    difficulty: 2,
    title: '',
    category_id: categories.value[0]?.id || 1,
    options: defaultOptions(),
    answer: '',
    analysis: '',
    tags: '',
    score: 2,
  }
  dialogVisible.value = true
}

function openEdit(row: any) {
  editing.value = normalizeQuestion(row)
  dialogVisible.value = true
}

async function onSubmit() {
  // 1. 表单校验
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    ElMessage.warning('请检查必填项')
    return
  }

  // 2. 类型特定校验
  const t = editing.value.type
  if ([1, 2, 5].includes(t)) {
    const opts = editing.value.options || []
    if (opts.length < 2) {
      ElMessage.error('单选/多选/不定项至少需要 2 个选项')
      return
    }
    if (opts.some((o: any) => !o.text?.trim())) {
      ElMessage.error('选项内容不能为空')
      return
    }
  }

  const payload: any = { ...editing.value }
  if ([1, 2, 5].includes(payload.type)) {
    payload.options = JSON.stringify(payload.options || [])
  } else {
    payload.options = null
  }
  let ans: string[] = []
  if (typeof payload.answer === 'string') {
    if (payload.type === 3) {
      ans = ['true'].includes(String(payload.answer).toLowerCase()) ? ['true'] : ['false']
    } else if (payload.type === 4) {
      ans = String(payload.answer).split(/[,，]s*/).filter(Boolean)
    } else if (payload.type === 2 || payload.type === 5) {
      ans = String(payload.answer).split(/[s,，]+/).filter(Boolean)
    } else {
      ans = [String(payload.answer)]
    }
  } else if (Array.isArray(payload.answer)) {
    ans = payload.answer
  }
  payload.answer = JSON.stringify(ans)

  try {
    if (payload.id) {
      await questionApi.update(payload.id, payload)
    } else {
      await questionApi.create(payload)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    fetchList()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  }
}

function onRemove(id: number) {
  ElMessageBox.confirm('确定删除该题目？', '提示', { type: 'warning' })
    .then(async () => {
      await questionApi.remove(id)
      fetchList()
    })
}

onMounted(() => {
  fetchCategories()
  fetchList()
})
</script>

<template>
  <div class="koala-page">
    <el-card>
      <div class="filter">
        <el-input v-model="query.keyword" placeholder="题干关键词" style="width:200px" clearable />
        <el-select v-model="query.category_id" placeholder="分类" clearable style="width:160px">
          <el-option v-for="c in categories" :key="c.id" :value="c.id" :label="c.name" />
        </el-select>
        <el-select v-model="query.type" placeholder="题型" clearable style="width:140px">
          <el-option :value="1" label="单选" />
          <el-option :value="2" label="多选" />
          <el-option :value="3" label="判断" />
          <el-option :value="4" label="填空" />
          <el-option :value="5" label="不定项" />
        </el-select>
        <el-button type="primary" @click="fetchList">搜索</el-button>
        <el-button type="success" @click="openCreate">+ 新建题目</el-button>
        <el-button>批量导入（Excel）</el-button>
      </div>

      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="分类" width="120">
          <template #default="{ row }">
            <span class="cat-text">{{ getCategoryName(row.category_id) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="题型" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="typeTagType(row.type)" effect="light">{{ typeName(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="题干" show-overflow-tooltip />
        <el-table-column prop="difficulty" label="难度" width="80" />
        <el-table-column prop="score" label="分值" width="80" />
        <el-table-column label="收藏" width="80">
          <template #default="{ row }">
            <FavoriteStar :target-type="1" :target-id="row.id" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="onRemove(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="total"
          :page-sizes="[15, 30, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editing.id ? '编辑题目' : '新建题目'" width="700">
      <el-form ref="formRef" :model="editing" :rules="rules" label-width="80px">
        <el-form-item label="分类" prop="category_id">
          <el-select v-model="editing.category_id" placeholder="请选择分类">
            <el-option v-for="c in categories" :key="c.id" :value="c.id" :label="c.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="题型" prop="type">
          <el-select v-model="editing.type" placeholder="请选择题型">
            <el-option :value="1" label="单选" />
            <el-option :value="2" label="多选" />
            <el-option :value="3" label="判断" />
            <el-option :value="4" label="填空" />
            <el-option :value="5" label="不定项" />
          </el-select>
        </el-form-item>
        <el-form-item label="难度" prop="difficulty">
          <el-rate v-model="editing.difficulty" :max="3" show-score :score-template="(v: number) => ['', '简单', '中等', '困难'][v] || '未知'" />
        </el-form-item>
        <el-form-item label="题干" prop="title">
          <el-input v-model="editing.title" type="textarea" :rows="3" placeholder="请输入题干" />
        </el-form-item>

        <template v-if="[1, 2, 5].includes(editing.type)" prop="options">
          <el-form-item v-for="opt in (editing.options || [])" :key="opt.key" :label="'选项 ' + opt.key">
            <el-input v-model="opt.text" :placeholder="'选项 ' + opt.key + ' 内容'" />
          </el-form-item>
        </template>

        <el-form-item label="答案" prop="answer">
          <el-radio-group v-if="editing.type === 3" v-model="editing.answer">
            <el-radio label="true">正确（✓）</el-radio>
            <el-radio label="false">错误（✗）</el-radio>
          </el-radio-group>
          <el-input
            v-else
            v-model="editing.answer"
            :placeholder="editing.type === 2 || editing.type === 5 ? '多选用逗号分隔，如 A,B,C' : editing.type === 4 ? '填空答案' : '单选答案，如 A'"
          />
        </el-form-item>

        <el-form-item label="分值" prop="score">
          <el-input-number v-model="editing.score" :min="1" :max="100" />
        </el-form-item>
        <el-form-item label="解析">
          <el-input v-model="editing.analysis" type="textarea" :rows="2" placeholder="答案解析（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="onSubmit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.filter { display: flex; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  padding: 8px 0;
}
</style>
