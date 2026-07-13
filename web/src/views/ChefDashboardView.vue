<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { api, page, ApiError } from '@/api/client'
import { statusClass, nextAction, canDecline } from '@/lib/status'
import { POLL_MS } from '@/stores/notifications'
import ChefOnboarding from '@/components/ChefOnboarding.vue'

const profile = ref(null)
const needsProfile = ref(false)
const orders = ref([])
const earnings = ref(null)
const loading = ref(true)
const error = ref('')
const toggling = ref(false)

async function load(silent = false) {
  if (!silent) loading.value = true
  error.value = ''
  try {
    // No profile yet -> onboarding instead of a broken dashboard.
    try {
      profile.value = await api.get('/chefs/me')
      needsProfile.value = false
    } catch (e) {
      if (e instanceof ApiError && e.status === 404) {
        needsProfile.value = true
        return
      }
      throw e
    }
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

async function toggleOnline() {
  toggling.value = true
  error.value = ''
  try {
    profile.value = await api.patch('/chefs/me/status', { is_online: !profile.value.is_online })
  } catch (e) {
    error.value = e.message
  } finally {
    toggling.value = false
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

// The chef acts on their own slice of the order: badge and action buttons are
// driven by the caller's sub-order status, not the order-level (derived) one.
function myStatus(order) {
  const sub = order.sub_orders?.find((s) => s.chef_id === profile.value?.id)
  return sub ? sub.status : order.status
}

// Incoming orders refresh themselves — no manual reload needed (issue #55).
let poll = null
onMounted(() => {
  load()
  poll = setInterval(() => {
    if (!document.hidden && !needsProfile.value) load(true)
  }, POLL_MS)
})
onBeforeUnmount(() => clearInterval(poll))
</script>

<template>
  <p v-if="loading" class="text-gray-500">Loading…</p>

  <ChefOnboarding v-else-if="needsProfile" @created="load" />

  <div v-else class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold">{{ profile?.business_name }}</h1>
        <p class="text-sm text-gray-500">{{ $t('dashboard.subtitle') }}</p>
      </div>
      <div class="flex items-center gap-2">
        <span class="badge" :class="profile?.is_online ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'">
          {{ profile?.is_online ? $t('dashboard.online') : $t('dashboard.offline') }}
        </span>
        <button class="btn-ghost" :disabled="toggling" @click="toggleOnline">
          {{ profile?.is_online ? $t('dashboard.goOffline') : $t('dashboard.goOnline') }}
        </button>
        <RouterLink to="/chef/menus" class="btn-primary">{{ $t('nav.myMenus') }}</RouterLink>
        <button class="btn-ghost" @click="load">{{ $t('dashboard.refresh') }}</button>
      </div>
    </div>

    <div v-if="earnings" class="grid gap-4 sm:grid-cols-3">
      <div class="card">
        <p class="text-sm text-gray-500">{{ $t('dashboard.earnings') }}</p>
        <p class="text-2xl font-bold">${{ earnings.total_earnings?.toFixed(2) }}</p>
      </div>
      <div class="card">
        <p class="text-sm text-gray-500">{{ $t('dashboard.deliveredOrders') }}</p>
        <p class="text-2xl font-bold">{{ earnings.delivered_orders }}</p>
      </div>
      <div class="card">
        <p class="text-sm text-gray-500">{{ $t('dashboard.itemsSold') }}</p>
        <p class="text-2xl font-bold">{{ earnings.items_sold }}</p>
      </div>
    </div>

    <h2 class="text-lg font-semibold">{{ $t('dashboard.incoming') }}</h2>
    <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>
    <div v-if="!orders.length" class="empty-state">
      <span class="empty-state-emoji">👨‍🍳</span>
      <p class="font-medium text-gray-600">{{ $t('dashboard.empty') }}</p>
      <p class="text-sm">{{ $t('dashboard.emptyHint') }}</p>
    </div>

    <div v-for="order in orders" :key="order.id" class="card space-y-2">
      <div class="flex items-center justify-between">
        <div>
          <span class="font-mono text-sm text-gray-500">{{ order.order_code }}</span>
          <span class="badge ml-2" :class="statusClass(myStatus(order))">{{ $t(`status.${myStatus(order)}`) }}</span>
          <span v-if="order.sub_orders?.length > 1" class="ml-2 text-xs text-gray-400">{{ $t('dashboard.multiChef') }}</span>
        </div>
        <span class="font-semibold">${{ order.total_price?.toFixed(2) }}</span>
      </div>
      <ul class="text-sm text-gray-600">
        <li v-for="it in order.items" :key="it.id">{{ it.quantity }}× {{ it.item_name }}</li>
      </ul>
      <div class="flex justify-end gap-2">
        <button v-if="canDecline(myStatus(order))" class="btn-ghost" @click="advance(order, 'decline')">{{ $t('actions.decline') }}</button>
        <button
          v-if="nextAction(myStatus(order))"
          class="btn-primary"
          @click="advance(order, nextAction(myStatus(order)).action)"
        >
          {{ $t(nextAction(myStatus(order)).labelKey) }}
        </button>
      </div>
    </div>
  </div>
</template>
