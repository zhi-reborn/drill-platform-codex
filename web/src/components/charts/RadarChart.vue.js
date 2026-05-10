import { watch } from 'vue';
import { useECharts } from '@/composables/useECharts';
import { applyDarkTheme } from './theme';
const props = withDefaults(defineProps(), {
    height: '300px',
    loading: false,
});
const { domRef, setOptions, showLoading, hideLoading } = useECharts();
watch(() => props.data, (newData) => {
    const indicator = newData.map((item) => ({
        name: item.name,
        max: item.max ?? 100,
    }));
    const values = newData.map((item) => item.value);
    const option = applyDarkTheme({
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
    });
    setOptions(option);
}, { immediate: true });
watch(() => props.loading, (newLoading) => {
    if (newLoading) {
        showLoading();
    }
    else {
        hideLoading();
    }
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_withDefaultsArg = (function (t) { return t; })({
    height: '300px',
    loading: false,
});
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ref: "domRef",
    ...{ class: "chart-container" },
    ...{ style: ({ height: __VLS_ctx.height }) },
});
/** @type {typeof __VLS_ctx.domRef} */ ;
/** @type {__VLS_StyleScopedClasses['chart-container']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            domRef: domRef,
        };
    },
    __typeProps: {},
    props: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeProps: {},
    props: {},
});
; /* PartiallyEnd: #4569/main.vue */
