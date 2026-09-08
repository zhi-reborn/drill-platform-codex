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

  it('owns its implementation and never renders task names', () => {
    expect(source).not.toContain('PhaseRing')
    expect(source).not.toContain('ScreenView2')
    expect(template).not.toContain('node-steps')
    expect(template).not.toContain('step.name')
  })

  it('provides distinct status colors and reduced motion', () => {
    expect(styles).toContain('.runway-node.is-completed')
    expect(styles).toContain('.runway-node.is-running')
    expect(styles).toContain('.runway-node.is-issue')
    expect(styles).toContain('@media (prefers-reduced-motion: reduce)')
  })

  it('keeps SVG positioning transforms out of group animations', () => {
    expect(styles).toMatch(/@keyframes node-arrive\s*\{\s*from \{ opacity: 0; \}\s*to \{ opacity: 1; \}\s*\}/)
    expect(styles).toMatch(/@keyframes baton-hover\s*\{\s*0%, 100% \{ opacity: \.62; \}\s*50% \{ opacity: 1; \}\s*\}/)
  })
})
