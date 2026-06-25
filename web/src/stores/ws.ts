import { defineStore } from 'pinia'
import { ref } from 'vue'
import { wsClient } from '@/websocket'

export const useWsStore = defineStore('ws', () => {
  const status = ref<'connecting' | 'connected' | 'disconnected'>('disconnected')
  const statusText = ref('WebSocket 未连接')
  // reconnectCount increments each time the WebSocket reconnects after a drop.
  // Views can watch this to trigger a drill state refetch — the single
  // realtime event path may have missed events while disconnected.
  const reconnectCount = ref(0)

  function update() {
    status.value = wsClient.status
    const texts: Record<string, string> = {
      connecting: 'WebSocket 连接中...',
      connected: 'WebSocket 已连接',
      disconnected: 'WebSocket 已断开',
    }
    statusText.value = texts[wsClient.status] || ''
  }

  // Listen for reconnect events dispatched by the WebSocket client. Each
  // reconnect increments the counter so views watching it can refetch.
  wsClient.subscribe('reconnect', () => {
    reconnectCount.value++
  })

  return { status, statusText, reconnectCount, update }
})
