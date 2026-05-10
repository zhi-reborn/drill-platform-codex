export interface Task {
  id: number
  drill_id: number
  drill_name: string
  step_id: number
  step_name: string
  step_description: string
  status: 'pending' | 'assigned' | 'in_progress' | 'completed' | 'issued' | 'skipped'
  assigned_to: number
  assigned_to_name: string
  deadline?: string
  script?: string
  result?: string
  error_message?: string
  created_at: string
  updated_at: string
}

export interface TaskAction {
  action: 'complete' | 'issue' | 'skip' | 'force_complete'
  result?: string
  error_message?: string
  comment: string
}
