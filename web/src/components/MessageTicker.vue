<template>
  <div class="message-ticker" role="log" aria-label="消息滚动" aria-live="polite">
    <div class="ticker-header" v-if="title">
      <span class="ticker-title">{{ title }}</span>
      <el-button
        v-if="showClear"
        text
        size="small"
        @click="clearMessages"
        aria-label="清空消息"
      >
        <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2" />
        </svg>
      </el-button>
    </div>
    <div class="ticker-container" ref="containerRef">
      <transition-group name="message-list" tag="div" class="message-list">
        <div
          v-for="message in displayedMessages"
          :key="message.id"
          class="message-item"
          :class="getMessageClass(message.type)"
          role="article"
        >
          <span class="message-icon" aria-hidden="true">
            <svg v-if="message.type === 'info'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10" />
              <path d="M12 16v-4M12 8h.01" />
            </svg>
            <svg v-else-if="message.type === 'success'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 11.08V12a10 10 0 11-5.93-9.14" />
              <polyline points="22 4 12 14.01 9 11.01" />
            </svg>
            <svg v-else-if="message.type === 'warning'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
              <line x1="12" y1="9" x2="12" y2="13" />
              <line x1="12" y1="17" x2="12.01" y2="17" />
            </svg>
            <svg v-else-if="message.type === 'error'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10" />
              <line x1="15" y1="9" x2="9" y2="15" />
              <line x1="9" y1="9" x2="15" y2="15" />
            </svg>
          </span>
          <span class="message-time">{{ formatTime(message.timestamp) }}</span>
          <span class="message-content">{{ message.content }}</span>
        </div>
      </transition-group>
      <div v-if="displayedMessages.length === 0" class="empty-state">
        <span>暂无消息</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElButton } from 'element-plus'

interface MessageItem {
  id: string | number
  type: 'info' | 'success' | 'warning' | 'error'
  content: string
  timestamp: number
}

interface MessageTickerProps {
  messages: MessageItem[]
  title?: string
  maxMessages?: number
  showClear?: boolean
  autoScroll?: boolean
}

const props = withDefaults(defineProps<MessageTickerProps>(), {
  title: '消息通知',
  maxMessages: 100,
  showClear: true,
  autoScroll: true
})

const emit = defineEmits<{
  clear: []
}>()

const containerRef = ref<HTMLElement | null>(null)

// 限制消息数量
const displayedMessages = computed(() => {
  return props.messages.slice(-props.maxMessages)
})

// 消息样式类
const getMessageClass = (type: MessageItem['type']) => {
  const classMap = {
    info: 'message-info',
    success: 'message-success',
    warning: 'message-warning',
    error: 'message-error'
  }
  return classMap[type]
}

// 格式化时间
const formatTime = (timestamp: number) => {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  // 1 分钟内显示"刚刚"
  if (diff < 60000) {
    return '刚刚'
  }
  
  // 1 小时内显示分钟
  if (diff < 3600000) {
    return `${Math.floor(diff / 60000)}分钟前`
  }
  
  // 今天显示时分
  if (diff < 86400000 && date.getDate() === now.getDate()) {
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }
  
  // 其他显示日期
  return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' })
}

// 清空消息
const clearMessages = () => {
  emit('clear')
}

// 自动滚动到底部
watch(
  () => displayedMessages.value.length,
  () => {
    if (props.autoScroll && containerRef.value) {
      // 使用 nextTick 确保 DOM 更新后滚动
      requestAnimationFrame(() => {
        containerRef.value?.scrollTo({
          top: containerRef.value.scrollHeight,
          behavior: 'smooth'
        })
      })
    }
  }
)
</script>

<style scoped>
.message-ticker {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: var(--color-muted, #1A1E2F);
  border-radius: 8px;
  border: 1px solid var(--color-border, #334155);
  overflow: hidden;
}

.ticker-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--color-border, #334155);
  background-color: var(--color-secondary, #1E293B);
}

.ticker-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-foreground, #F8FAFC);
  font-family: 'Fira Sans', sans-serif;
}

.ticker-container {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.message-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.message-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 6px;
  font-size: 13px;
  line-height: 1.5;
  transition: all 0.2s ease;
}

.message-info {
  background-color: rgba(59, 130, 246, 0.1);
  border-left: 3px solid #3B82F6;
  color: var(--color-foreground, #F8FAFC);
}

.message-success {
  background-color: rgba(34, 197, 94, 0.1);
  border-left: 3px solid #22C55E;
  color: var(--color-foreground, #F8FAFC);
}

.message-warning {
  background-color: rgba(245, 158, 11, 0.1);
  border-left: 3px solid #F59E0B;
  color: var(--color-foreground, #F8FAFC);
}

.message-error {
  background-color: rgba(239, 68, 68, 0.1);
  border-left: 3px solid #EF4444;
  color: var(--color-foreground, #F8FAFC);
}

.message-icon {
  flex-shrink: 0;
  width: 16px;
  height: 16px;
  margin-top: 2px;
}

.message-icon svg {
  width: 100%;
  height: 100%;
}

.message-info .message-icon { color: #3B82F6; }
.message-success .message-icon { color: #22C55E; }
.message-warning .message-icon { color: #F59E0B; }
.message-error .message-icon { color: #EF4444; }

.message-time {
  flex-shrink: 0;
  font-size: 11px;
  color: var(--color-muted-foreground, #94A3B8);
  font-family: 'Fira Sans', sans-serif;
}

.message-content {
  flex: 1;
  word-break: break-word;
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px 20px;
  color: var(--color-muted-foreground, #94A3B8);
  font-size: 14px;
}

/* 列表动画 */
.message-list-enter-active,
.message-list-leave-active {
  transition: all 0.3s ease;
}

.message-list-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.message-list-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

/* 滚动条样式 */
.ticker-container::-webkit-scrollbar {
  width: 6px;
}

.ticker-container::-webkit-scrollbar-track {
  background: var(--color-muted, #1A1E2F);
}

.ticker-container::-webkit-scrollbar-thumb {
  background: var(--color-border, #334155);
  border-radius: 3px;
}

.ticker-container::-webkit-scrollbar-thumb:hover {
  background: #475569;
}
</style>
