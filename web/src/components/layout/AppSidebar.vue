<template>
  <aside class="app-sidebar" :class="{ 'is-collapsed': collapsed }">
    <el-menu :default-active="activeMenu" :collapse="collapsed" :collapse-transition="true" router class="sidebar-menu">
      <template v-for="menu in visibleMenus" :key="menu.path">
        <el-menu-item v-if="!menu.children" :index="menu.path">
          <el-icon><component :is="menu.icon" /></el-icon>
          <template #title>{{ menu.title }}</template>
        </el-menu-item>
        <el-sub-menu v-else :index="menu.path">
          <template #title>
            <el-icon><component :is="menu.icon" /></el-icon>
            <span>{{ menu.title }}</span>
          </template>
          <el-menu-item v-for="child in menu.children" :key="child.path" :index="child.path">
            <el-icon><component :is="child.icon" /></el-icon>
            <template #title>{{ child.title }}</template>
          </el-menu-item>
        </el-sub-menu>
      </template>
    </el-menu>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import type { Role } from '@/types'

defineProps<{ collapsed: boolean }>()

const route = useRoute()
const authStore = useAuthStore()
const activeMenu = computed(() => {
  const path = route.path
  if (path.startsWith('/director/monitor/')) return path
  if (path.startsWith('/executor/tasks/')) return path
  return path
})

interface MenuItem { path: string; title: string; icon: string; requiresRole?: Role | Role[]; children?: MenuItem[] }

const menuConfig: Record<string, MenuItem[]> = {
  admin: [
    { path: '/admin', title: '系统概览', icon: 'DataAnalysis' },
    { path: '/admin/users', title: '用户管理', icon: 'User' },
    { path: '/admin/drills', title: '全部演练', icon: 'Monitor' },
  ],
  director: [
    { path: '/director', title: '指挥概览', icon: 'DataAnalysis' },
    { path: '/director/templates', title: '模板管理', icon: 'Document' },
    { path: '/director/create', title: '创建演练', icon: 'Plus' },
    { path: '/director/drills', title: '演练实例', icon: 'Monitor' },
  ],
  executor: [
    { path: '/executor', title: '我的任务', icon: 'Tickets' },
  ],
  viewer: [
    { path: '/viewer', title: '演练概览', icon: 'View' },
  ],
}

const visibleMenus = computed<MenuItem[]>(() => {
  const role = authStore.role as string
  return menuConfig[role] ?? menuConfig.viewer
})
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

// 深色侧边栏
.app-sidebar {
  width: $sidebar-width;
  height: calc(100vh - $header-height);
  position: fixed;
  left: 0;
  top: $header-height;
  background: #1E293B;
  border-right: 1px solid #334155;
  z-index: $z-index-sidebar;
  transition: width 0.3s ease;
  overflow: hidden;

  &.is-collapsed {
    width: $sidebar-collapsed-width;
  }
}

.sidebar-menu {
  border-right: none;
  background: transparent;

  :deep(.el-menu) {
    background: transparent;
  }

  :deep(.el-menu-item),
  :deep(.el-sub-menu__title) {
    height: 40px;
    min-height: 40px;
    color: #94A3B8;
    font-size: $font-size-sm;

    &:hover {
      color: #F1F5F9;
      background: #334155;
    }

    &.is-active {
      color: #F1F5F9;
      background: #55C3D3;
      border-right: 2px solid #0891B2;
    }
  }

  :deep(.el-sub-menu__title) {
    height: 40px;
    min-height: 40px;
  }

  :deep(.el-icon) {
    color: inherit;
  }
}
</style>
