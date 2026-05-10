import { ref, computed } from 'vue';
import { ElMessage } from 'element-plus';
import BarChart from '@/components/charts/BarChart.vue';
import dashboardData from '@/mock/data/dashboard.json';
const loading = ref(false);
// 柱状图数据
const barChartData = computed(() => {
    return dashboardData.by_category.map(item => ({
        name: getCategoryLabel(item.category),
        value: item.count,
        category: item.category,
    }));
});
// 最近活动数据
const recentActivity = computed(() => {
    return dashboardData.recent_activity.slice(0, 5);
});
const rowHasStep = computed(() => {
    return recentActivity.value.some(item => item.step_name);
});
// 分类标签映射
function getCategoryLabel(category) {
    const map = {
        disaster_recovery: '灾备切换',
        degradation: '服务降级',
        release: '发布演练',
        security: '安全事件',
    };
    return map[category] || category;
}
// 活动类型标签
function getActivityTypeTag(type) {
    const map = {
        drill_start: 'primary',
        drill_complete: 'success',
        drill_terminate: 'danger',
        step_start: 'info',
        step_complete: 'success',
    };
    return map[type] || 'info';
}
function getActivityLabel(type) {
    const map = {
        drill_start: '演练开始',
        drill_complete: '演练完成',
        drill_terminate: '演练终止',
        step_start: '步骤开始',
        step_complete: '步骤完成',
    };
    return map[type] || type;
}
function formatTime(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
    });
}
// 加载数据
async function loadDashboardData() {
    loading.value = true;
    try {
        // mock 数据已直接导入
        ElMessage.success('数据加载成功');
    }
    catch (error) {
        ElMessage.error('加载数据失败');
        console.error('Failed to load dashboard data:', error);
    }
    finally {
        loading.value = false;
    }
}
loadDashboardData();
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['table-card']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-container" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-content" },
});
const __VLS_0 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    gutter: (20),
    ...{ class: "stats-row" },
}));
const __VLS_2 = __VLS_1({
    gutter: (20),
    ...{ class: "stats-row" },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
const __VLS_4 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
    span: (6),
}));
const __VLS_6 = __VLS_5({
    span: (6),
}, ...__VLS_functionalComponentArgsRest(__VLS_5));
__VLS_7.slots.default;
const __VLS_8 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    ...{ class: "stat-card" },
}));
const __VLS_10 = __VLS_9({
    ...{ class: "stat-card" },
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-icon total" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.svg, __VLS_intrinsicElements.svg)({
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    'stroke-width': "2",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
    d: "M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.polyline)({
    points: "14,2 14,8 20,8",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-content" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-value" },
});
(__VLS_ctx.dashboardData.stats.total_drills);
var __VLS_11;
var __VLS_7;
const __VLS_12 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    span: (6),
}));
const __VLS_14 = __VLS_13({
    span: (6),
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
__VLS_15.slots.default;
const __VLS_16 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    ...{ class: "stat-card" },
}));
const __VLS_18 = __VLS_17({
    ...{ class: "stat-card" },
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-icon active" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.svg, __VLS_intrinsicElements.svg)({
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    'stroke-width': "2",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.circle)({
    cx: "12",
    cy: "12",
    r: "10",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.polyline)({
    points: "12,6 12,12 16,14",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-content" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-value" },
});
(__VLS_ctx.dashboardData.stats.active_drills);
var __VLS_19;
var __VLS_15;
const __VLS_20 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    span: (6),
}));
const __VLS_22 = __VLS_21({
    span: (6),
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_23.slots.default;
const __VLS_24 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    ...{ class: "stat-card" },
}));
const __VLS_26 = __VLS_25({
    ...{ class: "stat-card" },
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
__VLS_27.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-icon success" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.svg, __VLS_intrinsicElements.svg)({
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    'stroke-width': "2",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
    d: "M22 11.08V12a10 10 0 1 1-5.93-9.14",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.polyline)({
    points: "22,4 12,14.01 9,11.01",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-content" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-value" },
});
(__VLS_ctx.dashboardData.stats.success_rate);
var __VLS_27;
var __VLS_23;
const __VLS_28 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    span: (6),
}));
const __VLS_30 = __VLS_29({
    span: (6),
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
__VLS_31.slots.default;
const __VLS_32 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    ...{ class: "stat-card" },
}));
const __VLS_34 = __VLS_33({
    ...{ class: "stat-card" },
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
__VLS_35.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-icon team" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.svg, __VLS_intrinsicElements.svg)({
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    'stroke-width': "2",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
    d: "M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.circle)({
    cx: "9",
    cy: "7",
    r: "4",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
    d: "M23 21v-2a4 4 0 0 0-3-3.87",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
    d: "M16 3.13a4 4 0 0 1 0 7.75",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-content" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-value" },
});
(__VLS_ctx.dashboardData.stats.team_online_count);
(__VLS_ctx.dashboardData.stats.team_total_count);
var __VLS_35;
var __VLS_31;
var __VLS_3;
const __VLS_36 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
    gutter: (20),
    ...{ class: "charts-row" },
}));
const __VLS_38 = __VLS_37({
    gutter: (20),
    ...{ class: "charts-row" },
}, ...__VLS_functionalComponentArgsRest(__VLS_37));
__VLS_39.slots.default;
const __VLS_40 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
    span: (12),
}));
const __VLS_42 = __VLS_41({
    span: (12),
}, ...__VLS_functionalComponentArgsRest(__VLS_41));
__VLS_43.slots.default;
const __VLS_44 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
    ...{ class: "chart-card" },
}));
const __VLS_46 = __VLS_45({
    ...{ class: "chart-card" },
}, ...__VLS_functionalComponentArgsRest(__VLS_45));
__VLS_47.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_47.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "card-title" },
    });
}
/** @type {[typeof BarChart, ]} */ ;
// @ts-ignore
const __VLS_48 = __VLS_asFunctionalComponent(BarChart, new BarChart({
    data: (__VLS_ctx.barChartData),
    height: "320px",
}));
const __VLS_49 = __VLS_48({
    data: (__VLS_ctx.barChartData),
    height: "320px",
}, ...__VLS_functionalComponentArgsRest(__VLS_48));
var __VLS_47;
var __VLS_43;
const __VLS_51 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_52 = __VLS_asFunctionalComponent(__VLS_51, new __VLS_51({
    span: (12),
}));
const __VLS_53 = __VLS_52({
    span: (12),
}, ...__VLS_functionalComponentArgsRest(__VLS_52));
__VLS_54.slots.default;
const __VLS_55 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_56 = __VLS_asFunctionalComponent(__VLS_55, new __VLS_55({
    ...{ class: "table-card" },
}));
const __VLS_57 = __VLS_56({
    ...{ class: "table-card" },
}, ...__VLS_functionalComponentArgsRest(__VLS_56));
__VLS_58.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_58.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "card-title" },
    });
}
const __VLS_59 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
    data: (__VLS_ctx.recentActivity),
    stripe: true,
    ...{ style: {} },
}));
const __VLS_61 = __VLS_60({
    data: (__VLS_ctx.recentActivity),
    stripe: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_60));
