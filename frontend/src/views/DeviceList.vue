<template>
  <div class="device-list-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>设备台账</span>
          <el-button type="primary" v-if="userStore.isAdmin" @click="showCreateDialog = true">
            添加设备
          </el-button>
        </div>
      </template>

      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item label="关键词">
          <el-input
            v-model="filters.keyword"
            placeholder="设备编号/名称/型号"
            clearable
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部状态" clearable style="width: 120px">
            <el-option label="正常" value="active" />
            <el-option label="维护中" value="maintenance" />
            <el-option label="已报废" value="scrapped" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="devices" v-loading="loading" style="width: 100%">
        <el-table-column prop="device_code" label="设备编号" width="150" />
        <el-table-column prop="name" label="设备名称" width="150" />
        <el-table-column prop="model" label="型号" width="150" />
        <el-table-column prop="location" label="位置" width="150" />
        <el-table-column label="购入日期" width="120">
          <template #default="{ row }">
            {{ row.purchase_date ? new Date(row.purchase_date).toLocaleDateString('zh-CN') : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="保修到期日" width="120">
          <template #default="{ row }">
            {{ row.warranty_expire_date ? new Date(row.warranty_expire_date).toLocaleDateString('zh-CN') : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleViewDetail(row.id)">
              详情
            </el-button>
            <el-button type="primary" link v-if="userStore.isAdmin" @click="handleEdit(row)">
              编辑
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
      :title="editingDevice ? '编辑设备' : '添加设备'"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
      >
        <el-form-item label="设备编号" prop="device_code">
          <el-input v-model="form.device_code" placeholder="请输入设备编号" />
        </el-form-item>
        <el-form-item label="设备名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入设备名称" />
        </el-form-item>
        <el-form-item label="型号">
          <el-input v-model="form.model" placeholder="请输入型号" />
        </el-form-item>
        <el-form-item label="位置">
          <el-input v-model="form.location" placeholder="请输入位置" />
        </el-form-item>
        <el-form-item label="购入日期" prop="purchase_date">
          <el-date-picker
            v-model="form.purchase_date"
            type="date"
            placeholder="选择日期"
            value-format="YYYY-MM-DD"
            style="width: 100%"
            @change="handlePurchaseDateChange"
          />
        </el-form-item>
        <el-form-item label="保修到期日" prop="warranty_expire_date">
          <el-date-picker
            v-model="form.warranty_expire_date"
            type="date"
            placeholder="选择日期"
            value-format="YYYY-MM-DD"
            style="width: 100%"
            :disabled-date="disabledWarrantyDate"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio value="active">正常</el-radio>
            <el-radio value="maintenance">维护中</el-radio>
            <el-radio value="scrapped">已报废</el-radio>
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
import { ref, reactive, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { deviceApi } from '@/api'
import type { Device, DeviceStatus } from '@/types'
import { useUserStore } from '@/store/user'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref<FormInstance>()

const loading = ref(false)
const actionLoading = ref(false)
const showCreateDialog = ref(false)
const editingDevice = ref<Device | null>(null)

const devices = ref<Device[]>([])

const filters = reactive({
  keyword: '',
  status: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const form = reactive({
  device_code: '',
  name: '',
  model: '',
  location: '',
  purchase_date: '',
  warranty_expire_date: '',
  status: 'active' as DeviceStatus
})

const validateWarrantyDate = (rule: any, value: string, callback: any) => {
  if (form.purchase_date && value) {
    const purchaseDate = new Date(form.purchase_date)
    const warrantyDate = new Date(value)
    if (warrantyDate < purchaseDate) {
      callback(new Error('保修到期日不能早于购入日期'))
      return
    }
  }
  callback()
}

const rules: FormRules = {
  device_code: [{ required: true, message: '请输入设备编号', trigger: 'blur' }],
  name: [{ required: true, message: '请输入设备名称', trigger: 'blur' }],
  warranty_expire_date: [
    { validator: validateWarrantyDate, trigger: 'change' }
  ]
}

const disabledWarrantyDate = computed(() => {
  return (date: Date) => {
    if (form.purchase_date) {
      const purchaseDate = new Date(form.purchase_date)
      purchaseDate.setHours(0, 0, 0, 0)
      return date < purchaseDate
    }
    return false
  }
})

const handlePurchaseDateChange = () => {
  if (form.purchase_date && form.warranty_expire_date) {
    const purchaseDate = new Date(form.purchase_date)
    const warrantyDate = new Date(form.warranty_expire_date)
    if (warrantyDate < purchaseDate) {
      form.warranty_expire_date = ''
    }
  }
}

const statusLabelMap: Record<string, string> = {
  active: '正常',
  maintenance: '维护中',
  scrapped: '已报废'
}

const statusTypeMap: Record<string, string> = {
  active: 'success',
  maintenance: 'warning',
  scrapped: 'info'
}

const getStatusLabel = (status?: DeviceStatus) => {
  if (!status) return '-'
  return statusLabelMap[status] || status
}

const getStatusType = (status?: DeviceStatus) => {
  if (!status) return 'info'
  return statusTypeMap[status] || 'info'
}

const fetchDevices = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (filters.keyword) {
      params.keyword = filters.keyword
    }
    if (filters.status) {
      params.status = filters.status
    }

    const response = await deviceApi.getDevices(params)
    devices.value = response.data.data
    pagination.total = response.data.total
  } catch (error) {
    console.error('Failed to fetch devices:', error)
    ElMessage.error('获取设备列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchDevices()
}

const handleReset = () => {
  filters.keyword = ''
  filters.status = ''
  pagination.page = 1
  fetchDevices()
}

const handleSizeChange = (val: number) => {
  pagination.pageSize = val
  fetchDevices()
}

const handleCurrentChange = (val: number) => {
  pagination.page = val
  fetchDevices()
}

const handleViewDetail = (id: number) => {
  router.push(`/devices/${id}`)
}

const handleEdit = (device: Device) => {
  editingDevice.value = device
  form.device_code = device.device_code
  form.name = device.name
  form.model = device.model || ''
  form.location = device.location || ''
  form.purchase_date = device.purchase_date || ''
  form.warranty_expire_date = device.warranty_expire_date || ''
  form.status = device.status
  showCreateDialog.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      actionLoading.value = true
      try {
        const data: any = {
          device_code: form.device_code,
          name: form.name,
          model: form.model,
          location: form.location,
          status: form.status
        }
        if (form.purchase_date) {
          data.purchase_date = form.purchase_date
        }
        if (form.warranty_expire_date) {
          data.warranty_expire_date = form.warranty_expire_date
        }

        if (editingDevice.value) {
          await deviceApi.updateDevice(editingDevice.value.id, data)
          ElMessage.success('更新成功')
        } else {
          await deviceApi.createDevice(data)
          ElMessage.success('创建成功')
        }

        showCreateDialog.value = false
        fetchDevices()
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
    editingDevice.value = null
    formRef.value?.resetFields()
    form.device_code = ''
    form.name = ''
    form.model = ''
    form.location = ''
    form.purchase_date = ''
    form.warranty_expire_date = ''
    form.status = 'active'
  }
})

onMounted(() => {
  fetchDevices()
})
</script>

<style scoped>
.device-list-container {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-form {
  margin-bottom: 20px;
}
</style>
