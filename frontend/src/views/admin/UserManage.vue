<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { userApi } from '@/api/modules/user'

const list = ref<any[]>([])
const total = ref(0)
const loading = ref(false)
const query = ref({ page: 1, page_size: 20, role: 0, keyword: '' })

// 创建/编辑对话框
const dialogVisible = ref(false)
const dialogTitle = ref('新建用户')
const editingId = ref<number | null>(null)
const submitting = ref(false)
const formRef = ref<FormInstance>()

const form = reactive<any>({
  username: '',
  nickname: '',
  password: '',
  email: '',
  phone: '',
  role: 3,
  status: 1,
})

const rules: FormRules = {
  username: [
    { required: true, message: '请输入账号', trigger: 'blur' },
    { min: 3, max: 32, message: '账号长度 3-32 位', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_]+$/, message: '账号只能包含字母、数字、下划线', trigger: 'blur' },
  ],
  nickname: [
    { required: true, message: '请输入昵称', trigger: 'blur' },
    { min: 2, max: 32, message: '昵称长度 2-32 位', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 64, message: '密码长度 6-64 位', trigger: 'blur' },
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
  ],
  phone: [
    { pattern: /^$|^1[3-9]\d{9}$/, message: '手机号格式不正确', trigger: 'blur' },
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' },
  ],
}

async function fetchList() {
  loading.value = true
  try {
    const data: any = await userApi.list(query.value)
    list.value = data?.list || []
    total.value = data?.total || 0
  } finally {
    loading.value = false
  }
}

function resetPwd(id: number) {
  ElMessageBox.confirm('确定重置该用户的密码？', '提示', { type: 'warning' })
    .then(async () => {
      const data: any = await userApi.resetPwd(id)
      ElMessage.success(`新密码：${data?.new_password}`)
    })
    .catch(() => {})
}

function roleText(r: number) { return ['', '超管', '教师', '学生'][r] || '未知' }

function openCreate() {
  dialogTitle.value = '新建用户'
  editingId.value = null
  Object.assign(form, {
    username: '', nickname: '', password: '', email: '', phone: '', role: 3, status: 1,
  })
  dialogVisible.value = true
  // 清除校验状态
  setTimeout(() => formRef.value?.clearValidate(), 0)
}

async function openEdit(row: any) {
  dialogTitle.value = '编辑用户'
  editingId.value = row.id
  Object.assign(form, {
    username: row.username || '',
    nickname: row.nickname || '',
    password: '', // 编辑时不显示密码
    email: row.email || '',
    phone: row.phone || '',
    role: row.role || 3,
    status: row.status ?? 1,
  })
  dialogVisible.value = true
  setTimeout(() => formRef.value?.clearValidate(), 0)
}

async function onSubmit() {
  if (!formRef.value) return
  await formRef.value.validate()
  submitting.value = true
  try {
    if (editingId.value) {
      // 编辑：不传密码
      const { password, ...payload } = form
      await userApi.update(editingId.value, payload)
      ElMessage.success('更新成功')
    } else {
      await userApi.create(form)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (e: any) {
    if (e?.message) ElMessage.error(e.message)
  } finally {
    submitting.value = false
  }
}

async function onDelete(row: any) {
  try {
    await ElMessageBox.confirm(`确认删除用户 "${row.username}"？此操作不可恢复`, '删除确认', {
      type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消', confirmButtonClass: 'el-button--danger',
    })
    await userApi.remove(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
  }
}

onMounted(fetchList)
</script>

<template>
  <div class="koala-page">
    <el-card>
      <div class="filter">
        <el-input v-model="query.keyword" placeholder="账号/昵称" style="width:200px" clearable @keyup.enter="fetchList" />
        <el-select v-model="query.role" placeholder="角色" clearable style="width:140px">
          <el-option :value="1" label="超管" />
          <el-option :value="2" label="教师" />
          <el-option :value="3" label="学生" />
        </el-select>
        <el-button type="primary" @click="fetchList">
          <el-icon><Search /></el-icon> 搜索
        </el-button>
        <el-button type="success" @click="openCreate">
          <el-icon><Plus /></el-icon> 新建用户
        </el-button>
      </div>

      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="账号" />
        <el-table-column prop="nickname" label="昵称" />
        <el-table-column label="角色" width="100">
          <template #default="{ row }">
            <el-tag>{{ roleText(row.role) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="email" label="邮箱" show-overflow-tooltip />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" effect="plain">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">{{ row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" link @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="warning" link @click="resetPwd(row.id)">重置密码</el-button>
            <el-button size="small" type="danger" link @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.page_size"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @current-change="fetchList"
          @size-change="fetchList"
        />
      </div>
    </el-card>

    <!-- 创建/编辑用户对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="560"
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" label-position="right">
        <el-form-item label="账号" prop="username">
          <el-input v-model="form.username" placeholder="3-32位字母/数字/下划线" maxlength="32" show-word-limit :disabled="!!editingId" />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="form.nickname" placeholder="2-32位" maxlength="32" show-word-limit />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!editingId">
          <el-input v-model="form.password" type="password" placeholder="6-64位" show-password />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-radio-group v-model="form.role">
            <el-radio :value="1">超管</el-radio>
            <el-radio :value="2">教师</el-radio>
            <el-radio :value="3">学生</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="user@example.com" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" placeholder="可选，11位手机号" maxlength="11" />
        </el-form-item>
        <el-form-item label="状态" prop="status" v-if="editingId">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
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
