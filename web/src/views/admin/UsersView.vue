<template>
  <div class="page-container">
    <div class="page-header">
      <h2>用户管理</h2>
      <div class="header-actions">
        <el-input
          v-model="searchQuery"
          placeholder="搜索用户名/姓名/部门"
          class="search-input"
          clearable
          @input="handleSearch"
        >
          <template #prefix>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
              <circle cx="11" cy="11" r="8"/>
              <path d="m21 21-4.35-4.35"/>
            </svg>
          </template>
        </el-input>
        <el-button type="primary" @click="showCreateDialog = true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
            <path d="M12 5v14M5 12h14"/>
          </svg>
          创建用户
        </el-button>
      </div>
    </div>

    <div class="page-content">
      <DataTable
        :columns="columns"
        :data="filteredUsers"
        :loading="loading"
        pagination
        :total="total"
      >
        <template #role="{ row }">
          <el-tag :type="getRoleTagType(row.role)" size="small">
            {{ getRoleLabel(row.role) }}
          </el-tag>
        </template>
        <template #status="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ row.status === 1 ? '正常' : '禁用' }}
          </el-tag>
        </template>
        <template #last_login_at="{ row }">
          {{ row.last_login_at ? formatTime(row.last_login_at) : '从未登录' }}
        </template>
        <template #actions="{ row }">
          <el-button
            type="primary"
            size="small"
            @click="handleEditUser(row)"
          >
            编辑
          </el-button>
          <ActionConfirm
            message="确认要删除该用户吗？删除后无法恢复。"
            danger
            size="small"
            @confirm="handleDeleteUser(row)"
          >
            删除
          </ActionConfirm>
          <ActionConfirm
            v-if="row.status === 1"
            message="确认要禁用该用户吗？禁用后将无法登录系统。"
            type="warning"
            size="small"
            @confirm="handleDisableUser(row)"
          >
            禁用
          </ActionConfirm>
          <el-button
            v-else
            type="success"
            size="small"
            @click="handleEnableUser(row)"
          >
            启用
          </el-button>
        </template>
      </DataTable>

      <EmptyBox
        v-if="filteredUsers.length === 0 && !loading"
        title="暂无用户数据"
        description="尝试调整搜索条件或创建新用户"
      />
    </div>

    <!-- 创建用户对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      title="创建用户"
      width="500px"
      :close-on-click-modal="false"
      @close="resetForm"
    >
      <el-form
        ref="formRef"
        :model="createForm"
        :rules="formRules"
        label-width="80px"
        label-position="top"
      >
        <el-form-item label="用户名" prop="username">
          <el-input v-model="createForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="姓名" prop="name">
          <el-input v-model="createForm.name" placeholder="请输入真实姓名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="createForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="createForm.role" placeholder="请选择角色" style="width: 100%">
            <el-option label="系统管理员" value="admin" />
            <el-option label="指挥长" value="director" />
            <el-option label="执行员" value="executor" />
            <el-option label="观察员" value="viewer" />
          </el-select>
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="createForm.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="部门" prop="department">
          <el-input v-model="createForm.department" placeholder="请输入部门" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="createForm.password" type="password" placeholder="请输入初始密码" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateUser" :loading="submitting">
          确认创建
        </el-button>
      </template>
    </el-dialog>

    <!-- 编辑用户抽屉 -->
    <el-drawer
      v-model="showEditDrawer"
      title="编辑用户"
      size="500px"
      :close-on-click-modal="false"
      @close="resetEditForm"
    >
      <el-form
        ref="editFormRef"
        :model="editForm"
        :rules="editFormRules"
        label-width="80px"
        label-position="top"
      >
        <el-form-item label="用户名" prop="username">
          <el-input v-model="editForm.username" disabled />
        </el-form-item>
        <el-form-item label="姓名" prop="real_name">
          <el-input v-model="editForm.real_name" placeholder="请输入真实姓名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="editForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="editForm.role" placeholder="请选择角色" style="width: 100%">
            <el-option label="系统管理员" value="admin" />
            <el-option label="指挥长" value="director" />
            <el-option label="执行员" value="executor" />
            <el-option label="观察员" value="viewer" />
          </el-select>
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="editForm.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="部门" prop="department">
          <el-input v-model="editForm.department" placeholder="请输入部门" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div style="display: flex; justify-content: flex-end; gap: 10px;">
          <el-button @click="showEditDrawer = false">取消</el-button>
          <el-button type="primary" @click="handleUpdateUser" :loading="submitting">
            确认保存
          </el-button>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import DataTable, { type TableColumn } from '@/components/common/DataTable.vue'
import ActionConfirm from '@/components/common/ActionConfirm.vue'
import EmptyBox from '@/components/common/EmptyBox.vue'
import { userApi } from '@/api/modules/user'
import type { User } from '@/types'

const loading = ref(false)
const submitting = ref(false)
const showCreateDialog = ref(false)
const showEditDrawer = ref(false)
const searchQuery = ref('')
const formRef = ref<FormInstance>()
const editFormRef = ref<FormInstance>()
const editingUserId = ref<number | null>(null)

const users = ref<User[]>([])
const total = ref(0)

