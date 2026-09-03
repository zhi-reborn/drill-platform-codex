import { onScopeDispose, watch } from 'vue'
import { authApi } from '@/api/modules/auth'

export function usePresenceHeartbeat(getToken: () => string) {
  let timer: ReturnType<typeof setInterval> | undefined
  let controller: AbortController | null = null
  let session = 0
  let disposed = false

  async function heartbeat() {
    if (disposed || !getToken() || controller) return
    const requestSession = session
    controller = new AbortController()
    try {
      await authApi.heartbeat(controller.signal)
    } catch {
      // Presence is best-effort; the request layer still handles authentication errors.
    } finally {
      controller = null
      if (!disposed && session !== requestSession && getToken()) void heartbeat()
    }
  }

  watch(getToken, (token) => {
    session += 1
    clearInterval(timer)
    controller?.abort()
    if (token) {
      void heartbeat()
      timer = setInterval(() => void heartbeat(), 60000)
    }
  }, { immediate: true })

  onScopeDispose(() => {
    disposed = true
    clearInterval(timer)
    controller?.abort()
  })
}
