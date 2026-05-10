import { ref, computed, reactive } from 'vue';
import { ElMessage } from 'element-plus';
import { Refresh, Plus, Delete, Setting, Upload, Download, Top, Bottom, Edit } from '@element-plus/icons-vue';
import * as XLSX from 'xlsx';
import { useAuthStore } from '@/stores/auth';
import templatesData from '@/mock/data/templates.json';
const authStore = useAuthStore();
const defaultCategories = [
    { value: 'disaster_recovery', label: '灾备切换', tagType: 'primary' },
    { value: 'degradation', label: '服务降级', tagType: 'warning' },
    { value: 'release', label: '发布演练', tagType: 'success' },
    { value: 'security', label: '安全事件', tagType: 'danger' },
];
const activeCategory = ref('all');
const templates = ref([]);
const categories = ref([...defaultCategories]);
const formVisible = ref(false);
const stepsVisible = ref(false);
const deleteVisible = ref(false);
const categoryVisible = ref(false);
const importVisible = ref(false);
const editableCategories = ref([]);
const isEditing = ref(false);
const editingId = ref(null);
const editingSteps = ref([]);
const editingTemplateId = ref(null);
const editingTemplateName = ref('');
const deleteTarget = ref(null);
const filteredTemplates = computed(() => {
    if (activeCategory.value === 'all')
        return templates.value;
    return templates.value.filter(t => t.category === activeCategory.value);
});
const form = reactive({
    name: '',
    category: 'disaster_recovery',
    description: '',
});
const singleStepForm = reactive({
    name: '',
    description: '',
    step_type: 'serial',
    timeout_seconds: 300,
    assignee: '',
});
const singleStepEditIndex = ref(null);
const singleAddVisible = ref(false);
function openBatchImportDialog() {
    importVisible.value = true;
}
function openSingleAddDialog() {
    resetSingleStepForm();
    singleStepEditIndex.value = null;
    singleAddVisible.value = true;
}
function resetSingleStepForm() {
    singleStepForm.name = '';
    singleStepForm.description = '';
    singleStepForm.step_type = 'serial';
    singleStepForm.timeout_seconds = 300;
    singleStepForm.assignee = '';
}
function handleAddSingleStep() {
    if (!singleStepForm.name.trim()) {
        ElMessage.warning('请输入步骤名称');
        return;
    }
    if (singleStepEditIndex.value !== null) {
        // 编辑模式
        const step = editingSteps.value[singleStepEditIndex.value];
        step.name = singleStepForm.name.trim();
        step.description = singleStepForm.description.trim();
        step.step_type = singleStepForm.step_type;
        step.timeout_seconds = singleStepForm.timeout_seconds;
        step.assignee = singleStepForm.assignee.trim();
        ElMessage.success('步骤已更新');
    }
    else {
        // 新增模式
        editingSteps.value.push({
            id: Date.now(),
            template_id: editingTemplateId.value || 0,
            name: singleStepForm.name.trim(),
            description: singleStepForm.description.trim(),
            step_type: singleStepForm.step_type,
            timeout_seconds: singleStepForm.timeout_seconds,
            assignee: singleStepForm.assignee.trim(),
            order_index: editingSteps.value.length + 1,
            created_at: new Date().toISOString(),
        });
        ElMessage.success('步骤已添加');
    }
    resetSingleStepForm();
    singleAddVisible.value = false;
}
function moveStep(index, direction) {
    const newIndex = index + direction;
    if (newIndex < 0 || newIndex >= editingSteps.value.length)
        return;
    const temp = editingSteps.value[index];
    editingSteps.value[index] = editingSteps.value[newIndex];
    editingSteps.value[newIndex] = temp;
    editingSteps.value.forEach((s, i) => { s.order_index = i + 1; });
}
function openStepEditDialog(index) {
    const step = editingSteps.value[index];
    singleStepForm.name = step.name;
    singleStepForm.description = step.description || '';
    singleStepForm.step_type = step.step_type;
    singleStepForm.timeout_seconds = step.timeout_seconds;
    singleStepForm.assignee = step.assignee || '';
    singleStepEditIndex.value = index;
    singleAddVisible.value = true;
}
function getCategoryLabel(value) {
    const cat = categories.value.find(c => c.value === value);
    return cat?.label || value;
}
function getCategoryTagType(value) {
    const cat = categories.value.find(c => c.value === value);
    return cat?.tagType || 'info';
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
function formatTime(dateStr) {
    return new Date(dateStr).toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
    });
}
function loadTemplates() {
    templates.value = JSON.parse(JSON.stringify(templatesData));
    ElMessage.success('模板加载成功');
}
function openCategoryDialog() {
    editableCategories.value = JSON.parse(JSON.stringify(categories.value));
    categoryVisible.value = true;
}
function addCategory() {
    const newValue = `custom_${Date.now()}`;
    editableCategories.value.push({
        value: newValue,
        label: '新分类',
        tagType: 'info',
    });
}
function removeCategory(index) {
    const cat = editableCategories.value[index];
    if (templates.value.some(t => t.category === cat.value)) {
        ElMessage.warning('该分类下有模板，请先移除或转移模板');
        return;
    }
    editableCategories.value.splice(index, 1);
}
function moveCategory(index, direction) {
    const newIndex = index + direction;
    if (newIndex < 0 || newIndex >= editableCategories.value.length)
        return;
    const temp = editableCategories.value[index];
    editableCategories.value[index] = editableCategories.value[newIndex];
    editableCategories.value[newIndex] = temp;
}
function handleSaveCategories() {
    const labels = editableCategories.value.map(c => c.label);
    if (new Set(labels).size !== labels.length) {
        ElMessage.warning('分类名称不能重复');
        return;
    }
    for (const cat of editableCategories.value) {
        if (!cat.label.trim()) {
            ElMessage.warning('分类名称不能为空');
            return;
        }
    }
    categories.value = JSON.parse(JSON.stringify(editableCategories.value));
    ElMessage.success('分类已保存');
    categoryVisible.value = false;
}
function openCreateDialog() {
    isEditing.value = false;
    editingId.value = null;
    form.name = '';
    form.category = (categories.value[0]?.value || 'disaster_recovery');
    form.description = '';
    formVisible.value = true;
}
function openEditDialog(template) {
    isEditing.value = true;
    editingId.value = template.id;
    form.name = template.name;
    form.category = template.category;
    form.description = template.description || '';
    formVisible.value = true;
}
function handleSave() {
    if (!form.name.trim()) {
        ElMessage.warning('请输入模板名称');
        return;
    }
    if (!form.category) {
        ElMessage.warning('请选择分类');
        return;
    }
    if (isEditing.value && editingId.value) {
        const idx = templates.value.findIndex(t => t.id === editingId.value);
        if (idx !== -1) {
            templates.value[idx] = {
                ...templates.value[idx],
                name: form.name,
                category: form.category,
                description: form.description,
                updated_at: new Date().toISOString(),
            };
            ElMessage.success('模板已更新');
        }
    }
    else {
        const newTemplate = {
            id: Math.max(...templates.value.map(t => t.id), 0) + 1,
            name: form.name,
            category: form.category,
            description: form.description,
            version: '1.0.0',
            status: 'draft',
            created_by: authStore.user?.id || 1,
            created_by_name: authStore.userName || '当前用户',
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
            steps: [],
        };
        templates.value.push(newTemplate);
        ElMessage.success('模板已创建');
    }
    formVisible.value = false;
}
function openStepsDialog(template) {
    editingTemplateId.value = template.id;
    editingTemplateName.value = template.name;
    editingSteps.value = JSON.parse(JSON.stringify(template.steps || []));
    stepsVisible.value = true;
}
function removeStep(index) {
    editingSteps.value.splice(index, 1);
    editingSteps.value.forEach((s, i) => { s.order_index = i + 1; });
}
function handleSaveSteps() {
    const template = templates.value.find(t => t.id === editingTemplateId.value);
    if (template) {
        template.steps = editingSteps.value;
        template.updated_at = new Date().toISOString();
        ElMessage.success('步骤已保存');
    }
    stepsVisible.value = false;
}
function downloadTemplate() {
    const header = ['步骤名称', '描述', '步骤类型', '超时时间(秒)', '操作人'];
    const wb = XLSX.utils.book_new();
    const ws = XLSX.utils.aoa_to_sheet([header]);
    const colWidths = [
        { wch: 20 }, { wch: 30 }, { wch: 10 }, { wch: 12 }, { wch: 15 }
    ];
    ws['!cols'] = colWidths;
    XLSX.utils.book_append_sheet(wb, ws, '步骤导入');
    XLSX.writeFile(wb, `步骤导入模板_${editingTemplateName.value}.xlsx`);
    ElMessage.success('模板已下载');
}
function handleExcelUpload(file) {
    const reader = new FileReader();
    reader.onload = (e) => {
        try {
            const data = new Uint8Array(e.target?.result);
            const workbook = XLSX.read(data, { type: 'array', cellDates: true });
            const sheetName = workbook.SheetNames[0];
            const sheet = workbook.Sheets[sheetName];
            const rows = XLSX.utils.sheet_to_json(sheet, { header: 1 });
            if (rows.length < 2) {
                ElMessage.warning('Excel 文件内容为空');
                return false;
            }
            const steps = [];
            const errors = [];
            for (let i = 1; i < rows.length; i++) {
                const row = rows[i];
                const rowNum = i + 1;
                const name = String(row[0] || '').trim();
                const description = String(row[1] || '').trim();
                const stepTypeRaw = String(row[2] || '').trim();
                const timeoutSeconds = parseInt(String(row[3] || '300')) || 300;
                const assignee = String(row[4] || '').trim();
                if (!name) {
                    errors.push(`第${rowNum}行：步骤名称不能为空`);
                    continue;
                }
                const stepTypeMap = {
                    '串行': 'serial', '并行': 'parallel', '任选': 'any_of', '条件': 'condition',
                    'serial': 'serial', 'parallel': 'parallel', 'any_of': 'any_of', 'condition': 'condition',
                };
                const stepType = stepTypeMap[stepTypeRaw] || 'serial';
                steps.push({
                    id: Date.now() + Math.random(),
                    template_id: editingTemplateId.value || 0,
                    name,
                    description,
                    step_type: stepType,
                    timeout_seconds: Math.min(3600, Math.max(30, timeoutSeconds)),
                    assignee,
                    order_index: editingSteps.value.length + steps.length + 1,
                    created_at: new Date().toISOString(),
                });
            }
            if (errors.length > 0) {
                ElMessage.warning(errors.join('\n'));
            }
            if (steps.length > 0) {
                editingSteps.value.push(...steps);
                ElMessage.success(`成功导入 ${steps.length} 个步骤`);
            }
            else if (errors.length === 0) {
                ElMessage.warning('未找到有效数据');
            }
        }
        catch {
            ElMessage.error('Excel 文件解析失败');
        }
    };
    reader.readAsArrayBuffer(file);
    return false;
}
function handleDelete(template) {
    deleteTarget.value = template;
    deleteVisible.value = true;
}
function confirmDelete() {
    if (deleteTarget.value) {
        templates.value = templates.value.filter(t => t.id !== deleteTarget.value.id);
        ElMessage.success('模板已删除');
    }
    deleteVisible.value = false;
    deleteTarget.value = null;
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
    ...{ class: "header-actions" },
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
const __VLS_16 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    ...{ 'onClick': {} },
}));
const __VLS_18 = __VLS_17({
    ...{ 'onClick': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
let __VLS_20;
let __VLS_21;
let __VLS_22;
const __VLS_23 = {
    onClick: (__VLS_ctx.openCategoryDialog)
};
__VLS_19.slots.default;
const __VLS_24 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({}));
const __VLS_26 = __VLS_25({}, ...__VLS_functionalComponentArgsRest(__VLS_25));
__VLS_27.slots.default;
const __VLS_28 = {}.Setting;
/** @type {[typeof __VLS_components.Setting, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({}));
const __VLS_30 = __VLS_29({}, ...__VLS_functionalComponentArgsRest(__VLS_29));
var __VLS_27;
var __VLS_19;
const __VLS_32 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    ...{ 'onClick': {} },
    type: "primary",
}));
const __VLS_34 = __VLS_33({
    ...{ 'onClick': {} },
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
let __VLS_36;
let __VLS_37;
let __VLS_38;
const __VLS_39 = {
    onClick: (__VLS_ctx.openCreateDialog)
};
__VLS_35.slots.default;
const __VLS_40 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({}));
const __VLS_42 = __VLS_41({}, ...__VLS_functionalComponentArgsRest(__VLS_41));
__VLS_43.slots.default;
const __VLS_44 = {}.Plus;
/** @type {[typeof __VLS_components.Plus, ]} */ ;
// @ts-ignore
const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({}));
const __VLS_46 = __VLS_45({}, ...__VLS_functionalComponentArgsRest(__VLS_45));
var __VLS_43;
var __VLS_35;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "page-content" },
});
const __VLS_48 = {}.ElTabs;
/** @type {[typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, ]} */ ;
// @ts-ignore
const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
    modelValue: (__VLS_ctx.activeCategory),
    ...{ class: "category-tabs" },
}));
const __VLS_50 = __VLS_49({
    modelValue: (__VLS_ctx.activeCategory),
    ...{ class: "category-tabs" },
}, ...__VLS_functionalComponentArgsRest(__VLS_49));
__VLS_51.slots.default;
const __VLS_52 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({
    label: "全部",
    name: "all",
}));
const __VLS_54 = __VLS_53({
    label: "全部",
    name: "all",
}, ...__VLS_functionalComponentArgsRest(__VLS_53));
for (const [cat] of __VLS_getVForSourceType((__VLS_ctx.categories))) {
    const __VLS_56 = {}.ElTabPane;
    /** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
    // @ts-ignore
    const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
        key: (cat.value),
        label: (cat.label),
        name: (cat.value),
    }));
    const __VLS_58 = __VLS_57({
        key: (cat.value),
        label: (cat.label),
        name: (cat.value),
    }, ...__VLS_functionalComponentArgsRest(__VLS_57));
}
var __VLS_51;
const __VLS_60 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
    data: (__VLS_ctx.filteredTemplates),
    ...{ style: {} },
    ...{ class: "templates-table" },
}));
const __VLS_62 = __VLS_61({
    data: (__VLS_ctx.filteredTemplates),
    ...{ style: {} },
    ...{ class: "templates-table" },
}, ...__VLS_functionalComponentArgsRest(__VLS_61));
__VLS_63.slots.default;
const __VLS_64 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
    prop: "name",
    label: "模板名",
    minWidth: "200",
}));
const __VLS_66 = __VLS_65({
    prop: "name",
    label: "模板名",
    minWidth: "200",
}, ...__VLS_functionalComponentArgsRest(__VLS_65));
const __VLS_68 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({
    prop: "category",
    label: "分类",
    width: "120",
}));
const __VLS_70 = __VLS_69({
    prop: "category",
    label: "分类",
    width: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_69));
__VLS_71.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_71.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_72 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
        type: (__VLS_ctx.getCategoryTagType(row.category)),
        size: "small",
    }));
    const __VLS_74 = __VLS_73({
        type: (__VLS_ctx.getCategoryTagType(row.category)),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_73));
    __VLS_75.slots.default;
    (__VLS_ctx.getCategoryLabel(row.category));
    var __VLS_75;
}
var __VLS_71;
const __VLS_76 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({
    prop: "version",
    label: "版本",
    width: "100",
}));
const __VLS_78 = __VLS_77({
    prop: "version",
    label: "版本",
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_77));
const __VLS_80 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
    prop: "status",
    label: "状态",
    width: "100",
}));
const __VLS_82 = __VLS_81({
    prop: "status",
    label: "状态",
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_81));
__VLS_83.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_83.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_84 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
        type: (row.status === 'published' ? 'success' : 'info'),
        size: "small",
    }));
    const __VLS_86 = __VLS_85({
        type: (row.status === 'published' ? 'success' : 'info'),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_85));
    __VLS_87.slots.default;
    (row.status === 'published' ? '已发布' : '草稿');
    var __VLS_87;
}
var __VLS_83;
const __VLS_88 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
    prop: "created_by_name",
    label: "创建人",
    width: "120",
}));
const __VLS_90 = __VLS_89({
    prop: "created_by_name",
    label: "创建人",
    width: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_89));
const __VLS_92 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_93 = __VLS_asFunctionalComponent(__VLS_92, new __VLS_92({
    prop: "updated_at",
    label: "更新时间",
    width: "160",
}));
const __VLS_94 = __VLS_93({
    prop: "updated_at",
    label: "更新时间",
    width: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_93));
__VLS_95.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_95.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatTime(row.updated_at));
}
var __VLS_95;
const __VLS_96 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
    label: "操作",
    width: "280",
    fixed: "right",
}));
const __VLS_98 = __VLS_97({
    label: "操作",
    width: "280",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_97));
__VLS_99.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_99.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_100 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_101 = __VLS_asFunctionalComponent(__VLS_100, new __VLS_100({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
    }));
    const __VLS_102 = __VLS_101({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_101));
    let __VLS_104;
    let __VLS_105;
    let __VLS_106;
    const __VLS_107 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openEditDialog(row);
        }
    };
    __VLS_103.slots.default;
    var __VLS_103;
    const __VLS_108 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_109 = __VLS_asFunctionalComponent(__VLS_108, new __VLS_108({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
    }));
    const __VLS_110 = __VLS_109({
        ...{ 'onClick': {} },
        text: true,
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_109));
    let __VLS_112;
    let __VLS_113;
    let __VLS_114;
    const __VLS_115 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openStepsDialog(row);
        }
    };
    __VLS_111.slots.default;
    var __VLS_111;
    const __VLS_116 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_117 = __VLS_asFunctionalComponent(__VLS_116, new __VLS_116({
        ...{ 'onClick': {} },
        text: true,
        type: "danger",
    }));
    const __VLS_118 = __VLS_117({
        ...{ 'onClick': {} },
        text: true,
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_117));
    let __VLS_120;
    let __VLS_121;
    let __VLS_122;
    const __VLS_123 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleDelete(row);
        }
    };
    __VLS_119.slots.default;
    var __VLS_119;
}
var __VLS_99;
var __VLS_63;
const __VLS_124 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_125 = __VLS_asFunctionalComponent(__VLS_124, new __VLS_124({
    modelValue: (__VLS_ctx.categoryVisible),
    title: "分类管理",
    width: "500px",
}));
const __VLS_126 = __VLS_125({
    modelValue: (__VLS_ctx.categoryVisible),
    title: "分类管理",
    width: "500px",
}, ...__VLS_functionalComponentArgsRest(__VLS_125));
__VLS_127.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "category-list" },
});
for (const [cat, index] of __VLS_getVForSourceType((__VLS_ctx.editableCategories))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (cat.value),
        ...{ class: "category-item" },
    });
    const __VLS_128 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_129 = __VLS_asFunctionalComponent(__VLS_128, new __VLS_128({
        type: (cat.tagType),
        size: "small",
        ...{ style: {} },
    }));
    const __VLS_130 = __VLS_129({
        type: (cat.tagType),
        size: "small",
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_129));
    __VLS_131.slots.default;
    (cat.label);
    var __VLS_131;
    const __VLS_132 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_133 = __VLS_asFunctionalComponent(__VLS_132, new __VLS_132({
        modelValue: (cat.label),
        size: "small",
        ...{ class: "category-input" },
        placeholder: "分类名称",
    }));
    const __VLS_134 = __VLS_133({
        modelValue: (cat.label),
        size: "small",
        ...{ class: "category-input" },
        placeholder: "分类名称",
    }, ...__VLS_functionalComponentArgsRest(__VLS_133));
    const __VLS_136 = {}.ElButtonGroup;
    /** @type {[typeof __VLS_components.ElButtonGroup, typeof __VLS_components.elButtonGroup, typeof __VLS_components.ElButtonGroup, typeof __VLS_components.elButtonGroup, ]} */ ;
    // @ts-ignore
    const __VLS_137 = __VLS_asFunctionalComponent(__VLS_136, new __VLS_136({}));
    const __VLS_138 = __VLS_137({}, ...__VLS_functionalComponentArgsRest(__VLS_137));
    __VLS_139.slots.default;
    const __VLS_140 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_141 = __VLS_asFunctionalComponent(__VLS_140, new __VLS_140({
        ...{ 'onClick': {} },
        size: "small",
        disabled: (index === 0),
    }));
    const __VLS_142 = __VLS_141({
        ...{ 'onClick': {} },
        size: "small",
        disabled: (index === 0),
    }, ...__VLS_functionalComponentArgsRest(__VLS_141));
    let __VLS_144;
    let __VLS_145;
    let __VLS_146;
    const __VLS_147 = {
        onClick: (...[$event]) => {
            __VLS_ctx.moveCategory(index, -1);
        }
    };
    __VLS_143.slots.default;
    const __VLS_148 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_149 = __VLS_asFunctionalComponent(__VLS_148, new __VLS_148({}));
    const __VLS_150 = __VLS_149({}, ...__VLS_functionalComponentArgsRest(__VLS_149));
    __VLS_151.slots.default;
    const __VLS_152 = {}.ArrowUp;
    /** @type {[typeof __VLS_components.ArrowUp, ]} */ ;
    // @ts-ignore
    const __VLS_153 = __VLS_asFunctionalComponent(__VLS_152, new __VLS_152({}));
    const __VLS_154 = __VLS_153({}, ...__VLS_functionalComponentArgsRest(__VLS_153));
    var __VLS_151;
    var __VLS_143;
    const __VLS_156 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_157 = __VLS_asFunctionalComponent(__VLS_156, new __VLS_156({
        ...{ 'onClick': {} },
        size: "small",
        disabled: (index === __VLS_ctx.editableCategories.length - 1),
    }));
    const __VLS_158 = __VLS_157({
        ...{ 'onClick': {} },
        size: "small",
        disabled: (index === __VLS_ctx.editableCategories.length - 1),
    }, ...__VLS_functionalComponentArgsRest(__VLS_157));
    let __VLS_160;
    let __VLS_161;
    let __VLS_162;
    const __VLS_163 = {
        onClick: (...[$event]) => {
            __VLS_ctx.moveCategory(index, 1);
        }
    };
    __VLS_159.slots.default;
    const __VLS_164 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_165 = __VLS_asFunctionalComponent(__VLS_164, new __VLS_164({}));
    const __VLS_166 = __VLS_165({}, ...__VLS_functionalComponentArgsRest(__VLS_165));
    __VLS_167.slots.default;
    const __VLS_168 = {}.ArrowDown;
    /** @type {[typeof __VLS_components.ArrowDown, ]} */ ;
    // @ts-ignore
    const __VLS_169 = __VLS_asFunctionalComponent(__VLS_168, new __VLS_168({}));
    const __VLS_170 = __VLS_169({}, ...__VLS_functionalComponentArgsRest(__VLS_169));
    var __VLS_167;
    var __VLS_159;
    const __VLS_172 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_173 = __VLS_asFunctionalComponent(__VLS_172, new __VLS_172({
        ...{ 'onClick': {} },
        size: "small",
        type: "danger",
    }));
    const __VLS_174 = __VLS_173({
        ...{ 'onClick': {} },
        size: "small",
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_173));
    let __VLS_176;
    let __VLS_177;
    let __VLS_178;
    const __VLS_179 = {
        onClick: (...[$event]) => {
            __VLS_ctx.removeCategory(index);
        }
    };
    __VLS_175.slots.default;
    const __VLS_180 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_181 = __VLS_asFunctionalComponent(__VLS_180, new __VLS_180({}));
    const __VLS_182 = __VLS_181({}, ...__VLS_functionalComponentArgsRest(__VLS_181));
    __VLS_183.slots.default;
    const __VLS_184 = {}.Delete;
    /** @type {[typeof __VLS_components.Delete, ]} */ ;
    // @ts-ignore
    const __VLS_185 = __VLS_asFunctionalComponent(__VLS_184, new __VLS_184({}));
    const __VLS_186 = __VLS_185({}, ...__VLS_functionalComponentArgsRest(__VLS_185));
    var __VLS_183;
    var __VLS_175;
    var __VLS_139;
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "add-category" },
});
const __VLS_188 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_189 = __VLS_asFunctionalComponent(__VLS_188, new __VLS_188({
    ...{ 'onClick': {} },
    type: "primary",
    plain: true,
}));
const __VLS_190 = __VLS_189({
    ...{ 'onClick': {} },
    type: "primary",
    plain: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_189));
