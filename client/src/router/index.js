import { createRouter, createWebHistory } from 'vue-router'
import UserForm from '@/components/UserForm.vue'
import UserList from '@/components/UserList.vue'
import AuditLog from '@/components/AuditLog.vue'
import LoginForm from '@/components/LoginForm.vue'
import { userService } from '@/services/api'
import AccountDashboard from '@/components/AccountDashboard.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/users/create',
      name: 'createUser',
      component: UserForm
    },
    {
      path: '/login',
      name: 'login',
      component: LoginForm
    },
    {
      path: '/users',
      name: 'users',
      component: UserList
    },
    {
      path: '/accounts',
      name: 'accounts',
      component: AccountDashboard
    },
    {
      path: '/audits',
      name: 'audits',
      component: AuditLog
    }
  ]
})

router.beforeEach((to, from, next) => {
  const publicPages = ['/login', '/users/create']
  const authRequired = !publicPages.includes(to.path)
  const user = userService.getCurrentUser()

  if (authRequired && !user) {
    return next('/login')
  }

  next()
})

export default router
