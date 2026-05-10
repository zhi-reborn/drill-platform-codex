import { watch } from 'vue';
import { useECharts } from '@/composables/useECharts';
import { applyDarkTheme } from './theme';
const props = withDefaults(defineProps(), {
    height: '300px',
    loading: false,
});
const { domRef, setOptions, showLoading, hideLoading } = useECharts();
watch(() => props.data, (newData) => {
    const categories = newData.map((item) => item.category || item.name);
    const seriesData = newData.map((item) => ({
        name: item.name,
        value: item.value,
    }));
    const option = applyDarkTheme({
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
                formatter: function (value) {
                    if (value.length > 4) {
                        return value.substring(0, 4) + '..';
                    }
                    return value;
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
