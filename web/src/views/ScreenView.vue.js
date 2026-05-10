import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
import ScreenLayout from '@/components/screen/ScreenLayout.vue';
import MetricsBar from '@/components/screen/MetricsBar.vue';
import StatusPanel from '@/components/screen/StatusPanel.vue';
import AlertFeed from '@/components/screen/AlertFeed.vue';
import GaugeChart from '@/components/charts/GaugeChart.vue';
import PieChart from '@/components/charts/PieChart.vue';
import BarChart from '@/components/charts/BarChart.vue';
import TimelineChart from '@/components/charts/TimelineChart.vue';
// Mock data imports
import dashboardData from '@/mock/data/dashboard.json';
import instancesData from '@/mock/data/instances.json';
import stepsData from '@/mock/data/steps.json';
import notificationsData from '@/mock/data/notifications.json';
const refreshTimer = ref();
const stats = ref(dashboardData.stats);
// === KPI ===
const kpiMetrics = computed(() => [
    { label: '总演练', value: stats.value.total_drills, icon: 'Monitor' },
    { label: '活跃', value: stats.value.active_drills, icon: 'VideoPlay' },
    { label: '成功率', value: `${stats.value.success_rate}%`, icon: 'CircleCheck' },
    { label: '失败率', value: `${stats.value.failure_rate}%`, icon: 'CircleClose' },
    { label: '平均耗时', value: `${Math.round(stats.value.avg_step_duration_seconds / 60)}min`, icon: 'Timer' },
    { label: '团队在线', value: `${stats.value.team_online_count}/${stats.value.team_total_count}`, icon: 'User' },
]);
// === 活跃演练 ===
const activeDrills = computed(() => {
    return instancesData.filter(d => d.status === 'running' || d.status === 'paused');
});
// === 步骤状态分布 ===
const stepStatusData = computed(() => {
    const steps = stepsData;
    const counts = {};
    steps.forEach(s => {
        const status = s.status;
        counts[status] = (counts[status] || 0) + 1;
    });
    return Object.entries(counts).map(([name, value]) => ({ name, value }));
});
// === 各分类演练数 ===
const categoryData = computed(() => {
    const labels = {
        disaster_recovery: '灾备切换',
        degradation: '服务降级',
        release: '发布演练',
        security: '安全事件',
    };
    return dashboardData.by_category.map((c) => ({
        name: labels[c.category] || c.category,
        value: c.count,
    }));
});
// === 演练时间线 ===
const timelineData = computed(() => {
    const grouped = {};
    const steps = stepsData;
    steps.forEach(s => {
        const drillId = s.drill_id;
        if (!grouped[drillId])
            grouped[drillId] = [];
        grouped[drillId].push(s);
    });
    return Object.entries(grouped).map(([drillId, items]) => ({
        name: `演练 #${drillId}`,
        items: items.map(s => ({
            name: s.step_name,
            startTime: s.started_at || '',
            endTime: s.completed_at || '',
            status: s.status || 'pending',
        })),
    }));
});
// === 告警流 (丰富数据) ===
const alerts = computed(() => {
    const items = [];
    // 1. 从 steps 中提取 issue 和 timeout
    const steps = stepsData;
    steps.forEach((s, i) => {
        if (s.status === 'issue') {
            items.push({
                id: i,
                level: 'error',
                message: `步骤「${s.step_name}」异常: ${s.error_message || '未知错误'}`,
                icon: 'Warning',
                created_at: s.started_at || '2024-12-20T10:04:00Z',
            });
        }
        if (s.status === 'timeout') {
            items.push({
                id: i + 1000,
                level: 'warning',
                message: `步骤「${s.step_name}」超时`,
                icon: 'Clock',
                created_at: s.started_at || '2024-12-20T10:04:00Z',
            });
        }
    });
    // 2. 从通知中补充
    const notifs = notificationsData;
    notifs.forEach((n, i) => {
        const type = n.type;
        if (type === 'drill_started') {
            items.push({
                id: i + 2000,
                level: 'info',
                message: `演练「${n.title}」已启动`,
                icon: 'VideoPlay',
                created_at: n.created_at,
            });
        }
        if (type === 'drill_completed') {
            items.push({
                id: i + 3000,
                level: 'info',
                message: `演练「${n.title}」全部通过`,
                icon: 'CircleCheck',
                created_at: n.created_at,
            });
        }
        if (type === 'drill_paused') {
            items.push({
                id: i + 4000,
                level: 'warning',
                message: `演练「${n.title}」已暂停`,
                icon: 'VideoPause',
                created_at: n.created_at,
            });
        }
        if (type === 'drill_terminated') {
            items.push({
                id: i + 5000,
                level: 'error',
                message: `演练「${n.title}」已终止`,
                icon: 'CircleClose',
                created_at: n.created_at,
            });
        }
    });
    items.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
    return items.slice(0, 30);
});
// 自动刷新
function refreshData() {
    stats.value = dashboardData.stats;
}
onMounted(() => {
    refreshTimer.value = window.setInterval(refreshData, 5000);
});
onBeforeUnmount(() => {
    if (refreshTimer.value)
        clearInterval(refreshTimer.value);
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
/** @type {[typeof ScreenLayout, typeof ScreenLayout, ]} */ ;
// @ts-ignore
const __VLS_0 = __VLS_asFunctionalComponent(ScreenLayout, new ScreenLayout({}));
const __VLS_1 = __VLS_0({}, ...__VLS_functionalComponentArgsRest(__VLS_0));
var __VLS_3 = {};
__VLS_2.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "screen-content" },
});
/** @type {[typeof MetricsBar, ]} */ ;
// @ts-ignore
const __VLS_4 = __VLS_asFunctionalComponent(MetricsBar, new MetricsBar({
    metrics: (__VLS_ctx.kpiMetrics),
}));
const __VLS_5 = __VLS_4({
    metrics: (__VLS_ctx.kpiMetrics),
}, ...__VLS_functionalComponentArgsRest(__VLS_4));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "screen-body" },
});
/** @type {[typeof StatusPanel, ]} */ ;
// @ts-ignore
const __VLS_7 = __VLS_asFunctionalComponent(StatusPanel, new StatusPanel({
    drills: (__VLS_ctx.activeDrills),
}));
const __VLS_8 = __VLS_7({
    drills: (__VLS_ctx.activeDrills),
}, ...__VLS_functionalComponentArgsRest(__VLS_7));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "charts-area" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "chart-row chart-row--top" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "chart-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ class: "chart-title" },
});
/** @type {[typeof GaugeChart, ]} */ ;
// @ts-ignore
const __VLS_10 = __VLS_asFunctionalComponent(GaugeChart, new GaugeChart({
    data: ({ name: '成功率', value: __VLS_ctx.stats.success_rate, max: 100 }),
    height: "160px",
}));
const __VLS_11 = __VLS_10({
    data: ({ name: '成功率', value: __VLS_ctx.stats.success_rate, max: 100 }),
    height: "160px",
}, ...__VLS_functionalComponentArgsRest(__VLS_10));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "chart-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ class: "chart-title" },
});
/** @type {[typeof PieChart, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(PieChart, new PieChart({
    data: (__VLS_ctx.stepStatusData),
    height: "160px",
}));
const __VLS_14 = __VLS_13({
    data: (__VLS_ctx.stepStatusData),
    height: "160px",
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "chart-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ class: "chart-title" },
});
/** @type {[typeof BarChart, ]} */ ;
// @ts-ignore
const __VLS_16 = __VLS_asFunctionalComponent(BarChart, new BarChart({
    data: (__VLS_ctx.categoryData),
    height: "160px",
}));
const __VLS_17 = __VLS_16({
    data: (__VLS_ctx.categoryData),
    height: "160px",
}, ...__VLS_functionalComponentArgsRest(__VLS_16));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "chart-row chart-row--bottom" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "chart-card chart-card--wide" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ class: "chart-title" },
});
/** @type {[typeof TimelineChart, ]} */ ;
// @ts-ignore
const __VLS_19 = __VLS_asFunctionalComponent(TimelineChart, new TimelineChart({
    data: (__VLS_ctx.timelineData),
    height: "280px",
}));
const __VLS_20 = __VLS_19({
    data: (__VLS_ctx.timelineData),
    height: "280px",
}, ...__VLS_functionalComponentArgsRest(__VLS_19));
/** @type {[typeof AlertFeed, ]} */ ;
// @ts-ignore
const __VLS_22 = __VLS_asFunctionalComponent(AlertFeed, new AlertFeed({
    alerts: (__VLS_ctx.alerts),
}));
const __VLS_23 = __VLS_22({
    alerts: (__VLS_ctx.alerts),
}, ...__VLS_functionalComponentArgsRest(__VLS_22));
var __VLS_2;
/** @type {__VLS_StyleScopedClasses['screen-content']} */ ;
/** @type {__VLS_StyleScopedClasses['screen-body']} */ ;
/** @type {__VLS_StyleScopedClasses['charts-area']} */ ;
/** @type {__VLS_StyleScopedClasses['chart-row']} */ ;
/** @type {__VLS_StyleScopedClasses['chart-row--top']} */ ;
/** @type {__VLS_StyleScopedClasses['chart-card']} */ ;
/** @type {__VLS_StyleScopedClasses['chart-title']} */ ;
/** @type {__VLS_StyleScopedClasses['chart-card']} */ ;
/** @type {__VLS_StyleScopedClasses['chart-title']} */ ;
/** @type {__VLS_StyleScopedClasses['chart-card']} */ ;
/** @type {__VLS_StyleScopedClasses['chart-title']} */ ;
/** @type {__VLS_StyleScopedClasses['chart-row']} */ ;
/** @type {__VLS_StyleScopedClasses['chart-row--bottom']} */ ;
/** @type {__VLS_StyleScopedClasses['chart-card']} */ ;
/** @type {__VLS_StyleScopedClasses['chart-card--wide']} */ ;
/** @type {__VLS_StyleScopedClasses['chart-title']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            ScreenLayout: ScreenLayout,
            MetricsBar: MetricsBar,
            StatusPanel: StatusPanel,
            AlertFeed: AlertFeed,
            GaugeChart: GaugeChart,
            PieChart: PieChart,
            BarChart: BarChart,
            TimelineChart: TimelineChart,
            stats: stats,
            kpiMetrics: kpiMetrics,
            activeDrills: activeDrills,
            stepStatusData: stepStatusData,
            categoryData: categoryData,
            timelineData: timelineData,
            alerts: alerts,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
