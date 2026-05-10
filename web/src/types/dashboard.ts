export interface DashboardStats {
  total_drills: number
  active_drills: number
  today_completed: number
  success_rate: number
  avg_step_duration_seconds: number
  failure_rate: number
  team_online_count: number
  team_total_count: number
}

export interface DrillStatsByCategory {
  category: string
  count: number
  success_rate: number
}

export interface ActivityItem {
  id: number
  type: string
  drill_name: string
  step_name?: string
  operator: string
  created_at: string
}

export interface HourlyError {
  hour: string
  count: number
}

export interface DashboardData {
  stats: DashboardStats
  by_category: DrillStatsByCategory[]
  recent_activity: ActivityItem[]
  hourly_errors: HourlyError[]
}
