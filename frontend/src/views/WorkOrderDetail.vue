<template>
  <div class="work-order-detail-container">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <el-button type="primary" link @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
            返回
          </el-button>
          <span>工单详情 - {{ workOrder?.order_number }}</span>
          <el-tag :type="statusTypeMap[workOrder?.status || '']" size="large">
            {{ statusMap[workOrder?.status || ''] }}
          </el-tag>
        </div>
      </template>

      <el-descriptions :column="2" border>
        <el-descriptions-item label="工单号">{{ workOrder?.order_number }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTypeMap[workOrder?.status || '']">
            {{ statusMap[workOrder?.status || ''] }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="设备信息">
          {{ workOrder?.device?.device_code }} - {{ workOrder?.device?.name }}
        </el-descriptions-item>
        <el-descriptions-item label="故障类型">
          {{ faultTypeMap[workOrder?.fault_type || ''] }}
        </el-descriptions-item>
        <el-descriptions-item label="紧急程度">
          <el-tag :type="urgencyTypeMap[workOrder?.urgency || '']">
            {{ urgencyMap[workOrder?.urgency || ''] }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="报修人">
          {{ workOrder?.employee?.real_name }}
        </el-descriptions-item>
        <el-descriptions-item label="维修技师">
          {{ workOrder?.technician?.real_name || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ formatDate(workOrder?.created_at) }}
        </el-descriptions-item>
        <el-descriptions-item label="故障描述" :span="2">
          {{ workOrder?.fault_description }}
        </el-descriptions-item>
        <el-descriptions-item v-if="workOrder?.repair_measures" label="维修措施" :span="2">
          {{ workOrder?.repair_measures }}
        </el-descriptions-item>
        <el-descriptions-item v-if="workOrder?.replaced_parts" label="更换配件" :span="2">
          {{ workOrder?.replaced_parts }}
        </el-descriptions-item>
        <el-descriptions-item v-if="workOrder?.repair_duration" label="维修耗时">
          {{ workOrder?.repair_duration }} 分钟
        </el-descriptions-item>
      </el-descriptions>

      <el-divider content-position="left">现场照片</el-divider>
      <el-image v-if="beforeImages.length > 0"
        v-for="(img, index) in beforeImages"
        :key="img.id"
        :src="`/uploads/${img.file_path}`"
        :preview-src-list="beforeImages.map(i => `/uploads/${i.file_path}`)"
        :initial-index="index"
        style="width: 150px; height: 150px; margin-right: 10px; margin-bottom: 10px;"
        fit="cover"
        :preview-teleported="true"
      />
      <el-empty v-else description="暂无现场照片" :image-size="60" />

      <el-divider content-position="left" v-if="afterImages.length > 0">维修后照片</el-divider>
      <el-image v-if="afterImages.length > 0"
        v-for="(img, index) in afterImages"
        :key="img.id"
        :src="`/uploads/${img.file_path}`"
        :preview-src-list="afterImages.map(i => `/uploads/${i.file_path}`)"
        :initial-index="index"
        style="width: 150px; height: 150px; margin-right: 10px; margin-bottom: 10px;"
        fit="cover"
        :preview-teleported="true"
      />

      <el-divider content-position="left">操作按钮</el-divider>
      <div class="action-buttons">
        <el-button type="primary" v-if="canAssign" @click="showAssignDialog = true">
          指派技师
        </el-button>
        <el-button type="primary" v-if="canAccept" @click="handleAccept">
          接单处理
        </el-button>
        <el-button type="primary" v-if="canSubmit" @click="showSubmitDialog = true">
          提交维修完成
        </el-button>
        <el-button type="success" v-if="canConfirm" @click="handleConfirm">
          确认完成
        </el-button>
        <el-button type="warning" v-if="canReject" @click="handleReject">
          打回重修
        </el-button>
      </div>

      <el-divider content-position="left">状态流转时间线</el-divider>
      <el-timeline>
        <el-timeline-item
          v-for="log in sortedLogs"
          :key="log.id"
          :timestamp="formatDate(log.created_at)"
          placement="top"
        >
          <el-card>
            <h4>{{ operationMap[log.operation] }}</h4>
            <p v-if="log.user">操作人：{{ log.user.real_name }}</p>
            <p v-if="log.old_status">
              状态变更：{{ statusMap[log.old_status] }} → {{ statusMap[log.new_status || ''] }}
            </p>
          </el-card>
        </el-timeline-item>
      </el-timeline>
    </el-card>

    <el-dialog v-model="showAssignDialog" title="指派技师" width="500px">
      <el-form label-width="100px">
        <el-form-item label="选择技师">
          <el-select v-model="selectedTechnicianId" placeholder="请选择技师" style="width: 100%">
            <el-option
              v-for="tech in technicians"
              :key="tech.id"
              :label="tech.real_name"
              :value="tech.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAssignDialog = false">取消</el-button>
        <el-button type="primary" :loading="actionLoading" @click="handleAssign">确认</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showSubmitDialog" title="提交维修完成" width="600px">
      <el-form ref="submitFormRef" :model="submitForm" :rules="submitRules" label-width="100px">
        <el-form-item label="维修措施" prop="repair_measures">
          <el-input
            v-model="submitForm.repair_measures"
            type="textarea"
            :rows="4"
            placeholder="请详细描述维修措施"
          />
        </el-form-item>
        <el-form-item label="更换配件">
          <el-input
            v-model="submitForm.replaced_parts"
            type="textarea"
            :rows="2"
            placeholder="请输入更换的配件（可选）"
          />
        </el-form-item>
        <el-form-item label="维修耗时">
          <el-input-number v-model="submitForm.repair_duration" :min="0" placeholder="分钟" />
          <span style="margin-left: 10px;">分钟</span>
        </el-form-item>
        <el-form-item label="维修后照片">
          <el-upload
            :action="uploadAction"
            :headers="uploadHeaders"
            :limit="4"
            list-type="picture-card"
            :on-success="handleAfterUploadSuccess"
            :on-remove="handleAfterRemove"
            :file-list="afterFileList"
            :before-upload="beforeUpload"
          >
            <el-icon v-if="afterFileList.length < 4"><Plus /></el-icon>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSubmitDialog = false">取消</el-button>
        <el-button type="primary" :loading="actionLoading" @click="handleSubmitRepair">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules, type UploadFile, type UploadUserFile } from 'element-plus'
import { workOrderApi, userApi } from '@/api'
import type { WorkOrder, User, OperationLog, Image } from '@/types'
import { useUserStore } from '@/store/user'
import { ArrowLeft, Plus } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const loading = ref(false)
const actionLoading = ref(false)
const workOrder = ref<WorkOrder | null>(null)
const technicians = ref<User[]>([])
const selectedTechnicianId = ref<number | null>(null)

const showAssignDialog = ref(false)
const showSubmitDialog = ref(false)

const submitFormRef = ref<FormInstance>()
const submitForm = reactive({
  repair_measures: '',
  replaced_parts: '',
  repair_duration: 0
})

const submitRules: FormRules = {
  repair_measures: [{ required: true, message: '请输入维修措施', trigger: 'blur' }]
}

const afterFileList = ref<UploadUserFile[]>([])
const uploadedAfterImageIds = ref<number[]>([])

const uploadAction = '/api/images/upload'

const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${localStorage.getItem('token')}`
}))

const statusMap: Record<string, string> = {
  pending_assign: '待分配',
  assigned: '已分配',
  processing: '处理中',
  pending_confirm: '待确认',
  closed: '已关闭'
}

const statusTypeMap: Record<string, string> = {
  pending_assign: 'info',
  assigned: 'warning',
  processing: 'primary',
  pending_confirm: 'warning',
  closed: 'success'
}

const urgencyMap: Record<string, string> = {
  low: '低',
  medium: '中',
  high: '高',
  urgent: '紧急'
}

const urgencyTypeMap: Record<string, string> = {
  low: 'info',
  medium: 'warning',
  high: 'danger',
  urgent: 'danger'
}

const faultTypeMap: Record<string, string> = {
  hardware: '硬件故障',
  software: '软件问题',
  network: '网络异常',
  other: '其他'
}

const operationMap: Record<string, string> = {
  create: '创建工单',
  assign: '指派工单',
  accept: '接单处理',
  process: '处理中',
  submit: '提交维修完成',
  confirm: '确认完成',
  reject: '打回重修'
}

const beforeImages = computed<Image[]>(() => {
  return workOrder.value?.images?.filter(img => img.image_type === 'before') || []
})

const afterImages = computed<Image[]>(() => {
  return workOrder.value?.images?.filter(img => img.image_type === 'after') || []
})

const sortedLogs = computed<OperationLog[]>(() => {
  return [...(workOrder.value?.logs || [])].sort((a, b) => {
    return new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
  })
})

const canAssign = computed(() => {
  return userStore.isAdmin && workOrder.value?.status === 'pending_assign'
})

const canAccept = computed(() => {
  return userStore.isTechnician &&
         workOrder.value?.status === 'assigned' &&
         workOrder.value?.technician_id === userStore.user?.id
})

const canSubmit = computed(() => {
  return userStore.isTechnician &&
         workOrder.value?.status === 'processing' &&
         workOrder.value?.technician_id === userStore.user?.id
})

const canConfirm = computed(() => {
  return userStore.isEmployee &&
         workOrder.value?.status === 'pending_confirm' &&
         workOrder.value?.employee_id === userStore.user?.id
})

const canReject = computed(() => {
  return userStore.isEmployee &&
         workOrder.value?.status === 'pending_confirm' &&
         workOrder.value?.employee_id === userStore.user?.id
})

const formatDate = (date?: string) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

const goBack = () => {
  router.back()
}

const fetchWorkOrder = async () => {
  const id = route.params.id
  if (!id) return

  loading.value = true
  try {
    const response = await workOrderApi.getWorkOrderById(Number(id))
    workOrder.value = response.data
  } catch (error) {
    console.error('Failed to fetch work order:', error)
    ElMessage.error('获取工单详情失败')
  } finally {
    loading.value = false
  }
}

const fetchTechnicians = async () => {
  try {
    const response = await userApi.getTechnicians()
    technicians.value = response.data
  } catch (error) {
    console.error('Failed to fetch technicians:', error)
  }
}

const beforeUpload = (file: File) => {
  const isImage = file.type.startsWith('image/')
  if (!isImage) {
    ElMessage.error('只能上传图片文件！')
    return false
  }
  const isLt10M = file.size / 1024 / 1024 < 10
  if (!isLt10M) {
    ElMessage.error('图片大小不能超过10MB！')
    return false
  }
  return true
}

const handleAfterUploadSuccess = (response: any, file: UploadFile) => {
  if (response && response.id) {
    uploadedAfterImageIds.value.push(response.id)
    afterFileList.value = afterFileList.value.map(item => {
      if (item.uid === file.uid) {
        return { ...item, response }
      }
      return item
    })
    ElMessage.success('上传成功')
  } else {
    ElMessage.error('上传失败')
  }
}

const handleAfterRemove = (file: UploadFile) => {
  if (file.response && file.response.id) {
    const index = uploadedAfterImageIds.value.indexOf(file.response.id)
    if (index > -1) {
      uploadedAfterImageIds.value.splice(index, 1)
    }
  }
}

const handleAssign = async () => {
  if (!selectedTechnicianId.value) {
    ElMessage.warning('请选择技师')
    return
  }

  actionLoading.value = true
  try {
    await workOrderApi.assignWorkOrder(workOrder.value!.id, selectedTechnicianId.value)
    ElMessage.success('指派成功')
    showAssignDialog.value = false
    fetchWorkOrder()
  } catch (error: any) {
    const message = error.response?.data?.error || '指派失败'
    ElMessage.error(message)
  } finally {
    actionLoading.value = false
  }
}

const handleAccept = async () => {
  try {
    await ElMessageBox.confirm('确定要接单处理吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    actionLoading.value = true
    await workOrderApi.acceptWorkOrder(workOrder.value!.id)
    ElMessage.success('接单成功')
    fetchWorkOrder()
  } catch (error: any) {
    if (error !== 'cancel') {
      const message = error.response?.data?.error || '接单失败'
      ElMessage.error(message)
    }
  } finally {
    actionLoading.value = false
  }
}

const handleSubmitRepair = async () => {
  if (!submitFormRef.value) return

  await submitFormRef.value.validate(async (valid) => {
    if (valid) {
      actionLoading.value = true
      try {
        const data = {
          repair_measures: submitForm.repair_measures,
          replaced_parts: submitForm.replaced_parts,
          repair_duration: submitForm.repair_duration,
          after_image_ids: uploadedAfterImageIds.value.length > 0 ? uploadedAfterImageIds.value : undefined
        }
        await workOrderApi.submitRepair(workOrder.value!.id, data)
        ElMessage.success('提交成功')
        showSubmitDialog.value = false
        fetchWorkOrder()
      } catch (error: any) {
        const message = error.response?.data?.error || '提交失败'
        ElMessage.error(message)
      } finally {
        actionLoading.value = false
      }
    }
  })
}

const handleConfirm = async () => {
  try {
    await ElMessageBox.confirm('确认维修完成吗？确认后工单将关闭。', '提示', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'success'
    })

    actionLoading.value = true
    await workOrderApi.confirmWorkOrder(workOrder.value!.id)
    ElMessage.success('确认成功')
    fetchWorkOrder()
  } catch (error: any) {
    if (error !== 'cancel') {
      const message = error.response?.data?.error || '确认失败'
      ElMessage.error(message)
    }
  } finally {
    actionLoading.value = false
  }
}

const handleReject = async () => {
  try {
    await ElMessageBox.confirm('确定要打回重修吗？工单将返回处理中状态。', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    actionLoading.value = true
    await workOrderApi.rejectWorkOrder(workOrder.value!.id)
    ElMessage.success('打回成功')
    fetchWorkOrder()
  } catch (error: any) {
    if (error !== 'cancel') {
      const message = error.response?.data?.error || '打回失败'
      ElMessage.error(message)
    }
  } finally {
    actionLoading.value = false
  }
}

watch(showSubmitDialog, (val) => {
  if (!val) {
    submitForm.repair_measures = ''
    submitForm.replaced_parts = ''
    submitForm.repair_duration = 0
    afterFileList.value = []
    uploadedAfterImageIds.value = []
  }
})

onMounted(() => {
  fetchWorkOrder()
  if (userStore.isAdmin) {
    fetchTechnicians()
  }
})
</script>

<style scoped>
.work-order-detail-container {
  padding: 0;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.action-buttons {
  display: flex;
  gap: 10px;
}

:deep(.el-timeline-item__tail) {
  border-left: 2px solid #e4e7ed;
}

:deep(.el-timeline-item__node) {
  background-color: #409EFF;
}
</style>
