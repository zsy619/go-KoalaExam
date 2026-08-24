import { defineStore } from 'pinia'
import { examApi, type StartExamResp } from '@/api/modules/exam'

// 考试进行中状态
export const useExamStore = defineStore('exam', {
  state: () => ({
    current: null as StartExamResp | null,
    answers: {} as Record<number, unknown>,
    startTimestamp: 0,
    remainingSeconds: 0,
  }),
  actions: {
    async start(examId: number) {
      const { data } = await examApi.start(examId)
      this.current = data!
      this.answers = {}
      this.startTimestamp = Date.now()
      this.remainingSeconds = data!.duration * 60
      return data!
    },
    saveAnswer(questionId: number, answer: unknown) {
      this.answers[questionId] = answer
    },
    tick() {
      if (this.remainingSeconds > 0) this.remainingSeconds--
    },
    reset() {
      this.current = null
      this.answers = {}
      this.startTimestamp = 0
      this.remainingSeconds = 0
    },
  },
})
