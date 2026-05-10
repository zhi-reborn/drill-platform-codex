const __VLS_props = defineProps();
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "metrics-bar" },
});
for (const [metric] of __VLS_getVForSourceType((__VLS_ctx.metrics))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (metric.label),
        ...{ class: "metric-pill" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "metric-icon" },
    });
    const __VLS_0 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
        size: (20),
    }));
    const __VLS_2 = __VLS_1({
        size: (20),
    }, ...__VLS_functionalComponentArgsRest(__VLS_1));
    __VLS_3.slots.default;
    const __VLS_4 = ((metric.icon));
    // @ts-ignore
    const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({}));
    const __VLS_6 = __VLS_5({}, ...__VLS_functionalComponentArgsRest(__VLS_5));
    var __VLS_3;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "metric-content" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "metric-label" },
    });
    (metric.label);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "metric-value" },
    });
    (metric.value);
    if (metric.trend) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "metric-trend" },
            ...{ class: (metric.trend > 0 ? 'up' : 'down') },
        });
        (metric.trend > 0 ? '↑' : '↓');
        (Math.abs(metric.trend));
    }
}
/** @type {__VLS_StyleScopedClasses['metrics-bar']} */ ;
/** @type {__VLS_StyleScopedClasses['metric-pill']} */ ;
/** @type {__VLS_StyleScopedClasses['metric-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['metric-content']} */ ;
/** @type {__VLS_StyleScopedClasses['metric-label']} */ ;
/** @type {__VLS_StyleScopedClasses['metric-value']} */ ;
/** @type {__VLS_StyleScopedClasses['metric-trend']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeProps: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeProps: {},
});
; /* PartiallyEnd: #4569/main.vue */
