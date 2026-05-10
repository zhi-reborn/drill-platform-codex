import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue';
import instancesData from '@/mock/data/instances.json';
const router = useRouter();
const statusFilter = ref('');
const instances = ref([]);
const filteredInstances = computed(() => {
    if (!statusFilter.value) {
        return instances.value;
    }
    return instances.value.filter(i => i.status === statusFilter.value);
});
function formatTime(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
    });
}
async function loadInstances() {
    try {
        instances.value = instancesData;
        ElMessage.success('演练列表加载成功');
    }
    catch (error) {
        ElMessage.error('加载演练列表失败');
        console.error('Failed to load instances:', error);
    }
}
function viewDetail(instance) {
    router.push(`/director/monitor/${instance.id}`);
}
loadInstances();
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
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-content" },
});
const __VLS_0 = {}.ElRadioGroup;
/** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    modelValue: (__VLS_ctx.statusFilter),
    ...{ class: "status-filter" },
}));
const __VLS_2 = __VLS_1({
    modelValue: (__VLS_ctx.statusFilter),
    ...{ class: "status-filter" },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
const __VLS_4 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
    value: "",
}));
const __VLS_6 = __VLS_5({
    value: "",
}, ...__VLS_functionalComponentArgsRest(__VLS_5));
__VLS_7.slots.default;
var __VLS_7;
const __VLS_8 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    value: "pending",
}));
const __VLS_10 = __VLS_9({
    value: "pending",
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
var __VLS_11;
const __VLS_12 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    value: "running",
}));
const __VLS_14 = __VLS_13({
    value: "running",
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
__VLS_15.slots.default;
var __VLS_15;
const __VLS_16 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    value: "paused",
}));
const __VLS_18 = __VLS_17({
    value: "paused",
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
var __VLS_19;
const __VLS_20 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    value: "completed",
}));
const __VLS_22 = __VLS_21({
    value: "completed",
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_23.slots.default;
var __VLS_23;
const __VLS_24 = {}.ElRadioButton;
/** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    value: "terminated",
}));
const __VLS_26 = __VLS_25({
    value: "terminated",
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
__VLS_27.slots.default;
var __VLS_27;
var __VLS_3;
const __VLS_28 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    data: (__VLS_ctx.filteredInstances),
    ...{ style: {} },
    ...{ class: "drills-table" },
}));
const __VLS_30 = __VLS_29({
    data: (__VLS_ctx.filteredInstances),
    ...{ style: {} },
    ...{ class: "drills-table" },
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
__VLS_31.slots.default;
const __VLS_32 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    prop: "name",
    label: "演练名",
    minWidth: "200",
}));
const __VLS_34 = __VLS_33({
    prop: "name",
    label: "演练名",
    minWidth: "200",
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
const __VLS_36 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
    prop: "template_name",
    label: "模板",
    width: "180",
}));
const __VLS_38 = __VLS_37({
    prop: "template_name",
    label: "模板",
    width: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_37));
const __VLS_40 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
    prop: "status",
    label: "状态",
    width: "120",
}));
const __VLS_42 = __VLS_41({
    prop: "status",
    label: "状态",
    width: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_41));
__VLS_43.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_43.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    /** @type {[typeof DrillStatusBadge, ]} */ ;
    // @ts-ignore
    const __VLS_44 = __VLS_asFunctionalComponent(DrillStatusBadge, new DrillStatusBadge({
        status: (row.status),
        type: "drill",
    }));
    const __VLS_45 = __VLS_44({
        status: (row.status),
        type: "drill",
    }, ...__VLS_functionalComponentArgsRest(__VLS_44));
}
var __VLS_43;
const __VLS_47 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_48 = __VLS_asFunctionalComponent(__VLS_47, new __VLS_47({
    prop: "created_by_name",
    label: "创建人",
    width: "120",
}));
const __VLS_49 = __VLS_48({
    prop: "created_by_name",
    label: "创建人",
    width: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_48));
const __VLS_51 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_52 = __VLS_asFunctionalComponent(__VLS_51, new __VLS_51({
    label: "进度",
    width: "150",
}));
const __VLS_53 = __VLS_52({
    label: "进度",
    width: "150",
}, ...__VLS_functionalComponentArgsRest(__VLS_52));
__VLS_54.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_54.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_55 = {}.ElProgress;
    /** @type {[typeof __VLS_components.ElProgress, typeof __VLS_components.elProgress, ]} */ ;
    // @ts-ignore
    const __VLS_56 = __VLS_asFunctionalComponent(__VLS_55, new __VLS_55({
        percentage: (Math.round(row.completed_steps / row.total_steps * 100)),
        strokeWidth: (6),
        status: (row.status === 'completed' ? 'success' : undefined),
    }));
    const __VLS_57 = __VLS_56({
        percentage: (Math.round(row.completed_steps / row.total_steps * 100)),
        strokeWidth: (6),
        status: (row.status === 'completed' ? 'success' : undefined),
    }, ...__VLS_functionalComponentArgsRest(__VLS_56));
}
var __VLS_54;
const __VLS_59 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
    prop: "started_at",
    label: "开始时间",
    width: "180",
}));
const __VLS_61 = __VLS_60({
    prop: "started_at",
    label: "开始时间",
    width: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_60));
__VLS_62.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_62.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.started_at ? __VLS_ctx.formatTime(row.started_at) : '-');
}
var __VLS_62;
const __VLS_63 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_64 = __VLS_asFunctionalComponent(__VLS_63, new __VLS_63({
    label: "操作",
    width: "100",
    fixed: "right",
}));
const __VLS_65 = __VLS_64({
    label: "操作",
    width: "100",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_64));
__VLS_66.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_66.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_67 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
    }));
    const __VLS_69 = __VLS_68({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_68));
    let __VLS_71;
    let __VLS_72;
    let __VLS_73;
    const __VLS_74 = {
        onClick: (...[$event]) => {
            __VLS_ctx.viewDetail(row);
        }
    };
    __VLS_70.slots.default;
    var __VLS_70;
}
var __VLS_66;
var __VLS_31;
/** @type {__VLS_StyleScopedClasses['page-container']} */ ;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
/** @type {__VLS_StyleScopedClasses['page-title']} */ ;
/** @type {__VLS_StyleScopedClasses['page-content']} */ ;
/** @type {__VLS_StyleScopedClasses['status-filter']} */ ;
/** @type {__VLS_StyleScopedClasses['drills-table']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            DrillStatusBadge: DrillStatusBadge,
            statusFilter: statusFilter,
            filteredInstances: filteredInstances,
            formatTime: formatTime,
            viewDetail: viewDetail,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
