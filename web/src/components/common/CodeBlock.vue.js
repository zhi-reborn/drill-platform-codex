import { ref } from 'vue';
const props = withDefaults(defineProps(), {
    showCopy: false
});
const copied = ref(false);
async function copyCode() {
    await navigator.clipboard.writeText(props.code);
    copied.value = true;
    setTimeout(() => copied.value = false, 2000);
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_withDefaultsArg = (function (t) { return t; })({
    showCopy: false
});
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.pre, __VLS_intrinsicElements.pre)({
    ...{ class: "code-block" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.code, __VLS_intrinsicElements.code)({
    ...{ class: "code-content" },
});
(__VLS_ctx.code);
if (__VLS_ctx.showCopy) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (__VLS_ctx.copyCode) },
        ...{ class: "copy-btn" },
    });
    (__VLS_ctx.copied ? '已复制' : '复制');
}
/** @type {__VLS_StyleScopedClasses['code-block']} */ ;
/** @type {__VLS_StyleScopedClasses['code-content']} */ ;
/** @type {__VLS_StyleScopedClasses['copy-btn']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            copied: copied,
            copyCode: copyCode,
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
