<script setup>
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { api, page } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { useFavoritesStore } from '@/stores/favorites'

const auth = useAuthStore()
const favorites = useFavoritesStore()

const chefs = ref([])
const loading = ref(true)
const error = ref('')

// Optional "deliver to me" filter by coordinates.
const lat = ref('')
const lng = ref('')
const nearbyActive = ref(false)

// Customers see and use the favorite hearts; chefs don't favorite themselves.
const canFavorite = auth.isAuthenticated && !auth.isChef

async function toggleFavorite(chef) {
  try {
    await favorites.toggle(chef.id)
  } catch (e) {
    error.value = e.message
  }
}

async function loadAll() {
  loading.value = true
  error.value = ''
  nearbyActive.value = false
  try {
    chefs.value = page(await api.get('/chefs?limit=50')).items
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function loadNearby() {
  if (!lat.value || !lng.value) return loadAll()
  loading.value = true
  error.value = ''
  try {
    chefs.value = page(await api.get(`/chefs/nearby?lat=${lat.value}&lng=${lng.value}&limit=50`)).items
    nearbyActive.value = true
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

const initial = (chef) => (chef.business_name || '?')[0].toUpperCase()

onMounted(() => {
  loadAll()
  if (canFavorite) favorites.load().catch(() => {})
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 class="page-title">{{ $t('browse.title') }}</h1>
        <p class="page-subtitle">{{ $t('browse.subtitle') }}</p>
      </div>
      <form class="flex items-end gap-2" @submit.prevent="loadNearby">
        <div>
          <label class="label">{{ $t('browse.lat') }}</label>
          <input v-model="lat" class="input w-28" placeholder="41.0082" />
        </div>
        <div>
          <label class="label">{{ $t('browse.lng') }}</label>
          <input v-model="lng" class="input w-28" placeholder="28.9784" />
        </div>
        <button class="btn-ghost">{{ $t('browse.nearby') }}</button>
        <button v-if="nearbyActive" type="button" class="btn-ghost" @click="loadAll">{{ $t('browse.showAll') }}</button>
      </form>
    </div>

    <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>

    <!-- Loading skeletons -->
    <div v-if="loading" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <div v-for="n in 6" :key="n" class="skeleton h-36"></div>
    </div>

    <div v-else-if="!chefs.length" class="empty-state">
      <span class="empty-state-emoji">🍳</span>
      <p class="font-medium text-gray-600">{{ $t('browse.empty') }}</p>
      <p class="text-sm">{{ $t('browse.emptyHint') }}</p>
    </div>

    <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <RouterLink v-for="chef in chefs" :key="chef.id" :to="`/chefs/${chef.id}`" class="card-hover relative">
        <button
          v-if="canFavorite"
          class="absolute right-3 top-3 text-lg transition hover:scale-110"
          :title="favorites.has(chef.id) ? $t('favorites.removeTitle') : $t('favorites.addTitle')"
          @click.prevent.stop="toggleFavorite(chef)"
        >
          {{ favorites.has(chef.id) ? '❤️' : '🤍' }}
        </button>
        <div class="flex items-start gap-3" :class="canFavorite && 'pr-7'">
          <div class="avatar relative">
            {{ initial(chef) }}
            <span
              v-if="chef.is_online"
              class="absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-white bg-green-500"
              :title="$t('dashboard.online')"
            ></span>
          </div>
          <div class="min-w-0">
            <h2 class="truncate font-semibold">{{ chef.business_name }}</h2>
            <p v-if="chef.specialty" class="truncate text-sm text-gray-500">{{ chef.specialty }}</p>
          </div>
        </div>
        <div class="mt-3 flex items-center justify-between text-sm">
          <span class="truncate text-gray-500">📍 {{ chef.kitchen_city || chef.kitchen_address }}</span>
          <span class="badge bg-amber-50 text-amber-700">★ {{ chef.rating?.toFixed(1) ?? '—' }} ({{ chef.total_reviews }})</span>
        </div>
      </RouterLink>
    </div>
  </div>
</template>
