<template>
  <div class="work-order-list-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>工单列表</span>
          <el-button type="primary" v-if="userStore.isEmployee" @click="goToCreate">
            <el-icon><Plus /></el-icon>
            提交报修
          </el-button>
        </div>
      </template>

      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部状态" clearable>
            <el-option label="待分配" value="pending_assign" />
            <el-option label="已分配" value="assigned" />
            <el-option label="处理中" value="processing" />
            <el-option label="待确认" value="pending_confirm" />
            <el-option label="已关闭" value="closed" />
          </el-select>
        </el-form-item>
        <el-form-item label="紧急程度">
          <el-select v-model="searchForm.urgency" placeholder="全部" clearable>
            <el-option label="低" value="low" />
            <el-option label="中" value="medium" />
            <el-option label="高" value="high" />
            <el-option label="紧急" value="urgent" />
          </el-select>
        </el-form-item>
        <el-form-item label="故障类型">
          <el-select v-model="searchForm.fault_type" placeholder="全部" clearable>
            <el-option label="硬件故障" value="hardware" />
            <el-option label="软件问题" value="software" />
            <el-option label="网络异常" value="network" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="创建时间">
          <el-date-picker
            v-model="searchForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="workOrders" style="width: 100%" v-loading="loading">
        <el-table-column prop="order_number" label="工单号" width="180" />
        <el-table-column label="设备信息" min-width="150">
          <template #default="{ row }">
            <div>{{ row.device?.device_code }} - {{ row.device?.name }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="fault_description" label="故障描述" min-width="150" show-overflow-tooltip />
        <el-table-column prop="fault_type" label="故障类型" width="100">
          <template #default="{ row }">
            {{ faultTypeMap[row.fault_type] }}
          </template>
        </el-table-column>
        <el-table-column prop="urgency" label="紧急程度" width="100">
          <template #default="{ row }">
            <el-tag :type="urgencyTypeMap[row.urgency]">
              {{ urgencyMap[row.urgency] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTypeMap[row.status]">
              {{ statusMap[row.status] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="报修人" width="100">
          <template #default="{ row }">
            {{ row.employee?.real_name }}
          </template>
        </el-table-column>
        <el-table-column label="维修技师" width="100">
          <template #default="{ row }">
            {{ row.technician?.real_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="goToDetail(row.id)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { workOrderApi } from '@/api'
import type { WorkOrder } from '@/types'
import { useUserStore } from '@/store/user'
import { Plus } from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const workOrders = ref<WorkOrder[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)

const searchForm = reactive({
  status: '',
  urgency: '',
  fault_type: '',
  dateRange: [] as string[]
})

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

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('zh-CN')
}

const fetchWorkOrders = async () => {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: currentPage.value,
      page_size: pageSize.value
    }

    if (searchForm.status) params.status = searchForm.status
    if (searchForm.urgency) params.urgency = searchForm.urgency
    if (searchForm.fault_type) params.fault_type = searchForm.fault_type
    if (searchForm.dateRange && searchForm.dateRange.length === 2) {
      params.start_date = searchForm.dateRange[0]
      params.end_date = searchForm.dateRange[1]
    }

    const response = await workOrderApi.getWorkOrders(params)
    workOrders.value = response.data.data
    total.value = response.data.total
  } catch (error) {
    console.error('Failed to fetch work orders:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  currentPage.value = 1
  fetchWorkOrders()
}

const handleReset = () => {
  searchForm.status = ''
  searchForm.urgency = ''
  searchForm.fault_type = ''
  searchForm.dateRange = []
  currentPage.value = 1
  fetchWorkOrders()
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  fetchWorkOrders()
}

const handleCurrentChange = (page: number) => {
  currentPage.value = page
  fetchWorkOrders()
}

const goToCreate = () => {
  router.push('/work-orders/create')
}

const goToDetail = (id: number) => {
  router.push(`/work-orders/${id}`)
}

onMounted(() => {
  fetchWorkOrders()
})
</script>

<style scoped>
.work-order-list-container {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-form {
  margin-bottom: 20px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
