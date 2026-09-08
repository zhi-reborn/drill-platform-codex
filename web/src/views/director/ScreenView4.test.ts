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

  it('renders its own runway instead of the old task-card flow', () => {
    expect(screen4Source).toContain("import Screen4Runway from './Screen4Runway.vue'")
    expect(screen4Source).toContain('<Screen4Runway')
    expect(screen4Source).toContain(':phase-name="currentPhaseData?.name || \'当前阶段\'"')
    expect(screen4Source).toContain(':nodes="runwayNodes"')
    expect(screen4Source).not.toContain('ref="flowTrackRef"')
    expect(screen4Source).not.toContain('class="node-steps"')
    expect(screen4Source).not.toContain("from '@/components/screen/PhaseRing.vue'")
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
