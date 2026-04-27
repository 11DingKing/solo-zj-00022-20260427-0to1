<template>
  <div class="device-detail-container">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <el-button type="primary" link @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
            返回
          </el-button>
          <span>设备详情</span>
        </div>
      </template>

      <el-descriptions :column="2" border>
        <el-descriptions-item label="设备编号">{{ device?.device_code }}</el-descriptions-item>
        <el-descriptions-item label="设备名称">{{ device?.name }}</el-descriptions-item>
        <el-descriptions-item label="型号">{{ device?.model || '-' }}</el-descriptions-item>
        <el-descriptions-item label="位置">{{ device?.location || '-' }}</el-descriptions-item>
        <el-descriptions-item label="购入日期">
          {{ device?.purchase_date ? new Date(device.purchase_date).toLocaleDateString('zh-CN') : '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="保修到期日">
          {{ device?.warranty_expire_date ? new Date(device.warranty_expire_date).toLocaleDateString('zh-CN') : '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(device?.status)">
            {{ getStatusLabel(device?.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ device?.created_at ? new Date(device.created_at).toLocaleString('zh-CN') : '-' }}
        </el-descriptions-item>
      </el-descriptions>

      <el-divider content-position="left">历史工单</el-divider>
      <el-table :data="workOrders" v-loading="workOrdersLoading" style="width: 100%">
        <el-table-column prop="order_number" label="工单号" width="180" />
        <el-table-column label="故障类型" width="120">
          <template #default="{ row }">
            {{ faultTypeMap[row.fault_type] }}
          </template>
        </el-table-column>
        <el-table-column label="紧急程度" width="100">
          <template #default="{ row }">
            <el-tag :type="urgencyTypeMap[row.urgency]" size="small">
              {{ urgencyMap[row.urgency] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTypeMap[row.status]" size="small">
              {{ statusMap[row.status] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="维修技师" width="120">
          <template #default="{ row }">
            {{ row.technician?.real_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">
            {{ new Date(row.created_at).toLocaleString('zh-CN') }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="viewWorkOrder(row.id)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && workOrders.length === 0"
        description="暂无历史工单"
        :image-size="80"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { deviceApi } from '@/api'
import type { Device, WorkOrder, DeviceStatus } from '@/types'
import { ArrowLeft } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()

const loading = ref(false)
const workOrdersLoading = ref(false)
const device = ref<Device | null>(null)
const workOrders = ref<WorkOrder[]>([])

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

const deviceStatusLabelMap: Record<string, string> = {
  active: '正常',
  maintenance: '维护中',
  scrapped: '已报废'
}

const deviceStatusTypeMap: Record<string, string> = {
  active: 'success',
  maintenance: 'warning',
  scrapped: 'info'
}

const getStatusLabel = (status?: DeviceStatus) => {
  if (!status) return '-'
  return deviceStatusLabelMap[status] || status
}

const getStatusType = (status?: DeviceStatus) => {
  if (!status) return 'info'
  return deviceStatusTypeMap[status] || 'info'
}

const goBack = () => {
  router.back()
}

const fetchDevice = async () => {
  const id = route.params.id
  if (!id) return

  loading.value = true
  try {
    const response = await deviceApi.getDeviceById(Number(id))
    device.value = response.data.device
    workOrders.value = response.data.work_orders || []
  } catch (error) {
    console.error('Failed to fetch device:', error)
    ElMessage.error('获取设备详情失败')
  } finally {
    loading.value = false
  }
}

const viewWorkOrder = (id: number) => {
  router.push(`/work-orders/${id}`)
}

onMounted(() => {
  fetchDevice()
})
</script>

<style scoped>
.device-detail-container {
  padding: 0;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
}
</style>