__VLS_62.slots.default;
const __VLS_63 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_64 = __VLS_asFunctionalComponent(__VLS_63, new __VLS_63({
    prop: "type",
    label: "类型",
    width: "120",
}));
const __VLS_65 = __VLS_64({
    prop: "type",
    label: "类型",
    width: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_64));
__VLS_66.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_66.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_67 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({
        type: (__VLS_ctx.getActivityTypeTag(row.type)),
        size: "small",
    }));
    const __VLS_69 = __VLS_68({
        type: (__VLS_ctx.getActivityTypeTag(row.type)),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_68));
    __VLS_70.slots.default;
    (__VLS_ctx.getActivityLabel(row.type));
    var __VLS_70;
}
var __VLS_66;
const __VLS_71 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({
    prop: "drill_name",
    label: "演练名称",
    minWidth: "180",
}));
const __VLS_73 = __VLS_72({
    prop: "drill_name",
    label: "演练名称",
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_72));
if (__VLS_ctx.rowHasStep) {
    const __VLS_75 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_76 = __VLS_asFunctionalComponent(__VLS_75, new __VLS_75({
        prop: "step_name",
        label: "步骤",
        minWidth: "150",
    }));
    const __VLS_77 = __VLS_76({
        prop: "step_name",
        label: "步骤",
        minWidth: "150",
    }, ...__VLS_functionalComponentArgsRest(__VLS_76));
}
const __VLS_79 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_80 = __VLS_asFunctionalComponent(__VLS_79, new __VLS_79({
    prop: "operator",
    label: "操作人",
    width: "100",
}));
const __VLS_81 = __VLS_80({
    prop: "operator",
    label: "操作人",
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_80));
const __VLS_83 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_84 = __VLS_asFunctionalComponent(__VLS_83, new __VLS_83({
    prop: "created_at",
    label: "时间",
    width: "160",
}));
const __VLS_85 = __VLS_84({
    prop: "created_at",
    label: "时间",
    width: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_84));
__VLS_86.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_86.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatTime(row.created_at));
}
var __VLS_86;
var __VLS_62;
var __VLS_58;
var __VLS_54;
var __VLS_39;
/** @type {__VLS_StyleScopedClasses['page-container']} */ ;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
/** @type {__VLS_StyleScopedClasses['page-content']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-row']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['total']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-content']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-value']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['active']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-content']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-value']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['success']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-content']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-value']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['team']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-content']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-value']} */ ;
/** @type {__VLS_StyleScopedClasses['charts-row']} */ ;
/** @type {__VLS_StyleScopedClasses['chart-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-title']} */ ;
/** @type {__VLS_StyleScopedClasses['table-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-title']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            BarChart: BarChart,
            dashboardData: dashboardData,
            barChartData: barChartData,
            recentActivity: recentActivity,
            rowHasStep: rowHasStep,
            getActivityTypeTag: getActivityTypeTag,
            getActivityLabel: getActivityLabel,
            formatTime: formatTime,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