let __VLS_192;
let __VLS_193;
let __VLS_194;
const __VLS_195 = {
    onClick: (__VLS_ctx.addCategory)
};
__VLS_191.slots.default;
const __VLS_196 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_197 = __VLS_asFunctionalComponent(__VLS_196, new __VLS_196({}));
const __VLS_198 = __VLS_197({}, ...__VLS_functionalComponentArgsRest(__VLS_197));
__VLS_199.slots.default;
const __VLS_200 = {}.Plus;
/** @type {[typeof __VLS_components.Plus, ]} */ ;
// @ts-ignore
const __VLS_201 = __VLS_asFunctionalComponent(__VLS_200, new __VLS_200({}));
const __VLS_202 = __VLS_201({}, ...__VLS_functionalComponentArgsRest(__VLS_201));
var __VLS_199;
var __VLS_191;
{
    const { footer: __VLS_thisSlot } = __VLS_127.slots;
    const __VLS_204 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_205 = __VLS_asFunctionalComponent(__VLS_204, new __VLS_204({
        ...{ 'onClick': {} },
    }));
    const __VLS_206 = __VLS_205({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_205));
    let __VLS_208;
    let __VLS_209;
    let __VLS_210;
    const __VLS_211 = {
        onClick: (...[$event]) => {
            __VLS_ctx.categoryVisible = false;
        }
    };
    __VLS_207.slots.default;
    var __VLS_207;
    const __VLS_212 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_213 = __VLS_asFunctionalComponent(__VLS_212, new __VLS_212({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_214 = __VLS_213({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_213));
    let __VLS_216;
    let __VLS_217;
    let __VLS_218;
    const __VLS_219 = {
        onClick: (__VLS_ctx.handleSaveCategories)
    };
    __VLS_215.slots.default;
    var __VLS_215;
}
var __VLS_127;
const __VLS_220 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_221 = __VLS_asFunctionalComponent(__VLS_220, new __VLS_220({
    modelValue: (__VLS_ctx.formVisible),
    title: (__VLS_ctx.isEditing ? '编辑模板' : '新建模板'),
    width: "500px",
}));
const __VLS_222 = __VLS_221({
    modelValue: (__VLS_ctx.formVisible),
    title: (__VLS_ctx.isEditing ? '编辑模板' : '新建模板'),
    width: "500px",
}, ...__VLS_functionalComponentArgsRest(__VLS_221));
__VLS_223.slots.default;
const __VLS_224 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_225 = __VLS_asFunctionalComponent(__VLS_224, new __VLS_224({
    model: (__VLS_ctx.form),
    labelWidth: "80px",
}));
const __VLS_226 = __VLS_225({
    model: (__VLS_ctx.form),
    labelWidth: "80px",
}, ...__VLS_functionalComponentArgsRest(__VLS_225));
__VLS_227.slots.default;
const __VLS_228 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_229 = __VLS_asFunctionalComponent(__VLS_228, new __VLS_228({
    label: "模板名称",
    required: true,
}));
const __VLS_230 = __VLS_229({
    label: "模板名称",
    required: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_229));
__VLS_231.slots.default;
const __VLS_232 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_233 = __VLS_asFunctionalComponent(__VLS_232, new __VLS_232({
    modelValue: (__VLS_ctx.form.name),
    placeholder: "请输入模板名称",
}));
const __VLS_234 = __VLS_233({
    modelValue: (__VLS_ctx.form.name),
    placeholder: "请输入模板名称",
}, ...__VLS_functionalComponentArgsRest(__VLS_233));
var __VLS_231;
const __VLS_236 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_237 = __VLS_asFunctionalComponent(__VLS_236, new __VLS_236({
    label: "分类",
    required: true,
}));
const __VLS_238 = __VLS_237({
    label: "分类",
    required: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_237));
__VLS_239.slots.default;
const __VLS_240 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_241 = __VLS_asFunctionalComponent(__VLS_240, new __VLS_240({
    modelValue: (__VLS_ctx.form.category),
    placeholder: "请选择分类",
}));
const __VLS_242 = __VLS_241({
    modelValue: (__VLS_ctx.form.category),
    placeholder: "请选择分类",
}, ...__VLS_functionalComponentArgsRest(__VLS_241));
__VLS_243.slots.default;
for (const [cat] of __VLS_getVForSourceType((__VLS_ctx.categories))) {
    const __VLS_244 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_245 = __VLS_asFunctionalComponent(__VLS_244, new __VLS_244({
        key: (cat.value),
        label: (cat.label),
        value: (cat.value),
    }));
    const __VLS_246 = __VLS_245({
        key: (cat.value),
        label: (cat.label),
        value: (cat.value),
    }, ...__VLS_functionalComponentArgsRest(__VLS_245));
}
var __VLS_243;
var __VLS_239;
const __VLS_248 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_249 = __VLS_asFunctionalComponent(__VLS_248, new __VLS_248({
    label: "描述",
}));
const __VLS_250 = __VLS_249({
    label: "描述",
}, ...__VLS_functionalComponentArgsRest(__VLS_249));
__VLS_251.slots.default;
const __VLS_252 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_253 = __VLS_asFunctionalComponent(__VLS_252, new __VLS_252({
    modelValue: (__VLS_ctx.form.description),
    type: "textarea",
    rows: (3),
    placeholder: "请输入模板描述",
}));
const __VLS_254 = __VLS_253({
    modelValue: (__VLS_ctx.form.description),
    type: "textarea",
    rows: (3),
    placeholder: "请输入模板描述",
}, ...__VLS_functionalComponentArgsRest(__VLS_253));
var __VLS_251;
var __VLS_227;
{
    const { footer: __VLS_thisSlot } = __VLS_223.slots;
    const __VLS_256 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_257 = __VLS_asFunctionalComponent(__VLS_256, new __VLS_256({
        ...{ 'onClick': {} },
    }));
    const __VLS_258 = __VLS_257({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_257));
    let __VLS_260;
    let __VLS_261;
    let __VLS_262;
    const __VLS_263 = {
        onClick: (...[$event]) => {
            __VLS_ctx.formVisible = false;
        }
    };
    __VLS_259.slots.default;
    var __VLS_259;
    const __VLS_264 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_265 = __VLS_asFunctionalComponent(__VLS_264, new __VLS_264({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_266 = __VLS_265({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_265));
    let __VLS_268;
    let __VLS_269;
    let __VLS_270;
    const __VLS_271 = {
        onClick: (__VLS_ctx.handleSave)
    };
    __VLS_267.slots.default;
    var __VLS_267;
}
var __VLS_223;
const __VLS_272 = {}.ElDrawer;
/** @type {[typeof __VLS_components.ElDrawer, typeof __VLS_components.elDrawer, typeof __VLS_components.ElDrawer, typeof __VLS_components.elDrawer, ]} */ ;
// @ts-ignore
const __VLS_273 = __VLS_asFunctionalComponent(__VLS_272, new __VLS_272({
    modelValue: (__VLS_ctx.stepsVisible),
    title: (`编辑步骤 - ${__VLS_ctx.editingTemplateName}`),
    size: "900px",
}));
const __VLS_274 = __VLS_273({
    modelValue: (__VLS_ctx.stepsVisible),
    title: (`编辑步骤 - ${__VLS_ctx.editingTemplateName}`),
    size: "900px",
}, ...__VLS_functionalComponentArgsRest(__VLS_273));
__VLS_275.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_275.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "steps-drawer-header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.editingTemplateName);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "header-right" },
    });
    const __VLS_276 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_277 = __VLS_asFunctionalComponent(__VLS_276, new __VLS_276({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_278 = __VLS_277({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_277));
    let __VLS_280;
    let __VLS_281;
    let __VLS_282;
    const __VLS_283 = {
        onClick: (__VLS_ctx.openBatchImportDialog)
    };
    __VLS_279.slots.default;
    const __VLS_284 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_285 = __VLS_asFunctionalComponent(__VLS_284, new __VLS_284({}));
    const __VLS_286 = __VLS_285({}, ...__VLS_functionalComponentArgsRest(__VLS_285));
    __VLS_287.slots.default;
    const __VLS_288 = {}.Download;
    /** @type {[typeof __VLS_components.Download, ]} */ ;
    // @ts-ignore
    const __VLS_289 = __VLS_asFunctionalComponent(__VLS_288, new __VLS_288({}));
    const __VLS_290 = __VLS_289({}, ...__VLS_functionalComponentArgsRest(__VLS_289));
    var __VLS_287;
    var __VLS_279;
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "steps-editor" },
});
if (__VLS_ctx.editingSteps.length > 0) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "steps-table" },
    });
    const __VLS_292 = {}.ElTable;
    /** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
    // @ts-ignore
    const __VLS_293 = __VLS_asFunctionalComponent(__VLS_292, new __VLS_292({
        data: (__VLS_ctx.editingSteps),
        border: true,
        size: "small",
        rowKey: "id",
    }));
    const __VLS_294 = __VLS_293({
        data: (__VLS_ctx.editingSteps),
        border: true,
        size: "small",
        rowKey: "id",
    }, ...__VLS_functionalComponentArgsRest(__VLS_293));
    __VLS_295.slots.default;
    const __VLS_296 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_297 = __VLS_asFunctionalComponent(__VLS_296, new __VLS_296({
        type: "index",
        label: "序号",
        width: "60",
        align: "center",
    }));
    const __VLS_298 = __VLS_297({
        type: "index",
        label: "序号",
        width: "60",
        align: "center",
    }, ...__VLS_functionalComponentArgsRest(__VLS_297));
    const __VLS_300 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_301 = __VLS_asFunctionalComponent(__VLS_300, new __VLS_300({
        prop: "name",
        label: "步骤名称",
        minWidth: "150",
    }));
    const __VLS_302 = __VLS_301({
        prop: "name",
        label: "步骤名称",
        minWidth: "150",
    }, ...__VLS_functionalComponentArgsRest(__VLS_301));
    __VLS_303.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_303.slots;
        const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
        (row.name || '-');
    }
    var __VLS_303;
    const __VLS_304 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_305 = __VLS_asFunctionalComponent(__VLS_304, new __VLS_304({
        prop: "description",
        label: "描述",
        minWidth: "180",
        showOverflowTooltip: true,
    }));
    const __VLS_306 = __VLS_305({
        prop: "description",
        label: "描述",
        minWidth: "180",
        showOverflowTooltip: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_305));
    __VLS_307.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_307.slots;
        const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
        (row.description || '-');
    }
    var __VLS_307;
    const __VLS_308 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_309 = __VLS_asFunctionalComponent(__VLS_308, new __VLS_308({
        prop: "step_type",
        label: "类型",
        width: "80",
        align: "center",
    }));
    const __VLS_310 = __VLS_309({
        prop: "step_type",
        label: "类型",
        width: "80",
        align: "center",
    }, ...__VLS_functionalComponentArgsRest(__VLS_309));
    __VLS_311.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_311.slots;
        const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
        (__VLS_ctx.getStepTypeLabel(row.step_type));
    }
    var __VLS_311;
    const __VLS_312 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_313 = __VLS_asFunctionalComponent(__VLS_312, new __VLS_312({
        prop: "timeout_seconds",
        label: "超时(秒)",
        width: "90",
        align: "center",
    }));
    const __VLS_314 = __VLS_313({
        prop: "timeout_seconds",
        label: "超时(秒)",
        width: "90",
        align: "center",
    }, ...__VLS_functionalComponentArgsRest(__VLS_313));
    const __VLS_316 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_317 = __VLS_asFunctionalComponent(__VLS_316, new __VLS_316({
        prop: "assignee",
        label: "操作人",
        width: "100",
        showOverflowTooltip: true,
    }));
    const __VLS_318 = __VLS_317({
        prop: "assignee",
        label: "操作人",
        width: "100",
        showOverflowTooltip: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_317));
    __VLS_319.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_319.slots;
        const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
        (row.assignee || '-');
    }
    var __VLS_319;
    const __VLS_320 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_321 = __VLS_asFunctionalComponent(__VLS_320, new __VLS_320({
        label: "操作",
        width: "160",
        align: "center",
        fixed: "right",
    }));
    const __VLS_322 = __VLS_321({
        label: "操作",
        width: "160",
        align: "center",
        fixed: "right",
    }, ...__VLS_functionalComponentArgsRest(__VLS_321));
    __VLS_323.slots.default;
    {
        const { default: __VLS_thisSlot } = __VLS_323.slots;
        const [{ $index }] = __VLS_getSlotParams(__VLS_thisSlot);
        const __VLS_324 = {}.ElButtonGroup;
        /** @type {[typeof __VLS_components.ElButtonGroup, typeof __VLS_components.elButtonGroup, typeof __VLS_components.ElButtonGroup, typeof __VLS_components.elButtonGroup, ]} */ ;
        // @ts-ignore
        const __VLS_325 = __VLS_asFunctionalComponent(__VLS_324, new __VLS_324({
            size: "small",
        }));
        const __VLS_326 = __VLS_325({
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_325));
        __VLS_327.slots.default;
        const __VLS_328 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_329 = __VLS_asFunctionalComponent(__VLS_328, new __VLS_328({
            ...{ 'onClick': {} },
            text: true,
            type: "primary",
            title: "编辑",
        }));
        const __VLS_330 = __VLS_329({
            ...{ 'onClick': {} },
            text: true,
            type: "primary",
            title: "编辑",
        }, ...__VLS_functionalComponentArgsRest(__VLS_329));
        let __VLS_332;
        let __VLS_333;
        let __VLS_334;
        const __VLS_335 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.editingSteps.length > 0))
                    return;
                __VLS_ctx.openStepEditDialog($index);
            }
        };
        __VLS_331.slots.default;
        const __VLS_336 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_337 = __VLS_asFunctionalComponent(__VLS_336, new __VLS_336({}));
        const __VLS_338 = __VLS_337({}, ...__VLS_functionalComponentArgsRest(__VLS_337));
        __VLS_339.slots.default;
        const __VLS_340 = {}.Edit;
        /** @type {[typeof __VLS_components.Edit, ]} */ ;
        // @ts-ignore
        const __VLS_341 = __VLS_asFunctionalComponent(__VLS_340, new __VLS_340({}));
        const __VLS_342 = __VLS_341({}, ...__VLS_functionalComponentArgsRest(__VLS_341));
        var __VLS_339;
        var __VLS_331;
        const __VLS_344 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_345 = __VLS_asFunctionalComponent(__VLS_344, new __VLS_344({
            ...{ 'onClick': {} },
            disabled: ($index === 0),
            text: true,
            title: "上移",
        }));
        const __VLS_346 = __VLS_345({
            ...{ 'onClick': {} },
            disabled: ($index === 0),
            text: true,
            title: "上移",
        }, ...__VLS_functionalComponentArgsRest(__VLS_345));
        let __VLS_348;
        let __VLS_349;
        let __VLS_350;
        const __VLS_351 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.editingSteps.length > 0))
                    return;
                __VLS_ctx.moveStep($index, -1);
            }
        };
        __VLS_347.slots.default;
        const __VLS_352 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_353 = __VLS_asFunctionalComponent(__VLS_352, new __VLS_352({}));
        const __VLS_354 = __VLS_353({}, ...__VLS_functionalComponentArgsRest(__VLS_353));
        __VLS_355.slots.default;
        const __VLS_356 = {}.Top;
        /** @type {[typeof __VLS_components.Top, ]} */ ;
        // @ts-ignore
        const __VLS_357 = __VLS_asFunctionalComponent(__VLS_356, new __VLS_356({}));
        const __VLS_358 = __VLS_357({}, ...__VLS_functionalComponentArgsRest(__VLS_357));
        var __VLS_355;
        var __VLS_347;
        const __VLS_360 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_361 = __VLS_asFunctionalComponent(__VLS_360, new __VLS_360({
            ...{ 'onClick': {} },
            disabled: ($index === __VLS_ctx.editingSteps.length - 1),
            text: true,
            title: "下移",
        }));
        const __VLS_362 = __VLS_361({
            ...{ 'onClick': {} },
            disabled: ($index === __VLS_ctx.editingSteps.length - 1),
            text: true,
            title: "下移",
        }, ...__VLS_functionalComponentArgsRest(__VLS_361));
        let __VLS_364;
        let __VLS_365;
        let __VLS_366;
        const __VLS_367 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.editingSteps.length > 0))
                    return;
                __VLS_ctx.moveStep($index, 1);
            }
        };
        __VLS_363.slots.default;
        const __VLS_368 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_369 = __VLS_asFunctionalComponent(__VLS_368, new __VLS_368({}));
        const __VLS_370 = __VLS_369({}, ...__VLS_functionalComponentArgsRest(__VLS_369));
        __VLS_371.slots.default;
        const __VLS_372 = {}.Bottom;
        /** @type {[typeof __VLS_components.Bottom, ]} */ ;
        // @ts-ignore
        const __VLS_373 = __VLS_asFunctionalComponent(__VLS_372, new __VLS_372({}));
        const __VLS_374 = __VLS_373({}, ...__VLS_functionalComponentArgsRest(__VLS_373));
        var __VLS_371;
        var __VLS_363;
        const __VLS_376 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_377 = __VLS_asFunctionalComponent(__VLS_376, new __VLS_376({
            ...{ 'onClick': {} },
            text: true,
            type: "danger",
            title: "删除",
        }));
        const __VLS_378 = __VLS_377({
            ...{ 'onClick': {} },
            text: true,
            type: "danger",
            title: "删除",
        }, ...__VLS_functionalComponentArgsRest(__VLS_377));
        let __VLS_380;
        let __VLS_381;
        let __VLS_382;
        const __VLS_383 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.editingSteps.length > 0))
                    return;
                __VLS_ctx.removeStep($index);
            }
        };
        __VLS_379.slots.default;
        const __VLS_384 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_385 = __VLS_asFunctionalComponent(__VLS_384, new __VLS_384({}));
        const __VLS_386 = __VLS_385({}, ...__VLS_functionalComponentArgsRest(__VLS_385));
        __VLS_387.slots.default;
        const __VLS_388 = {}.Delete;
        /** @type {[typeof __VLS_components.Delete, ]} */ ;
        // @ts-ignore
        const __VLS_389 = __VLS_asFunctionalComponent(__VLS_388, new __VLS_388({}));
        const __VLS_390 = __VLS_389({}, ...__VLS_functionalComponentArgsRest(__VLS_389));
        var __VLS_387;
        var __VLS_379;
        var __VLS_327;
    }
    var __VLS_323;
    var __VLS_295;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "single-add-wrapper" },
    });
    const __VLS_392 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_393 = __VLS_asFunctionalComponent(__VLS_392, new __VLS_392({
        ...{ 'onClick': {} },
        type: "primary",
        plain: true,
    }));
    const __VLS_394 = __VLS_393({
        ...{ 'onClick': {} },
        type: "primary",
        plain: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_393));
    let __VLS_396;
    let __VLS_397;
    let __VLS_398;
    const __VLS_399 = {
        onClick: (__VLS_ctx.openSingleAddDialog)
    };
    __VLS_395.slots.default;
    const __VLS_400 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_401 = __VLS_asFunctionalComponent(__VLS_400, new __VLS_400({}));
    const __VLS_402 = __VLS_401({}, ...__VLS_functionalComponentArgsRest(__VLS_401));
    __VLS_403.slots.default;
    const __VLS_404 = {}.Plus;
    /** @type {[typeof __VLS_components.Plus, ]} */ ;
    // @ts-ignore
    const __VLS_405 = __VLS_asFunctionalComponent(__VLS_404, new __VLS_404({}));
    const __VLS_406 = __VLS_405({}, ...__VLS_functionalComponentArgsRest(__VLS_405));
    var __VLS_403;
    var __VLS_395;
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "steps-empty" },
    });
    const __VLS_408 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_409 = __VLS_asFunctionalComponent(__VLS_408, new __VLS_408({
        description: "暂无步骤",
    }));
    const __VLS_410 = __VLS_409({
        description: "暂无步骤",
    }, ...__VLS_functionalComponentArgsRest(__VLS_409));
    __VLS_411.slots.default;
    const __VLS_412 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_413 = __VLS_asFunctionalComponent(__VLS_412, new __VLS_412({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_414 = __VLS_413({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_413));
    let __VLS_416;
    let __VLS_417;
    let __VLS_418;
    const __VLS_419 = {
        onClick: (__VLS_ctx.openBatchImportDialog)
    };
    __VLS_415.slots.default;
    const __VLS_420 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_421 = __VLS_asFunctionalComponent(__VLS_420, new __VLS_420({}));
    const __VLS_422 = __VLS_421({}, ...__VLS_functionalComponentArgsRest(__VLS_421));
    __VLS_423.slots.default;
    const __VLS_424 = {}.Download;
    /** @type {[typeof __VLS_components.Download, ]} */ ;
    // @ts-ignore
    const __VLS_425 = __VLS_asFunctionalComponent(__VLS_424, new __VLS_424({}));
    const __VLS_426 = __VLS_425({}, ...__VLS_functionalComponentArgsRest(__VLS_425));
    var __VLS_423;
    var __VLS_415;
    var __VLS_411;
}
{
    const { footer: __VLS_thisSlot } = __VLS_275.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "drawer-footer" },
    });
    const __VLS_428 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_429 = __VLS_asFunctionalComponent(__VLS_428, new __VLS_428({
        ...{ 'onClick': {} },
    }));
    const __VLS_430 = __VLS_429({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_429));
    let __VLS_432;
    let __VLS_433;
    let __VLS_434;
    const __VLS_435 = {
        onClick: (...[$event]) => {
            __VLS_ctx.stepsVisible = false;
        }
    };
    __VLS_431.slots.default;
    var __VLS_431;
    const __VLS_436 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_437 = __VLS_asFunctionalComponent(__VLS_436, new __VLS_436({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_438 = __VLS_437({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_437));
    let __VLS_440;
    let __VLS_441;
    let __VLS_442;
    const __VLS_443 = {
        onClick: (__VLS_ctx.handleSaveSteps)
    };
    __VLS_439.slots.default;
    var __VLS_439;
}
var __VLS_275;
const __VLS_444 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_445 = __VLS_asFunctionalComponent(__VLS_444, new __VLS_444({
    modelValue: (__VLS_ctx.importVisible),
    title: "批量导入步骤",
    width: "520px",
}));
const __VLS_446 = __VLS_445({
    modelValue: (__VLS_ctx.importVisible),
    title: "批量导入步骤",
    width: "520px",
}, ...__VLS_functionalComponentArgsRest(__VLS_445));
__VLS_447.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "excel-upload" },
});
const __VLS_448 = {}.ElUpload;
/** @type {[typeof __VLS_components.ElUpload, typeof __VLS_components.elUpload, typeof __VLS_components.ElUpload, typeof __VLS_components.elUpload, ]} */ ;
// @ts-ignore
const __VLS_449 = __VLS_asFunctionalComponent(__VLS_448, new __VLS_448({
    ref: "uploadRef",
    beforeUpload: (__VLS_ctx.handleExcelUpload),
    showFileList: (false),
    accept: ".xlsx,.xls",
    ...{ class: "upload-area" },
}));
const __VLS_450 = __VLS_449({
    ref: "uploadRef",
    beforeUpload: (__VLS_ctx.handleExcelUpload),
    showFileList: (false),
    accept: ".xlsx,.xls",
    ...{ class: "upload-area" },
}, ...__VLS_functionalComponentArgsRest(__VLS_449));
/** @type {typeof __VLS_ctx.uploadRef} */ ;
var __VLS_452 = {};
__VLS_451.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "upload-content" },
});
const __VLS_454 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_455 = __VLS_asFunctionalComponent(__VLS_454, new __VLS_454({
    ...{ class: "upload-icon" },
}));
const __VLS_456 = __VLS_455({
    ...{ class: "upload-icon" },
}, ...__VLS_functionalComponentArgsRest(__VLS_455));
__VLS_457.slots.default;
const __VLS_458 = {}.Upload;
/** @type {[typeof __VLS_components.Upload, ]} */ ;
// @ts-ignore
const __VLS_459 = __VLS_asFunctionalComponent(__VLS_458, new __VLS_458({}));
const __VLS_460 = __VLS_459({}, ...__VLS_functionalComponentArgsRest(__VLS_459));
var __VLS_457;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "upload-text" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "upload-hint" },
});
var __VLS_451;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "template-download" },
});
const __VLS_462 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_463 = __VLS_asFunctionalComponent(__VLS_462, new __VLS_462({
    ...{ 'onClick': {} },
    type: "info",
    plain: true,
}));
const __VLS_464 = __VLS_463({
    ...{ 'onClick': {} },
    type: "info",
    plain: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_463));
