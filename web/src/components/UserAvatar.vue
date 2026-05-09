<template>
  <div class="user-avatar" :class="avatarClass" :style="avatarStyle">
    <!-- 头像图片 -->
    <img
      v-if="src && !hasError"
      :src="src"
      :alt="alt"
      class="avatar-image"
      @error="handleError"
    />
    <!-- 默认头像 (首字母) -->
    <span v-else class="avatar-fallback">
      {{ fallbackText }}
    </span>
    <!-- 状态指示器 -->
    <span
      v-if="showStatus && status !== 'offline'"
      class="status-indicator"
      :class="`status-${status}`"
      :aria-label="`状态：${statusText}`"
    ></span>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface UserAvatarProps {
  // 头像图片 URL
  src?: string
  // 用户姓名 (用于生成首字母)
  name?: string
  // 替代文本
  alt?: string
  // 头像大小
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  // 形状
  shape?: 'circle' | 'square' | 'rounded'
  // 显示状态指示器
  showStatus?: boolean
  // 在线状态
  status?: 'online' | 'busy' | 'away' | 'offline'
  // 背景颜色 (可选，会自动生成)
  backgroundColor?: string
}

const props = withDefaults(defineProps<UserAvatarProps>(), {
  src: '',
  name: '',
  alt: '用户头像',
  size: 'md',
  shape: 'circle',
  showStatus: false,
  status: 'offline',
  backgroundColor: ''
})

const hasError = ref(false)

// 尺寸映射
const sizeMap: Record<string, string> = {
  xs: '24px',
  sm: '32px',
  md: '40px',
  lg: '48px',
  xl: '64px'
}

// 状态文本映射
const statusTextMap: Record<string, string> = {
  online: '在线',
  busy: '忙碌',
  away: '离开',
  offline: '离线'
}

// 状态颜色映射
const statusColorMap: Record<string, string> = {
  online: '#22C55E',
  busy: '#EF4444',
  away: '#F59E0B',
  offline: '#64748B'
}

// 生成首字母
const fallbackText = computed(() => {
  if (props.name) {
    const parts = props.name.trim().split(/\s+/)
    if (parts.length >= 2) {
      return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
    }
    return props.name.slice(0, 2).toUpperCase()
  }
  return 'U'
})

// 生成背景色 (基于姓名哈希)
const generatedBackgroundColor = computed(() => {
  if (props.backgroundColor) return props.backgroundColor
  
  if (!props.name) return '#3B82F6'
  
  let hash = 0
  for (let i = 0; i < props.name.length; i++) {
    hash = props.name.charCodeAt(i) + ((hash << 5) - hash)
  }
  
  const hue = hash % 360
  return `hsl(${hue}, 60%, 25%)`
})

// 头像样式
const avatarStyle = computed(() => ({
  width: sizeMap[props.size],
  height: sizeMap[props.size],
  backgroundColor: generatedBackgroundColor.value
}))

// 头像类名
const avatarClass = computed(() => ({
  [`shape-${props.shape}`]: true,
  [`size-${props.size}`]: true
}))

// 状态文本
const statusText = computed(() => statusTextMap[props.status] || '离线')

// 处理图片加载错误
const handleError = () => {
  hasError.value = true
}
</script>

<style scoped>
.user-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  flex-shrink: 0;
  font-family: 'Fira Sans', sans-serif;
  font-weight: 600;
  font-size: 14px;
  color: #FFFFFF;
  transition: all 0.2s ease;
}

/* 形状 */
.shape-circle {
  border-radius: 50%;
}

.shape-square {
  border-radius: 4px;
}

.shape-rounded {
  border-radius: 12px;
}

/* 尺寸对应的字体大小 */
.size-xs { font-size: 10px; }
.size-sm { font-size: 12px; }
.size-md { font-size: 14px; }
.size-lg { font-size: 16px; }
.size-xl { font-size: 20px; }

.avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  text-transform: uppercase;
  letter-spacing: 1px;
}

/* 状态指示器 */
.status-indicator {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 25%;
  height: 25%;
  min-width: 8px;
  min-height: 8px;
  border-radius: 50%;
  border: 2px solid var(--color-background, #020617);
  transform: translate(10%, 10%);
}

.status-online {
  background-color: #22C55E;
}

.status-busy {
  background-color: #EF4444;
}

.status-away {
  background-color: #F59E0B;
}

.status-offline {
  background-color: #64748B;
}

/* 触摸反馈 */
@media (hover: hover) {
  .user-avatar:hover {
    transform: scale(1.05);
  }
}

/* 无障碍支持 */
.user-avatar:focus {
  outline: 2px solid var(--color-primary, #0F172A);
  outline-offset: 2px;
}

/* 减少动画 */
@media (prefers-reduced-motion: reduce) {
  .user-avatar {
    transition: none;
  }
}
</style>
