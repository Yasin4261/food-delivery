<script setup>
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { api, page } from '@/api/client'
import { useFavoritesStore } from '@/stores/favorites'

const favorites = useFavoritesStore()

const chefs = ref([])
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    chefs.value = page(await api.get('/favorites?limit=100')).items
    favorites.ids = chefs.value.map((c) => c.id)
    favorites.loaded = true
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function unfavorite(chef) {
  try {
    await favorites.toggle(chef.id)
    chefs.value = chefs.value.filter((c) => c.id !== chef.id)
  } catch (e) {
    error.value = e.message
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="page-title">Favorite chefs ❤️</h1>
      <p class="page-subtitle">Your go-to kitchens, one tap away.</p>
    </div>

    <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>

    <div v-if="loading" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <div v-for="n in 3" :key="n" class="skeleton h-36"></div>
    </div>

    <div v-else-if="!chefs.length" class="empty-state">
      <span class="empty-state-emoji">🤍</span>
      <p class="font-medium text-gray-600">No favorites yet</p>
      <p class="text-sm">
        Tap the heart on a chef you love while
        <RouterLink to="/" class="text-brand-600 hover:underline">browsing</RouterLink>.
      </p>
    </div>

    <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <RouterLink v-for="chef in chefs" :key="chef.id" :to="`/chefs/${chef.id}`" class="card-hover relative">
        <button
          class="absolute right-3 top-3 text-lg transition hover:scale-110"
          title="Remove from favorites"
          @click.prevent.stop="unfavorite(chef)"
        >
          ❤️
        </button>
        <div class="flex items-start gap-3 pr-7">
          <div class="avatar relative">
            {{ (chef.business_name || '?')[0].toUpperCase() }}
            <span
              v-if="chef.is_online"
              class="absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-white bg-green-500"
              title="online"
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
