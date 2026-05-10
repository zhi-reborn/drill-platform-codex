<template>
  <div ref="domRef" class="chart-container" :style="{ height: height }"></div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import type { EChartsOption } from 'echarts'
import { useECharts } from '@/composables/useECharts'
import { applyDarkTheme } from './theme'

interface BarChartData {
  name: string
  value: number
  category?: string
}

const props = withDefaults(
  defineProps<{
    data: BarChartData[]
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
    const categories = newData.map((item) => item.category || item.name)
    const seriesData = newData.map((item) => ({
      name: item.name,
      value: item.value,
    }))

    const option: EChartsOption = applyDarkTheme({
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'shadow',
        },
      },
      legend: {
        data: ['数值'],
      },
      xAxis: {
        type: 'category',
        data: categories,
        axisLabel: {
          rotate: 0,
          interval: 0,
          fontSize: 11,
          color: '#8B949E',
          formatter: function (value: string) {
            if (value.length > 4) {
              return value.substring(0, 4) + '..'
            }
            return value
          },
        },
      },
      yAxis: {
        type: 'value',
        axisLabel: {
          fontSize: 11,
          color: '#8B949E',
        },
        splitLine: {
          lineStyle: { color: '#21262D', type: 'dashed', width: 1 },
        },
      },
      series: [
        {
          name: '数值',
          type: 'bar',
          data: seriesData,
          barMaxWidth: 50,
          itemStyle: {
            borderRadius: [8, 8, 0, 0],
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
