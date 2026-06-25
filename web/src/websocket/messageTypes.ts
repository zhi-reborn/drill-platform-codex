// WSMessage mirrors the canonical versioned event envelope shared by Redis
// pub/sub and WebSocket delivery on the backend:
//   {"id":"event-id","version":1,"type":"step.updated","drill_id":72,
//    "occurred_at":"2026-06-25T12:00:00Z","data":{}}
//
// The backend may batch multiple messages into one JSON array frame:
//   [msg1, msg2, ...]
// The WebSocket client splits these and dispatches each message individually.
export interface WsMessage {
  id: string
  version: number
  type: string
  drill_id?: number
  user_id?: number
  occurred_at: string
  data: Record<string, unknown>
}

// Legacy field aliases kept for backward compatibility with view code that
// still reads event_type / payload. The WebSocket client normalizes every
// incoming canonical envelope to also populate these fields.
export interface NormalizedWsMessage extends WsMessage {
  event_type: string
  payload: Record<string, unknown>
  timestamp: number
}

// RawMessage is the loose shape accepted from JSON.parse — it tolerates both
// the canonical envelope (type/data) and the legacy format (event_type/payload).
export interface RawMessage {
  id?: string
  version?: number
  type?: string
  event_type?: string
  event?: string
  drill_id?: number
  user_id?: number
  occurred_at?: string
  data?: Record<string, unknown>
  payload?: Record<string, unknown>
  timestamp?: number
}

export const WS_EVENTS = {
  STEP_START: 'step_start',
  STEP_COMPLETE: 'step_complete',
  STEP_TIMEOUT: 'step_timeout',
  DRILL_STARTED: 'drill_started',
  DRILL_PAUSED: 'drill_paused',
  DRILL_RESUMED: 'drill_resumed',
  DRILL_COMPLETED: 'drill_completed',
  DRILL_TERMINATED: 'drill_terminated',
  TASK_ASSIGNED: 'task_assigned',
  SYSTEM_ALERT: 'system_alert',
  PING: 'ping',
  PONG: 'pong',
  RECONNECT: 'reconnect',
} as const
