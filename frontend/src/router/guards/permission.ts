import type { Router } from 'vue-router'
import { useUserStore } from '@/store/modules/user'
import { ElMessage } from 'element-plus'

export default function setupGuards(router: Router) {
  router.beforeEach(async (to, from, next) => {
    const userStore = useUserStore()
    document.title = (to.meta.title as string) ? `${to.meta.title} - 考拉智测` : '考拉智测'

    if (to.meta.public) return next()
    if (!userStore.isLogin) return next({ path: '/login', query: { redirect: to.fullPath } })
    if (!userStore.profile) {
      try {
        await userStore.fetchProfile()
      } catch (e) {
        userStore.logout()
        return next({ path: '/login' })
      }
    }

    const required = to.meta.role
    if (required) {
      const allow = Array.isArray(required) ? required.includes(userStore.role) : userStore.role === required
      if (!allow) {
        ElMessage.error('无权限访问')
        return next('/403')
      }
    }
    next()
  })
}
