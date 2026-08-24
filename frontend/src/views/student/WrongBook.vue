<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { favoriteApi } from '@/api/modules/favorite'
import MasteryTag from '@/components/business/MasteryTag.vue'
import FavoriteStar from '@/components/business/FavoriteStar.vue'
import dayjs from 'dayjs'

const list = ref<any[]>([])
const total = ref(0)
const loading = ref(false)
const filterMastery = ref<number>(0)

async function fetchList() {
  loading.value = true
  try {
    const { data } = await favoriteApi.wrongBook({ mastery_level: filterMastery.value, page: 1, page_size: 50 })
    list.value = data!.list || []
    total.value = data!.total
  } finally { loading.value = false }
}

async function markReviewed(logId: number, level: number) {
  try {
    await favoriteApi.markReviewed(logId, level)
    ElMessage.success('已标记复习')
    fetchList()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

onMounted(fetchList)
</script>

<template>
  <div class="koala-page">
    <h2>📒 错题本</h2>
    <p class="tip">系统自动收录每次考试中答错的题目，复习后掌握度会逐步提升 💪</p>
    <el-card>
      <div class="filter">
        <span>掌握度筛选：</span>
        <el-radio-group v-model="filterMastery" @change="fetchList">
          <el-radio :value="0">全部</el-radio>
          <el-radio :value="1">未掌握</el-radio>
          <el-radio :value="2">薄弱</el-radio>
          <el-radio :value="3">一般</el-radio>
          <el-radio :value="4">良好</el-radio>
          <el-radio :value="5">熟练</el-radio>
        </el-radio-group>
      </div>
      <div v-loading="loading">
        <el-empty v-if="!loading && list.length === 0" description="错题本是空的，继续保持 💯" />
        <div v-for="item in list" :key="item.log_id" class="wrong-item koala-card">
          <div class="head">
            <span class="tag">{{ ['单选','多选','判断','填空','不定项','编程'][item.question.type] }}</span>
            <MasteryTag :level="item.mastery_level" />
            <span class="time">{{ dayjs(item.last_wrong_at).format('YYYY-MM-DD HH:mm') }}</span>
            <span class="count">错 {{ item.wrong_count }} 次</span>
            <FavoriteStar :target-type="1" :target-id="item.question.id" size="small" />
          </div>
          <div class="title" v-html="item.question.title"></div>
          <div class="answer-row">
            <div><span class="lbl">你的答案：</span><el-tag type="danger">{{ String(item.user_answer ?? '未作答') }}</el-tag></div>
            <div><span class="lbl">正确答案：</span><el-tag type="success">{{ String(item.correct_answer) }}</el-tag></div>
          </div>
          <div class="actions">
            <el-button size="small" type="success" @click="markReviewed(item.log_id, Math.min(item.mastery_level + 1, 5))" :disabled="item.is_reviewed">
              {{ item.is_reviewed ? '已复习' : '已掌握' }}
            </el-button>
            <el-button size="small" @click="markReviewed(item.log_id, 1)">需要再练</el-button>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.tip { color: #999; margin-bottom: 16px; }
.filter { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
.wrong-item { margin-bottom: 16px; }
.head { display: flex; align-items: center; gap: 12px; padding-bottom: 8px; }
.head .time { color: #999; font-size: 12px; }
.head .count { color: #f56c6c; font-size: 12px; }
.title { padding: 8px 0; }
.answer-row { display: flex; gap: 24px; padding: 8px 0; }
.lbl { color: #666; margin-right: 8px; }
.actions { padding-top: 8px; border-top: 1px dashed #eee; }
</style>
