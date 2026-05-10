<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">演练详情</h2>
      <el-button @click="router.back()">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
    </div>

    <div v-if="instance" class="page-content">
      <!-- 基本信息 -->
      <el-card class="info-card">
        <div class="info-header">
          <div class="info-title">
            <h3>{{ instance.name }}</h3>
            <DrillStatusBadge :status="instance.status" type="drill" />
          </div>
          <div class="info-progress">
            <span class="progress-label">
              进度：{{ instance.completed_steps }} / {{ instance.total_steps }}
            </span>
            <el-progress
              :percentage="Math.round(instance.completed_steps / instance.total_steps * 100)"
              :stroke-width="8"
              :status="instance.status === 'completed' ? 'success' : undefined"
            />
          </div>
        </div>
        <el-descriptions :column="2" border class="info-descriptions">
          <el-descriptions-item label="演练模板">{{ instance.template_name }}</el-descriptions-item>
          <el-descriptions-item label="创建人">{{ instance.created_by_name }}</el-descriptions-item>
          <el-descriptions-item label="开始时间">
            {{ instance.started_at ? formatTime(instance.started_at) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatTime(instance.created_at) }}
          </el-descriptions-item>
          <el-descriptions-item v-if="instance.completed_at" label="完成时间" :span="2">
            {{ formatTime(instance.completed_at) }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 步骤列表 -->
      <el-card class="steps-card">
        <template #header>
          <span class="card-title">步骤列表</span>
        </template>
        <el-table :data="drillSteps" style="width: 100%">
          <el-table-column prop="order_index" label="序号" width="80" />
          <el-table-column prop="step_name" label="步骤名" min-width="200" />
          <el-table-column prop="step_type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag :type="getStepTypeTag(row.step_type)" size="small">
                {{ getStepTypeLabel(row.step_type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="120">
            <template #default="{ row }">
              <DrillStatusBadge :status="row.status" type="step" />
            </template>
          </el-table-column>
          <el-table-column prop="assignee_name" label="执行人" width="120" />
          <el-table-column label="耗时" width="120">
            <template #default="{ row }">
              {{ calculateDuration(row) }}
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 时间线图表 -->
      <el-card class="timeline-card">
        <template #header>
          <span class="card-title">执行时间线</span>
        </template>
        <TimelineChart :data="timelineData" height="200px" />
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import type { DrillInstance, StepInstance } from '@/types'

interface TimelineItem {
  startTime: string
  endTime?: string
  status: 'pending' | 'running' | 'completed' | 'timeout' | 'skipped' | 'issue'
}
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue'
import TimelineChart from '@/components/charts/TimelineChart.vue'
import instancesData from '@/mock/data/instances.json'
import stepsData from '@/mock/data/steps.json'

const route = useRoute()
const router = useRouter()

const instance = ref<DrillInstance | null>(null)
const steps = ref<StepInstance[]>([])

const drillId = computed(() => {
  const id = route.params.id
  return typeof id === 'string' ? parseInt(id, 10) : 0
})

const drillSteps = computed(() => {
  return steps.value.filter(s => s.drill_id === drillId.value).sort((a, b) => a.order_index - b.order_index)
})

const timelineData = computed(() => {
  return drillSteps.value.map(step => ({
    name: step.step_name,
    items: [{
      startTime: step.started_at || new Date().toISOString(),
      endTime: step.completed_at,
      status: step.status as TimelineItem['status'],
    }],
  }))
})

function getStepTypeLabel(type: string): string {
  const map: Record<string, string> = {
    serial: '串行',
    parallel: '并行',
    any_of: '任选',
    condition: '条件',
  }
  return map[type] || type
}

function getStepTypeTag(type: string): 'primary' | 'success' | 'warning' | 'info' {
  const map: Record<string, any> = {
    serial: 'primary',
    parallel: 'success',
    any_of: 'warning',
    condition: 'info',
  }
  return map[type] || 'info'
}

function calculateDuration(step: StepInstance): string {
  if (!step.started_at || !step.completed_at) {
    return '-'
  }
  const start = new Date(step.started_at).getTime()
  const end = new Date(step.completed_at).getTime()
  const diff = Math.floor((end - start) / 1000)
  if (diff < 60) {
    return `${diff}s`
  }
  const mins = Math.floor(diff / 60)
  const secs = diff % 60
  return `${mins}m ${secs}s`
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

async function loadDrillData() {
  try {
    instance.value = (instancesData.find(i => i.id === drillId.value) as DrillInstance) || null
    steps.value = stepsData as StepInstance[]
    if (!instance.value) {
      ElMessage.error('演练不存在')
      router.back()
    }
  } catch (error) {
    ElMessage.error('加载数据失败')
    console.error('Failed to load drill data:', error)
  }
}

onMounted(() => {
  loadDrillData()
})
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
    display: flex;
    flex-direction: column;
    gap: $spacing-base;

    .info-card {
      @include card-compact;

      .info-header {
        margin-bottom: $spacing-base;

        .info-title {
          display: flex;
          align-items: center;
          gap: $spacing-sm;
          margin-bottom: $spacing-sm;

          h3 {
            font-size: $font-size-lg;
            color: $text-primary;
            margin: 0;
          }
        }

        .info-progress {
          .progress-label {
            display: block;
            font-size: $font-size-sm;
            color: $text-secondary;
            margin-bottom: $spacing-xs;
          }
        }
      }

      .info-descriptions {
        :deep(.el-descriptions__label) {
          background: $bg-tertiary;
          color: $text-primary;
        }

        :deep(.el-descriptions__content) {
          color: $text-secondary;
        }
      }
    }

    .steps-card,
    .timeline-card {
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
