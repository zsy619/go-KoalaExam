<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { userApi } from '@/api/modules/user'
import { useUserStore } from '@/store/modules/user'
import type { User } from '@/types/entity'

const userStore = useUserStore()
const profile = ref<User | null>(null)
const editing = ref(false)
const form = ref({ nickname: '', email: '', phone: '', avatar: '' })
const pwdForm = ref({ old_password: '', new_password: '' })

async function loadProfile() {
  const { data } = await userApi.profile()
  profile.value = data!
  form.value = { nickname: data!.nickname, email: data!.email, phone: data!.phone, avatar: data!.avatar }
}

async function saveProfile() {
  await userApi.profileUpdate(form.value)
  ElMessage.success('资料已更新')
  editing.value = false
  loadProfile()
}

async function changePwd() {
  if (!pwdForm.value.old_password || !pwdForm.value.new_password) {
    ElMessage.warning('请填写完整')
    return
  }
  await userApi.changePassword(pwdForm.value)
  ElMessage.success('密码已修改，请重新登录')
  setTimeout(() => userStore.logout(), 1500)
}

onMounted(loadProfile)
</script>

<template>
  <div class="koala-page">
    <el-row :gutter="16">
      <el-col :span="12">
        <el-card>
          <template #header>个人资料</template>
          <template v-if="profile">
            <el-descriptions :column="1" border>
              <el-descriptions-item label="账号">{{ profile.username }}</el-descriptions-item>
              <el-descriptions-item label="昵称">{{ profile.nickname }}</el-descriptions-item>
              <el-descriptions-item label="角色">
                <el-tag>{{ { 1: '超管', 2: '教师', 3: '学员' }[profile.role] }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="邮箱">{{ profile.email }}</el-descriptions-item>
              <el-descriptions-item label="手机">{{ profile.phone }}</el-descriptions-item>
              <el-descriptions-item label="注册时间">{{ profile.created_at }}</el-descriptions-item>
              <el-descriptions-item label="最后登录">{{ profile.last_login_at || '-' }} · {{ profile.last_login_ip || '' }}</el-descriptions-item>
            </el-descriptions>
            <el-button style="margin-top:16px" type="primary" @click="editing = true">编辑资料</el-button>
          </template>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>修改密码</template>
          <el-form label-width="100px">
            <el-form-item label="原密码">
              <el-input v-model="pwdForm.old_password" type="password" show-password />
            </el-form-item>
            <el-form-item label="新密码">
              <el-input v-model="pwdForm.new_password" type="password" show-password />
            </el-form-item>
            <el-form-item>
              <el-button type="warning" @click="changePwd">修改密码</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
    <el-dialog v-model="editing" title="编辑资料" width="500px">
      <el-form label-width="80px">
        <el-form-item label="昵称"><el-input v-model="form.nickname" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item>
        <el-form-item label="手机"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="头像 URL"><el-input v-model="form.avatar" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editing = false">取消</el-button>
        <el-button type="primary" @click="saveProfile">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>
