<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { userApi } from '@/api/modules/user'

const list = ref<any[]>([])
const total = ref(0)
const loading = ref(false)
const query = ref({ page: 1, page_size: 20, role: 0, keyword: '' })

async function fetchList() {
  loading.value = true
  try {
    const { data } = await userApi.list(query.value)
    list.value = data!.list || []
    total.value = data!.total
  } finally {
    loading.value = false
  }
}

function resetPwd(id: number) {
  ElMessageBox.confirm('确定重置该用户的密码？', '提示', { type: 'warning' })
    .then(async () => {
      const { data } = await userApi.resetPwd(id)
      ElMessage.success(`新密码：${data!.new_password}`)
    })
}

function roleText(r: number) { return ['', '超管', '教师', '学生'][r] || '未知' }

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
        <el-button type="primary" @click="fetchList">搜索</el-button>
      </div>
      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="账号" />
        <el-table-column prop="nickname" label="昵称" />
        <el-table-column label="角色">
          <template #default="{ row }">
            <el-tag>{{ roleText(row.role) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="160">
          <template #default="{ row }">
            <el-button size="small" type="danger" @click="resetPwd(row.id)">重置密码</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.page_size"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="fetchList"
      />
    </el-card>
  </div>
</template>

<style scoped>
.filter { display: flex; gap: 12px; margin-bottom: 16px; }
</style>
