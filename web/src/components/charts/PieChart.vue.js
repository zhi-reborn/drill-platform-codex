import { watch } from 'vue';
import { useECharts } from '@/composables/useECharts';
import { applyDarkTheme } from './theme';
const props = withDefaults(defineProps(), {
    height: '300px',
    loading: false,
});
const { domRef, setOptions, showLoading, hideLoading } = useECharts();
watch(() => props.data, (newData) => {
    const option = applyDarkTheme({
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