let __VLS_466;
let __VLS_467;
let __VLS_468;
const __VLS_469 = {
    onClick: (__VLS_ctx.downloadTemplate)
};
__VLS_465.slots.default;
const __VLS_470 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_471 = __VLS_asFunctionalComponent(__VLS_470, new __VLS_470({}));
const __VLS_472 = __VLS_471({}, ...__VLS_functionalComponentArgsRest(__VLS_471));
__VLS_473.slots.default;
const __VLS_474 = {}.Download;
/** @type {[typeof __VLS_components.Download, ]} */ ;
// @ts-ignore
const __VLS_475 = __VLS_asFunctionalComponent(__VLS_474, new __VLS_474({}));
const __VLS_476 = __VLS_475({}, ...__VLS_functionalComponentArgsRest(__VLS_475));
var __VLS_473;
var __VLS_465;
var __VLS_447;
const __VLS_478 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_479 = __VLS_asFunctionalComponent(__VLS_478, new __VLS_478({
    modelValue: (__VLS_ctx.singleAddVisible),
    title: "添加步骤",
    width: "520px",
}));
const __VLS_480 = __VLS_479({
    modelValue: (__VLS_ctx.singleAddVisible),
    title: "添加步骤",
    width: "520px",
}, ...__VLS_functionalComponentArgsRest(__VLS_479));
__VLS_481.slots.default;
const __VLS_482 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_483 = __VLS_asFunctionalComponent(__VLS_482, new __VLS_482({
    model: (__VLS_ctx.singleStepForm),
    labelWidth: "90px",
    ...{ class: "single-step-form" },
}));
const __VLS_484 = __VLS_483({
    model: (__VLS_ctx.singleStepForm),
    labelWidth: "90px",
    ...{ class: "single-step-form" },
}, ...__VLS_functionalComponentArgsRest(__VLS_483));
__VLS_485.slots.default;
const __VLS_486 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_487 = __VLS_asFunctionalComponent(__VLS_486, new __VLS_486({
    label: "步骤名称",
    required: true,
}));
const __VLS_488 = __VLS_487({
    label: "步骤名称",
    required: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_487));
