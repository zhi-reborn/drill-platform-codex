<template>
  <div ref="chartContainer" class="drill-flow-chart" aria-label="演练流程图"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'

interface FlowNode {
  id: string | number
  name: string
  status: 'pending' | 'running' | 'completed' | 'issue' | 'skipped'
  x: number
  y: number
}

interface FlowLink {
  source: string | number
  target: string | number
}

interface FlowChartProps {
  nodes: FlowNode[]
  links: FlowLink[]
  width?: number
  height?: number
}

const props = withDefaults(defineProps<FlowChartProps>(), {
  width: 800,
  height: 600
})

const chartContainer = ref<HTMLElement | null>(null)
let chartInstance: echarts.ECharts | null = null

// 状态颜色映射 (符合设计系统暗色主题)
const statusColors: Record<string, string> = {
  pending: '#64748B',    // 灰色 - 待命
  running: '#3B82F6',    // 蓝色 - 执行中
  completed: '#22C55E',  // 绿色 - 已完成
  issue: '#EF4444',      // 红色 - 异常
  skipped: '#94A3B8'     // 浅灰 - 跳过
}

// 初始化图表
const initChart = () => {
  if (!chartContainer.value) return

  chartInstance = echarts.init(chartContainer.value, null, {
    renderer: 'svg',
    width: props.width,
    height: props.height
  })

  updateChart()
}

// 更新图表数据
const updateChart = () => {
  if (!chartInstance) return

  const option: EChartsOption = {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'item',
      formatter: (params: any) => {
        if (params.dataType === 'node') {
          const statusMap: Record<string, string> = {
            pending: '待命',
            running: '执行中',
            completed: '已完成',
            issue: '异常',
            skipped: '跳过'
          }
          return `${params.name}<br/>状态：${statusMap[params.value.status]}`
        }
        return params.name || ''
      }
    },
    series: [
      {
        type: 'graph',
        layout: 'none',
        symbol: 'circle',
        symbolSize: 50,
        roam: true,
        label: {
          show: true,
          position: 'bottom',
          color: '#F8FAFC',
          fontSize: 12,
          fontFamily: 'Fira Sans, sans-serif'
        },
        edgeSymbol: ['circle', 'arrow'],
        edgeSymbolSize: [4, 10],
        edgeLabel: {
          fontSize: 10,
          color: '#94A3B8'
        },
        data: props.nodes.map(node => ({
          id: String(node.id),
          name: node.name,
          value: { status: node.status },
          x: node.x,
          y: node.y,
          itemStyle: {
            color: statusColors[node.status],
            shadowBlur: 10,
            shadowColor: statusColors[node.status]
          }
        })),
        links: props.links.map(link => ({
          source: link.source,
          target: link.target,
          lineStyle: {
            color: '#475569',
            width: 2,
            curveness: 0.1
          }
        })),
        lineStyle: {
          opacity: 0.9,
          width: 2,
          curveness: 0
        }
      }
    ]
  }

  chartInstance.setOption(option, true)
}

// 监听数据变化
watch(
  () => [props.nodes, props.links],
  () => {
    updateChart()
  },
  { deep: true }
)

// 监听尺寸变化
watch(
  () => [props.width, props.height],
  ([newWidth, newHeight]) => {
    if (chartInstance) {
      chartInstance.resize({
        width: newWidth,
        height: newHeight
      })
    }
  }
)

// 生命周期
onMounted(() => {
  initChart()
})

onBeforeUnmount(() => {
  if (chartInstance) {
    chartInstance.dispose()
    chartInstance = null
  }
})

// 暴露方法供父组件调用
defineExpose({
  resize: () => {
    chartInstance?.resize()
  },
  getInstance: () => chartInstance
})
</script>

<style scoped>
.drill-flow-chart {
  width: 100%;
  height: 100%;
  min-height: 400px;
  background-color: var(--color-background, #020617);
  border-radius: 8px;
  border: 1px solid var(--color-border, #334155);
}
</style>
