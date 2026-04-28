<template>
  <div class="user-list-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户管理</span>
          <el-button type="primary" @click="showCreateDialog = true">
            添加用户
          </el-button>
        </div>
      </template>

      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item label="角色">
          <el-select v-model="filters.role" placeholder="全部角色" clearable style="width: 150px">
            <el-option label="普通员工" value="employee" />
            <el-option label="维修技师" value="technician" />
            <el-option label="管理员" value="admin" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input
            v-model="filters.keyword"
            placeholder="用户名/真实姓名"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="users" v-loading="loading" style="width: 100%">
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="real_name" label="真实姓名" width="150" />
        <el-table-column label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="roleTypeMap[row.role]" size="small">
              {{ roleMap[row.role] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
              {{ row.status === 'active' ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">
            {{ new Date(row.created_at).toLocaleString('zh-CN') }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button type="primary" link @click="handleResetPassword(row)">
              重置密码
            </el-button>
            <el-button
              v-if="row.status === 'active'"
              type="danger"
              link
              @click="handleToggleStatus(row)"
            >
              禁用
            </el-button>
            <el-button
              v-else
              type="success"
              link
              @click="handleToggleStatus(row)"
            >
              启用
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        style="margin-top: 20px; justify-content: flex-end;"
      />
    </el-card>

    <el-dialog
      v-model="showCreateDialog"
      :title="editingUser ? '编辑用户' : '添加用户'"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="form.username"
            placeholder="请输入用户名"
            :disabled="!!editingUser"
          />
        </el-form-item>
        <el-form-item
          v-if="!editingUser"
          label="密码"
          prop="password"
        >
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="真实姓名" prop="real_name">
          <el-input v-model="form.real_name" placeholder="请输入真实姓名" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select
            v-model="form.role"
            placeholder="请选择角色"
            style="width: 100%"
            :disabled="editingUser?.role === 'admin'"
          >
            <el-option label="普通员工" value="employee" />
            <el-option label="维修技师" value="technician" />
            <el-option label="管理员" value="admin" />
          </el-select>
          <el-text v-if="editingUser?.role === 'admin'" type="warning" size="small" class="ml-2">
            管理员角色不可修改
          </el-text>
        </el-form-item>
        <el-form-item v-if="editingUser" label="状态">
          <el-radio-group v-model="form.status">
            <el-radio value="active">正常</el-radio>
            <el-radio value="inactive">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" :loading="actionLoading" @click="handleSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { userApi } from '@/api'
import type { User } from '@/types'

const formRef = ref<FormInstance>()

const loading = ref(false)
const actionLoading = ref(false)
const showCreateDialog = ref(false)
const editingUser = ref<User | null>(null)

const users = ref<User[]>([])

const filters = reactive({
  role: '',
  keyword: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const form = reactive({
  username: '',
  password: '',
  real_name: '',
  role: 'employee' as const,
  status: 'active' as const
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  real_name: [{ required: true, message: '请输入真实姓名', trigger: 'blur' }],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }]
}

const roleMap: Record<string, string> = {
  employee: '普通员工',
  technician: '维修技师',
  admin: '管理员'
}

const roleTypeMap: Record<string, string> = {
  employee: 'info',
  technician: 'warning',
  admin: 'danger'
}

const fetchUsers = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (filters.role) {
      params.role = filters.role
    }
    if (filters.keyword) {
      params.keyword = filters.keyword
    }

    const response = await userApi.getUsers(params)
    users.value = response.data.data
    pagination.total = response.data.total
  } catch (error) {
    console.error('Failed to fetch users:', error)
    ElMessage.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchUsers()
}

const handleReset = () => {
  filters.role = ''
  filters.keyword = ''
  pagination.page = 1
  fetchUsers()
}

const handleSizeChange = (val: number) => {
  pagination.pageSize = val
  fetchUsers()
}

const handleCurrentChange = (val: number) => {
  pagination.page = val
  fetchUsers()
}

const handleEdit = (user: User) => {
  editingUser.value = user
  form.username = user.username
  form.real_name = user.real_name
  form.role = user.role
  form.status = user.status
  showCreateDialog.value = true
}

const handleResetPassword = async (user: User) => {
  try {
    await ElMessageBox.confirm(
      `确定要将用户 \"${user.real_name}\" 的密码重置为 \"123456\" 吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    actionLoading.value = true
    await userApi.resetPassword(user.id)
    ElMessage.success('密码重置成功')
    fetchUsers()
  } catch (error: any) {
    if (error !== 'cancel') {
      const message = error.response?.data?.error || '操作失败'
      ElMessage.error(message)
    }
  } finally {
    actionLoading.value = false
  }
}

const handleToggleStatus = async (user: User) => {
  const newStatus = user.status === 'active' ? 'inactive' : 'active'
  const action = newStatus === 'active' ? '启用' : '禁用'

  try {
    await ElMessageBox.confirm(
      `确定要${action}用户 \"${user.real_name}\" 吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    actionLoading.value = true
    await userApi.updateUser(user.id, { status: newStatus })
    ElMessage.success(`${action}成功`)
    fetchUsers()
  } catch (error: any) {
    if (error !== 'cancel') {
      const message = error.response?.data?.error || '操作失败'
      ElMessage.error(message)
    }
  } finally {
    actionLoading.value = false
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      actionLoading.value = true
      try {
        const data: any = {
          real_name: form.real_name,
          role: form.role
        }

        if (editingUser.value) {
          data.status = form.status
          await userApi.updateUser(editingUser.value.id, data)
          ElMessage.success('更新成功')
        } else {
          data.username = form.username
          data.password = form.password
          await userApi.createUser(data)
          ElMessage.success('创建成功')
        }

        showCreateDialog.value = false
        fetchUsers()
      } catch (error: any) {
        const message = error.response?.data?.error || '操作失败'
        ElMessage.error(message)
      } finally {
        actionLoading.value = false
      }
    }
  })
}

watch(showCreateDialog, (val) => {
  if (!val) {
    editingUser.value = null
    formRef.value?.resetFields()
    form.username = ''
    form.password = ''
    form.real_name = ''
    form.role = 'employee'
    form.status = 'active'
  }
})

onMounted(() => {
  fetchUsers()
})
</script>

<style scoped>
.user-list-container {
  padding: 0;
  width: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-form {
  margin-bottom: 20px;
}

:deep(.el-card__body) {
  padding: 20px;
}

.ml-2 {
  margin-left: 8px;
}
</style>
