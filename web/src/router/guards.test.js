import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { router } from './index'
import { useAuthStore } from '@/stores/auth'

// The guards read the auth store at navigation time, so a fresh pinia per test
// gives each test a clean session.
describe('router guards', () => {
  beforeEach(async () => {
    localStorage.clear()
    setActivePinia(createPinia())
    // Park on a guest route so each test starts from neutral ground.
    await router.push('/login').catch(() => {})
    await router.isReady()
  })

  it('login shows first: anonymous visits to / are redirected (issue #24)', async () => {
    await router.push('/')
    expect(router.currentRoute.value.name).toBe('login')
    expect(router.currentRoute.value.query.redirect).toBe('/')
  })

  it('protected routes redirect anonymous users with a return path', async () => {
    await router.push('/orders')
    expect(router.currentRoute.value.name).toBe('login')
    expect(router.currentRoute.value.query.redirect).toBe('/orders')
  })

  it('customers browse; guest-only pages bounce them home', async () => {
    const auth = useAuthStore()
    auth.token = 'jwt'
    auth.user = { id: 1, role: 'customer' }

    await router.push('/')
    expect(router.currentRoute.value.name).toBe('chefs')

    await router.push('/login')
    expect(router.currentRoute.value.name).toBe('chefs')
  })

  it('role guard: customers cannot enter the chef dashboard', async () => {
    const auth = useAuthStore()
    auth.token = 'jwt'
    auth.user = { id: 1, role: 'customer' }

    await router.push('/chef')
    expect(router.currentRoute.value.name).toBe('chefs')
  })

  it('chefs land on their dashboard from guest-only pages and may enter /chef', async () => {
    const auth = useAuthStore()
    auth.token = 'jwt'
    auth.user = { id: 2, role: 'chef' }

    // Move off /login first — pushing the current route is a no-op navigation.
    await router.push('/chef/menus')
    expect(router.currentRoute.value.name).toBe('chef-menus')

    await router.push('/login')
    expect(router.currentRoute.value.name).toBe('chef-dashboard')
  })
})
