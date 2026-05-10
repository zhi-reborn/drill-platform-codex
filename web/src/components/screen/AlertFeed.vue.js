import { ref } from 'vue';
const __VLS_props = defineProps();
const feedRef = ref();
function formatTime(iso) {
    const d = new Date(iso);
    return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "alert-feed" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({
    ...{ class: "panel-title" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "feed-list" },
    ref: "feedRef",
});
/** @type {typeof __VLS_ctx.feedRef} */ ;
for (const [alert] of __VLS_getVForSourceType((__VLS_ctx.alerts))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (alert.id),
        ...{ class: "alert-item" },
        ...{ class: (alert.level) },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "alert-time" },
    });
    (__VLS_ctx.formatTime(alert.created_at));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "alert-icon" },
    });
    const __VLS_0 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
        size: (14),
    }));
    const __VLS_2 = __VLS_1({
        size: (14),
    }, ...__VLS_functionalComponentArgsRest(__VLS_1));
    __VLS_3.slots.default;
    const __VLS_4 = ((alert.icon));
    // @ts-ignore
    const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({}));
    const __VLS_6 = __VLS_5({}, ...__VLS_functionalComponentArgsRest(__VLS_5));
    var __VLS_3;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "alert-message" },
    });
    (alert.message);
}
/** @type {__VLS_StyleScopedClasses['alert-feed']} */ ;
/** @type {__VLS_StyleScopedClasses['panel-title']} */ ;
/** @type {__VLS_StyleScopedClasses['feed-list']} */ ;
/** @type {__VLS_StyleScopedClasses['alert-item']} */ ;
/** @type {__VLS_StyleScopedClasses['alert-time']} */ ;
/** @type {__VLS_StyleScopedClasses['alert-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['alert-message']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            feedRef: feedRef,
            formatTime: formatTime,
        };
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
