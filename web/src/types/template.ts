export type TemplateCategory = 'disaster_recovery' | 'degradation' | 'release' | 'security'
export type StepType = 'serial' | 'parallel' | 'any_of' | 'condition'

export const CATEGORY_LABELS: Record<TemplateCategory, string> = {
  disaster_recovery: '灾备切换',
  degradation: '服务降级',
  release: '发布演练',
  security: '安全事件',
}

export const STEP_TYPE_LABELS: Record<StepType, string> = {
  serial: '串行',
  parallel: '并行',
  any_of: '任意',
  condition: '条件',
}

export interface StepTemplate {
  id: number
  template_id: number
  parent_step_id?: number
  name: string
  description?: string
  step_type: StepType
  script?: string
  timeout_minutes?: number
  order_index: number
  pre_check?: string
  rollback_script?: string
  guide_content?: string
  default_assignee_role?: string
  executor_team?: string
  created_at: string
}

export interface DrillTemplate {
  id: number
  name: string
  category: TemplateCategory
  description: string
  version: string
  created_by: number
  created_by_name: string
  steps: StepTemplate[]
  status: 'draft' | 'published' | 'archived'
  status_label?: string
  created_at: string
  updated_at: string
}
