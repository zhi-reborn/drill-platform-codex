<template>
  <aside class="app-sidebar" :class="{ 'is-collapsed': collapsed }">
    <el-menu :default-active="activeMenu" :collapse="collapsed" :collapse-transition="true" class="sidebar-menu" @select="handleMenuSelect">
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
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getVisibleMenus } from '@/config/menu'
import type { Role } from '@/types'

defineProps<{ collapsed: boolean }>()

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const activeMenu = computed(() => {
  const { path } = route
  if (path.startsWith('/director/monitor/')) return path
  if (path.startsWith('/executor/tasks/')) return path
  if (path.startsWith('/viewer/drills/')) return path
  return path
})

const visibleMenus = computed(() => {
  const role = authStore.role
  return getVisibleMenus(role as Role)
})

function handleMenuSelect(index: string) {
  if (index !== route.path) {
    router.push(index)
  }
}
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

.app-sidebar {
  width: $sidebar-width;
  height: calc(100vh - $header-height);
  position: fixed;
  left: 0;
  top: $header-height;
  background: $sidebar-bg;
  border-right: 1px solid rgba(255, 255, 255, 0.06);
  z-index: $z-index-sidebar;
  transition: width 0.3s cubic-bezier(0.2, 0, 0, 1);
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
    border-right: none;
  }

  // 一级菜单标题
  :deep(.el-sub-menu__title) {
    color: $sidebar-text !important;

    &:hover {
      background: $sidebar-bg-hover !important;
      color: $sidebar-text-active !important;
    }

    .el-icon {
      color: $sidebar-text;
    }
  }

  // 一级菜单激活
  :deep(.el-sub-menu.is-active > .el-sub-menu__title) {
    color: $sidebar-text-active !important;

    .el-icon {
      color: $sidebar-text-active !important;
    }
  }

  // 子菜单项
  :deep(.el-menu--inline .el-menu-item) {
    background-color: rgba(0, 0, 0, 0.12) !important;
    color: $sidebar-text;

    &:hover {
      background-color: $sidebar-bg-hover !important;
      color: $sidebar-text-active !important;
    }

    &.is-active {
      background-color: $sidebar-active-bg !important;
      color: $color-accent !important;
    }
  }

  // 无子菜单的菜单项
  :deep(.el-menu-item) {
    color: $sidebar-text;

    &:hover {
      background: $sidebar-bg-hover !important;
      color: $sidebar-text-active !important;
    }

    &.is-active {
      background-color: $sidebar-active-bg !important;
      color: $color-accent !important;
    }
  }

  // 折叠状态
  :deep(.el-menu--collapse) {
    width: $sidebar-collapsed-width;

    .el-menu-item.is-active {
      background-color: $sidebar-active-bg !important;
    }
  }
}
</style>
