<script setup>
import { onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { api, page } from '@/api/client'
import { statusClass } from '@/lib/status'
import OrderReviewPanel from '@/components/OrderReviewPanel.vue'

const route = useRoute()
const router = useRouter()

// Which orders have their review panel open, keyed by order id.
const reviewing = ref({})

// Payment-result banner from the gateway redirect (?payment=success|failed|error).
const paymentBanner = ref('')
const paying = ref(0)

const PAYMENT_BADGES = {
  paid: 'bg-green-100 text-green-700',
  pending: 'bg-amber-100 text-amber-700',
  refunded: 'bg-blue-100 text-blue-700',
  failed: 'bg-red-100 text-red-700',
}
const paymentClass = (s) => PAYMENT_BADGES[s] || 'bg-gray-100 text-gray-600'

const payable = (o) => o.payment_method === 'card' && o.payment_status === 'pending' && o.status !== 'cancelled'

async function payNow(order) {
  paying.value = order.id
  try {
    const { payment_page_url } = await api.post(`/orders/${order.id}/pay`)
    window.location.href = payment_page_url
  } catch (e) {
    error.value = e.message
    paying.value = 0
  }
}

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

onMounted(() => {
  const outcome = route.query.payment
  if (outcome === 'success') paymentBanner.value = '🎉 Payment received — thank you!'
  else if (outcome === 'failed') paymentBanner.value = '❌ Payment failed — you can retry from the order below.'
  else if (outcome === 'error') paymentBanner.value = '⚠️ We could not verify that payment. If you were charged, contact support.'
  if (outcome) router.replace({ query: {} })
  load()
})
</script>

<template>
  <div class="space-y-4">
    <div>
      <h1 class="page-title">My orders</h1>
      <p class="page-subtitle">Track your deliveries and past orders.</p>
    </div>
    <div
      v-if="paymentBanner"
      class="flex items-center justify-between rounded-lg border border-brand-200 bg-brand-50 px-3 py-2 text-sm"
    >
      <span>{{ paymentBanner }}</span>
      <button class="text-gray-400 hover:text-gray-600" @click="paymentBanner = ''">✕</button>
    </div>
    <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>
    <div v-if="loading" class="space-y-3"><div class="skeleton h-24"></div><div class="skeleton h-24"></div></div>
    <div v-else-if="!orders.length" class="empty-state">
      <span class="empty-state-emoji">🍽️</span>
      <p class="font-medium text-gray-600">No orders yet</p>
      <p class="text-sm">Hungry? <RouterLink to="/" class="text-brand-600 hover:underline">Browse chefs</RouterLink> and place your first order.</p>
    </div>

    <div v-for="order in orders" :key="order.id" class="card space-y-2">
      <div class="flex items-center justify-between">
        <div>
          <span class="font-mono text-sm text-gray-500">{{ order.order_code }}</span>
          <span class="badge ml-2" :class="statusClass(order.status)">{{ order.status }}</span>
          <span class="badge ml-1" :class="paymentClass(order.payment_status)">
            {{ order.payment_method === 'card' ? '💳' : '💵' }} {{ order.payment_status }}
          </span>
        </div>
        <span class="font-semibold">${{ order.total_price?.toFixed(2) }}</span>
      </div>
      <ul class="text-sm text-gray-600">
        <li v-for="it in order.items" :key="it.id">{{ it.quantity }}× {{ it.item_name }}</li>
      </ul>
      <div class="flex justify-end gap-2">
        <button v-if="payable(order)" class="btn-primary" :disabled="paying === order.id" @click="payNow(order)">
          {{ paying === order.id ? 'Redirecting…' : '💳 Pay now' }}
        </button>
        <button
          v-if="order.status === 'delivered'"
          class="btn-ghost"
          @click="reviewing[order.id] = !reviewing[order.id]"
        >
          {{ reviewing[order.id] ? 'Hide rating' : '⭐ Rate order' }}
        </button>
        <button v-if="cancellable(order.status)" class="btn-ghost" @click="cancel(order)">Cancel</button>
      </div>
      <OrderReviewPanel v-if="reviewing[order.id]" :order="order" />
    </div>
  </div>
</template>
