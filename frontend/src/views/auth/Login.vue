<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/store/modules/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const loading = ref(false)

const form = reactive({ username: '', password: '' })

async function onSubmit() {
  if (!form.username || !form.password) {
    ElMessage.warning('请输入账号和密码')
    return
  }
  loading.value = true
  try {
    await userStore.login(form.username, form.password)
    ElMessage.success('登录成功')
    const redirect = (route.query.redirect as string) || ((userStore.user?.role === 1 || userStore.user?.role === 2) ? '/admin/dashboard' : '/student/exam-hall')
    router.push(redirect)
  } catch (e: any) {
    ElMessage.error(e?.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-wrap">
    <el-card class="login-card">
      <h2>🐨 考拉智测</h2>
      <p class="subtitle">基于 Hertz + Vue3 的在线考试系统</p>
      <el-form :model="form" label-width="80px" @submit.prevent="onSubmit">
        <el-form-item label="账号">
          <el-input v-model="form.username" placeholder="admin / teacher / student" clearable />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" placeholder="默认 koala123" show-password @keyup.enter="onSubmit" />
        </el-form-item>
        <el-button type="primary" :loading="loading" @click="onSubmit" style="width:100%">登 录</el-button>
      </el-form>
      <div class="hint">
        <p>默认账号：</p>
        <p>🔑 超管 admin / 教师 teacher / 学员 student</p>
        <p>🔒 密码统一：koala123</p>
      </div>
    </el-card>
  </div>
</template>

<style scoped lang="scss">
.login-wrap { min-height: 100vh; display: flex; align-items: center; justify-content: center; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
.login-card { width: 420px; padding: 32px; border-radius: 12px; box-shadow: 0 12px 32px rgba(0,0,0,0.15); }
h2 { margin: 0 0 8px; text-align: center; }
.subtitle { color: #999; text-align: center; margin-bottom: 24px; font-size: 13px; }
.hint { margin-top: 20px; padding: 12px; background: #fafafa; border-radius: 6px; color: #666; font-size: 12px; line-height: 1.8; }
</style>
