import { nextTick, type Component } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createTestApp, deferred, nodesWithClass, nodeText, withClientTemplate } from '@/test-utils/vueRenderer'
import AdminDashboard from './admin/DashboardView.vue'
import DirectorDashboard from './director/DashboardView.vue'
import ViewerDashboard from './viewer/DashboardView.vue'
import adminSource from './admin/DashboardView.vue?raw'
import directorSource from './director/DashboardView.vue?raw'
import viewerSource from './viewer/DashboardView.vue?raw'
import * as elementPlus from '@/test-utils/elementPlus'

const api = vi.hoisted(() => ({
  getTeam: vi.fn(), getMyTemplates: vi.fn(), getStepDuration: vi.fn(), getList: vi.fn(),
}))
vi.mock('@/api/modules/dashboard', () => ({ dashboardApi: api }))
vi.mock('@/api/modules/drill', () => ({ drillApi: { getList: api.getList } }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }) }))
vi.mock('element-plus', () => ({ ElMessage: { error: vi.fn() } }))
vi.mock('element-plus/es', async () => ({ ...await import('@/test-utils/elementPlus'), ElMessage: { error: vi.fn() } }))
vi.mock('element-plus/es/components/row/style/css', () => ({}))
vi.mock('element-plus/es/components/base/style/css', () => ({}))
vi.mock('element-plus/es/components/col/style/css', () => ({}))
vi.mock('element-plus/es/components/card/style/css', () => ({}))
vi.mock('element-plus/es/components/table/style/css', () => ({}))
vi.mock('element-plus/es/components/table-column/style/css', () => ({}))
vi.mock('element-plus/es/components/tag/style/css', () => ({}))
vi.mock('element-plus/es/components/progress/style/css', () => ({}))
vi.mock('element-plus/es/components/button/style/css', () => ({}))
vi.mock('element-plus/es/components/icon/style/css', () => ({}))
vi.mock('element-plus/es/components/tooltip/style/css', () => ({}))
vi.mock('element-plus/es/components/message/style/css', () => ({}))

describe('dashboard metric cards', () => {
  let unmount = () => {}
  beforeEach(() => {
    vi.useFakeTimers()
    Object.values(api).forEach(fn => fn.mockReset())
    api.getList.mockResolvedValue({ list: [], total: 6 })
  })
  afterEach(() => { unmount(); vi.useRealTimers(); vi.restoreAllMocks() })

  function mountDashboard(component: Component) {
    const source = component === AdminDashboard ? adminSource : component === DirectorDashboard ? directorSource : viewerSource
    const harness = createTestApp(withClientTemplate(component, source))
    Object.entries(elementPlus).forEach(([name, child]) => harness.app.component(name, child))
    const vm = harness.mount()
    unmount = () => harness.app.unmount()
    return {
      statValues: () => nodesWithClass(harness.root, 'stat-value').map(nodeText),
      refresh: () => (vm.$ as unknown as { setupState: { loadDashboard: () => Promise<void> } }).setupState.loadDashboard(),
    }
  }

  it.each([
    ['admin', AdminDashboard, 'getTeam', { team_online_count: 0, team_total_count: 4 }, '0/4'],
    ['director', DirectorDashboard, 'getMyTemplates', { my_template_count: 0 }, '0'],
    ['viewer', ViewerDashboard, 'getStepDuration', { avg_step_duration_seconds: 0 }, '0s'],
  ] as const)('%s renders a loading placeholder then genuine backend zero', async (_role, component, method, data, display) => {
    const response = deferred<unknown>()
    api[method].mockReturnValueOnce(response.promise)
    const dashboard = mountDashboard(component)
    expect(dashboard.statValues()[3]).toBe(method === 'getTeam' ? '—/—' : '—')
    expect(api[method]).toHaveBeenCalledTimes(1)
    response.resolve(data)
    await nextTick()
    await nextTick()
    expect(dashboard.statValues()[3]).toBe(display)
    expect(dashboard.statValues().slice(0, 3)).toEqual(method === 'getMyTemplates' ? ['0', '0', '0%'] : ['6', '0', '0%'])
  })

  it('shows unknown online count independently from a known team total', async () => {
    api.getTeam.mockResolvedValue({ team_online_count: null, team_total_count: 8 })
    const dashboard = mountDashboard(AdminDashboard)
    await nextTick()
    await nextTick()
    expect(dashboard.statValues()[3]).toBe('—/8')
  })

  it('shows no-sample duration as unknown and formats completed-step average', async () => {
    api.getStepDuration.mockResolvedValueOnce({ avg_step_duration_seconds: null })
    const dashboard = mountDashboard(ViewerDashboard)
    await nextTick()
    await nextTick()
    expect(dashboard.statValues()[3]).toBe('—')
    api.getStepDuration.mockResolvedValueOnce({ avg_step_duration_seconds: 125 })
    await vi.advanceTimersByTimeAsync(60000)
    expect(dashboard.statValues()[3]).toBe('2m 5s')
  })

  it.each([
    [AdminDashboard, 'getTeam', { team_online_count: 3, team_total_count: 8 }, '3/8'],
    [DirectorDashboard, 'getMyTemplates', { my_template_count: 12 }, '12'],
    [ViewerDashboard, 'getStepDuration', { avg_step_duration_seconds: 30 }, '30s'],
  ] as const)('refreshes metrics independently of the drill list and recovers from errors', async (component, method, data, display) => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    api.getList.mockRejectedValue(new Error('drills unavailable'))
    api[method].mockRejectedValueOnce(new Error('metric unavailable'))
    const dashboard = mountDashboard(component)
    await nextTick()
    expect(dashboard.statValues()[3]).toBe(method === 'getTeam' ? '—/—' : '—')
    api[method].mockResolvedValueOnce(data)
    await dashboard.refresh()
    await nextTick()
    expect(api[method]).toHaveBeenCalledTimes(2)
    expect(dashboard.statValues()[3]).toBe(display)
  })
})
