<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/store/modules/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const menus = [
  { path: '/student/exam-hall', title: '考试大厅', icon: 'Calendar' },
  { path: '/student/records', title: '我的考试', icon: 'List' },
  { path: '/student/wrong-book', title: '错题本', icon: 'Warning' },
  { path: '/student/favorites', title: '我的收藏', icon: 'Star' },
]

const activeMenu = computed(() => route.path)

function handleCommand(cmd: string) {
  if (cmd === "logout") { userStore.logout(); router.push("/login") }
  else if (cmd === "profile") { router.push("/student/profile") }
}
</script>

<template>
  <el-container class="stu-layout">
    <el-header class="header">
      <span class="logo">🐨 考拉智测</span>
      <el-menu :default-active="activeMenu" mode="horizontal" router>
        <el-menu-item v-for="m in menus" :key="m.path" :index="m.path">
          <el-icon><component :is="m.icon" /></el-icon>{{ m.title }}
        </el-menu-item>
      </el-menu>
      <el-dropdown @command="(c: string) => c === 'logout' && logout()">
        <span class="user">{{ userStore.profile?.nickname }} <el-icon><ArrowDown /></el-icon></span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="profile">个人资料</el-dropdown-item>
          <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
          </template>
        </el-dropdown>
      </el-dropdown>
    </el-header>
    <el-main>
      <router-view />
    </el-main>
  </el-container>
</template>

<style scoped lang="scss">
.stu-layout { height: 100vh; }
.header { display: flex; align-items: center; background: #fff; padding: 0 24px; border-bottom: 1px solid #eee; .logo { font-size: 18px; font-weight: 600; margin-right: 40px; } .el-menu { flex: 1; border: none; } .user { cursor: pointer; } }
.el-main { background: #f5f7fa; }
</style>