__VLS_489.slots.default;
const __VLS_490 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_491 = __VLS_asFunctionalComponent(__VLS_490, new __VLS_490({
    modelValue: (__VLS_ctx.singleStepForm.name),
    placeholder: "请输入步骤名称",
    maxlength: "100",
    showWordLimit: true,
}));
const __VLS_492 = __VLS_491({
    modelValue: (__VLS_ctx.singleStepForm.name),
    placeholder: "请输入步骤名称",
    maxlength: "100",
    showWordLimit: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_491));
var __VLS_489;
const __VLS_494 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_495 = __VLS_asFunctionalComponent(__VLS_494, new __VLS_494({
    label: "描述",
}));
const __VLS_496 = __VLS_495({
    label: "描述",
}, ...__VLS_functionalComponentArgsRest(__VLS_495));
__VLS_497.slots.default;
const __VLS_498 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_499 = __VLS_asFunctionalComponent(__VLS_498, new __VLS_498({
    modelValue: (__VLS_ctx.singleStepForm.description),
    type: "textarea",
    placeholder: "步骤描述",
    rows: (2),
    maxlength: "500",
    showWordLimit: true,
}));
const __VLS_500 = __VLS_499({
    modelValue: (__VLS_ctx.singleStepForm.description),
    type: "textarea",
    placeholder: "步骤描述",
    rows: (2),
    maxlength: "500",
    showWordLimit: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_499));
