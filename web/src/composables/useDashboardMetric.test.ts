import { defineComponent, nextTick } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createTestApp, deferred } from '@/test-utils/vueRenderer'
import { useDashboardMetric } from './useDashboardMetric'

describe('useDashboardMetric', () => {
  let unmount = () => {}
  beforeEach(() => vi.useFakeTimers())
  afterEach(() => { unmount(); vi.useRealTimers() })

  async function mountMetric(fetchMetric: (signal: AbortSignal) => Promise<number | null>) {
    let metric!: ReturnType<typeof useDashboardMetric<number | null>>
    const harness = createTestApp(defineComponent({
      setup() { metric = useDashboardMetric(fetchMetric); return () => null },
    }))
    harness.mount()
    unmount = () => harness.app.unmount()
    return metric
  }

  it('starts empty, fetches on mount and keeps a genuine zero', async () => {
    const response = deferred<number>()
    const fetchMetric = vi.fn((_signal: AbortSignal) => response.promise)
    const metric = await mountMetric(fetchMetric)
    expect(metric.value.value).toBeNull()
    expect(fetchMetric).toHaveBeenCalledTimes(1)
    response.resolve(0)
    await nextTick()
    expect(metric.value.value).toBe(0)
  })

  it('refreshes manually and every 60 seconds, resetting loading and errors to null', async () => {
    const fetchMetric = vi.fn<() => Promise<number | null>>().mockResolvedValueOnce(7)
    const metric = await mountMetric(fetchMetric)
    await nextTick()
    expect(metric.value.value).toBe(7)
    const response = deferred<number>()
    fetchMetric.mockReturnValueOnce(response.promise)
    const refreshing = metric.refresh()
    expect(metric.value.value).toBeNull()
    response.reject(new Error('unavailable'))
    await refreshing
    expect(metric.value.value).toBeNull()
    fetchMetric.mockResolvedValueOnce(null)
    await vi.advanceTimersByTimeAsync(60000)
    expect(fetchMetric).toHaveBeenCalledTimes(3)
    expect(metric.value.value).toBeNull()
    fetchMetric.mockResolvedValueOnce(9)
    await vi.advanceTimersByTimeAsync(60000)
    expect(metric.value.value).toBe(9)
  })

  it('does not overlap requests or update after unmount', async () => {
    const response = deferred<number>()
    const fetchMetric = vi.fn((_signal: AbortSignal) => response.promise)
    const metric = await mountMetric(fetchMetric)
    await metric.refresh()
    await vi.advanceTimersByTimeAsync(120000)
    expect(fetchMetric).toHaveBeenCalledTimes(1)
    const signal = fetchMetric.mock.calls[0][0] as unknown as AbortSignal
    unmount()
    unmount = () => {}
    expect(signal.aborted).toBe(true)
    response.resolve(12)
    await nextTick()
    expect(metric.value.value).toBeNull()
    await metric.refresh()
    await vi.advanceTimersByTimeAsync(120000)
    expect(fetchMetric).toHaveBeenCalledTimes(1)
  })
})
