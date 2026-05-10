<template>
  <div class="dashboard">
    <h1 class="page-title">仪表盘</h1>
    <div class="dashboard-content">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="stat-card">
            <template #header>
              <span>演练总数</span>
            </template>
            <div class="stat-value">0</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <template #header>
              <span>执行中</span>
            </template>
            <div class="stat-value running">0</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <template #header>
              <span>已完成</span>
            </template>
            <div class="stat-value completed">0</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <template #header>
              <span>异常</span>
            </template>
            <div class="stat-value issue">0</div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
// Dashboard view - stats overview
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

.dashboard {
  .page-title {
    font-size: $font-size-2xl;
    font-weight: $font-weight-bold;
    color: $text-primary;
    margin-bottom: $spacing-xl;
  }
  
  .stat-card {
    .stat-value {
      font-size: $font-size-3xl;
      font-weight: $font-weight-bold;
      color: $text-primary;
      
      &.running {
        color: $color-primary;
      }
      
      &.completed {
        color: $color-success;
      }
      
      &.issue {
        color: $color-error;
      }
    }
  }
}
</style>
