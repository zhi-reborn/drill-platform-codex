import EmptyBox from '@/components/common/EmptyBox.vue';
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "tasks-page" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h1, __VLS_intrinsicElements.h1)({
    ...{ class: "page-title" },
});
/** @type {[typeof EmptyBox, ]} */ ;
// @ts-ignore
const __VLS_0 = __VLS_asFunctionalComponent(EmptyBox, new EmptyBox({
    description: "暂无待处理任务",
}));
const __VLS_1 = __VLS_0({
    description: "暂无待处理任务",
}, ...__VLS_functionalComponentArgsRest(__VLS_0));
/** @type {__VLS_StyleScopedClasses['tasks-page']} */ ;
/** @type {__VLS_StyleScopedClasses['page-title']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            EmptyBox: EmptyBox,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
