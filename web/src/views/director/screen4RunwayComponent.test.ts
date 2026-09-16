import { describe, expect, it } from 'vitest'
import { compileTemplate, parse } from '@vue/compiler-sfc'
import source from './Screen4Runway.vue?raw'

const { descriptor } = parse(source)
const template = descriptor.template!.content
const styles = descriptor.styles.map(style => style.content).join('\n')

function extractCssBlock(source: string, atRule: string) {
  const start = source.indexOf(atRule)
  if (start === -1) return ''
  const openBrace = source.indexOf('{', start)
  if (openBrace === -1) return ''

  let depth = 0
  for (let index = openBrace; index < source.length; index++) {
    if (source[index] === '{') depth++
    if (source[index] === '}') depth--
    if (depth === 0) return source.slice(start, index + 1)
  }

  return ''
}

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
    expect(template).toContain('{{ truncateScreen4RunwayText(step.name, 25) }}')
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
    expect(template).toContain('rotate(${activeHop.angle}deg)')
    expect(template).not.toContain("activePoint.direction === 'right' ? 0 : 180")
    expect(template).not.toContain('translate(${activePoint.x} ${activePoint.y})')
    expect(template).toContain('d="M -30 -7 L -9 -7 L 0 0 L -9 7 L -30 7 Z"')
    expect(styles).toMatch(/\.runway-baton-anchor\s*\{[^}]*transition: transform \.55s/s)
  })

  it('animates completed task cards through a staged milestone absorption', () => {
    expect(source).toContain('defineExpose({ playTaskCompletions })')
    expect(source).toContain("flyer.className = 'runway-flyer'")
    expect(source).toContain("done.className = 'runway-fly-done'")
    expect(source).toContain("dot.className = 'runway-fly-trail'")
    expect(source).toContain("ring.className = 'runway-hub-shockwave'")
    expect(source).toContain("particle.className = 'runway-burst-particle'")
    expect(source).toContain('spawnRunwayDoneBanner')
    expect(source).toContain('triggerRunwayAbsorption')
    expect(styles).toContain('.runway-fly-done')
    expect(styles).toContain('.runway-hub-shockwave')
    expect(styles).toContain('.runway-burst-particle')
  })

  it('connects completed nodes with a continuous green energy rail', () => {
    expect(template).toContain('class="runway-complete-flow"')
    expect(source).toContain('getScreen4RunwayCompletedPathEnd')
    expect(source).toContain('isScreen4RunwayNodeCompleted')
    expect(styles).toContain('.runway-complete-flow')
    expect(styles).toMatch(/\.runway-complete-path\s*\{[^}]*stroke-width: 15;[^}]*drop-shadow\(0 0 7px rgba\(29, 236, 147, \.58\)\)/s)
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

  it('pauses decorative motion while the document is hidden', () => {
    expect(template).toContain("'is-motion-paused': motionPaused")
    expect(source).toContain("document.addEventListener('visibilitychange', syncMotionVisibility)")
    expect(source).toContain("document.removeEventListener('visibilitychange', syncMotionVisibility)")
    expect(source).toContain('if (!root || !dial || motionPaused.value || prefersReducedMotion()) return')
    expect(styles).toMatch(/\.screen4-runway\.is-motion-paused[\s\S]*\*::after[\s\S]*animation-play-state: paused !important;/)
  })

  it('keeps motion visible with bounded low-cost paint layers', () => {
    expect(styles).toMatch(/\.screen4-runway\s*\{[^}]*contain: paint style;/s)
    expect(styles).toMatch(/\.runway-svg\s*\{[^}]*contain: paint;/s)
    expect(styles).toMatch(/\.ticker-track\s*\{[^}]*contain: paint;/s)
    expect(styles).not.toMatch(/\.runway-deck\s*\{[^}]*contain: paint;/s)
    expect(styles).toMatch(/\.runway-complete-flow\s*\{[^}]*filter: none;/s)
    expect(styles).toMatch(/\.runway-active-path\s*\{[^}]*filter: none;/s)
    expect(styles).toMatch(/\.progress-head-core\s*\{[^}]*filter: none;/s)
    expect(styles).toMatch(/\.runway-baton\s*\{[^}]*filter: none;/s)
    expect(styles).toMatch(/\.runway-progress-head\s*\{[^}]*will-change: transform, opacity;/s)
    expect(styles).toMatch(/\.milestone-scan\s*\{[^}]*will-change: transform;/s)
    expect(extractCssBlock(styles, '@media (max-width: 1280px)')).not.toContain('animation: none !important')
  })

  it('shows row turns as a downward continuation', () => {
    expect(template).toContain('class="turn-indicator-arrow"')
    expect(template).toContain('d="M -7 -9 L 0 -2 L 7 -9 M -7 1 L 0 8 L 7 1"')
    expect(template).not.toContain("indicator.direction === 'right'")
  })

  it('places each phase name above its node and progress directly below', () => {
    expect(template).toContain('class="runway-node-label"')
    expect(template).toContain("item.labelLines.length > 1 ? -66 : -51")
    expect(template).toContain("item.labelLines.length > 1 ? 'M -24 27 L 24 27' : 'M -24 10 L 24 10'")
    expect(template).toContain('class="runway-node-count" transform="translate(0 54)"')
    expect(template).not.toContain('class="runway-node-name" y="49"')
    expect(template).not.toContain("item.labelLines.length > 1 ? 91 : 72")
    expect(styles).toMatch(/\.runway-node-name\s*\{[^}]*font-size: 18px;/s)
    expect(template).toContain('<title>{{ item.node.name }}</title>')
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
    expect(styles).toMatch(/\.runway-summary\s*\{[^}]*position: absolute;[^}]*top: 12px;[^}]*right: 12px;/s)
    expect(styles).toMatch(/\.runway-summary\s*\{[^}]*width: fit-content;[^}]*max-width: calc\(100% - 430px\);[^}]*height: 38px;[^}]*justify-content: flex-start;/s)
    // 摘要主标与数字与跑道节点名称（18px）同规格。
    expect(styles).toMatch(/\.summary-kicker\s*\{[^}]*font-size: 18px;/s)
    expect(styles).toMatch(/\.summary-kicker\s*\{[^}]*flex: 0 0 auto;/s)
    expect(styles).toMatch(/\.summary-progress[\s\S]*?strong\s*\{[^}]*font-size: 18px;/)
    expect(styles).toMatch(/\.summary-progress[\s\S]*?strong\s*\{[^}]*color: #58f0b6;[^}]*text-shadow: 0 0 10px rgba\(56, 231, 167, \.42\);/)
    expect(styles).toMatch(/\.summary-progress[\s\S]*?span\s*\{[^}]*font-size: 18px;/)
    expect(styles).toMatch(/\.runway-svg\s*\{[^}]*inset: 2px 4px 82px;/s)
    expect(styles).toMatch(/\.runway-svg\s*\{[^}]*transform: translateY\(clamp\(-54px, -5\.5vh, -34px\)\) scale\(1\.08\);/s)
    expect(template).not.toContain('runway-node-index')
    expect(template).not.toContain('String(item.index + 1)')
  })

  it('uses the inherited intranet-safe font stack for all runway text', () => {
    expect(styles).not.toMatch(/DIN Alternate|Arial Narrow|Courier New/)
    expect(styles).toMatch(/\.screen4-runway\s*\{[^}]*font-family: inherit;/s)
    expect(styles).toMatch(/\.summary-progress\s*\{[^}]*font-variant-numeric: tabular-nums;/s)
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
    expect(styles).toMatch(/\.runway-legend\s*\{[^}]*top: 12px;[^}]*left: 12px;/s)
    // 图例与右上整体进度同高，保证两个 HUD 锚点严格处于同一水平线。
    expect(styles).toMatch(/\.runway-legend\s*\{[^}]*height: 38px;/s)
    // 图例文字与跑道节点名称（18px）同规格，圆点随之放大保持光学平衡。
    expect(styles).toMatch(/\.legend-item[\s\S]*?i\s*\{[\s\S]*?width: 11px;[\s\S]*?height: 11px;/)
    expect(styles).toMatch(/\.legend-item\s*\{[^}]*font-size: 18px;/)
    expect(styles).toContain('&.is-done i { color: #38e7a7; background: currentColor; }')
    expect(styles).toContain('@keyframes legend-arrive')
    expect(template).not.toContain('<span><i class="is-issue"></i>异常</span>')
  })

  it('keeps unfinished tasks static with running tasks first', () => {
    expect(template).toContain('class="deck-cell deck-ticker"')
    expect(template).toContain('aria-label="当前环节任务列表"')
    expect(template).toContain('class="ticker-chip"')
    expect(template).toContain('v-for="step in visibleSteps"')
    expect(template).toContain(':data-step-id="step.id"')
    expect(template).toContain(':title="step.name"')
    expect(template).toContain('{{ truncateScreen4RunwayText(step.name, 25) }}')
    // 两行任务铭牌：上行任务名，下行操作人 + 状态徽章。
    expect(template).toContain('class="chip-body"')
    expect(template).toContain('class="chip-operator"')
    expect(template).toContain('{{ tickerOperatorText(step) }}')
    expect(source).toContain('truncateScreen4RunwayText,')
    expect(template).not.toContain('class="chip-copy"')
    expect(template).toContain('class="ticker-sequence"')
    // 详情标题展示当前环节的完成比例，卡片仍只展示待处理任务。
    expect(template).toContain('class="ticker-standby"')
    expect(template).toContain('任务详情')
    expect(template).toContain("{{ runningSteps.length ? tickerDoneCount : '—' }}")
    expect(template).toContain("{{ runningSteps.length || '—' }}")
    expect(template).toContain('<em>已完成</em>')
    expect(template).toContain('class="ticker-complete"')
    expect(template).toContain('v-else-if="allTasksCompleted"')
    expect(template).toContain('当前环节所有任务已完成')
    expect(source).toContain('allTasksCompleted')
    // 无运行环节但已有环节收束时同样宣告完成，避免收束宣告一闪而过。
    expect(source).toContain('props.nodes.some(isScreen4RunwayNodeCompleted)')
    expect(template).not.toContain('暂无进行中环节任务')
    expect(template).not.toContain('本环节任务已全部完成')
    expect(template).not.toContain('activeNode?.name')
    expect(source).toContain('runningSteps')
    expect(source).toContain('!isAbsorbedStatus(step.status)')
    expect(source).toContain("running: 0")
    expect(source).toContain("pending: 2")
    expect(source).toContain('tickerStatusPriority[left.status]')
    expect(source).toContain('const tickerDoneCount = computed')
    expect(template).not.toContain('tickerScrolling')
    expect(template).not.toContain('ticker-repeat-')
    expect(source).not.toContain('measureTicker')
    expect(styles).not.toContain('.ticker-track.is-scrolling')
    expect(styles).not.toContain('@keyframes ticker-scroll')
    expect(styles).toMatch(/\.ticker-track\s*\{[^}]*width: 100%;/s)
    expect(styles).toMatch(/\.ticker-sequence\s*\{[^}]*flex: 1 1 auto;/s)
    expect(styles).toMatch(/\.ticker-chip\s*\{[^}]*display: inline-flex;[^}]*border-radius: 12px;/s)
    expect(styles).toMatch(/\.chip-name\s*\{[^}]*white-space: nowrap;/s)
    expect(styles).toContain('.ticker-standby')
    expect(styles).toMatch(/\.ticker-count\s*\{[^}]*display: inline-flex;/s)
    // 头部主标与计数与跑道节点名称（18px）同规格。
    expect(styles).toMatch(/\.ticker-node\s*\{[^}]*font-size: 18px;/s)
    expect(styles).toMatch(/\.ticker-count[\s\S]*?strong\s*\{[^}]*font-size: 18px;/)
    expect(styles).toMatch(/\.ticker-complete\s*\{[^}]*font-size: 18px;/s)
    expect(styles).toContain('@keyframes standby-drift')
    expect(styles).toContain('@keyframes standby-bead')
    // 完成宣告：分隔线右侧的绿色对勾徽记，与完成链路同色系。
    expect(styles).toContain('.ticker-complete')
    expect(styles).not.toContain('.ticker-head.is-complete')
    expect(styles).not.toContain('.ticker-empty')
  })

  it('flies finished tasks into the milestone dial with an absorb pulse', () => {
    expect(template).toContain('class="milestone-dial"')
    expect(source).toContain('playTaskCompletions')
    expect(source).toContain('triggerRunwayAbsorption')
    expect(source).toContain('flyer.animate')
    expect(source).toContain('pulseMilestoneDial')
    expect(styles).toContain('.milestone-dial.is-absorbing')
    expect(styles).toContain('@keyframes milestone-absorb')
  })

  it('advances the milestone progress only after a finished card is absorbed', () => {
    expect(source).toContain('const sourcePhaseProgress = computed')
    expect(source).toContain('const displayedPhaseCompleted = ref(sourcePhaseProgress.value.completed)')
    expect(source).toContain('pendingCompletionAnimations += launches.length')
    expect(source).toContain('if (pendingCompletionAnimations === 0) displayedPhaseCompleted.value = next.completed')

    const absorptionIndex = source.indexOf('triggerRunwayAbsorption(root, hubX, hubY)')
    const progressIndex = source.indexOf('commitMilestoneProgress()', absorptionIndex)
    const pulseIndex = source.indexOf('pulseMilestoneDial()', absorptionIndex)
    expect(absorptionIndex).toBeGreaterThan(-1)
    expect(progressIndex).toBeGreaterThan(absorptionIndex)
    expect(pulseIndex).toBeGreaterThan(progressIndex)
  })

  it('guards queued absorb flyers against hidden or stale roots', () => {
    expect(source).toContain('if (motionPaused.value || rootRef.value !== root) return')
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
