<script setup lang="ts">
import type { Question } from '@/types/entity'

const props = defineProps<{
  question: Question
  index: number
  answer: unknown
  mode?: 'do' | 'review'   // 做题 / 回顾
}>()

const emit = defineEmits<{
  (e: 'update:answer', value: unknown): void
  (e: 'mark', value: boolean): void
  (e: 'favorite'): void
}>()

function onSingle(optKey: string) {
  emit('update:answer', optKey)
}

function onMultiple(optKey: string) {
  const cur = Array.isArray(props.answer) ? [...(props.answer as string[])] : []
  const idx = cur.indexOf(optKey)
  if (idx >= 0) cur.splice(idx, 1); else cur.push(optKey)
  emit('update:answer', cur)
}

function onJudge(v: boolean) { emit('update:answer', v) }

function isChecked(optKey: string): boolean {
  if (props.question.type === 2 || props.question.type === 5) {
    return Array.isArray(props.answer) && (props.answer as string[]).includes(optKey)
  }
  return props.answer === optKey
}

function isCorrect(optKey: string): boolean | null {
  if (props.mode !== 'review' || !props.question.answer) return null
  const correct = Array.isArray(props.question.answer) ? props.question.answer : [props.question.answer]
  return correct.map(String).includes(optKey)
}
</script>

<template>
  <div class="q-card koala-card">
    <div class="q-head">
      <span class="q-no">{{ index + 1 }}.</span>
      <el-tag size="small" v-if="question.type === 1">单选</el-tag>
      <el-tag size="small" type="success" v-else-if="question.type === 2">多选</el-tag>
      <el-tag size="small" type="warning" v-else-if="question.type === 3">判断</el-tag>
      <el-tag size="small" type="info" v-else-if="question.type === 4">填空</el-tag>
      <el-tag size="small" type="danger" v-else-if="question.type === 5">不定项</el-tag>
      <el-tag size="small" type="primary" v-else>编程</el-tag>
      <span class="q-score">({{ question.score }}分)</span>
      <span style="flex:1"></span>
      <slot name="actions" />
    </div>
    <div class="q-title" v-html="question.title"></div>

    <div v-if="question.type === 1 || question.type === 2 || question.type === 5">
      <div v-for="opt in question.options" :key="opt.key" class="opt"
           :class="{ active: isChecked(opt.key), 'opt-correct': isCorrect(opt.key) === true, 'opt-wrong': isCorrect(opt.key) === false && isChecked(opt.key) }"
           @click="mode === 'do' && (question.type === 2 || question.type === 5 ? onMultiple(opt.key) : onSingle(opt.key))()">
        <el-checkbox v-if="question.type === 2 || question.type === 5" :model-value="isChecked(opt.key)" @change="onMultiple(opt.key)" />
        <el-radio v-else :model-value="isChecked(opt.key)" @change="onSingle(opt.key)" />
        <span class="opt-key">{{ opt.key }}.</span>
        <span>{{ opt.text }}</span>
      </div>
    </div>

    <div v-else-if="question.type === 3">
      <el-radio-group :model-value="answer as boolean" @change="onJudge">
        <el-radio :value="true">正确</el-radio>
        <el-radio :value="false">错误</el-radio>
      </el-radio-group>
    </div>

    <div v-else-if="question.type === 4">
      <el-input v-model="(answer as string)" placeholder="请输入答案" @input="(v: any) => emit('update:answer', v)" />
    </div>

    <div v-if="mode === 'review' && question.analysis" class="analysis">
      <strong>解析：</strong>{{ question.analysis }}
    </div>
  </div>
</template>

<style scoped lang="scss">
.q-card { margin-bottom: 16px; }
.q-head { display: flex; align-items: center; gap: 8px; }
.q-no { font-weight: 600; }
.q-score { color: #999; font-size: 12px; }
.q-title { padding: 12px 0; line-height: 1.7; }
.opt { display: flex; align-items: center; gap: 8px; padding: 10px 12px; border: 1px solid #eee; border-radius: 4px; margin-bottom: 8px; cursor: pointer; transition: all .2s; }
.opt:hover { background: #f5f7fa; }
.opt.active { border-color: #409eff; background: #ecf5ff; }
.opt.opt-correct { border-color: #67c23a; background: #f0f9eb; }
.opt.opt-wrong { border-color: #f56c6c; background: #fef0f0; }
.opt-key { font-weight: 600; width: 24px; }
.analysis { margin-top: 12px; padding: 12px; background: #fafafa; border-radius: 4px; color: #666; }
</style>
