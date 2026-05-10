import { ref, computed } from 'vue';
import { ElMessage } from 'element-plus';
import { Refresh } from '@element-plus/icons-vue';
import templatesData from '@/mock/data/templates.json';
const activeCategory = ref('all');
const templates = ref([]);
const detailVisible = ref(false);
const cloneVisible = ref(false);
const selectedTemplate = ref(null);
const cloneForm = ref({ newName: '' });
const filteredTemplates = computed(() => {
    if (activeCategory.value === 'all') {
        return templates.value;
    }
    return templates.value.filter(t => t.category === activeCategory.value);
});
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
function getStepTypeLabel(type) {
    const map = {
        serial: '串行',
        parallel: '并行',
        any_of: '任选',
        condition: '条件',
    };
    return map[type] || type;
}
function getStepTypeTag(type) {
    const map = {
        serial: 'primary',
        parallel: 'success',
        any_of: 'warning',
        condition: 'info',
    };
    return map[type] || 'info';
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
async function loadTemplates() {
    try {
        templates.value = templatesData;
        ElMessage.success('模板加载成功');
    }
    catch (error) {
        ElMessage.error('加载模板失败');
        console.error('Failed to load templates:', error);
    }
}
function viewDetail(template) {
    selectedTemplate.value = template;
    detailVisible.value = true;
}
function cloneTemplate(template) {
    selectedTemplate.value = template;
    cloneForm.value.newName = `${template.name} (副本)`;
    cloneVisible.value = true;
}
function confirmClone() {
    if (!cloneForm.value.newName.trim()) {
        ElMessage.warning('请输入新模板名称');
        return;
    }
    ElMessage.success(`模板「${selectedTemplate.value?.name}」已克隆为「${cloneForm.value.newName}」`);
    cloneVisible.value = false;
    selectedTemplate.value = null;
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
    onClick: (__VLS_ctx.loadTemplates)
};
__VLS_3.slots.default;
const __VLS_8 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({}));
const __VLS_10 = __VLS_9({}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
const __VLS_12 = {}.Refresh;
/** @type {[typeof __VLS_components.Refresh, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({}));
const __VLS_14 = __VLS_13({}, ...__VLS_functionalComponentArgsRest(__VLS_13));
var __VLS_11;
var __VLS_3;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-content" },
});
const __VLS_16 = {}.ElTabs;
/** @type {[typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    modelValue: (__VLS_ctx.activeCategory),
    ...{ class: "category-tabs" },
}));
const __VLS_18 = __VLS_17({
    modelValue: (__VLS_ctx.activeCategory),
    ...{ class: "category-tabs" },
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
const __VLS_20 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    label: "全部",
    name: "all",
}));
const __VLS_22 = __VLS_21({
    label: "全部",
    name: "all",
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
const __VLS_24 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    label: "灾备切换",
    name: "disaster_recovery",
}));
const __VLS_26 = __VLS_25({
    label: "灾备切换",
    name: "disaster_recovery",
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
const __VLS_28 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    label: "服务降级",
    name: "degradation",
}));
const __VLS_30 = __VLS_29({
    label: "服务降级",
    name: "degradation",
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
const __VLS_32 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    label: "发布演练",
    name: "release",
}));
const __VLS_34 = __VLS_33({
    label: "发布演练",
    name: "release",
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
const __VLS_36 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
    label: "安全事件",
    name: "security",
}));
const __VLS_38 = __VLS_37({
    label: "安全事件",
    name: "security",
}, ...__VLS_functionalComponentArgsRest(__VLS_37));
var __VLS_19;
const __VLS_40 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
    data: (__VLS_ctx.filteredTemplates),
    ...{ style: {} },
    ...{ class: "templates-table" },
}));
const __VLS_42 = __VLS_41({
    data: (__VLS_ctx.filteredTemplates),
    ...{ style: {} },
    ...{ class: "templates-table" },
}, ...__VLS_functionalComponentArgsRest(__VLS_41));
__VLS_43.slots.default;
const __VLS_44 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
    prop: "name",
    label: "模板名",
    minWidth: "200",
}));
const __VLS_46 = __VLS_45({
    prop: "name",
    label: "模板名",
    minWidth: "200",
}, ...__VLS_functionalComponentArgsRest(__VLS_45));
const __VLS_48 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
    prop: "category",
    label: "分类",
    width: "120",
}));
const __VLS_50 = __VLS_49({
    prop: "category",
    label: "分类",
    width: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_49));
__VLS_51.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_51.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_52 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({
        type: (__VLS_ctx.getCategoryTagType(row.category)),
        size: "small",
    }));
    const __VLS_54 = __VLS_53({
        type: (__VLS_ctx.getCategoryTagType(row.category)),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_53));
    __VLS_55.slots.default;
    (__VLS_ctx.getCategoryLabel(row.category));
    var __VLS_55;
}
var __VLS_51;
const __VLS_56 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
    prop: "version",
    label: "版本",
    width: "100",
}));
const __VLS_58 = __VLS_57({
    prop: "version",
    label: "版本",
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_57));
const __VLS_60 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
    prop: "status",
    label: "状态",
    width: "100",
}));
const __VLS_62 = __VLS_61({
    prop: "status",
    label: "状态",
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_61));
__VLS_63.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_63.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_64 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
        type: (row.status === 'published' ? 'success' : 'info'),
        size: "small",
    }));
    const __VLS_66 = __VLS_65({
        type: (row.status === 'published' ? 'success' : 'info'),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_65));
    __VLS_67.slots.default;
    (row.status === 'published' ? '已发布' : '草稿');
    var __VLS_67;
}
var __VLS_63;
const __VLS_68 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({
    prop: "created_by_name",
    label: "创建人",
    width: "120",
}));
const __VLS_70 = __VLS_69({
    prop: "created_by_name",
    label: "创建人",
    width: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_69));
const __VLS_72 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
    prop: "updated_at",
    label: "更新时间",
    width: "160",
}));
const __VLS_74 = __VLS_73({
    prop: "updated_at",
    label: "更新时间",
    width: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_73));
__VLS_75.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_75.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatTime(row.updated_at));
}
var __VLS_75;
const __VLS_76 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({
    label: "操作",
    width: "180",
    fixed: "right",
}));
const __VLS_78 = __VLS_77({
    label: "操作",
    width: "180",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_77));
__VLS_79.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_79.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_80 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
    }));
    const __VLS_82 = __VLS_81({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_81));
    let __VLS_84;
    let __VLS_85;
    let __VLS_86;
    const __VLS_87 = {
        onClick: (...[$event]) => {
            __VLS_ctx.viewDetail(row);
        }
    };
    __VLS_83.slots.default;
    var __VLS_83;
    const __VLS_88 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
    }));
    const __VLS_90 = __VLS_89({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_89));
    let __VLS_92;
    let __VLS_93;
    let __VLS_94;
    const __VLS_95 = {
        onClick: (...[$event]) => {
            __VLS_ctx.cloneTemplate(row);
        }
    };
    __VLS_91.slots.default;
    var __VLS_91;
}
var __VLS_79;
var __VLS_43;
const __VLS_96 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
    modelValue: (__VLS_ctx.detailVisible),
    title: "模板详情",
    width: "600px",
}));
const __VLS_98 = __VLS_97({
    modelValue: (__VLS_ctx.detailVisible),
    title: "模板详情",
    width: "600px",
}, ...__VLS_functionalComponentArgsRest(__VLS_97));
__VLS_99.slots.default;
if (__VLS_ctx.selectedTemplate) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "template-detail" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.h3, __VLS_intrinsicElements.h3)({});
    (__VLS_ctx.selectedTemplate.name);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
        ...{ class: "description" },
    });
    (__VLS_ctx.selectedTemplate.description);
    const __VLS_100 = {}.ElDescriptions;
    /** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
    // @ts-ignore
    const __VLS_101 = __VLS_asFunctionalComponent(__VLS_100, new __VLS_100({
        column: (2),
        border: true,
    }));
    const __VLS_102 = __VLS_101({
        column: (2),
        border: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_101));
    __VLS_103.slots.default;
    const __VLS_104 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_105 = __VLS_asFunctionalComponent(__VLS_104, new __VLS_104({
        label: "分类",
    }));
    const __VLS_106 = __VLS_105({
        label: "分类",
    }, ...__VLS_functionalComponentArgsRest(__VLS_105));
    __VLS_107.slots.default;
    const __VLS_108 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_109 = __VLS_asFunctionalComponent(__VLS_108, new __VLS_108({
        type: (__VLS_ctx.getCategoryTagType(__VLS_ctx.selectedTemplate.category)),
        size: "small",
    }));
    const __VLS_110 = __VLS_109({
        type: (__VLS_ctx.getCategoryTagType(__VLS_ctx.selectedTemplate.category)),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_109));
    __VLS_111.slots.default;
    (__VLS_ctx.getCategoryLabel(__VLS_ctx.selectedTemplate.category));
    var __VLS_111;
    var __VLS_107;
    const __VLS_112 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_113 = __VLS_asFunctionalComponent(__VLS_112, new __VLS_112({
        label: "版本",
    }));
    const __VLS_114 = __VLS_113({
        label: "版本",
    }, ...__VLS_functionalComponentArgsRest(__VLS_113));
    __VLS_115.slots.default;
    (__VLS_ctx.selectedTemplate.version);
    var __VLS_115;
    const __VLS_116 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_117 = __VLS_asFunctionalComponent(__VLS_116, new __VLS_116({
        label: "创建人",
    }));
    const __VLS_118 = __VLS_117({
        label: "创建人",
    }, ...__VLS_functionalComponentArgsRest(__VLS_117));
    __VLS_119.slots.default;
    (__VLS_ctx.selectedTemplate.created_by_name);
    var __VLS_119;
    const __VLS_120 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_121 = __VLS_asFunctionalComponent(__VLS_120, new __VLS_120({
        label: "更新时间",
    }));
    const __VLS_122 = __VLS_121({
        label: "更新时间",
    }, ...__VLS_functionalComponentArgsRest(__VLS_121));
    __VLS_123.slots.default;
    (__VLS_ctx.formatTime(__VLS_ctx.selectedTemplate.updated_at));
    var __VLS_123;
    var __VLS_103;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.h4, __VLS_intrinsicElements.h4)({});
    const __VLS_124 = {}.ElTimeline;
    /** @type {[typeof __VLS_components.ElTimeline, typeof __VLS_components.elTimeline, typeof __VLS_components.ElTimeline, typeof __VLS_components.elTimeline, ]} */ ;
    // @ts-ignore
    const __VLS_125 = __VLS_asFunctionalComponent(__VLS_124, new __VLS_124({}));
    const __VLS_126 = __VLS_125({}, ...__VLS_functionalComponentArgsRest(__VLS_125));
    __VLS_127.slots.default;
    for (const [step] of __VLS_getVForSourceType((__VLS_ctx.selectedTemplate.steps))) {
        const __VLS_128 = {}.ElTimelineItem;
        /** @type {[typeof __VLS_components.ElTimelineItem, typeof __VLS_components.elTimelineItem, typeof __VLS_components.ElTimelineItem, typeof __VLS_components.elTimelineItem, ]} */ ;
        // @ts-ignore
        const __VLS_129 = __VLS_asFunctionalComponent(__VLS_128, new __VLS_128({
            key: (step.id),
            timestamp: (`步骤${step.order_index}`),
            placement: "top",
        }));
        const __VLS_130 = __VLS_129({
            key: (step.id),
            timestamp: (`步骤${step.order_index}`),
            placement: "top",
        }, ...__VLS_functionalComponentArgsRest(__VLS_129));
        __VLS_131.slots.default;
        const __VLS_132 = {}.ElCard;
        /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
        // @ts-ignore
        const __VLS_133 = __VLS_asFunctionalComponent(__VLS_132, new __VLS_132({}));
        const __VLS_134 = __VLS_133({}, ...__VLS_functionalComponentArgsRest(__VLS_133));
        __VLS_135.slots.default;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.h4, __VLS_intrinsicElements.h4)({});
        (step.name);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
        (step.description);
        const __VLS_136 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_137 = __VLS_asFunctionalComponent(__VLS_136, new __VLS_136({
            type: (__VLS_ctx.getStepTypeTag(step.step_type)),
            size: "small",
        }));
        const __VLS_138 = __VLS_137({
            type: (__VLS_ctx.getStepTypeTag(step.step_type)),
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_137));
        __VLS_139.slots.default;
        (__VLS_ctx.getStepTypeLabel(step.step_type));
        var __VLS_139;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "timeout-info" },
        });
        (step.timeout_seconds);
        var __VLS_135;
        var __VLS_131;
    }
    var __VLS_127;
}
var __VLS_99;
const __VLS_140 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_141 = __VLS_asFunctionalComponent(__VLS_140, new __VLS_140({
    modelValue: (__VLS_ctx.cloneVisible),
    title: "克隆模板",
    width: "400px",
}));
const __VLS_142 = __VLS_141({
    modelValue: (__VLS_ctx.cloneVisible),
    title: "克隆模板",
    width: "400px",
}, ...__VLS_functionalComponentArgsRest(__VLS_141));
__VLS_143.slots.default;
const __VLS_144 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_145 = __VLS_asFunctionalComponent(__VLS_144, new __VLS_144({
    model: (__VLS_ctx.cloneForm),
    labelWidth: "80px",
}));
const __VLS_146 = __VLS_145({
    model: (__VLS_ctx.cloneForm),
    labelWidth: "80px",
}, ...__VLS_functionalComponentArgsRest(__VLS_145));
__VLS_147.slots.default;
const __VLS_148 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_149 = __VLS_asFunctionalComponent(__VLS_148, new __VLS_148({
    label: "新模板名称",
    required: true,
}));
const __VLS_150 = __VLS_149({
    label: "新模板名称",
    required: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_149));
