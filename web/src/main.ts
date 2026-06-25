import { createApp } from 'vue'
import { createPinia } from 'pinia'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import App from './App.vue'
import router from './router'
import { setupGuards } from './router/guards'

import './styles/_fonts.scss'
import './styles/_reset.scss'
import './styles/_variables.scss'
import './styles/_mixins.scss'
import './styles/_element-overrides.scss'

const app = createApp(App)

app.use(createPinia())
app.use(router)
setupGuards(router)
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

if (import.meta.env.VITE_USE_MOCK === 'true') {
  const { setupMock } = await import('./mock/handlers')
  setupMock()
}

app.mount('#app')

// Initialize WebSocket after mount
import { useAuthStore } from './stores/auth'
import { useWebSocket } from './composables/useWebSocket'
import { useNotificationStore } from './stores/notifications'
import type { NormalizedWsMessage } from './websocket/messageTypes'

const authStore = useAuthStore()
const notifStore = useNotificationStore()
authStore.restoreSession()

if (authStore.isAuthenticated && authStore.token) {
  const ws = useWebSocket()
  ws.init(authStore.token)

  // Subscribe to WebSocket events and update notifications
  ws.subscribe('*', (msg: NormalizedWsMessage) => {
    // Map WebSocket events to notification types
    const eventToNotifType: Record<string, string> = {
      'drill_started': 'drill_started',
      'drill_paused': 'drill_paused',
      'drill_resumed': 'drill_resumed',
      'drill_completed': 'drill_completed',
      'drill_terminated': 'drill_terminated',
      'step_complete': 'step_complete',
      'step_timeout': 'step_timeout',
      'task_assigned': 'task_assigned',
    }

    const notifType = eventToNotifType[msg.type]
    if (notifType && msg.data) {
      const payload = msg.data as Record<string, any>
      notifStore.addNotification({
        id: Date.now(), // temporary ID
        user_id: authStore.user?.id || 0,
        type: notifType as any,
        title: payload.title || '系统通知',
        content: payload.content || '',
        drill_id: payload.drill_id,
        drill_name: payload.drill_name,
        step_id: payload.step_id,
        step_name: payload.step_name,
        is_read: false,
        created_at: new Date().toISOString(),
      })
    }
  })
}

// Make ws available globally for header component
;(window as unknown as Record<string, unknown>).__ws = useWebSocket
