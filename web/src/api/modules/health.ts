import { apiRequest } from '../request'

// 组件检查结果（db/redis/publisher/subscriber）
export interface ComponentStatus {
  status: 'ok' | 'fail'
  latency_ms: number
}

// MySQL 连接池统计
export interface DBPoolStats {
  open: number
  in_use: number
  idle: number
  wait_count: number
  wait_duration_ms: number
}

// 单节点健康快照（与后端 health.Snapshot 对齐）
export interface HealthSnapshot {
  instance_id: string
  role: string
  ready: boolean
  worker_status: string
  components: Record<string, ComponentStatus>
  ws_connections?: Record<string, number>
  goroutines: number
  heap_alloc_bytes: number
  sys_bytes: number
  uptime_seconds: number
  started_at: number
  db_pool?: DBPoolStats
}

// 节点健康 = 快照 + 注册中心状态
export interface NodeHealth extends HealthSnapshot {
  heartbeat_at: number
  state: 'online' | 'offline'
}

// GET /v1/system/health 聚合响应
export interface SystemHealth {
  overall: 'healthy' | 'degraded'
  updated_at: number
  warning?: string
  nodes: NodeHealth[]
}

export const healthApi = {
  getSystemHealth: () =>
    apiRequest<SystemHealth>({ url: '/v1/system/health', method: 'GET' }),
}
