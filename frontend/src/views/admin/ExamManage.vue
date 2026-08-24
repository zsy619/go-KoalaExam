<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { examApi } from '@/api/modules/exam'
import { paperApi } from '@/api/modules/paper'

const list = ref<any[]>([])
const total = ref(0)
const papers = ref<any[]>([])
const dialogVisible = ref(false)
const editing = ref<any>({ duration: 60, shuffle_q: true, shuffle_opt: true, anti_cheat: true })

async function fetchList() {
  const { data } = await examApi.list({ page: 1, page_size: 50 })
  list.value = data!.list || []
  total.value = data!.total
}

async function fetchPapers() {
  const { data } = await paperApi.list({ page: 1, page_size: 100 })
  papers.value = data!.list || []
}

function openCreate() {
  editing.value = {
    duration: 60,
    shuffle_q: true,
    shuffle_opt: true,
    anti_cheat: true,
    start_time: new Date().toISOString(),
    end_time: new Date(Date.now() + 86400000 * 7).toISOString(),
  }
  dialogVisible.value = true
}

async function onSubmit() {
  await examApi.create(editing.value)
  dialogVisible.value = false
  fetchList()
}

onMounted(() => { fetchPapers(); fetchList() })
</script>

<template>
  <div class="koala-page">
    <el-card>
      <el-button type="success" @click="openCreate">+ 新建考试</el-button>
      <el-table :data="list" style="margin-top:16px">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="考试名称" />
        <el-table-column prop="start_time" label="开始时间" width="180" />
        <el-table-column prop="end_time" label="结束时间" width="180" />
        <el-table-column prop="duration" label="时长（分）" width="120" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="['','success','info'][row.status]">{{ ['未发布','进行中','已结束'][row.status] }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" title="新建考试" width="640">
      <el-form :model="editing" label-width="100px">
        <el-form-item label="标题"><el-input v-model="editing.title" /></el-form-item>
        <el-form-item label="试卷">
          <el-select v-model="editing.paper_id">
            <el-option v-for="p in papers" :key="p.id" :value="p.id" :label="p.title" />
          </el-select>
        </el-form-item>
        <el-form-item label="开始时间"><el-input v-model="editing.start_time" /></el-form-item>
        <el-form-item label="结束时间"><el-input v-model="editing.end_time" /></el-form-item>
        <el-form-item label="时长（分）"><el-input-number v-model="editing.duration" :min="1" :max="600" /></el-form-item>
        <el-form-item label="防作弊">
          <el-checkbox v-model="editing.shuffle_q">题目乱序</el-checkbox>
          <el-checkbox v-model="editing.shuffle_opt">选项乱序</el-checkbox>
          <el-checkbox v-model="editing.anti_cheat">切屏检测</el-checkbox>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="onSubmit">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>
