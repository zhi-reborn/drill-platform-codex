import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { ArrowLeft } from '@element-plus/icons-vue';
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue';
import TimelineChart from '@/components/charts/TimelineChart.vue';
import instancesData from '@/mock/data/instances.json';
import stepsData from '@/mock/data/steps.json';
const route = useRoute();
const router = useRouter();
const instance = ref(null);
const steps = ref([]);
const drillId = computed(() => {
    const id = route.params.id;
    return typeof id === 'string' ? parseInt(id, 10) : 0;
});
const drillSteps = computed(() => {
    return steps.value.filter(s => s.drill_id === drillId.value).sort((a, b) => a.order_index - b.order_index);
});
const timelineData = computed(() => {
    return drillSteps.value.map(step => ({
        name: step.step_name,
        items: [{
                startTime: step.started_at || new Date().toISOString(),
                endTime: step.completed_at,
                status: step.status,
            }],
    }));
});
function getStepTypeLabel(type) {
    const map = {
        serial: '串行',
        parallel: '并行',
        any_of: '任选',
        condition: '条件',
    };
    return map[type] || type;
}
function getStepTypeTag(type) {
    const map = {
        serial: 'primary',
        parallel: 'success',
        any_of: 'warning',
        condition: 'info',
    };
    return map[type] || 'info';
}
function calculateDuration(step) {
    if (!step.started_at || !step.completed_at) {
        return '-';
    }
    const start = new Date(step.started_at).getTime();
    const end = new Date(step.completed_at).getTime();
    const diff = Math.floor((end - start) / 1000);
    if (diff < 60) {
        return `${diff}s`;
    }
    const mins = Math.floor(diff / 60);
    const secs = diff % 60;
    return `${mins}m ${secs}s`;
}
function formatTime(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
    });
}
async function loadDrillData() {
    try {
        instance.value = instancesData.find(i => i.id === drillId.value) || null;
        steps.value = stepsData;
        if (!instance.value) {
            ElMessage.error('演练不存在');
            router.back();
        }
    }
    catch (error) {
        ElMessage.error('加载数据失败');
        console.error('Failed to load drill data:', error);
    }
}
onMounted(() => {
    loadDrillData();
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
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
}));
const __VLS_2 = __VLS_1({
    ...{ 'onClick': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onClick: (...[$event]) => {
        __VLS_ctx.router.back();
    }
};
__VLS_3.slots.default;
const __VLS_8 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({}));
const __VLS_10 = __VLS_9({}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
const __VLS_12 = {}.ArrowLeft;
/** @type {[typeof __VLS_components.ArrowLeft, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({}));
const __VLS_14 = __VLS_13({}, ...__VLS_functionalComponentArgsRest(__VLS_13));
var __VLS_11;
var __VLS_3;
if (__VLS_ctx.instance) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "page-content" },
    });
    const __VLS_16 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
        ...{ class: "info-card" },
    }));
    const __VLS_18 = __VLS_17({
        ...{ class: "info-card" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_17));
    __VLS_19.slots.default;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "info-header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "info-title" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({});
    (__VLS_ctx.instance.name);
    /** @type {[typeof DrillStatusBadge, ]} */ ;
    // @ts-ignore
    const __VLS_20 = __VLS_asFunctionalComponent(DrillStatusBadge, new DrillStatusBadge({
        status: (__VLS_ctx.instance.status),
        type: "drill",
    }));
    const __VLS_21 = __VLS_20({
        status: (__VLS_ctx.instance.status),
        type: "drill",
    }, ...__VLS_functionalComponentArgsRest(__VLS_20));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "info-progress" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "progress-label" },
    });
    (__VLS_ctx.instance.completed_steps);
    (__VLS_ctx.instance.total_steps);
    const __VLS_23 = {}.ElProgress;
    /** @type {[typeof __VLS_components.ElProgress, typeof __VLS_components.elProgress, ]} */ ;
    // @ts-ignore
    const __VLS_24 = __VLS_asFunctionalComponent(__VLS_23, new __VLS_23({
        percentage: (Math.round(__VLS_ctx.instance.completed_steps / __VLS_ctx.instance.total_steps * 100)),
        strokeWidth: (8),
        status: (__VLS_ctx.instance.status === 'completed' ? 'success' : undefined),
    }));
    const __VLS_25 = __VLS_24({
        percentage: (Math.round(__VLS_ctx.instance.completed_steps / __VLS_ctx.instance.total_steps * 100)),
        strokeWidth: (8),
        status: (__VLS_ctx.instance.status === 'completed' ? 'success' : undefined),
    }, ...__VLS_functionalComponentArgsRest(__VLS_24));
    const __VLS_27 = {}.ElDescriptions;
    /** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
    // @ts-ignore
    const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
        column: (2),
        border: true,
        ...{ class: "info-descriptions" },
    }));
    const __VLS_29 = __VLS_28({
        column: (2),
        border: true,
        ...{ class: "info-descriptions" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_28));
    __VLS_30.slots.default;
    const __VLS_31 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({
        label: "演练模板",
    }));
    const __VLS_33 = __VLS_32({
        label: "演练模板",
    }, ...__VLS_functionalComponentArgsRest(__VLS_32));
    __VLS_34.slots.default;
    (__VLS_ctx.instance.template_name);
    var __VLS_34;
    const __VLS_35 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
        label: "创建人",
    }));
    const __VLS_37 = __VLS_36({
        label: "创建人",
    }, ...__VLS_functionalComponentArgsRest(__VLS_36));
    __VLS_38.slots.default;
    (__VLS_ctx.instance.created_by_name);
    var __VLS_38;
    const __VLS_39 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_40 = __VLS_asFunctionalComponent(__VLS_39, new __VLS_39({
        label: "开始时间",
    }));
    const __VLS_41 = __VLS_40({
        label: "开始时间",
    }, ...__VLS_functionalComponentArgsRest(__VLS_40));
    __VLS_42.slots.default;
    (__VLS_ctx.instance.started_at ? __VLS_ctx.formatTime(__VLS_ctx.instance.started_at) : '-');
    var __VLS_42;
    const __VLS_43 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_44 = __VLS_asFunctionalComponent(__VLS_43, new __VLS_43({
        label: "创建时间",
    }));
    const __VLS_45 = __VLS_44({
        label: "创建时间",
    }, ...__VLS_functionalComponentArgsRest(__VLS_44));
    __VLS_46.slots.default;
    (__VLS_ctx.formatTime(__VLS_ctx.instance.created_at));
    var __VLS_46;
    if (__VLS_ctx.instance.completed_at) {
        const __VLS_47 = {}.ElDescriptionsItem;
        /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
        // @ts-ignore
        const __VLS_48 = __VLS_asFunctionalComponent(__VLS_47, new __VLS_47({
            label: "完成时间",
            span: (2),
        }));
        const __VLS_49 = __VLS_48({
            label: "完成时间",
            span: (2),
        }, ...__VLS_functionalComponentArgsRest(__VLS_48));
        __VLS_50.slots.default;
        (__VLS_ctx.formatTime(__VLS_ctx.instance.completed_at));
        var __VLS_50;
    }
    var __VLS_30;
    var __VLS_19;
    const __VLS_51 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_52 = __VLS_asFunctionalComponent(__VLS_51, new __VLS_51({
        ...{ class: "steps-card" },
    }));
    const __VLS_53 = __VLS_52({
        ...{ class: "steps-card" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_52));
    __VLS_54.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_54.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "card-title" },
        });
    }
    const __VLS_55 = {}.ElTable;
    /** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
    // @ts-ignore
    const __VLS_56 = __VLS_asFunctionalComponent(__VLS_55, new __VLS_55({
        data: (__VLS_ctx.drillSteps),
        ...{ style: {} },
    }));
    const __VLS_57 = __VLS_56({
        data: (__VLS_ctx.drillSteps),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_56));
    __VLS_58.slots.default;
    const __VLS_59 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
        prop: "order_index",
        label: "序号",
        width: "80",
    }));
    const __VLS_61 = __VLS_60({
        prop: "order_index",
        label: "序号",
        width: "80",
    }, ...__VLS_functionalComponentArgsRest(__VLS_60));
    const __VLS_63 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_64 = __VLS_asFunctionalComponent(__VLS_63, new __VLS_63({
        prop: "step_name",
        label: "步骤名",
        minWidth: "200",
    }));
    const __VLS_65 = __VLS_64({
        prop: "step_name",
        label: "步骤名",
        minWidth: "200",
    }, ...__VLS_functionalComponentArgsRest(__VLS_64));
    const __VLS_67 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({
        prop: "step_type",
        label: "类型",
        width: "100",
    }));
    const __VLS_69 = __VLS_68({
        prop: "step_type",
        label: "类型",
        width: "100",
    }, ...__VLS_functionalComponentArgsRest(__VLS_68));
    __VLS_70.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_70.slots;
        const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
        const __VLS_71 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({
            type: (__VLS_ctx.getStepTypeTag(row.step_type)),
            size: "small",
        }));
        const __VLS_73 = __VLS_72({
            type: (__VLS_ctx.getStepTypeTag(row.step_type)),
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_72));
        __VLS_74.slots.default;
        (__VLS_ctx.getStepTypeLabel(row.step_type));
        var __VLS_74;
    }
    var __VLS_70;
    const __VLS_75 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_76 = __VLS_asFunctionalComponent(__VLS_75, new __VLS_75({
        prop: "status",
        label: "状态",
        width: "120",
    }));
    const __VLS_77 = __VLS_76({
        prop: "status",
        label: "状态",
        width: "120",
    }, ...__VLS_functionalComponentArgsRest(__VLS_76));
    __VLS_78.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_78.slots;
        const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
        /** @type {[typeof DrillStatusBadge, ]} */ ;
        // @ts-ignore
        const __VLS_79 = __VLS_asFunctionalComponent(DrillStatusBadge, new DrillStatusBadge({
            status: (row.status),
            type: "step",
        }));
        const __VLS_80 = __VLS_79({
            status: (row.status),
            type: "step",
        }, ...__VLS_functionalComponentArgsRest(__VLS_79));
    }
    var __VLS_78;
    const __VLS_82 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_83 = __VLS_asFunctionalComponent(__VLS_82, new __VLS_82({
        prop: "assignee_name",
        label: "执行人",
        width: "120",
    }));
    const __VLS_84 = __VLS_83({
        prop: "assignee_name",
        label: "执行人",
        width: "120",
    }, ...__VLS_functionalComponentArgsRest(__VLS_83));
    const __VLS_86 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_87 = __VLS_asFunctionalComponent(__VLS_86, new __VLS_86({
        label: "耗时",
        width: "120",
    }));
    const __VLS_88 = __VLS_87({
        label: "耗时",
        width: "120",
    }, ...__VLS_functionalComponentArgsRest(__VLS_87));
    __VLS_89.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_89.slots;
        const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
        (__VLS_ctx.calculateDuration(row));
    }
    var __VLS_89;
    var __VLS_58;
    var __VLS_54;
    const __VLS_90 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_91 = __VLS_asFunctionalComponent(__VLS_90, new __VLS_90({
        ...{ class: "timeline-card" },
    }));
    const __VLS_92 = __VLS_91({
        ...{ class: "timeline-card" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_91));
    __VLS_93.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_93.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "card-title" },
        });
    }
    /** @type {[typeof TimelineChart, ]} */ ;
    // @ts-ignore
    const __VLS_94 = __VLS_asFunctionalComponent(TimelineChart, new TimelineChart({
        data: (__VLS_ctx.timelineData),
        height: "200px",
    }));
    const __VLS_95 = __VLS_94({
        data: (__VLS_ctx.timelineData),
        height: "200px",
    }, ...__VLS_functionalComponentArgsRest(__VLS_94));
    var __VLS_93;
}
/** @type {__VLS_StyleScopedClasses['page-container']} */ ;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
/** @type {__VLS_StyleScopedClasses['page-title']} */ ;
/** @type {__VLS_StyleScopedClasses['page-content']} */ ;
/** @type {__VLS_StyleScopedClasses['info-card']} */ ;
/** @type {__VLS_StyleScopedClasses['info-header']} */ ;
/** @type {__VLS_StyleScopedClasses['info-title']} */ ;
/** @type {__VLS_StyleScopedClasses['info-progress']} */ ;
/** @type {__VLS_StyleScopedClasses['progress-label']} */ ;
/** @type {__VLS_StyleScopedClasses['info-descriptions']} */ ;
/** @type {__VLS_StyleScopedClasses['steps-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-title']} */ ;
/** @type {__VLS_StyleScopedClasses['timeline-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-title']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            ArrowLeft: ArrowLeft,
            DrillStatusBadge: DrillStatusBadge,
            TimelineChart: TimelineChart,
            router: router,
            instance: instance,
            drillSteps: drillSteps,
            timelineData: timelineData,
            getStepTypeLabel: getStepTypeLabel,
            getStepTypeTag: getStepTypeTag,
            calculateDuration: calculateDuration,
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
