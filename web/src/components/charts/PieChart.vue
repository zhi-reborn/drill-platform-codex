<template>
  <div ref="domRef" class="chart-container" :style="{ height: height }"></div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import type { EChartsOption } from 'echarts'
import { useECharts } from '@/composables/useECharts'
import { applyDarkTheme } from './theme'

interface PieChartData {
  name: string
  value: number
}

const props = withDefaults(
  defineProps<{
    data: PieChartData[]
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
    const option: EChartsOption = applyDarkTheme({
      tooltip: {
        trigger: 'item',
        formatter: '{a} <br/>{b}: {c} ({d}%)',
      },
      legend: {
        orient: 'horizontal',
        bottom: 0,
        left: 'center',
        textStyle: { color: '#8B949E', fontSize: 11 },
        itemWidth: 10,
        itemHeight: 10,
        itemGap: 12,
      },
      series: [
        {
          name: '占比',
          type: 'pie',
          radius: ['40%', '70%'],
          avoidLabelOverlap: false,
          itemStyle: {
            borderRadius: 10,
            borderColor: '#161B22',
            borderWidth: 2,
          },
          label: {
            show: false,
            position: 'center',
          },
          emphasis: {
            label: {
              show: true,
              fontSize: 20,
              fontWeight: 'bold',
            },
          },
          labelLine: {
            show: false,
          },
          data: newData,
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
