<template>
  <el-container class="layout-container">
    <el-aside :width="isCollapse ? '64px' : '220px'" class="layout-aside">
      <div class="logo">
        <el-icon v-if="isCollapse" :size="32"><Tools /></el-icon>
        <span v-else class="logo-text">设备报修系统</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        :unique-opened="true"
        router
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
      >
        <el-menu-item index="/dashboard" v-if="userStore.isAdmin">
          <el-icon><DataAnalysis /></el-icon>
          <span>数据看板</span>
        </el-menu-item>

        <el-menu-item index="/work-orders">
          <el-icon><Document /></el-icon>
          <span>工单列表</span>
        </el-menu-item>

        <el-menu-item index="/work-orders/create" v-if="userStore.isEmployee">
          <el-icon><Edit /></el-icon>
          <span>提交报修</span>
        </el-menu-item>

        <el-menu-item index="/devices">
          <el-icon><Monitor /></el-icon>
          <span>设备台账</span>
        </el-menu-item>

        <el-menu-item index="/users" v-if="userStore.isAdmin">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="layout-header">
        <el-icon class="collapse-icon" @click="toggleCollapse">
          <Fold v-if="!isCollapse" />
          <Expand v-else />
        </el-icon>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-icon><UserFilled /></el-icon>
              <span>{{ userStore.user?.real_name }}</span>
              <el-tag :type="roleTagType" size="small" effect="plain">{{ userStore.roleName }}</el-tag>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人信息</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="layout-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/store/user'
import {
  Fold,
  Expand,
  DataAnalysis,
  Document,
  Edit,
  Monitor,
  User,
  UserFilled,
  Tools
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const isCollapse = ref(false)

const activeMenu = computed(() => route.path)

const roleTagType = computed(() => {
  const typeMap: Record<string, string> = {
    admin: 'danger',
    technician: 'warning',
    employee: 'primary'
  }
  return userStore.user ? typeMap[userStore.user.role] : 'info'
})

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

const handleCommand = (command: string) => {
  if (command === 'logout') {
    userStore.logout()
    router.push('/login')
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.layout-aside {
  background-color: #304156;
  transition: width 0.3s;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #2b3a4a;
  color: white;
}

.logo-text {
  font-size: 18px;
  font-weight: bold;
}

.el-menu {
  border-right: none;
}

.layout-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background-color: white;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  padding: 0 20px;
}

.collapse-icon {
  font-size: 20px;
  cursor: pointer;
  color: #606266;
}

.collapse-icon:hover {
  color: #409EFF;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #606266;
}

.user-info:hover {
  color: #409EFF;
}

.layout-main {
  background-color: #f0f2f5;
  padding: 20px;
  overflow-y: auto;
}
</style>