var __VLS_497;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "form-row" },
});
const __VLS_502 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_503 = __VLS_asFunctionalComponent(__VLS_502, new __VLS_502({
    label: "步骤类型",
}));
const __VLS_504 = __VLS_503({
    label: "步骤类型",
}, ...__VLS_functionalComponentArgsRest(__VLS_503));
__VLS_505.slots.default;
const __VLS_506 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_507 = __VLS_asFunctionalComponent(__VLS_506, new __VLS_506({
    modelValue: (__VLS_ctx.singleStepForm.step_type),
}));
const __VLS_508 = __VLS_507({
    modelValue: (__VLS_ctx.singleStepForm.step_type),
}, ...__VLS_functionalComponentArgsRest(__VLS_507));
__VLS_509.slots.default;
const __VLS_510 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_511 = __VLS_asFunctionalComponent(__VLS_510, new __VLS_510({
    label: "串行",
    value: "serial",
}));
const __VLS_512 = __VLS_511({
    label: "串行",
    value: "serial",
}, ...__VLS_functionalComponentArgsRest(__VLS_511));
const __VLS_514 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_515 = __VLS_asFunctionalComponent(__VLS_514, new __VLS_514({
    label: "并行",
    value: "parallel",
}));
const __VLS_516 = __VLS_515({
    label: "并行",
    value: "parallel",
}, ...__VLS_functionalComponentArgsRest(__VLS_515));
const __VLS_518 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_519 = __VLS_asFunctionalComponent(__VLS_518, new __VLS_518({
    label: "任选",
    value: "any_of",
}));
const __VLS_520 = __VLS_519({
    label: "任选",
    value: "any_of",
}, ...__VLS_functionalComponentArgsRest(__VLS_519));
const __VLS_522 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_523 = __VLS_asFunctionalComponent(__VLS_522, new __VLS_522({
    label: "条件",
    value: "condition",
}));
const __VLS_524 = __VLS_523({
    label: "条件",
    value: "condition",
}, ...__VLS_functionalComponentArgsRest(__VLS_523));
var __VLS_509;
var __VLS_505;
const __VLS_526 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_527 = __VLS_asFunctionalComponent(__VLS_526, new __VLS_526({
    label: "超时时间",
}));
const __VLS_528 = __VLS_527({
    label: "超时时间",
}, ...__VLS_functionalComponentArgsRest(__VLS_527));
__VLS_529.slots.default;
const __VLS_530 = {}.ElInputNumber;
/** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
// @ts-ignore
const __VLS_531 = __VLS_asFunctionalComponent(__VLS_530, new __VLS_530({
    modelValue: (__VLS_ctx.singleStepForm.timeout_seconds),
    min: (30),
    max: (3600),
    controlsPosition: "right",
}));
const __VLS_532 = __VLS_531({
    modelValue: (__VLS_ctx.singleStepForm.timeout_seconds),
    min: (30),
    max: (3600),
    controlsPosition: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_531));
var __VLS_529;
const __VLS_534 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_535 = __VLS_asFunctionalComponent(__VLS_534, new __VLS_534({
    label: "操作人",
}));
const __VLS_536 = __VLS_535({
    label: "操作人",
}, ...__VLS_functionalComponentArgsRest(__VLS_535));
__VLS_537.slots.default;
const __VLS_538 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_539 = __VLS_asFunctionalComponent(__VLS_538, new __VLS_538({
    modelValue: (__VLS_ctx.singleStepForm.assignee),
    placeholder: "填写操作人",
}));
const __VLS_540 = __VLS_539({
    modelValue: (__VLS_ctx.singleStepForm.assignee),
    placeholder: "填写操作人",
}, ...__VLS_functionalComponentArgsRest(__VLS_539));
var __VLS_537;
var __VLS_485;
{
    const { footer: __VLS_thisSlot } = __VLS_481.slots;
    const __VLS_542 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_543 = __VLS_asFunctionalComponent(__VLS_542, new __VLS_542({
        ...{ 'onClick': {} },
    }));
    const __VLS_544 = __VLS_543({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_543));
    let __VLS_546;
    let __VLS_547;
    let __VLS_548;
    const __VLS_549 = {
        onClick: (...[$event]) => {
            __VLS_ctx.singleAddVisible = false;
        }
    };
    __VLS_545.slots.default;
    var __VLS_545;
    const __VLS_550 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_551 = __VLS_asFunctionalComponent(__VLS_550, new __VLS_550({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_552 = __VLS_551({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_551));
    let __VLS_554;
    let __VLS_555;
    let __VLS_556;
    const __VLS_557 = {
        onClick: (__VLS_ctx.handleAddSingleStep)
    };
    __VLS_553.slots.default;
    var __VLS_553;
}
var __VLS_481;
const __VLS_558 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_559 = __VLS_asFunctionalComponent(__VLS_558, new __VLS_558({
    modelValue: (__VLS_ctx.deleteVisible),
    title: "确认删除",
    width: "400px",
}));
const __VLS_560 = __VLS_559({
    modelValue: (__VLS_ctx.deleteVisible),
    title: "确认删除",
    width: "400px",
}, ...__VLS_functionalComponentArgsRest(__VLS_559));
__VLS_561.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
(__VLS_ctx.deleteTarget?.name);
{
    const { footer: __VLS_thisSlot } = __VLS_561.slots;
    const __VLS_562 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_563 = __VLS_asFunctionalComponent(__VLS_562, new __VLS_562({
        ...{ 'onClick': {} },
    }));
    const __VLS_564 = __VLS_563({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_563));
    let __VLS_566;
    let __VLS_567;
    let __VLS_568;
    const __VLS_569 = {
        onClick: (...[$event]) => {
            __VLS_ctx.deleteVisible = false;
        }
    };
    __VLS_565.slots.default;
    var __VLS_565;
    const __VLS_570 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_571 = __VLS_asFunctionalComponent(__VLS_570, new __VLS_570({
        ...{ 'onClick': {} },
        type: "danger",
    }));
    const __VLS_572 = __VLS_571({
        ...{ 'onClick': {} },
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_571));
    let __VLS_574;
    let __VLS_575;
    let __VLS_576;
    const __VLS_577 = {
        onClick: (__VLS_ctx.confirmDelete)
    };
    __VLS_573.slots.default;
    var __VLS_573;
}
var __VLS_561;
/** @type {__VLS_StyleScopedClasses['page-container']} */ ;
/** @type {__VLS_StyleScopedClasses['page-header']} */ ;
/** @type {__VLS_StyleScopedClasses['page-title']} */ ;
/** @type {__VLS_StyleScopedClasses['header-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['page-content']} */ ;
/** @type {__VLS_StyleScopedClasses['category-tabs']} */ ;
/** @type {__VLS_StyleScopedClasses['templates-table']} */ ;
/** @type {__VLS_StyleScopedClasses['category-list']} */ ;
/** @type {__VLS_StyleScopedClasses['category-item']} */ ;
/** @type {__VLS_StyleScopedClasses['category-input']} */ ;
/** @type {__VLS_StyleScopedClasses['add-category']} */ ;
/** @type {__VLS_StyleScopedClasses['steps-drawer-header']} */ ;
/** @type {__VLS_StyleScopedClasses['header-right']} */ ;
/** @type {__VLS_StyleScopedClasses['steps-editor']} */ ;
/** @type {__VLS_StyleScopedClasses['steps-table']} */ ;
/** @type {__VLS_StyleScopedClasses['single-add-wrapper']} */ ;
/** @type {__VLS_StyleScopedClasses['steps-empty']} */ ;
/** @type {__VLS_StyleScopedClasses['drawer-footer']} */ ;
/** @type {__VLS_StyleScopedClasses['excel-upload']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-area']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-content']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-icon']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-text']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-hint']} */ ;
/** @type {__VLS_StyleScopedClasses['template-download']} */ ;
/** @type {__VLS_StyleScopedClasses['single-step-form']} */ ;
/** @type {__VLS_StyleScopedClasses['form-row']} */ ;
// @ts-ignore
var __VLS_453 = __VLS_452;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Refresh: Refresh,
            Plus: Plus,
            Delete: Delete,
            Setting: Setting,
            Upload: Upload,
            Download: Download,
            Top: Top,
            Bottom: Bottom,
            Edit: Edit,
            activeCategory: activeCategory,
            categories: categories,
            formVisible: formVisible,
            stepsVisible: stepsVisible,
            deleteVisible: deleteVisible,
            categoryVisible: categoryVisible,
            importVisible: importVisible,
            editableCategories: editableCategories,
            isEditing: isEditing,
            editingSteps: editingSteps,
            editingTemplateName: editingTemplateName,
            deleteTarget: deleteTarget,
            filteredTemplates: filteredTemplates,
            form: form,
            singleStepForm: singleStepForm,
            singleAddVisible: singleAddVisible,
            openBatchImportDialog: openBatchImportDialog,
            openSingleAddDialog: openSingleAddDialog,
            handleAddSingleStep: handleAddSingleStep,
            moveStep: moveStep,
            openStepEditDialog: openStepEditDialog,
            getCategoryLabel: getCategoryLabel,
            getCategoryTagType: getCategoryTagType,
            getStepTypeLabel: getStepTypeLabel,
            formatTime: formatTime,
            loadTemplates: loadTemplates,
            openCategoryDialog: openCategoryDialog,
            addCategory: addCategory,
            removeCategory: removeCategory,
            moveCategory: moveCategory,
            handleSaveCategories: handleSaveCategories,
            openCreateDialog: openCreateDialog,
            openEditDialog: openEditDialog,
            handleSave: handleSave,
            openStepsDialog: openStepsDialog,
            removeStep: removeStep,
            handleSaveSteps: handleSaveSteps,
            downloadTemplate: downloadTemplate,
            handleExcelUpload: handleExcelUpload,
            handleDelete: handleDelete,
            confirmDelete: confirmDelete,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
