// 考试 Pinia Store
//
// 特性：
//   - 答题进度实时持久化（localStorage + 后端同步）
//   - 倒计时管理（自动提交）
//   - 防作弊事件统计
//   - 断线续考支持
import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'

export const useExamStore = defineStore('exam', () => {
  // 当前考试状态
  const currentExam = ref<any>(null)
  const currentRecord = ref<any>(null)
  const questions = ref<any[]>([])
  const answers = ref<Record<string, any>>({})
  const startTime = ref<number>(0)
  const endTime = ref<number>(0)
  const remaining = ref<number>(0)
  const auditLog = ref<any[]>([])

  // 防作弊统计
  const antiCheat = ref({
    switchTabCount: 0,
    fullscreenExit: 0,
    copyCount: 0,
    pasteCount: 0,
    devtoolsOpen: 0
  })

  // 持久化键
  const STORAGE_KEY = 'koala_exam_state'

  // 加载本地状态（断线续考）
  function loadFromStorage(recordID: number) {
    try {
      const data = localStorage.getItem(STORAGE_KEY + ':' + recordID)
      if (data) {
        const state = JSON.parse(data)
        answers.value = state.answers || {}
        auditLog.value = state.auditLog || []
        antiCheat.value = state.antiCheat || antiCheat.value
        return true
      }
    } catch {}
    return false
  }

  // 保存到本地
  function saveToStorage() {
    if (!currentRecord.value) return
    const data = {
      answers: answers.value,
      auditLog: auditLog.value,
      antiCheat: antiCheat.value
    }
    localStorage.setItem(STORAGE_KEY + ':' + currentRecord.value.id, JSON.stringify(data))
  }

  // 开始考试
  function start(exam: any, record: any, qs: any[]) {
    currentExam.value = exam
    currentRecord.value = record
    questions.value = qs
    startTime.value = Date.now()
    endTime.value = startTime.value + (exam.duration || 60) * 60 * 1000
    remaining.value = Math.max(0, endTime.value - Date.now())
    answers.value = {}
    auditLog.value = []
    antiCheat.value = {
      switchTabCount: 0, fullscreenExit: 0,
      copyCount: 0, pasteCount: 0, devtoolsOpen: 0
    }
    loadFromStorage(record.id)
  }

  // 设置答案
  function setAnswer(questionID: number, answer: any) {
    answers.value[questionID] = answer
    saveToStorage()
  }

  // 倒计时 tick
  function tick() {
    if (!endTime.value) return
    remaining.value = Math.max(0, endTime.value - Date.now())
  }

  // 是否已超时
  const isTimeout = computed(() => remaining.value <= 0)

  // 进度百分比
  const progress = computed(() => {
    if (questions.value.length === 0) return 0
    const answered = questions.value.filter((q) => answers.value[q.id] !== undefined).length
    return Math.round((answered / questions.value.length) * 100)
  })

  // 已答题数
  const answeredCount = computed(() =>
    questions.value.filter((q) => answers.value[q.id] !== undefined).length
  )

  // 记录防作弊事件
  function logAudit(eventType: string, data?: any) {
    const event = {
      type: eventType,
      ts: Date.now(),
      data
    }
    auditLog.value.push(event)
    saveToStorage()
  }

  // 记录切屏
  function onVisibilityChange(visible: boolean) {
    if (!visible) {
      antiCheat.value.switchTabCount++
      logAudit('switch_tab')
    }
  }

  // 记录全屏退出
  function onFullscreenExit() {
    antiCheat.value.fullscreenExit++
    logAudit('fullscreen_exit')
  }

  // 记录复制/粘贴
  function onCopy() {
    antiCheat.value.copyCount++
    logAudit('copy')
  }

  function onPaste() {
    antiCheat.value.pasteCount++
    logAudit('paste')
  }

  // 记录开发者工具
  function onDevtoolsOpen() {
    antiCheat.value.devtoolsOpen++
    logAudit('devtools_open')
  }

  // 是否作弊
  const isCheating = computed(() =>
    antiCheat.value.switchTabCount >= 3 ||
    antiCheat.value.fullscreenExit >= 2 ||
    antiCheat.value.devtoolsOpen >= 1
  )

  // 提交考试
  async function submit() {
    const payload = {
      record_id: currentRecord.value?.id,
      answers: answers.value,
      audit_log: auditLog.value,
      anti_cheat: antiCheat.value
    }
    // 清理本地
    if (currentRecord.value) {
      localStorage.removeItem(STORAGE_KEY + ':' + currentRecord.value.id)
    }
    return payload
  }

  // 重置
  function reset() {
    currentExam.value = null
    currentRecord.value = null
    questions.value = []
    answers.value = {}
    startTime.value = 0
    endTime.value = 0
    remaining.value = 0
    auditLog.value = []
    antiCheat.value = {
      switchTabCount: 0, fullscreenExit: 0,
      copyCount: 0, pasteCount: 0, devtoolsOpen: 0
    }
  }

  // 自动保存（防抖）
  let saveTimer: any = null
  watch(answers, () => {
    clearTimeout(saveTimer)
    saveTimer = setTimeout(() => saveToStorage(), 1000)
  }, { deep: true })

  return {
    // 状态
    currentExam, currentRecord, questions, answers,
    startTime, endTime, remaining, auditLog, antiCheat,
    // 计算
    isTimeout, progress, answeredCount, isCheating,
    // 方法
    start, setAnswer, tick, loadFromStorage, saveToStorage,
    logAudit, onVisibilityChange, onFullscreenExit,
    onCopy, onPaste, onDevtoolsOpen,
    submit, reset
  }
})
