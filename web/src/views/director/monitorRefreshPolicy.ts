export type MonitorRefreshReason =
  | 'initial-load'
  | 'fallback-poll'
  | 'step-action-fallback'
  | 'step-patch-miss'
  | 'drill-event'

export interface MonitorRefreshPlan {
  detail: boolean
  steps: boolean
  logs: boolean
}

export function getMonitorRefreshPlan(reason: MonitorRefreshReason): MonitorRefreshPlan {
  switch (reason) {
    case 'initial-load':
      return { detail: true, steps: true, logs: true }
    case 'drill-event':
      return { detail: true, steps: false, logs: true }
    case 'fallback-poll':
    case 'step-action-fallback':
    case 'step-patch-miss':
      return { detail: false, steps: true, logs: true }
  }
}
