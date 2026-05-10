import { ref, computed, onMounted } from 'vue';
import { Fold, Expand, Monitor, Bell, ArrowDown, SwitchButton } from '@element-plus/icons-vue';
import { useAuthStore } from '@/stores/auth';
import { useNotificationStore } from '@/stores/notifications';
import { useWsStore } from '@/stores/ws';
import { useRouter } from 'vue-router';
const router = useRouter();
const props = withDefaults(defineProps(), {
    collapsed: false
});
const __VLS_emit = defineEmits();
const authStore = useAuthStore();
const notifStore = useNotificationStore();
const wsStore = useWsStore();
const showNotifications = ref(false);
const userInitial = computed(() => authStore.userInitial);
const userName = computed(() => authStore.userName);
const roleName = computed(() => authStore.roleName);
const roleType = computed(() => authStore.roleType);
const unreadCount = computed(() => notifStore.unreadCount);
const wsStatus = computed(() => wsStore.status);
const wsStatusText = computed(() => wsStore.statusText);
onMounted(() => {
    authStore.restoreSession();
    notifStore.fetchNotifications();
    wsStore.update();
});
function handleUserCommand(command) {
    if (command === 'logout') {
        authStore.logout();
    }
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_withDefaultsArg = (function (t) { return t; })({
    collapsed: false
});
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['status-dot']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.header, __VLS_intrinsicElements.header)({
    ...{ class: "app-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "header-left" },
});
const __VLS_0 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onClick': {} },
    text: true,
    ...{ class: "collapse-btn" },
}));
const __VLS_2 = __VLS_1({
    ...{ 'onClick': {} },
    text: true,
    ...{ class: "collapse-btn" },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onClick: (...[$event]) => {
        __VLS_ctx.$emit('toggle-sidebar');
    }
};
__VLS_3.slots.default;
const __VLS_8 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    size: (20),
}));
const __VLS_10 = __VLS_9({
    size: (20),
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
if (!__VLS_ctx.collapsed) {
    const __VLS_12 = {}.Fold;
    /** @type {[typeof __VLS_components.Fold, ]} */ ;
    // @ts-ignore
    const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({}));
    const __VLS_14 = __VLS_13({}, ...__VLS_functionalComponentArgsRest(__VLS_13));
}
else {
    const __VLS_16 = {}.Expand;
    /** @type {[typeof __VLS_components.Expand, ]} */ ;
    // @ts-ignore
    const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({}));
    const __VLS_18 = __VLS_17({}, ...__VLS_functionalComponentArgsRest(__VLS_17));
}
var __VLS_11;
var __VLS_3;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "logo" },
});
const __VLS_20 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    size: (24),
    color: "#55C3D3",
}));
const __VLS_22 = __VLS_21({
    size: (24),
    color: "#55C3D3",
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_23.slots.default;
const __VLS_24 = {}.Monitor;
/** @type {[typeof __VLS_components.Monitor, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({}));
const __VLS_26 = __VLS_25({}, ...__VLS_functionalComponentArgsRest(__VLS_25));
var __VLS_23;
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "logo-text" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "header-right" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "ws-status" },
    ...{ class: (__VLS_ctx.wsStatus) },
    title: (__VLS_ctx.wsStatusText),
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "status-dot" },
});
const __VLS_28 = {}.ElPopover;
/** @type {[typeof __VLS_components.ElPopover, typeof __VLS_components.elPopover, typeof __VLS_components.ElPopover, typeof __VLS_components.elPopover, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    visible: (__VLS_ctx.showNotifications),
    placement: "bottom-end",
    width: (360),
    trigger: "click",
}));
const __VLS_30 = __VLS_29({
    visible: (__VLS_ctx.showNotifications),
    placement: "bottom-end",
    width: (360),
    trigger: "click",
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
__VLS_31.slots.default;
{
    const { reference: __VLS_thisSlot } = __VLS_31.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "notification-bell" },
    });
    const __VLS_32 = {}.ElBadge;
    /** @type {[typeof __VLS_components.ElBadge, typeof __VLS_components.elBadge, typeof __VLS_components.ElBadge, typeof __VLS_components.elBadge, ]} */ ;
    // @ts-ignore
    const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
        value: (__VLS_ctx.unreadCount),
        hidden: (__VLS_ctx.unreadCount === 0),
        max: (99),
    }));
    const __VLS_34 = __VLS_33({
        value: (__VLS_ctx.unreadCount),
        hidden: (__VLS_ctx.unreadCount === 0),
        max: (99),
    }, ...__VLS_functionalComponentArgsRest(__VLS_33));
    __VLS_35.slots.default;
    const __VLS_36 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
        text: true,
    }));
    const __VLS_38 = __VLS_37({
        text: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_37));
    __VLS_39.slots.default;
    const __VLS_40 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
        size: (18),
    }));
    const __VLS_42 = __VLS_41({
        size: (18),
    }, ...__VLS_functionalComponentArgsRest(__VLS_41));
    __VLS_43.slots.default;
    const __VLS_44 = {}.Bell;
    /** @type {[typeof __VLS_components.Bell, ]} */ ;
    // @ts-ignore
    const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({}));
    const __VLS_46 = __VLS_45({}, ...__VLS_functionalComponentArgsRest(__VLS_45));
    var __VLS_43;
    var __VLS_39;
    var __VLS_35;
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
const __VLS_48 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
    ...{ 'onClick': {} },
    text: true,
    size: "small",
}));
const __VLS_50 = __VLS_49({
    ...{ 'onClick': {} },
    text: true,
    size: "small",
}, ...__VLS_functionalComponentArgsRest(__VLS_49));
let __VLS_52;
let __VLS_53;
let __VLS_54;
const __VLS_55 = {
    onClick: (__VLS_ctx.notifStore.markAllAsRead)
};
__VLS_51.slots.default;
var __VLS_51;
const __VLS_56 = {}.ElScrollbar;
/** @type {[typeof __VLS_components.ElScrollbar, typeof __VLS_components.elScrollbar, typeof __VLS_components.ElScrollbar, typeof __VLS_components.elScrollbar, ]} */ ;
// @ts-ignore
const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
    maxHeight: "400px",
}));
const __VLS_58 = __VLS_57({
    maxHeight: "400px",
}, ...__VLS_functionalComponentArgsRest(__VLS_57));
__VLS_59.slots.default;
if (__VLS_ctx.notifStore.recentNotifications.length === 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "empty-state" },
    });
    const __VLS_60 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
        description: "暂无通知",
        imageSize: (80),
    }));
    const __VLS_62 = __VLS_61({
        description: "暂无通知",
        imageSize: (80),
    }, ...__VLS_functionalComponentArgsRest(__VLS_61));
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "notification-list" },
    });
    for (const [n] of __VLS_getVForSourceType((__VLS_ctx.notifStore.recentNotifications))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ onClick: (...[$event]) => {
                    if (!!(__VLS_ctx.notifStore.recentNotifications.length === 0))
                        return;
                    __VLS_ctx.notifStore.markAsRead(n.id);
                } },
            key: (n.id),
            ...{ class: "notification-item" },
            ...{ class: ({ 'is-unread': !n.is_read }) },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "notification-content" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "notification-title" },
        });
        (n.title);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "notification-meta" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "notification-type" },
        });
        (n.type);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "notification-time" },
        });
        (new Date(n.created_at).toLocaleString('zh-CN'));
    }
}
var __VLS_59;
var __VLS_31;
const __VLS_64 = {}.ElDropdown;
/** @type {[typeof __VLS_components.ElDropdown, typeof __VLS_components.elDropdown, typeof __VLS_components.ElDropdown, typeof __VLS_components.elDropdown, ]} */ ;
// @ts-ignore
const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
    ...{ 'onCommand': {} },
    trigger: "click",
}));
const __VLS_66 = __VLS_65({
    ...{ 'onCommand': {} },
    trigger: "click",
}, ...__VLS_functionalComponentArgsRest(__VLS_65));
let __VLS_68;
let __VLS_69;
let __VLS_70;
const __VLS_71 = {
    onCommand: (__VLS_ctx.handleUserCommand)
};
__VLS_67.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "user-info" },
});
const __VLS_72 = {}.ElAvatar;
/** @type {[typeof __VLS_components.ElAvatar, typeof __VLS_components.elAvatar, typeof __VLS_components.ElAvatar, typeof __VLS_components.elAvatar, ]} */ ;
// @ts-ignore
const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
    size: (28),
    ...{ class: "user-avatar" },
}));
const __VLS_74 = __VLS_73({
    size: (28),
    ...{ class: "user-avatar" },
}, ...__VLS_functionalComponentArgsRest(__VLS_73));
__VLS_75.slots.default;
(__VLS_ctx.userInitial);
var __VLS_75;
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "user-name" },
});
(__VLS_ctx.userName);
const __VLS_76 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({}));
const __VLS_78 = __VLS_77({}, ...__VLS_functionalComponentArgsRest(__VLS_77));
__VLS_79.slots.default;
const __VLS_80 = {}.ArrowDown;
/** @type {[typeof __VLS_components.ArrowDown, ]} */ ;
// @ts-ignore
const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({}));
const __VLS_82 = __VLS_81({}, ...__VLS_functionalComponentArgsRest(__VLS_81));
var __VLS_79;
{
    const { dropdown: __VLS_thisSlot } = __VLS_67.slots;
    const __VLS_84 = {}.ElDropdownMenu;
    /** @type {[typeof __VLS_components.ElDropdownMenu, typeof __VLS_components.elDropdownMenu, typeof __VLS_components.ElDropdownMenu, typeof __VLS_components.elDropdownMenu, ]} */ ;
    // @ts-ignore
    const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({}));
    const __VLS_86 = __VLS_85({}, ...__VLS_functionalComponentArgsRest(__VLS_85));
    __VLS_87.slots.default;
    const __VLS_88 = {}.ElDropdownItem;
    /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
    // @ts-ignore
    const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
        disabled: true,
    }));
    const __VLS_90 = __VLS_89({
        disabled: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_89));
    __VLS_91.slots.default;
    const __VLS_92 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_93 = __VLS_asFunctionalComponent(__VLS_92, new __VLS_92({
        size: "small",
        type: __VLS_ctx.roleType,
    }));
    const __VLS_94 = __VLS_93({
        size: "small",
        type: __VLS_ctx.roleType,
    }, ...__VLS_functionalComponentArgsRest(__VLS_93));
    __VLS_95.slots.default;
    (__VLS_ctx.roleName);
    var __VLS_95;
    var __VLS_91;
    const __VLS_96 = {}.ElDropdownItem;
    /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
    // @ts-ignore
    const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
        divided: true,
        command: "logout",
    }));
    const __VLS_98 = __VLS_97({
        divided: true,
        command: "logout",
    }, ...__VLS_functionalComponentArgsRest(__VLS_97));
    __VLS_99.slots.default;
    const __VLS_100 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_101 = __VLS_asFunctionalComponent(__VLS_100, new __VLS_100({}));
    const __VLS_102 = __VLS_101({}, ...__VLS_functionalComponentArgsRest(__VLS_101));
    __VLS_103.slots.default;
    const __VLS_104 = {}.SwitchButton;
    /** @type {[typeof __VLS_components.SwitchButton, ]} */ ;
    // @ts-ignore
    const __VLS_105 = __VLS_asFunctionalComponent(__VLS_104, new __VLS_104({}));
    const __VLS_106 = __VLS_105({}, ...__VLS_functionalComponentArgsRest(__VLS_105));
    var __VLS_103;
    var __VLS_99;
    var __VLS_87;
}
var __VLS_67;
/** @type {__VLS_StyleScopedClasses['app-header']} */ ;
/** @type {__VLS_StyleScopedClasses['header-left']} */ ;
/** @type {__VLS_StyleScopedClasses['collapse-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['logo']} */ ;
/** @type {__VLS_StyleScopedClasses['logo-text']} */ ;
/** @type {__VLS_StyleScopedClasses['header-right']} */ ;
/** @type {__VLS_StyleScopedClasses['ws-status']} */ ;
/** @type {__VLS_StyleScopedClasses['status-dot']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-bell']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-popover']} */ ;
/** @type {__VLS_StyleScopedClasses['popover-header']} */ ;
/** @type {__VLS_StyleScopedClasses['popover-title']} */ ;
/** @type {__VLS_StyleScopedClasses['empty-state']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-list']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-item']} */ ;
/** @type {__VLS_StyleScopedClasses['is-unread']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-content']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-title']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-type']} */ ;
/** @type {__VLS_StyleScopedClasses['notification-time']} */ ;
/** @type {__VLS_StyleScopedClasses['user-info']} */ ;
/** @type {__VLS_StyleScopedClasses['user-avatar']} */ ;
/** @type {__VLS_StyleScopedClasses['user-name']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Fold: Fold,
            Expand: Expand,
            Monitor: Monitor,
            Bell: Bell,
            ArrowDown: ArrowDown,
            SwitchButton: SwitchButton,
            notifStore: notifStore,
            showNotifications: showNotifications,
            userInitial: userInitial,
            userName: userName,
            roleName: roleName,
            roleType: roleType,
            unreadCount: unreadCount,
            wsStatus: wsStatus,
            wsStatusText: wsStatusText,
            handleUserCommand: handleUserCommand,
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
