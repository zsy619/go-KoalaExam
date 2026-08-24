import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'

/**
 * 防作弊 Composable
 * - 切屏检测
 * - 复制粘贴拦截
 * - 全屏监控（可选）
 */
export function useAntiCheat(opts: {
  maxTabSwitch?: number
  blockCopy?: boolean
  onAudit: (event: { type: string; payload?: unknown }) => void
}) {
  const tabSwitchCount = ref(0)
  const max = opts.maxTabSwitch ?? Number(import.meta.env.VITE_MAX_TAB_SWITCH) || 3

  function onVisibilityChange() {
    if (document.hidden) {
      tabSwitchCount.value++
      opts.onAudit({ type: 'tab_switch', payload: { count: tabSwitchCount.value, ts: Date.now() } })
      ElMessage.warning(`切屏检测 ${tabSwitchCount.value}/${max}次`)
      if (tabSwitchCount.value > max) {
        opts.onAudit({ type: 'tab_switch_exceed', payload: { count: tabSwitchCount.value } })
      }
    }
  }

  function onCopy(e: ClipboardEvent) {
    if (!opts.blockCopy) return
    e.preventDefault()
    ElMessage.warning('考试中禁止复制')
    opts.onAudit({ type: 'copy_block', payload: { ts: Date.now() } })
  }

  function onPaste(e: ClipboardEvent) {
    if (!opts.blockCopy) return
    e.preventDefault()
    ElMessage.warning('考试中禁止粘贴')
    opts.onAudit({ type: 'paste_block', payload: { ts: Date.now() } })
  }

  function enterFullscreen() {
    const el = document.documentElement
    if (el.requestFullscreen) el.requestFullscreen().catch(() => {})
  }

  onMounted(() => {
    document.addEventListener('visibilitychange', onVisibilityChange)
    if (opts.blockCopy) {
      document.addEventListener('copy', onCopy)
      document.addEventListener('paste', onPaste)
    }
  })

  onUnmounted(() => {
    document.removeEventListener('visibilitychange', onVisibilityChange)
    document.removeEventListener('copy', onCopy)
    document.removeEventListener('paste', onPaste)
  })

  return { tabSwitchCount, enterFullscreen }
}
