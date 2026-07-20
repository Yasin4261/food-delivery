<script setup>
import { onMounted, reactive, ref } from 'vue'
import { formatMoney as money } from '@/lib/money'
import { api, page } from '@/api/client'
import { statusClass } from '@/lib/status'

const tab = ref('overview')
const tabs = ['overview', 'users', 'chefs', 'orders', 'promos']

const stats = ref(null)
const users = ref([])
const chefs = ref([])
const orders = ref([])
const promos = ref([])
const newPromo = reactive({ code: '', discount_type: 'percent', discount_value: '', min_order: '', usage_limit: '' })
const error = ref('')

async function loadStats() {
  try {
    stats.value = await api.get('/admin/stats')
  } catch (e) {
    error.value = e.message
  }
}
async function loadUsers() {
  try {
    users.value = page(await api.get('/admin/users?limit=100')).items
  } catch (e) {
    error.value = e.message
  }
}
async function loadChefs() {
  try {
    chefs.value = page(await api.get('/admin/chefs?limit=100')).items
  } catch (e) {
    error.value = e.message
  }
}
async function loadOrders() {
  try {
    orders.value = page(await api.get('/admin/orders?limit=100')).items
  } catch (e) {
    error.value = e.message
  }
}

async function loadPromos() {
  try {
    promos.value = page(await api.get('/admin/promos?limit=100')).items
  } catch (e) {
    error.value = e.message
  }
}
async function createPromo() {
  error.value = ''
  try {
    const p = {
      code: newPromo.code,
      discount_type: newPromo.discount_type,
      discount_value: Number(newPromo.discount_value),
      min_order: Number(newPromo.min_order) || 0,
      usage_limit: Number(newPromo.usage_limit) || 0,
    }
    await api.post('/admin/promos', p)
    newPromo.code = newPromo.discount_value = newPromo.min_order = newPromo.usage_limit = ''
    await loadPromos()
  } catch (e) {
    error.value = e.message
  }
}
async function togglePromo(p) {
  error.value = ''
  try {
    const r = await api.patch(`/admin/promos/${p.id}/active`, { active: !p.is_active })
    p.is_active = r.active
  } catch (e) {
    error.value = e.message
  }
}

function select(t) {
  tab.value = t
  error.value = ''
  if (t === 'users' && !users.value.length) loadUsers()
  if (t === 'chefs' && !chefs.value.length) loadChefs()
  if (t === 'orders' && !orders.value.length) loadOrders()
  if (t === 'promos' && !promos.value.length) loadPromos()
}

async function toggleUser(u) {
  error.value = ''
  try {
    const r = await api.patch(`/admin/users/${u.id}/active`, { active: !u.is_active })
    u.is_active = r.active
  } catch (e) {
    error.value = e.message
  }
}
async function toggleChef(c) {
  error.value = ''
  try {
    const r = await api.patch(`/admin/chefs/${c.id}/active`, { active: !c.is_active })
    c.is_active = r.active
  } catch (e) {
    error.value = e.message
  }
}

