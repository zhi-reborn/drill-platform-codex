<template>
  <div ref="domRef" class="chart-container" :style="{ height: height }"></div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import type { EChartsOption } from 'echarts'
import { useECharts } from '@/composables/useECharts'
import { applyDarkTheme } from './theme'

interface LineChartData {
  name: string
  value: number
  timestamp?: string
}

const props = withDefaults(
  defineProps<{
    data: LineChartData[]
    height?: string
    loading?: boolean
  }>(),
  {
    height: '300px',
    loading: false,
  }
)

const { domRef, setOptions, showLoading, hideLoading } = useECharts()

watch(
  () => props.data,
  (newData) => {
    const categories = newData.map((item) => item.timestamp || item.name)
    const seriesData = newData.map((item) => item.value)

    const option: EChartsOption = applyDarkTheme({
      tooltip: {
        trigger: 'axis',
      },
      legend: {
        data: ['趋势'],
      },
      xAxis: {
        type: 'category',
        data: categories,
        boundaryGap: false,
      },
      yAxis: {
        type: 'value',
      },
      series: [
        {
          name: '趋势',
          type: 'line',
          smooth: true,
          data: seriesData,
          areaStyle: {
            opacity: 0.3,
          },
          lineStyle: {
            width: 3,
          },
        },
      ],
    })

    setOptions(option)
  },
  { immediate: true }
)

watch(
  () => props.loading,
  (newLoading) => {
    if (newLoading) {
      showLoading()
    } else {
      hideLoading()
    }
  }
)
</script>

<style scoped>
.chart-container {
  width: 100%;
  min-height: 200px;
}
</style>
