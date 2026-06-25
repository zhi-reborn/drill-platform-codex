import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import type { WsMessage } from './messageTypes'

// Mock WebSocket for testing reconnection and batching behavior.
class MockWebSocket {
  static instances: MockWebSocket[] = []
  static OPEN = 1
  static CLOSED = 3

  url: string
  readyState = MockWebSocket.OPEN
  onopen: (() => void) | null = null
  onmessage: ((event: { data: string }) => void) | null = null
  onclose: (() => void) | null = null
  onerror: (() => void) | null = null

  constructor(url: string) {
    this.url = url
    MockWebSocket.instances.push(this)
  }

  send() {}
  close() {
    this.readyState = MockWebSocket.CLOSED
    this.onclose?.()
  }

  // Test helpers
  simulateOpen() { this.onopen?.() }
  simulateMessage(data: string) { this.onmessage?.({ data }) }
  simulateClose() { this.close() }
}

describe('WebSocket reconnect refetch', () => {
  let originalWebSocket: typeof WebSocket

  beforeEach(() => {
    originalWebSocket = globalThis.WebSocket
    MockWebSocket.instances = []
    ;(globalThis as any).WebSocket = MockWebSocket
  })

  afterEach(() => {
    ;(globalThis as any).WebSocket = originalWebSocket
    vi.restoreAllMocks()
  })

  it('dispatches a reconnect event when the connection is restored', async () => {
    vi.useFakeTimers()
    // We need to dynamically import the module so the MockWebSocket is in
    // place before the WebSocketClient constructor runs.
    const { wsClient } = await import('./index')
    wsClient.disconnect()

    const reconnectHandler = vi.fn()
    wsClient.connect('test-token')

    const socket = MockWebSocket.instances[MockWebSocket.instances.length - 1]
    socket.simulateOpen()

    // Subscribe to the reconnect event AFTER the initial connection so we
    // only observe reconnects, not the initial connect.
    wsClient.subscribe('reconnect', reconnectHandler)

    // Simulate a drop + reconnect cycle. scheduleReconnect uses setTimeout
    // with exponential backoff; advance fake timers to trigger it.
    socket.simulateClose()
    vi.advanceTimersByTime(2000)

    // scheduleReconnect created a new WebSocket; open it to trigger onopen.
    const socket2 = MockWebSocket.instances[MockWebSocket.instances.length - 1]
    if (socket2 !== socket) {
      socket2.simulateOpen()
    }

    expect(reconnectHandler).toHaveBeenCalledTimes(1)
    const msg = reconnectHandler.mock.calls[0][0] as WsMessage
    expect(msg.type).toBe('reconnect')
    vi.useRealTimers()
    wsClient.disconnect()
  })
})

describe('WebSocket JSON array batching', () => {
  let originalWebSocket: typeof WebSocket

  beforeEach(() => {
    originalWebSocket = globalThis.WebSocket
    MockWebSocket.instances = []
    ;(globalThis as any).WebSocket = MockWebSocket
  })

  afterEach(() => {
    ;(globalThis as any).WebSocket = originalWebSocket
    vi.restoreAllMocks()
  })

  it('parses a JSON array frame and dispatches each message individually', async () => {
    const { wsClient } = await import('./index')
    wsClient.disconnect()

    const handler = vi.fn()
    wsClient.connect('test-token')

    const socket = MockWebSocket.instances[MockWebSocket.instances.length - 1]
    socket.simulateOpen()

    wsClient.subscribe('step.updated', handler)

    // Backend sends a JSON array of canonical WSMessages.
    const batch = JSON.stringify([
      { id: 'evt-1', version: 1, type: 'step.updated', drill_id: 42, occurred_at: '2026-06-25T12:00:00Z', data: { step_id: 1 } },
      { id: 'evt-2', version: 1, type: 'step.updated', drill_id: 42, occurred_at: '2026-06-25T12:00:01Z', data: { step_id: 2 } },
    ])
    socket.simulateMessage(batch)

    expect(handler).toHaveBeenCalledTimes(2)
    expect(handler.mock.calls[0][0].type).toBe('step.updated')
    expect(handler.mock.calls[0][0].drill_id).toBe(42)
    expect(handler.mock.calls[1][0].data.step_id).toBe(2)
    wsClient.disconnect()
  })
})
