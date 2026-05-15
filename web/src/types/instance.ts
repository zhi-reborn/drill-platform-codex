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
  return steps.filter(s => s.status === 'completed').length
}

export function getTotalSteps(steps: StepInstance[]): number {
  return steps.length || 1
}

export interface StepInstance {
  id: number
  drill_instance_id: number
  step_template_id: number
  name: string
  seq: number
  status: StepStatus
  assignee_ids?: string
  actual_operator?: number
  start_time?: string
  end_time?: string
  timeout_at?: string
  remark?: string
  issue_desc?: string
  created_at: string
  logs?: StepInstanceLog[]
  drill_id?: number
  template_step_id?: number
  step_name?: string
  step_type?: string
  assignee_id?: number
  assignee_name?: string
  result_json?: string
  error_message?: string
  started_at?: string
  completed_at?: string
  timeout_seconds?: number
  order_index?: number
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
