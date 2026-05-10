import { ref, computed, watch } from 'vue';
import { Document } from '@element-plus/icons-vue';
import { NOTIFICATION_TYPE_LABELS } from '@/types/notification';
import EmptyBox from '@/components/common/EmptyBox.vue';
const props = withDefaults(defineProps(), {
    loading: false,
    showPagination: false,
    total: 0,
    pageSize: 20,
});
const emit = defineEmits();
const currentPage = ref(1);
const notifications = computed(() => props.notifications);
function getTypeTag(type) {
    const map = {
        drill_started: 'primary',
        drill_completed: 'success',
        drill_paused: 'warning',
        drill_resumed: 'primary',
        drill_terminated: 'danger',
        task_assigned: 'info',
        step_complete: 'success',
        step_timeout: 'warning',
        system_alert: 'danger',
    };
    return (map[type] || 'info');
}
function getTypeLabel(type) {
    return NOTIFICATION_TYPE_LABELS[type] || type;
}
function formatTime(dateStr) {
    const date = new Date(dateStr);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);
    if (minutes < 1)
        return '刚刚';
    if (minutes < 60)
        return `${minutes} 分钟前`;
    if (hours < 24)
        return `${hours} 小时前`;
    if (days < 7)
        return `${days} 天前`;
    return date.toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
    });
}
function handleClick(notification) {
    emit('read', notification);
}
function handlePageChange(page) {
    emit('loadMore', page);
}
// 重置页码
watch(() => props.total, () => {
    currentPage.value = 1;
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_withDefaultsArg = (function (t) { return t; })({
    loading: false,
    showPagination: false,
    total: 0,
    pageSize: 20,
});
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "notification-list-container" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "notifications-list" },
});
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
if (!__VLS_ctx.loading && __VLS_ctx.notifications.length === 0) {
    /** @type {[typeof EmptyBox, ]} */ ;
    // @ts-ignore
    const __VLS_0 = __VLS_asFunctionalComponent(EmptyBox, new EmptyBox({
        title: "暂无消息",
        description: "当前没有通知消息",
    }));
    const __VLS_1 = __VLS_0({
        title: "暂无消息",
        description: "当前没有通知消息",
    }, ...__VLS_functionalComponentArgsRest(__VLS_0));
}
for (const [notification] of __VLS_getVForSourceType((__VLS_ctx.notifications))) {
    const __VLS_3 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_4 = __VLS_asFunctionalComponent(__VLS_3, new __VLS_3({
        ...{ 'onClick': {} },
        key: (notification.id),
        ...{ class: "notification-card" },
        ...{ class: ({ unread: !notification.is_read }) },
    }));
    const __VLS_5 = __VLS_4({
        ...{ 'onClick': {} },
        key: (notification.id),
        ...{ class: "notification-card" },
        ...{ class: ({ unread: !notification.is_read }) },
    }, ...__VLS_functionalComponentArgsRest(__VLS_4));
    let __VLS_7;
    let __VLS_8;
    let __VLS_9;
    const __VLS_10 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleClick(notification);
        }
    };
    __VLS_6.slots.default;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "notification-header" },
    });
    const __VLS_11 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_12 = __VLS_asFunctionalComponent(__VLS_11, new __VLS_11({
        type: (__VLS_ctx.getTypeTag(notification.type)),
        size: "small",
    }));
    const __VLS_13 = __VLS_12({
        type: (__VLS_ctx.getTypeTag(notification.type)),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_12));
    __VLS_14.slots.default;
    (__VLS_ctx.getTypeLabel(notification.type));
    var __VLS_14;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "notification-time" },
    });
    (__VLS_ctx.formatTime(notification.created_at));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "notification-title" },
    });
    (notification.title);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "notification-content" },
    });
    (notification.content);
    if (notification.drill_name) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "notification-drill" },
        });
        const __VLS_15 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_16 = __VLS_asFunctionalComponent(__VLS_15, new __VLS_15({}));
        const __VLS_17 = __VLS_16({}, ...__VLS_functionalComponentArgsRest(__VLS_16));
        __VLS_18.slots.default;
        const __VLS_19 = {}.Document;
        /** @type {[typeof __VLS_components.Document, ]} */ ;
        // @ts-ignore
        const __VLS_20 = __VLS_asFunctionalComponent(__VLS_19, new __VLS_19({}));
        const __VLS_21 = __VLS_20({}, ...__VLS_functionalComponentArgsRest(__VLS_20));
        var __VLS_18;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
        (notification.drill_name);
    }
    var __VLS_6;
}
if (__VLS_ctx.showPagination && __VLS_ctx.total > __VLS_ctx.pageSize) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "pagination-wrapper" },
    });
    const __VLS_23 = {}.ElPagination;
    /** @type {[typeof __VLS_components.ElPagination, typeof __VLS_components.elPagination, ]} */ ;
    // @ts-ignore
    const __VLS_24 = __VLS_asFunctionalComponent(__VLS_23, new __VLS_23({
        ...{ 'onCurrentChange': {} },
        currentPage: (__VLS_ctx.currentPage),
        pageSize: (__VLS_ctx.pageSize),
        total: (__VLS_ctx.total),
        pagerCount: (5),
        layout: "prev, pager, next",
    }));
    const __VLS_25 = __VLS_24({
        ...{ 'onCurrentChange': {} },
        currentPage: (__VLS_ctx.currentPage),
        pageSize: (__VLS_ctx.pageSize),
        total: (__VLS_ctx.total),
        pagerCount: (5),
        layout: "prev, pager, next",
    }, ...__VLS_functionalComponentArgsRest(__VLS_24));
    let __VLS_27;
    let __VLS_28;
    let __VLS_29;
    const __VLS_30 = {
        onCurrentChange: (__VLS_ctx.handlePageChange)
    };
    var __VLS_26;
}
/** @type {__VLS_StyleScopedClasses['notification-list-container']} */ ;
/** @type {__VLS_StyleScopedClasses['notifications-list']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-card']} */ ;
/** @type {__VLS_StyleScopedClasses['unread']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-header']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-time']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-title']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-content']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-drill']} */ ;
/** @type {__VLS_StyleScopedClasses['pagination-wrapper']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Document: Document,
            EmptyBox: EmptyBox,
            currentPage: currentPage,
            notifications: notifications,
            getTypeTag: getTypeTag,
            getTypeLabel: getTypeLabel,
            formatTime: formatTime,
            handleClick: handleClick,
            handlePageChange: handlePageChange,
        };
    },
    __typeEmits: {},
    __typeProps: {},
    props: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeEmits: {},
    __typeProps: {},
    props: {},
});
; /* PartiallyEnd: #4569/main.vue */
