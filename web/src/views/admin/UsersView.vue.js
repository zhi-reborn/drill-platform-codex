import { ref, computed } from 'vue';
import { ElMessage } from 'element-plus';
import DataTable from '@/components/common/DataTable.vue';
import ActionConfirm from '@/components/common/ActionConfirm.vue';
import EmptyBox from '@/components/common/EmptyBox.vue';
import { userApi } from '@/api/modules/user';
import usersData from '@/mock/data/users.json';
const loading = ref(false);
const submitting = ref(false);
const showCreateDialog = ref(false);
const searchQuery = ref('');
const formRef = ref();
const users = ref(usersData);
// 表格列定义
const columns = [
    { prop: 'username', label: '用户名', width: 120 },
    { prop: 'name', label: '真实姓名', width: 100 },
    { prop: 'role', label: '角色', width: 100, slot: true },
    { prop: 'status', label: '状态', width: 80, slot: true },
    { prop: 'department', label: '部门' },
    { prop: 'phone', label: '手机号', width: 130 },
    { prop: 'last_login_at', label: '最后登录', width: 160, slot: true },
    { prop: 'actions', label: '操作', width: 120, slot: true },
];
// 创建用户表单
const createForm = ref({
    username: '',
    name: '',
    email: '',
    role: 'executor',
    phone: '',
    department: '',
    password: '',
});
// 表单验证规则
const formRules = {
    username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
    name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
    email: [
        { required: true, message: '请输入邮箱', trigger: 'blur' },
        { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
    ],
    role: [{ required: true, message: '请选择角色', trigger: 'change' }],
    password: [
        { required: true, message: '请输入密码', trigger: 'blur' },
        { min: 6, message: '密码长度至少 6 位', trigger: 'blur' },
    ],
};
// 过滤后的用户列表
const filteredUsers = computed(() => {
    if (!searchQuery.value)
        return users.value;
    const query = searchQuery.value.toLowerCase();
    return users.value.filter((user) => user.username.toLowerCase().includes(query) ||
        user.name.toLowerCase().includes(query) ||
        user.department.toLowerCase().includes(query));
});
// 角色标签类型
function getRoleTagType(role) {
    const map = {
        admin: 'danger',
        director: 'primary',
        executor: 'success',
        viewer: 'info',
    };
    return map[role] || 'info';
}
// 角色标签文本
function getRoleLabel(role) {
    const map = {
        admin: '管理员',
        director: '指挥长',
        executor: '执行员',
        viewer: '观察员',
    };
    return map[role] || role;
}
function formatTime(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
    });
}
function handleSearch() {
    // 搜索逻辑已在 computed 中处理
}
async function handleDisableUser(user) {
    try {
        await userApi.update(user.id, { status: 'disabled' });
        user.status = 'disabled';
        ElMessage.success('用户已禁用');
    }
    catch (error) {
        ElMessage.error('操作失败');
        console.error('Failed to disable user:', error);
    }
}
async function handleEnableUser(user) {
    try {
        await userApi.update(user.id, { status: 'active' });
        user.status = 'active';
        ElMessage.success('用户已启用');
    }
    catch (error) {
        ElMessage.error('操作失败');
        console.error('Failed to enable user:', error);
    }
}
async function handleCreateUser() {
    if (!formRef.value)
        return;
    try {
        await formRef.value.validate();
        submitting.value = true;
        await userApi.create({
            username: createForm.value.username,
            name: createForm.value.name,
            email: createForm.value.email,
            role: createForm.value.role,
            phone: createForm.value.phone || undefined,
            department: createForm.value.department || undefined,
            password: createForm.value.password,
        });
        ElMessage.success('用户创建成功');
        showCreateDialog.value = false;
        // 刷新列表（mock 模式下添加本地数据）
        users.value.unshift({
            id: Date.now(),
            ...createForm.value,
            status: 'active',
            last_login_at: '',
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
        });
    }
    catch (error) {
        if (error.message && !error.message.includes('validate')) {
            ElMessage.error('创建失败');
            console.error('Failed to create user:', error);
        }
    }
    finally {
        submitting.value = false;
    }
}
function resetForm() {
    formRef.value?.resetFields();
    createForm.value = {
        username: '',
        name: '',
        email: '',
        role: 'executor',
        phone: '',
        department: '',
        password: '',
    };
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['el-input__wrapper']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-container" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "header-actions" },
});
const __VLS_0 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onInput': {} },
    modelValue: (__VLS_ctx.searchQuery),
    placeholder: "搜索用户名/姓名/部门",
    ...{ class: "search-input" },
    clearable: true,
}));
const __VLS_2 = __VLS_1({
    ...{ 'onInput': {} },
    modelValue: (__VLS_ctx.searchQuery),
    placeholder: "搜索用户名/姓名/部门",
    ...{ class: "search-input" },
    clearable: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    onInput: (__VLS_ctx.handleSearch)
};
__VLS_3.slots.default;
{
    const { prefix: __VLS_thisSlot } = __VLS_3.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.svg, __VLS_intrinsicElements.svg)({
        viewBox: "0 0 24 24",
        fill: "none",
        stroke: "currentColor",
        'stroke-width': "2",
        width: "16",
        height: "16",
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.circle)({
        cx: "11",
        cy: "11",
        r: "8",
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
        d: "m21 21-4.35-4.35",
    });
}
var __VLS_3;
const __VLS_8 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    ...{ 'onClick': {} },
    type: "primary",
}));
const __VLS_10 = __VLS_9({
    ...{ 'onClick': {} },
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
let __VLS_12;
let __VLS_13;
let __VLS_14;
const __VLS_15 = {
    onClick: (...[$event]) => {
        __VLS_ctx.showCreateDialog = true;
    }
};
__VLS_11.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.svg, __VLS_intrinsicElements.svg)({
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    'stroke-width': "2",
    width: "16",
    height: "16",
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.path)({
    d: "M12 5v14M5 12h14",
});
var __VLS_11;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-content" },
});
/** @type {[typeof DataTable, typeof DataTable, ]} */ ;
// @ts-ignore
const __VLS_16 = __VLS_asFunctionalComponent(DataTable, new DataTable({
    columns: (__VLS_ctx.columns),
    data: (__VLS_ctx.filteredUsers),
    loading: (__VLS_ctx.loading),
    pagination: true,
    total: (__VLS_ctx.filteredUsers.length),
}));
const __VLS_17 = __VLS_16({
    columns: (__VLS_ctx.columns),
    data: (__VLS_ctx.filteredUsers),
    loading: (__VLS_ctx.loading),
    pagination: true,
    total: (__VLS_ctx.filteredUsers.length),
}, ...__VLS_functionalComponentArgsRest(__VLS_16));
__VLS_18.slots.default;
{
    const { role: __VLS_thisSlot } = __VLS_18.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_19 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_20 = __VLS_asFunctionalComponent(__VLS_19, new __VLS_19({
        type: (__VLS_ctx.getRoleTagType(row.role)),
        size: "small",
    }));
    const __VLS_21 = __VLS_20({
        type: (__VLS_ctx.getRoleTagType(row.role)),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_20));
    __VLS_22.slots.default;
    (__VLS_ctx.getRoleLabel(row.role));
    var __VLS_22;
}
{
    const { status: __VLS_thisSlot } = __VLS_18.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_23 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_24 = __VLS_asFunctionalComponent(__VLS_23, new __VLS_23({
        type: (row.status === 'active' ? 'success' : 'danger'),
        size: "small",
    }));
    const __VLS_25 = __VLS_24({
        type: (row.status === 'active' ? 'success' : 'danger'),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_24));
    __VLS_26.slots.default;
    (row.status === 'active' ? '正常' : '禁用');
    var __VLS_26;
}
{
    const { last_login_at: __VLS_thisSlot } = __VLS_18.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.last_login_at ? __VLS_ctx.formatTime(row.last_login_at) : '从未登录');
}
{
    const { actions: __VLS_thisSlot } = __VLS_18.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    if (row.status === 'active') {
        /** @type {[typeof ActionConfirm, typeof ActionConfirm, ]} */ ;
        // @ts-ignore
        const __VLS_27 = __VLS_asFunctionalComponent(ActionConfirm, new ActionConfirm({
            ...{ 'onConfirm': {} },
            message: "确认要禁用该用户吗？禁用后将无法登录系统。",
            danger: true,
            size: "small",
        }));
        const __VLS_28 = __VLS_27({
            ...{ 'onConfirm': {} },
            message: "确认要禁用该用户吗？禁用后将无法登录系统。",
            danger: true,
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_27));
        let __VLS_30;
        let __VLS_31;
        let __VLS_32;
        const __VLS_33 = {
            onConfirm: (...[$event]) => {
                if (!(row.status === 'active'))
                    return;
                __VLS_ctx.handleDisableUser(row);
            }
        };
        __VLS_29.slots.default;
        var __VLS_29;
    }
    else {
        const __VLS_34 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_35 = __VLS_asFunctionalComponent(__VLS_34, new __VLS_34({
            ...{ 'onClick': {} },
            type: "success",
            size: "small",
        }));
        const __VLS_36 = __VLS_35({
            ...{ 'onClick': {} },
            type: "success",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_35));
        let __VLS_38;
        let __VLS_39;
        let __VLS_40;
        const __VLS_41 = {
            onClick: (...[$event]) => {
                if (!!(row.status === 'active'))
                    return;
                __VLS_ctx.handleEnableUser(row);
            }
        };
        __VLS_37.slots.default;
        var __VLS_37;
    }
}
var __VLS_18;
if (__VLS_ctx.filteredUsers.length === 0 && !__VLS_ctx.loading) {
    /** @type {[typeof EmptyBox, ]} */ ;
    // @ts-ignore
    const __VLS_42 = __VLS_asFunctionalComponent(EmptyBox, new EmptyBox({
        title: "暂无用户数据",
        description: "尝试调整搜索条件或创建新用户",
    }));
    const __VLS_43 = __VLS_42({
        title: "暂无用户数据",
        description: "尝试调整搜索条件或创建新用户",
    }, ...__VLS_functionalComponentArgsRest(__VLS_42));
}
const __VLS_45 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_46 = __VLS_asFunctionalComponent(__VLS_45, new __VLS_45({
    ...{ 'onClose': {} },
    modelValue: (__VLS_ctx.showCreateDialog),
    title: "创建用户",
    width: "500px",
    closeOnClickModal: (false),
}));
const __VLS_47 = __VLS_46({
    ...{ 'onClose': {} },
    modelValue: (__VLS_ctx.showCreateDialog),
    title: "创建用户",
    width: "500px",
    closeOnClickModal: (false),
}, ...__VLS_functionalComponentArgsRest(__VLS_46));
let __VLS_49;
let __VLS_50;
let __VLS_51;
const __VLS_52 = {
    onClose: (__VLS_ctx.resetForm)
};
__VLS_48.slots.default;
const __VLS_53 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_54 = __VLS_asFunctionalComponent(__VLS_53, new __VLS_53({
    ref: "formRef",
    model: (__VLS_ctx.createForm),
    rules: (__VLS_ctx.formRules),
    labelWidth: "80px",
    labelPosition: "top",
}));
const __VLS_55 = __VLS_54({
    ref: "formRef",
    model: (__VLS_ctx.createForm),
    rules: (__VLS_ctx.formRules),
    labelWidth: "80px",
    labelPosition: "top",
}, ...__VLS_functionalComponentArgsRest(__VLS_54));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_57 = {};
__VLS_56.slots.default;
const __VLS_59 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
    label: "用户名",
    prop: "username",
}));
const __VLS_61 = __VLS_60({
    label: "用户名",
    prop: "username",
}, ...__VLS_functionalComponentArgsRest(__VLS_60));
__VLS_62.slots.default;
const __VLS_63 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_64 = __VLS_asFunctionalComponent(__VLS_63, new __VLS_63({
    modelValue: (__VLS_ctx.createForm.username),
    placeholder: "请输入用户名",
}));
const __VLS_65 = __VLS_64({
    modelValue: (__VLS_ctx.createForm.username),
    placeholder: "请输入用户名",
}, ...__VLS_functionalComponentArgsRest(__VLS_64));
var __VLS_62;
const __VLS_67 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({
    label: "姓名",
    prop: "name",
}));
const __VLS_69 = __VLS_68({
    label: "姓名",
    prop: "name",
}, ...__VLS_functionalComponentArgsRest(__VLS_68));
__VLS_70.slots.default;
const __VLS_71 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({
    modelValue: (__VLS_ctx.createForm.name),
    placeholder: "请输入真实姓名",
}));
const __VLS_73 = __VLS_72({
    modelValue: (__VLS_ctx.createForm.name),
    placeholder: "请输入真实姓名",
}, ...__VLS_functionalComponentArgsRest(__VLS_72));
var __VLS_70;
const __VLS_75 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_76 = __VLS_asFunctionalComponent(__VLS_75, new __VLS_75({
    label: "邮箱",
    prop: "email",
}));
const __VLS_77 = __VLS_76({
    label: "邮箱",
    prop: "email",
}, ...__VLS_functionalComponentArgsRest(__VLS_76));
__VLS_78.slots.default;
const __VLS_79 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_80 = __VLS_asFunctionalComponent(__VLS_79, new __VLS_79({
    modelValue: (__VLS_ctx.createForm.email),
    placeholder: "请输入邮箱",
}));
const __VLS_81 = __VLS_80({
    modelValue: (__VLS_ctx.createForm.email),
    placeholder: "请输入邮箱",
}, ...__VLS_functionalComponentArgsRest(__VLS_80));
var __VLS_78;
const __VLS_83 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_84 = __VLS_asFunctionalComponent(__VLS_83, new __VLS_83({
    label: "角色",
    prop: "role",
}));
const __VLS_85 = __VLS_84({
    label: "角色",
    prop: "role",
}, ...__VLS_functionalComponentArgsRest(__VLS_84));
__VLS_86.slots.default;
const __VLS_87 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_88 = __VLS_asFunctionalComponent(__VLS_87, new __VLS_87({
    modelValue: (__VLS_ctx.createForm.role),
    placeholder: "请选择角色",
    ...{ style: {} },
}));
const __VLS_89 = __VLS_88({
    modelValue: (__VLS_ctx.createForm.role),
    placeholder: "请选择角色",
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_88));
__VLS_90.slots.default;
const __VLS_91 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_92 = __VLS_asFunctionalComponent(__VLS_91, new __VLS_91({
    label: "系统管理员",
    value: "admin",
}));
const __VLS_93 = __VLS_92({
    label: "系统管理员",
    value: "admin",
}, ...__VLS_functionalComponentArgsRest(__VLS_92));
const __VLS_95 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_96 = __VLS_asFunctionalComponent(__VLS_95, new __VLS_95({
    label: "指挥长",
    value: "director",
}));
const __VLS_97 = __VLS_96({
    label: "指挥长",
    value: "director",
}, ...__VLS_functionalComponentArgsRest(__VLS_96));
const __VLS_99 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_100 = __VLS_asFunctionalComponent(__VLS_99, new __VLS_99({
    label: "执行员",
    value: "executor",
}));
const __VLS_101 = __VLS_100({
    label: "执行员",
    value: "executor",
}, ...__VLS_functionalComponentArgsRest(__VLS_100));
const __VLS_103 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_104 = __VLS_asFunctionalComponent(__VLS_103, new __VLS_103({
    label: "观察员",
    value: "viewer",
}));
const __VLS_105 = __VLS_104({
    label: "观察员",
    value: "viewer",
}, ...__VLS_functionalComponentArgsRest(__VLS_104));
var __VLS_90;
var __VLS_86;
const __VLS_107 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_108 = __VLS_asFunctionalComponent(__VLS_107, new __VLS_107({
    label: "手机号",
    prop: "phone",
}));
const __VLS_109 = __VLS_108({
    label: "手机号",
    prop: "phone",
}, ...__VLS_functionalComponentArgsRest(__VLS_108));
__VLS_110.slots.default;
const __VLS_111 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_112 = __VLS_asFunctionalComponent(__VLS_111, new __VLS_111({
    modelValue: (__VLS_ctx.createForm.phone),
    placeholder: "请输入手机号",
}));
const __VLS_113 = __VLS_112({
    modelValue: (__VLS_ctx.createForm.phone),
    placeholder: "请输入手机号",
}, ...__VLS_functionalComponentArgsRest(__VLS_112));
var __VLS_110;
const __VLS_115 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_116 = __VLS_asFunctionalComponent(__VLS_115, new __VLS_115({
    label: "部门",
    prop: "department",
}));
const __VLS_117 = __VLS_116({
    label: "部门",
    prop: "department",
}, ...__VLS_functionalComponentArgsRest(__VLS_116));
__VLS_118.slots.default;
const __VLS_119 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_120 = __VLS_asFunctionalComponent(__VLS_119, new __VLS_119({
    modelValue: (__VLS_ctx.createForm.department),
    placeholder: "请输入部门",
}));
const __VLS_121 = __VLS_120({
    modelValue: (__VLS_ctx.createForm.department),
    placeholder: "请输入部门",
}, ...__VLS_functionalComponentArgsRest(__VLS_120));
var __VLS_118;
const __VLS_123 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_124 = __VLS_asFunctionalComponent(__VLS_123, new __VLS_123({
    label: "密码",
    prop: "password",
}));
const __VLS_125 = __VLS_124({
    label: "密码",
    prop: "password",
}, ...__VLS_functionalComponentArgsRest(__VLS_124));
__VLS_126.slots.default;
const __VLS_127 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_128 = __VLS_asFunctionalComponent(__VLS_127, new __VLS_127({
    modelValue: (__VLS_ctx.createForm.password),
    type: "password",
    placeholder: "请输入初始密码",
    showPassword: true,
}));
const __VLS_129 = __VLS_128({
    modelValue: (__VLS_ctx.createForm.password),
    type: "password",
    placeholder: "请输入初始密码",
    showPassword: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_128));
