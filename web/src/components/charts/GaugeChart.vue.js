import { watch } from 'vue';
import { useECharts } from '@/composables/useECharts';
import { applyDarkTheme } from './theme';
const props = withDefaults(defineProps(), {
    height: '250px',
    loading: false,
});
const { domRef, setOptions, showLoading, hideLoading } = useECharts();
watch(() => props.data, (newData) => {
    const max = newData.max ?? 100;
    const min = newData.min ?? 0;
    const value = newData.value;
    const option = applyDarkTheme({
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
                        color: function (params) {
                            const val = params.value;
                            const percent = val / max;
                            if (percent >= 0.8) {
                                return '#2EA043';
                            }
                            else if (percent >= 0.5) {
                                return '#B8860B';
                            }
                            else {
                                return '#DA3633';
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
    height: '250px',
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
