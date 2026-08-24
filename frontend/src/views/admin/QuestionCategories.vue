<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  Plus, Search, Refresh, Folder, FolderOpened, Edit, Delete,
  Connection, Document, Promotion
} from '@element-plus/icons-vue'
import { questionApi, type Category } from '@/api/modules/question'

// ===== 状态 =====
const loading = ref(false)
const submitting = ref(false)
const list = ref<Category[]>([])
const treeRef = ref()

const dialogVisible = ref(false)
const editing = ref<Category | null>(null)
const formRef = ref<FormInstance>()
const form = reactive<any>({
  id: 0,
  name: '',
  code: '',
  parent_id: 0,
  sort: 0,
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入分类名称', trigger: 'blur' },
    { min: 1, max: 64, message: '名称长度 1-64 字', trigger: 'blur' },
  ],
  code: [
    { required: true, message: '请输入分类编码', trigger: 'blur' },
    { pattern: /^[a-z0-9_]+$/i, message: '编码只能包含字母、数字、下划线', trigger: 'blur' },
    { max: 32, message: '编码最多 32 位', trigger: 'blur' },
  ],
  sort: [
    { type: 'number', min: 0, max: 999, message: '排序 0-999', trigger: 'blur' },
  ],
}

const keyword = ref('')
const expandAll = ref(true)

// 题目数统计
const questionCountMap = reactive<Record<number, number>>({})

async function loadQuestionCounts() {
  // 为每个顶级分类异步加载题目数（懒加载，不阻塞渲染）
  for (const c of list.value) {
    try {
      const data: any = await questionApi.list({ category_id: c.id, page: 1, page_size: 1 })
      questionCountMap[c.id] = data?.total || 0
    } catch {}
  }
}

// 树形结构
const tree = computed(() => buildTree(list.value))

// 过滤树（关键词搜索）
const filteredTree = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return tree.value
  return filterTree(tree.value, (c) =>
    c.name.toLowerCase().includes(kw) ||
    (c.code || '').toLowerCase().includes(kw)
  )
})

// 父级选择树（排除当前编辑节点及其子节点）
const parentTree = computed(() => {
  if (!editing.value) return tree.value
  const excludeIds = new Set<number>([editing.value.id])
  const collectExclude = (cs: Category[]) => {
    cs?.forEach((c) => {
      if (c.children) {
        if (c.id === editing.value!.id) {
          c.children.forEach((cc) => excludeIds.add(cc.id))
        }
        collectExclude(c.children)
      }
    })
  }
  collectExclude(tree.value)
  return filterTree(tree.value, (c) => !excludeIds.has(c.id))
})

function buildTree(items: Category[]): (Category & { children: Category[] })[] {
  const map = new Map<number, Category & { children: Category[] }>()
  items.forEach((it) => map.set(it.id, { ...it, children: [] }))
  const roots: (Category & { children: Category[] })[] = []
  map.forEach((node) => {
    if (node.parent_id && map.has(node.parent_id)) {
      map.get(node.parent_id)!.children.push(node)
    } else {
      roots.push(node)
    }
  })
  // 按 sort 排序
  const sortTree = (ns: any[]) => {
    ns.sort((a, b) => (a.sort || 0) - (b.sort || 0))
    ns.forEach((n) => n.children && sortTree(n.children))
  }
  sortTree(roots)
  return roots
}

function filterTree(items: Category[], predicate: (c: Category) => boolean): any[] {
  const result: any[] = []
  for (const item of items) {
    if (predicate(item)) {
      result.push({ ...item, children: item.children ? filterTree(item.children, predicate) : [] })
    } else if (item.children && item.children.length > 0) {
      const filtered = filterTree(item.children, predicate)
      if (filtered.length > 0) {
        result.push({ ...item, children: filtered })
      }
    }
  }
  return result
}

// ===== 数据加载 =====
async function loadList() {
  loading.value = true
  try {
    const data: any = await questionApi.listCategories()
    list.value = Array.isArray(data) ? data : []
    await nextTick()
    // 默认展开所有节点
    expandAll.value && treeRef.value?.expandAll?.()
    // 异步加载题目数（不阻塞）
    loadQuestionCounts()
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
    list.value = []
  } finally {
    loading.value = false
  }
}

// ===== 统计 =====
const totalCategories = computed(() => list.value.length)
const rootCategories = computed(() => list.value.filter(c => !c.parent_id).length)
const subCategories = computed(() => list.value.filter(c => c.parent_id).length)
const totalQuestions = computed(() =>
  Object.values(questionCountMap).reduce((s, n) => s + n, 0)
)

