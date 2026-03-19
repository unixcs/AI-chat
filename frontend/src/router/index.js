import { createRouter, createWebHistory } from 'vue-router'
import AdminLayout from '../views/admin/AdminLayout.vue'
import UserLayout from '../views/user/UserLayout.vue'
import LandingView from '../views/public/LandingView.vue'
import LoginView from '../views/public/LoginView.vue'
import RegisterView from '../views/public/RegisterView.vue'
import ChatView from '../views/user/ChatView.vue'
import ProfileView from '../views/user/ProfileView.vue'
import SettingsView from '../views/user/SettingsView.vue'
import AdminLoginView from '../views/admin/AdminLoginView.vue'
import AdminDashboardView from '../views/admin/AdminDashboardView.vue'
import AdminUserView from '../views/admin/AdminUserView.vue'
import AdminRoleView from '../views/admin/AdminRoleView.vue'
import AdminMenuView from '../views/admin/AdminMenuView.vue'
import AdminMemberView from '../views/admin/AdminMemberView.vue'
import AdminRedeemCodeView from '../views/admin/AdminRedeemCodeView.vue'
import AdminRedeemRecordView from '../views/admin/AdminRedeemRecordView.vue'
import AdminConversationView from '../views/admin/AdminConversationView.vue'
import { useAuthStore } from '../stores/auth'

const routes = [
  { path: '/', name: 'landing', component: LandingView },
  { path: '/login', name: 'login', component: LoginView, meta: { guestOnly: true } },
  { path: '/register', name: 'register', component: RegisterView, meta: { guestOnly: true } },
  {
    path: '/app',
    component: UserLayout,
    meta: { requiresUser: true },
    children: [
      { path: 'chat', name: 'chat', component: ChatView },
      { path: 'profile', name: 'profile', component: ProfileView },
      { path: 'settings', name: 'settings', component: SettingsView }
    ]
  },
  { path: '/admin/login', name: 'adminLogin', component: AdminLoginView, meta: { guestOnlyAdmin: true } },
  {
    path: '/admin',
    component: AdminLayout,
    meta: { requiresAdmin: true },
    children: [
      { path: '', name: 'adminDashboard', component: AdminDashboardView },
      { path: 'users', name: 'adminUsers', component: AdminUserView },
      { path: 'roles', name: 'adminRoles', component: AdminRoleView },
      { path: 'menus', name: 'adminMenus', component: AdminMenuView },
      { path: 'members', name: 'adminMembers', component: AdminMemberView },
      { path: 'redeem-codes', name: 'adminRedeemCodes', component: AdminRedeemCodeView },
      { path: 'redeem-records', name: 'adminRedeemRecords', component: AdminRedeemRecordView },
      { path: 'conversations', name: 'adminConversations', component: AdminConversationView }
    ]
  },
  { path: '/:pathMatch(.*)*', redirect: '/' }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to) => {
  const authStore = useAuthStore()

  if (to.meta.requiresUser && !authStore.userToken) {
    return { name: 'login' }
  }

  if (to.meta.requiresAdmin && !authStore.adminToken) {
    return { name: 'adminLogin' }
  }

  if (to.meta.guestOnly && authStore.userToken) {
    return { name: 'chat' }
  }

  if (to.meta.guestOnlyAdmin && authStore.adminToken) {
    return { name: 'adminDashboard' }
  }

  return true
})

export default router
