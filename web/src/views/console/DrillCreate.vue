<template>
  <div class="drill-create-page">
    <div class="page-header">
      <div class="header-left">
        <el-page-header @back="handleBack">
          <template #title>
            <span class="page-title">创建演练</span>
          </template>
        </el-page-header>
      </div>
    </div>

    <div class="page-content">
      <el-card class="form-card">
        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-width="120px"
          class="create-form"
        >
          <el-form-item label="演练名称" prop="name">
            <el-input
              v-model="form.name"
              placeholder="请输入演练名称"
              maxlength="50"
              show-word-limit
            />
          </el-form-item>

          <el-form-item label="演练模板" prop="templateId">
            <el-select
              v-model="form.templateId"
              placeholder="请选择演练模板"
              style="width: 100%"
              @change="handleTemplateChange"
            >
              <el-option
                v-for="template in templates"
                :key="template.id"
                :label="template.name"
                :value="template.id"
              >
                <span>{{ template.name }}</span>
                <el-tag size="small" style="margin-left: 8px">
                  {{ getCategoryText(template.category) }}
                </el-tag>
              </el-option>
            </el-select>
          </el-form-item>

          <el-form-item label="演练分类" prop="category">
            <el-select v-model="form.category" placeholder="请选择演练分类" style="width: 100%">
              <el-option label="灾备切换" value="disaster_recovery" />
              <el-option label="服务降级" value="degradation" />
              <el-option label="发布回滚" value="rollback" />
              <el-option label="安全演练" value="security" />
            </el-select>
          </el-form-item>

          <el-form-item label="计划开始时间" prop="scheduledStartTime">
            <el-date-picker
              v-model="form.scheduledStartTime"
              type="datetime"
              placeholder="选择开始时间"
              style="width: 100%"
              :disabled-date="disabledDate"
            />
          </el-form-item>

          <el-form-item label="预计时长 (分钟)" prop="estimatedDuration">
            <el-input-number
              v-model="form.estimatedDuration"
              :min="5"
              :max="1440"
              :step="5"
              style="width: 100%"
            />
          </el-form-item>

          <el-form-item label="参演人员" prop="assigneeIds">
            <el-select
              v-model="form.assigneeIds"
              multiple
              placeholder="请选择参演人员"
              style="width: 100%"
            >
              <el-option
                v-for="user in users"
                :key="user.id"
                :label="user.name"
                :value="user.id"
              >
                <span>{{ user.name }}</span>
                <span style="color: var(--el-text-color-secondary); margin-left: 8px">
                  {{ getRoleText(user.role) }}
                </span>
              </el-option>
            </el-select>
          </el-form-item>

          <el-form-item label="备注说明" prop="remark">
            <el-input
              v-model="form.remark"
              type="textarea"
              :rows="4"
              placeholder="请输入备注说明（可选）"
              maxlength="500"
              show-word-limit
            />
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="handleSubmit" :loading="submitting">
              创建演练
            </el-button>
            <el-button @click="handleBack">取消</el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <!-- 模板预览 -->
      <el-card v-if="selectedTemplate" class="preview-card">
        <template #header>
          <span class="card-title">模板预览</span>
        </template>
        <div class="template-info">
          <div class="info-row">
            <span class="info-label">模板描述：</span>
            <span class="info-value">{{ selectedTemplate.description }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">步骤数量：</span>
            <span class="info-value">{{ selectedTemplate.stepCount }} 个步骤</span>
          </div>
          <div class="info-row">
            <span class="info-label">预计时长：</span>
            <span class="info-value">{{ selectedTemplate.estimatedDuration }} 分钟</span>
          </div>
        </div>
        <div class="step-list">
          <div class="step-list-title">流程步骤</div>
          <div
            v-for="(step, index) in selectedTemplate.steps"
            :key="step.id"
            class="step-item"
          >
            <span class="step-index">{{ index + 1 }}</span>
            <span class="step-name">{{ step.name }}</span>
            <el-tag size="small" :type="getStepType(step.stepType)">
              {{ getStepTypeText(step.stepType) }}
            </el-tag>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'

interface Template {
  id: number
  name: string
  category: string
  description: string
  stepCount: number
  estimatedDuration: number
  steps: Array<{
    id: number
    name: string
    stepType: string
  }>
}

interface User {
  id: number
  name: string
  role: string
}

const router = useRouter()
const route = useRoute()

const formRef = ref<FormInstance>()
const submitting = ref(false)

const form = reactive({
  name: '',
  templateId: 0,
  category: '',
  scheduledStartTime: null as Date | null,
  estimatedDuration: 60,
  assigneeIds: [] as number[],
  remark: ''
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入演练名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  templateId: [
    { required: true, message: '请选择演练模板', trigger: 'change' }
  ],
  category: [
    { required: true, message: '请选择演练分类', trigger: 'change' }
  ],
  estimatedDuration: [
    { required: true, message: '请输入预计时长', trigger: 'blur' }
  ]
}

// 模拟数据
const templates = ref<Template[]>([
  {
    id: 1,
    name: 'DB 主从切换模板',
    category: 'disaster_recovery',
    description: '数据库主从切换标准流程',
    stepCount: 7,
    estimatedDuration: 120,
    steps: [
      { id: 1, name: '开始', stepType: 'serial' },
      { id: 2, name: '备份检查', stepType: 'serial' },
      { id: 3, name: '主库降级', stepType: 'serial' },
      { id: 4, name: '从库升级', stepType: 'serial' },
      { id: 5, name: '流量切换', stepType: 'serial' },
      { id: 6, name: '验证测试', stepType: 'serial' },
      { id: 7, name: '结束', stepType: 'serial' }
    ]
  },
  {
    id: 2,
    name: '服务降级标准流程',
    category: 'degradation',
    description: '服务降级和熔断标准流程',
    stepCount: 5,
    estimatedDuration: 60,
    steps: [
      { id: 1, name: '开始', stepType: 'serial' },
      { id: 2, name: '流量评估', stepType: 'serial' },
      { id: 3, name: '降级执行', stepType: 'parallel' },
      { id: 4, name: '验证确认', stepType: 'serial' },
      { id: 5, name: '结束', stepType: 'serial' }
    ]
  }
])

const users = ref<User[]>([
  { id: 1, name: '张三', role: 'admin' },
  { id: 2, name: '李四', role: 'director' },
  { id: 3, name: '王五', role: 'executor' },
  { id: 4, name: '赵六', role: 'executor' },
  { id: 5, name: '钱七', role: 'viewer' }
])

const selectedTemplate = computed(() => {
  return templates.value.find(t => t.id === form.templateId)
})

// 获取分类文本
const getCategoryText = (category: string) => {
  const textMap: Record<string, string> = {
    disaster_recovery: '灾备切换',
    degradation: '服务降级',
    rollback: '发布回滚',
    security: '安全演练'
  }
  return textMap[category] || category
}

// 获取角色文本
const getRoleText = (role: string) => {
  const textMap: Record<string, string> = {
    admin: '指挥员',
    director: '导演',
    executor: '执行者',
    viewer: '观察者'
  }
  return textMap[role] || role
}

// 获取步骤类型
const getStepType = (type: string) => {
  const typeMap: Record<string, any> = {
    serial: '',
    parallel: 'warning',
    any_of: 'success',
    condition: 'info'
  }
  return typeMap[type] || ''
}

// 获取步骤类型文本
const getStepTypeText = (type: string) => {
  const textMap: Record<string, string> = {
    serial: '串行',
    parallel: '并行',
    any_of: '或签',
    condition: '条件'
  }
  return textMap[type] || type
}

// 禁用日期
const disabledDate = (date: Date) => {
  return date.getTime() < Date.now() - 86400000
}

// 模板变化
const handleTemplateChange = () => {
  if (selectedTemplate.value) {
    form.category = selectedTemplate.value.category
    form.estimatedDuration = selectedTemplate.value.estimatedDuration
  }
}

// 提交
const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      // 模拟 API 调用
      await new Promise(resolve => setTimeout(resolve, 1000))
      ElMessage.success('演练创建成功')
      router.push('/console')
    } catch (error) {
      ElMessage.error('创建失败，请重试')
    } finally {
      submitting.value = false
    }
  })
}

