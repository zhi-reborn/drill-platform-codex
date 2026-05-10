const __VLS_props = defineProps();
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['running']} */ ;
/** @type {__VLS_StyleScopedClasses['paused']} */ ;
/** @type {__VLS_StyleScopedClasses['completed']} */ ;
/** @type {__VLS_StyleScopedClasses['terminated']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "status-panel" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ class: "panel-title" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "status-list" },
});
for (const [drill] of __VLS_getVForSourceType((__VLS_ctx.drills))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (drill.id),
        ...{ class: "status-card" },
        ...{ class: (drill.status) },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "card-header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "drill-name" },
    });
    (drill.name);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "status-dot" },
        ...{ class: (drill.status) },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "card-progress" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "progress-bar" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "progress-fill" },
        ...{ style: ({ width: `${(drill.completed_steps / drill.total_steps) * 100}%` }) },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "progress-text" },
    });
    (drill.completed_steps);
    (drill.total_steps);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "card-footer" },
    });
    if (drill.created_by_name) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "assignee" },
        });
        const __VLS_0 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({}));
        const __VLS_2 = __VLS_1({}, ...__VLS_functionalComponentArgsRest(__VLS_1));
        __VLS_3.slots.default;
        const __VLS_4 = {}.User;
        /** @type {[typeof __VLS_components.User, ]} */ ;
        // @ts-ignore
        const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({}));
        const __VLS_6 = __VLS_5({}, ...__VLS_functionalComponentArgsRest(__VLS_5));
        var __VLS_3;
        (drill.created_by_name);
    }
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "template" },
    });
    (drill.template_name);
}
/** @type {__VLS_StyleScopedClasses['status-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['panel-title']} */ ;
/** @type {__VLS_StyleScopedClasses['status-list']} */ ;
/** @type {__VLS_StyleScopedClasses['status-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-header']} */ ;
/** @type {__VLS_StyleScopedClasses['drill-name']} */ ;
/** @type {__VLS_StyleScopedClasses['status-dot']} */ ;
/** @type {__VLS_StyleScopedClasses['card-progress']} */ ;
/** @type {__VLS_StyleScopedClasses['progress-bar']} */ ;
/** @type {__VLS_StyleScopedClasses['progress-fill']} */ ;
/** @type {__VLS_StyleScopedClasses['progress-text']} */ ;
/** @type {__VLS_StyleScopedClasses['card-footer']} */ ;
/** @type {__VLS_StyleScopedClasses['assignee']} */ ;
/** @type {__VLS_StyleScopedClasses['template']} */ ;
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