var __VLS_126;
var __VLS_56;
{
    const { footer: __VLS_thisSlot } = __VLS_48.slots;
    const __VLS_131 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_132 = __VLS_asFunctionalComponent(__VLS_131, new __VLS_131({
        ...{ 'onClick': {} },
    }));
    const __VLS_133 = __VLS_132({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_132));
    let __VLS_135;
    let __VLS_136;
    let __VLS_137;
    const __VLS_138 = {
        onClick: (...[$event]) => {
            __VLS_ctx.showCreateDialog = false;
        }
    };
    __VLS_134.slots.default;
    var __VLS_134;
    const __VLS_139 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_140 = __VLS_asFunctionalComponent(__VLS_139, new __VLS_139({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.submitting),
    }));
    const __VLS_141 = __VLS_140({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.submitting),
    }, ...__VLS_functionalComponentArgsRest(__VLS_140));
    let __VLS_143;
    let __VLS_144;
    let __VLS_145;
    const __VLS_146 = {
        onClick: (__VLS_ctx.handleCreateUser)
    };
    __VLS_142.slots.default;
    var __VLS_142;
}
var __VLS_48;
/** @type {__VLS_StyleScopedClasses['page-container']} */ ;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
/** @type {__VLS_StyleScopedClasses['header-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['search-input']} */ ;
/** @type {__VLS_StyleScopedClasses['page-content']} */ ;
// @ts-ignore
var __VLS_58 = __VLS_57;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            DataTable: DataTable,
            ActionConfirm: ActionConfirm,
            EmptyBox: EmptyBox,
            loading: loading,
            submitting: submitting,
            showCreateDialog: showCreateDialog,
            searchQuery: searchQuery,
            formRef: formRef,
            columns: columns,
            createForm: createForm,
            formRules: formRules,
            filteredUsers: filteredUsers,
            getRoleTagType: getRoleTagType,
            getRoleLabel: getRoleLabel,
            formatTime: formatTime,
            handleSearch: handleSearch,
            handleDisableUser: handleDisableUser,
            handleEnableUser: handleEnableUser,
            handleCreateUser: handleCreateUser,
            resetForm: resetForm,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
