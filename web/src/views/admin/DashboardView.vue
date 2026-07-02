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
                <div class="stat-value">{{ stats.total_drills }}</div>
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
                <div class="stat-value">{{ stats.active_drills }}</div>
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
                <div class="stat-value">{{ stats.success_rate }}%</div>
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
                <div class="stat-value">{{ stats.team_online_count }}/{{ stats.team_total_count }}</div>
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
                  :percentage="drill.progress_pct"
                  :stroke-width="8"
                />
              </div>
              <div class="drill-current-step">
                <span class="label">当前步骤:</span>
                <span class="value">{{ getCurrentStepName(drill.id) }}</span>
              </div>
              <div class="drill-actions">
                <el-button type="primary" size="small" @click="viewMonitor(drill.id)">
                  详情
                </el-button>
                <el-button type="success" size="small" @click="viewScreen(drill.id)">
                  <el-icon><Monitor /></el-icon>
                  大屏
                </el-button>
                <el-button type="warning" size="small" @click.stop="viewScreen2(drill.id)">
                  <el-icon><DataBoard /></el-icon>
                  大屏2
                </el-button>
                <el-button type="warning" size="small" @click.stop="viewScreen3(drill.id)">
                  <el-icon><DataBoard /></el-icon>
                  大屏3
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
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Monitor, DataBoard } from '@element-plus/icons-vue'
import type { DrillInstance, StepInstance } from '@/types'
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue'
import { drillApi } from '@/api/modules/drill'

const router = useRouter()

const instances = ref<DrillInstance[]>([])
const stepsMap = ref<Map<number, StepInstance[]>>(new Map())

const stats = ref({
  total_drills: 0,
  active_drills: 0,
  success_rate: 0,
  team_online_count: 0,
  team_total_count: 0,
})

const activeDrills = computed(() => {
  return instances.value.filter(i => i.status === 'running' || i.status === 'paused')
})

const recentActivity = ref<any[]>([])

const rowHasStep = computed(() => {
  return recentActivity.value.some(item => item.step_name)
})

function getCurrentStepName(drillId: number): string {
  const drillSteps = stepsMap.value.get(drillId) || []
  const runningStep = drillSteps.find(s => s.status === 'running')
  if (runningStep) {
    return runningStep.name
  }
  const pendingStep = drillSteps.find(s => s.status === 'pending')
  if (pendingStep) {
    return pendingStep.name
  }
  return '无'
}

function getActivityTypeTag(type: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, any> = {
    start: 'primary',
    pause: 'warning',
    resume: 'primary',
    terminate: 'danger',
    drill_start: 'primary',
    drill_complete: 'success',
    drill_terminate: 'danger',
    drill_pause: 'warning',
    drill_resume: 'primary',
    step_start: 'info',
    step_complete: 'success',
  }
  return map[type] || 'info'
}

function getActivityLabel(type: string): string {
  const map: Record<string, string> = {
    start: '演练启动',
    pause: '演练暂停',
    resume: '演练恢复',
    terminate: '演练终止',
    drill_start: '演练开始',
    drill_complete: '演练完成',
    drill_terminate: '演练终止',
    drill_pause: '演练暂停',
    drill_resume: '演练恢复',
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
  router.push(`/director/monitor/${drillId}`)
}

function viewScreen(drillId: number) {
  router.push(`/screen/${drillId}`)
}

function viewScreen2(drillId: number) {
  router.push(`/admin/screen/${drillId}`)
}

function viewScreen3(drillId: number) {
  router.push(`/screen3/${drillId}`)
}

async function loadDashboard() {
  try {
    // 加载演练列表
    const result = await drillApi.getList({ page: 1, page_size: 50 })
    instances.value = result.list
    
    // 计算统计数据
    stats.value.total_drills = result.total
    stats.value.active_drills = activeDrills.value.length
    const completed = instances.value.filter(i => i.status === 'completed').length
    stats.value.success_rate = instances.value.length > 0 
      ? Math.round((completed / instances.value.length) * 100) 
      : 0
    stats.value.team_online_count = 0 // TODO: 需要用户在线状态 API
    stats.value.team_total_count = 0 // TODO: 需要用户总数 API
    
    // 加载每个演练的步骤
    for (const drill of activeDrills.value) {
      try {
        const steps = await drillApi.getSteps(drill.id)
        stepsMap.value.set(drill.id, steps)
      } catch (e) {
        console.error(`Failed to load steps for drill ${drill.id}`, e)
      }
    }
    
    // 加载演练日志作为最近活动
    const allLogs: any[] = []
    for (const drill of instances.value.slice(0, 10)) {
      try {
        const logs = await drillApi.getLogs(drill.id)
        logs.forEach((log: any) => {
          allLogs.push({
            type: log.action,
            drill_name: drill.name,
            step_name: null,
            operator: log.operator_name || '流程引擎',
            created_at: log.created_at,
          })
        })
      } catch (e) {
        // 忽略错误
      }
    }
    // 按时间排序，取最新 5 条
    recentActivity.value = allLogs
      .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
      .slice(0, 5)
  } catch (error) {
    ElMessage.error('加载数据失败')
    console.error('Failed to load dashboard:', error)
  }
}

onMounted(() => {
  loadDashboard()
})
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
            background: $color-accent-bg;
            color: $color-accent;
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
            background: rgba($text-tertiary, 0.2);
            color: $text-tertiary;
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
