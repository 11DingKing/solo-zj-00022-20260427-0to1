import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'
import type {
  User,
  Device,
  WorkOrder,
  DashboardStats,
  FaultTypeDistribution,
  DailyTrend,
  TechnicianRanking,
  PaginatedResponse
} from '@/types'

const api: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

api.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export const authApi = {
  login: (username: string, password: string) => {
    return api.post('/auth/login', { username, password })
  },

  getCurrentUser: () => {
    return api.get<User>('/user/me')
  }
}

export const userApi = {
  getUsers: (params?: { page?: number; page_size?: number; role?: string; keyword?: string }) => {
    return api.get<PaginatedResponse<User>>('/users', { params })
  },

  getUserById: (id: number) => {
    return api.get<User>(`/users/${id}`)
  },

  createUser: (data: { username: string; password: string; real_name: string; role: string; phone?: string }) => {
    return api.post<User>('/users', data)
  },

  updateUser: (id: number, data: { real_name?: string; role?: string; phone?: string; password?: string; status?: string }) => {
    return api.put<User>(`/users/${id}`, data)
  },

  deleteUser: (id: number) => {
    return api.delete(`/users/${id}`)
  },

  resetPassword: (id: number) => {
    return api.put(`/users/${id}/reset-password`)
  },

  getTechnicians: () => {
    return api.get<User[]>('/technicians')
  }
}

export const deviceApi = {
  getDevices: (params?: { page?: number; page_size?: number; keyword?: string; status?: string }) => {
    return api.get<PaginatedResponse<Device>>('/devices', { params })
  },

  getDeviceById: (id: number) => {
    return api.get<{ device: Device; work_orders: WorkOrder[] }>(`/devices/${id}`)
  },

  createDevice: (data: {
    device_code: string
    name: string
    model?: string
    location?: string
    purchase_date?: string
    warranty_expire_date?: string
  }) => {
    return api.post<Device>('/devices', data)
  },

  updateDevice: (id: number, data: {
    name?: string
    model?: string
    location?: string
    purchase_date?: string
    warranty_expire_date?: string
    status?: string
  }) => {
    return api.put<Device>(`/devices/${id}`, data)
  },

  deleteDevice: (id: number) => {
    return api.delete(`/devices/${id}`)
  }
}

export const workOrderApi = {
  getWorkOrders: (params?: {
    page?: number
    page_size?: number
    status?: string
    urgency?: string
    fault_type?: string
    start_date?: string
    end_date?: string
  }) => {
    return api.get<PaginatedResponse<WorkOrder>>('/work-orders', { params })
  },

  getWorkOrderById: (id: number) => {
    return api.get<WorkOrder>(`/work-orders/${id}`)
  },

  createWorkOrder: (data: {
    device_id: number
    fault_type: string
    fault_description: string
    urgency?: string
    before_image_ids?: number[]
  }) => {
    return api.post<WorkOrder>('/work-orders', data)
  },

  assignWorkOrder: (id: number, technicianId: number) => {
    return api.put<WorkOrder>(`/work-orders/${id}/assign`, { technician_id: technicianId })
  },

  acceptWorkOrder: (id: number) => {
    return api.put<WorkOrder>(`/work-orders/${id}/accept`)
  },

  submitRepair: (id: number, data: {
    repair_measures: string
    replaced_parts?: string
    repair_duration?: number
    after_image_ids?: number[]
  }) => {
    return api.put<WorkOrder>(`/work-orders/${id}/submit`, data)
  },

  confirmWorkOrder: (id: number) => {
    return api.put<WorkOrder>(`/work-orders/${id}/confirm`)
  },

  rejectWorkOrder: (id: number) => {
    return api.put<WorkOrder>(`/work-orders/${id}/reject`)
  }
}

export const imageApi = {
  uploadImage: (file: File, onUploadProgress?: (progressEvent: any) => void) => {
    const formData = new FormData()
    formData.append('image', file)
    return api.post<{
      id: number
      file_path: string
      file_name: string
      file_size: number
      access_url: string
    }>('/images/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      },
      onUploadProgress
    })
  }
}

export const dashboardApi = {
  getStats: () => {
    return api.get<DashboardStats>('/dashboard/stats')
  },

  getFaultTypeDistribution: () => {
    return api.get<FaultTypeDistribution[]>('/dashboard/fault-type-distribution')
  },

  getLast30DaysTrend: () => {
    return api.get<DailyTrend[]>('/dashboard/30-days-trend')
  },

  getTechnicianRanking: () => {
    return api.get<TechnicianRanking[]>('/dashboard/technician-ranking')
  }
}

export default api
