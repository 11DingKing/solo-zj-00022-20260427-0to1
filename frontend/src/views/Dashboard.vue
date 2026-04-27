<template>
  <div class="dashboard-container">
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stats-icon" style="background: #409EFF;">
            <el-icon :size="30"><Document /></el-icon>
          </div>
          <div class="stats-info">
            <div class="stats-value">{{ stats.today_new_orders }}</div>
            <div class="stats-label">今日新增工单</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stats-icon" style="background: #E6A23C;">
            <el-icon :size="30"><Clock /></el-icon>
          </div>
          <div class="stats-info">
            <div class="stats-value">{{ stats.pending_orders }}</div>
            <div class="stats-label">待处理工单</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stats-icon" style="background: #67C23A;">
            <el-icon :size="30"><Timer /></el-icon>
          </div>
          <div class="stats-info">
            <div class="stats-value">{{ avgTime }}</div>
            <div class="stats-label">平均处理时长(分钟)</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stats-icon" style="background: #F56C6C;">
            <el-icon :size="30"><TrendCharts /></el-icon>
          </div>
          <div class="stats-info">
            <div class="stats-value">{{ technicianCount }}</div>
            <div class="stats-label">在职技师人数</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>各故障类型分布</span>
          </template>
          <div ref="pieChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>近30天工单趋势</span>
          </template>
          <div ref="lineChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="24">
        <el-card>
          <template #header>
            <span>技师工作量排行榜</span>
          </template>
          <el-table :data="technicianRanking" style="width: 100%">
            <el-table-column type="index" label="排名" width="80" align="center">
              <template #default="{ $index }">
                <el-tag v-if="$index === 0" type="danger" effect="dark">1</el-tag>
                <el-tag v-else-if="$index === 1" type="warning" effect="dark">2</el-tag>
                <el-tag v-else-if="$index === 2" type="success" effect="dark">3</el-tag>
                <span v-else>{{ $index + 1 }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="technician_name" label="技师姓名" />
            <el-table-column prop="completed_count" label="完成工单数">
              <template #default="{ row }">
                <el-badge :value="row.completed_count" class="item" :type="row.completed_count > 0 ? 'primary' : 'info'" />
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, nextTick } from 'vue'
import * as echarts from 'echarts'
import { dashboardApi } from '@/api'
import type { DashboardStats, FaultTypeDistribution, DailyTrend, TechnicianRanking } from '@/types'
import { Document, Clock, Timer, TrendCharts } from '@element-plus/icons-vue'

const pieChartRef = ref<HTMLElement>()
const lineChartRef = ref<HTMLElement>()

const stats = ref<DashboardStats>({
  today_new_orders: 0,
  pending_orders: 0,
  avg_processing_time: 0
})

const faultTypeDistribution = ref<FaultTypeDistribution[]>([])
const last30DaysTrend = ref<DailyTrend[]>([])
const technicianRanking = ref<TechnicianRanking[]>([])

const avgTime = computed(() => {
  return stats.value.avg_processing_time.toFixed(1)
})

const technicianCount = computed(() => {
  return technicianRanking.value.length
})

const faultTypeMap: Record<string, string> = {
  hardware: '硬件故障',
  software: '软件问题',
  network: '网络异常',
  other: '其他'
}

const fetchData = async () => {
  try {
    const [statsRes, faultRes, trendRes, rankingRes] = await Promise.all([
      dashboardApi.getStats(),
      dashboardApi.getFaultTypeDistribution(),
      dashboardApi.getLast30DaysTrend(),
      dashboardApi.getTechnicianRanking()
    ])

    stats.value = statsRes.data
    faultTypeDistribution.value = faultRes.data
    last30DaysTrend.value = trendRes.data
    technicianRanking.value = rankingRes.data
  } catch (error) {
    console.error('Failed to fetch dashboard data:', error)
  }
}

const initPieChart = () => {
  if (!pieChartRef.value) return

  const chart = echarts.init(pieChartRef.value)

  const data = faultTypeDistribution.value.map(item => ({
    name: faultTypeMap[item.fault_type] || item.fault_type,
    value: item.count
  }))

  const option = {
    tooltip: {
      trigger: 'item'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: false,
          position: 'center'
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 20,
            fontWeight: 'bold'
          }
        },
        labelLine: {
          show: false
        },
        data: data
      }
    ]
  }

  chart.setOption(option)
}

const initLineChart = () => {
  if (!lineChartRef.value) return

  const chart = echarts.init(lineChartRef.value)

  const dates = last30DaysTrend.value.map(item => item.date)
  const counts = last30DaysTrend.value.map(item => item.count)

  const option = {
    tooltip: {
      trigger: 'axis'
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: dates
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '工单数量',
        type: 'line',
        stack: 'Total',
        data: counts,
        smooth: true,
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(80, 141, 255, 0.5)' },
            { offset: 1, color: 'rgba(80, 141, 255, 0.05)' }
          ])
        },
        lineStyle: {
          color: '#409EFF',
          width: 2
        }
      }
    ]
  }

  chart.setOption(option)
}

onMounted(async () => {
  await fetchData()
  await nextTick()
  initPieChart()
  initLineChart()
})
</script>

<style scoped>
.dashboard-container {
  padding: 0;
}

.stats-row {
  margin-bottom: 20px;
}

.stats-card {
  display: flex;
  align-items: center;
  padding: 20px;
}

.stats-icon {
  width: 60px;
  height: 60px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.stats-info {
  margin-left: 20px;
}

.stats-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
}

.stats-label {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}

.chart-container {
  height: 350px;
  width: 100%;
}
</style>
