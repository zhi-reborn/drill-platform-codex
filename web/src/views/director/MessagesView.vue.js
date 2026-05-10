import { ref, computed, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { Check, Document } from '@element-plus/icons-vue';
import EmptyBox from '@/components/common/EmptyBox.vue';
import notificationsData from '@/mock/data/notifications.json';
const loading = ref(false);
const notifications = ref([]);
// 模拟当前用户 ID（实际应从 store 获取）
const currentUserId = 2;
const filteredNotifications = computed(() => {
    return notifications.value.filter(n => n.user_id === currentUserId);
});
function getNotificationTypeTag(type) {
    const map = {
        drill_started: 'primary',
        drill_completed: 'success',
        drill_paused: 'warning',
        task_assigned: 'info',
        step_complete: 'success',
        step_timeout: 'warning',
        system_alert: 'danger',
    };
    return map[type] || 'info';
}
function getNotificationTypeLabel(type) {
    const map = {
        drill_started: '演练开始',
        drill_completed: '演练完成',
        drill_paused: '演练暂停',
        task_assigned: '任务分配',
        step_complete: '步骤完成',
        step_timeout: '步骤超时',
        system_alert: '系统通知',
    };
    return map[type] || type;
}
function formatTime(dateStr) {
    const date = new Date(dateStr);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const hours = Math.floor(diff / (1000 * 60 * 60));
    if (hours < 1) {
        return '刚刚';
    }
    if (hours < 24) {
        return `${hours}小时前`;
    }
    return date.toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
    });
}
async function loadNotifications() {
    loading.value = true;
    try {
        notifications.value = notificationsData;
    }
    catch (error) {
        ElMessage.error('加载通知失败');
        console.error('Failed to load notifications:', error);
    }
    finally {
        loading.value = false;
    }
}
function markAsRead(notification) {
    notification.is_read = true;
}
function markAllAsRead() {
    filteredNotifications.value.forEach(n => {
        n.is_read = true;
    });
    ElMessage.success('已全部标为已读');
}
onMounted(() => {
    loadNotifications();
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-container" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({
    ...{ class: "page-title" },
});
const __VLS_0 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onClick': {} },
}));
const __VLS_2 = __VLS_1({
    ...{ 'onClick': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onClick: (__VLS_ctx.markAllAsRead)
};
__VLS_3.slots.default;
const __VLS_8 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({}));
const __VLS_10 = __VLS_9({}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
const __VLS_12 = {}.Check;
/** @type {[typeof __VLS_components.Check, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({}));
const __VLS_14 = __VLS_13({}, ...__VLS_functionalComponentArgsRest(__VLS_13));
var __VLS_11;
var __VLS_3;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-content" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "notifications-list" },
});
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.loading) }, null, null);
if (!__VLS_ctx.loading && __VLS_ctx.filteredNotifications.length === 0) {
    /** @type {[typeof EmptyBox, ]} */ ;
    // @ts-ignore
    const __VLS_16 = __VLS_asFunctionalComponent(EmptyBox, new EmptyBox({
        title: "暂无消息",
        description: "当前没有通知消息",
    }));
    const __VLS_17 = __VLS_16({
        title: "暂无消息",
        description: "当前没有通知消息",
    }, ...__VLS_functionalComponentArgsRest(__VLS_16));
}
for (const [notification] of __VLS_getVForSourceType((__VLS_ctx.filteredNotifications))) {
    const __VLS_19 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_20 = __VLS_asFunctionalComponent(__VLS_19, new __VLS_19({
        ...{ 'onClick': {} },
        key: (notification.id),
        ...{ class: "notification-card" },
        ...{ class: ({ unread: !notification.is_read }) },
    }));
    const __VLS_21 = __VLS_20({
        ...{ 'onClick': {} },
        key: (notification.id),
        ...{ class: "notification-card" },
        ...{ class: ({ unread: !notification.is_read }) },
    }, ...__VLS_functionalComponentArgsRest(__VLS_20));
    let __VLS_23;
    let __VLS_24;
    let __VLS_25;
    const __VLS_26 = {
        onClick: (...[$event]) => {
            __VLS_ctx.markAsRead(notification);
        }
    };
    __VLS_22.slots.default;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "notification-header" },
    });
    const __VLS_27 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
        type: (__VLS_ctx.getNotificationTypeTag(notification.type)),
        size: "small",
    }));
    const __VLS_29 = __VLS_28({
        type: (__VLS_ctx.getNotificationTypeTag(notification.type)),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_28));
    __VLS_30.slots.default;
    (__VLS_ctx.getNotificationTypeLabel(notification.type));
    var __VLS_30;
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
        const __VLS_31 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({}));
        const __VLS_33 = __VLS_32({}, ...__VLS_functionalComponentArgsRest(__VLS_32));
        __VLS_34.slots.default;
        const __VLS_35 = {}.Document;
        /** @type {[typeof __VLS_components.Document, ]} */ ;
        // @ts-ignore
        const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({}));
        const __VLS_37 = __VLS_36({}, ...__VLS_functionalComponentArgsRest(__VLS_36));
        var __VLS_34;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
        (notification.drill_name);
    }
    var __VLS_22;
}
/** @type {__VLS_StyleScopedClasses['page-container']} */ ;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
/** @type {__VLS_StyleScopedClasses['page-title']} */ ;
/** @type {__VLS_StyleScopedClasses['page-content']} */ ;
/** @type {__VLS_StyleScopedClasses['notifications-list']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-card']} */ ;
/** @type {__VLS_StyleScopedClasses['unread']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-header']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-time']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-title']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-content']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-drill']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Check: Check,
            Document: Document,
            EmptyBox: EmptyBox,
            loading: loading,
            filteredNotifications: filteredNotifications,
            getNotificationTypeTag: getNotificationTypeTag,
            getNotificationTypeLabel: getNotificationTypeLabel,
            formatTime: formatTime,
            markAsRead: markAsRead,
            markAllAsRead: markAllAsRead,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
