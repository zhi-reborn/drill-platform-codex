import { onMounted, onScopeDispose, shallowRef } from 'vue'

export function useDashboardMetric<T>(fetchMetric: (signal: AbortSignal) => Promise<T>) {
  const value = shallowRef<T | null>(null)
  let controller: AbortController | null = null
  let timer: ReturnType<typeof setInterval> | undefined
  let disposed = false

  async function refresh() {
    if (disposed || controller) return
    controller = new AbortController()
    value.value = null
    try {
      const result = await fetchMetric(controller.signal)
      if (!disposed) value.value = result
    } catch {
      if (!disposed) value.value = null
    } finally {
      controller = null
    }
  }

  onMounted(() => {
    void refresh()
    timer = setInterval(() => void refresh(), 60000)
  })

  onScopeDispose(() => {
    disposed = true
    clearInterval(timer)
    controller?.abort()
  })

  return { value, refresh }
}
