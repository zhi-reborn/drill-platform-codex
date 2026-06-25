import type { WsMessage, NormalizedWsMessage, RawMessage } from './messageTypes'

type Handler = (message: NormalizedWsMessage) => void

// normalizeEnvelope maps the canonical backend envelope (type/data) to the
// NormalizedWsMessage shape that existing view code expects (event_type/payload).
// This keeps view code backward-compatible without requiring every view to
// migrate at once.
function normalizeEnvelope(raw: RawMessage): NormalizedWsMessage {
  const type = raw.type || raw.event_type || raw.event || ''
  const data = raw.data || raw.payload || {}
  return {
    id: raw.id || '',
    version: raw.version || 1,
    type,
    drill_id: raw.drill_id,
    user_id: raw.user_id,
    occurred_at: raw.occurred_at || '',
    data,
    // Legacy aliases for backward compatibility.
    event_type: type,
    payload: data,
    timestamp: raw.timestamp || (raw.occurred_at ? new Date(raw.occurred_at).getTime() : Date.now()),
  }
}

class WebSocketClient {
  private ws: WebSocket | null = null
  private url = ''
  private handlers: Map<string, Set<Handler>> = new Map()
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  public status: 'connecting' | 'connected' | 'disconnected' = 'disconnected'

  connect(token: string) {
    // 默认连接到 tasks 通道（个人通知）
    this.url = `/ws/tasks?token=${token}&t=${Date.now()}`
    this.connectWs()
  }

  private connectWs() {
    this.status = 'connecting'
    try {
      this.ws = new WebSocket(this.url)

      this.ws.onopen = () => {
        this.status = 'connected'
        // If this is a reconnect (not the initial connection), dispatch a
        // reconnect event so views can refetch drill state and resync after
        // the gap. This is the single realtime event path: the client missed
        // events while disconnected, so a full state refetch is required.
        if (this.reconnectAttempts > 0) {
          this.dispatchReconnect()
        }
        this.reconnectAttempts = 0
        this.startHeartbeat()
        this.notifyStatusChange()
      }

      this.ws.onmessage = (event) => {
        try {
          const parsed = JSON.parse(event.data)
          // The backend batches multiple messages into a JSON array frame
          // ([msg1, msg2, ...]) via json.Marshal([]WSMessage). Parse and
          // dispatch each message individually. A single-object frame is
          // also accepted for backward compatibility.
          const messages = Array.isArray(parsed) ? parsed : [parsed]
          for (const raw of messages) {
            if (raw && raw.type === 'pong') continue
            if (!raw || (typeof raw !== 'object')) continue
            const msg = normalizeEnvelope(raw)
            if (msg.type) this.dispatchMessage(msg)
          }
        } catch { /* ignored */ }
      }

      this.ws.onclose = () => {
        this.status = 'disconnected'
        this.stopHeartbeat()
        this.notifyStatusChange()
        this.scheduleReconnect()
      }

      this.ws.onerror = () => {
        this.status = 'disconnected'
        this.notifyStatusChange()
      }
    } catch {
      this.scheduleReconnect()
    }
  }

  private dispatchReconnect() {
    const msg = normalizeEnvelope({
      id: `reconnect-${Date.now()}`,
      version: 1,
      type: 'reconnect',
      occurred_at: new Date().toISOString(),
      data: {},
    })
    this.dispatchMessage(msg)
  }

  private notifyStatusChange() {
    const msg = normalizeEnvelope({
      type: 'connection_status',
      data: { status: this.status },
    })
    const handlers = this.handlers.get('connection_status') || new Set()
    for (const h of handlers) h(msg)
  }

  private dispatchMessage(message: NormalizedWsMessage) {
    const globalHandlers = this.handlers.get('*') || new Set()
    const channelHandlers = this.handlers.get(message.type) || new Set()
    for (const h of globalHandlers) h(message)
    for (const h of channelHandlers) h(message)
  }

  subscribe(eventType: string | '*', handler: Handler) {
    if (!this.handlers.has(eventType)) this.handlers.set(eventType, new Set())
    this.handlers.get(eventType)!.add(handler)
  }

  unsubscribe(eventType: string, handler: Handler) {
    const h = this.handlers.get(eventType)
    if (h) h.delete(handler)
  }

  send(data: Record<string, unknown>) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    }
  }

  private startHeartbeat() {
    this.heartbeatTimer = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.send({ event_type: 'ping', timestamp: Date.now() })
      }
    }, 30000)
  }

  private stopHeartbeat() {
    if (this.heartbeatTimer) clearInterval(this.heartbeatTimer)
    this.heartbeatTimer = null
  }

  private scheduleReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) return
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 16000)
    this.reconnectAttempts++
    this.reconnectTimer = setTimeout(() => {
      this.connectWs()
    }, delay)
  }

  disconnect() {
    this.stopHeartbeat()
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer)
    this.ws?.close()
    this.ws = null
    this.status = 'disconnected'
  }
}

export const wsClient = new WebSocketClient()
