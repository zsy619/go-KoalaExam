import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import setupGuards from './guards/permission'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/Login.vue'),
    meta: { public: true, title: '登录' },
  },
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    redirect: '/dashboard',
  },
  {
    path: '/student',
    component: () => import('@/layouts/StudentLayout.vue'),
    meta: { requiresAuth: true, role: 3 },
    children: [
      { path: 'exam-hall', name: 'ExamHall', component: () => import('@/views/student/ExamHall.vue'), meta: { title: '考试大厅' } },
      { path: 'exam-room/:id', name: 'ExamRoom', component: () => import('@/views/student/ExamRoom.vue'), meta: { title: '答题中', fullscreen: true } },
      { path: 'records', name: 'ExamRecords', component: () => import('@/views/student/ExamRecords.vue'), meta: { title: '我的考试' } },
      { path: 'wrong-book', name: 'WrongBook', component: () => import('@/views/student/WrongBook.vue'), meta: { title: '错题本' } },
      { path: 'favorites', name: 'MyFavorites', component: () => import('@/views/student/Favorites.vue'), meta: { title: '我的收藏' } },
      { path: 'profile', name: 'Profile', component: () => import('@/views/Profile.vue'), meta: { title: '个人资料' } },
    ],
  },
  {
    path: '/admin',
    component: () => import('@/layouts/AdminLayout.vue'),
    meta: { requiresAuth: true, role: [1, 2] },
    children: [
      { path: 'dashboard', name: 'Dashboard', component: () => import('@/views/admin/Dashboard.vue'), meta: { title: '仪表盘' } },
      { path: 'users', name: 'UserManage', component: () => import('@/views/admin/UserManage.vue'), meta: { title: '用户管理', role: [1] } },
      { path: 'questions', name: 'QuestionBank', component: () => import('@/views/admin/QuestionBank.vue'), meta: { title: '题库管理' } },
      { path: 'papers', name: 'PaperManage', component: () => import('@/views/admin/PaperManage.vue'), meta: { title: '试卷管理' } },
      { path: 'exams', name: 'ExamManage', component: () => import('@/views/admin/ExamManage.vue'), meta: { title: '考试管理' } },
      { path: 'statistics', name: 'Statistics', component: () => import('@/views/admin/Statistics.vue'), meta: { title: '成绩统计' } },
      { path: 'profile', name: 'AdminProfile', component: () => import('@/views/Profile.vue'), meta: { title: '个人资料' } },
    ],
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/error/Forbidden.vue'),
    meta: { public: true },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/error/NotFound.vue'),
    meta: { public: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

setupGuards(router)

export default router
