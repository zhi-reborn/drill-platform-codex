<template>
  <div class="step-group">
    <div class="step-row" :class="[depthClass, { 'step-selected': selectedStep?.id === node.id }]"
      @click="$emit('select', node)">
      <span class="drag-handle" v-if="depth === 0">⠿</span>
      <span class="seq-badge" :class="depthBadgeClass">{{ computeSEQ(node) }}</span>
      <button v-if="node.hasChildren" class="expand-btn" @click.stop="$emit('toggle-collapse', node.id)">
        {{ isCollapsed ? '▶' : '▼' }}
      </button>
      <span v-else class="expand-placeholder"></span>
      <el-tag v-if="depthLabel" size="small" :type="depthTagType" class="depth-tag">{{ depthLabel }}</el-tag>
      <span class="step-name">{{ node.name || '-' }}</span>
      <el-tag size="small" type="info">{{ getStepTypeLabel(node.step_type) }}</el-tag>
      <el-button text type="danger" size="small" @click.stop="$emit('remove', node)" title="删除">
        <el-icon>
          <Delete />
        </el-icon>
      </el-button>
    </div>
    <div v-if="node.children?.length && !isCollapsed" class="children-list" :style="childrenIndentStyle">
      <StepTreeNodeItem v-for="child in node.children" :key="child.id" :node="child" :depth="depth + 1"
        :selected-step="selectedStep" :collapsed-ids="collapsedIds" @select="$emit('select', $event)"
        @toggle-collapse="$emit('toggle-collapse', $event)" @remove="$emit('remove', $event)" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Delete } from '@element-plus/icons-vue'
import type { StepTemplate, StepType } from '@/types'

interface StepTreeNode extends StepTemplate {
  children?: StepTreeNode[]
  hasChildren?: boolean
}

const props = defineProps<{
  node: StepTreeNode
  depth: number
  selectedStep?: (StepTemplate & { parent_seq_display?: string }) | null
  collapsedIds: Set<number>
}>()

defineEmits<{
  select: [node: StepTreeNode]
  'toggle-collapse': [id: number]
  remove: [node: StepTreeNode]
}>()

const isCollapsed = computed(() => props.collapsedIds.has(props.node.id))

const depthClass = computed(() => `depth-${props.depth}`)

const depthBadgeClass = computed(() => `badge-depth-${Math.min(props.depth, 3)}`)

const depthLabel = computed(() => {
  const labels = ['阶段', '环节', '任务', '步骤']
  return labels[props.depth] || ''
})

const depthTagType = computed(() => {
  const types: Record<number, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = { 0: 'primary', 1: 'success', 2: 'warning', 3: 'info' }
  return types[props.depth] || 'info'
})

const childrenIndentStyle = computed(() => ({
  marginLeft: '4px',
  paddingLeft: '8px',
  borderLeft: '2px solid #E4E7ED',
}))

function computeSEQ(step: StepTreeNode): number {
  return step.order_index ?? 0
}

function getStepTypeLabel(type: string): string {
  const map: Record<string, string> = { serial: '串行', parallel: '并行' }
  return map[type] || type
}
</script>

<style scoped lang="scss">
.step-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  background: white;
  border: 1px solid transparent;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;

  &:hover {
    background: #ecf5ff;
    border-color: var(--el-color-primary-light-7);
  }
}

// Depth-based visual differentiation
.depth-0 {
  font-weight: 500;
  background: #f0f5ff;

  .step-name { color: #1d2129; }

  &:hover { background: #dbe8ff; }
}

.depth-1 {
  background: #f7f8fa;
  margin-left: 24px;
  border-left: 3px solid var(--el-color-primary-light-5);

  .step-name { color: #4e5969; }
}

.depth-2 {
  background: #fafafa;
  margin-left: 48px;
  border-left: 3px solid var(--el-color-success-light-5);

  .step-name { color: #606266; font-size: 12.5px; }
}

.depth-3 {
  background: #fdfdfd;
  margin-left: 72px;
  border-left: 3px solid var(--el-color-warning-light-5);

  .step-name { color: #86909c; font-size: 12px; }
}

.step-selected {
  color: var(--el-color-primary) !important;
  font-weight: 600;
  border-color: var(--el-color-primary-light-5) !important;
  background: #ecf5ff !important;
}

.drag-handle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: grab;
  color: #c0c4cc;
  font-size: 14px;
  width: 16px;
  height: 24px;
  user-select: none;
  flex-shrink: 0;

  &:hover { color: var(--el-color-primary); }
  &:active { cursor: grabbing; }
}

.seq-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  color: white;
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
}

.badge-depth-0 { background: var(--el-color-primary); }
.badge-depth-1 { background: var(--el-color-success); }
.badge-depth-2 { background: var(--el-color-warning); }
.badge-depth-3 { background: var(--el-color-info); }

.expand-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: #606266;
  font-size: 10px;
  width: 16px;
  text-align: center;
  padding: 0;
  flex-shrink: 0;

  &:hover { color: var(--el-color-primary); }
}

.expand-placeholder {
  width: 16px;
  flex-shrink: 0;
}

.step-name {
  flex: 1;
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.depth-tag {
  flex-shrink: 0;
}

.children-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-top: 2px;
}
</style>