async function loadUsers() {
  loading.value = true
  try {
    const res = await userApi.getList({ page: 1, page_size: 100 })
    users.value = res.items || []
    total.value = res.total || 0
  } catch (error) {
    console.error('Failed to load users:', error)
    ElMessage.error('加载用户列表失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadUsers()
})

// 表格列定义
const columns: TableColumn[] = [
  { prop: 'username', label: '用户名', width: 140 },
  { prop: 'real_name', label: '真实姓名', width: 160 },
  { prop: 'role', label: '角色', width: 120, slot: true },
  { prop: 'status', label: '状态', width: 100, slot: true },
  { prop: 'department', label: '部门' },
  { prop: 'phone', label: '手机号', width: 160 },
  { prop: 'created_at', label: '创建时间', width: 180 },
]

// 创建用户表单
const createForm = ref({
  username: '',
  name: '',
  email: '',
  role: 'executor',
  phone: '',
  department: '',
  password: '',
})

// 表单验证规则
const formRules: FormRules = {
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
}

// 编辑用户表单
const editForm = ref({
  username: '',
  real_name: '',
  email: '',
  role: '',
  phone: '',
  department: '',
})

// 编辑表单验证规则
const editFormRules: FormRules = {
  real_name: [
    { required: true, message: '请输入姓名', trigger: 'blur' },
    { min: 2, message: '姓名至少 2 个字符', trigger: 'blur' },
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
  ],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }],
}

// 过滤后的用户列表
const filteredUsers = computed(() => {
  if (!searchQuery.value) return users.value
  const query = searchQuery.value.toLowerCase()
  return users.value.filter(
    (user) =>
      user.username.toLowerCase().includes(query) ||
      user.real_name.toLowerCase().includes(query) ||
      (user.department && user.department.toLowerCase().includes(query))
  )
})

// 角色标签类型
function getRoleTagType(role: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, any> = {
    admin: 'danger',
    director: 'primary',
    executor: 'success',
    viewer: 'info',
  }
  return map[role] || 'info'
}

// 角色标签文本
function getRoleLabel(role: string): string {
  const map: Record<string, string> = {
    admin: '管理员',
    director: '指挥长',
    executor: '执行员',
    viewer: '观察员',
  }
  return map[role] || role
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function handleSearch() {
  // 搜索逻辑已在 computed 中处理
}

async function handleDisableUser(user: User) {
  try {
    await userApi.update(user.id, { status: 0 as any })
    ElMessage.success('用户已禁用')
    loadUsers()
  } catch (error) {
    ElMessage.error('操作失败')
    console.error('Failed to disable user:', error)
  }
}

async function handleEnableUser(user: User) {
  try {
    await userApi.update(user.id, { status: 1 as any })
    ElMessage.success('用户已启用')
    loadUsers()
  } catch (error) {
    ElMessage.error('操作失败')
    console.error('Failed to enable user:', error)
  }
}

async function handleCreateUser() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    submitting.value = true

    await userApi.create({
      username: createForm.value.username,
      name: createForm.value.name,
      email: createForm.value.email,
      role: createForm.value.role,
      phone: createForm.value.phone || undefined,
      department: createForm.value.department || undefined,
      password: createForm.value.password,
    } as any)

    ElMessage.success('用户创建成功')
    showCreateDialog.value = false
    loadUsers()
  } catch (error: any) {
    if (error.message && !error.message.includes('validate')) {
      ElMessage.error('创建失败')
      console.error('Failed to create user:', error)
    }
  } finally {
    submitting.value = false
  }
}

function resetForm() {
  formRef.value?.resetFields()
  createForm.value = {
    username: '',
    name: '',
    email: '',
    role: 'executor',
    phone: '',
    department: '',
    password: '',
  }
}

// 编辑用户相关函数
function handleEditUser(user: User) {
  editingUserId.value = user.id
  editForm.value = {
    username: user.username,
    real_name: user.real_name || '',
    email: user.email || '',
    role: user.role,
    phone: user.phone || '',
    department: user.department || '',
  }
  showEditDrawer.value = true
}

async function handleUpdateUser() {
  if (!editFormRef.value || !editingUserId.value) return

  try {
    await editFormRef.value.validate()
    submitting.value = true

    await userApi.update(editingUserId.value, {
      real_name: editForm.value.real_name,
      email: editForm.value.email,
      role: editForm.value.role,
      phone: editForm.value.phone || undefined,
      department: editForm.value.department || undefined,
    })

    ElMessage.success('用户信息已更新')
    showEditDrawer.value = false
    loadUsers()
  } catch (error: any) {
    if (error.message && !error.message.includes('validate')) {
      ElMessage.error('更新失败')
      console.error('Failed to update user:', error)
    }
  } finally {
    submitting.value = false
  }
}

function resetEditForm() {
  editFormRef.value?.resetFields()
  editingUserId.value = null
  editForm.value = {
    username: '',
    real_name: '',
    email: '',
    role: '',
    phone: '',
    department: '',
  }
}

async function handleDeleteUser(user: User) {
  try {
    await userApi.delete(user.id)
    ElMessage.success('用户已删除')
    loadUsers()
  } catch (error) {
    ElMessage.error('删除失败')
    console.error('Failed to delete user:', error)
  }
}
</script>

<style scoped lang="scss">
@use '@/styles/layout' as *;
@use '@/styles/variables' as *;

.page-container {
  @include page-container;

  .page-header {
    @include page-header;

    h2 {
      font-size: $font-size-xl;
      color: $text-primary;
      font-weight: $font-weight-semibold;
      margin: 0;
    }

    .header-actions {
      display: flex;
      align-items: center;
      gap: $spacing-sm;

      .search-input {
        width: 240px;

        :deep(.el-input__wrapper) {
          background: $bg-secondary;
          border-color: $border-color;

          .el-input__inner {
            color: $text-primary;
          }
        }
      }
    }
  }

  .page-content {
    @include page-content;

    :deep(.el-card) {
      @include card-compact;
    }

    :deep(.el-dialog__body) {
      .el-form-item__label {
        color: $text-secondary;
      }

      .el-input__wrapper,
      .el-textarea__inner {
        background: $bg-tertiary;
        border-color: $border-color;
        color: $text-primary;
      }
    }
  }
}
</style>
