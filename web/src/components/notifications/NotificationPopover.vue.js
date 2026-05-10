import { computed, onMounted } from 'vue';
import { Bell, Close } from '@element-plus/icons-vue';
import { useRouter } from 'vue-router';
import { useNotificationStore } from '@/stores/notifications';
import { NOTIFICATION_TYPE_LABELS } from '@/types/notification';
import { useAuthStore } from '@/stores/auth';
const router = useRouter();
const notifStore = useNotificationStore();
const authStore = useAuthStore();
onMounted(async () => {
    authStore.restoreSession();
    console.log('[NotificationPopover] Auth restored:', {
        isAuthenticated: authStore.isAuthenticated,
        user: authStore.user,
        role: authStore.role,
    });
    await notifStore.fetchNotifications();
    console.log('[NotificationPopover] Notifications loaded:', {
        total: notifStore.notifications.length,
        unread: notifStore.unreadCount,
        items: notifStore.notifications.slice(0, 3),
    });
});
const notifications = computed(() => notifStore.notifications.slice(0, 10));
const unreadCount = computed(() => notifStore.unreadCount);
const hasMore = computed(() => notifStore.notifications.length > 10);
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
    return date.toLocaleDateString('zh-CN');
}
function handleItemClick(n) {
    if (!n.is_read) {
        notifStore.markAsRead(n.id);
    }
}
function handleMarkAllAsRead() {
    notifStore.markAllAsRead();
}
function handleDelete(id) {
    notifStore.deleteNotification(id);
}
function handleViewAll() {
    const role = authStore.role;
    router.push(`/${role}/messages`);
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['delete-btn']} */ ;
// CSS variable injection 
// CSS variable injection end 
const __VLS_0 = {}.ElPopover;
/** @type {[typeof __VLS_components.ElPopover, typeof __VLS_components.elPopover, typeof __VLS_components.ElPopover, typeof __VLS_components.elPopover, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    placement: "bottom-end",
    width: (380),
    trigger: "click",
}));
const __VLS_2 = __VLS_1({
    placement: "bottom-end",
    width: (380),
    trigger: "click",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
var __VLS_4 = {};
__VLS_3.slots.default;
{
    const { reference: __VLS_thisSlot } = __VLS_3.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "notification-bell" },
    });
    const __VLS_5 = {}.ElBadge;
    /** @type {[typeof __VLS_components.ElBadge, typeof __VLS_components.elBadge, typeof __VLS_components.ElBadge, typeof __VLS_components.elBadge, ]} */ ;
    // @ts-ignore
    const __VLS_6 = __VLS_asFunctionalComponent(__VLS_5, new __VLS_5({
        value: (__VLS_ctx.unreadCount),
        hidden: (__VLS_ctx.unreadCount === 0),
        max: (99),
    }));
    const __VLS_7 = __VLS_6({
        value: (__VLS_ctx.unreadCount),
        hidden: (__VLS_ctx.unreadCount === 0),
        max: (99),
    }, ...__VLS_functionalComponentArgsRest(__VLS_6));
    __VLS_8.slots.default;
    const __VLS_9 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_10 = __VLS_asFunctionalComponent(__VLS_9, new __VLS_9({
        text: true,
    }));
    const __VLS_11 = __VLS_10({
        text: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_10));
    __VLS_12.slots.default;
    const __VLS_13 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_14 = __VLS_asFunctionalComponent(__VLS_13, new __VLS_13({
        size: (18),
    }));
    const __VLS_15 = __VLS_14({
        size: (18),
    }, ...__VLS_functionalComponentArgsRest(__VLS_14));
    __VLS_16.slots.default;
    const __VLS_17 = {}.Bell;
    /** @type {[typeof __VLS_components.Bell, ]} */ ;
    // @ts-ignore
    const __VLS_18 = __VLS_asFunctionalComponent(__VLS_17, new __VLS_17({}));
    const __VLS_19 = __VLS_18({}, ...__VLS_functionalComponentArgsRest(__VLS_18));
    var __VLS_16;
    var __VLS_12;
    var __VLS_8;
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "notification-popover" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "popover-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "popover-title" },
});
if (__VLS_ctx.notifications.length > 0) {
    const __VLS_21 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_22 = __VLS_asFunctionalComponent(__VLS_21, new __VLS_21({
        ...{ 'onClick': {} },
        text: true,
        size: "small",
    }));
    const __VLS_23 = __VLS_22({
        ...{ 'onClick': {} },
        text: true,
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_22));
    let __VLS_25;
    let __VLS_26;
    let __VLS_27;
    const __VLS_28 = {
        onClick: (__VLS_ctx.handleMarkAllAsRead)
    };
    __VLS_24.slots.default;
    var __VLS_24;
}
const __VLS_29 = {}.ElScrollbar;
/** @type {[typeof __VLS_components.ElScrollbar, typeof __VLS_components.elScrollbar, typeof __VLS_components.ElScrollbar, typeof __VLS_components.elScrollbar, ]} */ ;
// @ts-ignore
const __VLS_30 = __VLS_asFunctionalComponent(__VLS_29, new __VLS_29({
    maxHeight: "400px",
}));
const __VLS_31 = __VLS_30({
    maxHeight: "400px",
}, ...__VLS_functionalComponentArgsRest(__VLS_30));
__VLS_32.slots.default;
if (__VLS_ctx.notifications.length === 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "empty-state" },
    });
    const __VLS_33 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_34 = __VLS_asFunctionalComponent(__VLS_33, new __VLS_33({
        description: "暂无消息",
        imageSize: (60),
    }));
    const __VLS_35 = __VLS_34({
        description: "暂无消息",
        imageSize: (60),
    }, ...__VLS_functionalComponentArgsRest(__VLS_34));
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "notification-list" },
    });
    for (const [n] of __VLS_getVForSourceType((__VLS_ctx.notifications))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ onClick: (...[$event]) => {
                    if (!!(__VLS_ctx.notifications.length === 0))
                        return;
                    __VLS_ctx.handleItemClick(n);
                } },
            key: (n.id),
            ...{ class: "notification-item" },
            ...{ class: ({ 'is-unread': !n.is_read }) },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "notification-content" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "notification-header" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "notification-type" },
        });
        (__VLS_ctx.getTypeLabel(n.type));
        const __VLS_37 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_38 = __VLS_asFunctionalComponent(__VLS_37, new __VLS_37({
            ...{ 'onClick': {} },
            ...{ class: "delete-btn" },
            text: true,
            size: "small",
        }));
        const __VLS_39 = __VLS_38({
            ...{ 'onClick': {} },
            ...{ class: "delete-btn" },
            text: true,
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_38));
        let __VLS_41;
        let __VLS_42;
        let __VLS_43;
        const __VLS_44 = {
            onClick: (...[$event]) => {
                if (!!(__VLS_ctx.notifications.length === 0))
                    return;
                __VLS_ctx.handleDelete(n.id);
            }
        };
        __VLS_40.slots.default;
        const __VLS_45 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_46 = __VLS_asFunctionalComponent(__VLS_45, new __VLS_45({}));
        const __VLS_47 = __VLS_46({}, ...__VLS_functionalComponentArgsRest(__VLS_46));
        __VLS_48.slots.default;
        const __VLS_49 = {}.Close;
        /** @type {[typeof __VLS_components.Close, ]} */ ;
        // @ts-ignore
        const __VLS_50 = __VLS_asFunctionalComponent(__VLS_49, new __VLS_49({}));
        const __VLS_51 = __VLS_50({}, ...__VLS_functionalComponentArgsRest(__VLS_50));
        var __VLS_48;
        var __VLS_40;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "notification-title" },
        });
        (n.title);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "notification-time" },
        });
        (__VLS_ctx.formatTime(n.created_at));
    }
    if (__VLS_ctx.hasMore) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "popover-footer" },
        });
        const __VLS_53 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_54 = __VLS_asFunctionalComponent(__VLS_53, new __VLS_53({
            ...{ 'onClick': {} },
            text: true,
            size: "small",
        }));
        const __VLS_55 = __VLS_54({
            ...{ 'onClick': {} },
            text: true,
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_54));
        let __VLS_57;
        let __VLS_58;
        let __VLS_59;
        const __VLS_60 = {
            onClick: (__VLS_ctx.handleViewAll)
        };
        __VLS_56.slots.default;
        var __VLS_56;
    }
}
var __VLS_32;
var __VLS_3;
/** @type {__VLS_StyleScopedClasses['notification-bell']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-popover']} */ ;
/** @type {__VLS_StyleScopedClasses['popover-header']} */ ;
/** @type {__VLS_StyleScopedClasses['popover-title']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-list']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-item']} */ ;
/** @type {__VLS_StyleScopedClasses['is-unread']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-content']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-header']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-type']} */ ;
/** @type {__VLS_StyleScopedClasses['delete-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-title']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-time']} */ ;
/** @type {__VLS_StyleScopedClasses['popover-footer']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Bell: Bell,
            Close: Close,
            notifications: notifications,
            unreadCount: unreadCount,
            hasMore: hasMore,
            getTypeLabel: getTypeLabel,
            formatTime: formatTime,
            handleItemClick: handleItemClick,
            handleMarkAllAsRead: handleMarkAllAsRead,
            handleDelete: handleDelete,
            handleViewAll: handleViewAll,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
