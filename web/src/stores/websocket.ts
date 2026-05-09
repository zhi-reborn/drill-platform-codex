import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { WS_CONFIG } from '@/utils/constants'
import { useUserStore } from './user'

// WebSocket 消息类型
export interface WSMessage {
  event_type: string
  drill_id?: number
  payload: any
  timestamp: number
}

// WebSocket 事件类型
export type WSEventType = 
  | 'drill_start'
  | 'drill_complete'
  | 'drill_terminate'
  | 'step_start'
  | 'step_complete'
  | 'step_issue'
  | 'task_assigned'
  | 'system_message'

// 事件处理器类型
export type WSEventHandler = (message: WSMessage) => void

export const useWebSocketStore = defineStore('websocket', () => {
  // 状态
  const ws = ref<WebSocket | null>(null)
  const isConnected = ref(false)
  const isConnecting = ref(false)
  const reconnectAttempts = ref(0)
  const lastMessage = ref<WSMessage | null>(null)
  
  // 事件处理器映射
  const eventHandlers = ref<Map<string, Set<WSEventHandler>>>(new Map())
  
  // 心跳定时器
  let heartbeatTimer: ReturnType<typeof setInterval> | null = null
  
  // 计算属性
  const connectionStatus = computed(() => {
    if (isConnecting.value) return 'connecting'
    if (isConnected.value) return 'connected'
    return 'disconnected'
  })
  
  // 注册事件处理器
  const on = (eventType: string, handler: WSEventHandler): void => {
    if (!eventHandlers.value.has(eventType)) {
      eventHandlers.value.set(eventType, new Set())
    }
    eventHandlers.value.get(eventType)!.add(handler)
  }
  
  // 移除事件处理器
  const off = (eventType: string, handler: WSEventHandler): void => {
    const handlers = eventHandlers.value.get(eventType)
    if (handlers) {
      handlers.delete(handler)
    }
  }
  
  // 触发事件
  const emit = (message: WSMessage): void => {
    lastMessage.value = message
    
    // 触发特定事件处理器
    const handlers = eventHandlers.value.get(message.event_type)
    handlers?.forEach(handler => handler(message))
    
    // 触发通用消息处理器
    const allHandlers = eventHandlers.value.get('*')
    allHandlers?.forEach(handler => handler(message))
  }
  
  // 发送消息
  const send = (data: any): void => {
    if (ws.value && ws.value.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify(data))
    } else {
      console.warn('WebSocket 未连接，无法发送消息')
    }
  }
  
  // 连接 WebSocket
  const connect = (drillId?: number): void => {
    if (isConnected.value || isConnecting.value) {
      return
    }
    
    isConnecting.value = true
    
    const userStore = useUserStore()
    const token = userStore.token
    
    // 构建连接 URL
    let url = `${WS_CONFIG.BASE_URL}/ws/tasks`
    if (drillId) {
      url = `${WS_CONFIG.BASE_URL}/ws/display/${drillId}`
    }
    
    // 添加 token 参数
    const separator = url.includes('?') ? '&' : '?'
    url = `${url}${separator}token=${encodeURIComponent(token)}`
    
    try {
      ws.value = new WebSocket(url)
      
      ws.value.onopen = () => {
        isConnected.value = true
        isConnecting.value = false
        reconnectAttempts.value = 0
        console.log('WebSocket 连接成功')
        ElMessage.success('实时通信已连接')
        
        // 启动心跳
        startHeartbeat()
      }
      
      ws.value.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data)
          emit(message)
        } catch (error) {
          console.error('解析 WebSocket 消息失败:', error)
        }
      }
      
      ws.value.onerror = (error) => {
        console.error('WebSocket 错误:', error)
        ElMessage.error('实时通信连接失败')
      }
      
      ws.value.onclose = () => {
        isConnected.value = false
        isConnecting.value = false
        console.log('WebSocket 连接关闭')
        
        // 停止心跳
        stopHeartbeat()
        
        // 尝试重连
        attemptReconnect(drillId)
      }
    } catch (error) {
      console.error('创建 WebSocket 连接失败:', error)
      isConnecting.value = false
    }
  }
  
  // 断开连接
  const disconnect = (): void => {
    stopHeartbeat()
    
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
    
    isConnected.value = false
    isConnecting.value = false
    reconnectAttempts.value = 0
    eventHandlers.value.clear()
    
    console.log('WebSocket 已断开')
  }
  
  // 尝试重连
  const attemptReconnect = (drillId?: number): void => {
    if (reconnectAttempts.value >= WS_CONFIG.MAX_RECONNECT_ATTEMPTS) {
      console.log('达到最大重连次数，停止重连')
      ElMessage.warning('实时通信断开，请检查网络连接')
      return
    }
    
    reconnectAttempts.value++
    const delay = WS_CONFIG.RECONNECT_INTERVAL * reconnectAttempts.value
    
    console.log(`准备重连 (${reconnectAttempts.value}/${WS_CONFIG.MAX_RECONNECT_ATTEMPTS}), 延迟 ${delay}ms`)
    
    setTimeout(() => {
      connect(drillId)
    }, delay)
  }
  
  // 启动心跳
  const startHeartbeat = (): void => {
    stopHeartbeat()
    
    heartbeatTimer = setInterval(() => {
      if (ws.value && ws.value.readyState === WebSocket.OPEN) {
        send({ event_type: 'heartbeat', timestamp: Date.now() })
      }
    }, WS_CONFIG.HEARTBEAT_INTERVAL)
  }
  
  // 停止心跳
  const stopHeartbeat = (): void => {
    if (heartbeatTimer) {
      clearInterval(heartbeatTimer)
      heartbeatTimer = null
    }
  }
  
  // 订阅演练状态
  const subscribeDrill = (drillId: number): void => {
    connect(drillId)
  }
  
  // 取消订阅
  const unsubscribeDrill = (): void => {
    disconnect()
  }
  
  return {
    // 状态
    ws,
    isConnected,
    isConnecting,
    reconnectAttempts,
    lastMessage,
    // 计算属性
    connectionStatus,
    // 方法
    on,
    off,
    send,
    connect,
    disconnect,
    subscribeDrill,
    unsubscribeDrill,
  }
})
