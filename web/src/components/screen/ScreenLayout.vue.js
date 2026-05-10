import { ref, onMounted, onBeforeUnmount } from 'vue';
const cursorHidden = ref(false);
const containerRef = ref();
let hideTimer = null;
function resetHideTimer() {
    cursorHidden.value = false;
    if (hideTimer)
        clearTimeout(hideTimer);
    hideTimer = setTimeout(() => { cursorHidden.value = true; }, 3000);
}
function handleMouseMove() {
    resetHideTimer();
}
onMounted(() => resetHideTimer());
onBeforeUnmount(() => { if (hideTimer)
    clearTimeout(hideTimer); });
// ESC to exit fullscreen
function handleKeydown(e) {
    if (e.key === 'Escape') {
        document.exitFullscreen?.();
    }
    if (e.key === 'F' || e.key === 'f') {
        containerRef.value?.requestFullscreen?.();
    }
}
onMounted(() => window.addEventListener('keydown', handleKeydown));
onBeforeUnmount(() => window.removeEventListener('keydown', handleKeydown));
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['screen-layout']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ onMousemove: (__VLS_ctx.handleMouseMove) },
    ...{ onMouseleave: (...[$event]) => {
            __VLS_ctx.cursorHidden = true;
        } },
    ref: "containerRef",
    ...{ class: "screen-layout" },
    ...{ class: ({ 'cursor-hidden': __VLS_ctx.cursorHidden }) },
});
/** @type {typeof __VLS_ctx.containerRef} */ ;
var __VLS_0 = {};
/** @type {__VLS_StyleScopedClasses['screen-layout']} */ ;
/** @type {__VLS_StyleScopedClasses['cursor-hidden']} */ ;
// @ts-ignore
var __VLS_1 = __VLS_0;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            cursorHidden: cursorHidden,
            containerRef: containerRef,
            handleMouseMove: handleMouseMove,
        };
    },
});
const __VLS_component = (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
export default {};
; /* PartiallyEnd: #4569/main.vue */
