<script setup>
import { onMounted, ref } from 'vue'
import { api, page } from '@/api/client'
import { statusClass } from '@/lib/status'

const orders = ref([])
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  try {
    orders.value = page(await api.get('/orders?limit=50')).items
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function cancel(order) {
  try {
    await api.post(`/orders/${order.id}/cancel`)
    await load()
  } catch (e) {
    error.value = e.message
  }
}

const cancellable = (s) => s === 'pending' || s === 'confirmed'

onMounted(load)
</script>

<template>
  <div class="space-y-4">
    <h1 class="text-2xl font-bold">My orders</h1>
    <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
    <p v-if="loading" class="text-gray-500">Loading…</p>
    <p v-else-if="!orders.length" class="text-gray-500">You have no orders yet.</p>

    <div v-for="order in orders" :key="order.id" class="card space-y-2">
      <div class="flex items-center justify-between">
        <div>
          <span class="font-mono text-sm text-gray-500">{{ order.order_code }}</span>
          <span class="badge ml-2" :class="statusClass(order.status)">{{ order.status }}</span>
        </div>
        <span class="font-semibold">${{ order.total_price?.toFixed(2) }}</span>
      </div>
      <ul class="text-sm text-gray-600">
        <li v-for="it in order.items" :key="it.id">{{ it.quantity }}× {{ it.item_name }}</li>
      </ul>
      <div class="flex justify-end">
        <button v-if="cancellable(order.status)" class="btn-ghost" @click="cancel(order)">Cancel</button>
      </div>
    </div>
  </div>
</template>
