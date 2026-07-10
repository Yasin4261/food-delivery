import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

// Everything except the auth screens requires a logged-in user, so an
// unauthenticated visitor lands on the login page first (issue #24).
const routes = [
  { path: '/', name: 'chefs', component: () => import('@/views/ChefsView.vue'), meta: { requiresAuth: true } },
  {
    path: '/chefs/:id',
    name: 'chef-detail',
    component: () => import('@/views/ChefDetailView.vue'),
    props: true,
    meta: { requiresAuth: true },
  },
  { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue'), meta: { guestOnly: true } },
  { path: '/register', name: 'register', component: () => import('@/views/RegisterView.vue'), meta: { guestOnly: true } },
  {
    path: '/forgot-password',
    name: 'forgot-password',
    component: () => import('@/views/ForgotPasswordView.vue'),
    meta: { guestOnly: true },
  },
  {
    path: '/reset-password',
    name: 'reset-password',
    component: () => import('@/views/ResetPasswordView.vue'),
    meta: { guestOnly: true },
  },
  { path: '/cart', name: 'cart', component: () => import('@/views/CartView.vue'), meta: { requiresAuth: true } },
  { path: '/orders', name: 'orders', component: () => import('@/views/OrdersView.vue'), meta: { requiresAuth: true } },
  {
    path: '/favorites',
    name: 'favorites',
    component: () => import('@/views/FavoritesView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/search',
    name: 'search',
    component: () => import('@/views/SearchView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/chat',
    name: 'chat',
    component: () => import('@/views/ChatView.vue'),
    meta: { requiresAuth: true },
  },
  {
    // Dev-only stand-in for the gateway's hosted payment page (mock gateway).
    path: '/mock-pay',
    name: 'mock-pay',
    component: () => import('@/views/MockPayView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/chef',
    name: 'chef-dashboard',
    component: () => import('@/views/ChefDashboardView.vue'),
    meta: { requiresAuth: true, role: 'chef' },
  },
  {
    path: '/chef/menus',
    name: 'chef-menus',
    component: () => import('@/views/ChefMenusView.vue'),
    meta: { requiresAuth: true, role: 'chef' },
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

// home returns the landing route for a logged-in user by role.
function home(auth) {
  return auth.isChef ? { name: 'chef-dashboard' } : { name: 'chefs' }
}

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.meta.role && auth.role !== to.meta.role) {
    return home(auth)
  }
  if (to.meta.guestOnly && auth.isAuthenticated) {
    return home(auth)
  }
  return true
})