// ===== 操作 =====
function onCreate(parent: Category | null) {
  editing.value = null
  Object.assign(form, {
    id: 0,
    name: '',
    code: '',
    parent_id: parent?.id || 0,
    sort: (list.value.filter(c => c.parent_id === (parent?.id || 0)).length) + 1,
  })
  dialogVisible.value = true
  setTimeout(() => formRef.value?.clearValidate(), 0)
}

function onEdit(row: Category) {
  editing.value = row
  Object.assign(form, { ...row })
  dialogVisible.value = true
  setTimeout(() => formRef.value?.clearValidate(), 0)
}

async function onDelete(row: Category) {
  const children = list.value.filter(c => c.parent_id === row.id)
  if (children.length > 0) {
    ElMessage.warning(`"${row.name}" 下还有 ${children.length} 个子分类，请先删除子分类`)
    return
  }
  const count = questionCountMap[row.id]
  const msg = count && count > 0
    ? `"${row.name}" 下还有 ${count} 道题目，删除后题目将变为无分类状态，确定删除吗？`
    : `确认删除分类 "${row.name}"？`

  try {
    await ElMessageBox.confirm(msg, '提示', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      confirmButtonClass: 'el-button--danger'
    })
    await questionApi.deleteCategory(row.id)
    ElMessage.success('已删除')
    await loadList()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
  }
}