__VLS_151.slots.default;
const __VLS_152 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_153 = __VLS_asFunctionalComponent(__VLS_152, new __VLS_152({
    modelValue: (__VLS_ctx.cloneForm.newName),
    placeholder: "请输入新模板名称",
}));
const __VLS_154 = __VLS_153({
    modelValue: (__VLS_ctx.cloneForm.newName),
    placeholder: "请输入新模板名称",
}, ...__VLS_functionalComponentArgsRest(__VLS_153));
var __VLS_151;
var __VLS_147;
{
    const { footer: __VLS_thisSlot } = __VLS_143.slots;
    const __VLS_156 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_157 = __VLS_asFunctionalComponent(__VLS_156, new __VLS_156({
        ...{ 'onClick': {} },
    }));
    const __VLS_158 = __VLS_157({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_157));
    let __VLS_160;
    let __VLS_161;
    let __VLS_162;
    const __VLS_163 = {
        onClick: (...[$event]) => {
            __VLS_ctx.cloneVisible = false;
        }
    };
    __VLS_159.slots.default;
    var __VLS_159;
    const __VLS_164 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_165 = __VLS_asFunctionalComponent(__VLS_164, new __VLS_164({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_166 = __VLS_165({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_165));
    let __VLS_168;
    let __VLS_169;
    let __VLS_170;
    const __VLS_171 = {
        onClick: (__VLS_ctx.confirmClone)
    };
    __VLS_167.slots.default;
    var __VLS_167;
}
var __VLS_143;
/** @type {__VLS_StyleScopedClasses['page-container']} */ ;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
/** @type {__VLS_StyleScopedClasses['page-title']} */ ;
/** @type {__VLS_StyleScopedClasses['page-content']} */ ;
/** @type {__VLS_StyleScopedClasses['category-tabs']} */ ;
/** @type {__VLS_StyleScopedClasses['templates-table']} */ ;
/** @type {__VLS_StyleScopedClasses['template-detail']} */ ;
/** @type {__VLS_StyleScopedClasses['description']} */ ;
/** @type {__VLS_StyleScopedClasses['timeout-info']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Refresh: Refresh,
            activeCategory: activeCategory,
            detailVisible: detailVisible,
            cloneVisible: cloneVisible,
            selectedTemplate: selectedTemplate,
            cloneForm: cloneForm,
            filteredTemplates: filteredTemplates,
            getCategoryLabel: getCategoryLabel,
            getCategoryTagType: getCategoryTagType,
            getStepTypeLabel: getStepTypeLabel,
            getStepTypeTag: getStepTypeTag,
            formatTime: formatTime,
            loadTemplates: loadTemplates,
            viewDetail: viewDetail,
            cloneTemplate: cloneTemplate,
            confirmClone: confirmClone,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
