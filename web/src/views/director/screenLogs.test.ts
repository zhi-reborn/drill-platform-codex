import { describe, expect, it } from 'vitest'
import { getLatestTaskOutcomeLogs, getLogPresentation } from './screenLogs'

describe('screen log presentation', () => {
  const steps = [{ id: 42, name: '防火墙规则准备', status: 'completed' }]

  it('resolves task names from the log step ID instead of generic operator text', () => {
    expect(getLogPresentation({ step_instance_id: 42, action: 'complete', content: '指挥组完成任务' }, steps))
      .toEqual({ message: '防火墙规则准备', status: '已完成', tone: 'completed' })
  })

  it.each([
    ['force_complete', '已完成', 'completed'],
    ['skip', '已跳过', 'skipped'],
    ['timeout', '已超时', 'timeout'],
    ['issue', '异常', 'issue'],
  ])('preserves the outcome of %s even if the task later changes', (action, status, tone) => {
    expect(getLogPresentation({ StepInstanceID: 42, Action: action }, steps))
      .toEqual({ message: '防火墙规则准备', status, tone })
  })

  it('keeps drill events and unavailable task details readable without inventing a task', () => {
    expect(getLogPresentation({ action: 'pause', content: '演练暂停' }, steps).message).toBe('演练暂停')
    expect(getLogPresentation({ step_instance_id: 99, action: 'complete', content: '指挥组完成任务' }, steps).message).toBe('指挥组完成任务')
  })

  it('keeps only completed or abnormal task outcomes and deduplicates each task', () => {
    const latest = { id: 3, source: { step_instance_id: 42, action: 'complete' } }
    const earlier = { id: 2, source: { step_instance_id: 42, action: 'start' } }
    const drillEvent = { id: 1, source: { action: 'pause' } }
    const issue = { id: 4, source: { step_instance_id: 43, action: 'issue' } }

    expect(getLatestTaskOutcomeLogs([latest, earlier, drillEvent, issue])).toEqual([latest, issue])
  })
})
