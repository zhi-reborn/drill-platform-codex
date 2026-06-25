import { ref } from 'vue'
import type { NormalizedWsMessage } from '@/websocket/messageTypes'
import { wsClient } from '@/websocket'

export function useWebSocket() {
  const connectionStatus = ref<'connecting' | 'connected' | 'disconnected'>('disconnected')

  function subscribe(channel: string | '*', handler: (msg: NormalizedWsMessage) => void) {
    wsClient.subscribe(channel, handler)
  }

  function unsubscribe(channel: string, handler: (msg: NormalizedWsMessage) => void) {
    wsClient.unsubscribe(channel, handler)
  }

  function init(token: string) {
    wsClient.connect(token)
    wsClient.subscribe('connection_status', (msg) => {
      connectionStatus.value = (msg.data as { status: string }).status as 'connecting' | 'connected' | 'disconnected'
    })
  }

  return { connectionStatus, subscribe, unsubscribe, init }
}
