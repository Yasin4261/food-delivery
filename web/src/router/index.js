import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  { path: '/', name: 'chefs', component: () => import('@/views/ChefsView.vue') },
  { path: '/chefs/:id', name: 'chef-detail', component: () => import('@/views/ChefDetailView.vue'), props: true },
  { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue'), meta: { guestOnly: true } },
  { path: '/register', name: 'register', component: () => import('@/views/RegisterView.vue'), meta: { guestOnly: true } },
  { path: '/cart', name: 'cart', component: () => import('@/views/CartView.vue'), meta: { requiresAuth: true } },
  { path: '/orders', name: 'orders', component: () => import('@/views/OrdersView.vue'), meta: { requiresAuth: true } },
  {
    path: '/chef',
    name: 'chef-dashboard',
    component: () => import('@/views/ChefDashboardView.vue'),
    meta: { requiresAuth: true, role: 'chef' },
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.meta.role && auth.role !== to.meta.role) {
    return { name: 'chefs' }
  }
  if (to.meta.guestOnly && auth.isAuthenticated) {
    return { name: auth.isChef ? 'chef-dashboard' : 'chefs' }
  }
  return true
})
