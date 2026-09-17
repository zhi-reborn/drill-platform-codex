import { describe, expect, it } from 'vitest'
import screen4Source from './ScreenView4.vue?raw'
import monitorSource from './MonitorView.vue?raw'
import routerSource from '../../router/index.ts?raw'

describe('independent screen4 entry', () => {
  it('owns an independent page implementation', () => {
    expect(screen4Source).toContain('class="screen-root cyber-command-screen"')
    expect(screen4Source).toContain("from '@/api/modules/drill'")
    expect(screen4Source).not.toMatch(/import\s+ScreenView2|<ScreenView2/)
  })

  it('matches the primary screen header and runtime branding', () => {
    expect(screen4Source).toContain('class="screen-header"')
    expect(screen4Source).toContain('class="header-frame" viewBox="0 0 1200 82"')
    expect(screen4Source).toContain('class="header-title-block"')
    expect(screen4Source).toContain('class="drill-title">应急指挥中心</h1>')
    expect(screen4Source).toContain('class="btn-icon btn-fullscreen-mark"')
    expect(screen4Source).toContain('{{ screenBrand.fullscreenIconText }}')
    expect(screen4Source).toContain('{{ screenBrand.companyName }}')
    expect(screen4Source).toContain("fetch('/screen-brand.json', { cache: 'no-store' })")
    expect(screen4Source).not.toContain('class="command-header"')
    expect(screen4Source).not.toContain('class="header-title-shell"')
    expect(screen4Source).not.toContain('<FullScreen />')
    expect(screen4Source).toMatch(/\.screen-header\s*\{[\s\S]*?height:\s*74px;/)
    expect(screen4Source).toMatch(/\.screen-header\s*\{[\s\S]*?margin:\s*12px 18px 0;/)
    expect(screen4Source).toContain('grid-template-rows: 86px minmax(0, 1fr);')
    expect(screen4Source).toContain('@keyframes header-flow-to-center')
  })

  it('counts only visible business phase steps with its own presentation', () => {
    expect(screen4Source).toContain('const visiblePhaseSteps = phase.phaseSteps.filter(ps => ps.name !== phase.name)')
    expect(screen4Source).toContain("visiblePhaseSteps.filter(ps => getPhaseStepStatus(ps) === 'done').length")
    expect(screen4Source).toContain('const totalPhaseSteps = visiblePhaseSteps.length')
    expect(screen4Source).not.toContain('const totalPhaseSteps = phase.phaseSteps.length || 1')
    expect(screen4Source).toContain('class="stat-divider"')
    expect(screen4Source).toContain('class="stat-total"')
    expect(screen4Source).toMatch(/\.phase-stats\s*\{[^}]*font-variant-numeric:\s*tabular-nums/)
    expect(screen4Source).toContain('.stat-divider')
    expect(screen4Source).toContain('.stat-total')
  })

  it('uses a common local font stack without display-only numeric fonts', () => {
    expect(screen4Source).toContain('--screen4-font-family:')
    expect(screen4Source).toContain('"Microsoft YaHei"')
    expect(screen4Source).toContain('"PingFang SC"')
    expect(screen4Source).toContain('"SimHei"')
    expect(screen4Source).not.toMatch(/Courier New|DIN Alternate|Arial Narrow/)
  })

  it('renders its own runway instead of the old task-card flow', () => {
    expect(screen4Source).toContain("import Screen4Runway from './Screen4Runway.vue'")
    expect(screen4Source).toContain('<Screen4Runway')
    expect(screen4Source).not.toContain(':phase-name=')
    expect(screen4Source).toContain(':nodes="runwayNodes"')
    expect(screen4Source).toContain(':completed-steps="overallStepProgress.completed"')
    expect(screen4Source).toContain(':total-steps="overallStepProgress.total"')
    expect(screen4Source).toContain('const overallStepProgress = computed(() =>')
    expect(screen4Source).not.toContain('ref="flowTrackRef"')
    expect(screen4Source).not.toContain('class="node-steps"')
    expect(screen4Source).not.toContain("from '@/components/screen/PhaseRing.vue'")
  })

  it('sends completed task cards into the runway milestone without opening a modal', () => {
    expect(screen4Source).toContain('ref="runwayRef"')
    expect(screen4Source).toContain('runwayRef.value?.playTaskCompletions(completed)')
    expect(screen4Source).toContain("previous.status === 'completed' || step.status !== 'completed'")
    expect(screen4Source).toContain("{ flush: 'sync' }")
    expect(screen4Source).not.toContain('completionModal')
    expect(screen4Source).not.toContain('class="completion-modal"')
  })

  it('gives the runway full visual priority without a redundant board header', () => {
    expect(screen4Source).not.toContain('<header class="flow-board-heading"')
    expect(screen4Source).not.toContain('class="board-signal"')
    expect(screen4Source).not.toContain('.flow-board-heading')
    expect(screen4Source).toContain('padding: clamp(10px, 1.3vh, 14px) clamp(12px, 1.5vw, 28px) clamp(8px, 1vh, 14px)')
  })

  it('shows the configured operator name in runway task cards', () => {
    expect(screen4Source).toContain('operator: getScreen4StepOperatorName(s)')
    expect(screen4Source).toContain('step.operator_name = payload.executor')
    expect(screen4Source).not.toContain('if (payload.executor) step.assignee_names = payload.executor')
  })

  it('keeps the ambient scan narrow and compositor-friendly for software rendering', () => {
    const sweepStyles = screen4Source.match(/\.main-rect-sweep\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''
    const sweepKeyframes = screen4Source.match(/@keyframes rect-sweep-lite\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''
    expect(sweepStyles).toContain('width: 18px')
    expect(sweepStyles).toContain('contain: strict')
    expect(sweepStyles).toContain('translate3d(')
    expect(sweepStyles).toContain('10s linear infinite')
    expect(sweepStyles).toContain('will-change: transform')
    expect(sweepStyles).not.toContain('will-change: transform, opacity')
    expect(sweepStyles).not.toContain('box-shadow')
    expect(sweepStyles).not.toContain('repeating-linear-gradient')
    expect(sweepKeyframes).not.toContain('opacity')
  })

  it('removes the pending task area so the runway uses the released space', () => {
    expect(screen4Source).not.toContain("from './screen4PendingTasks'")
    expect(screen4Source).not.toContain('class="pending-task-panel"')
    expect(screen4Source).not.toContain('aria-label="当前环节待完成任务"')
    expect(screen4Source).not.toContain('pendingTaskPanel')
    expect(screen4Source).not.toContain('class="flow-brief"')
    expect(screen4Source).not.toContain('class="flow-log-panel"')
    expect(screen4Source).not.toContain("from './screenLogs'")
  })

  it('places the screen4 button after screen3 and opens the independent route', () => {
    expect(monitorSource).toMatch(/@click="viewScreen3"[\s\S]*?大屏3[\s\S]*?@click="viewScreen4"[\s\S]*?大屏4/)
    expect(monitorSource).toContain('window.open(`/director/screen4/${drillId.value}`, \'_blank\')')
    expect(monitorSource).toContain('.screen-entry-aqua')
  })

  it('routes screen4 to its own component', () => {
    expect(routerSource).toMatch(/path: 'screen4\/:id\(\\\\d\+\)'[\s\S]*?name: 'DirectorScreen4'[\s\S]*?ScreenView4\.vue/)
  })
})
