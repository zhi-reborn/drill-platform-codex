import { ref, reactive, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { User, Lock, Connection, CircleCheck, Monitor } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { useAuthStore } from '@/stores/auth';
import usersData from '@/mock/data/users.json';
const router = useRouter();
const authStore = useAuthStore();
const authMode = ref(import.meta.env.VITE_AUTH_MODE || 'dev');
const loading = ref(false);
const error = ref('');
const remember = ref(false);
const formRef = ref();
// Dev mode
const selectedUser = ref(0);
const activeUsers = usersData.filter((u) => u.status === 'active');
// Local mode
const form = reactive({ username: '', password: '' });
const rules = {
    username: [{ required: true, message: '请输入用户名', trigger: 'blur' }, { min: 3, max: 50, message: '3-50 个字符', trigger: 'blur' }],
    password: [{ required: true, message: '请输入密码', trigger: 'blur' }, { min: 6, message: '至少 6 个字符', trigger: 'blur' }],
};
const canSubmit = computed(() => form.username.length >= 3 && form.password.length >= 6);
function roleLabel(role) {
    const map = { admin: '管理员', director: '指挥员', executor: '执行者', viewer: '观察者' };
    return map[role] || role;
}
function roleTagType(role) {
    const map = { admin: 'danger', director: 'warning', executor: 'success', viewer: 'info' };
    return map[role];
}
const roleDashboards = {
    admin: '/admin',
    director: '/director',
    executor: '/executor',
    viewer: '/viewer',
};
async function handleDevLogin() {
    if (!selectedUser.value) {
        ElMessage.warning('请选择一个用户');
        return;
    }
    loading.value = true;
    error.value = '';
    try {
        const user = activeUsers.find(u => u.id === selectedUser.value);
        await authStore.loginWithUser(user);
        ElMessage.success(`欢迎回来，${user.name}`);
        router.push(roleDashboards[user.role]);
    }
    catch (e) {
        error.value = e instanceof Error ? e.message : '登录失败';
    }
    finally {
        loading.value = false;
    }
}
async function handleLocalLogin() {
    loading.value = true;
    error.value = '';
    try {
        await authStore.loginWithCredentials(form);
        ElMessage.success('登录成功');
        const user = authStore.user;
        router.push(user ? roleDashboards[user.role] : '/viewer');
    }
    catch (e) {
        error.value = '用户名或密码错误';
    }
    finally {
        loading.value = false;
    }
}
async function handleCasLogin() {
    window.location.href = '/api/v1/auth/cas?redirect=' + encodeURIComponent(window.location.origin + '/cas/callback');
}
// Restore session on mount
onMounted(() => {
    authStore.restoreSession();
    if (authStore.isAuthenticated && authStore.user) {
        router.push(roleDashboards[authStore.user.role]);
    }
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "login-page" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "login-brand" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "brand-content" },
});
const __VLS_0 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    size: (64),
    color: "#55C3D3",
}));
const __VLS_2 = __VLS_1({
    size: (64),
    color: "#55C3D3",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
const __VLS_4 = {}.Monitor;
/** @type {[typeof __VLS_components.Monitor, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({}));
const __VLS_6 = __VLS_5({}, ...__VLS_functionalComponentArgsRest(__VLS_5));
var __VLS_3;
__VLS_asFunctionalElement(__VLS_intrinsicElements.h1, __VLS_intrinsicElements.h1)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "brand-tagline" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "brand-desc" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "brand-features" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "feature" },
});
const __VLS_8 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({}));
const __VLS_10 = __VLS_9({}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
const __VLS_12 = {}.CircleCheck;
/** @type {[typeof __VLS_components.CircleCheck, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({}));
const __VLS_14 = __VLS_13({}, ...__VLS_functionalComponentArgsRest(__VLS_13));
var __VLS_11;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "feature" },
});
const __VLS_16 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({}));
const __VLS_18 = __VLS_17({}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
const __VLS_20 = {}.CircleCheck;
/** @type {[typeof __VLS_components.CircleCheck, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({}));
const __VLS_22 = __VLS_21({}, ...__VLS_functionalComponentArgsRest(__VLS_21));
var __VLS_19;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "feature" },
});
const __VLS_24 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({}));
const __VLS_26 = __VLS_25({}, ...__VLS_functionalComponentArgsRest(__VLS_25));
__VLS_27.slots.default;
const __VLS_28 = {}.CircleCheck;
/** @type {[typeof __VLS_components.CircleCheck, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({}));
const __VLS_30 = __VLS_29({}, ...__VLS_functionalComponentArgsRest(__VLS_29));
var __VLS_27;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "brand-bg-pattern" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "login-area" },
});
if (__VLS_ctx.authMode === 'dev') {
    const __VLS_32 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
        ...{ class: "dev-badge" },
        type: "success",
        effect: "dark",
        size: "small",
    }));
    const __VLS_34 = __VLS_33({
        ...{ class: "dev-badge" },
        type: "success",
        effect: "dark",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_33));
    __VLS_35.slots.default;
    var __VLS_35;
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "login-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({
    ...{ class: "login-title" },
});
if (__VLS_ctx.authMode === 'cas') {
}
else if (__VLS_ctx.authMode === 'dev') {
}
else {
}
if (__VLS_ctx.authMode === 'cas') {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "cas-mode" },
    });
    const __VLS_36 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
        ...{ 'onClick': {} },
        type: "primary",
        size: "large",
        ...{ class: "cas-btn" },
        loading: (__VLS_ctx.loading),
    }));
    const __VLS_38 = __VLS_37({
        ...{ 'onClick': {} },
        type: "primary",
        size: "large",
        ...{ class: "cas-btn" },
        loading: (__VLS_ctx.loading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_37));
    let __VLS_40;
    let __VLS_41;
    let __VLS_42;
    const __VLS_43 = {
        onClick: (__VLS_ctx.handleCasLogin)
    };
    __VLS_39.slots.default;
    const __VLS_44 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
        size: (20),
    }));
    const __VLS_46 = __VLS_45({
        size: (20),
    }, ...__VLS_functionalComponentArgsRest(__VLS_45));
    __VLS_47.slots.default;
    const __VLS_48 = {}.Connection;
    /** @type {[typeof __VLS_components.Connection, ]} */ ;
    // @ts-ignore
    const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({}));
    const __VLS_50 = __VLS_49({}, ...__VLS_functionalComponentArgsRest(__VLS_49));
    var __VLS_47;
    var __VLS_39;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "hint" },
    });
}
else if (__VLS_ctx.authMode === 'dev') {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "dev-mode" },
    });
    const __VLS_52 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({
        modelValue: (__VLS_ctx.selectedUser),
        filterable: true,
        placeholder: "选择登录用户",
        size: "large",
        ...{ class: "user-select" },
    }));
    const __VLS_54 = __VLS_53({
        modelValue: (__VLS_ctx.selectedUser),
        filterable: true,
        placeholder: "选择登录用户",
        size: "large",
        ...{ class: "user-select" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_53));
    __VLS_55.slots.default;
    for (const [u] of __VLS_getVForSourceType((__VLS_ctx.activeUsers))) {
        const __VLS_56 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
            key: (u.id),
            label: (u.id),
            value: (u.id),
        }));
        const __VLS_58 = __VLS_57({
            key: (u.id),
            label: (u.id),
            value: (u.id),
        }, ...__VLS_functionalComponentArgsRest(__VLS_57));
        __VLS_59.slots.default;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "user-option" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
        (u.name);
        (u.username);
        const __VLS_60 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
            size: ('small'),
            type: __VLS_ctx.roleTagType(u.role),
        }));
        const __VLS_62 = __VLS_61({
            size: ('small'),
            type: __VLS_ctx.roleTagType(u.role),
        }, ...__VLS_functionalComponentArgsRest(__VLS_61));
        __VLS_63.slots.default;
        (__VLS_ctx.roleLabel(u.role));
        var __VLS_63;
        var __VLS_59;
    }
    var __VLS_55;
    const __VLS_64 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
        ...{ 'onClick': {} },
        type: "primary",
        size: "large",
        ...{ class: "login-btn" },
        loading: (__VLS_ctx.loading),
    }));
    const __VLS_66 = __VLS_65({
        ...{ 'onClick': {} },
        type: "primary",
        size: "large",
        ...{ class: "login-btn" },
        loading: (__VLS_ctx.loading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_65));
    let __VLS_68;
    let __VLS_69;
    let __VLS_70;
    const __VLS_71 = {
        onClick: (__VLS_ctx.handleDevLogin)
    };
    __VLS_67.slots.default;
    var __VLS_67;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ onClick: (...[$event]) => {
                if (!!(__VLS_ctx.authMode === 'cas'))
                    return;
                if (!(__VLS_ctx.authMode === 'dev'))
                    return;
                __VLS_ctx.authMode = 'local';
                __VLS_ctx.selectedUser = 0;
            } },
        ...{ class: "link-btn" },
    });
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "local-mode" },
    });
    const __VLS_72 = {}.ElForm;
    /** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
    // @ts-ignore
    const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
        ...{ 'onSubmit': {} },
        ref: "formRef",
        model: (__VLS_ctx.form),
        rules: (__VLS_ctx.rules),
    }));
    const __VLS_74 = __VLS_73({
        ...{ 'onSubmit': {} },
        ref: "formRef",
        model: (__VLS_ctx.form),
        rules: (__VLS_ctx.rules),
    }, ...__VLS_functionalComponentArgsRest(__VLS_73));
    let __VLS_76;
    let __VLS_77;
    let __VLS_78;
    const __VLS_79 = {
        onSubmit: (__VLS_ctx.handleLocalLogin)
    };
    /** @type {typeof __VLS_ctx.formRef} */ ;
    var __VLS_80 = {};
    __VLS_75.slots.default;
    const __VLS_82 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_83 = __VLS_asFunctionalComponent(__VLS_82, new __VLS_82({
        prop: "username",
    }));
    const __VLS_84 = __VLS_83({
        prop: "username",
    }, ...__VLS_functionalComponentArgsRest(__VLS_83));
    __VLS_85.slots.default;
    const __VLS_86 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_87 = __VLS_asFunctionalComponent(__VLS_86, new __VLS_86({
        modelValue: (__VLS_ctx.form.username),
        placeholder: "用户名",
        size: "large",
        prefixIcon: (__VLS_ctx.User),
        autocomplete: "username",
    }));
    const __VLS_88 = __VLS_87({
        modelValue: (__VLS_ctx.form.username),
        placeholder: "用户名",
        size: "large",
        prefixIcon: (__VLS_ctx.User),
        autocomplete: "username",
    }, ...__VLS_functionalComponentArgsRest(__VLS_87));
    var __VLS_85;
    const __VLS_90 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_91 = __VLS_asFunctionalComponent(__VLS_90, new __VLS_90({
        prop: "password",
    }));
    const __VLS_92 = __VLS_91({
        prop: "password",
    }, ...__VLS_functionalComponentArgsRest(__VLS_91));
    __VLS_93.slots.default;
    const __VLS_94 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_95 = __VLS_asFunctionalComponent(__VLS_94, new __VLS_94({
        ...{ 'onKeyup': {} },
        modelValue: (__VLS_ctx.form.password),
        type: "password",
        placeholder: "密码",
        size: "large",
        prefixIcon: (__VLS_ctx.Lock),
        showPassword: true,
        autocomplete: "current-password",
    }));
    const __VLS_96 = __VLS_95({
        ...{ 'onKeyup': {} },
        modelValue: (__VLS_ctx.form.password),
        type: "password",
        placeholder: "密码",
        size: "large",
        prefixIcon: (__VLS_ctx.Lock),
        showPassword: true,
        autocomplete: "current-password",
    }, ...__VLS_functionalComponentArgsRest(__VLS_95));
    let __VLS_98;
    let __VLS_99;
    let __VLS_100;
    const __VLS_101 = {
        onKeyup: (__VLS_ctx.handleLocalLogin)
    };
    var __VLS_97;
    var __VLS_93;
    const __VLS_102 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_103 = __VLS_asFunctionalComponent(__VLS_102, new __VLS_102({}));
    const __VLS_104 = __VLS_103({}, ...__VLS_functionalComponentArgsRest(__VLS_103));
    __VLS_105.slots.default;
    const __VLS_106 = {}.ElCheckbox;
    /** @type {[typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, typeof __VLS_components.ElCheckbox, typeof __VLS_components.elCheckbox, ]} */ ;
    // @ts-ignore
    const __VLS_107 = __VLS_asFunctionalComponent(__VLS_106, new __VLS_106({
        modelValue: (__VLS_ctx.remember),
    }));
    const __VLS_108 = __VLS_107({
        modelValue: (__VLS_ctx.remember),
    }, ...__VLS_functionalComponentArgsRest(__VLS_107));
    __VLS_109.slots.default;
    var __VLS_109;
    var __VLS_105;
    const __VLS_110 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_111 = __VLS_asFunctionalComponent(__VLS_110, new __VLS_110({}));
    const __VLS_112 = __VLS_111({}, ...__VLS_functionalComponentArgsRest(__VLS_111));
    __VLS_113.slots.default;
    const __VLS_114 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_115 = __VLS_asFunctionalComponent(__VLS_114, new __VLS_114({
        ...{ 'onClick': {} },
        type: "primary",
        size: "large",
        ...{ class: "login-btn" },
        loading: (__VLS_ctx.loading),
        disabled: (!__VLS_ctx.canSubmit),
    }));
    const __VLS_116 = __VLS_115({
        ...{ 'onClick': {} },
        type: "primary",
        size: "large",
        ...{ class: "login-btn" },
        loading: (__VLS_ctx.loading),
        disabled: (!__VLS_ctx.canSubmit),
    }, ...__VLS_functionalComponentArgsRest(__VLS_115));
    let __VLS_118;
    let __VLS_119;
    let __VLS_120;
    const __VLS_121 = {
        onClick: (__VLS_ctx.handleLocalLogin)
    };
    __VLS_117.slots.default;
    var __VLS_117;
    var __VLS_113;
    var __VLS_75;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ onClick: (...[$event]) => {
                if (!!(__VLS_ctx.authMode === 'cas'))
                    return;
                if (!!(__VLS_ctx.authMode === 'dev'))
                    return;
                __VLS_ctx.authMode = 'dev';
                __VLS_ctx.selectedUser = 0;
            } },
        ...{ class: "link-btn" },
    });
}
if (__VLS_ctx.error) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "login-error" },
    });
    const __VLS_122 = {}.ElAlert;
    /** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
    // @ts-ignore
    const __VLS_123 = __VLS_asFunctionalComponent(__VLS_122, new __VLS_122({
        type: "error",
        title: (__VLS_ctx.error),
        showIcon: true,
        closable: (false),
    }));
    const __VLS_124 = __VLS_123({
        type: "error",
        title: (__VLS_ctx.error),
        showIcon: true,
        closable: (false),
    }, ...__VLS_functionalComponentArgsRest(__VLS_123));
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "copyright" },
});
/** @type {__VLS_StyleScopedClasses['login-page']} */ ;
/** @type {__VLS_StyleScopedClasses['login-brand']} */ ;
/** @type {__VLS_StyleScopedClasses['brand-content']} */ ;
/** @type {__VLS_StyleScopedClasses['brand-tagline']} */ ;
/** @type {__VLS_StyleScopedClasses['brand-desc']} */ ;
/** @type {__VLS_StyleScopedClasses['brand-features']} */ ;
/** @type {__VLS_StyleScopedClasses['feature']} */ ;
/** @type {__VLS_StyleScopedClasses['feature']} */ ;
/** @type {__VLS_StyleScopedClasses['feature']} */ ;
/** @type {__VLS_StyleScopedClasses['brand-bg-pattern']} */ ;
/** @type {__VLS_StyleScopedClasses['login-area']} */ ;
/** @type {__VLS_StyleScopedClasses['dev-badge']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card']} */ ;
/** @type {__VLS_StyleScopedClasses['login-title']} */ ;
/** @type {__VLS_StyleScopedClasses['cas-mode']} */ ;
/** @type {__VLS_StyleScopedClasses['cas-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['hint']} */ ;
/** @type {__VLS_StyleScopedClasses['dev-mode']} */ ;
/** @type {__VLS_StyleScopedClasses['user-select']} */ ;
/** @type {__VLS_StyleScopedClasses['user-option']} */ ;
/** @type {__VLS_StyleScopedClasses['login-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['link-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['local-mode']} */ ;
/** @type {__VLS_StyleScopedClasses['login-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['link-btn']} */ ;
/** @type {__VLS_StyleScopedClasses['login-error']} */ ;
/** @type {__VLS_StyleScopedClasses['copyright']} */ ;
// @ts-ignore
var __VLS_81 = __VLS_80;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            User: User,
            Lock: Lock,
            Connection: Connection,
            CircleCheck: CircleCheck,
            Monitor: Monitor,
            authMode: authMode,
            loading: loading,
            error: error,
            remember: remember,
            formRef: formRef,
            selectedUser: selectedUser,
            activeUsers: activeUsers,
            form: form,
            rules: rules,
            canSubmit: canSubmit,
            roleLabel: roleLabel,
            roleTagType: roleTagType,
            handleDevLogin: handleDevLogin,
            handleLocalLogin: handleLocalLogin,
            handleCasLogin: handleCasLogin,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
