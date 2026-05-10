<template>
  <div class="app-layout">
    <AppHeader @toggle-sidebar="collapsed = !collapsed" />
    <AppSidebar :collapsed="collapsed" />
    <main class="app-main">
      <div class="app-breadcrumb">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item v-for="item in breadcrumbs" :key="item.path">
            {{ item.title }}
          </el-breadcrumb-item>
        </el-breadcrumb>
      </div>
      <div class="app-content">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import AppHeader from './AppHeader.vue'
import AppSidebar from './AppSidebar.vue'

const collapsed = ref(false)
const route = useRoute()

const breadcrumbs = computed(() => {
  return route.matched
    .filter(r => r.meta?.title)
    .map(r => ({ path: r.path, title: r.meta.title as string }))
})
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

.app-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.app-main {
  margin-left: $sidebar-width;
  padding-top: $header-height;
  transition: margin-left 0.3s ease;
  min-height: calc(100vh - $header-height);
  background: #F8FAFC;
}

.app-breadcrumb {
  padding: $spacing-sm $spacing-base;
  background: #FFFFFF;
  border-bottom: 1px solid #E2E8F0;

  :deep(.el-breadcrumb__inner) {
    color: #64748B;
  }
  :deep(.el-breadcrumb__inner.is-link:hover) {
    color: #0891B2;
  }
}

.app-content {
  padding: $spacing-base;
  overflow-x: hidden;
}

.app-layout:has(.app-sidebar.is-collapsed) .app-main {
  margin-left: $sidebar-collapsed-width;
}
</style>
