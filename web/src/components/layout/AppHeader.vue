<template>
  <header class="app-header">
    <div class="header-left">
      <el-button text @click="$emit('toggle-sidebar')" class="collapse-btn">
        <el-icon :size="20"><Fold v-if="!collapsed" /><Expand v-else /></el-icon>
      </el-button>
      <div class="logo">
        <el-icon :size="24" color="#0891B2"><Monitor /></el-icon>
        <span class="logo-text">演练流程管理系统</span>
      </div>
    </div>

    <div class="header-right">
      <div class="ws-status" :class="wsStatus" :title="wsStatusText">
        <span class="status-dot"></span>
      </div>

      <NotificationPopover />

      <el-dropdown trigger="click" @command="handleUserCommand">
        <div class="user-info">
          <el-avatar :size="28" class="user-avatar">
            {{ userInitial }}
          </el-avatar>
          <span class="user-name">{{ userName }}</span>
          <span v-if="userDept" class="user-dept">{{ userDept }}</span>
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
import { computed, onMounted } from 'vue'
import { Fold, Expand, Monitor, ArrowDown, SwitchButton } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useWsStore } from '@/stores/ws'
import NotificationPopover from '@/components/notifications/NotificationPopover.vue'

interface Props {
  collapsed?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  collapsed: false
})

defineEmits<{ 'toggle-sidebar': [] }>()

const authStore = useAuthStore()
const wsStore = useWsStore()

const userInitial = computed(() => authStore.userInitial)
const userName = computed(() => authStore.userName)
const userDept = computed(() => authStore.userDept)
const roleName = computed(() => authStore.roleName)
const roleType = computed(() => authStore.roleType)
const wsStatus = computed(() => wsStore.status)
const wsStatusText = computed(() => wsStore.statusText)

onMounted(() => {
  authStore.restoreSession()
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

  .user-dept {
    font-size: 11px;
    color: #94A3B8;
    background: #334155;
    padding: 1px 6px;
    border-radius: 4px;
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