export type NotificationType =
  | 'task_assigned'
  | 'step_complete'
  | 'step_timeout'
  | 'drill_started'
  | 'drill_paused'
  | 'drill_resumed'
  | 'drill_completed'
  | 'drill_terminated'
  | 'system_alert'

export const NOTIFICATION_TYPE_LABELS: Record<NotificationType, string> = {
  task_assigned: '任务分配',
  step_complete: '步骤完成',
  step_timeout: '步骤超时',
  drill_started: '演练开始',
  drill_paused: '演练暂停',
  drill_resumed: '演练恢复',
  drill_completed: '演练完成',
  drill_terminated: '演练终止',
  system_alert: '系统公告',
}

export interface Notification {
  id: number
  user_id: number
  type: NotificationType
  title: string
  content: string
  drill_id?: number
  drill_name?: string
  step_id?: number
  step_name?: string
  is_read: boolean
  created_at: string
}
