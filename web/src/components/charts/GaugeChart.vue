<template>
  <div ref="domRef" class="chart-container" :style="{ height: height }"></div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import type { EChartsOption } from 'echarts'
import { useECharts } from '@/composables/useECharts'
import { applyDarkTheme } from './theme'

interface GaugeChartData {
  name: string
  value: number
  max?: number
  min?: number
}

const props = withDefaults(
  defineProps<{
    data: GaugeChartData
    height?: string
    loading?: boolean
  }>(),
  {
    height: '250px',
    loading: false,
  }
)

const { domRef, setOptions, showLoading, hideLoading } = useECharts()

watch(
  () => props.data,
  (newData) => {
    const max = newData.max ?? 100
    const min = newData.min ?? 0
    const value = newData.value

    const option: EChartsOption = applyDarkTheme({
      series: [
        {
          name: newData.name,
          type: 'gauge',
          min: min,
          max: max,
          startAngle: 200,
          endAngle: -20,
          radius: '90%',
          center: ['50%', '55%'],
          progress: {
            show: true,
            width: 18,
            itemStyle: {
              color: function (params: any) {
                const val = params.value as number
                const percent = val / max
                if (percent >= 0.8) {
                  return '#2EA043'
                } else if (percent >= 0.5) {
                  return '#B8860B'
                } else {
                  return '#DA3633'
                }
              },
            },
          },
          pointer: {
            itemStyle: {
              color: 'auto',
            },
          },
          axisLine: {
            lineStyle: {
              width: 18,
              color: [[1, '#30363D']],
            },
          },
          axisTick: {
            show: false,
          },
          splitLine: {
            length: 15,
            lineStyle: {
              width: 2,
              color: '#8B949E',
            },
          },
          axisLabel: {
            distance: 25,
            color: '#8B949E',
            fontSize: 12,
          },
          anchor: {
            show: true,
            showAbove: true,
            size: 25,
            itemStyle: {
              borderWidth: 10,
            },
          },
          title: {
            show: true,
            offsetCenter: [0, '70%'],
            fontSize: 14,
            color: '#8B949E',
          },
          detail: {
            valueAnimation: true,
            fontSize: 24,
            fontWeight: 'bold',
            color: '#E0E6ED',
            offsetCenter: [0, '30%'],
            formatter: '{value}%',
          },
          data: [
            {
              value: value,
              name: newData.name,
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
