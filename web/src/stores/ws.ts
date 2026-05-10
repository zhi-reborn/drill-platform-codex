import { defineStore } from 'pinia'
import { ref } from 'vue'
import { wsClient } from '@/websocket'

export const useWsStore = defineStore('ws', () => {
  const status = ref<'connecting' | 'connected' | 'disconnected'>('disconnected')
  const statusText = ref('WebSocket 未连接')

  function update() {
    status.value = wsClient.status
    const texts: Record<string, string> = {
      connecting: 'WebSocket 连接中...',
      connected: 'WebSocket 已连接',
      disconnected: 'WebSocket 已断开',
    }
    statusText.value = texts[wsClient.status] || ''
  }

  return { status, statusText, update }
})
