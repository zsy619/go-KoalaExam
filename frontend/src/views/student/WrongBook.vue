<template>
  <div class="wrongbook-page">
    <!-- 智能概览 -->
    <el-row :gutter="16" class="stats-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <el-icon class="stat-icon" color="#f56c6c"><WarningFilled /></el-icon>
            <div>
              <div class="stat-value">{{ wrongList.length }}</div>
              <div class="stat-label">累计错题</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <el-icon class="stat-icon" color="#e6a23c"><Edit /></el-icon>
            <div>
              <div class="stat-value">{{ masteryCounts.unmastered || 0 }}</div>
              <div class="stat-label">未掌握</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <el-icon class="stat-icon" color="#67c23a"><CircleCheckFilled /></el-icon>
            <div>
              <div class="stat-value">{{ masteryCounts.mastered || 0 }}</div>
              <div class="stat-label">已掌握</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <el-icon class="stat-icon" color="#409eff"><DataLine /></el-icon>
            <div>
              <div class="stat-value">{{ avgMastery.toFixed(1) }}</div>
              <div class="stat-label">平均掌握度</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 筛选区 -->
    <el-card class="filter-card">
      <div class="filter-row">
        <el-radio-group v-model="filterMastery" @change="loadList">
          <el-radio-button :value="0">全部</el-radio-button>
          <el-radio-button :value="1">未掌握</el-radio-button>
          <el-radio-button :value="3">部分掌握</el-radio-button>
          <el-radio-button :value="5">已掌握</el-radio-button>
        </el-radio-group>

        <el-checkbox v-model="onlyUnreviewed" @change="loadList" class="ml-3">
          仅显示未复习
        </el-checkbox>

        <el-input
          v-model="searchKw"
          placeholder="搜索错题..."
          clearable
          style="width: 240px; margin-left: auto;"
          @input="onSearchInput"
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-button type="primary" @click="loadList">刷新</el-button>
        <el-button type="success" @click="onPracticeMode" :disabled="wrongList.length === 0">
          <el-icon><VideoPlay /></el-icon>开始练习
        </el-button>
      </div>
    </el-card>

    <!-- 错题列表 -->
    <el-card>
      <el-empty v-if="!loading && wrongList.length === 0" description="恭喜，没有错题！" />
      <div v-else v-loading="loading">
        <div
          v-for="item in filteredList"
          :key="item.log_id"
          class="wrong-item"
          @click="onShowDetail(item)"
        >
          <div class="wrong-item-header">
            <MasteryTag :level="item.mastery_level" />
            <span class="wrong-title">{{ item.title || ('错题 #' + item.question_id) }}</span>
            <span class="wrong-meta">
              错误 {{ item.wrong_count || 1 }} 次
              <el-divider direction="vertical" />
              {{ formatTime(item.last_wrong_at) }}
            </span>
            <div class="wrong-actions" @click.stop>
              <el-button
                v-if="!item.is_reviewed"
                type="primary"
                size="small"
                @click="onMarkReviewed(item)"
              >
                标记已复习
              </el-button>
              <el-tag v-else type="success" size="small">已复习</el-tag>
            </div>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 错题详情对话框 -->
    <el-dialog v-model="detailVisible" title="错题详情" width="700px">
      <div v-if="currentItem">
        <div class="detail-section">
          <h4>题目</h4>
          <div class="detail-content">{{ currentItem.question?.title || currentItem.title }}</div>
        </div>
        <div class="detail-section" v-if="currentItem.question?.options">
          <h4>选项</h4>
          <div class="detail-options">
            <div v-for="(opt, idx) in parseOptions(currentItem.question?.options)" :key="idx">
              {{ idx }}: {{ opt }}
            </div>
          </div>
        </div>
        <div class="detail-section">
          <h4>你的答案</h4>
          <el-tag type="danger">{{ formatAnswer(currentItem.user_answer) }}</el-tag>
        </div>
        <div class="detail-section">
          <h4>正确答案</h4>
          <el-tag type="success">{{ formatAnswer(currentItem.correct_answer) }}</el-tag>
        </div>
        <div class="detail-section" v-if="currentItem.question?.analysis">
          <h4>解析</h4>
          <div class="detail-content">{{ currentItem.question.analysis }}</div>
        </div>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
        <el-button type="primary" @click="onMarkReviewed(currentItem)">标记已掌握</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { WarningFilled, Edit, CircleCheckFilled, DataLine, Search, VideoPlay } from '@element-plus/icons-vue'
import { useFavorite } from '@/composables/useFavorite'
import MasteryTag from '@/components/business/MasteryTag.vue'

