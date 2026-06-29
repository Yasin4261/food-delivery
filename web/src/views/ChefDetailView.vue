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

function addToCart(item) {
  if (!auth.isAuthenticated) {
    router.push({ name: 'login', query: { redirect: `/chefs/${props.id}` } })
    return
  }
  cart.add(item, chef.value)
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
    <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
    <p v-if="loading" class="text-gray-500">Loading…</p>

    <template v-else-if="chef">
      <div>
        <h1 class="text-2xl font-bold">{{ chef.business_name }}</h1>
        <p class="text-gray-600">{{ chef.bio }}</p>
        <p class="mt-1 text-sm text-gray-500">{{ chef.kitchen_address }}</p>
        <p class="mt-1 text-sm text-amber-600">★ {{ chef.rating?.toFixed(1) ?? '—' }} ({{ chef.total_reviews }} reviews)</p>
      </div>

      <div>
        <h2 class="mb-3 text-lg font-semibold">Menu</h2>
        <p v-if="!items.length" class="text-gray-500">This chef has no dishes yet.</p>
        <div v-else class="grid gap-3 sm:grid-cols-2">
          <div v-for="item in items" :key="item.id" class="card flex items-start justify-between gap-3">
            <div>
              <h3 class="font-medium">{{ item.name }}</h3>
              <p v-if="item.description" class="text-sm text-gray-500">{{ item.description }}</p>
              <p class="mt-1 text-sm font-semibold">${{ item.price?.toFixed(2) }}</p>
            </div>
            <button class="btn-primary shrink-0" :disabled="!item.is_available" @click="addToCart(item)">
              {{ item.is_available ? 'Add' : 'Sold out' }}
            </button>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
