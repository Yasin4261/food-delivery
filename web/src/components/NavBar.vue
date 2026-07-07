<script setup>
import { computed } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores/cart'

const auth = useAuthStore()
const cart = useCartStore()
const router = useRouter()

const cartCount = computed(() => cart.count)

async function logout() {
  await auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <header class="border-b border-gray-200 bg-white">
    <nav class="mx-auto flex max-w-5xl items-center gap-4 px-4 py-3">
      <RouterLink to="/" class="text-lg font-bold text-brand-600">🍲 Home Chef</RouterLink>

      <RouterLink to="/" class="text-sm text-gray-600 hover:text-gray-900">Browse</RouterLink>
      <RouterLink
        v-if="auth.isAuthenticated && auth.isChef"
        to="/chef"
        class="text-sm text-gray-600 hover:text-gray-900"
      >
        Chef dashboard
      </RouterLink>
      <RouterLink
        v-if="auth.isAuthenticated && auth.isChef"
        to="/chef/menus"
        class="text-sm text-gray-600 hover:text-gray-900"
      >
        My menus
      </RouterLink>
      <RouterLink
        v-if="auth.isAuthenticated && !auth.isChef"
        to="/orders"
        class="text-sm text-gray-600 hover:text-gray-900"
      >
        My orders
      </RouterLink>

      <div class="ml-auto flex items-center gap-3">
        <RouterLink v-if="!auth.isChef" to="/cart" class="relative text-sm text-gray-600 hover:text-gray-900">
          Cart
          <span
            v-if="cartCount"
            class="absolute -right-3 -top-2 rounded-full bg-brand-600 px-1.5 text-xs text-white"
            >{{ cartCount }}</span
          >
        </RouterLink>

        <template v-if="auth.isAuthenticated">
          <span class="text-sm text-gray-500">{{ auth.user?.email }}</span>
          <button class="btn-ghost" @click="logout">Log out</button>
        </template>
        <template v-else>
          <RouterLink to="/login" class="text-sm text-gray-600 hover:text-gray-900">Log in</RouterLink>
          <RouterLink to="/register" class="btn-primary">Sign up</RouterLink>
        </template>
      </div>
    </nav>
  </header>
</template>
