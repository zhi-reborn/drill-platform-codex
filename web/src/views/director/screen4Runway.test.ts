import { describe, expect, it } from 'vitest'
import {
  buildScreen4RunwayLayout,
  getScreen4PhaseStepProgress,
  getScreen4RunwayCompletedPathEnd,
  getScreen4RunwayHopProgress,
  getScreen4RunwayProgress,
  getScreen4TickerVisibleLimit,
  isScreen4RunwayNodeCompleted,
  splitScreen4RunwayName,
  truncateScreen4RunwayText,
} from './screen4Runway'

describe('screen4 runway geometry', () => {
  it('lays out a short runway on one forward lane with a milestone', () => {
    const layout = buildScreen4RunwayLayout(3)

    expect(layout.rowCount).toBe(1)
    expect(layout.points).toHaveLength(4)
    expect(layout.points.map(point => point.row)).toEqual([0, 0, 0, 0])
    expect(layout.points[0].x).toBeLessThan(layout.points[3].x)
  })

  it('folds a longer runway into alternating lanes', () => {
    const layout = buildScreen4RunwayLayout(7)

    expect(layout.rowCount).toBe(2)
    expect(layout.points).toHaveLength(8)
    expect(layout.points[4].direction).toBe('left')
    expect(layout.points[4].x).toBeGreaterThan(layout.points[5].x)
    expect(layout.trackPath).toContain('Q 998')
  })

  it('keeps a twelve-node runway readable by fitting it into three balanced lanes', () => {
    const layout = buildScreen4RunwayLayout(12)
    const laneSizes = layout.laneY.map((_, row) => (
      layout.points.filter(point => point.row === row).length
    ))

    expect(layout.rowCount).toBe(3)
    expect(laneSizes).toEqual([5, 4, 4])
    expect(layout.laneY[0]).toBeGreaterThanOrEqual(220)
    expect(layout.viewBoxHeight).toBeLessThanOrEqual(680)
    expect(layout.points[layout.points.length - 1]?.x).toBeLessThanOrEqual(760)
  })

  it('rounds each runway fold with horizontal and vertical tangent curves', () => {
    const layout = buildScreen4RunwayLayout(7)

    expect(layout.trackPath).toContain('L 970 240 Q 998 240 998 268')
    expect(layout.trackPath).toContain('L 998 372 Q 998 400 970 400')
  })

  it('calculates the selected phase percentage from its step totals', () => {
    expect(getScreen4PhaseStepProgress([
      { completed: 6, total: 6 },
      { completed: 0, total: 3 },
      { completed: 0, total: 3 },
      { completed: 0, total: 3 },
      { completed: 0, total: 4 },
      { completed: 0, total: 3 },
    ])).toEqual({ completed: 6, total: 22, percent: 27 })
  })

  it('counts terminal task states without exceeding the total', () => {
    expect(getScreen4RunwayProgress(['done', 'skipped', 'issue', 'running', 'pending']))
      .toEqual({ completed: 3, total: 5, ratio: 0.6 })
  })

  it('treats a node as completed as soon as all of its tasks finish', () => {
    expect(isScreen4RunwayNodeCompleted({ status: 'running', completed: 3, total: 3 })).toBe(true)
    expect(isScreen4RunwayNodeCompleted({ status: 'done', completed: 0, total: 3 })).toBe(true)
    expect(isScreen4RunwayNodeCompleted({ status: 'running', completed: 2, total: 3 })).toBe(false)
    expect(isScreen4RunwayNodeCompleted({ status: 'issue', completed: 2, total: 3 })).toBe(false)
    expect(isScreen4RunwayNodeCompleted({ status: 'issue', completed: 3, total: 3 })).toBe(false)
  })

  it('extends the completed energy rail through the next runway handoff point', () => {
    expect(getScreen4RunwayCompletedPathEnd([
      { status: 'done', completed: 6, total: 6 },
      { status: 'running', completed: 2, total: 3 },
      { status: 'pending', completed: 0, total: 3 },
    ])).toBe(1)
    expect(getScreen4RunwayCompletedPathEnd([
      { status: 'done', completed: 6, total: 6 },
      { status: 'running', completed: 3, total: 3 },
      { status: 'pending', completed: 0, total: 3 },
    ])).toBe(2)
    expect(getScreen4RunwayCompletedPathEnd([
      { status: 'running', completed: 0, total: 3 },
    ])).toBeNull()
  })

  it('fills the running hop in proportion to its step completion', () => {
    const layout = buildScreen4RunwayLayout(3)

    expect(getScreen4RunwayHopProgress(layout.points, 0, 1, 0.25)).toEqual({
      strokeLength: 60,
      cursor: { x: 220, y: 240 },
      angle: 0,
    })
    expect(getScreen4RunwayHopProgress(layout.points, 2, 3, 2 / 3)).toEqual({
      strokeLength: 160,
      cursor: { x: 800, y: 240 },
      angle: 0,
    })
  })

  it('parks the cursor on the running node and clamps the ratio', () => {
    const layout = buildScreen4RunwayLayout(3)

    expect(getScreen4RunwayHopProgress(layout.points, 0, 1, 0)).toEqual({
      strokeLength: 0,
      cursor: { x: 160, y: 240 },
      angle: 0,
    })
    expect(getScreen4RunwayHopProgress(layout.points, 0, 1, 1)?.cursor).toEqual({ x: 400, y: 240 })
    expect(getScreen4RunwayHopProgress(layout.points, 0, 1, 3)?.strokeLength).toBe(240)
    expect(getScreen4RunwayHopProgress(layout.points, 0, 1, -1)?.strokeLength).toBe(0)
    expect(getScreen4RunwayHopProgress(layout.points, 4, 5, 1)).toBeNull()
  })

  it('walks through lane turns along the rendered curve', () => {
    const layout = buildScreen4RunwayLayout(7)

    const mid = getScreen4RunwayHopProgress(layout.points, 3, 4, 0.5)
    expect(mid?.cursor).toEqual({ x: 998, y: 320 })
    expect(mid?.angle).toBe(90)
    const done = getScreen4RunwayHopProgress(layout.points, 3, 4, 1)
    expect(done?.cursor).toEqual({ x: 880, y: 400 })
    expect(done?.angle).toBe(180)
    expect(done?.strokeLength).toBeGreaterThan(370)
    expect(done?.strokeLength).toBeLessThan(380)
  })

  it('keeps runway labels to two concise lines', () => {
    expect(splitScreen4RunwayName('缓存与消息队列切换')).toEqual(['缓存与消息', '队列切换'])
    expect(splitScreen4RunwayName('非常非常长的环节节点名称需要截断')).toEqual(['非常非常长的', '环节节点名称…'])
  })

  it('truncates display text only after the configured limit', () => {
    expect(truncateScreen4RunwayText('123456789012')).toBe('123456789012')
    expect(truncateScreen4RunwayText('1234567890123')).toBe('123456789012…')
    expect(truncateScreen4RunwayText('一'.repeat(25), 25)).toBe('一'.repeat(25))
    expect(truncateScreen4RunwayText('一'.repeat(26), 25)).toBe(`${'一'.repeat(25)}…`)
  })

  it('adapts the ticker card count to the available width', () => {
    expect(getScreen4TickerVisibleLimit(0, 4)).toBe(1)
    expect(getScreen4TickerVisibleLimit(860, 2)).toBe(2)
    expect(getScreen4TickerVisibleLimit(860, 4)).toBe(2)
    expect(getScreen4TickerVisibleLimit(1160, 5)).toBe(3)
    expect(getScreen4TickerVisibleLimit(1600, 6)).toBe(5)
    expect(getScreen4TickerVisibleLimit(1600, 0)).toBe(0)
  })
})
