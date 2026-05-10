import { ref, reactive } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import templatesData from '@/mock/data/templates.json';
const router = useRouter();
const currentStep = ref(0);
const templates = ref([]);
const selectedTemplate = ref(null);
const formRef = ref();
const form = reactive({
    name: '',
    description: '',
});
const rules = {
    name: [
        { required: true, message: '请输入演练名称', trigger: 'blur' },
        { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' },
    ],
};
async function loadTemplates() {
    try {
        templates.value = templatesData;
    }
    catch (error) {
        ElMessage.error('加载模板失败');
        console.error('Failed to load templates:', error);
    }
}
function getCategoryLabel(category) {
    const map = {
        disaster_recovery: '灾备切换',
        degradation: '服务降级',
        release: '发布演练',
        security: '安全事件',
    };
    return map[category] || category;
}
function getCategoryTagType(category) {
    const map = {
        disaster_recovery: 'primary',
        degradation: 'warning',
        release: 'success',
        security: 'danger',
    };
    return map[category] || 'info';
}
function selectTemplate(template) {
    selectedTemplate.value = template;
}
async function prevStep() {
    currentStep.value--;
}
async function nextStep() {
    if (currentStep.value === 0) {
        if (!selectedTemplate.value) {
            ElMessage.warning('请先选择一个模板');
            return;
        }
    }
    if (currentStep.value === 1) {
        if (!formRef.value)
            return;
        await formRef.value.validate(valid => {
            if (valid) {
                currentStep.value++;
            }
            else {
                ElMessage.warning('请填写必填项');
            }
        });
        return;
    }
    currentStep.value++;
}
async function confirmCreate() {
    if (!selectedTemplate.value || !form.name.trim()) {
        ElMessage.warning('信息不完整，无法创建');
        return;
    }
    try {
        ElMessage.success('演练创建成功');
        router.push('/director/dashboard');
    }
    catch (error) {
        ElMessage.error('创建失败');
        console.error('Failed to create drill:', error);
    }
}
loadTemplates();
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
    ...{ class: "page-content" },
});
const __VLS_0 = {}.ElSteps;
/** @type {[typeof __VLS_components.ElSteps, typeof __VLS_components.elSteps, typeof __VLS_components.ElSteps, typeof __VLS_components.elSteps, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    active: (__VLS_ctx.currentStep),
    finishStatus: "success",
    alignCenter: true,
    ...{ class: "create-steps" },
}));
const __VLS_2 = __VLS_1({
    active: (__VLS_ctx.currentStep),
    finishStatus: "success",
    alignCenter: true,
    ...{ class: "create-steps" },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
const __VLS_4 = {}.ElStep;
/** @type {[typeof __VLS_components.ElStep, typeof __VLS_components.elStep, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
    title: "选择模板",
}));
const __VLS_6 = __VLS_5({
    title: "选择模板",
}, ...__VLS_functionalComponentArgsRest(__VLS_5));
const __VLS_8 = {}.ElStep;
/** @type {[typeof __VLS_components.ElStep, typeof __VLS_components.elStep, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    title: "配置基本信息",
}));
const __VLS_10 = __VLS_9({
    title: "配置基本信息",
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
const __VLS_12 = {}.ElStep;
/** @type {[typeof __VLS_components.ElStep, typeof __VLS_components.elStep, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    title: "确认创建",
}));
const __VLS_14 = __VLS_13({
    title: "确认创建",
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
var __VLS_3;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "step-content" },
});
__VLS_asFunctionalDirective(__VLS_directives.vShow)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.currentStep === 0) }, null, null);
const __VLS_16 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    gutter: (20),
}));
const __VLS_18 = __VLS_17({
    gutter: (20),
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
for (const [template] of __VLS_getVForSourceType((__VLS_ctx.templates))) {
    const __VLS_20 = {}.ElCol;
    /** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
    // @ts-ignore
    const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
        key: (template.id),
        xs: (24),
        sm: (12),
        lg: (8),
    }));
    const __VLS_22 = __VLS_21({
        key: (template.id),
        xs: (24),
        sm: (12),
        lg: (8),
    }, ...__VLS_functionalComponentArgsRest(__VLS_21));
    __VLS_23.slots.default;
    const __VLS_24 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
        ...{ 'onClick': {} },
        ...{ class: "template-card" },
        ...{ class: ({ selected: __VLS_ctx.selectedTemplate?.id === template.id }) },
    }));
    const __VLS_26 = __VLS_25({
        ...{ 'onClick': {} },
        ...{ class: "template-card" },
        ...{ class: ({ selected: __VLS_ctx.selectedTemplate?.id === template.id }) },
    }, ...__VLS_functionalComponentArgsRest(__VLS_25));
    let __VLS_28;
    let __VLS_29;
    let __VLS_30;
    const __VLS_31 = {
        onClick: (...[$event]) => {
            __VLS_ctx.selectTemplate(template);
        }
    };
    __VLS_27.slots.default;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "template-name" },
    });
    (template.name);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "template-description" },
    });
    (template.description);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "template-meta" },
    });
    const __VLS_32 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
        type: (__VLS_ctx.getCategoryTagType(template.category)),
        size: "small",
    }));
    const __VLS_34 = __VLS_33({
        type: (__VLS_ctx.getCategoryTagType(template.category)),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_33));
    __VLS_35.slots.default;
    (__VLS_ctx.getCategoryLabel(template.category));
    var __VLS_35;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "steps-count" },
    });
    (template.steps.length);
    var __VLS_27;
    var __VLS_23;
}
var __VLS_19;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "step-content" },
});
__VLS_asFunctionalDirective(__VLS_directives.vShow)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.currentStep === 1) }, null, null);
const __VLS_36 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
    ref: "formRef",
    model: (__VLS_ctx.form),
    rules: (__VLS_ctx.rules),
    labelWidth: "100px",
    ...{ class: "create-form" },
}));
const __VLS_38 = __VLS_37({
    ref: "formRef",
    model: (__VLS_ctx.form),
    rules: (__VLS_ctx.rules),
    labelWidth: "100px",
    ...{ class: "create-form" },
}, ...__VLS_functionalComponentArgsRest(__VLS_37));
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_40 = {};
__VLS_39.slots.default;
const __VLS_42 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_43 = __VLS_asFunctionalComponent(__VLS_42, new __VLS_42({
    label: "演练名称",
    prop: "name",
}));
const __VLS_44 = __VLS_43({
    label: "演练名称",
    prop: "name",
}, ...__VLS_functionalComponentArgsRest(__VLS_43));
__VLS_45.slots.default;
const __VLS_46 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_47 = __VLS_asFunctionalComponent(__VLS_46, new __VLS_46({
    modelValue: (__VLS_ctx.form.name),
    placeholder: "请输入演练名称",
    maxlength: "50",
    showWordLimit: true,
}));
const __VLS_48 = __VLS_47({
    modelValue: (__VLS_ctx.form.name),
    placeholder: "请输入演练名称",
    maxlength: "50",
    showWordLimit: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_47));
var __VLS_45;
const __VLS_50 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_51 = __VLS_asFunctionalComponent(__VLS_50, new __VLS_50({
    label: "描述",
    prop: "description",
}));
const __VLS_52 = __VLS_51({
    label: "描述",
    prop: "description",
}, ...__VLS_functionalComponentArgsRest(__VLS_51));
__VLS_53.slots.default;
const __VLS_54 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_55 = __VLS_asFunctionalComponent(__VLS_54, new __VLS_54({
    modelValue: (__VLS_ctx.form.description),
    type: "textarea",
    rows: (4),
    placeholder: "请输入演练描述（可选）",
    maxlength: "200",
    showWordLimit: true,
}));
const __VLS_56 = __VLS_55({
    modelValue: (__VLS_ctx.form.description),
    type: "textarea",
    rows: (4),
    placeholder: "请输入演练描述（可选）",
    maxlength: "200",
    showWordLimit: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_55));
var __VLS_53;
var __VLS_39;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "step-content" },
});
__VLS_asFunctionalDirective(__VLS_directives.vShow)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.currentStep === 2) }, null, null);
const __VLS_58 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_59 = __VLS_asFunctionalComponent(__VLS_58, new __VLS_58({
    ...{ class: "confirm-card" },
}));
const __VLS_60 = __VLS_59({
    ...{ class: "confirm-card" },
}, ...__VLS_functionalComponentArgsRest(__VLS_59));
__VLS_61.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({});
const __VLS_62 = {}.ElDescriptions;
/** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
// @ts-ignore
const __VLS_63 = __VLS_asFunctionalComponent(__VLS_62, new __VLS_62({
    column: (1),
    border: true,
}));
const __VLS_64 = __VLS_63({
    column: (1),
    border: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_63));
__VLS_65.slots.default;
const __VLS_66 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_67 = __VLS_asFunctionalComponent(__VLS_66, new __VLS_66({
    label: "选择模板",
}));
const __VLS_68 = __VLS_67({
    label: "选择模板",
}, ...__VLS_functionalComponentArgsRest(__VLS_67));
__VLS_69.slots.default;
(__VLS_ctx.selectedTemplate?.name);
var __VLS_69;
const __VLS_70 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_71 = __VLS_asFunctionalComponent(__VLS_70, new __VLS_70({
    label: "演练名称",
}));
const __VLS_72 = __VLS_71({
    label: "演练名称",
}, ...__VLS_functionalComponentArgsRest(__VLS_71));
__VLS_73.slots.default;
(__VLS_ctx.form.name);
var __VLS_73;
const __VLS_74 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_75 = __VLS_asFunctionalComponent(__VLS_74, new __VLS_74({
    label: "演练描述",
}));
const __VLS_76 = __VLS_75({
    label: "演练描述",
}, ...__VLS_functionalComponentArgsRest(__VLS_75));
__VLS_77.slots.default;
(__VLS_ctx.form.description || '无');
var __VLS_77;
const __VLS_78 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_79 = __VLS_asFunctionalComponent(__VLS_78, new __VLS_78({
    label: "步骤数量",
}));
const __VLS_80 = __VLS_79({
    label: "步骤数量",
}, ...__VLS_functionalComponentArgsRest(__VLS_79));
__VLS_81.slots.default;
(__VLS_ctx.selectedTemplate?.steps.length);
var __VLS_81;
var __VLS_65;
var __VLS_61;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "step-actions" },
});
if (__VLS_ctx.currentStep > 0) {
    const __VLS_82 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_83 = __VLS_asFunctionalComponent(__VLS_82, new __VLS_82({
        ...{ 'onClick': {} },
    }));
    const __VLS_84 = __VLS_83({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_83));
    let __VLS_86;
    let __VLS_87;
    let __VLS_88;
    const __VLS_89 = {
        onClick: (__VLS_ctx.prevStep)
    };
    __VLS_85.slots.default;
    var __VLS_85;
}
if (__VLS_ctx.currentStep < 2) {
    const __VLS_90 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_91 = __VLS_asFunctionalComponent(__VLS_90, new __VLS_90({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_92 = __VLS_91({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_91));
    let __VLS_94;
    let __VLS_95;
    let __VLS_96;
    const __VLS_97 = {
        onClick: (__VLS_ctx.nextStep)
    };
    __VLS_93.slots.default;
    var __VLS_93;
}
else {
    const __VLS_98 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_99 = __VLS_asFunctionalComponent(__VLS_98, new __VLS_98({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_100 = __VLS_99({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_99));
    let __VLS_102;
    let __VLS_103;
    let __VLS_104;
    const __VLS_105 = {
        onClick: (__VLS_ctx.confirmCreate)
    };
    __VLS_101.slots.default;
    var __VLS_101;
}
/** @type {__VLS_StyleScopedClasses['page-container']} */ ;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
/** @type {__VLS_StyleScopedClasses['page-title']} */ ;
/** @type {__VLS_StyleScopedClasses['page-content']} */ ;
/** @type {__VLS_StyleScopedClasses['create-steps']} */ ;
/** @type {__VLS_StyleScopedClasses['step-content']} */ ;
/** @type {__VLS_StyleScopedClasses['template-card']} */ ;
/** @type {__VLS_StyleScopedClasses['selected']} */ ;
/** @type {__VLS_StyleScopedClasses['template-name']} */ ;
/** @type {__VLS_StyleScopedClasses['template-description']} */ ;
/** @type {__VLS_StyleScopedClasses['template-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['steps-count']} */ ;
/** @type {__VLS_StyleScopedClasses['step-content']} */ ;
/** @type {__VLS_StyleScopedClasses['create-form']} */ ;
/** @type {__VLS_StyleScopedClasses['step-content']} */ ;
/** @type {__VLS_StyleScopedClasses['confirm-card']} */ ;
/** @type {__VLS_StyleScopedClasses['step-actions']} */ ;
// @ts-ignore
var __VLS_41 = __VLS_40;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            currentStep: currentStep,
            templates: templates,
            selectedTemplate: selectedTemplate,
            formRef: formRef,
            form: form,
            rules: rules,
            getCategoryLabel: getCategoryLabel,
            getCategoryTagType: getCategoryTagType,
            selectTemplate: selectTemplate,
            prevStep: prevStep,
            nextStep: nextStep,
            confirmCreate: confirmCreate,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
