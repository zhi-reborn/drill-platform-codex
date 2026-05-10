const __VLS_props = withDefaults(defineProps(), {
    rows: 3
});
function randomWidth() {
    const min = 60;
    const max = 100;
    return `${min + Math.floor(Math.random() * (max - min))}%`;
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_withDefaultsArg = (function (t) { return t; })({
    rows: 3
});
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "loading-card" },
    ...{ style: ({ height: __VLS_ctx.height, width: __VLS_ctx.width }) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div)({
    ...{ class: "skeleton-title" },
});
for (const [i] of __VLS_getVForSourceType((__VLS_ctx.rows))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div)({
        ...{ class: "skeleton-text" },
        key: (i),
        ...{ style: ({ width: __VLS_ctx.randomWidth() }) },
    });
}
/** @type {__VLS_StyleScopedClasses['loading-card']} */ ;
/** @type {__VLS_StyleScopedClasses['skeleton-title']} */ ;
/** @type {__VLS_StyleScopedClasses['skeleton-text']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            randomWidth: randomWidth,
        };
    },
    __typeProps: {},
    props: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeProps: {},
    props: {},
});
; /* PartiallyEnd: #4569/main.vue */
