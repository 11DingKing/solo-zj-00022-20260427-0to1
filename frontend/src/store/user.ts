import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, UserRole } from '@/types'
import { authApi } from '@/api'

export const useUserStore = defineStore('user', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<User | null>(null)

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isTechnician = computed(() => user.value?.role === 'technician')
  const isEmployee = computed(() => user.value?.role === 'employee')

  const roleName = computed(() => {
    const roleMap: Record<UserRole, string> = {
      employee: '普通员工',
      technician: '维修技师',
      admin: '管理员'
    }
    return user.value ? roleMap[user.value.role] : ''
  })

  async function login(username: string, password: string) {
    const response = await authApi.login(username, password)
    const { token: newToken, user: userData } = response.data
    token.value = newToken
    user.value = userData
    localStorage.setItem('token', newToken)
    localStorage.setItem('user', JSON.stringify(userData))
    return response.data
  }

  async function fetchCurrentUser() {
    if (!token.value) {
      throw new Error('No token')
    }
    const response = await authApi.getCurrentUser()
    user.value = response.data
    localStorage.setItem('user', JSON.stringify(response.data))
    return response.data
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  function initFromStorage() {
    const storedUser = localStorage.getItem('user')
    if (storedUser) {
      try {
        user.value = JSON.parse(storedUser)
      } catch {
        user.value = null
      }
    }
  }

  return {
    token,
    user,
    isLoggedIn,
    isAdmin,
    isTechnician,
    isEmployee,
    roleName,
    login,
    fetchCurrentUser,
    logout,
    initFromStorage
  }
})
