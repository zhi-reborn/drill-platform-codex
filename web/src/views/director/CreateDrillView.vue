<template>
  <div class="page-container">
    <div class="page-content">
      <el-steps :active="currentStep" finish-status="success" align-center class="create-steps">
        <el-step title="选择模板" />
        <el-step title="配置基本信息" />
        <el-step title="确认创建" />
      </el-steps>

      <!-- 步骤 1: 选择模板 -->
      <div v-show="currentStep === 0" class="step-content">
        <el-row :gutter="20">
          <el-col
            v-for="template in templates"
            :key="template.id"
            :xs="24"
            :sm="12"
            :lg="8"
          >
            <el-card
              class="template-card"
              :class="{ selected: selectedTemplate?.id === template.id }"
              @click="selectTemplate(template)"
            >
              <div class="template-name">{{ template.name }}</div>
              <div class="template-description">{{ template.description }}</div>
              <div class="template-meta">
                <el-tag :type="getCategoryTagType(template.category)" size="small">
                  {{ getCategoryLabel(template.category) }}
                </el-tag>
                <span class="steps-count">共 {{ template.steps.length }} 步</span>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <!-- 步骤 2: 配置基本信息 -->
      <div v-show="currentStep === 1" class="step-content">
        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-width="100px"
          class="create-form"
        >
          <el-form-item label="演练名称" prop="name">
            <el-input v-model="form.name" placeholder="请输入演练名称" maxlength="50" show-word-limit />
          </el-form-item>
          <el-form-item label="描述" prop="description">
            <el-input
              v-model="form.description"
              type="textarea"
              :rows="4"
              placeholder="请输入演练描述（可选）"
              maxlength="200"
              show-word-limit
            />
          </el-form-item>
        </el-form>
      </div>

      <!-- 步骤 3: 确认创建 -->
      <div v-show="currentStep === 2" class="step-content">
        <el-card class="confirm-card">
          <h3>信息摘要</h3>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="选择模板">
              {{ selectedTemplate?.name }}
            </el-descriptions-item>
            <el-descriptions-item label="演练名称">
              {{ form.name }}
            </el-descriptions-item>
            <el-descriptions-item label="演练描述">
              {{ form.description || '无' }}
            </el-descriptions-item>
            <el-descriptions-item label="步骤数量">
              {{ selectedTemplate?.steps.length }} 个步骤
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </div>

      <!-- 导航按钮 -->
      <div class="step-actions">
        <el-button v-if="currentStep > 0" @click="prevStep">
          上一步
        </el-button>
        <el-button v-if="currentStep < 2" type="primary" @click="nextStep">
          下一步
        </el-button>
        <el-button v-else type="primary" @click="confirmCreate">
          确认创建
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import type { DrillTemplate } from '@/types'
import { templateApi } from '@/api/modules/template'
import { drillApi } from '@/api/modules/drill'

const router = useRouter()

const currentStep = ref(0)
const templates = ref<DrillTemplate[]>([])
const selectedTemplate = ref<DrillTemplate | null>(null)
const formRef = ref<FormInstance>()

const form = reactive({
  name: '',
  description: '',
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入演练名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' },
  ],
}

async function loadTemplates() {
  try {
    const result = await templateApi.getList({ page: 1, page_size: 100 })
    // 只显示已启用的模板（status=2 或 status_label='enabled'）
    templates.value = (result.list || []).filter(t => t.status === 2 || t.status_label === 'enabled')
  } catch (error) {
    ElMessage.error('加载模板失败')
    console.error('Failed to load templates:', error)
  }
}

function getCategoryLabel(category: string): string {
  const map: Record<string, string> = {
    disaster_recovery: '灾备切换',
    degradation: '服务降级',
    release: '发布演练',
    security: '安全事件',
  }
  return map[category] || category
}

function getCategoryTagType(category: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, any> = {
    disaster_recovery: 'primary',
    degradation: 'warning',
    release: 'success',
    security: 'danger',
  }
  return map[category] || 'info'
}

function selectTemplate(template: DrillTemplate) {
  selectedTemplate.value = template
}

async function prevStep() {
  currentStep.value--
}

async function nextStep() {
  if (currentStep.value === 0) {
    if (!selectedTemplate.value) {
      ElMessage.warning('请先选择一个模板')
      return
    }
  }
  if (currentStep.value === 1) {
    if (!formRef.value) return
    await formRef.value.validate(valid => {
      if (valid) {
        currentStep.value++
      } else {
        ElMessage.warning('请填写必填项')
      }
    })
    return
  }
  currentStep.value++
}

async function confirmCreate() {
  if (!selectedTemplate.value || !form.name.trim()) {
    ElMessage.warning('信息不完整，无法创建')
    return
  }
  try {
    await drillApi.create({
      template_id: selectedTemplate.value.id,
      name: form.name.trim(),
      description: form.description?.trim() || '',
    })
    ElMessage.success('演练创建成功')
    router.push('/director/drills')
  } catch (error) {
    ElMessage.error('创建失败')
    console.error('Failed to create drill:', error)
  }
}

loadTemplates()
</script>

<style scoped lang="scss">
@use '@/styles/layout' as *;
@use '@/styles/variables' as *;

.page-container {
  @include page-container;

  .page-header {
    @include page-header;

    .page-title {
      font-size: $font-size-xl;
      font-weight: $font-weight-bold;
      color: $text-primary;
      margin: 0;
    }
  }

  .page-content {
    @include page-content;

    .create-steps {
      margin-bottom: $spacing-base;
      background: $bg-secondary;
      padding: $spacing-base;
      border-radius: $radius-base;
    }

    .step-content {
      min-height: 200px;
      margin-bottom: $spacing-base;

      .template-card {
        background: $bg-secondary;
        border-color: $border-color;
        cursor: pointer;
        transition: all 0.2s;
        margin-bottom: $spacing-sm;
        border-width: 2px;

        &:hover {
          border-color: $color-accent;
          box-shadow: $shadow-sm;
        }

        &.selected {
          border-color: $color-accent;
          background: $color-accent-bg;
        }

        .template-name {
          font-size: $font-size-base;
          font-weight: $font-weight-semibold;
          color: $text-primary;
          margin-bottom: $spacing-xs;
        }

        .template-description {
          font-size: $font-size-sm;
          color: $text-secondary;
          margin-bottom: $spacing-sm;
          line-height: 1.6;
        }

        .template-meta {
          display: flex;
          justify-content: space-between;
          align-items: center;

          .steps-count {
            font-size: $font-size-xs;
            color: $text-tertiary;
          }
        }
      }

      .create-form {
        max-width: 600px;
        margin: 0 auto;
        background: $bg-secondary;
        padding: $spacing-base;
        border-radius: $radius-base;
      }

      .confirm-card {
        max-width: 600px;
        margin: 0 auto;
        background: $bg-secondary;
        border-color: $border-color;

        h3 {
          font-size: $font-size-md;
          color: $text-primary;
          margin-bottom: $spacing-base;
        }
      }
    }

    .step-actions {
      text-align: center;
      padding-top: $spacing-base;
      border-top: 1px solid $border-color;
    }
  }
}
</style>
