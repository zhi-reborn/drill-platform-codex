import { describe, expect, it } from 'vitest'
import {
  buildScreen4RunwayLayout,
  getScreen4RunwayProgress,
  splitScreen4RunwayName,
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

  it('counts terminal task states without exceeding the total', () => {
    expect(getScreen4RunwayProgress(['done', 'skipped', 'issue', 'running', 'pending']))
      .toEqual({ completed: 3, total: 5, ratio: 0.6 })
  })

  it('keeps runway labels to two concise lines', () => {
    expect(splitScreen4RunwayName('缓存与消息队列切换')).toEqual(['缓存与消息', '队列切换'])
    expect(splitScreen4RunwayName('非常非常长的环节节点名称需要截断')).toEqual(['非常非常长的', '环节节点名…'])
  })
})
