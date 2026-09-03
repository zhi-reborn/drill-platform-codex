import { describe, expect, it } from 'vitest'
import { createSSRApp } from 'vue'
import { renderToString } from 'vue/server-renderer'
import PhaseRing from './PhaseRing.vue'

describe('active phase progress', () => {
  it.each([
    { name: 'missing completed count', counts: { total: 4 }, offset: 100 },
    { name: 'missing total count', counts: { completed: 2 }, offset: 100 },
    { name: 'missing both counts', counts: {}, offset: 100 },
    { name: 'zero total', counts: { completed: 0, total: 0 }, offset: 100 },
    { name: 'partial completion', counts: { completed: 1, total: 4 }, offset: 75 },
    { name: 'negative completed count', counts: { completed: -1, total: 4 }, offset: 100 },
    { name: 'completed count above total', counts: { completed: 5, total: 4 }, offset: 0 },
  ])('renders a finite progress line for $name', async ({ counts, offset }) => {
    const html = await renderToString(createSSRApp(PhaseRing, {
      phases: ['准备'],
      phaseNames: [['检查']],
      phaseNodeStatuses: [[{ status: 'running', progress: 0, ...counts }]],
      currentIndex: 0,
      progress: 0,
      centerNumerator: 0,
      centerDenominator: 1,
      centerHint: '准备',
    }))

    expect(html).not.toContain('NaN')
    const activePath = html.match(/<path\b[^>]*class="lane-active-path"[^>]*>/)?.[0]
    expect(activePath).toContain(`stroke-dashoffset="${offset}"`)
  })
})
