import { STEP_STATUS_LABELS, type StepStatus } from '@/types/instance'

const actionStatuses: Record<string, StepStatus> = {
  complete: 'completed', force_complete: 'completed', step_complete: 'completed', step_completed: 'completed',
  skip: 'skipped', step_skip: 'skipped', step_skipped: 'skipped',
  timeout: 'timeout', step_timeout: 'timeout', issue: 'issue', step_issue: 'issue',
  start: 'running', step_start: 'running', step_started: 'running',
}

function getLogStepId(log: Record<string, unknown>): number {
  return Number(log.step_instance_id || log.StepInstanceID || log.step_id || log.task_instance_id)
}

export function getLatestTaskOutcomeLogs<T extends { source?: Record<string, unknown> }>(logs: T[]): T[] {
  const seenStepIds = new Set<number>()
  return logs.filter(log => {
    if (!log.source) return false
    const stepId = getLogStepId(log.source)
    const action = String(log.source.action || log.source.Action || '')
    const status = actionStatuses[action]
    if (!stepId || !status || status === 'running') return false
    if (seenStepIds.has(stepId)) return false
    seenStepIds.add(stepId)
    return true
  })
}

export function getLogPresentation(log: Record<string, unknown>, steps: { id: number; name: string }[]) {
  const stepId = getLogStepId(log)
  const step = steps.find(item => item.id === stepId)
  const action = String(log.action || log.Action || '')
  const tone = stepId ? actionStatuses[action] || '' : ''
  return {
    message: step?.name || String(log.step_name || log.content || log.Content || action),
    status: tone ? STEP_STATUS_LABELS[tone] : '',
    tone,
  }
}
