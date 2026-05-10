<template>
  <div class="drill-create-page">
    <h1 class="page-title">创建演练</h1>
    
    <el-card>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
        label-position="top"
      >
        <el-form-item label="演练标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入演练标题" />
        </el-form-item>
        
        <el-form-item label="演练类型" prop="category">
          <el-select v-model="form.category" placeholder="请选择演练类型">
            <el-option label="灾备切换" value="disaster_recovery" />
            <el-option label="服务降级" value="degradation" />
            <el-option label="发布回滚" value="rollback" />
            <el-option label="安全演练" value="security" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="演练描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="4"
            placeholder="请输入演练描述"
          />
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="handleSubmit">创建演练</el-button>
          <el-button @click="$router.back()">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'

const formRef = ref<FormInstance>()

const form = reactive({
  title: '',
  category: '',
  description: ''
})

const rules: FormRules = {
  title: [
    { required: true, message: '请输入演练标题', trigger: 'blur' },
    { min: 2, max: 100, message: '长度在 2 到 100 个字符', trigger: 'blur' }
  ],
  category: [
    { required: true, message: '请选择演练类型', trigger: 'change' }
  ]
}

function handleSubmit() {
  formRef.value?.validate((valid) => {
    if (valid) {
      console.log('Submit:', form)
    }
  })
}
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

.drill-create-page {
  .page-title {
    font-size: $font-size-2xl;
    font-weight: $font-weight-bold;
    color: $text-primary;
    margin-bottom: $spacing-xl;
  }
}
</style>
