import { watch } from 'vue';
import { useECharts } from '@/composables/useECharts';
import { applyDarkTheme } from './theme';
const statusColorMap = {
    pending: '#7D7D7D',
    running: '#55C3D3',
    completed: '#2EA043',
    timeout: '#B8860B',
    skipped: '#646464',
    issue: '#DA3633',
};
const props = withDefaults(defineProps(), {
    height: '400px',
    loading: false,
});
const { domRef, setOptions, showLoading, hideLoading } = useECharts();
watch(() => props.data, (newData) => {
    if (!newData || newData.length === 0)
        return;
    const seriesData = [];
    const yCategories = [];
    let globalIndex = 0;
    newData.forEach((group) => {
        group.items.forEach((item) => {
            const startTime = item.startTime ? new Date(item.startTime).getTime() : 0;
            // For pending items, use start time + 1 minute as end
            const endTime = item.endTime ? new Date(item.endTime).getTime() : (startTime || Date.now()) + 60000;
            const categoryIndex = yCategories.indexOf(group.name);
            if (categoryIndex === -1) {
                yCategories.push(group.name);
            }
            seriesData.push({
                name: item.name || group.name,
                // [categoryIndex, startTime, endTime, order]
                value: [yCategories.indexOf(group.name), startTime, endTime, globalIndex],
                itemStyle: {
                    color: statusColorMap[item.status] || '#55C3D3',
                    borderRadius: 3,
                },
                groupId: globalIndex,
            });
            globalIndex++;
        });
    });
    if (seriesData.length === 0)
        return;
    const minTime = Math.min(...seriesData.map((item) => item.value[1]));
    const maxTime = Math.max(...seriesData.map((item) => item.value[2]));
    const timeRange = maxTime - minTime;
    const padding = Math.max(timeRange * 0.05, 60000);
    const option = applyDarkTheme({
        tooltip: {
            trigger: 'item',
            formatter: (params) => {
                const idx = params.dataIndex;
                const d = seriesData[idx];
                if (!d)
                    return '';
                const start = new Date(d.value[1]).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
                const end = new Date(d.value[2]).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
                return `${d.name}<br/>开始：${start}<br/>结束：${end}`;
            },
        },
        grid: {
            top: 10,
            bottom: 10,
            left: 140,
            right: 20,
        },
        xAxis: {
            type: 'time',
            min: minTime - padding,
            max: maxTime + padding,
            axisLabel: {
                fontSize: 11,
                color: '#8B949E',
                formatter: (value) => {
                    const date = new Date(value);
                    return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`;
                },
            },
            splitLine: {
                show: true,
                lineStyle: { color: '#21262D', type: 'dashed', width: 1 },
            },
            axisTick: { show: false },
            axisLine: { lineStyle: { color: '#30363D' } },
        },
        series: [
            {
                type: 'custom',
                renderItem: (_params, api) => {
                    const categoryIndex = api.value(0);
                    const startTime = api.value(1);
                    const endTime = api.value(2);
                    const xStart = api.coord([startTime, categoryIndex])[0];
                    const xEnd = api.coord([endTime, categoryIndex])[0];
                    const yPoint = api.coord([0, categoryIndex])[1];
                    // Get the height of each category row
                    const rowHeight = api.size([0, 1])[1] || 30;
                    const barHeight = rowHeight * 0.6;
                    const yOffset = yPoint - rowHeight / 2 + (rowHeight * 0.4) / 2;
                    return {
                        type: 'rect',
                        shape: {
                            x: xStart,
                            y: yOffset,
                            width: Math.max(xEnd - xStart, 4),
                            height: barHeight,
                            r: 3,
                        },
                        style: api.style(),
                    };
                },
                data: seriesData,
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
    height: '400px',
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
