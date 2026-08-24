<template>
  <div class="category-page">
    <el-card>
      <template #header>
        <div class="header">
          <span>题目分类管理</span>
          <el-button type="primary" @click="onCreate(null)">
            <el-icon><Plus /></el-icon>新增分类
          </el-button>
        </div>
      </template>

      <el-table :data="tree" row-key="id" :tree-props="{ children: 'children' }" v-loading="loading">
        <el-table-column prop="name" label="分类名称" min-width="200" />
        <el-table-column prop="code" label="编码" width="120" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="操作" width="220">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="onCreate(row)">添加子分类</el-button>
            <el-button type="warning" link size="small" @click="onEdit(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="editing ? '编辑分类' : '新增分类'" width="500px">
      <el-form :model="form" label-width="100px" :rules="rules" ref="formRef">
        <el-form-item label="父级分类" prop="parent_id">
          <el-tree-select
            v-model="form.parent_id"
            :data="parentTree"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            check-strictly
            clearable
            placeholder="不选则为顶级"
            style="width: 100%;"
          />
        </el-form-item>
        <el-form-item label="分类名称" prop="name">
          <el-input v-model="form.name" maxlength="64" show-word-limit />
        </el-form-item>
        <el-form-item label="编码" prop="code">
          <el-input v-model="form.code" placeholder="可选，如 FE/BE/ALGO" maxlength="64" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" :max="999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="onSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { questionApi } from '@/api/modules/question'

interface Category {
  id: number
  parent_id: number
  name: string
  code?: string
  sort: number
  children?: Category[]
}

const loading = ref(false)
const submitting = ref(false)
const list = ref<Category[]>([])
const dialogVisible = ref(false)
const editing = ref<Category | null>(null)
const formRef = ref()
const form = ref<Category>({
  id: 0,
  name: '',
  code: '',
  parent_id: 0,
  sort: 0
})

const rules = {
  name: [
    { required: true, message: '请输入分类名称', trigger: 'blur' },
    { min: 1, max: 64, message: '名称长度 1-64 字', trigger: 'blur' },
  ],
  code: [
    { required: true, message: '请输入分类编码', trigger: 'blur' },
    { pattern: /^[a-z0-9_]+$/i, message: '编码只能包含字母、数字、下划线', trigger: 'blur' },
    { max: 32, message: '编码最多 32 位', trigger: 'blur' },
  ],
}

// 树形结构
const tree = computed(() => buildTree(list.value))

// 父级选择树（排除当前编辑节点及其子节点）
const parentTree = computed(() => {
  if (!editing.value) return tree.value
  const excludeIds = new Set<number>([editing.value.id])
  const collectExclude = (cs: Category[]) => {
    cs?.forEach((c) => {
      if (c.id === editing.value!.id) {
        c.children?.forEach((cc) => excludeIds.add(cc.id))
      } else if (c.children) {
        collectExclude(c.children)
      }
    })
  }
  collectExclude(tree.value)
  return filterTree(tree.value, (c) => !excludeIds.has(c.id))
})

function buildTree(items: Category[]): Category[] {
  const map = new Map<number, Category>()
  items.forEach((it) => map.set(it.id, { ...it, children: [] }))
  const roots: Category[] = []
  map.forEach((node) => {
    if (node.parent_id && map.has(node.parent_id)) {
      map.get(node.parent_id)!.children!.push(node)
    } else {
      roots.push(node)
    }
  })
  return roots
}

function filterTree(items: Category[], predicate: (c: Category) => boolean): Category[] {
  return items.filter(predicate).map((c) => ({
    ...c,
    children: c.children ? filterTree(c.children, predicate) : []
  }))
}

async function loadList() {
  loading.value = true
  try {
    const list = await questionApi.listCategories()
    list.value = Array.isArray(list) ? list : []
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

function onCreate(parent: Category | null) {
  editing.value = null
  form.value = {
    id: 0,
    name: '',
    code: '',
    parent_id: parent?.id || 0,
    sort: 0
  }
  dialogVisible.value = true
}

function onEdit(row: Category) {
  editing.value = row
  form.value = { ...row }
  dialogVisible.value = true
}

async function onDelete(row: Category) {
  try {
    await ElMessageBox.confirm(`确定删除分类"${row.name}"？`, '提示', { type: 'warning' })
    await questionApi.deleteCategory(row.id)
    ElMessage.success('已删除')
    await loadList()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
  }
}

async function onSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    if (editing.value) {
      await questionApi.updateCategory(editing.value.id, form.value)
      ElMessage.success('已更新')
    } else {
      await questionApi.createCategory(form.value)
      ElMessage.success('已创建')
    }
    dialogVisible.value = false
    await loadList()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

onMounted(loadList)
</script>

<style scoped lang="scss">
.category-page { padding: 16px; }
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
