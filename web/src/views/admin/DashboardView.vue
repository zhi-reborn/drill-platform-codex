<template>
  <div class="page-container">
    <div class="page-header">
      <h2>系统概览</h2>
    </div>

    <div class="page-content">
      <!-- 统计卡片 -->
      <el-row :gutter="20" class="stats-row">
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-icon total">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                <polyline points="14,2 14,8 20,8"/>
              </svg>
            </div>
            <div class="stat-content">
              <div class="stat-label">总演练数</div>
              <div class="stat-value">{{ dashboardData.stats.total_drills }}</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-icon active">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <polyline points="12,6 12,12 16,14"/>
              </svg>
            </div>
            <div class="stat-content">
              <div class="stat-label">活跃演练</div>
              <div class="stat-value">{{ dashboardData.stats.active_drills }}</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-icon success">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
                <polyline points="22,4 12,14.01 9,11.01"/>
              </svg>
            </div>
            <div class="stat-content">
              <div class="stat-label">成功率</div>
              <div class="stat-value">{{ dashboardData.stats.success_rate }}%</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-icon team">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                <circle cx="9" cy="7" r="4"/>
                <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
                <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
              </svg>
            </div>
            <div class="stat-content">
              <div class="stat-label">团队在线</div>
              <div class="stat-value">{{ dashboardData.stats.team_online_count }}/{{ dashboardData.stats.team_total_count }}</div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 图表和表格 -->
      <el-row :gutter="20" class="charts-row">
        <el-col :span="24">
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
              <el-table-column prop="step_name" label="步骤" min-width="150" v-if="rowHasStep" />
              <el-table-column prop="operator" label="操作人" width="100" />
              <el-table-column prop="created_at" label="时间" width="160">
                <template #default="{ row }">
                  {{ formatTime(row.created_at) }}
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import BarChart from '@/components/charts/BarChart.vue'
import dashboardData from '@/mock/data/dashboard.json'

const loading = ref(false)

// 柱状图数据
const barChartData = computed(() => {
  return dashboardData.by_category.map(item => ({
    name: getCategoryLabel(item.category),
    value: item.count,
    category: item.category,
  }))
})

// 最近活动数据
const recentActivity = computed(() => {
  return dashboardData.recent_activity.slice(0, 5)
})

const rowHasStep = computed(() => {
  return recentActivity.value.some(item => item.step_name)
})

// 分类标签映射
function getCategoryLabel(category: string): string {
  const map: Record<string, string> = {
    disaster_recovery: '灾备切换',
    degradation: '服务降级',
    release: '发布演练',
    security: '安全事件',
  }
  return map[category] || category
}

// 活动类型标签
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

// 加载数据
async function loadDashboardData() {
  loading.value = true
  try {
    // mock 数据已直接导入
    ElMessage.success('数据加载成功')
  } catch (error) {
    ElMessage.error('加载数据失败')
    console.error('Failed to load dashboard data:', error)
  } finally {
    loading.value = false
  }
}

loadDashboardData()
</script>

<style scoped lang="scss">
@use '@/styles/layout' as *;
@use '@/styles/variables' as *;

.page-container {
  @include page-container;

  .page-header {
    @include page-header;
  }

  .page-content {
    @include page-content;

    .stats-row {
      margin-bottom: $spacing-base;

      .stat-card {
        @include stat-card;

        .stat-icon {
          width: 40px;
          height: 40px;
          border-radius: $radius-base;
          display: flex;
          align-items: center;
          justify-content: center;

          svg {
            width: 20px;
            height: 20px;
          }

          &.total {
            background: $color-primary-bg;
            color: $color-primary;
          }

          &.active {
            background: $color-info;
            color: $bg-primary;
          }

          &.success {
            background: $color-success-bg;
            color: $color-success;
          }

          &.team {
            background: rgba($color-secondary, 0.2);
            color: $color-secondary;
          }
        }

        .stat-content {
          flex: 1;

          .stat-label {
            font-size: $font-size-sm;
            color: $text-secondary;
            margin-bottom: $spacing-xs;
          }

          .stat-value {
            font-size: $font-size-xl;
            font-weight: $font-weight-bold;
            color: $text-primary;
          }
        }
      }
    }

    .charts-row {
      margin-top: $spacing-base;

      .chart-card,
      .table-card {
        @include card-compact;

        .card-title {
          font-size: $font-size-base;
          font-weight: $font-weight-semibold;
          color: $text-primary;
        }
      }

      .table-card {
        :deep(.el-table) {
          background: #{$bg-secondary};
          color: #{$text-primary};

          .el-table__header th {
            background: #{$bg-tertiary};
            color: #{$text-secondary};
          }

          .el-table__row td {
            background: #{$bg-secondary};
            border-color: #{$border-color-light};
          }

          .el-table__row--striped td {
            background: #{rgba(26, 31, 46, 0.5)};
          }
        }
      }
    }
  }
}
</style>
