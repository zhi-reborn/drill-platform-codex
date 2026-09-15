import { describe, expect, it } from 'vitest'
import screenSource from './ScreenView.vue?raw'
import phaseRingSource from '../components/screen/PhaseRing.vue?raw'

describe('main command screen typography', () => {
  it('uses one intranet-safe Chinese font stack across the page and phase runway', () => {
    const source = `${screenSource}\n${phaseRingSource}`

    expect(screenSource).toContain("$font-ui: 'Microsoft YaHei', 'PingFang SC', 'Hiragino Sans GB', SimHei, sans-serif;")
    expect(screenSource).toContain('$font-display: $font-ui;')
    expect(screenSource).toContain('$font-mono: $font-ui;')
    expect(screenSource).toContain('$font-cn: $font-ui;')
    expect(phaseRingSource).toMatch(/\.phase-ring\s*\{[^}]*font-family: inherit;/s)
    expect(source).not.toMatch(/Consolas|Menlo|Monaco|Courier New|Segoe UI|Arial/)
  })
})
