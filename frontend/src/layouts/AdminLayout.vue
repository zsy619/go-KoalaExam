<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { useUserStore } from '@/store/modules/user'
import { useLayoutStore } from '@/store/modules/layout'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const layout = useLayoutStore()
const fullscreen = ref(false)

// ====== 菜单（多级嵌套，芋道风） ======
interface MenuItem {
  path?: string
  title: string
  icon?: string
  role?: number | number[]
  children?: MenuItem[]
  hidden?: boolean
}

const menus: MenuItem[] = [
  {
    path: '/admin/dashboard',
    title: '仪表盘',
    icon: 'Odometer',
  },
  {
    title: '题库中心',
    icon: 'Notebook',
    children: [
      { path: '/admin/questions', title: '题库管理', icon: 'List' },
      { path: '/admin/questions/categories', title: '分类管理', icon: 'Folder' },
    ],
  },
  {
    title: '考试管理',
    icon: 'Calendar',
    children: [
      { path: '/admin/papers', title: '试卷管理', icon: 'Document' },
      { path: '/admin/exams', title: '考试计划', icon: 'Timer' },
      { path: '/admin/grading', title: '人工批改', icon: 'EditPen' },
    ],
  },
  {
    title: '统计分析',
    icon: 'DataAnalysis',
    children: [
      { path: '/admin/statistics', title: '成绩统计', icon: 'TrendCharts' },
      { path: '/admin/statistics/ranking', title: '排行榜', icon: 'Trophy' },
    ],
  },
  {
    title: '系统管理',
    icon: 'Setting',
    role: 1, // 仅管理员
    children: [
      { path: '/admin/users', title: '用户管理', icon: 'User', role: 1 },
    ],
  },
]

function hasRole(m: MenuItem): boolean {
  if (!m.role) return true
  const arr = Array.isArray(m.role) ? m.role : [m.role]
  return arr.includes(userStore.role)
}

function filterMenus(list: MenuItem[]): MenuItem[] {
  return list
    .filter(m => hasRole(m))
    .map(m => ({
      ...m,
      children: m.children ? filterMenus(m.children) : undefined,
    }))
}

const visibleMenus = computed(() => filterMenus(menus))

const breadcrumbs = computed(() => {
  const matched = route.matched.filter(r => r.meta?.title)
  return matched.map(r => ({ title: r.meta.title as string, path: r.path }))
})

const themeMenu = [
  { label: '浅色', value: 'light' },
  { label: '深色', value: 'dark' },
]

const colorList = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#722ed1', '#13c2c2']

const activeMenu = computed(() => route.path)

// 用户菜单
const userCommands = {
  profile: () => router.push('/admin/profile'),
  logout: async () => {
    try {
      await ElMessageBox.confirm('确认退出登录？', '提示', { type: 'warning' })
      userStore.logout()
      router.push('/login')
    } catch {}
  },
}

// 标签页操作
function closeTag(path: string) {
  layout.removeVisitedView(path)
  if (route.path === path) {
    const last = layout.visitedViews[layout.visitedViews.length - 1]
    router.push(last?.path || '/dashboard')
  }
}

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen()
    fullscreen.value = true
  } else {
    document.exitFullscreen()
    fullscreen.value = false
  }
}

function refreshPage() {
  router.go(0)
}

onMounted(() => {
  // 初始化主题
  layout.setTheme(layout.theme)
  layout.setPrimaryColor(layout.primaryColor)
  // 监听路由变化
})
</script>

