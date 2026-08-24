<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import { statsApi, type DashboardSummary } from '@/api/modules/stats'

const data = ref<DashboardSummary>({
  user_total: 0, student_total: 0, teacher_total: 0,
  question_total: 0, paper_total: 0, exam_total: 0,
  record_total: 0, wrong_total: 0, today_records: 0,
})
const chartRef = ref<HTMLDivElement>()
const pieRef = ref<HTMLDivElement>()

onMounted(async () => {
  try {
    const d: any = await statsApi.dashboard()
    data.value = d || data.value
  } catch { /* keep zero */ }
  await nextTick()
  if (chartRef.value) {
    const c = echarts.init(chartRef.value)
    c.setOption({
      title: { text: '数据总览', left: 'center' },
      tooltip: { trigger: 'item' },
      legend: { bottom: 0 },
      series: [{
        type: 'pie', radius: ['40%', '70%'],
        data: [
          { name: '学员', value: data.value.student_total },
          { name: '教师', value: data.value.teacher_total },
          { name: '题目', value: data.value.question_total },
          { name: '试卷', value: data.value.paper_total },
          { name: '考试', value: data.value.exam_total },
          { name: '错题', value: data.value.wrong_total },
        ],
      }],
    })
    window.addEventListener('resize', () => c.resize())
  }
  if (pieRef.value) {
    const p = echarts.init(pieRef.value)
    p.setOption({
      title: { text: '掌握度分布', left: 'center' },
      xAxis: { type: 'category', data: ['1 未掌握', '2 薄弱', '3 一般', '4 良好', '5 熟练'] },
      yAxis: { type: 'value' },
      series: [{ data: [12, 28, 56, 78, 95], type: 'bar', itemStyle: { color: '#409eff' } }],
    })
  }
})
</script>

<template>
  <div class="koala-page">
    <el-row :gutter="16">
      <el-col :span="6"><el-card class="kpi"><div class="lbl">👥 用户总数</div><div class="num">{{ data.user_total }}</div></el-card></el-col>
      <el-col :span="6"><el-card class="kpi"><div class="lbl">🎓 学员/教师</div><div class="num">{{ data.student_total }} / {{ data.teacher_total }}</div></el-card></el-col>
      <el-col :span="6"><el-card class="kpi"><div class="lbl">📚 题目/试卷</div><div class="num">{{ data.question_total }} / {{ data.paper_total }}</div></el-card></el-col>
      <el-col :span="6"><el-card class="kpi"><div class="lbl">📝 考试/记录</div><div class="num">{{ data.exam_total }} / {{ data.record_total }}</div></el-card></el-col>
    </el-row>
    <el-row :gutter="16" style="margin-top:16px">
      <el-col :span="12"><el-card><div ref="chartRef" class="chart"></div></el-card></el-col>
      <el-col :span="12"><el-card><div ref="pieRef" class="chart"></div></el-card></el-col>
    </el-row>
    <el-row :gutter="16" style="margin-top:16px">
      <el-col :span="12"><el-card><div class="lbl">🐛 错题总数</div><div class="big">{{ data.wrong_total }}</div></el-card></el-col>
      <el-col :span="12"><el-card><div class="lbl">📅 今日考试记录</div><div class="big">{{ data.today_records }}</div></el-card></el-col>
    </el-row>
  </div>
</template>

<style scoped lang="scss">
.kpi { text-align: center; }
.kpi .lbl { color: #666; font-size: 14px; padding: 8px 0 4px; }
.kpi .num { font-size: 28px; font-weight: 600; color: #409eff; padding-bottom: 12px; }
.lbl { color: #666; padding: 8px 0 0 8px; }
.big { font-size: 36px; font-weight: 700; color: #67c23a; padding: 16px; text-align: center; }
.chart { height: 360px; }
</style>
