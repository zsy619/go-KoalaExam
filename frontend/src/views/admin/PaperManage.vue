<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { paperApi } from '@/api/modules/paper'

const list = ref<any[]>([])
const total = ref(0)
const loading = ref(false)
const query = ref({ page: 1, page_size: 20, keyword: '' })

async function fetchList() {
  loading.value = true
  try {
    const { data } = await paperApi.list(query.value)
    list.value = data!.list || []
    total.value = data!.total
  } finally { loading.value = false }
}

const dialogVisible = ref(false)
const editing = ref<any>({ strategy: 1, duration: 60, total_score: 100 })

function openCreate() { editing.value = { strategy: 1, duration: 60, total_score: 100, question_ids: [] }; dialogVisible.value = true }

async function onSubmit() {
  await paperApi.create(editing.value)
  dialogVisible.value = false
  fetchList()
}

onMounted(fetchList)
</script>

<template>
  <div class="koala-page">
    <el-card>
      <div class="filter">
        <el-input v-model="query.keyword" placeholder="试卷名称" clearable style="width:200px" @keyup.enter="fetchList" />
        <el-button type="primary" @click="fetchList">搜索</el-button>
        <el-button type="success" @click="openCreate">+ 新建试卷</el-button>
      </div>
      <el-table :data="list" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="标题" />
        <el-table-column label="组卷策略">
          <template #default="{ row }">
            <el-tag>{{ ['固定', '随机', '遗传算法'][row.strategy] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="total_score" label="总分" width="100" />
        <el-table-column prop="duration" label="时长（分）" width="120" />
        <el-table-column prop="pass_score" label="及格分" width="100" />
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" title="新建试卷" width="600">
      <el-form :model="editing" label-width="100px">
        <el-form-item label="标题"><el-input v-model="editing.title" /></el-form-item>
        <el-form-item label="策略">
          <el-radio-group v-model="editing.strategy">
            <el-radio :value="1">固定</el-radio>
            <el-radio :value="2">随机</el-radio>
            <el-radio :value="3">遗传算法</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="时长"><el-input-number v-model="editing.duration" :min="10" :max="300" /></el-form-item>
        <el-form-item label="总分"><el-input-number v-model="editing.total_score" :min="1" :max="200" /></el-form-item>
        <el-form-item label="及格分"><el-input-number v-model="editing.pass_score" :min="0" :max="200" /></el-form-item>
        <el-form-item v-if="editing.strategy === 1" label="题目IDs（逗号分隔）">
          <el-input v-model="editing.question_ids_text" placeholder="1,2,3" @change="(v: any) => editing.question_ids = (v || '').split(',').map((x: string) => Number(x.trim()))" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="onSubmit">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>
