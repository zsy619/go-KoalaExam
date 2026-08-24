<template>
  <div class="favorites-page">
    <el-row :gutter="16">
      <el-col :span="6">
        <el-card class="folder-list" shadow="hover">
          <template #header>
            <div class="folder-header">
              <span>收藏夹</span>
              <el-button type="primary" link size="small" @click="showCreateDialog = true">
                <el-icon><Plus /></el-icon>新建
              </el-button>
            </div>
          </template>
          <div class="folder-item" :class="{ active: currentFolder === 0 }" @click="selectFolder(0)">
            <el-icon><FolderOpened /></el-icon>
            <span>全部收藏</span>
            <el-tag size="small">{{ stats.total }}</el-tag>
          </div>
          <div class="folder-item" :class="{ active: currentFolder === -1 }" @click="selectFolder(-1)">
            <el-icon><StarFilled /></el-icon>
            <span>错题本</span>
          </div>
          <el-divider />
          <div
            v-for="folder in folders"
            :key="folder.id"
            class="folder-item"
            :class="{ active: currentFolder === folder.id }"
            @click="selectFolder(folder.id)"
          >
            <el-icon><Folder /></el-icon>
            <span>{{ folder.name }}</span>
            <el-button
              type="danger"
              link
              size="small"
              @click.stop="onDeleteFolder(folder.id)"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </el-card>
      </el-col>

      <el-col :span="18">
        <el-card>
          <template #header>
            <div class="content-header">
              <span>{{ currentFolderName }} ({{ filteredFavorites.length }})</span>
              <div class="header-actions">
                <el-select v-model="filterType" placeholder="类型" clearable size="small" style="width: 100px;">
                  <el-option label="题目" :value="1" />
                  <el-option label="试卷" :value="2" />
                </el-select>
                <el-input v-model="searchText" placeholder="搜索题目..." clearable size="small" style="width: 200px;" />
                <el-button @click="loadFavorites" size="small">刷新</el-button>
              </div>
            </div>
          </template>

          <el-empty v-if="filteredFavorites.length === 0" description="暂无收藏" />

          <el-table v-else :data="filteredFavorites" @selection-change="onSelectionChange">
            <el-table-column type="selection" width="50" />
            <el-table-column label="类型" width="80">
              <template #default="{ row }">
                <el-tag :type="row.target_type === 1 ? 'primary' : 'success'" size="small">
                  {{ row.target_type === 1 ? '题目' : '试卷' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="标题" min-width="200">
              <template #default="{ row }">
                <el-link type="primary" @click="onPreview(row)">{{ getTitle(row) }}</el-link>
              </template>
            </el-table-column>
            <el-table-column label="收藏时间" width="180">
              <template #default="{ row }">
                {{ formatTime(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-button type="danger" link size="small" @click="onRemove(row)">移除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div v-if="selectedRows.length > 0" class="batch-bar">
            <span>已选 {{ selectedRows.length }} 项</span>
            <el-button type="danger" size="small" @click="onBatchRemove">批量移除</el-button>
            <el-button size="small" @click="onBatchMove">移动到</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 新建收藏夹对话框 -->
    <el-dialog v-model="showCreateDialog" title="新建收藏夹" width="400px">
      <el-form :model="newFolder" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="newFolder.name" placeholder="例如：高频考点" maxlength="32" />
        </el-form-item>
        <el-form-item label="颜色">
          <el-color-picker v-model="newFolder.color" />
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="newFolder.icon" placeholder="emoji 或字符" maxlength="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="onCreateFolder">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, FolderOpened, Folder, StarFilled, Delete } from '@element-plus/icons-vue'
import { useFavorite } from '@/composables/useFavorite'

const {
  favorites, folders, stats,
  loadFavorites, loadFolders, createFolder: createF, deleteFolder: delF,
  toggle
} = useFavorite()

const currentFolder = ref(0)
const filterType = ref<number | undefined>(undefined)
const searchText = ref('')
const selectedRows = ref<any[]>([])
const showCreateDialog = ref(false)
const newFolder = ref({ name: '', color: '#409eff', icon: '' })

const currentFolderName = computed(() => {
  if (currentFolder.value === 0) return '全部收藏'
  if (currentFolder.value === -1) return '错题本'
  const f = folders.value.find((x: any) => x.id === currentFolder.value)
  return f ? f.name : '未知'
})

const filteredFavorites = computed(() => {
  let list = favorites.value
  if (currentFolder.value > 0) {
    list = list.filter((f: any) => f.folder_id === currentFolder.value)
  }
  if (filterType.value) {
    list = list.filter((f: any) => f.target_type === filterType.value)
  }
  if (searchText.value) {
    const kw = searchText.value.toLowerCase()
    list = list.filter((f: any) => (f.title || '').toLowerCase().includes(kw))
  }
  return list
})

function selectFolder(id: number) {
  currentFolder.value = id
  if (id > 0) {
    loadFavorites(undefined, id)
  } else {
    loadFavorites()
  }
}

function getTitle(row: any): string {
  return row.title || (row.target_type === 1 ? '题目 #' + row.target_id : '试卷 #' + row.target_id)
}

function formatTime(t: string): string {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

function onSelectionChange(rows: any[]) {
  selectedRows.value = rows
}

async function onRemove(row: any) {
  await toggle(row.target_type, row.target_id, row.folder_id)
  await loadFavorites()
}

async function onBatchRemove() {
  await ElMessageBox.confirm(`确定移除 ${selectedRows.value.length} 项？`, '提示', { type: 'warning' })
  for (const row of selectedRows.value) {
    await toggle(row.target_type, row.target_id, row.folder_id)
  }
  await loadFavorites()
  ElMessage.success('已移除')
}

async function onBatchMove() {
  if (folders.value.length === 0) {
    ElMessage.warning('请先创建收藏夹')
    return
  }
  const { value: target } = await ElMessageBox.prompt('输入目标收藏夹 ID（开发中功能）', '移动', {
    inputPattern: /\d+/,
    inputErrorMessage: '请输入数字'
  }).catch(() => ({ value: '' }))
  if (!target) return
  ElMessage.success('已移动')
}

function onPreview(row: any) {
  // 路由跳转
  if (row.target_type === 1) {
    window.location.href = '/student/question-bank?id=' + row.target_id
  }
}

async function onCreateFolder() {
  if (!newFolder.value.name) {
    ElMessage.warning('请输入名称')
    return
  }
  await createF(newFolder.value.name, newFolder.value.color, newFolder.value.icon)
  showCreateDialog.value = false
  newFolder.value = { name: '', color: '#409eff', icon: '' }
}

async function onDeleteFolder(id: number) {
  await delF(id)
  if (currentFolder.value === id) {
    currentFolder.value = 0
  }
}

onMounted(async () => {
  await loadFolders()
  await loadFavorites()
})
</script>

<style scoped lang="scss">
.favorites-page {
  padding: 16px;
}
.folder-list {
  min-height: 600px;
}
.folder-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.folder-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
  margin-bottom: 4px;
  &:hover {
    background: var(--el-color-primary-light-9);
  }
  &.active {
    background: var(--el-color-primary-light-7);
    color: var(--el-color-primary);
  }
  .el-tag {
    margin-left: auto;
  }
}
.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.header-actions {
  display: flex;
  gap: 8px;
}
.batch-bar {
  margin-top: 12px;
  padding: 12px;
  background: var(--el-color-primary-light-9);
  border-radius: 6px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
