import { ref, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

/**
 * 考试倒计时 Composable
 * - 支持暂停/恢复
 * - 时间到自动触发回调
 */
export function useExamTimer(initialSeconds: number, onExpire: () => void) {
  const remaining = ref(initialSeconds)
  const running = ref(true)
  let timer: number | null = null

  function start() {
    if (timer) return
    running.value = true
    timer = window.setInterval(() => {
      if (!running.value) return
      if (remaining.value > 0) remaining.value--
      if (remaining.value === 0) {
        stop()
        ElMessage.warning('考试时间到，正在自动交卷...')
        onExpire()
      }
    }, 1000)
  }

  function pause() { running.value = false }
  function resume() { running.value = true }
  function stop() { if (timer) { clearInterval(timer); timer = null } }

  function format() {
    const h = Math.floor(remaining.value / 3600)
    const m = Math.floor((remaining.value % 3600) / 60)
    const s = remaining.value % 60
    const pad = (n: number) => n.toString().padStart(2, '0')
    return h > 0 ? `${pad(h)}:${pad(m)}:${pad(s)}` : `${pad(m)}:${pad(s)}`
  }

  onUnmounted(stop)
  return { remaining, running, start, pause, resume, stop, format }
}
