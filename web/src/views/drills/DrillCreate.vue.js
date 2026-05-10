import { ref, reactive } from 'vue';
const formRef = ref();
const form = reactive({
    title: '',
    category: '',
    description: ''
});
const rules = {
    title: [
        { required: true, message: '请输入演练标题', trigger: 'blur' },
        { min: 2, max: 100, message: '长度在 2 到 100 个字符', trigger: 'blur' }
    ],
    category: [
        { required: true, message: '请选择演练类型', trigger: 'change' }
    ]
};
function handleSubmit() {
    formRef.value?.validate((valid) => {
        if (valid) {
            console.log('Submit:', form);
        }
    });
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "drill-create-page" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h1, __VLS_intrinsicElements.h1)({
    ...{ class: "page-title" },
});
const __VLS_0 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({}));
const __VLS_2 = __VLS_1({}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
const __VLS_4 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
    ref: "formRef",
    model: (__VLS_ctx.form),
    rules: (__VLS_ctx.rules),
    labelWidth: "100px",
    labelPosition: "top",
}));
const __VLS_6 = __VLS_5({
    ref: "formRef",
    model: (__VLS_ctx.form),
    rules: (__VLS_ctx.rules),
    labelWidth: "100px",
    labelPosition: "top",
}, ...__VLS_functionalComponentArgsRest(__VLS_5));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_8 = {};
__VLS_7.slots.default;
const __VLS_10 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_11 = __VLS_asFunctionalComponent(__VLS_10, new __VLS_10({
    label: "演练标题",
    prop: "title",
}));
const __VLS_12 = __VLS_11({
    label: "演练标题",
    prop: "title",
}, ...__VLS_functionalComponentArgsRest(__VLS_11));
__VLS_13.slots.default;
const __VLS_14 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_15 = __VLS_asFunctionalComponent(__VLS_14, new __VLS_14({
    modelValue: (__VLS_ctx.form.title),
    placeholder: "请输入演练标题",
}));
const __VLS_16 = __VLS_15({
    modelValue: (__VLS_ctx.form.title),
    placeholder: "请输入演练标题",
}, ...__VLS_functionalComponentArgsRest(__VLS_15));
var __VLS_13;
const __VLS_18 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_19 = __VLS_asFunctionalComponent(__VLS_18, new __VLS_18({
    label: "演练类型",
    prop: "category",
}));
const __VLS_20 = __VLS_19({
    label: "演练类型",
    prop: "category",
}, ...__VLS_functionalComponentArgsRest(__VLS_19));
__VLS_21.slots.default;
const __VLS_22 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_23 = __VLS_asFunctionalComponent(__VLS_22, new __VLS_22({
    modelValue: (__VLS_ctx.form.category),
    placeholder: "请选择演练类型",
}));
const __VLS_24 = __VLS_23({
    modelValue: (__VLS_ctx.form.category),
    placeholder: "请选择演练类型",
}, ...__VLS_functionalComponentArgsRest(__VLS_23));
__VLS_25.slots.default;
const __VLS_26 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_27 = __VLS_asFunctionalComponent(__VLS_26, new __VLS_26({
    label: "灾备切换",
    value: "disaster_recovery",
}));
const __VLS_28 = __VLS_27({
    label: "灾备切换",
    value: "disaster_recovery",
}, ...__VLS_functionalComponentArgsRest(__VLS_27));
const __VLS_30 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_31 = __VLS_asFunctionalComponent(__VLS_30, new __VLS_30({
    label: "服务降级",
    value: "degradation",
}));
const __VLS_32 = __VLS_31({
    label: "服务降级",
    value: "degradation",
}, ...__VLS_functionalComponentArgsRest(__VLS_31));
const __VLS_34 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_35 = __VLS_asFunctionalComponent(__VLS_34, new __VLS_34({
    label: "发布回滚",
    value: "rollback",
}));
const __VLS_36 = __VLS_35({
    label: "发布回滚",
    value: "rollback",
}, ...__VLS_functionalComponentArgsRest(__VLS_35));
const __VLS_38 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_39 = __VLS_asFunctionalComponent(__VLS_38, new __VLS_38({
    label: "安全演练",
    value: "security",
}));
const __VLS_40 = __VLS_39({
    label: "安全演练",
    value: "security",
}, ...__VLS_functionalComponentArgsRest(__VLS_39));
var __VLS_25;
var __VLS_21;
const __VLS_42 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_43 = __VLS_asFunctionalComponent(__VLS_42, new __VLS_42({
    label: "演练描述",
    prop: "description",
}));
const __VLS_44 = __VLS_43({
    label: "演练描述",
    prop: "description",
}, ...__VLS_functionalComponentArgsRest(__VLS_43));
__VLS_45.slots.default;
const __VLS_46 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_47 = __VLS_asFunctionalComponent(__VLS_46, new __VLS_46({
    modelValue: (__VLS_ctx.form.description),
    type: "textarea",
    rows: (4),
    placeholder: "请输入演练描述",
}));
const __VLS_48 = __VLS_47({
    modelValue: (__VLS_ctx.form.description),
    type: "textarea",
    rows: (4),
    placeholder: "请输入演练描述",
}, ...__VLS_functionalComponentArgsRest(__VLS_47));
var __VLS_45;
const __VLS_50 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_51 = __VLS_asFunctionalComponent(__VLS_50, new __VLS_50({}));
const __VLS_52 = __VLS_51({}, ...__VLS_functionalComponentArgsRest(__VLS_51));
__VLS_53.slots.default;
const __VLS_54 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_55 = __VLS_asFunctionalComponent(__VLS_54, new __VLS_54({
    ...{ 'onClick': {} },
    type: "primary",
}));
const __VLS_56 = __VLS_55({
    ...{ 'onClick': {} },
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_55));
let __VLS_58;
let __VLS_59;
let __VLS_60;
const __VLS_61 = {
    onClick: (__VLS_ctx.handleSubmit)
};
__VLS_57.slots.default;
var __VLS_57;
const __VLS_62 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_63 = __VLS_asFunctionalComponent(__VLS_62, new __VLS_62({
    ...{ 'onClick': {} },
}));
const __VLS_64 = __VLS_63({
    ...{ 'onClick': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_63));
let __VLS_66;
let __VLS_67;
let __VLS_68;
const __VLS_69 = {
    onClick: (...[$event]) => {
        __VLS_ctx.$router.back();
    }
};
__VLS_65.slots.default;
var __VLS_65;
var __VLS_53;
var __VLS_7;
var __VLS_3;
/** @type {__VLS_StyleScopedClasses['drill-create-page']} */ ;
/** @type {__VLS_StyleScopedClasses['page-title']} */ ;
// @ts-ignore
var __VLS_9 = __VLS_8;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            formRef: formRef,
            form: form,
            rules: rules,
            handleSubmit: handleSubmit,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
