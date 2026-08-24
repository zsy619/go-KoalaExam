<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { examApi, type StartExamResp } from '@/api/modules/exam'
import { useExamTimer } from '@/composables/useExamTimer'
import { useAntiCheat } from '@/composables/useAntiCheat'
import QuestionRenderer from '@/components/business/QuestionRenderer.vue'
import FavoriteStar from '@/components/business/FavoriteStar.vue'

const route = useRoute()
const router = useRouter()
const examId = computed(() => Number(route.params.id))

const exam = ref<StartExamResp | null>(null)
const answers = ref<Record<number, unknown>>({})
const currentIdx = ref(0)
const marked = ref<Set<number>>(new Set())
const submitted = ref(false)

const { remaining, format: formatTime, start: startTimer, pause } = useExamTimer(0, () => onSubmit(true))

const { tabSwitchCount } = useAntiCheat({
  blockCopy: true,
  maxTabSwitch: 3,
  onAudit: (e) => {
    if (exam.value) examApi.audit({ record_id: exam.value.record_id, events: e })
  },
})

async function loadExam() {
  try {
    const { data } = await examApi.start(examId.value)
    exam.value = data!
    answers.value = {}
    startTimer()
  } catch (e: any) {
    ElMessage.error(e.message || '加载考试失败')
    router.push('/student/exam-hall')
  }
}

let syncTimer: number | null = null
function scheduleSync() {
  if (syncTimer) return
  syncTimer = window.setTimeout(() => { syncTimer = null; flushAnswers() }, 10000)
}

async function flushAnswers() {
  if (!exam.value) return
  for (const [qid, ans] of Object.entries(answers.value)) {
    try { await examApi.saveAnswer({ record_id: exam.value.record_id, question_id: Number(qid), answer: ans }) } catch { /* ignore */ }
  }
}

function onAnswer(qid: number, value: unknown) {
  answers.value[qid] = value
  scheduleSync()
}

function toggleMark(idx: number) {
  if (marked.value.has(idx)) marked.value.delete(idx); else marked.value.add(idx)
  marked.value = new Set(marked.value)
}

function jumpTo(idx: number) {
  if (idx < 0 || idx >= (exam.value?.questions.length || 0)) return
  currentIdx.value = idx
}

const answeredCount = computed(() => {
  return Object.entries(answers.value).filter(([_, v]) => {
    if (v === undefined || v === null || v === '') return false
    if (Array.isArray(v) && v.length === 0) return false
    return true
  }).length
})

async function onSubmit(force = false) {
  if (!exam.value || submitted.value) return
  if (!force) {
    try {
      await ElMessageBox.confirm(`已作答 ${answeredCount.value}/${exam.value.questions.length} 题，确定交卷？`, '确认交卷', { type: 'warning' })
    } catch { return }
  }
  submitted.value = true
  pause()
  try {
    await flushAnswers()
    const { data } = await examApi.submit(exam.value.record_id)
    ElMessage.success(`已交卷！得分 ${data?.total_score || 0}`)
    router.push('/student/records')
  } catch (e: any) {
    submitted.value = false
    ElMessage.error(e.message || '交卷失败')
  }
}

onMounted(loadExam)
onUnmounted(() => { if (syncTimer) clearTimeout(syncTimer) })
</script>

<template>
  <div class="exam-room" v-if="exam">
    <header class="exam-header koala-card">
      <div>
        <strong>{{ exam.title }}</strong>
        <div class="meta">已作答 {{ answeredCount }}/{{ exam.questions.length }} | 切屏 {{ tabSwitchCount }} 次</div>
      </div>
      <div class="timer" :class="{ danger: remaining < 60 }">⏱ {{ formatTime() }}</div>
      <el-button type="warning" @click="onSubmit(false)">交卷</el-button>
    </header>

    <div class="exam-body">
      <main class="main">
        <QuestionRenderer
          v-if="exam.questions[currentIdx]"
          :question="exam.questions[currentIdx]"
          :index="currentIdx"
          :answer="answers[exam.questions[currentIdx].id]"
          mode="do"
          @update:answer="(v: any) => onAnswer(exam.questions[currentIdx].id, v)"
        >
          <template #actions>
            <el-button link @click="toggleMark(currentIdx)">
              <el-icon><Flag v-if="marked.has(currentIdx)" /><FlagFilled v-else /></el-icon>
              {{ marked.has(currentIdx) ? '已标记' : '标记' }}
            </el-button>
            <FavoriteStar :target-type="1" :target-id="exam.questions[currentIdx].id" size="small" />
          </template>
        </QuestionRenderer>
        <div class="nav">
          <el-button :disabled="currentIdx === 0" @click="jumpTo(currentIdx - 1)">上一题</el-button>
          <el-button type="primary" :disabled="currentIdx === exam.questions.length - 1" @click="jumpTo(currentIdx + 1)">下一题</el-button>
        </div>
      </main>
      <aside class="aside koala-card">
        <h4>答题卡</h4>
        <div class="q-grid">
          <button
            v-for="(q, idx) in exam.questions"
            :key="q.id"
            :class="{
              active: idx === currentIdx,
              answered: answers[q.id] !== undefined && answers[q.id] !== '' && !(Array.isArray(answers[q.id]) && (answers[q.id] as any[]).length === 0),
              marked: marked.has(idx)
            }"
            @click="jumpTo(idx)"
          >{{ idx + 1 }}</button>
        </div>
        <div class="legend">
          <span><i class="dot answered"></i>已答</span>
          <span><i class="dot marked"></i>标记</span>
        </div>
        <el-button style="width:100%;margin-top:16px" type="warning" @click="onSubmit(false)">交卷</el-button>
      </aside>
    </div>
  </div>
</template>

<style scoped lang="scss">
.exam-room { padding: 16px; }
.exam-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
.exam-header .meta { color: #999; font-size: 12px; }
.timer { font-size: 24px; font-weight: 600; color: #409eff; }
.timer.danger { color: #f56c6c; }
.exam-body { display: grid; grid-template-columns: 1fr 280px; gap: 16px; }
.nav { display: flex; justify-content: space-between; margin-top: 16px; }
.aside h4 { margin: 0 0 12px; }
.q-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 8px; }
.q-grid button { padding: 6px; border: 1px solid #ddd; border-radius: 4px; background: #fff; cursor: pointer; }
.q-grid button.answered { background: #e1f3d8; border-color: #67c23a; color: #67c23a; }
.q-grid button.marked { background: #faecd8; border-color: #e6a23c; }
.q-grid button.active { border-color: #409eff; outline: 2px solid #409eff; }
.legend { margin-top: 12px; display: flex; gap: 12px; font-size: 12px; color: #666; }
.legend .dot { display: inline-block; width: 10px; height: 10px; border-radius: 2px; margin-right: 4px; }
.legend .answered { background: #67c23a; }
.legend .marked { background: #e6a23c; }
</style>
