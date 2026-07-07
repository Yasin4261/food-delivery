<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, page } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores/cart'

const props = defineProps({ id: { type: String, required: true } })
const auth = useAuthStore()
const cart = useCartStore()
const router = useRouter()

const chef = ref(null)
const items = ref([])
const loading = ref(true)
const error = ref('')
const justAdded = ref(0) // menu item id flashed after add

function addToCart(item) {
  if (!auth.isAuthenticated) {
    router.push({ name: 'login', query: { redirect: `/chefs/${props.id}` } })
    return
  }
  cart.add(item, chef.value)
  justAdded.value = item.id
  setTimeout(() => (justAdded.value = 0), 1200)
}

onMounted(async () => {
  try {
    chef.value = await api.get(`/chefs/${props.id}`)
    items.value = page(await api.get(`/chefs/${props.id}/menu-items?limit=100`)).items
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="space-y-6">
    <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>

    <div v-if="loading" class="space-y-4">
      <div class="skeleton h-28"></div>
      <div class="grid gap-3 sm:grid-cols-2"><div class="skeleton h-24"></div><div class="skeleton h-24"></div></div>
    </div>

    <template v-else-if="chef">
      <div class="card flex items-start gap-4">
        <div class="avatar h-14 w-14 text-2xl">{{ (chef.business_name || '?')[0].toUpperCase() }}</div>
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <h1 class="page-title">{{ chef.business_name }}</h1>
            <span v-if="chef.is_online" class="badge bg-green-100 text-green-700">● online</span>
          </div>
          <p v-if="chef.bio" class="mt-1 text-gray-600">{{ chef.bio }}</p>
          <p class="mt-1 text-sm text-gray-500">📍 {{ chef.kitchen_address }}</p>
          <p class="mt-1 text-sm">
            <span class="badge bg-amber-50 text-amber-700">★ {{ chef.rating?.toFixed(1) ?? '—' }}</span>
            <span class="ml-1 text-gray-400">{{ chef.total_reviews }} reviews</span>
          </p>
        </div>
      </div>

      <div>
        <h2 class="mb-3 text-lg font-semibold">Menu</h2>
        <div v-if="!items.length" class="empty-state">
          <span class="empty-state-emoji">🥘</span>
          <p class="font-medium text-gray-600">This chef hasn't added dishes yet</p>
          <p class="text-sm">Check back soon — good things take time to simmer.</p>
        </div>
        <div v-else class="grid gap-3 sm:grid-cols-2">
          <div v-for="item in items" :key="item.id" class="card-hover flex items-start justify-between gap-3">
            <div class="min-w-0">
              <h3 class="font-medium">{{ item.name }}</h3>
              <p v-if="item.description" class="truncate text-sm text-gray-500">{{ item.description }}</p>
              <p class="mt-1.5"><span class="badge bg-brand-50 text-brand-700">${{ item.price?.toFixed(2) }}</span></p>
            </div>
            <button
              class="btn-primary shrink-0"
              :class="justAdded === item.id && 'bg-green-600 hover:bg-green-600'"
              :disabled="!item.is_available"
              @click="addToCart(item)"
            >
              {{ justAdded === item.id ? '✓ Added' : item.is_available ? '+ Add' : 'Sold out' }}
            </button>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
