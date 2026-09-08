import type { StepInstance } from '@/types/instance'

export interface Screen4PhaseStep {
  name: string
  stepNodes: StepInstance[]
}

export interface Screen4PendingTaskPanel {
  state: 'ready' | 'empty' | 'complete'
  phaseStepName: string
  tasks: StepInstance[]
  primaryTask: StepInstance | null
  queuedTasks: StepInstance[]
}

const EMPTY_PANEL: Screen4PendingTaskPanel = {
  state: 'empty',
  phaseStepName: '',
  tasks: [],
  primaryTask: null,
  queuedTasks: [],
}

const COMPLETE_PANEL: Screen4PendingTaskPanel = {
  state: 'complete',
  phaseStepName: '',
  tasks: [],
  primaryTask: null,
  queuedTasks: [],
}

export function buildScreen4PendingTaskPanel(
  phaseName: string,
  phaseSteps: Screen4PhaseStep[],
  isLeafStep: (step: StepInstance) => boolean,
): Screen4PendingTaskPanel {
  const candidates = phaseSteps
    .filter(item => item.name !== phaseName)
    .map(item => ({
      ...item,
      tasks: item.stepNodes.filter(isLeafStep),
    }))

  if (!candidates.some(item => item.tasks.length > 0)) return EMPTY_PANEL

  const target = candidates.find(item => item.tasks.some(task => task.status === 'running'))
    ?? candidates.find(item => item.tasks.some(task => task.status === 'pending'))
  if (!target) return COMPLETE_PANEL

  const tasks = target.tasks
    .filter(task => task.status === 'running' || task.status === 'pending')
    .sort((left, right) => (
      Number(right.status === 'running') - Number(left.status === 'running')
      || left.seq - right.seq
    ))

  return {
    state: 'ready',
    phaseStepName: target.name,
    tasks,
    primaryTask: tasks[0] ?? null,
    queuedTasks: tasks.slice(1),
  }
}