<template>
  <div class="koala-layout">
    <!-- 左侧菜单 -->
    <aside class="koala-aside" :class="{ collapsed: layout.sidebarCollapsed }">
      <div class="koala-logo" :class="{ small: layout.sidebarCollapsed }">
        <span class="logo-icon">🐨</span>
        <span v-show="!layout.sidebarCollapsed" class="logo-text">考拉智测</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="layout.sidebarCollapsed"
        :collapse-transition="false"
        router
        background-color="var(--koala-aside-bg)"
        text-color="var(--koala-aside-text)"
        active-text-color="var(--koala-aside-text-active)"
        class="koala-menu"
      >
        <template v-for="m in visibleMenus" :key="m.path || m.title">
          <!-- 有子菜单 -->
          <el-sub-menu v-if="m.children && m.children.length" :index="m.path || m.title">
            <template #title>
              <el-icon><component :is="m.icon" /></el-icon>
              <span>{{ m.title }}</span>
            </template>
            <el-menu-item v-for="c in m.children" :key="c.path" :index="c.path">
              <el-icon><component :is="c.icon" /></el-icon>
              <template #title>{{ c.title }}</template>
            </el-menu-item>
          </el-sub-menu>
          <!-- 叶子节点 -->
          <el-menu-item v-else :index="m.path">
            <el-icon><component :is="m.icon" /></el-icon>
            <template #title>{{ m.title }}</template>
          </el-menu-item>
        </template>
      </el-menu>
    </aside>

    <!-- 右侧主区 -->
    <div class="koala-main-wrap">
      <!-- 顶栏 -->
      <header class="koala-header">
        <div class="header-left">
          <el-button text @click="layout.toggleSidebar">
            <el-icon :size="20">
              <component :is="layout.sidebarCollapsed ? 'Expand' : 'Fold'" />
            </el-icon>
          </el-button>
          <el-breadcrumb separator="/" class="header-breadcrumb">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-for="b in breadcrumbs" :key="b.path">{{ b.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <!-- 主题切换 -->
          <el-dropdown @command="(c: string) => layout.setTheme(c)">
            <el-button text>
              <el-icon :size="18"><Sunny v-if="layout.theme === 'dark'" /><Moon v-else /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-for="t in themeMenu" :key="t.value" :command="t.value">{{ t.label }}</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <!-- 主色 -->
          <el-dropdown @command="(c: string) => layout.setPrimaryColor(c)">
            <el-button text>
              <el-icon :size="18"><Brush /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-for="c in colorList" :key="c" :command="c">
                  <span :style="{ display: 'inline-block', width: '14px', height: '14px', borderRadius: '50%', background: c, marginRight: '6px' }" />
                  {{ c }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <!-- 全屏 -->
          <el-button text @click="toggleFullscreen">
            <el-icon :size="18"><FullScreen v-if="!fullscreen" /><Aim v-else /></el-icon>
          </el-button>
          <!-- 刷新 -->
          <el-button text @click="refreshPage">
            <el-icon :size="18"><Refresh /></el-icon>
          </el-button>
          <!-- 用户菜单 -->
          <el-dropdown @command="(c: string) => userCommands[c as keyof typeof userCommands]?.()">
            <span class="user-trigger">
              <el-avatar :size="32" :src="userStore.profile?.avatar">
                {{ userStore.profile?.nickname?.[0] || userStore.profile?.username?.[0] || 'U' }}
              </el-avatar>
              <span class="user-name">{{ userStore.profile?.nickname || userStore.profile?.username }}</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon>个人中心
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <!-- 标签页 -->
      <div class="koala-tags-view">
        <el-dropdown trigger="contextmenu" @command="(c: string) => {
          if (c === 'close') closeTag(route.path)
          else if (c === 'close-others') { layout.removeOthers(route.path) }
          else if (c === 'close-all') { layout.removeAll(); router.push('/dashboard') }
        }">
          <span
            v-for="tag in layout.visitedViews"
            :key="tag.path"
            class="tag-item"
            :class="{ active: tag.path === route.path }"
            @click="router.push(tag.path)"
          >
            {{ tag.title }}
            <el-icon v-if="tag.path !== '/dashboard'" class="close-icon" @click.stop="closeTag(tag.path)">
              <Close />
            </el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="close">关闭当前</el-dropdown-item>
              <el-dropdown-item command="close-others">关闭其他</el-dropdown-item>
              <el-dropdown-item command="close-all">关闭所有</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>

      <!-- 内容 -->
      <main class="koala-main">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<style scoped lang="scss">
.koala-menu {
  border-right: none;
  flex: 1;
  :deep(.el-menu-item), :deep(.el-sub-menu__title) {
    height: 48px;
    line-height: 48px;
  }
  :deep(.el-menu-item.is-active) {
    background: var(--el-color-primary) !important;
    color: #fff !important;
  }
}
.header-breadcrumb { font-size: 14px; }
.user-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background .2s;
  &:hover { background: rgba(0,0,0,.04); }
}
.user-name { font-size: 14px; }
.fade-enter-active, .fade-leave-active { transition: opacity .15s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
