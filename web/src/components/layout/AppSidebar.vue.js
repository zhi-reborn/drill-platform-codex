import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
const __VLS_props = defineProps();
const route = useRoute();
const authStore = useAuthStore();
const activeMenu = computed(() => {
    const path = route.path;
    if (path.startsWith('/director/monitor/'))
        return path;
    if (path.startsWith('/executor/tasks/'))
        return path;
    return path;
});
const menuConfig = {
    admin: [
        { path: '/admin', title: '系统概览', icon: 'DataAnalysis' },
        { path: '/admin/users', title: '用户管理', icon: 'User' },
        { path: '/admin/drills', title: '全部演练', icon: 'Monitor' },
    ],
    director: [
        { path: '/director', title: '指挥概览', icon: 'DataAnalysis' },
        { path: '/director/templates', title: '模板管理', icon: 'Document' },
        { path: '/director/create', title: '创建演练', icon: 'Plus' },
    ],
    executor: [
        { path: '/executor', title: '我的任务', icon: 'Tickets' },
    ],
    viewer: [
        { path: '/viewer', title: '演练概览', icon: 'View' },
    ],
};
const visibleMenus = computed(() => {
    const role = authStore.role;
    return menuConfig[role] ?? menuConfig.viewer;
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['el-sub-menu__title']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.aside, __VLS_intrinsicElements.aside)({
    ...{ class: "app-sidebar" },
    ...{ class: ({ 'is-collapsed': __VLS_ctx.collapsed }) },
});
const __VLS_0 = {}.ElMenu;
/** @type {[typeof __VLS_components.ElMenu, typeof __VLS_components.elMenu, typeof __VLS_components.ElMenu, typeof __VLS_components.elMenu, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    defaultActive: (__VLS_ctx.activeMenu),
    collapse: (__VLS_ctx.collapsed),
    collapseTransition: (true),
    router: true,
    ...{ class: "sidebar-menu" },
}));
const __VLS_2 = __VLS_1({
    defaultActive: (__VLS_ctx.activeMenu),
    collapse: (__VLS_ctx.collapsed),
    collapseTransition: (true),
    router: true,
    ...{ class: "sidebar-menu" },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
for (const [menu] of __VLS_getVForSourceType((__VLS_ctx.visibleMenus))) {
    (menu.path);
    if (!menu.children) {
        const __VLS_4 = {}.ElMenuItem;
        /** @type {[typeof __VLS_components.ElMenuItem, typeof __VLS_components.elMenuItem, typeof __VLS_components.ElMenuItem, typeof __VLS_components.elMenuItem, ]} */ ;
        // @ts-ignore
        const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
            index: (menu.path),
        }));
        const __VLS_6 = __VLS_5({
            index: (menu.path),
        }, ...__VLS_functionalComponentArgsRest(__VLS_5));
        __VLS_7.slots.default;
        const __VLS_8 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({}));
        const __VLS_10 = __VLS_9({}, ...__VLS_functionalComponentArgsRest(__VLS_9));
        __VLS_11.slots.default;
        const __VLS_12 = ((menu.icon));
        // @ts-ignore
        const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({}));
        const __VLS_14 = __VLS_13({}, ...__VLS_functionalComponentArgsRest(__VLS_13));
        var __VLS_11;
        {
            const { title: __VLS_thisSlot } = __VLS_7.slots;
            (menu.title);
        }
        var __VLS_7;
    }
    else {
        const __VLS_16 = {}.ElSubMenu;
        /** @type {[typeof __VLS_components.ElSubMenu, typeof __VLS_components.elSubMenu, typeof __VLS_components.ElSubMenu, typeof __VLS_components.elSubMenu, ]} */ ;
        // @ts-ignore
        const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
            index: (menu.path),
        }));
        const __VLS_18 = __VLS_17({
            index: (menu.path),
        }, ...__VLS_functionalComponentArgsRest(__VLS_17));
        __VLS_19.slots.default;
        {
            const { title: __VLS_thisSlot } = __VLS_19.slots;
            const __VLS_20 = {}.ElIcon;
            /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
            // @ts-ignore
            const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({}));
            const __VLS_22 = __VLS_21({}, ...__VLS_functionalComponentArgsRest(__VLS_21));
            __VLS_23.slots.default;
            const __VLS_24 = ((menu.icon));
            // @ts-ignore
            const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({}));
            const __VLS_26 = __VLS_25({}, ...__VLS_functionalComponentArgsRest(__VLS_25));
            var __VLS_23;
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
            (menu.title);
        }
        for (const [child] of __VLS_getVForSourceType((menu.children))) {
            const __VLS_28 = {}.ElMenuItem;
            /** @type {[typeof __VLS_components.ElMenuItem, typeof __VLS_components.elMenuItem, typeof __VLS_components.ElMenuItem, typeof __VLS_components.elMenuItem, ]} */ ;
            // @ts-ignore
            const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
                key: (child.path),
                index: (child.path),
            }));
            const __VLS_30 = __VLS_29({
                key: (child.path),
                index: (child.path),
            }, ...__VLS_functionalComponentArgsRest(__VLS_29));
            __VLS_31.slots.default;
            const __VLS_32 = {}.ElIcon;
            /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
            // @ts-ignore
            const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({}));
            const __VLS_34 = __VLS_33({}, ...__VLS_functionalComponentArgsRest(__VLS_33));
            __VLS_35.slots.default;
            const __VLS_36 = ((child.icon));
            // @ts-ignore
            const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({}));
            const __VLS_38 = __VLS_37({}, ...__VLS_functionalComponentArgsRest(__VLS_37));
            var __VLS_35;
            {
                const { title: __VLS_thisSlot } = __VLS_31.slots;
                (child.title);
            }
            var __VLS_31;
        }
        var __VLS_19;
    }
}
var __VLS_3;
/** @type {__VLS_StyleScopedClasses['app-sidebar']} */ ;
/** @type {__VLS_StyleScopedClasses['is-collapsed']} */ ;
/** @type {__VLS_StyleScopedClasses['sidebar-menu']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            activeMenu: activeMenu,
            visibleMenus: visibleMenus,
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
