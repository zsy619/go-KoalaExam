<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { examApi } from '@/api/modules/exam'
import { paperApi } from '@/api/modules/paper'
import { userApi } from '@/api/modules/user'

const list = ref<any[]>([])
const total = ref(0)
const loading = ref(false)
const papers = ref<any[]>([])
const students = ref<any[]>([])

const query = reactive({ page: 1, page_size: 10, keyword: '', status: -1 })
const currentPage = computed({
  get: () => query.page,
  set: (v: number) => { query.page = v }
})
const pageSize = computed({
  get: () => query.page_size,
  set: (v: number) => { query.page_size = v }
})

// 创建/编辑对话框
const dialogVisible = ref(false)
const dialogTitle = ref('新建考试')
const editingId = ref<number | null>(null)
const submitting = ref(false)
const formRef = ref<FormInstance>()

const form = reactive<any>({
  title: '',
  description: '',
  paper_id: null as number | null,
  start_time: '',
  end_time: '',
  duration: 60,
  max_attempts: 1,
  shuffle_q: true,
  shuffle_opt: true,
  anti_cheat: true,
  target_users: [] as number[],
})

const rules: FormRules = {
  title: [
    { required: true, message: '请输入考试名称', trigger: 'blur' },
    { min: 2, max: 128, message: '名称长度 2-128 字', trigger: 'blur' },
  ],
  paper_id: [
    { required: true, message: '请选择试卷', trigger: 'change' },
  ],
  start_time: [
    { required: true, message: '请选择开始时间', trigger: 'change' },
  ],
  end_time: [
    { required: true, message: '请选择结束时间', trigger: 'change' },
    {
      validator: (_: any, value: string, cb: any) => {
        if (value && form.start_time && new Date(value).getTime() <= new Date(form.start_time).getTime()) {
          cb(new Error('结束时间必须晚于开始时间'))
        } else {
          cb()
        }
      },
      trigger: 'change',
    },
  ],
  duration: [
    { required: true, message: '请设置考试时长', trigger: 'blur' },
    { type: 'number', min: 1, max: 600, message: '时长 1-600 分钟', trigger: 'blur' },
  ],
  max_attempts: [
    { type: 'number', min: 1, max: 10, message: '次数 1-10', trigger: 'blur' },
  ],
}

async function fetchList() {
  loading.value = true
  try {
    const params: any = { page: query.page, page_size: query.page_size }
    if (query.keyword) params.keyword = query.keyword
    if (query.status >= 0) params.status = query.status
    const data: any = await examApi.list(params)
    list.value = data?.list || []
    total.value = data?.total || 0
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

async function fetchPapers() {
  const data: any = await paperApi.list({ page: 1, page_size: 100 })
  papers.value = data?.list || []
}

async function fetchStudents() {
  try {
    const data: any = await userApi.list({ page: 1, page_size: 200, role: 3 })
    students.value = data?.list || []
  } catch {
    students.value = []
  }
}

function statusName(s: number) { return ['未发布', '进行中', '已结束'][s] || '未知' }
function statusType(s: number) { return ['danger', 'success', 'info'][s] || '' }

function formatTime(t?: string) { return t ? new Date(t).toLocaleString('zh-CN') : '-' }

function resetForm() {
  Object.assign(form, {
    title: '', description: '', paper_id: null,
    start_time: new Date().toISOString(),
    end_time: new Date(Date.now() + 86400000 * 7).toISOString(),
    duration: 60, max_attempts: 1,
    shuffle_q: true, shuffle_opt: true, anti_cheat: true,
    target_users: [],
  })
}

function openCreate() {
  dialogTitle.value = '新建考试'
  editingId.value = null
  resetForm()
  dialogVisible.value = true
  setTimeout(() => formRef.value?.clearValidate(), 0)
}

async function openEdit(row: any) {
  dialogTitle.value = '编辑考试'
  editingId.value = row.id
  try {
    const data: any = await examApi.detail(row.id)
    Object.assign(form, {
      title: data?.title || '',
      description: data?.description || '',
      paper_id: data?.paper_id || null,
      start_time: data?.start_time || '',
      end_time: data?.end_time || '',
      duration: data?.duration || 60,
      max_attempts: data?.max_attempts || 1,
      shuffle_q: data?.shuffle_q ?? true,
      shuffle_opt: data?.shuffle_opt ?? true,
      anti_cheat: data?.anti_cheat ?? true,
      target_users: typeof data?.target_users === 'string'
        ? (() => { try { return JSON.parse(data.target_users) || [] } catch { return [] } })()
        : (data?.target_users || []),
    })
    dialogVisible.value = true
    setTimeout(() => formRef.value?.clearValidate(), 0)
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
  }
}

async function onSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    ElMessage.warning('请检查必填项')
    return
  }
  submitting.value = true
  try {
    if (editingId.value) {
      await examApi.update(editingId.value, form)
      ElMessage.success('更新成功')
    } else {
      await examApi.create(form)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    submitting.value = false
  }
}

async function onDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      `确认删除考试 "${row.title}"？此操作不可恢复`,
      '删除确认',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消', confirmButtonClass: 'el-button--danger' }
    )
    await examApi.remove(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
  }
}

async function onArchive(row: any) {
  try {
    await ElMessageBox.confirm(
      `确认归档考试 "${row.title}"？归档后将无法再参加`,
      '归档确认',
      { type: 'warning' }
    )
    // 后端目前没有专门的 archive 接口，先用 update 把 status 设为 2
    await examApi.update(row.id, {
      ...row,
      status: 2,
      // 不传时间相关必填，避免报错
    })
    ElMessage.success('已归档')
    fetchList()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '操作失败')
  }
}

