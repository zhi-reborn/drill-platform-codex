<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">消息中心</h2>
      <div class="header-actions">
        <div class="ws-status" :class="wsStatus">
          <el-icon><Connection /></el-icon>
          <span>{{ wsStatusText }}</span>
        </div>
        <el-select v-model="filterType" placeholder="筛选类型" clearable style="width: 140px">
          <el-option label="全部" value="" />
          <el-option label="演练开始" value="drill_started" />
          <el-option label="演练完成" value="drill_completed" />
          <el-option label="演练暂停" value="drill_paused" />
          <el-option label="任务分配" value="task_assigned" />
          <el-option label="步骤完成" value="step_complete" />
          <el-option label="步骤超时" value="step_timeout" />
          <el-option label="系统公告" value="system_alert" />
        </el-select>
        <el-button @click="handleMarkAllAsRead" :disabled="unreadCount === 0">
          <el-icon><Check /></el-icon>
          全部标为已读
        </el-button>
      </div>
    </div>

    <div class="page-content">
      <NotificationList
        :notifications="filteredNotifications"
        :loading="loading"
        @read="handleMarkAsRead"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { Check, Connection } from '@element-plus/icons-vue'
import { useNotificationStore } from '@/stores/notifications'
import { useAuthStore } from '@/stores/auth'
import { useWsStore } from '@/stores/ws'
import NotificationList from '@/components/notifications/NotificationList.vue'
import type { Notification } from '@/types'

const notifStore = useNotificationStore()
const authStore = useAuthStore()
const wsStore = useWsStore()

const loading = ref(false)
const filterType = ref('')
const pageSize = 20

// 合并 store 数据 + 筛选
const allNotifications = computed(() => notifStore.notifications)

const filteredNotifications = computed(() => {
  if (!filterType.value) return allNotifications.value
  return allNotifications.value.filter(n => n.type === filterType.value)
})

const unreadCount = computed(() => notifStore.unreadCount)

const wsStatusText = computed(() => wsStore.statusText)
const wsStatus = computed(() => wsStore.status)

onMounted(async () => {
  loading.value = true
  try {
    await notifStore.fetchNotifications()
  } finally {
    loading.value = false
  }
  
  // 更新 WebSocket 状态
  wsStore.update()
})

onBeforeUnmount(() => {
  // 组件卸载时不清用连接，保持全局 WebSocket
})

function handleMarkAsRead(notification: Notification) {
  if (!notification.is_read) {
    notifStore.markAsRead(notification.id)
  }
}

function handleMarkAllAsRead() {
  notifStore.markAllAsRead()
  ElMessage.success('已全部标为已读')
}

function handleLoadMore(page: number) {
  // 页面模式下，store 已经加载了足够的数据
  // 如果需要分页获取更多，可以在这里调用 API
  console.log('Load page:', page)
}
</script>

<style scoped lang="scss">
@use '@/styles/layout' as *;
@use '@/styles/variables' as *;

.page-container {
  @include page-container;

  .page-header {
    @include page-header;

    .header-actions {
      display: flex;
      align-items: center;
      gap: $spacing-sm;

      .ws-status {
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 4px 10px;
        border-radius: 12px;
        font-size: 12px;
        background: $bg-tertiary;
        color: $text-secondary;

        &.connected {
          background: $color-success-bg;
          color: $color-success;
        }

        &.disconnected {
          background: $color-error-bg;
          color: $color-error;
        }

        &.connecting {
          background: $color-warning-bg;
          color: $color-warning;
        }
      }
    }
  }

  .page-content {
    @include page-content;
  }
}
</style>