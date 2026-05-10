import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { ArrowLeft, CircleCheck, Warning, CircleClose } from '@element-plus/icons-vue';
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue';
import ActionConfirm from '@/components/common/ActionConfirm.vue';
import tasksData from '@/mock/data/tasks.json';
import stepsData from '@/mock/data/steps.json';
const route = useRoute();
const router = useRouter();
const task = ref(null);
const stepInstance = ref(null);
const actionForm = ref({
    result: '',
    remark: '',
});
const taskId = computed(() => {
    const id = route.params.id;
    return typeof id === 'string' ? parseInt(id, 10) : 0;
});
function formatDeadline(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
    });
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
async function loadTaskData() {
    try {
        task.value = tasksData.find(t => t.id === taskId.value) || null;
        if (task.value) {
            stepInstance.value = stepsData.find(s => s.id === task.value.step_id) || null;
        }
        if (!task.value) {
            ElMessage.error('任务不存在');
            router.back();
        }
    }
    catch (error) {
        ElMessage.error('加载数据失败');
        console.error('Failed to load task data:', error);
    }
}
function handleComplete() {
    ElMessage.success('任务已完成');
    router.back();
}
function handleReportIssue() {
    if (!actionForm.value.remark.trim()) {
        ElMessage.warning('请填写异常说明');
        return;
    }
    ElMessage.success('异常已上报');
    router.back();
}
function handleSkip() {
    ElMessage.success('任务已跳过');
    router.back();
}
onMounted(() => {
    loadTaskData();
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['action-card']} */ ;
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
if (__VLS_ctx.task) {
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
    const __VLS_20 = {}.ElDescriptions;
    /** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
    // @ts-ignore
    const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
        column: (2),
        border: true,
    }));
    const __VLS_22 = __VLS_21({
        column: (2),
        border: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_21));
    __VLS_23.slots.default;
    const __VLS_24 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
        label: "演练名称",
    }));
    const __VLS_26 = __VLS_25({
        label: "演练名称",
    }, ...__VLS_functionalComponentArgsRest(__VLS_25));
    __VLS_27.slots.default;
    (__VLS_ctx.task.drill_name);
    var __VLS_27;
    const __VLS_28 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
        label: "步骤名称",
    }));
    const __VLS_30 = __VLS_29({
        label: "步骤名称",
    }, ...__VLS_functionalComponentArgsRest(__VLS_29));
    __VLS_31.slots.default;
    (__VLS_ctx.task.step_name);
    var __VLS_31;
    const __VLS_32 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
        label: "状态",
    }));
    const __VLS_34 = __VLS_33({
        label: "状态",
    }, ...__VLS_functionalComponentArgsRest(__VLS_33));
    __VLS_35.slots.default;
    /** @type {[typeof DrillStatusBadge, ]} */ ;
    // @ts-ignore
    const __VLS_36 = __VLS_asFunctionalComponent(DrillStatusBadge, new DrillStatusBadge({
        status: (__VLS_ctx.task.status),
        type: "step",
    }));
    const __VLS_37 = __VLS_36({
        status: (__VLS_ctx.task.status),
        type: "step",
    }, ...__VLS_functionalComponentArgsRest(__VLS_36));
    var __VLS_35;
    const __VLS_39 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_40 = __VLS_asFunctionalComponent(__VLS_39, new __VLS_39({
        label: "执行人",
    }));
    const __VLS_41 = __VLS_40({
        label: "执行人",
    }, ...__VLS_functionalComponentArgsRest(__VLS_40));
    __VLS_42.slots.default;
    (__VLS_ctx.task.assigned_to_name);
    var __VLS_42;
    if (__VLS_ctx.task.deadline) {
        const __VLS_43 = {}.ElDescriptionsItem;
        /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
        // @ts-ignore
        const __VLS_44 = __VLS_asFunctionalComponent(__VLS_43, new __VLS_43({
            label: "截止时间",
        }));
        const __VLS_45 = __VLS_44({
            label: "截止时间",
        }, ...__VLS_functionalComponentArgsRest(__VLS_44));
        __VLS_46.slots.default;
        (__VLS_ctx.formatDeadline(__VLS_ctx.task.deadline));
        var __VLS_46;
    }
    const __VLS_47 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_48 = __VLS_asFunctionalComponent(__VLS_47, new __VLS_47({
        label: "创建时间",
    }));
    const __VLS_49 = __VLS_48({
        label: "创建时间",
    }, ...__VLS_functionalComponentArgsRest(__VLS_48));
    __VLS_50.slots.default;
    (__VLS_ctx.formatTime(__VLS_ctx.task.created_at));
    var __VLS_50;
    var __VLS_23;
    var __VLS_19;
    const __VLS_51 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_52 = __VLS_asFunctionalComponent(__VLS_51, new __VLS_51({
        ...{ class: "detail-card" },
    }));
    const __VLS_53 = __VLS_52({
        ...{ class: "detail-card" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_52));
    __VLS_54.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_54.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "card-title" },
        });
    }
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "step-detail" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "detail-section" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.h4, __VLS_intrinsicElements.h4)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "description" },
    });
    (__VLS_ctx.task.step_description);
    if (__VLS_ctx.task.script) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "detail-section" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.h4, __VLS_intrinsicElements.h4)({});
        __VLS_asFunctionalElement(__VLS_intrinsicElements.pre, __VLS_intrinsicElements.pre)({
            ...{ class: "code-block" },
        });
        (__VLS_ctx.task.script);
    }
    if (__VLS_ctx.stepInstance?.error_message) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "detail-section error" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.h4, __VLS_intrinsicElements.h4)({});
        __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
            ...{ class: "error-message" },
        });
        (__VLS_ctx.stepInstance.error_message);
    }
    var __VLS_54;
    const __VLS_55 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_56 = __VLS_asFunctionalComponent(__VLS_55, new __VLS_55({
        ...{ class: "action-card" },
    }));
    const __VLS_57 = __VLS_56({
        ...{ class: "action-card" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_56));
    __VLS_58.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_58.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "card-title" },
        });
    }
    const __VLS_59 = {}.ElForm;
    /** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
    // @ts-ignore
    const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
        model: (__VLS_ctx.actionForm),
        labelWidth: "80px",
    }));
    const __VLS_61 = __VLS_60({
        model: (__VLS_ctx.actionForm),
        labelWidth: "80px",
    }, ...__VLS_functionalComponentArgsRest(__VLS_60));
    __VLS_62.slots.default;
    const __VLS_63 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_64 = __VLS_asFunctionalComponent(__VLS_63, new __VLS_63({
        label: "执行结果",
    }));
    const __VLS_65 = __VLS_64({
        label: "执行结果",
    }, ...__VLS_functionalComponentArgsRest(__VLS_64));
    __VLS_66.slots.default;
    const __VLS_67 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({
        modelValue: (__VLS_ctx.actionForm.result),
        type: "textarea",
        rows: (4),
        placeholder: "请输入执行结果（可选）",
    }));
    const __VLS_69 = __VLS_68({
        modelValue: (__VLS_ctx.actionForm.result),
        type: "textarea",
        rows: (4),
        placeholder: "请输入执行结果（可选）",
    }, ...__VLS_functionalComponentArgsRest(__VLS_68));
    var __VLS_66;
    const __VLS_71 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({
        label: "备注",
    }));
    const __VLS_73 = __VLS_72({
        label: "备注",
    }, ...__VLS_functionalComponentArgsRest(__VLS_72));
    __VLS_74.slots.default;
    const __VLS_75 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_76 = __VLS_asFunctionalComponent(__VLS_75, new __VLS_75({
        modelValue: (__VLS_ctx.actionForm.remark),
        type: "textarea",
        rows: (2),
        placeholder: "请输入备注说明（可选）",
    }));
    const __VLS_77 = __VLS_76({
        modelValue: (__VLS_ctx.actionForm.remark),
        type: "textarea",
        rows: (2),
        placeholder: "请输入备注说明（可选）",
    }, ...__VLS_functionalComponentArgsRest(__VLS_76));
    var __VLS_74;
    var __VLS_62;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "action-buttons" },
    });
    /** @type {[typeof ActionConfirm, typeof ActionConfirm, ]} */ ;
    // @ts-ignore
    const __VLS_79 = __VLS_asFunctionalComponent(ActionConfirm, new ActionConfirm({
        ...{ 'onConfirm': {} },
        title: "完成任务",
        message: "确定要标记此任务为完成状态吗？",
        type: "success",
    }));
    const __VLS_80 = __VLS_79({
        ...{ 'onConfirm': {} },
        title: "完成任务",
        message: "确定要标记此任务为完成状态吗？",
        type: "success",
    }, ...__VLS_functionalComponentArgsRest(__VLS_79));
    let __VLS_82;
    let __VLS_83;
    let __VLS_84;
    const __VLS_85 = {
        onConfirm: (__VLS_ctx.handleComplete)
    };
    __VLS_81.slots.default;
    const __VLS_86 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_87 = __VLS_asFunctionalComponent(__VLS_86, new __VLS_86({}));
    const __VLS_88 = __VLS_87({}, ...__VLS_functionalComponentArgsRest(__VLS_87));
    __VLS_89.slots.default;
    const __VLS_90 = {}.CircleCheck;
    /** @type {[typeof __VLS_components.CircleCheck, ]} */ ;
    // @ts-ignore
    const __VLS_91 = __VLS_asFunctionalComponent(__VLS_90, new __VLS_90({}));
    const __VLS_92 = __VLS_91({}, ...__VLS_functionalComponentArgsRest(__VLS_91));
    var __VLS_89;
    var __VLS_81;
    /** @type {[typeof ActionConfirm, typeof ActionConfirm, ]} */ ;
    // @ts-ignore
    const __VLS_94 = __VLS_asFunctionalComponent(ActionConfirm, new ActionConfirm({
        ...{ 'onConfirm': {} },
        title: "上报异常",
        message: "确定要上报此任务异常吗？",
        type: "danger",
    }));
    const __VLS_95 = __VLS_94({
        ...{ 'onConfirm': {} },
        title: "上报异常",
        message: "确定要上报此任务异常吗？",
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_94));
    let __VLS_97;
    let __VLS_98;
    let __VLS_99;
    const __VLS_100 = {
        onConfirm: (__VLS_ctx.handleReportIssue)
    };
    __VLS_96.slots.default;
    const __VLS_101 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_102 = __VLS_asFunctionalComponent(__VLS_101, new __VLS_101({}));
    const __VLS_103 = __VLS_102({}, ...__VLS_functionalComponentArgsRest(__VLS_102));
    __VLS_104.slots.default;
    const __VLS_105 = {}.Warning;
    /** @type {[typeof __VLS_components.Warning, ]} */ ;
    // @ts-ignore
    const __VLS_106 = __VLS_asFunctionalComponent(__VLS_105, new __VLS_105({}));
    const __VLS_107 = __VLS_106({}, ...__VLS_functionalComponentArgsRest(__VLS_106));
    var __VLS_104;
    var __VLS_96;
    /** @type {[typeof ActionConfirm, typeof ActionConfirm, ]} */ ;
    // @ts-ignore
    const __VLS_109 = __VLS_asFunctionalComponent(ActionConfirm, new ActionConfirm({
        ...{ 'onConfirm': {} },
        title: "跳过任务",
        message: "确定要跳过此任务吗？",
        type: "warning",
    }));
    const __VLS_110 = __VLS_109({
        ...{ 'onConfirm': {} },
        title: "跳过任务",
        message: "确定要跳过此任务吗？",
        type: "warning",
    }, ...__VLS_functionalComponentArgsRest(__VLS_109));
    let __VLS_112;
    let __VLS_113;
    let __VLS_114;
    const __VLS_115 = {
        onConfirm: (__VLS_ctx.handleSkip)
    };
    __VLS_111.slots.default;
    const __VLS_116 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_117 = __VLS_asFunctionalComponent(__VLS_116, new __VLS_116({}));
    const __VLS_118 = __VLS_117({}, ...__VLS_functionalComponentArgsRest(__VLS_117));
    __VLS_119.slots.default;
    const __VLS_120 = {}.CircleClose;
    /** @type {[typeof __VLS_components.CircleClose, ]} */ ;
    // @ts-ignore
    const __VLS_121 = __VLS_asFunctionalComponent(__VLS_120, new __VLS_120({}));
    const __VLS_122 = __VLS_121({}, ...__VLS_functionalComponentArgsRest(__VLS_121));
    var __VLS_119;
    var __VLS_111;
    var __VLS_58;
}
/** @type {__VLS_StyleScopedClasses['page-container']} */ ;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
/** @type {__VLS_StyleScopedClasses['page-title']} */ ;
/** @type {__VLS_StyleScopedClasses['page-content']} */ ;
/** @type {__VLS_StyleScopedClasses['info-card']} */ ;
/** @type {__VLS_StyleScopedClasses['detail-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-title']} */ ;
/** @type {__VLS_StyleScopedClasses['step-detail']} */ ;
/** @type {__VLS_StyleScopedClasses['detail-section']} */ ;
/** @type {__VLS_StyleScopedClasses['description']} */ ;
/** @type {__VLS_StyleScopedClasses['detail-section']} */ ;
/** @type {__VLS_StyleScopedClasses['code-block']} */ ;
/** @type {__VLS_StyleScopedClasses['detail-section']} */ ;
/** @type {__VLS_StyleScopedClasses['error']} */ ;
/** @type {__VLS_StyleScopedClasses['error-message']} */ ;
/** @type {__VLS_StyleScopedClasses['action-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-title']} */ ;
/** @type {__VLS_StyleScopedClasses['action-buttons']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            ArrowLeft: ArrowLeft,
            CircleCheck: CircleCheck,
            Warning: Warning,
            CircleClose: CircleClose,
            DrillStatusBadge: DrillStatusBadge,
            ActionConfirm: ActionConfirm,
            router: router,
            task: task,
            stepInstance: stepInstance,
            actionForm: actionForm,
            formatDeadline: formatDeadline,
            formatTime: formatTime,
            handleComplete: handleComplete,
            handleReportIssue: handleReportIssue,
            handleSkip: handleSkip,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
