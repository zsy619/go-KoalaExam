<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/store/modules/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const isCollapsed = ref(false)

function handleCommand(c: string) {
  if (c === "logout") { userStore.logout(); router.push("/login") }
  else if (c === "profile") { router.push("/admin/profile") }
}

const menus = [
  { path: '/admin/dashboard', title: '仪表盘', icon: 'Odometer' },
  { path: '/admin/users', title: '用户管理', icon: 'User', role: 1 },
  { path: '/admin/questions', title: '题库管理', icon: 'Notebook' },
  { path: '/admin/papers', title: '试卷管理', icon: 'Document' },
  { path: '/admin/exams', title: '考试管理', icon: 'Calendar' },
  { path: '/admin/statistics', title: '成绩统计', icon: 'DataLine' },
]

const visibleMenus = computed(() => menus.filter((m) => !m.role || userStore.role === m.role))

const activeMenu = computed(() => route.path)

</script>

<template>
  <el-container class="admin-layout">
    <el-aside :width="isCollapsed ? '64px' : '220px'" class="aside">
      <div class="logo">
        <span v-if="!isCollapsed">🐨 考拉智测</span>
        <span v-else>🐨</span>
      </div>
      <el-menu :default-active="activeMenu" :collapse="isCollapsed" router>
        <el-menu-item v-for="m in visibleMenus" :key="m.path" :index="m.path">
          <el-icon><component :is="m.icon" /></el-icon>
          <template #title>{{ m.title }}</template>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <el-button text @click="isCollapsed = !isCollapsed">
          <el-icon><Expand v-if="isCollapsed" /><Fold v-else /></el-icon>
        </el-button>
        <el-dropdown @command="(c: string) => c === 'logout' && logout()">
          <span class="user">
            {{ userStore.profile?.nickname || userStore.profile?.username }} <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped lang="scss">
.admin-layout { height: 100vh; }
.aside { background: #001529; transition: width .3s; .logo { height: 60px; line-height: 60px; color: #fff; text-align: center; font-size: 16px; font-weight: 600; } }
.header { background: #fff; display: flex; align-items: center; justify-content: space-between; padding: 0 16px; border-bottom: 1px solid #eee; .user { cursor: pointer; } }
.el-main { padding: 20px; background: #f5f7fa; }
</style>