onMounted(loadStats)
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="page-title">{{ $t('admin.title') }}</h1>
      <p class="page-subtitle">{{ $t('admin.subtitle') }}</p>
    </div>

    <div class="flex flex-wrap gap-1 border-b border-gray-200">
      <button
        v-for="t in tabs"
        :key="t"
        class="rounded-t-md px-4 py-2 text-sm font-medium transition"
        :class="tab === t ? 'border-b-2 border-brand-600 text-brand-700' : 'text-gray-500 hover:text-gray-700'"
        @click="select(t)"
      >
        {{ $t(`admin.${t}`) }}
      </button>
    </div>

    <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>

    <!-- Overview -->
    <div v-if="tab === 'overview' && stats" class="space-y-6">
      <div class="grid gap-4 sm:grid-cols-3 lg:grid-cols-4">
        <div class="card"><p class="text-sm text-gray-500">{{ $t('admin.gmv') }}</p><p class="text-2xl font-bold">{{ money(stats.gmv) }}</p></div>
        <div class="card"><p class="text-sm text-gray-500">{{ $t('admin.totalOrders') }}</p><p class="text-2xl font-bold">{{ stats.total_orders }}</p></div>
        <div class="card"><p class="text-sm text-gray-500">{{ $t('admin.deliveredOrders') }}</p><p class="text-2xl font-bold">{{ stats.delivered_orders }}</p></div>
        <div class="card"><p class="text-sm text-gray-500">{{ $t('admin.ordersToday') }}</p><p class="text-2xl font-bold">{{ stats.orders_today }}</p></div>
        <div class="card"><p class="text-sm text-gray-500">{{ $t('admin.totalUsers') }}</p><p class="text-2xl font-bold">{{ stats.total_users }}</p></div>
        <div class="card"><p class="text-sm text-gray-500">{{ $t('admin.totalChefs') }}</p><p class="text-2xl font-bold">{{ stats.total_chefs }}</p></div>
        <div class="card"><p class="text-sm text-gray-500">{{ $t('admin.activeChefs') }}</p><p class="text-2xl font-bold">{{ stats.active_chefs }}</p></div>
      </div>
      <div v-if="stats.top_chefs?.length" class="card">
        <h2 class="mb-2 font-semibold">{{ $t('admin.topChefs') }}</h2>
        <div v-for="c in stats.top_chefs" :key="c.chef_id" class="flex justify-between border-t border-gray-100 py-1.5 text-sm">
          <span>{{ c.business_name }}</span>
          <span class="text-gray-500">{{ c.orders }} {{ $t('admin.orders') }} · <span class="font-medium text-gray-700">{{ money(c.revenue) }}</span></span>
        </div>
      </div>
    </div>

    <!-- Users -->
    <div v-else-if="tab === 'users'" class="card overflow-x-auto">
      <table class="w-full text-sm">
        <thead class="text-left text-gray-500"><tr><th class="py-1">{{ $t('admin.user') }}</th><th>{{ $t('admin.role') }}</th><th>{{ $t('admin.status') }}</th><th></th></tr></thead>
        <tbody>
          <tr v-for="u in users" :key="u.id" class="border-t border-gray-100">
            <td class="py-1.5"><span class="font-medium">{{ u.username }}</span> <span class="text-gray-400">{{ u.email }}</span></td>
            <td><span class="badge bg-gray-100 text-gray-600">{{ u.role }}</span></td>
            <td><span class="badge" :class="u.is_active ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'">{{ u.is_active ? $t('admin.active') : $t('admin.inactive') }}</span></td>
            <td class="text-right"><button class="text-sm hover:underline" :class="u.is_active ? 'text-red-600' : 'text-green-600'" @click="toggleUser(u)">{{ u.is_active ? $t('admin.deactivate') : $t('admin.activate') }}</button></td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Chefs -->
    <div v-else-if="tab === 'chefs'" class="card overflow-x-auto">
      <table class="w-full text-sm">
        <thead class="text-left text-gray-500"><tr><th class="py-1">{{ $t('admin.kitchen') }}</th><th>★</th><th>{{ $t('admin.status') }}</th><th></th></tr></thead>
        <tbody>
          <tr v-for="c in chefs" :key="c.id" class="border-t border-gray-100">
            <td class="py-1.5 font-medium">{{ c.business_name }}</td>
            <td>{{ c.rating?.toFixed(1) ?? '—' }} ({{ c.total_reviews }})</td>
            <td><span class="badge" :class="c.is_active ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'">{{ c.is_active ? $t('admin.active') : $t('admin.inactive') }}</span></td>
            <td class="text-right"><button class="text-sm hover:underline" :class="c.is_active ? 'text-red-600' : 'text-green-600'" @click="toggleChef(c)">{{ c.is_active ? $t('admin.deactivate') : $t('admin.activate') }}</button></td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Orders overview -->
    <div v-else-if="tab === 'orders'" class="card overflow-x-auto">
      <table class="w-full text-sm">
        <thead class="text-left text-gray-500"><tr><th class="py-1">{{ $t('admin.order') }}</th><th>{{ $t('admin.status') }}</th><th>{{ $t('admin.payment') }}</th><th class="text-right">{{ $t('admin.total') }}</th></tr></thead>
        <tbody>
          <tr v-for="o in orders" :key="o.id" class="border-t border-gray-100">
            <td class="py-1.5 font-mono text-xs text-gray-500">{{ o.order_code }}</td>
            <td><span class="badge" :class="statusClass(o.status)">{{ $t(`status.${o.status}`) }}</span></td>
            <td>{{ o.payment_method === 'card' ? '💳' : '💵' }} {{ $t(`payment.${o.payment_status}`) }}</td>
            <td class="text-right font-semibold">{{ money(o.total_price) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Promo codes -->
    <div v-else-if="tab === 'promos'" class="space-y-4">
      <form class="card grid gap-2 sm:grid-cols-6" @submit.prevent="createPromo">
        <input v-model="newPromo.code" class="input uppercase sm:col-span-2" :placeholder="$t('admin.promoCode')" required />
        <select v-model="newPromo.discount_type" class="input">
          <option value="percent">%</option>
          <option value="fixed">$</option>
        </select>
        <input v-model="newPromo.discount_value" type="number" min="0" step="0.5" class="input" :placeholder="$t('admin.promoValue')" required />
        <input v-model="newPromo.min_order" type="number" min="0" class="input" :placeholder="$t('admin.promoMin')" />
        <input v-model="newPromo.usage_limit" type="number" min="0" class="input" :placeholder="$t('admin.promoLimit')" />
        <button class="btn-primary sm:col-span-6">{{ $t('admin.promoCreate') }}</button>
      </form>
      <div class="card overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="text-left text-gray-500"><tr><th class="py-1">{{ $t('admin.promoCode') }}</th><th>{{ $t('admin.promoDiscount') }}</th><th>{{ $t('admin.promoUsage') }}</th><th>{{ $t('admin.status') }}</th><th></th></tr></thead>
          <tbody>
            <tr v-for="p in promos" :key="p.id" class="border-t border-gray-100">
              <td class="py-1.5 font-mono font-medium">{{ p.code }}</td>
              <td>{{ p.discount_type === 'percent' ? `${p.discount_value}%` : money(p.discount_value) }}<span v-if="p.min_order > 0" class="text-gray-400"> · min {{ money(p.min_order) }}</span></td>
              <td>{{ p.used_count }}<span v-if="p.usage_limit > 0"> / {{ p.usage_limit }}</span></td>
              <td><span class="badge" :class="p.is_active ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'">{{ p.is_active ? $t('admin.active') : $t('admin.inactive') }}</span></td>
              <td class="text-right"><button class="text-sm hover:underline" :class="p.is_active ? 'text-red-600' : 'text-green-600'" @click="togglePromo(p)">{{ p.is_active ? $t('admin.deactivate') : $t('admin.activate') }}</button></td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
