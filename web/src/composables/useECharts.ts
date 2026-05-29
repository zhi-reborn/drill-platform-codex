import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
// Full import kept — 7 chart types used (line/bar/pie/gauge/radar/heatmap/custom)
// plus many components make tree-shaking ROI marginal. Already isolated in manualChunk.
import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'
import { useElementSize } from '@vueuse/core'

export function useECharts(initialOptions: EChartsOption = {}) {
  const domRef = ref<HTMLElement>()
  let chart: echarts.ECharts | null = null

  const { width, height } = useElementSize(domRef)

  function initChart() {
    if (!domRef.value) return
    chart = echarts.init(domRef.value, undefined, { renderer: 'canvas' })
    if (Object.keys(initialOptions).length > 0) {
      chart.setOption(initialOptions)
    }
  }

  function setOptions(options: EChartsOption, notMerge = true) {
    chart?.setOption(options, { notMerge })
  }

  function showLoading() {
    chart?.showLoading('default', {
      text: '加载中...',
      color: '#55C3D3',
      textColor: '#E0E6ED',
      maskColor: 'rgba(13, 17, 23, 0.8)',
    })
  }

  function hideLoading() {
    chart?.hideLoading()
  }

  function resize() {
    chart?.resize()
  }

  function dispose() {
    chart?.dispose()
    chart = null
  }

  onMounted(() => {
    nextTick(() => initChart())
  })

  onBeforeUnmount(() => {
    dispose()
  })

  watch([width, height], () => {
    resize()
  })

  return {
    domRef,
    chart,
    setOptions,
    showLoading,
    hideLoading,
    resize,
    dispose,
  }
}
