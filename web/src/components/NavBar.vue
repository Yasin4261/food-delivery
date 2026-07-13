<script setup>
import { computed, watch } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores/cart'
import { useNotificationsStore } from '@/stores/notifications'
import { setLocale } from '@/i18n'

const auth = useAuthStore()
const cart = useCartStore()
const notifications = useNotificationsStore()
const router = useRouter()
const { locale } = useI18n()

const cartCount = computed(() => cart.count)
const initial = computed(() => (auth.user?.username || auth.user?.email || '?')[0].toUpperCase())
const otherLocale = computed(() => (locale.value === 'tr' ? 'en' : 'tr'))

// Poll the badge counts while a session is active.
watch(
  () => auth.isAuthenticated,
  (loggedIn) => (loggedIn ? notifications.start() : notifications.stop()),
  { immediate: true },
)

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
          {{ $t('nav.browse') }}
        </RouterLink>
        <RouterLink
          v-if="auth.isAuthenticated && !auth.isChef"
          to="/search"
          class="nav-link"
          exact-active-class="router-link-active bg-brand-50"
        >
          {{ $t('nav.search') }}
        </RouterLink>
        <template v-if="auth.isAuthenticated && auth.isChef">
          <RouterLink to="/chef" class="nav-link relative" exact-active-class="router-link-active bg-brand-50">
            {{ $t('nav.dashboard') }}
            <span
              v-if="notifications.pendingChefOrders"
              class="absolute -right-2 -top-1.5 flex h-5 min-w-5 items-center justify-center rounded-full bg-red-500 px-1 text-xs font-bold text-white shadow-sm"
              :title="$t('nav.pendingTitle')"
              >{{ notifications.pendingChefOrders }}</span
            >
          </RouterLink>
          <RouterLink to="/chef/menus" class="nav-link" exact-active-class="router-link-active bg-brand-50">
            {{ $t('nav.myMenus') }}
          </RouterLink>
        </template>
        <RouterLink
          v-if="auth.isAuthenticated && !auth.isChef"
          to="/orders"
          class="nav-link relative"
          exact-active-class="router-link-active bg-brand-50"
        >
          {{ $t('nav.myOrders') }}
          <span
            v-if="notifications.activeOrders"
            class="absolute -right-2 -top-1.5 flex h-5 min-w-5 items-center justify-center rounded-full bg-brand-600 px-1 text-xs font-bold text-white shadow-sm"
            :title="$t('nav.activeTitle')"
            >{{ notifications.activeOrders }}</span
          >
        </RouterLink>
        <RouterLink
          v-if="auth.isAuthenticated && !auth.isChef"
          to="/favorites"
          class="nav-link"
          exact-active-class="router-link-active bg-brand-50"
        >
          {{ $t('nav.favorites') }}
        </RouterLink>
        <RouterLink
          v-if="auth.isAuthenticated"
          to="/chat"
          class="nav-link"
          exact-active-class="router-link-active bg-brand-50"
        >
          {{ $t('nav.chat') }}
        </RouterLink>
      </div>

      <div class="ml-auto flex items-center gap-3">
        <button
          class="rounded-md border border-gray-200 px-2 py-1 text-xs font-semibold uppercase text-gray-500 transition hover:bg-gray-50 hover:text-gray-800"
          :title="otherLocale === 'tr' ? 'Türkçe' : 'English'"
          @click="setLocale(otherLocale)"
        >
          {{ otherLocale }}
        </button>

        <RouterLink
          v-if="auth.isAuthenticated && !auth.isChef"
          to="/cart"
          class="nav-link relative"
          exact-active-class="router-link-active bg-brand-50"
        >
          {{ $t('nav.cart') }}
          <span
            v-if="cartCount"
            class="absolute -right-2 -top-1.5 flex h-5 min-w-5 items-center justify-center rounded-full bg-brand-600 px-1 text-xs font-bold text-white shadow-sm"
            >{{ cartCount }}</span
          >
        </RouterLink>

        <template v-if="auth.isAuthenticated">
          <RouterLink
            to="/profile"
            class="hidden items-center gap-2 rounded-full border border-gray-200 bg-white py-1 pl-1 pr-3 text-sm text-gray-600 transition hover:border-brand-200 hover:text-brand-700 sm:flex"
            :title="$t('nav.profileTitle')"
          >
            <span class="flex h-6 w-6 items-center justify-center rounded-full bg-brand-100 text-xs font-bold text-brand-700">
              {{ initial }}
            </span>
            {{ auth.user?.username || auth.user?.email }}
          </RouterLink>
          <button class="btn-ghost" @click="logout">{{ $t('nav.logout') }}</button>
        </template>
        <template v-else>
          <RouterLink to="/login" class="nav-link">{{ $t('nav.login') }}</RouterLink>
          <RouterLink to="/register" class="btn-primary">{{ $t('nav.signup') }}</RouterLink>
        </template>
      </div>
    </nav>
  </header>
</template>