const {
  wrongBook, loading,
  loadWrongBook, markReviewed
} = useFavorite()

const wrongList = ref<any[]>([])
const filterMastery = ref(0)
const onlyUnreviewed = ref(false)
const searchKw = ref('')
const detailVisible = ref(false)
const currentItem = ref<any>(null)

// 计算每种掌握度的数量
const masteryCounts = computed(() => {
  const counts: Record<string, number> = { mastered: 0, unmastered: 0, partial: 0 }
  wrongList.value.forEach((item: any) => {
    if (item.mastery_level >= 4) counts.mastered++
    else if (item.mastery_level <= 2) counts.unmastered++
    else counts.partial++
  })
  return counts
})

// 平均掌握度
const avgMastery = computed(() => {
  if (wrongList.value.length === 0) return 0
  const sum = wrongList.value.reduce((acc: number, item: any) => acc + (item.mastery_level || 1), 0)
  return sum / wrongList.value.length
})

// 过滤后的列表
const filteredList = computed(() => {
  let list = wrongList.value
  if (searchKw.value) {
    const kw = searchKw.value.toLowerCase()
    list = list.filter((i: any) =>
      (i.title || '').toLowerCase().includes(kw) ||
      String(i.question_id).includes(kw)
    )
  }
  return list
})

async function loadList() {
  const res: any = await loadWrongBook({
    masteryLevel: filterMastery.value || undefined,
    isReviewed: onlyUnreviewed.value ? false : undefined
  })
  const respData = res.data || res
      wrongList.value = respData.list || []
}

let searchTimer: any = null
function onSearchInput() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    // 搜索是前端的，过滤即时生效
  }, 300)
}

function formatTime(t: string) {
  if (!t) return '-'
  const date = new Date(t)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (24 * 3600 * 1000))
  if (days === 0) return '今天 ' + date.toTimeString().substring(0, 5)
  if (days === 1) return '昨天'
  if (days < 7) return days + ' 天前'
  return date.toLocaleDateString('zh-CN')
}

function formatAnswer(ans: any) {
  if (Array.isArray(ans)) return ans.join(', ')
  if (typeof ans === 'object') return JSON.stringify(ans)
  return String(ans || '-')
}

function parseOptions(opts: any): Record<string, string> {
  if (!opts) return {}
  if (typeof opts === 'string') {
    try { return JSON.parse(opts) } catch { return {} }
  }
  return opts
}

function onShowDetail(item: any) {
  currentItem.value = item
  detailVisible.value = true
}

async function onMarkReviewed(item: any) {
  await markReviewed(item.log_id, 4)
  await loadList()
}

function onPracticeMode() {
  ElMessage.info('练习模式（开发中）：将基于错题智能组卷')
  // 实际跳转: window.location.href = '/student/exam-room?mode=wrong-practice&ids=' + wrongList.value.map(i=>i.question_id).join(',')
}

onMounted(() => {
  loadList()
})
</script>

<style scoped lang="scss">
.wrongbook-page {
  padding: 16px;
}
.stats-row {
  margin-bottom: 16px;
}
.stat-card {
  .stat-content {
    display: flex;
    align-items: center;
    gap: 16px;
  }
  .stat-icon {
    font-size: 36px;
  }
  .stat-value {
    font-size: 24px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }
  .stat-label {
    font-size: 13px;
    color: var(--el-text-color-secondary);
    margin-top: 4px;
  }
}
.filter-card {
  margin-bottom: 16px;
}
.filter-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.ml-3 {
  margin-left: 12px;
}
.wrong-item {
  padding: 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  cursor: pointer;
  transition: background 0.2s;
  &:hover {
    background: var(--el-color-primary-light-9);
  }
  &:last-child {
    border-bottom: none;
  }
}
.wrong-item-header {
  display: flex;
  align-items: center;
  gap: 12px;
}
.wrong-title {
  font-weight: 500;
  flex: 1;
  color: var(--el-text-color-primary);
}
.wrong-meta {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
.wrong-actions {
  display: flex;
  gap: 8px;
}
.detail-section {
  margin-bottom: 16px;
  h4 {
    margin: 0 0 8px;
    color: var(--el-text-color-primary);
  }
  .detail-content {
    padding: 12px;
    background: var(--el-color-info-light-9);
    border-radius: 4px;
    line-height: 1.6;
  }
  .detail-options {
    padding: 8px 12px;
    background: var(--el-bg-color-page);
    border-radius: 4px;
    line-height: 1.8;
  }
}
</style>
