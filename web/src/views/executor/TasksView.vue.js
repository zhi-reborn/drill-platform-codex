import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { Clock, User } from '@element-plus/icons-vue';
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue';
import EmptyBox from '@/components/common/EmptyBox.vue';
import tasksData from '@/mock/data/tasks.json';
const router = useRouter();
const loading = ref(false);
const tasks = ref([]);
const filterStatus = ref('');
const filteredTasks = computed(() => {
    if (!filterStatus.value) {
        return tasks.value;
    }
    return tasks.value.filter((task) => task.status === filterStatus.value);
});
const getStatusClass = (status) => {
    const classMap = {
        in_progress: 'status-in-progress',
        issued: 'status-issued',
        completed: 'status-completed',
    };
    return classMap[status] || '';
};
// 处理筛选变化
function handleFilterChange() {
    // filter is reactive, filteredTasks computed handles it
}
// 格式化截止时间
function formatDeadline(d) {
    if (!d)
        return '无';
    const date = new Date(d);
    const now = new Date();
    const diff = date.getTime() - now.getTime();
    const hours = Math.floor(diff / (1000 * 60 * 60));
    if (hours < 0) {
        return '已过期';
    }
    if (hours < 1) {
        return '1 小时内';
    }
    return `${hours}小时`;
}
// 跳转到任务详情
function goToTaskDetail(taskId) {
    router.push(`/executor/tasks/${taskId}`);
}
// 加载数据
async function loadTasks() {
    loading.value = true;
    try {
        // 使用 mock 数据
        tasks.value = tasksData;
    }
    catch (error) {
        ElMessage.error('加载任务失败');
        console.error('Failed to load tasks:', error);
    }
    finally {
        loading.value = false;
    }
}
onMounted(() => {
    loadTasks();
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['el-radio-button__inner']} */ ;
/** @type {__VLS_StyleScopedClasses['task-deadline']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "executor-tasks" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({
    ...{ class: "page-title" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "filter-group" },
});
const __VLS_0 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.filterStatus),
    size: "default",
}));
const __VLS_2 = __VLS_1({
    ...{ 'onChange': {} },
    modelValue: (__VLS_ctx.filterStatus),
    size: "default",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onChange: (__VLS_ctx.handleFilterChange)
};
__VLS_3.slots.default;
const __VLS_8 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    value: "",
}));
const __VLS_10 = __VLS_9({
    value: "",
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
var __VLS_11;
const __VLS_12 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    value: "pending",
}));
const __VLS_14 = __VLS_13({
    value: "pending",
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
__VLS_15.slots.default;
var __VLS_15;
const __VLS_16 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    value: "assigned",
}));
const __VLS_18 = __VLS_17({
    value: "assigned",
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
var __VLS_19;
const __VLS_20 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    value: "in_progress",
}));
const __VLS_22 = __VLS_21({
    value: "in_progress",
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_23.slots.default;
var __VLS_23;
const __VLS_24 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    value: "completed",
}));
const __VLS_26 = __VLS_25({
    value: "completed",
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
__VLS_27.slots.default;
var __VLS_27;
const __VLS_28 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    value: "issued",
}));
const __VLS_30 = __VLS_29({
    value: "issued",
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
__VLS_31.slots.default;
var __VLS_31;
var __VLS_3;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "tasks-container" },
});
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
if (!__VLS_ctx.loading && __VLS_ctx.filteredTasks.length === 0) {
    /** @type {[typeof EmptyBox, ]} */ ;
    // @ts-ignore
    const __VLS_32 = __VLS_asFunctionalComponent(EmptyBox, new EmptyBox({
        title: "暂无任务",
        description: "当前没有分配给您的任务",
    }));
    const __VLS_33 = __VLS_32({
        title: "暂无任务",
        description: "当前没有分配给您的任务",
    }, ...__VLS_functionalComponentArgsRest(__VLS_32));
}
else {
    const __VLS_35 = {}.ElRow;
    /** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
    // @ts-ignore
    const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
        gutter: (20),
        ...{ class: "tasks-grid" },
    }));
    const __VLS_37 = __VLS_36({
        gutter: (20),
        ...{ class: "tasks-grid" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_36));
    __VLS_38.slots.default;
    for (const [task] of __VLS_getVForSourceType((__VLS_ctx.filteredTasks))) {
        const __VLS_39 = {}.ElCol;
        /** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
        // @ts-ignore
        const __VLS_40 = __VLS_asFunctionalComponent(__VLS_39, new __VLS_39({
            key: (task.id),
            xs: (24),
            sm: (12),
            lg: (8),
            xl: (6),
        }));
        const __VLS_41 = __VLS_40({
            key: (task.id),
            xs: (24),
            sm: (12),
            lg: (8),
            xl: (6),
        }, ...__VLS_functionalComponentArgsRest(__VLS_40));
        __VLS_42.slots.default;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "task-card" },
            ...{ class: (__VLS_ctx.getStatusClass(task.status)) },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "task-card-header" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "task-drill-name" },
        });
        (task.drill_name);
        /** @type {[typeof DrillStatusBadge, ]} */ ;
        // @ts-ignore
        const __VLS_43 = __VLS_asFunctionalComponent(DrillStatusBadge, new DrillStatusBadge({
            status: (task.status),
            type: "step",
        }));
        const __VLS_44 = __VLS_43({
            status: (task.status),
            type: "step",
        }, ...__VLS_functionalComponentArgsRest(__VLS_43));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "task-card-body" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "task-step-name" },
        });
        (task.step_name);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "task-description" },
        });
        (task.step_description);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "task-meta" },
        });
        if (task.deadline) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                ...{ class: "task-deadline" },
            });
            const __VLS_46 = {}.ElIcon;
            /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
            // @ts-ignore
            const __VLS_47 = __VLS_asFunctionalComponent(__VLS_46, new __VLS_46({}));
            const __VLS_48 = __VLS_47({}, ...__VLS_functionalComponentArgsRest(__VLS_47));
            __VLS_49.slots.default;
            const __VLS_50 = {}.Clock;
            /** @type {[typeof __VLS_components.Clock, ]} */ ;
            // @ts-ignore
            const __VLS_51 = __VLS_asFunctionalComponent(__VLS_50, new __VLS_50({}));
            const __VLS_52 = __VLS_51({}, ...__VLS_functionalComponentArgsRest(__VLS_51));
            var __VLS_49;
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
            (__VLS_ctx.formatDeadline(task.deadline));
        }
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "task-assignee" },
        });
        const __VLS_54 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_55 = __VLS_asFunctionalComponent(__VLS_54, new __VLS_54({}));
        const __VLS_56 = __VLS_55({}, ...__VLS_functionalComponentArgsRest(__VLS_55));
        __VLS_57.slots.default;
        const __VLS_58 = {}.User;
        /** @type {[typeof __VLS_components.User, ]} */ ;
        // @ts-ignore
        const __VLS_59 = __VLS_asFunctionalComponent(__VLS_58, new __VLS_58({}));
        const __VLS_60 = __VLS_59({}, ...__VLS_functionalComponentArgsRest(__VLS_59));
        var __VLS_57;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
        (task.assigned_to_name);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "task-card-footer" },
        });
        if (task.status === 'pending' || task.status === 'assigned') {
            const __VLS_62 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_63 = __VLS_asFunctionalComponent(__VLS_62, new __VLS_62({
                ...{ 'onClick': {} },
                type: "primary",
                ...{ class: "action-btn" },
            }));
            const __VLS_64 = __VLS_63({
                ...{ 'onClick': {} },
                type: "primary",
                ...{ class: "action-btn" },
            }, ...__VLS_functionalComponentArgsRest(__VLS_63));
            let __VLS_66;
            let __VLS_67;
            let __VLS_68;
            const __VLS_69 = {
                onClick: (...[$event]) => {
                    if (!!(!__VLS_ctx.loading && __VLS_ctx.filteredTasks.length === 0))
                        return;
                    if (!(task.status === 'pending' || task.status === 'assigned'))
                        return;
                    __VLS_ctx.goToTaskDetail(task.id);
                }
            };
            __VLS_65.slots.default;
            var __VLS_65;
        }
        else if (task.status === 'in_progress') {
            const __VLS_70 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_71 = __VLS_asFunctionalComponent(__VLS_70, new __VLS_70({
                ...{ 'onClick': {} },
                type: "primary",
                ...{ class: "action-btn" },
            }));
            const __VLS_72 = __VLS_71({
                ...{ 'onClick': {} },
                type: "primary",
                ...{ class: "action-btn" },
            }, ...__VLS_functionalComponentArgsRest(__VLS_71));
            let __VLS_74;
            let __VLS_75;
            let __VLS_76;
            const __VLS_77 = {
                onClick: (...[$event]) => {
                    if (!!(!__VLS_ctx.loading && __VLS_ctx.filteredTasks.length === 0))
                        return;
                    if (!!(task.status === 'pending' || task.status === 'assigned'))
                        return;
                    if (!(task.status === 'in_progress'))
                        return;
                    __VLS_ctx.goToTaskDetail(task.id);
                }
            };
            __VLS_73.slots.default;
            var __VLS_73;
        }
        else if (task.status === 'issued') {
            const __VLS_78 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_79 = __VLS_asFunctionalComponent(__VLS_78, new __VLS_78({
                ...{ 'onClick': {} },
                type: "warning",
                ...{ class: "action-btn" },
            }));
            const __VLS_80 = __VLS_79({
                ...{ 'onClick': {} },
                type: "warning",
                ...{ class: "action-btn" },
            }, ...__VLS_functionalComponentArgsRest(__VLS_79));
            let __VLS_82;
            let __VLS_83;
            let __VLS_84;
            const __VLS_85 = {
                onClick: (...[$event]) => {
                    if (!!(!__VLS_ctx.loading && __VLS_ctx.filteredTasks.length === 0))
                        return;
                    if (!!(task.status === 'pending' || task.status === 'assigned'))
                        return;
                    if (!!(task.status === 'in_progress'))
                        return;
                    if (!(task.status === 'issued'))
                        return;
                    __VLS_ctx.goToTaskDetail(task.id);
                }
            };
            __VLS_81.slots.default;
            var __VLS_81;
        }
        else {
            const __VLS_86 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_87 = __VLS_asFunctionalComponent(__VLS_86, new __VLS_86({
                disabled: true,
                ...{ class: "action-btn" },
            }));
            const __VLS_88 = __VLS_87({
                disabled: true,
                ...{ class: "action-btn" },
            }, ...__VLS_functionalComponentArgsRest(__VLS_87));
            __VLS_89.slots.default;
            var __VLS_89;
        }
        var __VLS_42;
    }
    var __VLS_38;
}
/** @type {__VLS_StyleScopedClasses['executor-tasks']} */ ;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
/** @type {__VLS_StyleScopedClasses['page-title']} */ ;
/** @type {__VLS_StyleScopedClasses['filter-group']} */ ;
/** @type {__VLS_StyleScopedClasses['tasks-container']} */ ;
/** @type {__VLS_StyleScopedClasses['tasks-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['task-card']} */ ;
/** @type {__VLS_StyleScopedClasses['task-card-header']} */ ;
/** @type {__VLS_StyleScopedClasses['task-drill-name']} */ ;
/** @type {__VLS_StyleScopedClasses['task-card-body']} */ ;
/** @type {__VLS_StyleScopedClasses['task-step-name']} */ ;
/** @type {__VLS_StyleScopedClasses['task-description']} */ ;
/** @type {__VLS_StyleScopedClasses['task-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['task-deadline']} */ ;
/** @type {__VLS_StyleScopedClasses['task-assignee']} */ ;
/** @type {__VLS_StyleScopedClasses['task-card-footer']} */ ;
/** @type {__VLS_StyleScopedClasses['action-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['action-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['action-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['action-btn']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Clock: Clock,
            User: User,
            DrillStatusBadge: DrillStatusBadge,
            EmptyBox: EmptyBox,
            loading: loading,
            filterStatus: filterStatus,
            filteredTasks: filteredTasks,
            getStatusClass: getStatusClass,
            handleFilterChange: handleFilterChange,
            formatDeadline: formatDeadline,
            goToTaskDetail: goToTaskDetail,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
