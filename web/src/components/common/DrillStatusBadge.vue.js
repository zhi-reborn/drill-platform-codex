import { computed } from 'vue';
const props = defineProps();
const drillStatusMap = {
    pending: { label: '待启动', color: '#6E7681' },
    running: { label: '执行中', color: '#55C3D3' },
    paused: { label: '已暂停', color: '#B8860B' },
    completed: { label: '已完成', color: '#2EA043' },
    terminated: { label: '已终止', color: '#DA3633' },
};
const stepStatusMap = {
    pending: { label: '待执行', color: '#6E7681' },
    running: { label: '执行中', color: '#55C3D3' },
    completed: { label: '已完成', color: '#2EA043' },
    timeout: { label: '已超时', color: '#B8860B' },
    skipped: { label: '已跳过', color: '#6E7681' },
    issue: { label: '异常', color: '#DA3633' },
};
const statusMap = props.type === 'step' ? stepStatusMap : drillStatusMap;
const info = computed(() => statusMap[props.status] ?? { label: props.status, color: '#6E7681' });
const label = computed(() => info.value.label);
const isRunning = computed(() => props.status === 'running');
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
__VLS_ctx.info.color;
__VLS_ctx.info.color;
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "status-badge" },
    ...{ class: ([__VLS_ctx.type, __VLS_ctx.status]) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "status-dot" },
    ...{ class: ({ pulse: __VLS_ctx.isRunning }) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "status-text" },
});
(__VLS_ctx.label);
/** @type {__VLS_StyleScopedClasses['status-badge']} */ ;
/** @type {__VLS_StyleScopedClasses['status-dot']} */ ;
/** @type {__VLS_StyleScopedClasses['pulse']} */ ;
/** @type {__VLS_StyleScopedClasses['status-text']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            info: info,
            label: label,
            isRunning: isRunning,
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
