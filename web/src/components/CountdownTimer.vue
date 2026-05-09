<template>
  <div class="countdown-timer" role="timer" :aria-label="`剩余时间 ${formattedTime}`">
    <div class="timer-display" :class="timerClass">
      <span class="time-segment">
        <span class="time-value">{{ hours }}</span>
        <span class="time-label">时</span>
      </span>
      <span class="time-separator">:</span>
      <span class="time-segment">
        <span class="time-value">{{ minutes }}</span>
        <span class="time-label">分</span>
      </span>
      <span class="time-separator">:</span>
      <span class="time-segment">
        <span class="time-value">{{ seconds }}</span>
        <span class="time-label">秒</span>
      </span>
    </div>
    <div v-if="showProgress && totalDuration > 0" class="timer-progress">
      <div class="progress-bar" :style="{ width: progressPercent + '%' }"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'

interface CountdownTimerProps {
  // 结束时间戳 (毫秒)
  endTime?: number
  // 或者剩余秒数
  remainingSeconds?: number
  // 是否显示进度条
  showProgress?: boolean
  // 总时长 (用于进度条计算，秒)
  totalDuration?: number
  // 超时回调
  onTimeout?: () => void
  // 是否自动开始
  autoStart?: boolean
}

const props = withDefaults(defineProps<CountdownTimerProps>(), {
  endTime: 0,
  remainingSeconds: 0,
  showProgress: false,
  totalDuration: 0,
  autoStart: true
})

const emit = defineEmits<{
  timeout: []
  tick: [remaining: number]
}>()

const remaining = ref(0)
const timerInterval = ref<number | null>(null)
const isRunning = ref(false)

// 格式化时间
const hours = computed(() => {
  const hrs = Math.floor(remaining.value / 3600)
  return hrs.toString().padStart(2, '0')
})

const minutes = computed(() => {
  const mins = Math.floor((remaining.value % 3600) / 60)
  return mins.toString().padStart(2, '0')
})

const seconds = computed(() => {
  const secs = Math.floor(remaining.value % 60)
  return secs.toString().padStart(2, '0')
})

const formattedTime = computed(() => {
  return `${hours.value}时${minutes.value}分${seconds.value}秒`
})

// 进度百分比
const progressPercent = computed(() => {
  if (props.totalDuration <= 0 || remaining.value <= 0) return 0
  return Math.min(100, (remaining.value / props.totalDuration) * 100)
})

// 定时器样式
const timerClass = computed(() => {
  if (remaining.value <= 0) return 'timer-expired'
  if (remaining.value <= 60) return 'timer-warning'
  if (remaining.value <= 300) return 'timer-caution'
  return 'timer-normal'
})

// 开始倒计时
const start = () => {
  if (isRunning.value) return

  isRunning.value = true

  // 初始化剩余时间
  if (props.endTime > 0) {
    remaining.value = Math.max(0, Math.floor((props.endTime - Date.now()) / 1000))
  } else if (props.remainingSeconds > 0) {
    remaining.value = props.remainingSeconds
  }

  if (remaining.value <= 0) {
    emit('timeout')
    return
  }

  timerInterval.value = window.setInterval(() => {
    remaining.value--

    emit('tick', remaining.value)

    if (remaining.value <= 0) {
      stop()
      emit('timeout')
      props.onTimeout?.()
    }
  }, 1000)
}

// 停止倒计时
const stop = () => {
  if (timerInterval.value !== null) {
    clearInterval(timerInterval.value)
    timerInterval.value = null
  }
  isRunning.value = false
}

// 重置倒计时
const reset = (newSeconds?: number) => {
  stop()
  remaining.value = newSeconds ?? props.remainingSeconds ?? 0
  if (props.autoStart && remaining.value > 0) {
    start()
  }
}

// 监听 props 变化
watch(
  () => [props.endTime, props.remainingSeconds],
  () => {
    if (isRunning.value) {
      if (props.endTime > 0) {
        remaining.value = Math.max(0, Math.floor((props.endTime - Date.now()) / 1000))
      } else if (props.remainingSeconds > 0) {
        remaining.value = props.remainingSeconds
      }
    }
  }
)

// 生命周期
onMounted(() => {
  if (props.autoStart) {
    start()
  }
})

onBeforeUnmount(() => {
  stop()
})

// 暴露方法
defineExpose({
  start,
  stop,
  reset,
  getRemaining: () => remaining.value,
  isRunning: () => isRunning.value
})
</script>

<style scoped>
.countdown-timer {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.timer-display {
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'Fira Code', monospace;
  font-size: 32px;
  font-weight: 600;
  padding: 12px 24px;
  background-color: var(--color-muted, #1A1E2F);
  border-radius: 8px;
  border: 2px solid var(--color-border, #334155);
  transition: all 0.3s ease;
}

.timer-normal {
  color: var(--color-foreground, #F8FAFC);
  border-color: var(--color-primary, #0F172A);
}

.timer-caution {
  color: #F59E0B;
  border-color: #F59E0B;
  animation: pulse-caution 2s infinite;
}

.timer-warning {
  color: #EF4444;
  border-color: #EF4444;
  animation: pulse-warning 1s infinite;
}

.timer-expired {
  color: #EF4444;
  border-color: #EF4444;
  opacity: 0.8;
}

.time-segment {
  display: flex;
  flex-direction: column;
  align-items: center;
  line-height: 1;
}

.time-value {
  font-size: 32px;
  font-weight: 700;
  letter-spacing: 2px;
}

.time-label {
  font-size: 10px;
  font-weight: 400;
  color: var(--color-muted-foreground, #94A3B8);
  margin-top: 4px;
  font-family: 'Fira Sans', sans-serif;
}

.time-separator {
  color: var(--color-foreground, #F8FAFC);
  margin: 0 4px;
  font-weight: 700;
}

.timer-progress {
  width: 100%;
  height: 4px;
  background-color: var(--color-border, #334155);
  border-radius: 2px;
  overflow: hidden;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(
    90deg,
    var(--color-accent, #22C55E) 0%,
    #F59E0B 50%,
    var(--color-destructive, #EF4444) 100%
  );
  transition: width 0.3s ease;
}

@keyframes pulse-caution {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(245, 158, 11, 0.4);
  }
  50% {
    box-shadow: 0 0 0 8px rgba(245, 158, 11, 0);
  }
}

@keyframes pulse-warning {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.4);
  }
  50% {
    box-shadow: 0 0 0 8px rgba(239, 68, 68, 0);
  }
}

/* 无障碍支持 - 减少动画 */
@media (prefers-reduced-motion: reduce) {
  .timer-caution,
  .timer-warning {
    animation: none;
  }
}
</style>
