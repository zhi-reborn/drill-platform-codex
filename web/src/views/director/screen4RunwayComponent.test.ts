import { describe, expect, it } from 'vitest'
import { compileTemplate, parse } from '@vue/compiler-sfc'
import source from './Screen4Runway.vue?raw'

const { descriptor } = parse(source)
const template = descriptor.template!.content
const styles = descriptor.styles.map(style => style.content).join('\n')

describe('screen4 runway component', () => {
  it('compiles an accessible serpentine energy runway', () => {
    expect(compileTemplate({
      source: template,
      filename: 'Screen4Runway.vue',
      id: 'screen4-runway',
    }).errors).toEqual([])
    expect(template).toContain('aria-label="当前阶段环节能量跑道"')
    expect(template).toContain('class="runway-svg"')
    expect(template).toContain(':d="layout.trackPath"')
    expect(template).toContain('class="runway-node-count"')
    expect(template).toContain('里程碑')
  })

  it('owns its implementation and renders task names only inside the ticker', () => {
    expect(source).not.toContain('PhaseRing')
    expect(source).not.toContain('ScreenView2')
    expect(template).not.toContain('node-steps')
    expect(template).toContain('class="ticker-chip"')
    expect(template).toContain('{{ step.name }}')
  })

  it('advances inter-node progress with the running node step ratio', () => {
    expect(template).toContain('class="runway-active-progress"')
    expect(template).toContain('class="runway-active-progress-core"')
    expect(template).toContain('class="runway-progress-head"')
    expect(template).toContain('strokeDasharray: `${activeHop.strokeLength} 99999`')
    expect(template).toContain('translate(${activeHop.cursor.x}px, ${activeHop.cursor.y}px)')
    expect(source).toContain('getScreen4RunwayHopProgress')
    expect(source).toContain('const activeRatio = computed(() =>')
    expect(styles).toContain('.runway-active-progress')
    expect(styles).toMatch(/\.runway-active-progress-core\s*\{[^}]*stroke: #ffe29a;[^}]*stroke-width: 3;/s)
    expect(styles).toMatch(/\.runway-active-path\s*\{[^}]*opacity: \.24;/s)
    expect(styles).toContain('transition: stroke-dasharray')
    expect(styles).toContain('@keyframes progress-head-pulse')
  })

  it('moves the running arrow with the progress frontier', () => {
    expect(template).toContain('class="runway-baton-anchor"')
    expect(template).toContain('translate(${activeHop.cursor.x}px, ${activeHop.cursor.y}px)')
    expect(template).not.toContain('translate(${activePoint.x} ${activePoint.y})')
    expect(template).toContain('d="M -30 -7 L -9 -7 L 0 0 L -9 7 L -30 7 Z"')
    expect(styles).toMatch(/\.runway-baton-anchor\s*\{[^}]*transition: transform \.55s/s)
  })

  it('connects completed nodes with a continuous green energy rail', () => {
    expect(template).toContain('class="runway-complete-flow"')
    expect(source).toContain('getScreen4RunwayCompletedPathEnd')
    expect(source).toContain('isScreen4RunwayNodeCompleted')
    expect(styles).toContain('.runway-complete-flow')
    expect(styles).toMatch(/\.runway-complete-path\s*\{[^}]*stroke-width: 15;[^}]*drop-shadow\(0 0 8px rgba\(29, 236, 147, \.72\)\)/s)
    expect(styles).toMatch(/\.runway-complete-core\s*\{[^}]*stroke: #42f0a4;/s)
    expect(styles).toMatch(/\.runway-complete-flow\s*\{[^}]*stroke: #42f0a4;/s)
    expect(styles).toContain('@keyframes completed-energy-flow')
  })

  it('renders the completed chain as an energized cable with lit node cores', () => {
    // 导轨正中的灯芯与导轨共用路径、同步描入。
    expect(template).toContain('ref="completeCoreRef"')
    expect(template).toContain('class="runway-complete-core"')
    expect(styles).toContain('.runway-complete-core')
    // 完成节点核心点亮，读作链上导通的灯。
    expect(template).toContain('id="s4-node-lit-core"')
    expect(styles).toContain(".runway-node.is-completed .node-core")
    expect(styles).toContain("url('#s4-node-lit-core')")
    expect(styles).toContain('@keyframes node-orbit-spin')
  })

  it('energizes row turns once the completed chain has passed through them', () => {
    expect(template).toContain("'is-energized': indicator.energized")
    expect(source).toContain('energized: end !== null && end >= index + 1')
    expect(styles).toContain('.turn-indicator.is-energized')
  })

  it('provides distinct status colors and reduced motion', () => {
    expect(styles).toContain('.runway-node.is-completed')
    expect(styles).toContain('.runway-node.is-running')
    expect(styles).toContain('.runway-node.is-issue')
    expect(styles).toContain('@media (prefers-reduced-motion: reduce)')
  })

  it('shows row turns as a downward continuation', () => {
    expect(template).toContain('class="turn-indicator-arrow"')
    expect(template).toContain('d="M -7 -9 L 0 -2 L 7 -9 M -7 1 L 0 8 L 7 1"')
    expect(template).not.toContain("indicator.direction === 'right'")
  })

  it('places each phase name above its node and progress directly below', () => {
    expect(template).toContain('class="runway-node-label"')
    expect(template).toContain("item.labelLines.length > 1 ? -73 : -51")
    expect(template).toContain('class="runway-node-count" transform="translate(0 58)"')
    expect(template).not.toContain('class="runway-node-name" y="49"')
    expect(template).not.toContain("item.labelLines.length > 1 ? 91 : 72")
  })

  it('floats the overall summary at the upper right and leaves the ticker in the bottom deck', () => {
    const summaryIndex = template.indexOf('class="deck-cell deck-summary runway-summary"')
    const svgIndex = template.indexOf('class="runway-svg"')
    const deckIndex = template.indexOf('class="runway-deck"')

    expect(template).toContain('class="runway-deck"')
    expect(summaryIndex).toBeGreaterThan(-1)
    expect(summaryIndex).toBeLessThan(svgIndex)
    expect(deckIndex).toBeGreaterThan(svgIndex)
    expect(template.slice(deckIndex)).not.toContain('deck-summary')
    expect(styles).toContain('.runway-deck')
    expect(styles).toMatch(/\.runway-summary\s*\{[^}]*position: absolute;[^}]*top: 8px;[^}]*right: 8px;/s)
    expect(styles).toMatch(/\.runway-summary\s*\{[^}]*width: min\(210px,[^}]*height: 32px;/s)
    expect(styles).toMatch(/\.runway-svg\s*\{[^}]*inset: 2px 4px 60px;/s)
    expect(styles).toMatch(/\.runway-svg\s*\{[^}]*transform: translateY\(-4px\) scale\(1\.02\);/s)
    expect(template).not.toContain('runway-node-index')
    expect(template).not.toContain('String(item.index + 1)')
  })

  it('presents the overall progress without a redundant meter', () => {
    expect(template).toContain('<span>整体进度</span>')
    expect(template).toContain('{{ completedSteps }}')
    expect(template).toContain('/ {{ totalSteps }} 步骤')
    expect(template).toContain('`整体进度，${completedSteps}/${totalSteps} 步骤`')
    expect(template).not.toContain('class="summary-meter"')
    expect(template).not.toContain('progressPercent')
    expect(template).not.toContain('class="summary-percent"')
    expect(source).toContain('completedSteps: number')
    expect(source).toContain('totalSteps: number')
    expect(source).not.toContain('const progressPercent = computed(() =>')
    expect(styles).not.toContain('.summary-meter')
    expect(styles).not.toContain('.summary-percent')
  })

  it('anchors the phase percentage ring on the milestone at the runway end', () => {
    expect(template).toContain('class="runway-milestone"')
    expect(template).toContain('`当前阶段进度 ${phaseProgressPercent}%`')
    expect(template).toContain(':aria-valuenow="phaseProgressPercent"')
    expect(template).toContain('{{ phaseProgressPercent }}')
    expect(template).toContain('MILESTONE_CIRCUMFERENCE')
    expect(source).toContain('getScreen4PhaseStepProgress')
    expect(styles).toContain('.milestone-progress')
    expect(styles).toContain('.milestone-value')
    expect(styles).not.toContain('.runway-progress-hub')
  })

  it('floats the status legend as a HUD plate in the top-left corner', () => {
    expect(template).toContain('class="runway-legend"')
    expect(template).toContain('class="legend-kicker"')
    expect(template).toContain('class="legend-item is-done"')
    expect(template).not.toContain('class="deck-cell deck-legend"')
    expect(styles).toMatch(/\.runway-legend\s*\{[^}]*top: 10px;[^}]*left: 12px;/s)
    expect(styles).toMatch(/\.legend-item[\s\S]*?i\s*\{[\s\S]*?width: 9px;[\s\S]*?height: 9px;/)
    expect(styles).toMatch(/\.legend-item\s*\{[^}]*font-size: 13px;/)
    expect(styles).toContain('&.is-done i { color: #38e7a7; background: currentColor; }')
    expect(styles).toContain('@keyframes legend-arrive')
    expect(template).not.toContain('<span><i class="is-issue"></i>异常</span>')
  })

  it('streams only the unfinished tasks along the bottom ticker', () => {
    expect(template).toContain('class="deck-cell deck-ticker"')
    expect(template).toContain('class="ticker-chip"')
    expect(template).toContain('v-for="step in visibleSteps"')
    expect(template).toContain(':data-step-id="step.id"')
    expect(template).toContain('class="ticker-sequence"')
    expect(template).toContain('暂无进行中环节任务')
    expect(template).toContain('本环节任务已全部完成')
    expect(source).toContain('runningSteps')
    expect(source).toContain('!isAbsorbedStatus(step.status)')
    expect(source).toContain('measureTicker')
    expect(styles).toContain('.ticker-track.is-scrolling')
    expect(styles).toContain('@keyframes ticker-scroll')
    expect(styles).toMatch(/\.ticker-empty\s*\{[^}]*padding-left: 18px;/s)
  })

  it('flies finished tasks into the milestone dial with an absorb pulse', () => {
    expect(template).toContain('class="milestone-dial"')
    expect(source).toContain('absorbedStepIds')
    expect(source).toContain('launchAbsorbFlyers')
    expect(source).toContain('flyer.animate')
    expect(source).toContain('pulseMilestoneDial')
    expect(styles).toContain('.milestone-dial.is-absorbing')
    expect(styles).toContain('@keyframes milestone-absorb')
  })

  it('spins the milestone scan ring and hides the pending helper text', () => {
    expect(template).toContain('<text v-if="phaseComplete" class="milestone-status" y="104">阶段达成</text>')
    expect(template).not.toContain("phaseComplete ? '阶段达成' : '等待到达'")
    expect(styles).toContain('animation: milestone-scan-spin')
    expect(styles).toContain('@keyframes milestone-scan-spin')
  })

  it('draws the milestone as a progress dial with its title below the rings', () => {
    expect(template).toContain('<circle class="milestone-scan" r="64" />')
    expect(template).toContain('<circle class="milestone-track" r="46" />')
    expect(template).toContain('class="milestone-progress"')
    expect(template).toContain('strokeDasharray: `${milestoneArcLength} ${MILESTONE_CIRCUMFERENCE}`')
    expect(template).toContain('<circle class="milestone-plate" r="36" />')
    expect(template).toContain('class="milestone-spark"')
    expect(template).toContain('<text class="milestone-title" y="82">里程碑</text>')
    expect(styles).toContain('transition: stroke-dasharray')
  })

  it('emphasizes the completed and total counts below each node', () => {
    expect(template).toContain('<rect x="-39" y="-14" width="78" height="28" rx="14" />')
    expect(styles).toContain('font-size: 17px')
  })

  it('keeps SVG positioning transforms out of group animations', () => {
    expect(styles).toMatch(/@keyframes node-arrive\s*\{\s*from \{ opacity: 0; \}\s*to \{ opacity: 1; \}\s*\}/)
    expect(styles).toMatch(/@keyframes baton-hover\s*\{\s*0%, 100% \{ opacity: \.62; \}\s*50% \{ opacity: 1; \}\s*\}/)
  })
})