// 返回
const handleBack = () => {
  router.back()
}
</script>

<style scoped>
.drill-create-page {
  padding: 24px;
  background-color: var(--color-background, #020617);
  min-height: 100vh;
}

.page-header {
  margin-bottom: 24px;
}

.page-title {
  font-size: 18px;
  font-weight: 600;
}

.page-content {
  display: grid;
  grid-template-columns: 1fr 400px;
  gap: 20px;
}

.form-card,
.preview-card {
  background-color: var(--color-muted, #1A1E2F);
  border: 1px solid var(--color-border, #334155);
}

:deep(.el-card__header) {
  background-color: var(--color-secondary, #1E293B);
  border-bottom: 1px solid var(--color-border, #334155);
}

.card-title {
  font-weight: 600;
  color: var(--color-foreground, #F8FAFC);
}

.create-form {
  max-width: 600px;
}

.template-info {
  margin-bottom: 20px;
}

.info-row {
  display: flex;
  margin-bottom: 12px;
}

.info-label {
  width: 100px;
  color: var(--color-muted-foreground, #94A3B8);
  flex-shrink: 0;
}

.info-value {
  color: var(--color-foreground, #F8FAFC);
}

.step-list {
  border-top: 1px solid var(--color-border, #334155);
  padding-top: 16px;
}

.step-list-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 12px;
  color: var(--color-foreground, #F8FAFC);
}

.step-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
}

.step-index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background-color: var(--color-secondary, #1E293B);
  color: var(--color-foreground, #F8FAFC);
  font-size: 12px;
  font-weight: 600;
}

.step-name {
  flex: 1;
  color: var(--color-foreground, #F8FAFC);
}

@media (max-width: 1024px) {
  .page-content {
    grid-template-columns: 1fr;
  }
}
</style>
