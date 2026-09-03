import { afterEach, describe, expect, it } from 'vitest'
import { effectScope, nextTick, ref } from 'vue'
import { getFlowFocusIndex, getPhaseChamberPath, getPhaseFlowNodes, getPhaseStripScrollLeft, useScreenPhaseSelection } from './screenPhaseFlow'

const scopes: ReturnType<typeof effectScope>[] = []

describe('continuous phase chamber outline', () => {
  it('raises the board outline around the selected tab without drawing its bottom edge', () => {
    const path = getPhaseChamberPath(1000, 600, 120, { left: 40, right: 260, top: 10 })
    expect(path).toContain('H 34 Q 40 121 40 115 V 20 Q 40 10 50 10 H 250 Q 260 10 260 20 V 115 Q 260 121 266 121')
    expect(path.endsWith('Z')).toBe(true)
    expect(path).not.toContain('H 260 V 121 H 40')
  })

  it('moves the raised opening with the selected phase', () => {
    const first = getPhaseChamberPath(1000, 600, 120, { left: 40, right: 260, top: 10 })
    const second = getPhaseChamberPath(1000, 600, 120, { left: 300, right: 520, top: 10 })
    expect(second).not.toBe(first)
    expect(second).toContain('H 294 Q 300 121 300 115')
    expect(second).toContain('Q 520 121 526 121')
  })

  it('keeps clipped tabs within the board and falls back to a plain board when hidden', () => {
    const clipped = getPhaseChamberPath(680, 600, 120, { left: -40, right: 160, top: 10 })
    expect(clipped).toContain('Q 20 121 20 115')
    const plain = getPhaseChamberPath(680, 600, 120)
    expect(getPhaseChamberPath(680, 600, 120, { left: -240, right: -20, top: 10 })).toBe(plain)
    expect(getPhaseChamberPath(680, 600, 120, { left: 700, right: 920, top: 10 })).toBe(plain)
    expect(plain).not.toContain('V 20')
  })
})
afterEach(() => scopes.splice(0).forEach(scope => scope.stop()))

function setup(statuses = ['done', 'running', 'pending']) {
  const phases = ref(statuses.map((status, i) => ({ name: ['准备', '恢复', '验证'][i], status })))
  const drillId = ref(89)
  const scope = effectScope()
  scopes.push(scope)
  const selection = scope.run(() => useScreenPhaseSelection(phases, drillId))!
  return { phases, drillId, ...selection }
}

describe('screen phase selection', () => {
  it('selects the running phase when data arrives after loading', async () => {
    const { phases, selectedPhaseIdx } = setup([])
    expect(selectedPhaseIdx.value).toBe(-1)
    phases.value = [{ name: '准备', status: 'done' }, { name: '恢复', status: 'running' }]
    await nextTick()
    expect(selectedPhaseIdx.value).toBe(1)
  })

  it('allows future and completed phase previews without same-phase refresh stealing selection', async () => {
    const { phases, selectedPhaseIdx } = setup()
    expect(selectedPhaseIdx.value).toBe(1)
    for (const index of [2, 0]) {
      selectedPhaseIdx.value = index
      phases.value = phases.value.map(phase => ({ ...phase }))
      await nextTick()
      expect(selectedPhaseIdx.value).toBe(index)
    }
  })

  it('follows a genuinely new running phase once, but allows subsequent preview', async () => {
    const { phases, selectedPhaseIdx } = setup()
    selectedPhaseIdx.value = 0
    phases.value[1].status = 'done'
    phases.value[2].status = 'running'
    await nextTick()
    expect(selectedPhaseIdx.value).toBe(2)
    selectedPhaseIdx.value = 1
    phases.value = phases.value.map(phase => ({ ...phase }))
    await nextTick()
    expect(selectedPhaseIdx.value).toBe(1)
  })

  it('does not treat a paused then resumed phase as a new phase', async () => {
    const { phases, selectedPhaseIdx } = setup()
    selectedPhaseIdx.value = 2
    phases.value[1].status = 'pending'
    await nextTick()
    expect(selectedPhaseIdx.value).toBe(2)
    phases.value[1].status = 'running'
    await nextTick()
    expect(selectedPhaseIdx.value).toBe(2)
  })

  it('preserves selection by name when phases reorder and repairs a removed selection', async () => {
    const { phases, selectedPhaseIdx } = setup()
    selectedPhaseIdx.value = 2
    phases.value.reverse()
    await nextTick()
    expect(selectedPhaseIdx.value).toBe(0)
    phases.value.shift()
    await nextTick()
    expect(selectedPhaseIdx.value).toBe(0)
    expect(phases.value[selectedPhaseIdx.value].name).toBe('恢复')
  })

  it('falls back to first incomplete or last completed phase and ignores invalid clicks', () => {
    expect(setup(['done', 'pending', 'pending']).selectedPhaseIdx.value).toBe(1)
    const { selectedPhaseIdx } = setup(['done', 'done', 'done'])
    expect(selectedPhaseIdx.value).toBe(2)
    selectedPhaseIdx.value = 99
    expect(selectedPhaseIdx.value).toBe(2)
  })

  it('resets manual preview for a different drill even with identical phase names', async () => {
    const { drillId, selectedPhaseIdx } = setup()
    selectedPhaseIdx.value = 2
    drillId.value = 90
    await nextTick()
    expect(selectedPhaseIdx.value).toBe(1)
  })
})

