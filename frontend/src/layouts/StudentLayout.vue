<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { useUserStore } from '@/store/modules/user'
import { useLayoutStore } from '@/store/modules/layout'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const layout = useLayoutStore()
const fullscreen = ref(false)

interface MenuItem {
  path?: string
  title: string
  icon?: string
  children?: MenuItem[]
}

const menus: MenuItem[] = [
  { path: '/student/exam-hall', title: '考试大厅', icon: 'Calendar' },
  { path: '/student/records', title: '我的考试', icon: 'List' },
  { path: '/student/wrong-book', title: '错题本', icon: 'Warning' },
  { path: '/student/favorites', title: '我的收藏', icon: 'Star' },
  { path: '/student/profile', title: '个人资料', icon: 'User' },
]

const breadcrumbs = computed(() => {
  const matched = route.matched.filter(r => r.meta?.title)
  return matched.map(r => ({ title: r.meta.title as string, path: r.path }))
})

const themeMenu = [
  { label: '浅色', value: 'light' },
  { label: '深色', value: 'dark' },
]

const colorList = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#722ed1', '#13c2c2']

const userCommands = {
  logout: async () => {
    try {
      await ElMessageBox.confirm('确认退出登录？', '提示', { type: 'warning' })
      userStore.logout()
      router.push('/login')
    } catch {}
  },
}

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
  layout.setTheme(layout.theme)
  layout.setPrimaryColor(layout.primaryColor)
})
</script>

<template>
  <div class="koala-layout">
    <aside class="koala-aside" :class="{ collapsed: layout.sidebarCollapsed }">
      <div class="koala-logo" :class="{ small: layout.sidebarCollapsed }">
        <span class="logo-icon">🐨</span>
        <span v-show="!layout.sidebarCollapsed" class="logo-text">考拉智测</span>
      </div>
      <el-menu
        :default-active="route.path"
        :collapse="layout.sidebarCollapsed"
        router
        background-color="var(--koala-aside-bg)"
        text-color="var(--koala-aside-text)"
        active-text-color="#fff"
        class="koala-menu"
      >
        <el-menu-item v-for="m in menus" :key="m.path" :index="m.path">
          <el-icon><component :is="m.icon" /></el-icon>
          <template #title>{{ m.title }}</template>
        </el-menu-item>
      </el-menu>
    </aside>

    <div class="koala-main-wrap">
      <header class="koala-header">
        <div class="header-left">
          <el-button text @click="layout.toggleSidebar">
            <el-icon :size="20">
              <component :is="layout.sidebarCollapsed ? 'Expand' : 'Fold'" />
            </el-icon>
          </el-button>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-for="b in breadcrumbs" :key="b.path">{{ b.title }}</el-breadcrumb-item>
            </el-breadcrumb>
        </div>
        <div class="header-right">
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
          <el-button text @click="toggleFullscreen">
            <el-icon :size="18"><FullScreen v-if="!fullscreen" /><Aim v-else /></el-icon>
          </el-button>
          <el-button text @click="refreshPage">
            <el-icon :size="18"><Refresh /></el-icon>
          </el-button>
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
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

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
  :deep(.el-menu-item) { height: 48px; line-height: 48px; }
  :deep(.el-menu-item.is-active) {
    background: var(--el-color-primary) !important;
    color: #fff !important;
  }
}
.user-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  &:hover { background: rgba(0,0,0,.04); }
}
.user-name { font-size: 14px; }
.fade-enter-active, .fade-leave-active { transition: opacity .15s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
