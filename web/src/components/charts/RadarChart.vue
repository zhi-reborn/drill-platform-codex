<template>
  <div ref="domRef" class="chart-container" :style="{ height: height }"></div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import type { EChartsOption } from 'echarts'
import { useECharts } from '@/composables/useECharts'
import { applyDarkTheme } from './theme'

interface RadarChartData {
  name: string
  value: number
  max?: number
}

const props = withDefaults(
  defineProps<{
    data: RadarChartData[]
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
    const indicator = newData.map((item) => ({
      name: item.name,
      max: item.max ?? 100,
    }))

    const values = newData.map((item) => item.value)

    const option: EChartsOption = applyDarkTheme({
      tooltip: {
        trigger: 'item',
      },
      legend: {
        data: ['维度对比'],
        orient: 'vertical',
        right: 10,
        top: 10,
      },
      radar: {
        indicator: indicator,
        shape: 'circle',
        splitNumber: 5,
        axisName: {
          color: '#8B949E',
          fontSize: 13,
        },
        splitLine: {
          lineStyle: {
            color: '#30363D',
          },
        },
        splitArea: {
          show: false,
        },
        axisLine: {
          lineStyle: {
            color: '#30363D',
          },
        },
      },
      series: [
        {
          name: '维度对比',
          type: 'radar',
          data: [
            {
              value: values,
              name: '维度对比',
              areaStyle: {
                color: 'rgba(85, 195, 211, 0.3)',
              },
              lineStyle: {
                color: '#55C3D3',
                width: 2,
              },
              itemStyle: {
                color: '#55C3D3',
              },
            },
          ],
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
