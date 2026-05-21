export type DrillStatus = 'pending' | 'running' | 'paused' | 'completed' | 'terminated'
export type StepStatus = 'pending' | 'running' | 'completed' | 'timeout' | 'skipped' | 'issue'

export const DRILL_STATUS_LABELS: Record<DrillStatus, string> = {
  pending: '待启动',
  running: '执行中',
  paused: '已暂停',
  completed: '已完成',
  terminated: '已终止',
}

export const STEP_STATUS_LABELS: Record<StepStatus, string> = {
  pending: '待执行',
  running: '执行中',
  completed: '已完成',
  timeout: '已超时',
  skipped: '已跳过',
  issue: '异常',
}

export interface DrillInstance {
  id: number
  template_id: number
  template?: {
    id: number
    name: string
  }
  name: string
  description: string
  status: DrillStatus
  created_by: number
  created_by_name?: string
  current_step_id?: number
  progress_pct: number
  started_at?: string
  paused_at?: string
  completed_at?: string
  created_at: string
  updated_at?: string
  template_name?: string
  completed_steps?: number
  total_steps?: number
}

export function getCompletedSteps(steps: StepInstance[]): number {
  return steps.filter(s => s.status === 'completed' || s.status === 'skipped').length
}

export function getTotalSteps(steps: StepInstance[]): number {
  return steps.length || 1
}

export interface StepInstance {
  id: number
  drill_instance_id: number
  step_template_id: number
  parent_step_id?: number
  name: string
  seq: number
  status: StepStatus
  assignee_ids: string
  assignee_names?: string
  actual_operator: number | null
  start_time: string | null
  end_time: string | null
  timeout_at: string | null
  remark: string
  issue_desc: string
  step_type: string
  timeout_minutes: number
  default_assignee_role: string
  executor_team: string
  phase?: string
  phase_step?: string
  execution_mode?: 'serial' | 'parallel'
  estimated_duration_minutes?: number
  estimated_start_offset?: number
  task_name?: string
  sub_task?: string
  responsible_department?: string
  responsible_person?: string
  executor?: string
  reviewer?: string
  created_at: string
  drill_instance?: DrillInstance
  logs?: StepInstanceLog[]
}

export interface StepLog {
  id: number
  step_instance_id: number
  action: 'complete' | 'issue' | 'force_complete' | 'skip' | 'timeout' | 'step_start' | 'step_complete' | 'step_issue' | 'step_skip'
  operator_id: number
  operator_name?: string
  comment?: string
  remark?: string
  created_at: string
}

export interface StepInstanceLog {
  ID: number
  StepInstanceID: number
  Action: string
  OperatorID: number
  OperatorName: string
  Remark: string
  CreatedAt: string
}
