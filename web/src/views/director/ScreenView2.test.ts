import { describe, expect, it } from 'vitest'
import { compileTemplate, parse } from '@vue/compiler-sfc'
import source from './ScreenView2.vue?raw'

const { descriptor } = parse(source)
const template = descriptor.template!.content

describe('phase chamber template wiring', () => {
  it('compiles and places the stage strip and flow board inside a shared chamber', () => {
    expect(compileTemplate({ source: template, filename: 'ScreenView2.vue', id: 'screen2' }).errors).toEqual([])
    expect(template).toMatch(/class="phase-flow-chamber"[\s\S]*?class="phase-card-strip"[\s\S]*?class="flow-board"/)
  })

  it('provides keyboard-operable stage selection and selected-phase ownership', () => {
    expect(template).toMatch(/<button[\s\S]*?class="phase-card"[\s\S]*?:aria-pressed="index === selectedPhaseIdx"[\s\S]*?@click="selectPhase\(index\)"/)
    expect(template).toContain('{{ phase.name }}')
    expect(template).toContain('aria-controls="selected-phase-flow"')
    expect(template).toContain('id="selected-phase-flow"')
    expect(source).toContain('treeData.value[selectedPhaseIdx.value] ?? null')
    expect(source).toContain('getPhaseFlowNodes(currentPhaseData.value, getPhaseStepStatus')
    expect(template).toContain('阶段预览')
    expect(source).toContain("querySelector<HTMLElement>('.phase-card.active')")
    expect(source).not.toContain("querySelector<HTMLElement>('.phase-card.is-running')")
  })

  it('reloads for a new drill and guards against stale responses', () => {
    expect(source.includes('watch(drillId,')).toBe(true)
    expect(source.includes('if (requestId !== drillId.value) return')).toBe(true)
    expect(source.includes('socket !== ws')).toBe(true)
  })

  it('keeps chamber selection styles separate from individual stage status colors', () => {
    expect(template.includes(':class="\'phase-state-\' + selectedPhaseStatus"')).toBe(true)
  })

  it('uses one raised-tab outline and directional links instead of a detached light bridge', () => {
    expect(template).toContain('class="phase-chamber-surface"')
    expect(template).toContain(':d="chamberPath"')
    expect(template).toMatch(/v-if="index < phaseCards.length - 1"[\s\S]*?class="phase-sequence-arrow"/)
    expect(template).not.toContain('phase-flow-bridge')
    const styles = descriptor.styles.map(style => style.content).join('\n')
    expect(styles).toMatch(/\.phase-card\.active\s*\{[^}]*border-color: transparent/)
    expect(styles).toMatch(/\.phase-card\.active\s*\{[^}]*box-shadow: none/)
  })
})
