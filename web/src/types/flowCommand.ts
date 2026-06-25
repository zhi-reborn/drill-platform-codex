export interface FlowCommand {
  id: number
  status: 'pending' | 'processing' | 'succeeded' | 'failed'
  result?: unknown
  error_code?: string
  error_message?: string
}
