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

  it('uses the business stage name as the card title without a numeric prefix', () => {
    expect(template).toContain('<span class="phase-name">{{ phase.name }}</span>')
    expect(template).not.toContain('class="phase-number"')
    expect(template).not.toContain('阶段{{ index + 1 }}')
    expect(template).toContain(':title="`查看${phase.name}`"')
  })

  it('counts only visible business phase steps and renders stable numeric hierarchy', () => {
    expect(source).toContain('const visiblePhaseSteps = phase.phaseSteps.filter(ps => ps.name !== phase.name)')
    expect(source).toContain("visiblePhaseSteps.filter(ps => getPhaseStepStatus(ps) === 'done').length")
    expect(source).toContain('const totalPhaseSteps = visiblePhaseSteps.length')
    expect(source).not.toContain('const totalPhaseSteps = phase.phaseSteps.length || 1')
    expect(template).toContain('class="stat-divider"')
    expect(template).toContain('class="stat-total"')

    const styles = descriptor.styles.map(style => style.content).join('\n')
    expect(styles).toMatch(/\.phase-stats\s*\{[^}]*font-variant-numeric:\s*tabular-nums/)
    expect(styles).toContain('.stat-divider')
    expect(styles).toContain('.stat-total')
  })

  it('reloads for a new drill and guards against stale responses', () => {
    expect(source.includes('watch(drillId,')).toBe(true)
    expect(source.includes('if (requestId !== drillId.value) return')).toBe(true)
    expect(source.includes('socket !== ws')).toBe(true)
  })

  it('presents the completed task name as the modal focus with its phase as context', () => {
    expect(source).toContain('getStepCompletionPresentation(payload, steps.value)')
    expect(template).toMatch(/class="completion-modal-content"[\s\S]*?role="status"[\s\S]*?aria-live="polite"/)
    expect(template).toContain('class="completion-task-plate"')
    expect(template).toContain('{{ completionModal.stepName }}')
    expect(template).toContain('所属环节')

    const styles = descriptor.styles.map(style => style.content).join('\n')
    const modalRule = styles.match(/\.completion-modal-content\s*\{([^}]*)\}/)?.[1]
    const stepRule = styles.match(/\.completion-step\s*\{([^}]*)\}/)?.[1]
    expect(modalRule).toMatch(/width:\s*min\(520px, calc\(100vw - 48px\)\)/)
    expect(modalRule).toMatch(/min-width:\s*0/)
    expect(stepRule).toMatch(/font-size:\s*clamp\(22px, 2\.2vw, 34px\)/)
    expect(stepRule).toMatch(/overflow-wrap:\s*anywhere/)
    expect(styles).toMatch(/\.completion-progress-bar\s*\{[^}]*animation:\s*progress-shrink 3s linear forwards/)
  })

  it('keeps chamber selection styles separate from individual stage status colors', () => {
    expect(template.includes(':class="\'phase-state-\' + selectedPhaseStatus"')).toBe(true)
  })

  it('highlights the running stage while keeping the running node label stable', () => {
    const styles = descriptor.styles.map(style => style.content).join('\n')
    expect(styles).toMatch(/\.phase-card\.is-running\.active\s*\{[^}]*border-color:\s*rgba\(255, 190, 92, 0\.96\)[^}]*box-shadow:/)

    const runningNodeRule = styles.match(/\.flow-node\.is-running \.node-tag\s*\{([^}]*)\}/)?.[1]
    expect(runningNodeRule).not.toMatch(/animation:/)
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

  it('renders restrained rail caps for virtual endpoints', () => {
    expect(template).toContain('class="rail-cap"')
    expect(template).toContain('<span class="virtual-name">开始</span>')
    expect(template).toContain('<span class="virtual-name">结束</span>')
    expect(template).toContain("index === flowNodes.length - 1 ? virtualArrowStyle('end') : arrowStyle(index)")
    expect(template).toContain(":class=\"['is-' + node.status, { 'is-virtual': index === flowNodes.length - 1 }]\"")
    expect(template).not.toContain('virtual-badge')
    expect(template).not.toContain('virtual-glyph')

    const styles = descriptor.styles.map(style => style.content).join('\n')
    expect(styles).toContain('.rail-cap')
    expect(styles).toContain('.rail-cap-core')
    expect(styles).toMatch(/\.flow-board\.all-done \.flow-node-wrap:not\(\.is-virtual\)/)
    expect(styles).toMatch(/\.flow-board\.all-done \.flow-node\.is-virtual-end \.rail-cap-core/)
    expect(styles).toMatch(/prefers-reduced-motion[\s\S]*?\.rail-cap-core/)
  })

  it('keeps arrow layout geometry stable while focus transitions', () => {
    const styles = descriptor.styles.map(style => style.content).join('\n')
    expect(styles).toMatch(
      /\.flow-arrow\s*\{[^}]*height:\s*4px;[^}]*transition:\s*opacity 0\.7s ease,\s*height 0\.45s ease;/,
    )
    expect(styles).toMatch(/\.flow-arrow\.is-running\s*\{[^}]*height:\s*5px;/)
  })

  it('shows up to six tasks in every node', () => {
    expect(template).toContain('v-for="step in getVisibleNodeSteps(node, NODE_STEP_LIMIT)"')
    expect(template).toContain('node.steps.length > NODE_STEP_LIMIT')
    expect(source).toContain('getVisibleNodeSteps')
    expect(source).toContain('const NODE_STEP_LIMIT = 6')
    expect(source).not.toContain('nodeStepLimit')
  })

  it('loads enough history to fill the task outcome log panel', () => {
    expect(source).toContain('drillApi.getLogs(drillId.value, 200)')
    expect(source).not.toContain('logData.slice(0, 50)')
    expect(source).toContain('const LOG_ROW_H = 24')

    const styles = descriptor.styles.map(style => style.content).join('\n')
    expect(styles).toMatch(/\.log-row\s*\{[^}]*min-height:\s*24px[^}]*padding:\s*2px 0/)
    expect(styles).toMatch(/\.flow-information \.log-row\s*\{[^}]*min-height:\s*24px[^}]*padding:\s*2px 0/)
  })

  it('keeps step lists in one column without scroll affordances', () => {
    const styles = descriptor.styles.map(style => style.content).join('\n')
    const listRule = styles.match(/\.node-steps\s*\{([^}]*)\}/)?.[1]

    expect(listRule).toMatch(/flex-direction:\s*column/)
    expect(template).not.toContain('node-steps-scroll-link')
    expect(template).not.toContain('is-scrollable')
    expect(styles).not.toContain('.flow-node.is-running .node-steps.is-scrollable')
  })

  it('reserves enough bottom breathing room for scaled rows', () => {
    const styles = descriptor.styles.map(style => style.content).join('\n')
    const viewportRule = styles.match(/\.flow-viewport\s*\{([^}]*)\}/)?.[1]
    expect(viewportRule).toContain('clamp(104px, 15.5vh, 132px)')
  })
})
