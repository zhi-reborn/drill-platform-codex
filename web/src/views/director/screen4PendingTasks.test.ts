import { describe, expect, it } from 'vitest'
import type { StepInstance } from '@/types/instance'
import { buildScreen4PendingTaskPanel, type Screen4PhaseStep } from './screen4PendingTasks'

function step(id: number, seq: number, status: StepInstance['status'], name = `任务${id}`): StepInstance {
  return {
    id,
    drill_instance_id: 90,
    step_template_id: id,
    name,
    seq,
    status,
    assignee_ids: '',
    actual_operator: null,
    start_time: null,
    end_time: null,
    timeout_at: null,
    remark: '',
    issue_desc: '',
    step_type: 'serial',
    timeout_minutes: 5,
    default_assignee_role: '',
    executor_team: '',
    created_at: '2026-09-08T00:00:00Z',
  }
}

function phaseStep(name: string, stepNodes: StepInstance[]): Screen4PhaseStep {
  return { name, stepNodes }
}

const isLeafStep = () => true

describe('screen4 pending task panel selection', () => {
  it('selects the running business phase step and excludes the phase placeholder', () => {
    const result = buildScreen4PendingTaskPanel('实施阶段', [
      phaseStep('实施阶段', [step(1, 1, 'running')]),
      phaseStep('故障注入执行', [step(11, 11, 'completed')]),
      phaseStep('主数据库切换', [
        step(21, 23, 'pending'),
        step(22, 21, 'running'),
        step(23, 22, 'pending'),
      ]),
    ], isLeafStep)

    expect(result).toMatchObject({
      state: 'ready',
      phaseStepName: '主数据库切换',
      primaryTask: { id: 22, status: 'running' },
    })
    expect(result.tasks.map(task => task.id)).toEqual([22, 23, 21])
    expect(result.queuedTasks.map(task => task.id)).toEqual([23, 21])
  })

  it('falls back to the first business phase step with pending work', () => {
    const result = buildScreen4PendingTaskPanel('实施阶段', [
      phaseStep('故障注入执行', [step(11, 1, 'completed')]),
      phaseStep('主数据库切换', [step(21, 2, 'pending')]),
      phaseStep('应用服务切换', [step(31, 3, 'pending')]),
    ], isLeafStep)

    expect(result.state).toBe('ready')
    expect(result.phaseStepName).toBe('主数据库切换')
    expect(result.primaryTask).toMatchObject({ id: 21, status: 'pending' })
  })

  it('filters terminal states from the selected phase step', () => {
    const result = buildScreen4PendingTaskPanel('实施阶段', [
      phaseStep('主数据库切换', [
        step(1, 1, 'completed'),
        step(2, 2, 'skipped'),
        step(3, 3, 'timeout'),
        step(4, 4, 'issue'),
        step(5, 5, 'pending'),
      ]),
    ], isLeafStep)

    expect(result.tasks.map(task => task.id)).toEqual([5])
  })

  it('keeps parallel running tasks before pending work in sequence order', () => {
    const result = buildScreen4PendingTaskPanel('实施阶段', [
      phaseStep('主数据库切换', [
        step(1, 4, 'pending'),
        step(2, 3, 'running'),
        step(3, 2, 'running'),
      ]),
    ], isLeafStep)

    expect(result.tasks.map(task => task.id)).toEqual([3, 2, 1])
    expect(result.queuedTasks.map(task => task.id)).toEqual([2, 1])
  })

  it('returns empty when the selected phase has no configured leaf task', () => {
    const result = buildScreen4PendingTaskPanel('实施阶段', [
      phaseStep('实施阶段', [step(1, 1, 'pending')]),
      phaseStep('主数据库切换', [step(2, 2, 'pending')]),
    ], current => current.id !== 2)

    expect(result).toEqual({
      state: 'empty',
      phaseStepName: '',
      tasks: [],
      primaryTask: null,
      queuedTasks: [],
    })
  })

  it('returns complete when all configured business tasks are terminal', () => {
    const result = buildScreen4PendingTaskPanel('实施阶段', [
      phaseStep('故障注入执行', [step(1, 1, 'completed')]),
      phaseStep('主数据库切换', [step(2, 2, 'skipped'), step(3, 3, 'issue')]),
    ], isLeafStep)

    expect(result).toEqual({
      state: 'complete',
      phaseStepName: '',
      tasks: [],
      primaryTask: null,
      queuedTasks: [],
    })
  })
})
