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
              进度：{{ instance?.completed_steps ?? 0 }} / {{ instance?.total_steps ?? 0 }}
            </span>
            <el-progress
              :percentage="Math.round((instance?.completed_steps ?? 0) / (instance?.total_steps ?? 1) * 100)"
              :stroke-width="8"
              :status="instance.status === 'completed' ? 'success' : undefined"
            />
          </div>
        </div>
        <el-descriptions :column="2" border class="info-descriptions">
          <el-descriptions-item label="演练模板">{{ instance.template_name }}</el-descriptions-item>
          <el-descriptions-item label="创建人">{{ instance.created_by_name }}</el-descriptions-item>
          <el-descriptions-item label="开始时间">
            {{ (instance as any)?.start_time || instance.started_at ? formatTime((instance as any)?.start_time || instance.started_at) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatTime(instance.created_at) }}
          </el-descriptions-item>
          <el-descriptions-item v-if="(instance as any)?.end_time || instance.completed_at" label="完成时间" :span="2">
            {{ formatTime((instance as any)?.end_time || instance.completed_at) }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 步骤列表（按 phase 分组） -->
      <el-card class="steps-card">
        <template #header>
          <span class="card-title">步骤列表</span>
        </template>
        <el-collapse v-model="activePhases" class="phase-collapse">
          <el-collapse-item v-for="(group, phaseName) in groupedSteps" :key="phaseName" :name="phaseName">
            <template #title>
              <div class="phase-title">
                <span class="phase-name">{{ phaseName || '未分组' }}</span>
                <el-tag size="small" type="info">{{ Object.keys(groupedSteps).length > 0 ? (group as any[]).length : 0 }} 步</el-tag>
              </div>
            </template>
            <div class="step-list">
              <div v-for="step in (group as any[])" :key="step.id" class="step-item">
                <div class="step-main">
                  <div class="step-header">
                    <span class="step-order">#{{ step.seq ?? step.order_index ?? 0 }}</span>
                    <span class="step-name">{{ step.name || step.step_name }}</span>
                    <DrillStatusBadge :status="step.status" type="step" />
                  </div>
                  <div class="step-meta">
                    <span v-if="step.attributes?.responsible_department" class="meta-item">
                      <el-icon><OfficeBuilding /></el-icon>
                      责任部门：{{ step.attributes.responsible_department }}
                    </span>
                    <span v-if="step.attributes?.operator" class="meta-item">
                      <el-icon><Avatar /></el-icon>
                      操作人：{{ step.attributes.operator }}
                    </span>
                    <span v-if="step.estimated_duration_minutes" class="meta-item">
                      <el-icon><Timer /></el-icon>
                      预计：{{ step.estimated_duration_minutes }}分钟
                    </span>
                    <span v-if="getStartTime(step) && getEndTime(step)" class="meta-item">
                      <el-icon><Clock /></el-icon>
                      实际：{{ calculateDuration(step) }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </el-collapse-item>
        </el-collapse>
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
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, UserFilled, Avatar, Timer, Clock } from '@element-plus/icons-vue'
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
const activePhases = ref<string[]>([])

const drillId = computed(() => {
  const id = route.params.id
  return typeof id === 'string' ? parseInt(id, 10) : 0
})

const drillSteps = computed(() => {
  return steps.value.filter(s => s.drill_instance_id === drillId.value).sort((a, b) => {
    const ai = (a as any).order_index ?? a.seq ?? 0
    const bi = (b as any).order_index ?? b.seq ?? 0
    return ai - bi
  })
})

// 按 phase 分组步骤
const groupedSteps = computed(() => {
  const groups: Record<string, StepInstance[]> = {}
  for (const step of drillSteps.value) {
    const phase = step.phase || '未分组'
    if (!groups[phase]) {
      groups[phase] = []
    }
    groups[phase].push(step)
  }
  // 按 phase 名称排序，未分组放最后
  const sorted: Record<string, StepInstance[]> = {}
  const keys = Object.keys(groups).sort((a, b) => {
    if (a === '未分组') return 1
    if (b === '未分组') return -1
    return a.localeCompare(b, 'zh-CN')
  })
  for (const key of keys) {
    sorted[key] = groups[key]
  }
  return sorted
})

interface StepWithChildren extends StepInstance {
  children?: StepWithChildren[]
}

// 将扁平步骤数据转换为树形结构
const drillStepTree = computed(() => {
  const sorted = drillSteps.value
  const stepMap = new Map<number, StepWithChildren>()

  for (const step of sorted) {
    stepMap.set(step.id, { ...(step as any), step_name: step.name })
  }

  const roots: StepWithChildren[] = []

  for (const step of sorted) {
    const node = stepMap.get(step.id)!
    const parentId = (step as any).parent_step_id
    if (parentId && stepMap.has(parentId)) {
      const parent = stepMap.get(parentId)!
      if (!parent.children) {
        parent.children = []
      }
      parent.children.push(node)
    } else {
      roots.push(node)
    }
  }

  return roots
})

// 父步骤状态聚合显示 (只读模式)
function getStepStatusText(row: StepWithChildren): string {
  if (!row.children || row.children.length === 0) {
    return row.status
  }
  const total = row.children.length
  const completed = row.children.filter(c => c.status === 'completed' || c.status === 'skipped').length
  return `${completed}/${total} 子任务已完成`
}

function isParentStep(row: StepWithChildren): boolean {
  return !!(row.children && row.children.length > 0)
}

const timelineData = computed(() => {
  return drillSteps.value.map(step => ({
    name: step.name,
    items: [{
      startTime: step.start_time || new Date().toISOString(),
      endTime: step.end_time ?? undefined,
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

function getStartTime(step: StepInstance): string | undefined {
  return (step as any).started_at || step.start_time
}

function getEndTime(step: StepInstance): string | undefined {
  return (step as any).completed_at || step.end_time
}

function calculateDuration(step: StepInstance): string {
  const startTime = (step as any).started_at || step.start_time
  const endTime = (step as any).completed_at || step.end_time
  if (!startTime || !endTime) {
    return '-'
  }
  const start = new Date(startTime).getTime()
  const end = new Date(endTime).getTime()
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
    const mockInstance = instancesData.find(i => i.id === drillId.value)
    instance.value = mockInstance ? mockInstance as unknown as DrillInstance : null
    steps.value = stepsData as unknown as StepInstance[]
    if (!instance.value) {
      ElMessage.error('演练不存在')
      router.back()
    }
  } catch (error) {
    ElMessage.error('加载数据失败')
    console.error('Failed to load drill data:', error)
  }
}

// 数据加载后默认展开所有分组
watch(groupedSteps, (groups) => {
  if (activePhases.value.length === 0) {
    activePhases.value = Object.keys(groups)
  }
}, { immediate: true })

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

      :deep(.el-collapse) {
        border: none;

        .el-collapse-item__header {
          background: $bg-tertiary;
          padding: $spacing-sm $spacing-base;
          border-radius: $radius-sm;
          margin-bottom: $spacing-xs;
          height: auto;
          line-height: normal;

          .phase-title {
            display: flex;
            align-items: center;
            gap: $spacing-sm;
            width: 100%;
            font-size: $font-size-base;
            font-weight: $font-weight-semibold;
            color: $text-primary;

            .phase-name {
              flex: 1;
            }
          }
        }

        .el-collapse-item__content {
          padding-bottom: $spacing-base;
        }
      }

      .step-list {
        display: flex;
        flex-direction: column;
        gap: $spacing-sm;

        .step-item {
          display: flex;
          align-items: flex-start;
          padding: $spacing-sm $spacing-base;
          background: $bg-tertiary;
          border-radius: $radius-sm;
          border-left: 3px solid var(--el-color-primary);
          transition: background 0.2s;

          &:hover {
            background: rgba(64, 158, 255, 0.08);
          }

          .step-main {
            flex: 1;
            min-width: 0;

            .step-header {
              display: flex;
              align-items: center;
              gap: $spacing-sm;
              margin-bottom: $spacing-xs;

              .step-order {
                font-size: $font-size-xs;
                color: $text-tertiary;
                font-family: monospace;
                min-width: 28px;
              }

              .step-name {
                font-size: $font-size-base;
                color: $text-primary;
                font-weight: $font-weight-medium;
                flex: 1;
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
              }
            }

            .step-meta {
              display: flex;
              flex-wrap: wrap;
              gap: $spacing-base;
              font-size: $font-size-xs;
              color: $text-secondary;

              .meta-item {
                display: inline-flex;
                align-items: center;
                gap: 4px;

                .el-icon {
                  font-size: 12px;
                  opacity: 0.7;
                }
              }
            }
          }
        }
      }
    }
  }
}
</style>
