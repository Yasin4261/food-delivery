<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, page } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores/cart'
import { useFavoritesStore } from '@/stores/favorites'

const props = defineProps({ id: { type: String, required: true } })
const auth = useAuthStore()
const cart = useCartStore()
const favorites = useFavoritesStore()
const router = useRouter()

const canFavorite = auth.isAuthenticated && !auth.isChef

async function toggleFavorite() {
  try {
    await favorites.toggle(chef.value.id)
  } catch (e) {
    error.value = e.message
  }
}

// Opens (or reuses) the thread with this chef and jumps into it.
async function startChat() {
  if (!auth.isAuthenticated) {
    router.push({ name: 'login', query: { redirect: `/chefs/${props.id}` } })
    return
  }
  try {
    const conv = await api.post('/chat/conversations', { chef_id: chef.value.id })
    router.push({ path: '/chat', query: { c: conv.id } })
  } catch (e) {
    error.value = e.message
  }
}

const chef = ref(null)
const items = ref([])
const reviews = ref([])
const reviewsTotal = ref(0)
const loading = ref(true)
const error = ref('')
const justAdded = ref(0) // menu item id flashed after add

const stars = (n) => '★'.repeat(n) + '☆'.repeat(5 - n)
const when = (iso) => new Date(iso).toLocaleDateString()

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
    const rp = page(await api.get(`/chefs/${props.id}/reviews?limit=20`))
    reviews.value = rp.items
    reviewsTotal.value = rp.total
    if (canFavorite) await favorites.load().catch(() => {})
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
        <img v-if="chef.image_url" :src="chef.image_url" :alt="chef.business_name" class="h-14 w-14 rounded-xl object-cover" />
        <div v-else class="avatar h-14 w-14 text-2xl">{{ (chef.business_name || '?')[0].toUpperCase() }}</div>
        <div class="min-w-0 grow">
          <div class="flex flex-wrap items-center gap-2">
            <h1 class="page-title">{{ chef.business_name }}</h1>
            <span v-if="chef.is_online" class="badge bg-green-100 text-green-700">{{ $t('chef.online') }}</span>
            <button
              v-if="canFavorite"
              class="ml-auto text-2xl transition hover:scale-110"
              :title="favorites.has(chef.id) ? $t('favorites.removeTitle') : $t('favorites.addTitle')"
              @click="toggleFavorite"
            >
              {{ favorites.has(chef.id) ? '❤️' : '🤍' }}
            </button>
          </div>
          <p v-if="chef.bio" class="mt-1 text-gray-600">{{ chef.bio }}</p>
          <p class="mt-1 text-sm text-gray-500">📍 {{ chef.kitchen_address }}</p>
          <p class="mt-1 text-sm">
            <span class="badge bg-amber-50 text-amber-700">★ {{ chef.rating?.toFixed(1) ?? '—' }}</span>
            <span class="ml-1 text-gray-400">{{ chef.total_reviews }} · {{ $t('chef.reviews').toLowerCase() }}</span>
          </p>
          <button v-if="!auth.isChef" class="btn-ghost mt-3" @click="startChat">{{ $t('chef.chatWithChef') }}</button>
        </div>
      </div>

      <div>
        <h2 class="mb-3 text-lg font-semibold">{{ $t('chef.menu') }}</h2>
        <div v-if="!items.length" class="empty-state">
          <span class="empty-state-emoji">🥘</span>
          <p class="font-medium text-gray-600">{{ $t('chef.noDishes') }}</p>
          <p class="text-sm">{{ $t('chef.noDishesHint') }}</p>
        </div>
        <div v-else class="grid gap-3 sm:grid-cols-2">
          <div v-for="item in items" :key="item.id" class="card-hover flex items-start justify-between gap-3">
            <img v-if="item.image_url" :src="item.image_url" :alt="item.name" class="h-16 w-16 shrink-0 rounded-lg object-cover" />
            <div class="min-w-0 grow">
              <h3 class="font-medium">{{ item.name }}</h3>
              <p v-if="item.description" class="truncate text-sm text-gray-500">{{ item.description }}</p>
              <p class="mt-1.5">
                <span class="badge bg-brand-50 text-brand-700">${{ item.price?.toFixed(2) }}</span>
                <span v-if="item.total_reviews" class="badge ml-1 bg-amber-50 text-amber-700">
                  ★ {{ item.rating?.toFixed(1) }}
                </span>
              </p>
            </div>
            <button
              class="btn-primary shrink-0"
              :class="justAdded === item.id && 'bg-green-600 hover:bg-green-600'"
              :disabled="!item.is_available"
              @click="addToCart(item)"
            >
              {{ justAdded === item.id ? $t('chef.added') : item.is_available ? $t('chef.add') : $t('chef.soldOut') }}
            </button>
          </div>
        </div>
      </div>

      <div>
        <h2 class="mb-3 text-lg font-semibold">{{ $t('chef.reviews') }} <span class="text-sm font-normal text-gray-400">({{ reviewsTotal }})</span></h2>
        <div v-if="!reviews.length" class="empty-state">
          <span class="empty-state-emoji">⭐</span>
          <p class="font-medium text-gray-600">{{ $t('chef.noReviews') }}</p>
          <p class="text-sm">{{ $t('chef.noReviewsHint') }}</p>
        </div>
        <div v-else class="space-y-3">
          <div v-for="r in reviews" :key="r.id" class="card py-3">
            <div class="flex items-center justify-between">
              <span class="text-amber-400">{{ stars(r.rating) }}</span>
              <span class="text-xs text-gray-400">{{ when(r.created_at) }}</span>
            </div>
            <p v-if="r.comment" class="mt-1 text-sm text-gray-600">{{ r.comment }}</p>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
