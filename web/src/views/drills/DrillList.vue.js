import { ref } from 'vue';
import { Plus } from '@element-plus/icons-vue';
import DataTable from '@/components/common/DataTable.vue';
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue';
import ActionConfirm from '@/components/common/ActionConfirm.vue';
const loading = ref(false);
const total = ref(0);
const columns = [
    { prop: 'id', label: 'ID', width: 80 },
    { prop: 'title', label: '演练标题', minWidth: 200 },
    { prop: 'category', label: '类型', width: 100 },
    { prop: 'status', label: '状态', width: 100, slot: true },
    { prop: 'created_at', label: '创建时间', width: 180 },
];
const data = ref([]);
function handlePageChange({ page, size }) {
    console.log('Page change:', page, size);
}
function handleDelete(row) {
    console.log('Delete:', row);
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "drill-list-page" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h1, __VLS_intrinsicElements.h1)({
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
    onClick: (...[$event]) => {
        __VLS_ctx.$router.push('/drills/create');
    }
};
__VLS_3.slots.default;
const __VLS_8 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({}));
const __VLS_10 = __VLS_9({}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
const __VLS_12 = {}.Plus;
/** @type {[typeof __VLS_components.Plus, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({}));
const __VLS_14 = __VLS_13({}, ...__VLS_functionalComponentArgsRest(__VLS_13));
var __VLS_11;
var __VLS_3;
/** @type {[typeof DataTable, typeof DataTable, ]} */ ;
// @ts-ignore
const __VLS_16 = __VLS_asFunctionalComponent(DataTable, new DataTable({
    ...{ 'onPageChange': {} },
    columns: (__VLS_ctx.columns),
    data: (__VLS_ctx.data),
    loading: (__VLS_ctx.loading),
    pagination: true,
    total: (__VLS_ctx.total),
}));
const __VLS_17 = __VLS_16({
    ...{ 'onPageChange': {} },
    columns: (__VLS_ctx.columns),
    data: (__VLS_ctx.data),
    loading: (__VLS_ctx.loading),
    pagination: true,
    total: (__VLS_ctx.total),
}, ...__VLS_functionalComponentArgsRest(__VLS_16));
let __VLS_19;
let __VLS_20;
let __VLS_21;
const __VLS_22 = {
    onPageChange: (__VLS_ctx.handlePageChange)
};
__VLS_18.slots.default;
{
    const { status: __VLS_thisSlot } = __VLS_18.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    /** @type {[typeof DrillStatusBadge, ]} */ ;
    // @ts-ignore
    const __VLS_23 = __VLS_asFunctionalComponent(DrillStatusBadge, new DrillStatusBadge({
        status: (row.status),
        type: "drill",
    }));
    const __VLS_24 = __VLS_23({
        status: (row.status),
        type: "drill",
    }, ...__VLS_functionalComponentArgsRest(__VLS_23));
}
{
    const { actions: __VLS_thisSlot } = __VLS_18.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_26 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_27 = __VLS_asFunctionalComponent(__VLS_26, new __VLS_26({
        link: true,
        type: "primary",
        size: "small",
    }));
    const __VLS_28 = __VLS_27({
        link: true,
        type: "primary",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_27));
    __VLS_29.slots.default;
    var __VLS_29;
    /** @type {[typeof ActionConfirm, typeof ActionConfirm, ]} */ ;
    // @ts-ignore
    const __VLS_30 = __VLS_asFunctionalComponent(ActionConfirm, new ActionConfirm({
        ...{ 'onConfirm': {} },
        message: "确定要删除此演练吗？",
        danger: true,
        size: "small",
    }));
    const __VLS_31 = __VLS_30({
        ...{ 'onConfirm': {} },
        message: "确定要删除此演练吗？",
        danger: true,
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_30));
    let __VLS_33;
    let __VLS_34;
    let __VLS_35;
    const __VLS_36 = {
        onConfirm: (...[$event]) => {
            __VLS_ctx.handleDelete(row);
        }
    };
    __VLS_32.slots.default;
    var __VLS_32;
}
var __VLS_18;
/** @type {__VLS_StyleScopedClasses['drill-list-page']} */ ;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
/** @type {__VLS_StyleScopedClasses['page-title']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Plus: Plus,
            DataTable: DataTable,
            DrillStatusBadge: DrillStatusBadge,
            ActionConfirm: ActionConfirm,
            loading: loading,
            total: total,
            columns: columns,
            data: data,
            handlePageChange: handlePageChange,
            handleDelete: handleDelete,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