function viewRecords(row: any) {
  ElMessage.info(`考试 #${row.id} "${row.title}" 共 ${row.record_count || 0} 条记录。可在"考试记录"中查看详情`)
}

function handlePageChange(p: number) { query.page = p; fetchList() }
function handleSizeChange(s: number) { query.page_size = s; query.page = 1; fetchList() }
function handleSearch() { query.page = 1; fetchList() }
function handleResetSearch() {
  Object.assign(query, { keyword: '', status: -1, page: 1 })
  fetchList()
}

onMounted(() => { fetchPapers(); fetchStudents(); fetchList() })
</script>

<template>
  <div class="koala-page">
    <el-card>
      <!-- 搜索栏 -->
      <div class="filter">
        <el-input
          v-model="query.keyword"
          placeholder="搜索考试名称"
          clearable
          style="width: 220px;"
          @keyup.enter="handleSearch"
        />
        <el-select v-model="query.status" placeholder="状态" clearable style="width: 130px;" @change="handleSearch">
          <el-option :value="-1" label="全部" />
          <el-option :value="0" label="未发布" />
          <el-option :value="1" label="进行中" />
          <el-option :value="2" label="已结束" />
        </el-select>
        <el-button type="primary" @click="handleSearch">
          <el-icon><Search /></el-icon> 搜索
        </el-button>
        <el-button @click="handleResetSearch">重置</el-button>
        <el-button type="success" @click="openCreate">
          <el-icon><Plus /></el-icon> 新建考试
        </el-button>
      </div>

      <!-- 列表 -->
      <el-table :data="list" v-loading="loading" stripe border style="margin-top: 16px;">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="考试名称" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <strong>{{ row.title }}</strong>
          </template>
        </el-table-column>
        <el-table-column label="关联试卷" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag size="small" effect="plain">
              {{ papers.find((p: any) => p.id === row.paper_id)?.title || '#' + row.paper_id }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="开始时间" width="170">
          <template #default="{ row }">{{ formatTime(row.start_time) }}</template>
        </el-table-column>
        <el-table-column label="结束时间" width="170">
          <template #default="{ row }">{{ formatTime(row.end_time) }}</template>
        </el-table-column>
        <el-table-column prop="duration" label="时长(分)" width="100" align="center" />
        <el-table-column prop="max_attempts" label="次数" width="80" align="center" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" effect="plain">{{ statusName(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" link @click="openEdit(row)">
              <el-icon><Edit /></el-icon>编辑
            </el-button>
            <el-button size="small" type="warning" link @click="onArchive(row)" v-if="row.status !== 2">
              <el-icon><Folder /></el-icon>归档
            </el-button>
            <el-button size="small" type="info" link @click="viewRecords(row)">
              <el-icon><Document /></el-icon>记录
            </el-button>
            <el-button size="small" type="danger" link @click="onDelete(row)">
              <el-icon><Delete /></el-icon>删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="640"
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" label-position="right">
        <el-form-item label="考试名称" prop="title">
          <el-input v-model="form.title" maxlength="128" show-word-limit placeholder="2-128 字" />
        </el-form-item>
        <el-form-item label="考试描述">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="可选，描述考试内容" />
        </el-form-item>
        <el-form-item label="试卷" prop="paper_id">
          <el-select v-model="form.paper_id" placeholder="请选择试卷" style="width: 100%;" filterable>
            <el-option
              v-for="p in papers"
              :key="p.id"
              :value="p.id"
              :label="p.title"
            >
              <span>{{ p.title }}</span>
              <span style="float:right; color:#909399; font-size:12px;">
                {{ p.duration }}分 · {{ p.total_score }}分
              </span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="开始时间" prop="start_time">
          <el-date-picker
            v-model="form.start_time"
            type="datetime"
            placeholder="选择开始时间"
            style="width: 100%;"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
          />
        </el-form-item>
        <el-form-item label="结束时间" prop="end_time">
          <el-date-picker
            v-model="form.end_time"
            type="datetime"
            placeholder="选择结束时间"
            style="width: 100%;"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
          />
        </el-form-item>
        <el-form-item label="时长（分）" prop="duration">
          <el-input-number v-model="form.duration" :min="1" :max="600" />
          <span style="margin-left: 12px; color: #909399; font-size: 12px;">
            总时长 = 结束时间 - 开始时间
          </span>
        </el-form-item>
        <el-form-item label="最大次数" prop="max_attempts">
          <el-input-number v-model="form.max_attempts" :min="1" :max="10" />
        </el-form-item>
        <el-form-item label="防作弊">
          <el-checkbox v-model="form.shuffle_q">题目乱序</el-checkbox>
          <el-checkbox v-model="form.shuffle_opt">选项乱序</el-checkbox>
          <el-checkbox v-model="form.anti_cheat">切屏检测</el-checkbox>
        </el-form-item>
        <el-form-item label="指定学员">
          <el-select v-model="form.target_users" multiple filterable placeholder="不选则全员可参加" style="width: 100%;">
            <el-option
              v-for="s in students"
              :key="s.id"
              :value="s.id"
              :label="s.username + (s.nickname ? ' (' + s.nickname + ')' : '')"
            />
          </el-select>
          <span style="margin-left: 12px; color: #909399; font-size: 12px;">
            不指定则全部学员可参加
          </span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">
          {{ editingId ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.filter { display: flex; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; }
</style>
