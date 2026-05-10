import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { Monitor } from '@element-plus/icons-vue';
import BarChart from '@/components/charts/BarChart.vue';
import LineChart from '@/components/charts/LineChart.vue';
import dashboardData from '@/mock/data/dashboard.json';
const router = useRouter();
const barChartData = computed(() => {
    return dashboardData.by_category.map(item => ({
        name: getCategoryLabel(item.category),
        value: item.count,
        category: item.category,
    }));
});
const lineChartData = computed(() => {
    return dashboardData.hourly_errors.map(item => ({
        name: item.hour,
        value: item.count,
    }));
});
const recentActivity = computed(() => {
    return dashboardData.recent_activity.slice(0, 5);
});
function getCategoryLabel(category) {
    const map = {
        disaster_recovery: '灾备切换',
        degradation: '服务降级',
        release: '发布演练',
        security: '安全事件',
    };
    return map[category] || category;
}
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
function formatDuration(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}m ${secs}s`;
}
function viewScreen() {
    router.push('/screen/100');
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['card-title']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-container" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({
    ...{ class: "page-title" },
});
const __VLS_0 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onClick': {} },
    type: "primary",
}));
const __VLS_2 = __VLS_1({
    ...{ 'onClick': {} },
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onClick: (__VLS_ctx.viewScreen)
};
__VLS_3.slots.default;
const __VLS_8 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({}));
const __VLS_10 = __VLS_9({}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
const __VLS_12 = {}.Monitor;
/** @type {[typeof __VLS_components.Monitor, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({}));
const __VLS_14 = __VLS_13({}, ...__VLS_functionalComponentArgsRest(__VLS_13));
var __VLS_11;
var __VLS_3;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-content" },
});
const __VLS_16 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    gutter: (20),
    ...{ class: "stats-row" },
}));
const __VLS_18 = __VLS_17({
    gutter: (20),
    ...{ class: "stats-row" },
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
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
    ...{ class: "stat-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-value" },
});
(__VLS_ctx.dashboardData.stats.total_drills);
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
    ...{ class: "stat-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-value" },
});
(__VLS_ctx.dashboardData.stats.active_drills);
var __VLS_35;
var __VLS_31;
const __VLS_36 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
    span: (6),
}));
const __VLS_38 = __VLS_37({
    span: (6),
}, ...__VLS_functionalComponentArgsRest(__VLS_37));
__VLS_39.slots.default;
const __VLS_40 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
    ...{ class: "stat-card" },
}));
const __VLS_42 = __VLS_41({
    ...{ class: "stat-card" },
}, ...__VLS_functionalComponentArgsRest(__VLS_41));
__VLS_43.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-value success" },
});
(__VLS_ctx.dashboardData.stats.success_rate);
var __VLS_43;
var __VLS_39;
const __VLS_44 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
    span: (6),
}));
const __VLS_46 = __VLS_45({
    span: (6),
}, ...__VLS_functionalComponentArgsRest(__VLS_45));
__VLS_47.slots.default;
const __VLS_48 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
    ...{ class: "stat-card" },
}));
const __VLS_50 = __VLS_49({
    ...{ class: "stat-card" },
}, ...__VLS_functionalComponentArgsRest(__VLS_49));
__VLS_51.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat-value" },
});
(__VLS_ctx.formatDuration(__VLS_ctx.dashboardData.stats.avg_step_duration_seconds));
var __VLS_51;
var __VLS_47;
var __VLS_19;
const __VLS_52 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({
    gutter: (20),
    ...{ class: "charts-row" },
}));
const __VLS_54 = __VLS_53({
    gutter: (20),
    ...{ class: "charts-row" },
}, ...__VLS_functionalComponentArgsRest(__VLS_53));
__VLS_55.slots.default;
const __VLS_56 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
    span: (12),
}));
const __VLS_58 = __VLS_57({
    span: (12),
}, ...__VLS_functionalComponentArgsRest(__VLS_57));
__VLS_59.slots.default;
const __VLS_60 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
    ...{ class: "chart-card" },
}));
const __VLS_62 = __VLS_61({
    ...{ class: "chart-card" },
}, ...__VLS_functionalComponentArgsRest(__VLS_61));
__VLS_63.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_63.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "card-title" },
    });
}
/** @type {[typeof BarChart, ]} */ ;
// @ts-ignore
const __VLS_64 = __VLS_asFunctionalComponent(BarChart, new BarChart({
    data: (__VLS_ctx.barChartData),
    height: "320px",
}));
const __VLS_65 = __VLS_64({
    data: (__VLS_ctx.barChartData),
    height: "320px",
}, ...__VLS_functionalComponentArgsRest(__VLS_64));
var __VLS_63;
var __VLS_59;
const __VLS_67 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({
    span: (12),
}));
const __VLS_69 = __VLS_68({
    span: (12),
}, ...__VLS_functionalComponentArgsRest(__VLS_68));
__VLS_70.slots.default;
const __VLS_71 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({
    ...{ class: "chart-card" },
}));
const __VLS_73 = __VLS_72({
    ...{ class: "chart-card" },
}, ...__VLS_functionalComponentArgsRest(__VLS_72));
__VLS_74.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_74.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "card-title" },
    });
}
/** @type {[typeof LineChart, ]} */ ;
// @ts-ignore
const __VLS_75 = __VLS_asFunctionalComponent(LineChart, new LineChart({
    data: (__VLS_ctx.lineChartData),
    height: "320px",
}));
const __VLS_76 = __VLS_75({
    data: (__VLS_ctx.lineChartData),
    height: "320px",
}, ...__VLS_functionalComponentArgsRest(__VLS_75));
var __VLS_74;
var __VLS_70;
var __VLS_55;
const __VLS_78 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_79 = __VLS_asFunctionalComponent(__VLS_78, new __VLS_78({
    ...{ class: "table-card" },
}));
const __VLS_80 = __VLS_79({
    ...{ class: "table-card" },
}, ...__VLS_functionalComponentArgsRest(__VLS_79));
__VLS_81.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_81.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "card-title" },
    });
}
const __VLS_82 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_83 = __VLS_asFunctionalComponent(__VLS_82, new __VLS_82({
    data: (__VLS_ctx.recentActivity),
    stripe: true,
    ...{ style: {} },
}));
const __VLS_84 = __VLS_83({
    data: (__VLS_ctx.recentActivity),
    stripe: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_83));
__VLS_85.slots.default;
const __VLS_86 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_87 = __VLS_asFunctionalComponent(__VLS_86, new __VLS_86({
    prop: "type",
    label: "类型",
    width: "120",
}));
const __VLS_88 = __VLS_87({
    prop: "type",
    label: "类型",
    width: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_87));
__VLS_89.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_89.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_90 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_91 = __VLS_asFunctionalComponent(__VLS_90, new __VLS_90({
        type: (__VLS_ctx.getActivityTypeTag(row.type)),
        size: "small",
    }));
    const __VLS_92 = __VLS_91({
        type: (__VLS_ctx.getActivityTypeTag(row.type)),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_91));
    __VLS_93.slots.default;
    (__VLS_ctx.getActivityLabel(row.type));
    var __VLS_93;
}
var __VLS_89;
const __VLS_94 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_95 = __VLS_asFunctionalComponent(__VLS_94, new __VLS_94({
    prop: "drill_name",
    label: "演练名称",
    minWidth: "180",
}));
const __VLS_96 = __VLS_95({
    prop: "drill_name",
    label: "演练名称",
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_95));
const __VLS_98 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_99 = __VLS_asFunctionalComponent(__VLS_98, new __VLS_98({
    prop: "operator",
    label: "操作人",
    width: "100",
}));
const __VLS_100 = __VLS_99({
    prop: "operator",
    label: "操作人",
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_99));
const __VLS_102 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_103 = __VLS_asFunctionalComponent(__VLS_102, new __VLS_102({
    prop: "created_at",
    label: "时间",
    width: "160",
}));
const __VLS_104 = __VLS_103({
    prop: "created_at",
    label: "时间",
    width: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_103));
__VLS_105.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_105.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatTime(row.created_at));
}
var __VLS_105;
var __VLS_85;
var __VLS_81;
/** @type {__VLS_StyleScopedClasses['page-container']} */ ;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
/** @type {__VLS_StyleScopedClasses['page-title']} */ ;
/** @type {__VLS_StyleScopedClasses['page-content']} */ ;
/** @type {__VLS_StyleScopedClasses['stats-row']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-value']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-value']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-value']} */ ;
/** @type {__VLS_StyleScopedClasses['success']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-card']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-label']} */ ;
/** @type {__VLS_StyleScopedClasses['stat-value']} */ ;
/** @type {__VLS_StyleScopedClasses['charts-row']} */ ;
/** @type {__VLS_StyleScopedClasses['chart-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-title']} */ ;
/** @type {__VLS_StyleScopedClasses['chart-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-title']} */ ;
/** @type {__VLS_StyleScopedClasses['table-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-title']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Monitor: Monitor,
            BarChart: BarChart,
            LineChart: LineChart,
            dashboardData: dashboardData,
            barChartData: barChartData,
            lineChartData: lineChartData,
            recentActivity: recentActivity,
            getActivityTypeTag: getActivityTypeTag,
            getActivityLabel: getActivityLabel,
            formatTime: formatTime,
            formatDuration: formatDuration,
            viewScreen: viewScreen,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
