<template>
  <div class="notification-list-container">
    <div v-loading="loading" class="notifications-list">
      <EmptyBox v-if="!loading && notifications.length === 0" title="暂无消息" description="当前没有通知消息" />

      <el-card
        v-for="notification in notifications"
        :key="notification.id"
        class="notification-card"
        :class="{ unread: !notification.is_read }"
        @click="handleClick(notification)"
      >
        <div class="notification-header">
          <el-tag :type="getTypeTag(notification.type)" size="small">
            {{ getTypeLabel(notification.type) }}
          </el-tag>
          <div class="notification-header-right">
            <el-button
              class="delete-btn"
              text
              size="small"
              @click.stop="handleDelete(notification)"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
            <span class="notification-time">{{ formatTime(notification.created_at) }}</span>
          </div>
        </div>
        <div class="notification-title">{{ notification.title }}</div>
        <div class="notification-content">{{ notification.content }}</div>
        <div v-if="notification.drill_name" class="notification-drill">
          <el-icon><Document /></el-icon>
          <span>{{ notification.drill_name }}</span>
        </div>
      </el-card>
    </div>

    <!-- 分页 -->
    <div v-if="showPagination && total > pageSize" class="pagination-wrapper">
      <el-pagination
        v-model:current-page="currentPage"
        :page-size="pageSize"
        :total="total"
        :pager-count="5"
        layout="prev, pager, next"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Document, Delete } from '@element-plus/icons-vue'
import { NOTIFICATION_TYPE_LABELS, type NotificationType } from '@/types/notification'
import type { Notification } from '@/types'
import EmptyBox from '@/components/common/EmptyBox.vue'

interface Props {
  notifications: Notification[]
  loading?: boolean
  showPagination?: boolean
  total?: number
  pageSize?: number
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  showPagination: false,
  total: 0,
  pageSize: 20,
})

const emit = defineEmits<{
  loadMore: [page: number]
  'read': [notification: Notification]
  'delete': [notification: Notification]
}>()

const currentPage = ref(1)

const notifications = computed(() => props.notifications)

function getTypeTag(type: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, string> = {
    drill_started: 'primary',
    drill_completed: 'success',
    drill_paused: 'warning',
    drill_resumed: 'primary',
    drill_terminated: 'danger',
    task_assigned: 'info',
    step_complete: 'success',
    step_timeout: 'warning',
    system_alert: 'danger',
  }
  return (map[type] || 'info') as 'primary' | 'success' | 'warning' | 'danger' | 'info'
}

function getTypeLabel(type: string): string {
  return NOTIFICATION_TYPE_LABELS[type as NotificationType] || type
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes} 分钟前`
  if (hours < 24) return `${hours} 小时前`
  if (days < 7) return `${days} 天前`
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}

function handleClick(notification: Notification) {
  emit('read', notification)
}

function handleDelete(notification: Notification) {
  emit('delete', notification)
}

function handlePageChange(page: number) {
  emit('loadMore', page)
}

// 重置页码
watch(() => props.total, () => {
  currentPage.value = 1
})
</script>

<style scoped lang="scss">
@use '@/styles/layout' as *;
@use '@/styles/variables' as *;

.notification-list-container {
  display: flex;
  flex-direction: column;
  gap: $spacing-md;
}

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

    .delete-btn {
      opacity: 1;
    }
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

    .notification-header-right {
      display: flex;
      align-items: center;
      gap: $spacing-xs;

      .delete-btn {
        opacity: 0;
        padding: 2px;
        transition: opacity 0.2s;
        color: $text-tertiary;

        &:hover {
          color: $color-error;
        }
      }
    }

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
    color: $text-primary;
    line-height: 1.6;
    margin-bottom: $spacing-sm;
    padding: $spacing-xs;
    background: $bg-tertiary;
    border-radius: $radius-base;
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

.pagination-wrapper {
  display: flex;
  justify-content: center;
  padding-top: $spacing-md;
}
</style>