import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { VideoPause, VideoPlay, VideoCamera } from '@element-plus/icons-vue';
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue';
import ActionConfirm from '@/components/common/ActionConfirm.vue';
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
// 模拟日志数据（实际应从 API 获取）
const drillLogs = computed(() => {
    const logs = [];
    drillSteps.value.forEach((step, index) => {
        if (step.started_at) {
            logs.push({
                id: `log-${step.id}-start`,
                action: 'step_start',
                step_name: step.step_name,
                operator: step.assignee_name,
                created_at: step.started_at,
                remark: `开始执行步骤`,
            });
        }
        if (step.completed_at) {
            logs.push({
                id: `log-${step.id}-complete`,
                action: 'step_complete',
                step_name: step.step_name,
                operator: step.assignee_name,
                created_at: step.completed_at,
                remark: step.result_json ? `执行结果：${step.result_json}` : '步骤执行完成',
            });
        }
    });
    return logs.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
});
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
function getLogTypeTag(action) {
    const map = {
        step_start: 'primary',
        step_complete: 'success',
        step_issue: 'danger',
        step_skip: 'info',
        force_complete: 'warning',
    };
    return map[action] || 'info';
}
function getLogActionLabel(action) {
    const map = {
        step_start: '步骤开始',
        step_complete: '步骤完成',
        step_issue: '步骤异常',
        step_skip: '步骤跳过',
        force_complete: '强制完成',
    };
    return map[action] || action;
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
function handlePause() {
    ElMessage.success('演练已暂停');
}
function handleResume() {
    ElMessage.success('演练已继续');
}
function handleTerminate() {
    ElMessage.success('演练已终止');
    router.push('/director/dashboard');
}
onMounted(() => {
    loadDrillData();
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['logs-card']} */ ;
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
(__VLS_ctx.instance?.name || '演练监控');
if (__VLS_ctx.instance) {
    /** @type {[typeof DrillStatusBadge, ]} */ ;
    // @ts-ignore
    const __VLS_0 = __VLS_asFunctionalComponent(DrillStatusBadge, new DrillStatusBadge({
        status: (__VLS_ctx.instance.status),
        type: "drill",
    }));
    const __VLS_1 = __VLS_0({
        status: (__VLS_ctx.instance.status),
        type: "drill",
    }, ...__VLS_functionalComponentArgsRest(__VLS_0));
}
if (__VLS_ctx.instance) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "page-content" },
    });
    const __VLS_3 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_4 = __VLS_asFunctionalComponent(__VLS_3, new __VLS_3({
        ...{ class: "info-card" },
    }));
    const __VLS_5 = __VLS_4({
        ...{ class: "info-card" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_4));
    __VLS_6.slots.default;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "progress-section" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "progress-label" },
    });
    (__VLS_ctx.instance.completed_steps);
    (__VLS_ctx.instance.total_steps);
    const __VLS_7 = {}.ElProgress;
    /** @type {[typeof __VLS_components.ElProgress, typeof __VLS_components.elProgress, ]} */ ;
    // @ts-ignore
    const __VLS_8 = __VLS_asFunctionalComponent(__VLS_7, new __VLS_7({
        percentage: (Math.round(__VLS_ctx.instance.completed_steps / __VLS_ctx.instance.total_steps * 100)),
        strokeWidth: (10),
        status: (__VLS_ctx.instance.status === 'completed' ? 'success' : undefined),
    }));
    const __VLS_9 = __VLS_8({
        percentage: (Math.round(__VLS_ctx.instance.completed_steps / __VLS_ctx.instance.total_steps * 100)),
        strokeWidth: (10),
        status: (__VLS_ctx.instance.status === 'completed' ? 'success' : undefined),
    }, ...__VLS_functionalComponentArgsRest(__VLS_8));
    var __VLS_6;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "control-section" },
    });
    /** @type {[typeof ActionConfirm, typeof ActionConfirm, ]} */ ;
    // @ts-ignore
    const __VLS_11 = __VLS_asFunctionalComponent(ActionConfirm, new ActionConfirm({
        ...{ 'onConfirm': {} },
        title: "暂停演练",
        message: "确定要暂停当前演练吗？",
        type: "warning",
    }));
    const __VLS_12 = __VLS_11({
        ...{ 'onConfirm': {} },
        title: "暂停演练",
        message: "确定要暂停当前演练吗？",
        type: "warning",
    }, ...__VLS_functionalComponentArgsRest(__VLS_11));
    let __VLS_14;
    let __VLS_15;
    let __VLS_16;
    const __VLS_17 = {
        onConfirm: (__VLS_ctx.handlePause)
    };
    __VLS_13.slots.default;
    const __VLS_18 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_19 = __VLS_asFunctionalComponent(__VLS_18, new __VLS_18({}));
    const __VLS_20 = __VLS_19({}, ...__VLS_functionalComponentArgsRest(__VLS_19));
    __VLS_21.slots.default;
    const __VLS_22 = {}.VideoPause;
    /** @type {[typeof __VLS_components.VideoPause, ]} */ ;
    // @ts-ignore
    const __VLS_23 = __VLS_asFunctionalComponent(__VLS_22, new __VLS_22({}));
    const __VLS_24 = __VLS_23({}, ...__VLS_functionalComponentArgsRest(__VLS_23));
    var __VLS_21;
    var __VLS_13;
    /** @type {[typeof ActionConfirm, typeof ActionConfirm, ]} */ ;
    // @ts-ignore
    const __VLS_26 = __VLS_asFunctionalComponent(ActionConfirm, new ActionConfirm({
        ...{ 'onConfirm': {} },
        title: "继续演练",
        message: "确定要继续执行演练吗？",
        type: "primary",
    }));
    const __VLS_27 = __VLS_26({
        ...{ 'onConfirm': {} },
        title: "继续演练",
        message: "确定要继续执行演练吗？",
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_26));
    let __VLS_29;
    let __VLS_30;
    let __VLS_31;
    const __VLS_32 = {
        onConfirm: (__VLS_ctx.handleResume)
    };
    __VLS_28.slots.default;
    const __VLS_33 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_34 = __VLS_asFunctionalComponent(__VLS_33, new __VLS_33({}));
    const __VLS_35 = __VLS_34({}, ...__VLS_functionalComponentArgsRest(__VLS_34));
    __VLS_36.slots.default;
    const __VLS_37 = {}.VideoPlay;
    /** @type {[typeof __VLS_components.VideoPlay, ]} */ ;
    // @ts-ignore
    const __VLS_38 = __VLS_asFunctionalComponent(__VLS_37, new __VLS_37({}));
    const __VLS_39 = __VLS_38({}, ...__VLS_functionalComponentArgsRest(__VLS_38));
    var __VLS_36;
    var __VLS_28;
    /** @type {[typeof ActionConfirm, typeof ActionConfirm, ]} */ ;
    // @ts-ignore
    const __VLS_41 = __VLS_asFunctionalComponent(ActionConfirm, new ActionConfirm({
        ...{ 'onConfirm': {} },
        title: "终止演练",
        message: "确定要终止当前演练吗？此操作不可恢复！",
        danger: true,
    }));
    const __VLS_42 = __VLS_41({
        ...{ 'onConfirm': {} },
        title: "终止演练",
        message: "确定要终止当前演练吗？此操作不可恢复！",
        danger: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_41));
    let __VLS_44;
    let __VLS_45;
    let __VLS_46;
    const __VLS_47 = {
        onConfirm: (__VLS_ctx.handleTerminate)
    };
    __VLS_43.slots.default;
    const __VLS_48 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({}));
    const __VLS_50 = __VLS_49({}, ...__VLS_functionalComponentArgsRest(__VLS_49));
    __VLS_51.slots.default;
    const __VLS_52 = {}.VideoCamera;
    /** @type {[typeof __VLS_components.VideoCamera, ]} */ ;
    // @ts-ignore
    const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({}));
    const __VLS_54 = __VLS_53({}, ...__VLS_functionalComponentArgsRest(__VLS_53));
    var __VLS_51;
    var __VLS_43;
    const __VLS_56 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
        ...{ class: "steps-card" },
    }));
    const __VLS_58 = __VLS_57({
        ...{ class: "steps-card" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_57));
    __VLS_59.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_59.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "card-title" },
        });
    }
    const __VLS_60 = {}.ElTable;
    /** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
    // @ts-ignore
    const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
        data: (__VLS_ctx.drillSteps),
        ...{ style: {} },
    }));
    const __VLS_62 = __VLS_61({
        data: (__VLS_ctx.drillSteps),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_61));
    __VLS_63.slots.default;
    const __VLS_64 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
        prop: "order_index",
        label: "序号",
        width: "80",
    }));
    const __VLS_66 = __VLS_65({
        prop: "order_index",
        label: "序号",
        width: "80",
    }, ...__VLS_functionalComponentArgsRest(__VLS_65));
    const __VLS_68 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({
        prop: "step_name",
        label: "步骤名",
        minWidth: "200",
    }));
    const __VLS_70 = __VLS_69({
        prop: "step_name",
        label: "步骤名",
        minWidth: "200",
    }, ...__VLS_functionalComponentArgsRest(__VLS_69));
    const __VLS_72 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
        prop: "status",
        label: "状态",
        width: "120",
    }));
    const __VLS_74 = __VLS_73({
        prop: "status",
        label: "状态",
        width: "120",
    }, ...__VLS_functionalComponentArgsRest(__VLS_73));
    __VLS_75.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_75.slots;
        const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
        /** @type {[typeof DrillStatusBadge, ]} */ ;
        // @ts-ignore
        const __VLS_76 = __VLS_asFunctionalComponent(DrillStatusBadge, new DrillStatusBadge({
            status: (row.status),
            type: "step",
        }));
        const __VLS_77 = __VLS_76({
            status: (row.status),
            type: "step",
        }, ...__VLS_functionalComponentArgsRest(__VLS_76));
    }
    var __VLS_75;
    const __VLS_79 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_80 = __VLS_asFunctionalComponent(__VLS_79, new __VLS_79({
        prop: "assignee_name",
        label: "执行人",
        width: "120",
    }));
    const __VLS_81 = __VLS_80({
        prop: "assignee_name",
        label: "执行人",
        width: "120",
    }, ...__VLS_functionalComponentArgsRest(__VLS_80));
    const __VLS_83 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_84 = __VLS_asFunctionalComponent(__VLS_83, new __VLS_83({
        label: "耗时",
        width: "120",
    }));
    const __VLS_85 = __VLS_84({
        label: "耗时",
        width: "120",
    }, ...__VLS_functionalComponentArgsRest(__VLS_84));
    __VLS_86.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_86.slots;
        const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
        (__VLS_ctx.calculateDuration(row));
    }
    var __VLS_86;
    var __VLS_63;
    var __VLS_59;
    const __VLS_87 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_88 = __VLS_asFunctionalComponent(__VLS_87, new __VLS_87({
        ...{ class: "logs-card" },
    }));
    const __VLS_89 = __VLS_88({
        ...{ class: "logs-card" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_88));
    __VLS_90.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_90.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "card-title" },
        });
    }
    const __VLS_91 = {}.ElTimeline;
    /** @type {[typeof __VLS_components.ElTimeline, typeof __VLS_components.elTimeline, typeof __VLS_components.ElTimeline, typeof __VLS_components.elTimeline, ]} */ ;
    // @ts-ignore
    const __VLS_92 = __VLS_asFunctionalComponent(__VLS_91, new __VLS_91({}));
    const __VLS_93 = __VLS_92({}, ...__VLS_functionalComponentArgsRest(__VLS_92));
    __VLS_94.slots.default;
    for (const [log] of __VLS_getVForSourceType((__VLS_ctx.drillLogs))) {
        const __VLS_95 = {}.ElTimelineItem;
        /** @type {[typeof __VLS_components.ElTimelineItem, typeof __VLS_components.elTimelineItem, typeof __VLS_components.ElTimelineItem, typeof __VLS_components.elTimelineItem, ]} */ ;
        // @ts-ignore
        const __VLS_96 = __VLS_asFunctionalComponent(__VLS_95, new __VLS_95({
            key: (log.id),
            timestamp: (__VLS_ctx.formatTime(log.created_at)),
            placement: "top",
        }));
        const __VLS_97 = __VLS_96({
            key: (log.id),
            timestamp: (__VLS_ctx.formatTime(log.created_at)),
            placement: "top",
        }, ...__VLS_functionalComponentArgsRest(__VLS_96));
        __VLS_98.slots.default;
        const __VLS_99 = {}.ElCard;
        /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
        // @ts-ignore
        const __VLS_100 = __VLS_asFunctionalComponent(__VLS_99, new __VLS_99({}));
        const __VLS_101 = __VLS_100({}, ...__VLS_functionalComponentArgsRest(__VLS_100));
        __VLS_102.slots.default;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "log-content" },
        });
        const __VLS_103 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_104 = __VLS_asFunctionalComponent(__VLS_103, new __VLS_103({
            type: (__VLS_ctx.getLogTypeTag(log.action)),
            size: "small",
        }));
        const __VLS_105 = __VLS_104({
            type: (__VLS_ctx.getLogTypeTag(log.action)),
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_104));
        __VLS_106.slots.default;
        (__VLS_ctx.getLogActionLabel(log.action));
        var __VLS_106;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "log-step" },
        });
        (log.step_name);
        if (log.operator) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "log-operator" },
            });
            (log.operator);
        }
        if (log.remark) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
                ...{ class: "log-remark" },
            });
            (log.remark);
        }
        var __VLS_102;
        var __VLS_98;
    }
    var __VLS_94;
    if (__VLS_ctx.drillLogs.length === 0) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "empty-tip" },
        });
    }
    var __VLS_90;
}
/** @type {__VLS_StyleScopedClasses['page-container']} */ ;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
/** @type {__VLS_StyleScopedClasses['page-title']} */ ;
/** @type {__VLS_StyleScopedClasses['page-content']} */ ;
/** @type {__VLS_StyleScopedClasses['info-card']} */ ;
/** @type {__VLS_StyleScopedClasses['progress-section']} */ ;
/** @type {__VLS_StyleScopedClasses['progress-label']} */ ;
/** @type {__VLS_StyleScopedClasses['control-section']} */ ;
/** @type {__VLS_StyleScopedClasses['steps-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-title']} */ ;
/** @type {__VLS_StyleScopedClasses['logs-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-title']} */ ;
/** @type {__VLS_StyleScopedClasses['log-content']} */ ;
/** @type {__VLS_StyleScopedClasses['log-step']} */ ;
/** @type {__VLS_StyleScopedClasses['log-operator']} */ ;
/** @type {__VLS_StyleScopedClasses['log-remark']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-tip']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            VideoPause: VideoPause,
            VideoPlay: VideoPlay,
            VideoCamera: VideoCamera,
            DrillStatusBadge: DrillStatusBadge,
            ActionConfirm: ActionConfirm,
            instance: instance,
            drillSteps: drillSteps,
            drillLogs: drillLogs,
            calculateDuration: calculateDuration,
            formatTime: formatTime,
            getLogTypeTag: getLogTypeTag,
            getLogActionLabel: getLogActionLabel,
            handlePause: handlePause,
            handleResume: handleResume,
            handleTerminate: handleTerminate,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
