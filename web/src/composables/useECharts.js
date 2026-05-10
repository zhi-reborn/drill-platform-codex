import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue';
import * as echarts from 'echarts';
import { useElementSize } from '@vueuse/core';
export function useECharts(initialOptions = {}) {
    const domRef = ref();
    let chart = null;
    const { width, height } = useElementSize(domRef);
    function initChart() {
        if (!domRef.value)
            return;
        chart = echarts.init(domRef.value, undefined, { renderer: 'canvas' });
        if (Object.keys(initialOptions).length > 0) {
            chart.setOption(initialOptions);
        }
    }
    function setOptions(options, notMerge = true) {
        chart?.setOption(options, { notMerge });
    }
    function showLoading() {
        chart?.showLoading('default', {
            text: '加载中...',
            color: '#55C3D3',
            textColor: '#E0E6ED',
            maskColor: 'rgba(13, 17, 23, 0.8)',
        });
    }
    function hideLoading() {
        chart?.hideLoading();
    }
    function resize() {
        chart?.resize();
    }
    function dispose() {
        chart?.dispose();
        chart = null;
    }
    onMounted(() => {
        nextTick(() => initChart());
    });
    onBeforeUnmount(() => {
        dispose();
    });
    watch([width, height], () => {
        resize();
    });
    return {
        domRef,
        chart,
        setOptions,
        showLoading,
        hideLoading,
        resize,
        dispose,
    };
}
