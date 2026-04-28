import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/store/user'

const getRedirectPath = (userStore: ReturnType<typeof useUserStore>) => {
  if (userStore.isAdmin) {
    return '/dashboard'
  } else if (userStore.isTechnician) {
    return '/work-orders'
  } else if (userStore.isEmployee) {
    return '/work-orders/create'
  }
  return '/work-orders'
}

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/403.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/404',
    name: 'NotFound',
    component: () => import('@/views/404.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/views/Layout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: (to) => {
          const userStore = useUserStore()
          userStore.initFromStorage()
          return getRedirectPath(userStore)
        }
      },
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { requiresAuth: true, roles: ['admin'] }
      },
      {
        path: 'work-orders',
        name: 'WorkOrders',
        component: () => import('@/views/WorkOrderList.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'work-orders/create',
        name: 'CreateWorkOrder',
        component: () => import('@/views/CreateWorkOrder.vue'),
        meta: { requiresAuth: true, roles: ['employee'] }
      },
      {
        path: 'work-orders/:id',
        name: 'WorkOrderDetail',
        component: () => import('@/views/WorkOrderDetail.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'devices',
        name: 'Devices',
        component: () => import('@/views/DeviceList.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'devices/:id',
        name: 'DeviceDetail',
        component: () => import('@/views/DeviceDetail.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/UserList.vue'),
        meta: { requiresAuth: true, roles: ['admin'] }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/404'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to, _from, next) => {
  const userStore = useUserStore()

  if (to.meta.requiresAuth !== false && !userStore.isLoggedIn) {
    userStore.initFromStorage()
    if (!userStore.isLoggedIn) {
      next({ name: 'Login', query: { redirect: to.fullPath } })
      return
    }
  }

  if (to.name === 'Login' && userStore.isLoggedIn) {
    next({ path: getRedirectPath(userStore) })
    return
  }

  if (to.meta.roles && Array.isArray(to.meta.roles)) {
    const roles = to.meta.roles as string[]
    if (userStore.user && !roles.includes(userStore.user.role)) {
      next({ path: '/403' })
      return
    }
  }

  next()
})

export default router
