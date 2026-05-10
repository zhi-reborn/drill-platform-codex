<template>
  <el-popover placement="bottom-end" :width="380" trigger="click">
    <template #reference>
      <div class="notification-bell">
        <el-badge :value="unreadCount" :hidden="unreadCount === 0" :max="99">
          <el-button text>
            <el-icon :size="18"><Bell /></el-icon>
          </el-button>
        </el-badge>
      </div>
    </template>

    <div class="notification-popover">
      <div class="popover-header">
        <span class="popover-title">消息中心</span>
        <el-button v-if="notifications.length > 0" text size="small" @click="handleMarkAllAsRead">
          全部已读
        </el-button>
      </div>

      <el-scrollbar max-height="400px">
        <div v-if="notifications.length === 0" class="empty-state">
          <el-empty description="暂无消息" :image-size="60" />
        </div>
        <template v-else>
          <div class="notification-list">
            <div
              v-for="n in notifications"
              :key="n.id"
              class="notification-item"
              :class="{ 'is-unread': !n.is_read }"
              @click="handleItemClick(n)"
            >
              <div class="notification-content">
                <div class="notification-header">
                  <span class="notification-type">{{ getTypeLabel(n.type) }}</span>
                  <el-button
                    class="delete-btn"
                    text
                    size="small"
                    @click.stop="handleDelete(n.id)"
                  >
                    <el-icon><Close /></el-icon>
                  </el-button>
                </div>
                <div class="notification-title">{{ n.title }}</div>
                <div class="notification-time">{{ formatTime(n.created_at) }}</div>
              </div>
            </div>
          </div>

          <div v-if="hasMore" class="popover-footer">
            <el-button text size="small" @click="handleViewAll">
              查看全部消息
            </el-button>
          </div>
        </template>
      </el-scrollbar>
    </div>
  </el-popover>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { Bell, Close } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import { useNotificationStore } from '@/stores/notifications'
import { NOTIFICATION_TYPE_LABELS, type NotificationType } from '@/types/notification'
import { useAuthStore } from '@/stores/auth'
import type { Notification } from '@/types'

const router = useRouter()
const notifStore = useNotificationStore()
const authStore = useAuthStore()

onMounted(async () => {
  authStore.restoreSession()
  console.log('[NotificationPopover] Auth restored:', {
    isAuthenticated: authStore.isAuthenticated,
    user: authStore.user,
    role: authStore.role,
  })
  await notifStore.fetchNotifications()
  console.log('[NotificationPopover] Notifications loaded:', {
    total: notifStore.notifications.length,
    unread: notifStore.unreadCount,
    items: notifStore.notifications.slice(0, 3),
  })
})

const notifications = computed(() => notifStore.notifications.slice(0, 10))
const unreadCount = computed(() => notifStore.unreadCount)
const hasMore = computed(() => notifStore.notifications.length > 10)

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
  return date.toLocaleDateString('zh-CN')
}

function handleItemClick(n: Notification) {
  if (!n.is_read) {
    notifStore.markAsRead(n.id)
  }
}

function handleMarkAllAsRead() {
  notifStore.markAllAsRead()
}

function handleDelete(id: number) {
  notifStore.deleteNotification(id)
}

function handleViewAll() {
  const role = authStore.role
  router.push(`/${role}/messages`)
}
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

.notification-bell {
  cursor: pointer;
  color: #F1F5F9;
}

.notification-popover {
  .popover-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-bottom: $spacing-sm;
    border-bottom: 1px solid $border-color;
    margin-bottom: $spacing-sm;

    .popover-title {
      font-size: $font-size-base;
      font-weight: $font-weight-semibold;
      color: $text-primary;
    }
  }

  .empty-state {
    padding: $spacing-lg 0;
  }

  .notification-list {
    .notification-item {
      padding: $spacing-sm;
      border-radius: $radius-base;
      cursor: pointer;
      transition: background 0.2s;

      &:hover {
        background: $bg-tertiary;

        .delete-btn {
          opacity: 1;
        }
      }

      &.is-unread {
        background: $color-primary-bg;
      }

      .notification-content {
        .notification-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: $spacing-xs;

          .notification-type {
            font-size: $font-size-xs;
            color: $text-secondary;
          }

          .delete-btn {
            opacity: 0;
            padding: 2px;
            transition: opacity 0.2s;
          }
        }

        .notification-title {
          font-size: $font-size-sm;
          color: $text-primary;
          margin-bottom: $spacing-xs;
          line-height: 1.4;
        }

        .notification-time {
          font-size: $font-size-xs;
          color: $text-tertiary;
        }
      }
    }
  }

  .popover-footer {
    padding: $spacing-sm 0;
    border-top: 1px solid $border-color;
    margin-top: $spacing-sm;
    text-align: center;
  }
}
</style>