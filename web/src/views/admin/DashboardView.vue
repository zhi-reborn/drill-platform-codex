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

      <!-- 活跃演练 -->
      <el-card class="section-card">
        <template #header>
          <span class="card-title">活跃演练</span>
        </template>
        <el-row :gutter="20">
          <el-col
            v-for="drill in activeDrills"
            :key="drill.id"
            :xs="24"
            :sm="12"
            :lg="8"
          >
            <el-card class="drill-card">
              <div class="drill-header">
                <span class="drill-name">{{ drill.name }}</span>
                <DrillStatusBadge :status="drill.status" type="drill" />
              </div>
              <div class="drill-progress">
                <el-progress
                  :percentage="Math.round(drill.completed_steps / drill.total_steps * 100)"
                  :stroke-width="8"
                />
              </div>
              <div class="drill-current-step">
                <span class="label">当前步骤:</span>
                <span class="value">{{ getCurrentStepName(drill.id) }}</span>
              </div>
              <div class="drill-actions">
                <el-button type="primary" size="small" @click="viewMonitor(drill.id)">
                  监控
                </el-button>
                <el-button type="success" size="small" @click="viewScreen(drill.id)">
                  <el-icon><Monitor /></el-icon>
                  大屏
                </el-button>
              </div>
            </el-card>
          </el-col>
        </el-row>
        <div v-if="activeDrills.length === 0" class="empty-tip">
          暂无活跃演练
        </div>
      </el-card>

      <!-- 最近活动 -->
      <el-card class="section-card">
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
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Monitor } from '@element-plus/icons-vue'
import type { DrillInstance, StepInstance } from '@/types'
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue'
import instancesData from '@/mock/data/instances.json'
import stepsData from '@/mock/data/steps.json'
import dashboardData from '@/mock/data/dashboard.json'

const router = useRouter()

const instances = ref<DrillInstance[]>([])

const activeDrills = computed(() => {
  return instances.value.filter(i => i.status === 'running' || i.status === 'paused')
})

const recentActivity = computed(() => {
  return dashboardData.recent_activity.slice(0, 5)
})

const rowHasStep = computed(() => {
  return recentActivity.value.some(item => item.step_name)
})

function getCurrentStepName(drillId: number): string {
  const drillSteps = (stepsData as StepInstance[]).filter(s => s.drill_id === drillId && s.status === 'running')
  if (drillSteps.length > 0) {
    return drillSteps[0].step_name
  }
  const pendingSteps = (stepsData as StepInstance[]).filter(s => s.drill_id === drillId && s.status === 'pending')
  if (pendingSteps.length > 0) {
    return pendingSteps[0].step_name
  }
  return '无'
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

function viewMonitor(drillId: number) {
  router.push(`/admin/drills`)
}

function viewScreen(drillId: number) {
  router.push(`/screen/${drillId}`)
}

async function loadInstances() {
  try {
    instances.value = instancesData as DrillInstance[]
  } catch (error) {
    ElMessage.error('加载数据失败')
    console.error('Failed to load instances:', error)
  }
}

loadInstances()
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

    .section-card {
      @include card-compact;
      margin-bottom: $spacing-base;

      .card-title {
        font-size: $font-size-base;
        font-weight: $font-weight-semibold;
        color: $text-primary;
      }

      .drill-card {
        background: $bg-tertiary;
        border-color: $border-color-light;
        margin-bottom: $spacing-sm;

        .drill-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: $spacing-sm;

          .drill-name {
            font-size: $font-size-base;
            font-weight: $font-weight-semibold;
            color: $text-primary;
          }
        }

        .drill-progress {
          margin-bottom: $spacing-sm;
        }

        .drill-current-step {
          font-size: $font-size-sm;
          color: $text-secondary;
          margin-bottom: $spacing-sm;

          .label {
            margin-right: $spacing-xs;
          }

          .value {
            color: $text-primary;
          }
        }

        .drill-actions {
          display: flex;
          gap: $spacing-xs;
          justify-content: flex-end;
        }
      }

      .empty-tip {
        text-align: center;
        color: $text-tertiary;
        padding: $spacing-base;
      }

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
</style>
