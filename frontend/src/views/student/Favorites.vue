<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { favoriteApi } from '@/api/modules/favorite'

const folders = ref<any[]>([])
const newFolderName = ref('')
const dialogVisible = ref(false)

async function fetchFolders() {
  const { data } = await favoriteApi.listFolders()
  folders.value = data || []
}

async function createFolder() {
  if (!newFolderName.value) return
  await favoriteApi.createFolder({ name: newFolderName.value })
  newFolderName.value = ''
  dialogVisible.value = false
  fetchFolders()
}

async function removeFolder(id: number) {
  await ElMessageBox.confirm('确定删除该收藏夹？', '提示', { type: 'warning' })
  await favoriteApi.deleteFolder(id)
  fetchFolders()
}

onMounted(fetchFolders)
</script>

<template>
  <div class="koala-page">
    <h2>⭐ 我的收藏</h2>
    <el-card>
      <el-button type="primary" @click="dialogVisible = true">+ 新建收藏夹</el-button>
      <el-row :gutter="20" style="margin-top:16px">
        <el-col :span="6" v-for="f in folders" :key="f.id" style="margin-bottom:16px">
          <el-card class="folder-card">
            <h3>
              <el-icon><Folder /></el-icon>
              {{ f.name }}
              <el-tag v-if="f.is_system" size="small" type="success">系统</el-tag>
            </h3>
            <p>共 {{ f.question_cnt }} 题</p>
            <el-button v-if="!f.is_system" size="small" type="danger" @click="removeFolder(f.id)">删除</el-button>
          </el-card>
        </el-col>
      </el-row>
    </el-card>
    <el-dialog v-model="dialogVisible" title="新建收藏夹" width="400">
      <el-input v-model="newFolderName" placeholder="收藏夹名称，如：高频错题" />
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createFolder">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.folder-card h3 { display: flex; align-items: center; gap: 8px; margin: 0; }
</style>
