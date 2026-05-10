<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">演练概览</h2>
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
import dashboardData from '@/mock/data/dashboard.json'
import instancesData from '@/mock/data/instances.json'
import stepsData from '@/mock/data/steps.json'

const router = useRouter()

const instances = ref<DrillInstance[]>([])

const activeDrills = computed(() => {
  return instances.value.filter(i => i.status === 'running' || i.status === 'paused')
})

const recentActivity = computed(() => {
  return dashboardData.recent_activity.slice(0, 5)
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
  const map: Record<string, 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
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
          text-align: right;
        }
      }

      .empty-tip {
        text-align: center;
        color: $text-tertiary;
        padding: $spacing-base;
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
