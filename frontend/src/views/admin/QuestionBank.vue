<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { questionApi } from '@/api/modules/question'
import FavoriteStar from '@/components/business/FavoriteStar.vue'

const list = ref<any[]>([])
const total = ref(0)
const loading = ref(false)
const categories = ref<any[]>([])
const query = ref({ page: 1, page_size: 20, category_id: 0, type: 0, difficulty: 0, keyword: '' })

async function fetchList() {
  loading.value = true
  try {
    const { data } = await questionApi.list(query.value)
    list.value = data!.list || []
    total.value = data!.total
  } finally { loading.value = false }
}

async function fetchCategories() {
  const { data } = await questionApi.categories()
  categories.value = data || []
}

const dialogVisible = ref(false)
const editing = ref<any>({})

function openCreate() {
  editing.value = { type: 1, difficulty: 2, options: [{ key: 'A', text: '' }, { key: 'B', text: '' }, { key: 'C', text: '' }, { key: 'D', text: '' }], answer: '' }
  dialogVisible.value = true
}

async function onSubmit() {
  if (editing.value.id) {
    await questionApi.update(editing.value.id, editing.value)
  } else {
    await questionApi.create(editing.value)
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  fetchList()
}

function onRemove(id: number) {
  ElMessageBox.confirm('确定删除该题目？', '提示', { type: 'warning' })
    .then(async () => { await questionApi.remove(id); fetchList() })
}

onMounted(() => { fetchCategories(); fetchList() })
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
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="题型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ ['','单选','多选','判断','填空','不定项','编程'][row.type] }}</el-tag>
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
            <el-button size="small" @click="editing = row; dialogVisible = true">编辑</el-button>
            <el-button size="small" type="danger" @click="onRemove(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" title="编辑题目" width="700">
      <el-form :model="editing" label-width="80px">
        <el-form-item label="分类">
          <el-select v-model="editing.category_id">
            <el-option v-for="c in categories" :key="c.id" :value="c.id" :label="c.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="题型">
          <el-select v-model="editing.type">
            <el-option :value="1" label="单选" />
            <el-option :value="2" label="多选" />
            <el-option :value="3" label="判断" />
            <el-option :value="4" label="填空" />
          </el-select>
        </el-form-item>
        <el-form-item label="难度">
          <el-rate v-model="editing.difficulty" :max="3" />
        </el-form-item>
        <el-form-item label="题干">
          <el-input v-model="editing.title" type="textarea" :rows="3" />
        </el-form-item>
        <template v-if="editing.type !== 3 && editing.type !== 4">
          <el-form-item v-for="opt in editing.options" :key="opt.key" :label="`选项 ${opt.key}`">
            <el-input v-model="opt.text" />
          </el-form-item>
        </template>
        <el-form-item label="答案">
          <el-input v-model="editing.answer" placeholder="多选用逗号分隔，如 A,C" />
        </el-form-item>
        <el-form-item label="解析">
          <el-input v-model="editing.analysis" type="textarea" :rows="2" />
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
.filter { display: flex; gap: 12px; margin-bottom: 16px; }
</style>
