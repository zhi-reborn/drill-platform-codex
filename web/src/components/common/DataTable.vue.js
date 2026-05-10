import { ref } from 'vue';
const props = withDefaults(defineProps(), {
    loading: false,
    selectable: false,
    pagination: false,
    total: 0
});
const emit = defineEmits();
const currentPage = ref(1);
const pageSize = ref(20);
function handleSelectionChange(selection) {
    emit('selection-change', selection);
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_withDefaultsArg = (function (t) { return t; })({
    loading: false,
    selectable: false,
    pagination: false,
    total: 0
});
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "data-table" },
});
const __VLS_0 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onSelectionChange': {} },
    data: (__VLS_ctx.data),
    stripe: true,
    border: true,
    ...{ style: {} },
}));
const __VLS_2 = __VLS_1({
    ...{ 'onSelectionChange': {} },
    data: (__VLS_ctx.data),
    stripe: true,
    border: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onSelectionChange: (__VLS_ctx.handleSelectionChange)
};
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
__VLS_3.slots.default;
if (__VLS_ctx.selectable) {
    const __VLS_8 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
        type: "selection",
        width: "48",
    }));
    const __VLS_10 = __VLS_9({
        type: "selection",
        width: "48",
    }, ...__VLS_functionalComponentArgsRest(__VLS_9));
}
for (const [col] of __VLS_getVForSourceType((__VLS_ctx.columns))) {
    const __VLS_12 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
        key: (col.prop),
        prop: (col.prop),
        label: (col.label),
        width: (col.width),
        minWidth: (col.minWidth),
        formatter: (col.formatter),
        sortable: (col.sortable),
        align: (col.align || 'left'),
    }));
    const __VLS_14 = __VLS_13({
        key: (col.prop),
        prop: (col.prop),
        label: (col.label),
        width: (col.width),
        minWidth: (col.minWidth),
        formatter: (col.formatter),
        sortable: (col.sortable),
        align: (col.align || 'left'),
    }, ...__VLS_functionalComponentArgsRest(__VLS_13));
    __VLS_15.slots.default;
    if (col.slot) {
        {
            const { default: __VLS_thisSlot } = __VLS_15.slots;
            const [scope] = __VLS_getSlotParams(__VLS_thisSlot);
            var __VLS_16 = {
                row: (scope.row),
            };
            var __VLS_17 = __VLS_tryAsConstant(col.prop);
        }
    }
    var __VLS_15;
}
if (__VLS_ctx.$slots.actions) {
    const __VLS_20 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
        label: "操作",
        width: "180",
        fixed: "right",
        align: "center",
    }));
    const __VLS_22 = __VLS_21({
        label: "操作",
        width: "180",
        fixed: "right",
        align: "center",
    }, ...__VLS_functionalComponentArgsRest(__VLS_21));
    __VLS_23.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_23.slots;
        const [scope] = __VLS_getSlotParams(__VLS_thisSlot);
        var __VLS_24 = {
            row: (scope.row),
        };
    }
    var __VLS_23;
}
var __VLS_3;
if (__VLS_ctx.pagination) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "table-pagination" },
    });
    const __VLS_26 = {}.ElPagination;
    /** @type {[typeof __VLS_components.ElPagination, typeof __VLS_components.elPagination, ]} */ ;
    // @ts-ignore
    const __VLS_27 = __VLS_asFunctionalComponent(__VLS_26, new __VLS_26({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        currentPage: (__VLS_ctx.currentPage),
        pageSize: (__VLS_ctx.pageSize),
        total: (__VLS_ctx.total),
        pageSizes: ([10, 20, 50, 100]),
        layout: "total, sizes, prev, pager, next, jumper",
    }));
    const __VLS_28 = __VLS_27({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        currentPage: (__VLS_ctx.currentPage),
        pageSize: (__VLS_ctx.pageSize),
        total: (__VLS_ctx.total),
        pageSizes: ([10, 20, 50, 100]),
        layout: "total, sizes, prev, pager, next, jumper",
    }, ...__VLS_functionalComponentArgsRest(__VLS_27));
    let __VLS_30;
    let __VLS_31;
    let __VLS_32;
    const __VLS_33 = {
        onCurrentChange: (...[$event]) => {
            if (!(__VLS_ctx.pagination))
                return;
            __VLS_ctx.emit('page-change', { page: __VLS_ctx.currentPage, size: __VLS_ctx.pageSize });
        }
    };
    const __VLS_34 = {
        onSizeChange: (...[$event]) => {
            if (!(__VLS_ctx.pagination))
                return;
            __VLS_ctx.emit('page-change', { page: 1, size: __VLS_ctx.pageSize });
        }
    };
    var __VLS_29;
}
/** @type {__VLS_StyleScopedClasses['data-table']} */ ;
/** @type {__VLS_StyleScopedClasses['table-pagination']} */ ;
// @ts-ignore
var __VLS_18 = __VLS_17, __VLS_19 = __VLS_16, __VLS_25 = __VLS_24;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            emit: emit,
            currentPage: currentPage,
            pageSize: pageSize,
            handleSelectionChange: handleSelectionChange,
        };
    },
    __typeEmits: {},
    __typeProps: {},
    props: {},
});
const __VLS_component = (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeEmits: {},
    __typeProps: {},
    props: {},
});
export default {};
; /* PartiallyEnd: #4569/main.vue */