describe('selected phase nodes', () => {
  const statusOf = (node: { status: string }) => node.status

  it('projects only this phase and excludes its header without truncating the thirteenth link', () => {
    const phase = {
      name: '准备',
      phaseSteps: [
        { name: '准备', status: 'running' },
        ...Array.from({ length: 14 }, (_, i) => ({ name: `准备环节${i + 1}`, status: i === 13 ? 'running' : 'done' })),
      ],
    }
    const nodes = getPhaseFlowNodes(phase, statusOf)
    expect(nodes).toHaveLength(14)
    expect(nodes[13]).toMatchObject({ name: '准备环节14', index: 14, status: 'running' })
    expect(getFlowFocusIndex(nodes)).toBe(13)
    expect(getPhaseFlowNodes({ name: '验证', phaseSteps: [{ name: '验证环节', status: 'pending' }] }, statusOf).map(n => n.name)).toEqual(['验证环节'])
  })

  it('never substitutes another phase for an empty phase', () => {
    expect(getPhaseFlowNodes<{ name: string; status: string }>(null, statusOf)).toEqual([])
    expect(getPhaseFlowNodes({ name: '空阶段', phaseSteps: [] }, statusOf)).toEqual([])
    expect(getPhaseFlowNodes({ name: '空阶段', phaseSteps: [{ name: '空阶段', status: 'pending' }] }, statusOf)).toEqual([])
  })

  it('focuses running, then first pending, then last completed, or no node', () => {
    expect(getFlowFocusIndex([{ status: 'pending' }, { status: 'running' }])).toBe(1)
    expect(getFlowFocusIndex([{ status: 'done' }, { status: 'pending' }])).toBe(1)
    expect(getFlowFocusIndex([{ status: 'done' }, { status: 'done' }])).toBe(1)
    expect(getFlowFocusIndex([])).toBe(-1)
  })
})

describe('selected card visibility after mount and resize', () => {
  it('reveals a late-loaded running phase beyond the initial viewport', () => {
    expect(getPhaseStripScrollLeft(0, 642, 685, 220)).toBe(263)
  })

  it('reveals the same selection when the viewport shrinks without an index change', () => {
    expect(getPhaseStripScrollLeft(0, 1000, 685, 220)).toBe(0)
    expect(getPhaseStripScrollLeft(0, 642, 685, 220)).toBe(263)
  })

  it('scrolls back for an earlier selection and preserves an already visible card', () => {
    expect(getPhaseStripScrollLeft(500, 642, 10, 220)).toBe(10)
    expect(getPhaseStripScrollLeft(263, 642, 460, 220)).toBe(263)
  })
})
