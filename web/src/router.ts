import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from './views/DashboardView.vue'
import LoginView from './views/LoginView.vue'
import RegisterView from './views/RegisterView.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/login' },
    { path: '/login', component: LoginView, meta: { title: 'Login' } },
    { path: '/register', component: RegisterView, meta: { title: 'Register' } },
    { path: '/dashboard', component: DashboardView, meta: { title: 'Dashboard' } },
  ],
})
