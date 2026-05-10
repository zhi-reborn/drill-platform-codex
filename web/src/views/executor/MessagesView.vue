<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">消息中心</h2>
      <el-button @click="markAllAsRead">
        <el-icon><Check /></el-icon>
        全部标为已读
      </el-button>
    </div>

    <div class="page-content">
      <div v-loading="loading" class="notifications-list">
        <EmptyBox v-if="!loading && filteredNotifications.length === 0" title="暂无消息" description="当前没有通知消息" />

        <el-card
          v-for="notification in filteredNotifications"
          :key="notification.id"
          class="notification-card"
          :class="{ unread: !notification.is_read }"
          @click="markAsRead(notification)"
        >
          <div class="notification-header">
            <el-tag :type="getNotificationTypeTag(notification.type)" size="small">
              {{ getNotificationTypeLabel(notification.type) }}
            </el-tag>
            <span class="notification-time">{{ formatTime(notification.created_at) }}</span>
          </div>
          <div class="notification-title">{{ notification.title }}</div>
          <div class="notification-content">{{ notification.content }}</div>
          <div v-if="notification.drill_name" class="notification-drill">
            <el-icon><Document /></el-icon>
            <span>{{ notification.drill_name }}</span>
          </div>
        </el-card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Check, Document } from '@element-plus/icons-vue'
import type { Notification } from '@/types'
import EmptyBox from '@/components/common/EmptyBox.vue'
import notificationsData from '@/mock/data/notifications.json'

const loading = ref(false)
const notifications = ref<Notification[]>([])

// 模拟当前用户 ID（实际应从 store 获取）
const currentUserId = 3

const filteredNotifications = computed(() => {
  return notifications.value.filter(n => n.user_id === currentUserId)
})

function getNotificationTypeTag(type: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, any> = {
    drill_started: 'primary',
    drill_completed: 'success',
    drill_paused: 'warning',
    task_assigned: 'info',
    step_complete: 'success',
    step_timeout: 'warning',
    system_alert: 'danger',
  }
  return map[type] || 'info'
}

function getNotificationTypeLabel(type: string): string {
  const map: Record<string, string> = {
    drill_started: '演练开始',
    drill_completed: '演练完成',
    drill_paused: '演练暂停',
    task_assigned: '任务分配',
    step_complete: '步骤完成',
    step_timeout: '步骤超时',
    system_alert: '系统通知',
  }
  return map[type] || type
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const hours = Math.floor(diff / (1000 * 60 * 60))

  if (hours < 1) {
    return '刚刚'
  }
  if (hours < 24) {
    return `${hours}小时前`
  }
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

async function loadNotifications() {
  loading.value = true
  try {
    notifications.value = notificationsData as Notification[]
  } catch (error) {
    ElMessage.error('加载通知失败')
    console.error('Failed to load notifications:', error)
  } finally {
    loading.value = false
  }
}

function markAsRead(notification: Notification) {
  notification.is_read = true
}

function markAllAsRead() {
  filteredNotifications.value.forEach(n => {
    n.is_read = true
  })
  ElMessage.success('已全部标为已读')
}

onMounted(() => {
  loadNotifications()
})
</script>

<style scoped lang="scss">
@use '@/styles/layout' as *;
@use '@/styles/variables' as *;

.page-container {
  @include page-container;

  .page-header {
    @include page-header;

    .page-title {
      font-size: $font-size-xl;
      font-weight: $font-weight-bold;
      color: $text-primary;
      margin: 0;
    }
  }

  .page-content {
    @include page-content;

    .notifications-list {
      display: flex;
      flex-direction: column;
      gap: $spacing-sm;
    }

    .notification-card {
      background: $bg-secondary;
      border-color: $border-color;
      cursor: pointer;
      transition: all 0.2s;
      border-left-width: 4px;
      border-left-color: transparent;

      &:hover {
        border-color: $color-primary;
        box-shadow: $shadow-sm;
      }

      &.unread {
        background: $color-primary-bg;
        border-left-color: $color-primary;
      }

      .notification-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: $spacing-xs;

        .notification-time {
          font-size: $font-size-xs;
          color: $text-tertiary;
        }
      }

      .notification-title {
        font-size: $font-size-base;
        font-weight: $font-weight-semibold;
        color: $text-primary;
        margin-bottom: $spacing-xs;
      }

      .notification-content {
        font-size: $font-size-sm;
        color: $text-secondary;
        line-height: 1.6;
        margin-bottom: $spacing-sm;
      }

      .notification-drill {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: $font-size-xs;
        color: $text-tertiary;

        .el-icon {
          font-size: 14px;
        }
      }
    }
  }
}
</style>
