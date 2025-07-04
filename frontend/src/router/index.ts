import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import Login from '../views/Login.vue'
import Register from '../views/Register.vue'
import MainLayout from '../views/MainLayout.vue'
import UserManagement from '../views/UserManagement.vue'
import ChangePassword from '../views/ChangePassword.vue'
import { useMainStore } from '../store'

const routes: Array<RouteRecordRaw> = [
  { path: '/login', name: 'Login', component: Login },
  { path: '/register', name: 'Register', component: Register },
  { 
    path: '/', 
    name: 'MainLayout', 
    component: MainLayout,
    meta: { requiresAuth: true },
    children: [
      { 
        path: 'users', 
        name: 'UserManagement', 
        component: UserManagement, 
        meta: { requiresAdmin: true } 
      },
      { 
        path: 'changepwd', 
        name: 'ChangePassword', 
        component: ChangePassword 
      }
    ]
  },
  // 可以在这里添加其他需要权限的路由
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, _from, next) => {
  const store = useMainStore()
  if (to.meta.requiresAuth && !store.token) {
    next('/login')
  } else if (to.meta.requiresAdmin && !store.isAdmin) {
    next({ name: 'MainLayout' })
  } else {
    next()
  }
})

export default router 