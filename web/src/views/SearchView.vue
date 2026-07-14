<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { api, page } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores/cart'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const cart = useCartStore()

const q = ref(typeof route.query.q === 'string' ? route.query.q : '')
const type = ref(route.query.type === 'food' ? 'food' : 'chef')
const results = ref([])
const total = ref(0)
const loading = ref(false)
const searched = ref(false)
const error = ref('')
const justAdded = ref(0)

// Filters/sort: whitelisted values, validated again server-side.
const sort = ref('')
const minRating = ref('')
const minPrice = ref('')
const maxPrice = ref('')

function filterParams() {
  const p = new URLSearchParams({ q: q.value.trim(), type: type.value, limit: '30' })
  if (sort.value) p.set('sort', sort.value)
  if (minRating.value) p.set('min_rating', minRating.value)
  if (type.value === 'food') {
    if (minPrice.value) p.set('min_price', minPrice.value)
    if (maxPrice.value) p.set('max_price', maxPrice.value)
  }
  return p
}

async function search() {
  const term = q.value.trim()
  if (!term) return
  loading.value = true
  error.value = ''
  // Keep the query shareable/bookmarkable.
  router.replace({ query: { q: term, type: type.value } })
  try {
    const p = page(await api.get(`/search?${filterParams()}`))
    results.value = p.items
    total.value = p.total
    searched.value = true
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function setType(t) {
  if (type.value === t) return
  type.value = t
  results.value = []
  if (sort.value.startsWith('price') && t === 'chef') sort.value = '' // chefs have no price sort
  if (q.value.trim()) search()
}

function addToCart(item) {
  cart.add(item)
  justAdded.value = item.id
  setTimeout(() => (justAdded.value = 0), 1200)
}

onMounted(() => {
  if (q.value.trim()) search()
})
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="page-title">{{ $t('search.title') }}</h1>
      <p class="page-subtitle">{{ $t('search.subtitle') }}</p>
    </div>

    <form class="flex flex-wrap items-center gap-2" @submit.prevent="search">
      <input v-model="q" class="input max-w-md" :placeholder="$t('search.placeholder')" autofocus />
      <button class="btn-primary" :disabled="loading || !q.trim()">{{ loading ? $t('search.searching') : $t('search.search') }}</button>
      <div class="ml-2 flex rounded-lg border border-gray-300 bg-white p-0.5">
        <button
          type="button"
          class="rounded-md px-3 py-1 text-sm font-medium transition"
          :class="type === 'chef' ? 'bg-brand-600 text-white shadow-sm' : 'text-gray-600 hover:bg-gray-50'"
          @click="setType('chef')"
        >
          {{ $t('search.chefs') }}
        </button>
        <button
          type="button"
          class="rounded-md px-3 py-1 text-sm font-medium transition"
          :class="type === 'food' ? 'bg-brand-600 text-white shadow-sm' : 'text-gray-600 hover:bg-gray-50'"
          @click="setType('food')"
        >
          {{ $t('search.dishes') }}
        </button>
      </div>
    </form>

    <!-- Filters + sort -->
    <div class="flex flex-wrap items-center gap-2 text-sm">
      <select v-model="sort" class="input w-auto py-1.5" @change="search">
        <option value="">{{ $t('filters.sortDefault') }}</option>
        <option value="rating">{{ $t('filters.sortRating') }}</option>
        <option value="popular">{{ $t('filters.sortPopular') }}</option>
        <template v-if="type === 'food'">
          <option value="price_asc">{{ $t('filters.sortPriceAsc') }}</option>
          <option value="price_desc">{{ $t('filters.sortPriceDesc') }}</option>
        </template>
      </select>
      <select v-model="minRating" class="input w-auto py-1.5" @change="search">
        <option value="">{{ $t('filters.anyRating') }}</option>
        <option value="3">★ 3+</option>
        <option value="4">★ 4+</option>
        <option value="4.5">★ 4.5+</option>
      </select>
      <template v-if="type === 'food'">
        <input v-model="minPrice" class="input w-24 py-1.5" type="number" min="0" step="0.5" :placeholder="$t('filters.minPrice')" @change="search" />
        <input v-model="maxPrice" class="input w-24 py-1.5" type="number" min="0" step="0.5" :placeholder="$t('filters.maxPrice')" @change="search" />
      </template>
    </div>

    <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>

    <div v-if="loading" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <div v-for="n in 3" :key="n" class="skeleton h-32"></div>
    </div>

    <div v-else-if="searched && !results.length" class="empty-state">
      <span class="empty-state-emoji">🫥</span>
      <p class="font-medium text-gray-600">{{ $t('search.empty', { q }) }}</p>
      <p class="text-sm">{{ $t('search.emptyHint') }}</p>
    </div>

    <template v-else-if="searched">
      <p class="text-sm text-gray-500">{{ $t('search.results', { n: total }, total) }}</p>

      <!-- Chef results -->
      <div v-if="type === 'chef'" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <RouterLink v-for="chef in results" :key="chef.id" :to="`/chefs/${chef.id}`" class="card-hover">
          <div class="flex items-start gap-3">
            <div class="avatar relative">
              {{ (chef.business_name || '?')[0].toUpperCase() }}
              <span
                v-if="chef.is_online"
                class="absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-white bg-green-500"
              ></span>
            </div>
            <div class="min-w-0">
              <h2 class="truncate font-semibold">{{ chef.business_name }}</h2>
              <p v-if="chef.specialty" class="truncate text-sm text-gray-500">{{ chef.specialty }}</p>
            </div>
          </div>
          <div class="mt-3 flex items-center justify-between text-sm">
            <span class="truncate text-gray-500">📍 {{ chef.kitchen_city || chef.kitchen_address }}</span>
            <span class="badge bg-amber-50 text-amber-700">★ {{ chef.rating?.toFixed(1) ?? '—' }}</span>
          </div>
        </RouterLink>
      </div>

      <!-- Dish results -->
      <div v-else class="grid gap-3 sm:grid-cols-2">
        <div v-for="item in results" :key="item.id" class="card-hover flex items-start justify-between gap-3">
          <div class="min-w-0">
            <h3 class="font-medium">{{ item.name }}</h3>
            <p v-if="item.description" class="truncate text-sm text-gray-500">{{ item.description }}</p>
            <p class="mt-1.5 flex items-center gap-1.5">
              <span class="badge bg-brand-50 text-brand-700">${{ item.price?.toFixed(2) }}</span>
              <RouterLink :to="`/chefs/${item.chef_id}`" class="text-xs text-brand-600 hover:underline">
                {{ $t('search.viewChef') }}
              </RouterLink>
            </p>
          </div>
          <button
            v-if="!auth.isChef"
            class="btn-primary shrink-0"
            :class="justAdded === item.id && 'bg-green-600 hover:bg-green-600'"
            :disabled="!item.is_available"
            @click="addToCart(item)"
          >
            {{ justAdded === item.id ? $t('chef.added') : item.is_available ? $t('chef.add') : $t('chef.soldOut') }}
          </button>
        </div>
      </div>
    </template>
  </div>
</template>
