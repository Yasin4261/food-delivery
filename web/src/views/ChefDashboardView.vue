<script setup>
import { onMounted, ref } from 'vue'
import { api, page } from '@/api/client'
import { statusClass, nextAction, canDecline } from '@/lib/status'

const orders = ref([])
const earnings = ref(null)
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [o, e] = await Promise.all([
      api.get('/chef/orders?limit=50'),
      api.get('/chefs/me/earnings'),
    ])
    orders.value = page(o).items
    earnings.value = e
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function advance(order, action) {
  try {
    await api.post(`/chef/orders/${order.id}/status`, { action })
    await load()
  } catch (e) {
    error.value = e.message
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">Chef dashboard</h1>
      <button class="btn-ghost" @click="load">Refresh</button>
    </div>

    <div v-if="earnings" class="grid gap-4 sm:grid-cols-3">
      <div class="card">
        <p class="text-sm text-gray-500">Earnings (delivered &amp; paid)</p>
        <p class="text-2xl font-bold">${{ earnings.total_earnings?.toFixed(2) }}</p>
      </div>
      <div class="card">
        <p class="text-sm text-gray-500">Delivered orders</p>
        <p class="text-2xl font-bold">{{ earnings.delivered_orders }}</p>
      </div>
      <div class="card">
        <p class="text-sm text-gray-500">Items sold</p>
        <p class="text-2xl font-bold">{{ earnings.items_sold }}</p>
      </div>
    </div>

    <h2 class="text-lg font-semibold">Incoming orders</h2>
    <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
    <p v-if="loading" class="text-gray-500">Loading…</p>
    <p v-else-if="!orders.length" class="text-gray-500">No orders yet.</p>

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
      <div class="flex justify-end gap-2">
        <button
          v-if="canDecline(order.status)"
          class="btn-ghost"
          @click="advance(order, 'decline')"
        >
          Decline
        </button>
        <button
          v-if="nextAction(order.status)"
          class="btn-primary"
          @click="advance(order, nextAction(order.status).action)"
        >
          {{ nextAction(order.status).label }}
        </button>
      </div>
    </div>
  </div>
</template>
