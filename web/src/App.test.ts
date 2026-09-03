import { createPinia } from 'pinia'
import { nextTick } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createTestApp, deferred, withClientTemplate } from '@/test-utils/vueRenderer'
import { useAuthStore } from '@/stores/auth'
import App from './App.vue'
import appSource from './App.vue?raw'
import * as elementPlus from '@/test-utils/elementPlus'

const heartbeat = vi.hoisted(() => vi.fn<(signal: AbortSignal) => Promise<void>>())
vi.mock('@/api/modules/auth', () => ({ authApi: { heartbeat } }))
vi.mock('element-plus/es', () => import('@/test-utils/elementPlus'))
vi.mock('element-plus/es/components/config-provider/style/css', () => ({}))
vi.mock('element-plus/es/components/base/style/css', () => ({}))

describe('App presence heartbeat lifecycle', () => {
  let unmount = () => {}
  beforeEach(() => { vi.useFakeTimers(); heartbeat.mockReset(); heartbeat.mockResolvedValue(undefined) })
  afterEach(() => { unmount(); vi.useRealTimers() })

  function mountApp(token = '') {
    const pinia = createPinia()
    const auth = useAuthStore(pinia)
    auth.token = token
    const harness = createTestApp(withClientTemplate(App, appSource))
    harness.app.use(pinia)
    Object.entries(elementPlus).forEach(([name, component]) => harness.app.component(name, component))
    harness.app.component('RouterView', { render: () => null })
    harness.mount()
    unmount = () => harness.app.unmount()
    return auth
  }

  it('starts immediately when authenticated and repeats every 60 seconds', async () => {
    mountApp('initial-token')
    expect(heartbeat).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(59999)
    expect(heartbeat).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(1)
    expect(heartbeat).toHaveBeenCalledTimes(2)
  })

  it('starts after login and stops after logout', async () => {
    const auth = mountApp()
    await vi.advanceTimersByTimeAsync(60000)
    expect(heartbeat).not.toHaveBeenCalled()
    auth.token = 'logged-in'
    await nextTick()
    expect(heartbeat).toHaveBeenCalledTimes(1)
    auth.token = ''
    await nextTick()
    await vi.advanceTimersByTimeAsync(120000)
    expect(heartbeat).toHaveBeenCalledTimes(1)
  })

  it('does not overlap slow requests and restarts a new token after the old flight settles', async () => {
    const response = deferred<void>()
    heartbeat.mockReturnValueOnce(response.promise)
    const auth = mountApp('first')
    await vi.advanceTimersByTimeAsync(120000)
    expect(heartbeat).toHaveBeenCalledTimes(1)
    const oldSignal = heartbeat.mock.calls[0][0]
    auth.token = 'second'
    await nextTick()
    expect(oldSignal.aborted).toBe(true)
    expect(heartbeat).toHaveBeenCalledTimes(1)
    response.resolve(undefined)
    await nextTick()
    expect(heartbeat).toHaveBeenCalledTimes(2)
    expect(heartbeat.mock.calls[1][0].aborted).toBe(false)
  })

  it('aborts an in-flight heartbeat and clears its timer on unmount', async () => {
    const response = deferred<void>()
    heartbeat.mockReturnValueOnce(response.promise)
    mountApp('token')
    expect(heartbeat).toHaveBeenCalledTimes(1)
    const signal = heartbeat.mock.calls[0][0]
    unmount()
    unmount = () => {}
    expect(signal.aborted).toBe(true)
    response.resolve(undefined)
    await vi.advanceTimersByTimeAsync(120000)
    expect(heartbeat).toHaveBeenCalledTimes(1)
  })

  it('aborts on logout and ignores the late response without restarting', async () => {
    const response = deferred<void>()
    heartbeat.mockReturnValueOnce(response.promise)
    const auth = mountApp('token')
    const signal = heartbeat.mock.calls[0][0]
    auth.token = ''
    await nextTick()
    expect(signal.aborted).toBe(true)
    response.resolve(undefined)
    await vi.advanceTimersByTimeAsync(120000)
    expect(heartbeat).toHaveBeenCalledTimes(1)
  })

  it('continues its cadence without unhandled rejections after temporary failures', async () => {
    heartbeat.mockRejectedValue(new Error('presence unavailable'))
    mountApp('token')
    await vi.advanceTimersByTimeAsync(120000)
    expect(heartbeat).toHaveBeenCalledTimes(3)
  })
})
