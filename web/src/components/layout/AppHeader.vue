<template>
  <header class="app-header">
    <div class="header-left">
      <el-button text @click="$emit('toggle-sidebar')" class="collapse-btn">
        <el-icon :size="20"><Fold v-if="!collapsed" /><Expand v-else /></el-icon>
      </el-button>
      <div class="logo">
        <el-icon :size="24" color="#22C55E"><Monitor /></el-icon>
        <span class="logo-text">应急处置流程管理系统</span>
      </div>
    </div>

    <div class="header-right">
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
import NotificationPopover from '@/components/notifications/NotificationPopover.vue'

interface Props {
  collapsed?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  collapsed: false
})

defineEmits<{ 'toggle-sidebar': [] }>()

const authStore = useAuthStore()

const userInitial = computed(() => authStore.userInitial)
const userName = computed(() => authStore.userName)
const userDept = computed(() => authStore.userDept)
const roleName = computed(() => authStore.roleName)
const roleType = computed(() => authStore.roleType)

onMounted(() => {
  authStore.restoreSession()
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
  background: $sidebar-bg;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 $spacing-xl;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: $z-index-header;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.15);
}

.header-left {
  display: flex;
  align-items: center;
  gap: $spacing-base;
}

.logo {
  display: flex;
  align-items: center;
  gap: $spacing-sm;

  .logo-text {
    font-size: $font-size-lg;
    font-weight: $font-weight-semibold;
    color: $sidebar-text-active;
    letter-spacing: 0.5px;
  }
}

.collapse-btn {
  color: $sidebar-text;
  
  &:hover {
    color: $sidebar-text-active;
    background: $sidebar-bg-hover;
  }
}

.header-right {
  display: flex;
  align-items: center;
  gap: $spacing-base;
}

.user-info {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  cursor: pointer;
  padding: $spacing-xs $spacing-sm;
  border-radius: $radius-base;
  color: $sidebar-text-active;

  &:hover {
    background: $sidebar-bg-hover;
  }

  .user-name {
    font-size: $font-size-sm;
    color: $sidebar-text-active;
  }

  .user-dept {
    font-size: 11px;
    color: $sidebar-text;
    background: rgba(255, 255, 255, 0.08);
    padding: 1px 6px;
    border-radius: 4px;
  }
}

.user-avatar {
  background: $color-accent;
  color: white;
  font-weight: $font-weight-semibold;
  font-size: $font-size-sm;
}

</style>

<style lang="scss">
/* Header 深色模式 - 强制覆盖所有 button hover 背景 */
@use '@/styles/variables' as *;

.app-header .el-button:hover,
.app-header .el-button.el-button--text:hover {
  background-color: #{$sidebar-bg-hover} !important;
}

.app-header .notification-bell .el-button {
  background-color: transparent !important;
}
</style>