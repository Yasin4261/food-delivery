<script setup>
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { api, page } from '@/api/client'

const chefs = ref([])
const loading = ref(true)
const error = ref('')

// Optional "deliver to me" filter by coordinates.
const lat = ref('')
const lng = ref('')

async function loadAll() {
  loading.value = true
  error.value = ''
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
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(loadAll)
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <h1 class="text-2xl font-bold">Chefs near you</h1>
      <form class="flex items-end gap-2" @submit.prevent="loadNearby">
        <div>
          <label class="label">Lat</label>
          <input v-model="lat" class="input w-28" placeholder="41.0082" />
        </div>
        <div>
          <label class="label">Lng</label>
          <input v-model="lng" class="input w-28" placeholder="28.9784" />
        </div>
        <button class="btn-ghost">Find nearby</button>
      </form>
    </div>

    <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
    <p v-if="loading" class="text-gray-500">Loading…</p>
    <p v-else-if="!chefs.length" class="text-gray-500">No chefs found.</p>

    <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <RouterLink v-for="chef in chefs" :key="chef.id" :to="`/chefs/${chef.id}`" class="card hover:border-brand-500">
        <div class="flex items-start justify-between">
          <h2 class="font-semibold">{{ chef.business_name }}</h2>
          <span v-if="chef.is_online" class="badge bg-green-100 text-green-700">online</span>
        </div>
        <p v-if="chef.specialty" class="text-sm text-gray-500">{{ chef.specialty }}</p>
        <p class="mt-2 text-sm text-gray-600">{{ chef.kitchen_city || chef.kitchen_address }}</p>
        <p class="mt-1 text-sm text-amber-600">★ {{ chef.rating?.toFixed(1) ?? '—' }} ({{ chef.total_reviews }})</p>
      </RouterLink>
    </div>
  </div>
</template>
