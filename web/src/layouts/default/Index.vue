<template>
  <el-container class="layout-container">
    <el-aside width="200px" class="sidebar">
      <div class="logo">演练平台</div>
      <el-menu :default-active="activeMenu" router background-color="#304156" text-color="#bfcbd9" active-text-color="#409EFF">
        <el-menu-item index="/display">
          <el-icon><Monitor /></el-icon>
          <span>指挥大屏</span>
        </el-menu-item>
        <el-menu-item index="/console/drill">
          <el-icon><List /></el-icon>
          <span>演练管理</span>
        </el-menu-item>
        <el-menu-item index="/workspace">
          <el-icon><User /></el-icon>
          <span>我的任务</span>
        </el-menu-item>
        <el-menu-item index="/template">
          <el-icon><Document /></el-icon>
          <span>模板管理</span>
        </el-menu-item>
        <el-menu-item index="/report">
          <el-icon><DataAnalysis /></el-icon>
          <span>报告中心</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <span class="breadcrumb">生产演练流程管理系统</span>
        <UserAvatar />
      </el-header>
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Monitor, List, User, Document, DataAnalysis } from '@element-plus/icons-vue'
import UserAvatar from '@/components/UserAvatar.vue'

const route = useRoute()
const activeMenu = computed(() => route.path)
</script>

<style scoped>
.layout-container {
  height: 100vh;
}
.sidebar {
  background-color: #304156;
}
.logo {
  height: 60px;
  line-height: 60px;
  text-align: center;
  font-size: 20px;
  font-weight: bold;
  color: #fff;
  background-color: #2b3a4b;
}
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #fff;
  border-bottom: 1px solid #e6e6e6;
}
.breadcrumb {
  font-size: 16px;
  font-weight: 500;
}
.main-content {
  background-color: #f0f2f5;
}
</style>
