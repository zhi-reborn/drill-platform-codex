import { ref, computed, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { Check } from '@element-plus/icons-vue';
import { useNotificationStore } from '@/stores/notifications';
import { useAuthStore } from '@/stores/auth';
import NotificationList from '@/components/notifications/NotificationList.vue';
const notifStore = useNotificationStore();
const authStore = useAuthStore();
const loading = ref(false);
const filterType = ref('');
const pageSize = 20;
// 合并 store 数据 + 筛选
const allNotifications = computed(() => notifStore.notifications);
const filteredNotifications = computed(() => {
    if (!filterType.value)
        return allNotifications.value;
    return allNotifications.value.filter(n => n.type === filterType.value);
});
const unreadCount = computed(() => notifStore.unreadCount);
onMounted(async () => {
    loading.value = true;
    try {
        await notifStore.fetchNotifications();
    }
    finally {
        loading.value = false;
    }
});
function handleMarkAsRead(notification) {
    if (!notification.is_read) {
        notifStore.markAsRead(notification.id);
    }
}
function handleMarkAllAsRead() {
    notifStore.markAllAsRead();
    ElMessage.success('已全部标为已读');
}
function handleLoadMore(page) {
    // 页面模式下，store 已经加载了足够的数据
    // 如果需要分页获取更多，可以在这里调用 API
    console.log('Load page:', page);
}
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
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "header-actions" },
});
const __VLS_0 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    modelValue: (__VLS_ctx.filterType),
    placeholder: "筛选类型",
    clearable: true,
    ...{ style: {} },
}));
const __VLS_2 = __VLS_1({
    modelValue: (__VLS_ctx.filterType),
    placeholder: "筛选类型",
    clearable: true,
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
const __VLS_4 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
    label: "全部",
    value: "",
}));
const __VLS_6 = __VLS_5({
    label: "全部",
    value: "",
}, ...__VLS_functionalComponentArgsRest(__VLS_5));
const __VLS_8 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    label: "演练开始",
    value: "drill_started",
}));
const __VLS_10 = __VLS_9({
    label: "演练开始",
    value: "drill_started",
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
const __VLS_12 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    label: "演练完成",
    value: "drill_completed",
}));
const __VLS_14 = __VLS_13({
    label: "演练完成",
    value: "drill_completed",
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
const __VLS_16 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    label: "演练暂停",
    value: "drill_paused",
}));
const __VLS_18 = __VLS_17({
    label: "演练暂停",
    value: "drill_paused",
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
const __VLS_20 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    label: "任务分配",
    value: "task_assigned",
}));
const __VLS_22 = __VLS_21({
    label: "任务分配",
    value: "task_assigned",
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
const __VLS_24 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    label: "步骤完成",
    value: "step_complete",
}));
const __VLS_26 = __VLS_25({
    label: "步骤完成",
    value: "step_complete",
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
const __VLS_28 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    label: "步骤超时",
    value: "step_timeout",
}));
const __VLS_30 = __VLS_29({
    label: "步骤超时",
    value: "step_timeout",
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
const __VLS_32 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    label: "系统公告",
    value: "system_alert",
}));
const __VLS_34 = __VLS_33({
    label: "系统公告",
    value: "system_alert",
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
var __VLS_3;
const __VLS_36 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
    ...{ 'onClick': {} },
    disabled: (__VLS_ctx.unreadCount === 0),
}));
const __VLS_38 = __VLS_37({
    ...{ 'onClick': {} },
    disabled: (__VLS_ctx.unreadCount === 0),
}, ...__VLS_functionalComponentArgsRest(__VLS_37));
let __VLS_40;
let __VLS_41;
let __VLS_42;
const __VLS_43 = {
    onClick: (__VLS_ctx.handleMarkAllAsRead)
};
__VLS_39.slots.default;
const __VLS_44 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({}));
const __VLS_46 = __VLS_45({}, ...__VLS_functionalComponentArgsRest(__VLS_45));
__VLS_47.slots.default;
const __VLS_48 = {}.Check;
/** @type {[typeof __VLS_components.Check, ]} */ ;
// @ts-ignore
const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({}));
const __VLS_50 = __VLS_49({}, ...__VLS_functionalComponentArgsRest(__VLS_49));
var __VLS_47;
var __VLS_39;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-content" },
});
/** @type {[typeof NotificationList, ]} */ ;
// @ts-ignore
const __VLS_52 = __VLS_asFunctionalComponent(NotificationList, new NotificationList({
    ...{ 'onRead': {} },
    notifications: (__VLS_ctx.filteredNotifications),
    loading: (__VLS_ctx.loading),
}));
const __VLS_53 = __VLS_52({
    ...{ 'onRead': {} },
    notifications: (__VLS_ctx.filteredNotifications),
    loading: (__VLS_ctx.loading),
}, ...__VLS_functionalComponentArgsRest(__VLS_52));
let __VLS_55;
let __VLS_56;
let __VLS_57;
const __VLS_58 = {
    onRead: (__VLS_ctx.handleMarkAsRead)
};
var __VLS_54;
/** @type {__VLS_StyleScopedClasses['page-container']} */ ;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
/** @type {__VLS_StyleScopedClasses['page-title']} */ ;
/** @type {__VLS_StyleScopedClasses['header-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['page-content']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Check: Check,
            NotificationList: NotificationList,
            loading: loading,
            filterType: filterType,
            filteredNotifications: filteredNotifications,
            unreadCount: unreadCount,
            handleMarkAsRead: handleMarkAsRead,
            handleMarkAllAsRead: handleMarkAllAsRead,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
