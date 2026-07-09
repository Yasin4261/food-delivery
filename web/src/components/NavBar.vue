<script setup>
import { computed } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores/cart'

const auth = useAuthStore()
const cart = useCartStore()
const router = useRouter()

const cartCount = computed(() => cart.count)
const initial = computed(() => (auth.user?.username || auth.user?.email || '?')[0].toUpperCase())

async function logout() {
  await auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <header class="sticky top-0 z-40 border-b border-gray-200 bg-white/80 backdrop-blur">
    <nav class="mx-auto flex max-w-5xl flex-wrap items-center gap-x-4 gap-y-2 px-4 py-3">
      <RouterLink to="/" class="flex items-center gap-2 text-lg font-bold tracking-tight text-gray-900">
        <span class="text-xl">🍲</span>
        <span>Home<span class="text-brand-600">Chef</span></span>
      </RouterLink>

      <div class="flex items-center gap-1">
        <RouterLink v-if="!auth.isChef" to="/" class="nav-link" exact-active-class="router-link-active bg-brand-50">
          Browse
        </RouterLink>
        <RouterLink
          v-if="auth.isAuthenticated && !auth.isChef"
          to="/search"
          class="nav-link"
          exact-active-class="router-link-active bg-brand-50"
        >
          🔍 Search
        </RouterLink>
        <template v-if="auth.isAuthenticated && auth.isChef">
          <RouterLink to="/chef" class="nav-link" exact-active-class="router-link-active bg-brand-50">
            Dashboard
          </RouterLink>
          <RouterLink to="/chef/menus" class="nav-link" exact-active-class="router-link-active bg-brand-50">
            My menus
          </RouterLink>
        </template>
        <RouterLink
          v-if="auth.isAuthenticated && !auth.isChef"
          to="/orders"
          class="nav-link"
          exact-active-class="router-link-active bg-brand-50"
        >
          My orders
        </RouterLink>
        <RouterLink
          v-if="auth.isAuthenticated && !auth.isChef"
          to="/favorites"
          class="nav-link"
          exact-active-class="router-link-active bg-brand-50"
        >
          Favorites
        </RouterLink>
        <RouterLink
          v-if="auth.isAuthenticated"
          to="/chat"
          class="nav-link"
          exact-active-class="router-link-active bg-brand-50"
        >
          💬 Chat
        </RouterLink>
      </div>

      <div class="ml-auto flex items-center gap-3">
        <RouterLink
          v-if="auth.isAuthenticated && !auth.isChef"
          to="/cart"
          class="nav-link relative"
          exact-active-class="router-link-active bg-brand-50"
        >
          🛒 Cart
          <span
            v-if="cartCount"
            class="absolute -right-2 -top-1.5 flex h-5 min-w-5 items-center justify-center rounded-full bg-brand-600 px-1 text-xs font-bold text-white shadow-sm"
            >{{ cartCount }}</span
          >
        </RouterLink>

        <template v-if="auth.isAuthenticated">
          <span
            class="hidden items-center gap-2 rounded-full border border-gray-200 bg-white py-1 pl-1 pr-3 text-sm text-gray-600 sm:flex"
          >
            <span class="flex h-6 w-6 items-center justify-center rounded-full bg-brand-100 text-xs font-bold text-brand-700">
              {{ initial }}
            </span>
            {{ auth.user?.username || auth.user?.email }}
          </span>
          <button class="btn-ghost" @click="logout">Log out</button>
        </template>
        <template v-else>
          <RouterLink to="/login" class="nav-link">Log in</RouterLink>
          <RouterLink to="/register" class="btn-primary">Sign up</RouterLink>
        </template>
      </div>
    </nav>
  </header>
</template>
