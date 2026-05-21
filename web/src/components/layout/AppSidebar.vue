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
import { getVisibleMenus } from '@/config/menu'
import type { Role } from '@/types'

defineProps<{ collapsed: boolean }>()

const route = useRoute()
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
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

.app-sidebar {
  width: $sidebar-width;
  height: calc(100vh - $header-height);
  position: fixed;
  left: 0;
  top: $header-height;
  background: $bg-secondary;
  border-right: 1px solid $border-color;
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
}
</style>
