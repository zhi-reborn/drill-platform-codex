<template>
  <header class="app-header">
    <div class="header-left">
      <el-button text @click="$emit('toggle-sidebar')" class="collapse-btn">
        <el-icon :size="20"><Fold v-if="!collapsed" /><Expand v-else /></el-icon>
      </el-button>
      <div class="logo">
        <el-icon :size="24" color="#0891B2"><Monitor /></el-icon>
        <span class="logo-text">Drill Platform</span>
      </div>
    </div>

    <div class="header-right">
      <!-- WebSocket status -->
      <div class="ws-status" :class="wsStatus" :title="wsStatusText">
        <span class="status-dot"></span>
      </div>

      <!-- Notifications bell -->
      <el-popover v-model:visible="showNotifications" placement="bottom-end" :width="360" trigger="click">
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
            <span class="popover-title">通知</span>
            <el-button text size="small" @click="notifStore.markAllAsRead">全部已读</el-button>
          </div>
          <el-scrollbar max-height="400px">
            <div v-if="notifStore.recentNotifications.length === 0" class="empty-state">
              <el-empty description="暂无通知" :image-size="80" />
            </div>
            <div v-else class="notification-list">
              <div
                v-for="n in notifStore.recentNotifications"
                :key="n.id"
                class="notification-item"
                :class="{ 'is-unread': !n.is_read }"
                @click="notifStore.markAsRead(n.id)"
              >
                <div class="notification-content">
                  <div class="notification-title">{{ n.title }}</div>
                  <div class="notification-meta">
                    <span class="notification-type">{{ n.type }}</span>
                    <span class="notification-time">{{ new Date(n.created_at).toLocaleString('zh-CN') }}</span>
                  </div>
                </div>
              </div>
            </div>
          </el-scrollbar>
        </div>
      </el-popover>

      <!-- User dropdown -->
      <el-dropdown trigger="click" @command="handleUserCommand">
        <div class="user-info">
          <el-avatar :size="28" class="user-avatar">
            {{ userInitial }}
          </el-avatar>
          <span class="user-name">{{ userName }}</span>
          <el-icon><ArrowDown /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item disabled>
              <el-tag size="small" :type="roleType as 'primary' | 'success' | 'warning' | 'info' | 'danger'">{{ roleName }}</el-tag>
            </el-dropdown-item>
            <el-dropdown-item divided command="logout">
              <el-icon><SwitchButton /></el-icon>
              退出登录
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Fold, Expand, Monitor, Bell, ArrowDown, SwitchButton } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notifications'
import { useWsStore } from '@/stores/ws'
import { useRouter } from 'vue-router'

const router = useRouter()

interface Props {
  collapsed?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  collapsed: false
})

defineEmits<{ 'toggle-sidebar': [] }>()

const authStore = useAuthStore()
const notifStore = useNotificationStore()
const wsStore = useWsStore()

const showNotifications = ref(false)

const userInitial = computed(() => authStore.userInitial)
const userName = computed(() => authStore.userName)
const roleName = computed(() => authStore.roleName)
const roleType = computed(() => authStore.roleType)
const unreadCount = computed(() => notifStore.unreadCount)
const wsStatus = computed(() => wsStore.status)
const wsStatusText = computed(() => wsStore.statusText)

onMounted(() => {
  authStore.restoreSession()
  notifStore.fetchNotifications()
  wsStore.update()
})

function handleUserCommand(command: string) {
  if (command === 'logout') {
    authStore.logout()
  }
}
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

// 深色顶部栏
.app-header {
  height: $header-height;
  background: #1E293B;
  border-bottom: 1px solid #334155;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 $spacing-base;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: $z-index-header;
}

.header-left {
  display: flex;
  align-items: center;
  gap: $spacing-md;
}

.logo {
  display: flex;
  align-items: center;
  gap: $spacing-sm;

  .logo-text {
    font-size: $font-size-lg;
    font-weight: $font-weight-bold;
    color: #F1F5F9;
    letter-spacing: 0.5px;
  }
}

.header-right {
  display: flex;
  align-items: center;
  gap: $spacing-md;
}

.ws-status {
  width: 10px;
  height: 10px;
  border-radius: 50%;

  &.connected .status-dot {
    background: $color-success;
    box-shadow: 0 0 4px $color-success;
  }

  &.disconnected .status-dot {
    background: $color-error;
  }
}

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

      &:hover {
        background: $bg-tertiary;
      }

      &.is-unread {
        background: $color-primary-bg;
      }

      .notification-content {
        .notification-title {
          font-size: $font-size-sm;
          color: $text-primary;
          margin-bottom: $spacing-xs;
        }

        .notification-meta {
          display: flex;
          justify-content: space-between;
          font-size: $font-size-xs;
          color: $text-secondary;

          .notification-type {
            text-transform: capitalize;
          }
        }
      }
    }
  }
}

.user-info {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  cursor: pointer;
  padding: $spacing-xs $spacing-sm;
  border-radius: $radius-base;
  color: #F1F5F9;

  &:hover {
    background: #334155;
  }

  .user-name {
    font-size: $font-size-sm;
    color: #F1F5F9;
  }
}

.user-avatar {
  background: $color-primary;
  color: white;
  font-weight: $font-weight-semibold;
  font-size: $font-size-sm;
}

:deep(.el-button) {
  color: #F1F5F9;
}
</style>