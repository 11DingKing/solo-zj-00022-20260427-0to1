<template>
  <div class="create-work-order-container">
    <el-card>
      <template #header>
        <span>提交报修</span>
      </template>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
        style="max-width: 800px;"
      >
        <el-form-item label="设备编号" prop="device_id">
          <el-select
            v-model="form.device_id"
            placeholder="请选择设备"
            style="width: 100%"
            filterable
          >
            <el-option
              v-for="device in devices"
              :key="device.id"
              :label="`${device.device_code} - ${device.name}"
              :value="device.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="故障类型" prop="fault_type">
          <el-select v-model="form.fault_type" placeholder="请选择故障类型" style="width: 100%">
            <el-option label="硬件故障" value="hardware" />
            <el-option label="软件问题" value="software" />
            <el-option label="网络异常" value="network" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>

        <el-form-item label="紧急程度" prop="urgency">
          <el-radio-group v-model="form.urgency">
            <el-radio value="low">低</el-radio>
            <el-radio value="medium">中</el-radio>
            <el-radio value="high">高</el-radio>
            <el-radio value="urgent">紧急</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="故障描述" prop="fault_description">
          <el-input
            v-model="form.fault_description"
            type="textarea"
            :rows="4"
            placeholder="请详细描述故障情况"
          />
        </el-form-item>

        <el-form-item label="现场照片">
          <el-upload
            :action="uploadAction"
            :headers="uploadHeaders"
            :limit="4"
            list-type="picture-card"
            :on-success="handleUploadSuccess"
            :on-remove="handleRemove"
            :file-list="fileList"
            :before-upload="beforeUpload"
          >
            <el-icon v-if="fileList.length < 4"><Plus /></el-icon>
            <template #tip>
              <div class="el-upload__tip">最多上传4张图片，支持jpg/png/gif格式，单张不超过10MB</div>
            </template>
          </el-upload>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleSubmit">提交</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules, type UploadFile, type UploadUserFile } from 'element-plus'
import { deviceApi, workOrderApi } from '@/api'
import type { Device } from '@/types'
import { Plus } from '@element-plus/icons-vue'

const router = useRouter()
const formRef = ref<FormInstance>()

const loading = ref(false)
const devices = ref<Device[]>([])
const fileList = ref<UploadUserFile[]>([])
const uploadedImageIds = ref<number[]>([])

const uploadAction = '/api/images/upload'

const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${localStorage.getItem('token')}`
}))

const form = reactive({
  device_id: null as number | null,
  fault_type: '',
  urgency: 'low' as const,
  fault_description: ''
})

const rules: FormRules = {
  device_id: [{ required: true, message: '请选择设备', trigger: 'change' }],
  fault_type: [{ required: true, message: '请选择故障类型', trigger: 'change' }],
  fault_description: [{ required: true, message: '请输入故障描述', trigger: 'blur' }]
}

const fetchDevices = async () => {
  try {
    const response = await deviceApi.getDevices({ page_size: 100 })
    devices.value = response.data.data
  } catch (error) {
    console.error('Failed to fetch devices:', error)
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

const handleUploadSuccess = (response: any, file: UploadFile) => {
  if (response && response.id) {
    uploadedImageIds.value.push(response.id)
    fileList.value = fileList.value.map(item => {
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

const handleRemove = (file: UploadFile) => {
  if (file.response && file.response.id) {
    const index = uploadedImageIds.value.indexOf(file.response.id)
    if (index > -1) {
      uploadedImageIds.value.splice(index, 1)
    }
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const data = {
          device_id: form.device_id!,
          fault_type: form.fault_type,
          urgency: form.urgency,
          fault_description: form.fault_description,
          before_image_ids: uploadedImageIds.value.length > 0 ? uploadedImageIds.value : undefined
        }
        await workOrderApi.createWorkOrder(data)
        ElMessage.success('提交成功')
        router.push('/work-orders')
      } catch (error: any) {
        const message = error.response?.data?.error || '提交失败'
        ElMessage.error(message)
      } finally {
        loading.value = false
      }
    }
  })
}

const handleReset = () => {
  formRef.value?.resetFields()
  fileList.value = []
  uploadedImageIds.value = []
}

onMounted(() => {
  fetchDevices()
})
</script>

<style scoped>
.create-work-order-container {
  padding: 0;
}

:deep(.el-upload--picture-card) {
  width: 148px;
  height: 148px;
}
</style>
