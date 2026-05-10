import { ElMessageBox } from 'element-plus';
const props = withDefaults(defineProps(), {
    title: '',
    danger: false,
    type: 'primary',
    size: 'default',
    disabled: false
});
const emit = defineEmits();
async function handleConfirm() {
    try {
        await ElMessageBox.confirm(props.message, props.title || (props.danger ? '危险操作确认' : '操作确认'), {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: props.danger ? 'warning' : 'info',
        });
        emit('confirm');
    }
    catch {
        // User cancelled
    }
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_withDefaultsArg = (function (t) { return t; })({
    title: '',
    danger: false,
    type: 'primary',
    size: 'default',
    disabled: false
});
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
const __VLS_0 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onClick': {} },
    type: (__VLS_ctx.danger ? 'danger' : __VLS_ctx.type),
    size: (__VLS_ctx.size),
    disabled: (__VLS_ctx.disabled),
}));
const __VLS_2 = __VLS_1({
    ...{ 'onClick': {} },
    type: (__VLS_ctx.danger ? 'danger' : __VLS_ctx.type),
    size: (__VLS_ctx.size),
    disabled: (__VLS_ctx.disabled),
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onClick: (__VLS_ctx.handleConfirm)
};
var __VLS_8 = {};
__VLS_3.slots.default;
var __VLS_9 = {};
var __VLS_3;
// @ts-ignore
var __VLS_10 = __VLS_9;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            handleConfirm: handleConfirm,
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
