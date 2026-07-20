<script setup>
import { onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { formatMoney as money } from '@/lib/money'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { api, page } from '@/api/client'
import { statusClass } from '@/lib/status'
import { POLL_MS } from '@/stores/notifications'
import { useCartStore } from '@/stores/cart'
import { reorderIntoCart } from '@/lib/reorder'
import OrderReviewPanel from '@/components/OrderReviewPanel.vue'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

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

// Per-order "save this card" opt-in for the checkout (#67).
const saveCard = reactive({})

async function payNow(order) {
  paying.value = order.id
  try {
    const { payment_page_url } = await api.post(`/orders/${order.id}/pay`, {
      save_card: !!saveCard[order.id],
    })
    window.location.href = payment_page_url
  } catch (e) {
    error.value = e.message
    paying.value = 0
  }
}

const orders = ref([])
const loading = ref(true)
const error = ref('')
let poll = null

async function load(silent = false) {
  if (!silent) loading.value = true
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

// ETA line: shown only for in-progress orders that have an estimate.
function etaLabel(order) {
  if (!order.estimated_delivery_time) return ''
  if (order.status === 'delivered' || order.status === 'cancelled') return ''
  const eta = new Date(order.estimated_delivery_time)
  const clock = eta.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  const mins = Math.round((eta - Date.now()) / 60000)
  return mins > 0
    ? t('orders.etaIn', { clock, mins })
    : t('orders.etaSoon', { clock })
}

// "Order again": repopulate the cart from a past order using CURRENT dish
// prices/availability, dropping anything no longer orderable.
const cart = useCartStore()
const reordering = ref(0)
async function orderAgain(order) {
  reordering.value = order.id
  error.value = ''
  try {
    const { added, dropped } = await reorderIntoCart(order, cart, api)
    if (!added) {
      error.value = t('orders.reorderNothing')
      return
    }
    if (dropped.length) error.value = t('orders.reorderDropped', { items: dropped.join(', ') })
    router.push({ name: 'cart' })
  } catch (e) {
    error.value = e.message
  } finally {
    reordering.value = 0
  }
}

onMounted(() => {
  const outcome = route.query.payment
  if (outcome === 'success') paymentBanner.value = t('orders.paySuccess')
  else if (outcome === 'failed') paymentBanner.value = t('orders.payFailed')
  else if (outcome === 'error') paymentBanner.value = t('orders.payError')
  if (outcome) router.replace({ query: {} })
  load()
  // Status changes appear without a manual reload (issue #55).
  poll = setInterval(() => {
    if (!document.hidden) load(true)
  }, POLL_MS)
})

onBeforeUnmount(() => clearInterval(poll))
</script>

<template>
  <div class="space-y-4">
    <div>
      <h1 class="page-title">{{ $t('orders.title') }}</h1>
      <p class="page-subtitle">{{ $t('orders.subtitle') }}</p>
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
      <p class="font-medium text-gray-600">{{ $t('orders.empty') }}</p>
      <p class="text-sm"><i18n-t keypath="orders.emptyHint" tag="span"><template #browse><RouterLink to="/" class="text-brand-600 hover:underline">{{ $t('orders.browseChefs') }}</RouterLink></template></i18n-t></p>
    </div>

    <div v-for="order in orders" :key="order.id" class="card space-y-2">
      <div class="flex items-center justify-between">
        <div>
          <span class="font-mono text-sm text-gray-500">{{ order.order_code }}</span>
          <span class="badge ml-2" :class="statusClass(order.status)">{{ $t(`status.${order.status}`) }}</span>
          <span class="badge ml-1" :class="paymentClass(order.payment_status)">
            {{ order.payment_method === 'card' ? '💳' : '💵' }} {{ $t(`payment.${order.payment_status}`) }}
          </span>
        </div>
        <span class="text-right">
          <span class="font-semibold">{{ money(order.total_price) }}</span>
          <span v-if="order.delivery_fee > 0" class="block text-xs text-gray-400">
            {{ $t('orders.inclDelivery', { fee: money(order.delivery_fee) }) }}
          </span>
          <span v-if="order.tip > 0" class="block text-xs text-gray-400">
            {{ $t('orders.inclTip', { tip: money(order.tip) }) }}
          </span>
          <span v-if="order.discount > 0" class="block text-xs text-green-600">
            {{ $t('orders.discount', { amount: money(order.discount), code: order.promo_code }) }}
          </span>
        </span>
      </div>
      <!-- Estimated delivery time, while the order is in progress. -->
      <p v-if="etaLabel(order)" class="text-sm text-brand-700">🕒 {{ etaLabel(order) }}</p>

      <!-- Multi-chef orders: each chef's slice advances on its own. -->
      <div v-if="order.sub_orders?.length > 1" class="flex flex-wrap gap-2 text-sm">
        <span
          v-for="sub in order.sub_orders"
          :key="sub.id"
          class="inline-flex items-center gap-1.5 rounded-full bg-gray-50 py-0.5 pl-2.5 pr-1 text-gray-600"
        >
          {{ sub.chef_name }}
          <span class="badge" :class="statusClass(sub.status)">{{ $t(`status.${sub.status}`) }}</span>
        </span>
      </div>
      <ul class="text-sm text-gray-600">
        <li v-for="it in order.items" :key="it.id">{{ it.quantity }}× {{ it.item_name }}</li>
      </ul>
      <div class="flex flex-wrap items-center justify-end gap-2">
        <label v-if="payable(order)" class="mr-auto flex items-center gap-2 text-sm text-gray-500">
          <input v-model="saveCard[order.id]" type="checkbox" class="rounded border-gray-300" />
          {{ $t('cards.saveThisCard') }}
        </label>
        <button v-if="payable(order)" class="btn-primary" :disabled="paying === order.id" @click="payNow(order)">
          {{ paying === order.id ? $t('orders.redirecting') : $t('orders.payNow') }}
        </button>
        <button
          v-if="order.status === 'delivered' || order.sub_orders?.some((s) => s.status === 'delivered')"
          class="btn-ghost"
          @click="reviewing[order.id] = !reviewing[order.id]"
        >
          {{ reviewing[order.id] ? $t('orders.hideRating') : $t('orders.rate') }}
        </button>
        <button v-if="cancellable(order.status)" class="btn-ghost" @click="cancel(order)">{{ $t('orders.cancel') }}</button>
        <button class="btn-ghost" :disabled="reordering === order.id" @click="orderAgain(order)">
          {{ reordering === order.id ? $t('orders.reordering') : $t('orders.orderAgain') }}
        </button>
      </div>
      <OrderReviewPanel v-if="reviewing[order.id]" :order="order" />
    </div>
  </div>
</template>
