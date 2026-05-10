import { ref, computed } from 'vue';
import { useRoute } from 'vue-router';
import AppHeader from './AppHeader.vue';
import AppSidebar from './AppSidebar.vue';
const collapsed = ref(false);
const route = useRoute();
const breadcrumbs = computed(() => {
    return route.matched
        .filter(r => r.meta?.title)
        .map(r => ({ path: r.path, title: r.meta.title }));
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['el-breadcrumb__inner']} */ ;
/** @type {__VLS_StyleScopedClasses['app-layout']} */ ;
/** @type {__VLS_StyleScopedClasses['app-main']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "app-layout" },
});
/** @type {[typeof AppHeader, ]} */ ;
// @ts-ignore
const __VLS_0 = __VLS_asFunctionalComponent(AppHeader, new AppHeader({
    ...{ 'onToggleSidebar': {} },
}));
const __VLS_1 = __VLS_0({
    ...{ 'onToggleSidebar': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_0));
let __VLS_3;
let __VLS_4;
let __VLS_5;
const __VLS_6 = {
    onToggleSidebar: (...[$event]) => {
        __VLS_ctx.collapsed = !__VLS_ctx.collapsed;
    }
};
var __VLS_2;
/** @type {[typeof AppSidebar, ]} */ ;
// @ts-ignore
const __VLS_7 = __VLS_asFunctionalComponent(AppSidebar, new AppSidebar({
    collapsed: (__VLS_ctx.collapsed),
}));
const __VLS_8 = __VLS_7({
    collapsed: (__VLS_ctx.collapsed),
}, ...__VLS_functionalComponentArgsRest(__VLS_7));
__VLS_asFunctionalElement(__VLS_intrinsicElements.main, __VLS_intrinsicElements.main)({
    ...{ class: "app-main" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "app-breadcrumb" },
});
const __VLS_10 = {}.ElBreadcrumb;
/** @type {[typeof __VLS_components.ElBreadcrumb, typeof __VLS_components.elBreadcrumb, typeof __VLS_components.ElBreadcrumb, typeof __VLS_components.elBreadcrumb, ]} */ ;
// @ts-ignore
const __VLS_11 = __VLS_asFunctionalComponent(__VLS_10, new __VLS_10({
    separator: "/",
}));
const __VLS_12 = __VLS_11({
    separator: "/",
}, ...__VLS_functionalComponentArgsRest(__VLS_11));
__VLS_13.slots.default;
for (const [item] of __VLS_getVForSourceType((__VLS_ctx.breadcrumbs))) {
    const __VLS_14 = {}.ElBreadcrumbItem;
    /** @type {[typeof __VLS_components.ElBreadcrumbItem, typeof __VLS_components.elBreadcrumbItem, typeof __VLS_components.ElBreadcrumbItem, typeof __VLS_components.elBreadcrumbItem, ]} */ ;
    // @ts-ignore
    const __VLS_15 = __VLS_asFunctionalComponent(__VLS_14, new __VLS_14({
        key: (item.path),
    }));
    const __VLS_16 = __VLS_15({
        key: (item.path),
    }, ...__VLS_functionalComponentArgsRest(__VLS_15));
    __VLS_17.slots.default;
    (item.title);
    var __VLS_17;
}
var __VLS_13;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "app-content" },
});
const __VLS_18 = {}.RouterView;
/** @type {[typeof __VLS_components.RouterView, typeof __VLS_components.routerView, ]} */ ;
// @ts-ignore
const __VLS_19 = __VLS_asFunctionalComponent(__VLS_18, new __VLS_18({}));
const __VLS_20 = __VLS_19({}, ...__VLS_functionalComponentArgsRest(__VLS_19));
/** @type {__VLS_StyleScopedClasses['app-layout']} */ ;
/** @type {__VLS_StyleScopedClasses['app-main']} */ ;
/** @type {__VLS_StyleScopedClasses['app-breadcrumb']} */ ;
/** @type {__VLS_StyleScopedClasses['app-content']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            AppHeader: AppHeader,
            AppSidebar: AppSidebar,
            collapsed: collapsed,
            breadcrumbs: breadcrumbs,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