async function onSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  submitting.value = true
  try {
    if (editing.value) {
      await questionApi.updateCategory(editing.value.id, form)
      ElMessage.success('已更新')
    } else {
      await questionApi.createCategory(form)
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

// 全部展开/折叠
function toggleExpandAll() {
  expandAll.value = !expandAll.value
  if (expandAll.value) {
    treeRef.value?.expandAll?.()
  } else {
    treeRef.value?.collapseAll?.()
  }
}

// 展开到指定节点
function focusKeyword() {
  if (!keyword.value.trim()) return
  const nodes = treeRef.value?.getNodeKey?.()
  // 让 Element Plus 自动展开包含匹配项的路径
  treeRef.value?.filter?.(keyword.value)
}

watch(keyword, () => {
  if (keyword.value.trim()) {
    nextTick(() => treeRef.value?.filter?.(keyword.value))
  } else {
    nextTick(() => treeRef.value?.filter?.(''))
  }
})

onMounted(loadList)
</script>

<template>
  <div class="category-page">
    <!-- 顶部统计 -->
    <div class="stats-row">
      <div class="stat-card stat-total">
        <div class="stat-icon"><Folder /></div>
        <div class="stat-info">
          <div class="stat-value">{{ totalCategories }}</div>
          <div class="stat-label">总分类数</div>
        </div>
      </div>
      <div class="stat-card stat-root">
        <div class="stat-icon"><FolderOpened /></div>
        <div class="stat-info">
          <div class="stat-value">{{ rootCategories }}</div>
          <div class="stat-label">顶级分类</div>
        </div>
      </div>
      <div class="stat-card stat-sub">
        <div class="stat-icon"><Connection /></div>
        <div class="stat-info">
          <div class="stat-value">{{ subCategories }}</div>
          <div class="stat-label">子分类</div>
        </div>
      </div>
      <div class="stat-card stat-question">
        <div class="stat-icon"><Document /></div>
        <div class="stat-info">
          <div class="stat-value">{{ totalQuestions }}</div>
          <div class="stat-label">题目总数</div>
        </div>
      </div>
    </div>

    <el-card>
      <template #header>
        <div class="header">
          <span class="title"><el-icon><Folder /></el-icon> 题目分类管理</span>
          <div class="header-actions">
            <el-input
              v-model="keyword"
              placeholder="搜索分类名称/编码"
              clearable
              style="width: 240px;"
            >
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
            <el-button @click="toggleExpandAll">
              <el-icon><Promotion /></el-icon>
              {{ expandAll ? '全部折叠' : '全部展开' }}
            </el-button>
            <el-button @click="loadList">
              <el-icon><Refresh /></el-icon> 刷新
            </el-button>
            <el-button type="primary" @click="onCreate(null)">
              <el-icon><Plus /></el-icon> 新增顶级
            </el-button>
          </div>
        </div>
      </template>

      <!-- el-tree 树形展示 -->
      <div class="tree-wrapper" v-loading="loading">
        <el-empty v-if="!loading && filteredTree.length === 0" description="暂无分类" />

        <el-tree
          v-else
          ref="treeRef"
          :data="filteredTree"
          node-key="id"
          :props="{ label: 'name', children: 'children' }"
          :expand-on-click-node="false"
          :default-expand-all="expandAll"
          empty-text="暂无分类"
          class="category-tree"
        >
          <template #default="{ node, data }">
            <div class="tree-node">
              <!-- 左侧：图标 + 名称 + tag -->
              <div class="node-left">
                <el-icon class="node-icon" :size="16">
                  <FolderOpened v-if="data.children && data.children.length" />
                  <Folder v-else />
                </el-icon>
                <span class="node-name">{{ data.name }}</span>
                <el-tag v-if="data.children && data.children.length" size="small" effect="plain" type="info" round>
                  {{ data.children.length }}
                </el-tag>
                <el-tag v-if="data.code" size="small" effect="dark" type="primary">{{ data.code }}</el-tag>
              </div>

              <!-- 中间：排序 + 题目数 -->
              <div class="node-meta">
                <el-tooltip content="排序值" placement="top">
                  <span class="sort-tag">#{{ data.sort }}</span>
                </el-tooltip>
                <el-tooltip content="该分类下的题目数" placement="top">
                  <el-tag
                    size="small"
                    effect="plain"
                    :type="(questionCountMap[data.id] || 0) > 0 ? 'success' : 'info'"
                    class="count-tag"
                  >
                    <el-icon><Document /></el-icon>
                    {{ questionCountMap[data.id] ?? '-' }}
                  </el-tag>
                </el-tooltip>
              </div>

              <!-- 右侧：操作按钮 -->
              <div class="node-actions">
                <el-button type="primary" link size="small" @click.stop="onCreate(data)">
                  <el-icon><Plus /></el-icon>子类
                </el-button>
                <el-button type="warning" link size="small" @click.stop="onEdit(data)">
                  <el-icon><Edit /></el-icon>编辑
                </el-button>
                <el-button type="danger" link size="small" @click.stop="onDelete(data)">
                  <el-icon><Delete /></el-icon>删除
                </el-button>
              </div>
            </div>
          </template>
        </el-tree>
      </div>
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="editing ? '编辑分类' : '新增分类'" width="520" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" label-position="right">
        <el-form-item label="父级分类" prop="parent_id">
          <el-tree-select
            v-model="form.parent_id"
            :data="parentTree"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            check-strictly
            clearable
            placeholder="不选则为顶级分类"
            style="width: 100%;"
          />
        </el-form-item>
        <el-form-item label="分类名称" prop="name">
          <el-input v-model="form.name" maxlength="64" show-word-limit placeholder="如：计算机基础" />
        </el-form-item>
        <el-form-item label="编码" prop="code">
          <el-input v-model="form.code" placeholder="如 CS-BASIC" maxlength="32" show-word-limit />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" :max="999" />
          <span style="margin-left:12px;color:#909399;font-size:12px;">数字小靠前</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">
          {{ editing ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.category-page { padding: 16px; }

/* ============ 统计卡片 ============ */
.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}
.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px 24px;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
  border-left: 4px solid #409eff;
  transition: all 0.3s;
}
.stat-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 6px 20px rgba(0,0,0,0.1);
}
.stat-card.stat-root { border-left-color: #67c23a; }
.stat-card.stat-sub { border-left-color: #e6a23c; }
.stat-card.stat-question { border-left-color: #909399; }
.stat-icon {
  font-size: 36px;
  color: #409eff;
}
.stat-root .stat-icon { color: #67c23a; }
.stat-sub .stat-icon { color: #e6a23c; }
.stat-question .stat-icon { color: #909399; }
.stat-info { flex: 1; }
.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
  line-height: 1;
}
.stat-label {
  font-size: 13px;
  color: #909399;
  margin-top: 4px;
}

/* ============ 头部 ============ */
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 6px;
}
.header-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

/* ============ 树形展示 ============ */
.tree-wrapper {
  min-height: 200px;
}
.category-tree {
  background: transparent;
}

.category-tree :deep(.el-tree-node__content) {
  height: 52px;
  border-radius: 6px;
  transition: background 0.2s;
}
.category-tree :deep(.el-tree-node__content:hover) {
  background: #f5f7fa;
}
.category-tree :deep(.el-tree-node.is-current > .el-tree-node__content) {
  background: #ecf5ff;
}

.category-tree :deep(.el-tree-node__expand-icon) {
  font-size: 16px;
}

/* 自定义节点 */
.tree-node {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0 8px;
  gap: 16px;
}
.node-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}
.node-icon {
  color: #e6a23c;
  flex-shrink: 0;
}
.node-name {
  font-weight: 500;
  color: #303133;
  font-size: 14px;
}

.node-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}
.sort-tag {
  font-size: 12px;
  color: #909399;
  font-family: 'Courier New', monospace;
  padding: 2px 6px;
  background: #f4f4f5;
  border-radius: 3px;
}
.count-tag {
  min-width: 60px;
  justify-content: center;
}

.node-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
  opacity: 0.6;
  transition: opacity 0.2s;
}
.tree-node:hover .node-actions {
  opacity: 1;
}

/* 节点连线 */
.category-tree :deep(.el-tree-node) {
  position: relative;
}
.category-tree :deep(.el-tree-node__children) {
  padding-left: 20px;
}
</style>
