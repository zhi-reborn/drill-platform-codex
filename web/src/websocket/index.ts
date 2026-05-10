import type { WsMessage } from './messageTypes'

type Handler = (message: WsMessage) => void

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
    this.url = `${import.meta.env.VITE_WS_URL}?token=${token}`
    this.connectWs()
  }

  private connectWs() {
    this.status = 'connecting'
    try {
      this.ws = new WebSocket(this.url)

      this.ws.onopen = () => {
        this.status = 'connected'
        this.reconnectAttempts = 0
        this.startHeartbeat()
        this.notifyStatusChange()
      }

      this.ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as WsMessage
          if (message.event_type === 'pong') return
          this.dispatchMessage(message)
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

  private dispatchMessage(message: WsMessage) {
    const globalHandlers = this.handlers.get('*') || new Set()
    const channelHandlers = this.handlers.get(message.event_type) || new Set()
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

  private notifyStatusChange() {
    const handlers = this.handlers.get('connection_status') || new Set()
    for (const h of handlers) h({ event_type: 'connection_status', payload: { status: this.status }, timestamp: Date.now() } as WsMessage)
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
