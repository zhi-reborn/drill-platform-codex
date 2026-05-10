<template>
  <div ref="domRef" class="chart-container" :style="{ height: height }"></div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import type { EChartsOption } from 'echarts'
import { useECharts } from '@/composables/useECharts'
import { applyDarkTheme } from './theme'

interface HeatmapChartData {
  x: string
  y: string
  value: number
}

const props = withDefaults(
  defineProps<{
    data: HeatmapChartData[]
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
    const xValues = Array.from(new Set(newData.map((item) => item.x)))
    const yValues = Array.from(new Set(newData.map((item) => item.y)))

    const heatmapData = newData.map((item) => {
      const xIndex = xValues.indexOf(item.x)
      const yIndex = yValues.indexOf(item.y)
      return [xIndex, yIndex, item.value]
    })

    const maxValue = Math.max(...newData.map((item) => item.value), 1)

    const option: EChartsOption = applyDarkTheme({
      tooltip: {
        position: 'top',
        formatter: (params: any) => {
          const xIndex = params.value[0]
          const yIndex = params.value[1]
          const value = params.value[2]
          return `${xValues[xIndex]} - ${yValues[yIndex]}: ${value}`
        },
      },
      grid: {
        height: '70%',
        top: '15%',
      },
      xAxis: {
        type: 'category',
        data: xValues,
        splitArea: {
          show: false,
        },
        axisLabel: {
          rotate: 45,
        },
      },
      yAxis: {
        type: 'category',
        data: yValues,
        splitArea: {
          show: false,
        },
      },
      visualMap: {
        min: 0,
        max: maxValue,
        calculable: true,
        orient: 'horizontal',
        left: 'center',
        bottom: '0%',
        inRange: {
          color: ['#1A1F2E', '#55C3D3', '#DA3633'],
        },
        textStyle: {
          color: '#8B949E',
        },
      },
      series: [
        {
          name: '热力图',
          type: 'heatmap',
          data: heatmapData,
          label: {
            show: true,
            color: '#E0E6ED',
            fontSize: 11,
          },
          itemStyle: {
            emphasis: {
              shadowBlur: 10,
              shadowColor: 'rgba(0, 0, 0, 0.5)',
            },
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
