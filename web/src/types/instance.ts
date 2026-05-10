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
  template_name: string
  name: string
  description: string
  status: DrillStatus
  created_by: number
  created_by_name: string
  current_step_index: number
  total_steps: number
  completed_steps: number
  started_at?: string
  paused_at?: string
  completed_at?: string
  created_at: string
}

export interface StepInstance {
  id: number
  drill_id: number
  template_step_id: number
  step_name: string
  step_type: string
  status: StepStatus
  assignee_id?: number
  assignee_name?: string
  result_json?: string
  error_message?: string
  started_at?: string
  completed_at?: string
  timeout_seconds: number
  order_index: number
}

export interface StepLog {
  id: number
  step_instance_id: number
  action: 'complete' | 'issue' | 'force_complete' | 'skip' | 'timeout'
  operator_id: number
  operator_name: string
  comment: string
  created_at: string
}
