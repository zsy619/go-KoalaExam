import { defineStore } from 'pinia'

export const useLayoutStore = defineStore('layout', {
  state: () => ({
    // 侧边栏
    sidebarCollapsed: localStorage.getItem('koala_sidebar') === '1',
    // 主题
    theme: localStorage.getItem('koala_theme') || 'light', // light | dark
    primaryColor: localStorage.getItem('koala_color') || '#409eff',
    // 设备
    isMobile: false,
    // 标签页
    visitedViews: JSON.parse(localStorage.getItem('koala_tags') || '[]'),
  }),
  actions: {
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
      localStorage.setItem('koala_sidebar', this.sidebarCollapsed ? '1' : '0')
    },
    setTheme(t: string) {
      this.theme = t
      localStorage.setItem('koala_theme', t)
      document.documentElement.classList.toggle('dark', t === 'dark')
    },
    setPrimaryColor(c: string) {
      this.primaryColor = c
      localStorage.setItem('koala_color', c)
      const style = document.documentElement.style
      style.setProperty('--el-color-primary', c)
      // 派生浅/暗色
      style.setProperty('--el-color-primary-light-3', c)
    },
    addVisitedView(view: { path: string; name?: string; title?: string }) {
      if (this.visitedViews.some(v => v.path === view.path)) return
      this.visitedViews.push(view)
      localStorage.setItem('koala_tags', JSON.stringify(this.visitedViews))
    },
    removeVisitedView(path: string) {
      const idx = this.visitedViews.findIndex(v => v.path === path)
      if (idx > -1) {
        this.visitedViews.splice(idx, 1)
        localStorage.setItem('koala_tags', JSON.stringify(this.visitedViews))
      }
    },
    removeOthers(path: string) {
      this.visitedViews = this.visitedViews.filter(v => v.path === path || v.path === '/dashboard')
      localStorage.setItem('koala_tags', JSON.stringify(this.visitedViews))
    },
    removeAll() {
      this.visitedViews = this.visitedViews.filter(v => v.path === '/dashboard')
      localStorage.setItem('koala_tags', JSON.stringify(this.visitedViews))
    },
  },
})
