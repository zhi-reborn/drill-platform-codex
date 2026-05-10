const __VLS_props = defineProps();
const __VLS_emit = defineEmits();
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "empty-box" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.svg, __VLS_intrinsicElements.svg)({
    ...{ class: "empty-icon" },
    viewBox: "0 0 200 150",
    fill: "none",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.rect)({
    x: "40",
    y: "30",
    width: "120",
    height: "90",
    rx: "8",
    stroke: "#484F58",
    'stroke-width': "2",
    fill: "none",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.line)({
    x1: "70",
    y1: "60",
    x2: "130",
    y2: "60",
    stroke: "#484F58",
    'stroke-width': "2",
    'stroke-linecap': "round",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.line)({
    x1: "70",
    y1: "80",
    x2: "110",
    y2: "80",
    stroke: "#484F58",
    'stroke-width': "2",
    'stroke-linecap': "round",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.circle)({
    cx: "100",
    cy: "110",
    r: "4",
    fill: "#484F58",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ class: "empty-title" },
});
(__VLS_ctx.title || '暂无数据');
if (__VLS_ctx.description) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "empty-desc" },
    });
    (__VLS_ctx.description);
}
if (__VLS_ctx.actionText) {
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
            if (!(__VLS_ctx.actionText))
                return;
            __VLS_ctx.$emit('action');
        }
    };
    __VLS_3.slots.default;
    (__VLS_ctx.actionText);
    var __VLS_3;
}
/** @type {__VLS_StyleScopedClasses['empty-box']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-title']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-desc']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeEmits: {},
    __typeProps: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeEmits: {},
    __typeProps: {},
});
; /* PartiallyEnd: #4569/main.vue */
