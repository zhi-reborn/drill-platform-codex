<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">演练概览</h2>
      <el-button type="primary" @click="viewScreen">
        <el-icon><Monitor /></el-icon>
        查看监控大屏
      </el-button>
    </div>

    <div class="page-content">
      <!-- 统计卡片 -->
      <el-row :gutter="20" class="stats-row">
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-label">总演练数</div>
            <div class="stat-value">{{ dashboardData.stats.total_drills }}</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-label">活跃演练</div>
            <div class="stat-value">{{ dashboardData.stats.active_drills }}</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-label">成功率</div>
            <div class="stat-value success">{{ dashboardData.stats.success_rate }}%</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-label">平均耗时</div>
            <div class="stat-value">{{ formatDuration(dashboardData.stats.avg_step_duration_seconds) }}</div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 图表 -->
      <el-row :gutter="20" class="charts-row">
        <el-col :span="12">
          <el-card class="chart-card">
            <template #header>
              <span class="card-title">演练分类统计</span>
            </template>
            <BarChart :data="barChartData" height="320px" />
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card class="chart-card">
            <template #header>
              <span class="card-title">每小时错误数</span>
            </template>
            <LineChart :data="lineChartData" height="320px" />
          </el-card>
        </el-col>
      </el-row>

      <!-- 最近活动 -->
      <el-card class="table-card">
        <template #header>
          <span class="card-title">最近活动</span>
        </template>
        <el-table :data="recentActivity" stripe style="width: 100%">
          <el-table-column prop="type" label="类型" width="120">
            <template #default="{ row }">
              <el-tag :type="getActivityTypeTag(row.type)" size="small">
                {{ getActivityLabel(row.type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="drill_name" label="演练名称" min-width="180" />
          <el-table-column prop="operator" label="操作人" width="100" />
          <el-table-column prop="created_at" label="时间" width="160">
            <template #default="{ row }">
              {{ formatTime(row.created_at) }}
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Monitor } from '@element-plus/icons-vue'
import BarChart from '@/components/charts/BarChart.vue'
import LineChart from '@/components/charts/LineChart.vue'
import dashboardData from '@/mock/data/dashboard.json'

const router = useRouter()

const barChartData = computed(() => {
  return dashboardData.by_category.map(item => ({
    name: getCategoryLabel(item.category),
    value: item.count,
    category: item.category,
  }))
})

const lineChartData = computed(() => {
  return dashboardData.hourly_errors.map(item => ({
    name: item.hour,
    value: item.count,
  }))
})

const recentActivity = computed(() => {
  return dashboardData.recent_activity.slice(0, 5)
})

function getCategoryLabel(category: string): string {
  const map: Record<string, string> = {
    disaster_recovery: '灾备切换',
    degradation: '服务降级',
    release: '发布演练',
    security: '安全事件',
  }
  return map[category] || category
}

function getActivityTypeTag(type: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, any> = {
    drill_start: 'primary',
    drill_complete: 'success',
    drill_terminate: 'danger',
    step_start: 'info',
    step_complete: 'success',
  }
  return map[type] || 'info'
}

function getActivityLabel(type: string): string {
  const map: Record<string, string> = {
    drill_start: '演练开始',
    drill_complete: '演练完成',
    drill_terminate: '演练终止',
    step_start: '步骤开始',
    step_complete: '步骤完成',
  }
  return map[type] || type
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatDuration(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}m ${secs}s`
}

function viewScreen() {
  router.push('/screen/100')
}
</script>

<style scoped lang="scss">
@use '@/styles/layout' as *;
@use '@/styles/variables' as *;

.page-container {
  @include page-container;

  .page-header {
    @include page-header;

    .page-title {
      font-size: $font-size-xl;
      font-weight: $font-weight-bold;
      color: $text-primary;
      margin: 0;
    }
  }

  .page-content {
    @include page-content;

    .stats-row {
      margin-bottom: $spacing-base;

      .stat-card {
        @include stat-card;
        text-align: center;

        .stat-value {
          font-size: $font-size-xl;
          font-weight: $font-weight-bold;
          color: $text-primary;

          &.success {
            color: $color-success;
          }
        }
      }
    }

    .charts-row {
      margin-bottom: $spacing-base;

      .chart-card {
        @include card-compact;

        .card-title {
          font-size: $font-size-base;
          font-weight: $font-weight-semibold;
          color: $text-primary;
        }
      }
    }

    .table-card {
      @include card-compact;

      .card-title {
        font-size: $font-size-base;
        font-weight: $font-weight-semibold;
        color: $text-primary;
      }

      :deep(.el-table) {
        background: $bg-secondary;
        color: $text-primary;

        .el-table__header th {
          background: $bg-tertiary;
          color: $text-secondary;
        }

        .el-table__row td {
          background: $bg-secondary;
          border-color: $border-color-light;
        }

        .el-table__row--striped td {
          background: rgba(26, 31, 46, 0.5);
        }
      }
    }
  }
}
</style>
