export type UserRole = 'employee' | 'technician' | 'admin'
export type UserStatus = 'active' | 'inactive'

export interface User {
  id: number
  username: string
  real_name: string
  role: UserRole
  status: UserStatus
  phone: string
  created_at: string
  updated_at: string
}

export type DeviceStatus = 'active' | 'maintenance' | 'scrapped'

export interface Device {
  id: number
  device_code: string
  name: string
  model: string
  location: string
  purchase_date?: string
  warranty_expire_date?: string
  status: DeviceStatus
  created_at: string
  updated_at: string
  work_orders?: WorkOrder[]
}

export type WorkOrderStatus = 'pending_assign' | 'assigned' | 'processing' | 'pending_confirm' | 'closed'
export type FaultType = 'hardware' | 'software' | 'network' | 'other'
export type UrgencyLevel = 'low' | 'medium' | 'high' | 'urgent'

export interface WorkOrder {
  id: number
  order_number: string
  device_id: number
  employee_id: number
  technician_id?: number
  fault_type: FaultType
  fault_description: string
  urgency: UrgencyLevel
  status: WorkOrderStatus
  repair_measures?: string
  replaced_parts?: string
  repair_duration?: number
  created_at: string
  updated_at: string
  device?: Device
  employee?: User
  technician?: User
  images?: Image[]
  logs?: OperationLog[]
}

export type ImageType = 'before' | 'after'

export interface Image {
  id: number
  work_order_id: number
  image_type: ImageType
  file_path: string
  file_name: string
  file_size: number
  created_at: string
}

export type OperationType = 'create' | 'assign' | 'accept' | 'process' | 'submit' | 'confirm' | 'reject'

export interface OperationLog {
  id: number
  work_order_id: number
  user_id: number
  operation: OperationType
  old_status?: WorkOrderStatus
  new_status?: WorkOrderStatus
  remark?: string
  created_at: string
  user?: User
}

export interface DashboardStats {
  today_new_orders: number
  pending_orders: number
  avg_processing_time: number
}

export interface FaultTypeDistribution {
  fault_type: string
  count: number
}

export interface DailyTrend {
  date: string
  count: number
}

export interface TechnicianRanking {
  technician_id: number
  technician_name: string
  completed_count: number
}

export interface ApiResponse<T = any> {
  data?: T
  message?: string
  error?: string
  total?: number
  page?: number
  page_size?: number
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  page_size: number
}
