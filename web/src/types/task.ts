export interface TaskAction {
  action: 'complete' | 'issue' | 'skip' | 'force_complete'
  result?: string
  error_message?: string
  comment: string
}
