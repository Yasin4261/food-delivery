<script setup>
import { onMounted, reactive, ref } from 'vue'
import { formatMoney as money } from '@/lib/money'
import { api, page } from '@/api/client'
import { statusClass } from '@/lib/status'

const tab = ref('overview')
const tabs = ['overview', 'users', 'chefs', 'orders', 'promos', 'audit']

const stats = ref(null)
const users = ref([])
const chefs = ref([])
const orders = ref([])
const promos = ref([])
const newPromo = reactive({ code: '', discount_type: 'percent', discount_value: '', min_order: '', usage_limit: '' })
const error = ref('')

// Search / filter / pagination state per tab (#118). The API returns
// {data,limit,offset,total}; `total` drives the pager, so a filtered list pages
// over its matches rather than the whole table.
const PAGE_SIZE = 20
const filters = reactive({
  users: { q: '', role: '', active: '', offset: 0, total: 0 },
  chefs: { q: '', active: '', offset: 0, total: 0 },
  orders: { status: '', payment_status: '', offset: 0, total: 0 },
})

// qs builds a query string from the non-empty filter fields plus paging.
function qs(state, keys) {
  const p = new URLSearchParams()
  for (const k of keys) {
    if (state[k] !== '' && state[k] != null) p.set(k, state[k])
  }
  p.set('limit', PAGE_SIZE)
  p.set('offset', state.offset)
  return `?${p}`
}

async function loadStats() {
  try {
    stats.value = await api.get('/admin/stats')
  } catch (e) {
    error.value = e.message
  }
}
async function loadUsers() {
  error.value = ''
  try {
    const res = page(await api.get(`/admin/users${qs(filters.users, ['q', 'role', 'active'])}`))
    users.value = res.items
    filters.users.total = res.total
  } catch (e) {
    error.value = e.message
  }
}
async function loadChefs() {
  error.value = ''
  try {
    const res = page(await api.get(`/admin/chefs${qs(filters.chefs, ['q', 'active'])}`))
    chefs.value = res.items
    filters.chefs.total = res.total
  } catch (e) {
    error.value = e.message
  }
}
async function loadOrders() {
  error.value = ''
  try {
    const res = page(await api.get(`/admin/orders${qs(filters.orders, ['status', 'payment_status'])}`))
    orders.value = res.items
    filters.orders.total = res.total
  } catch (e) {
    error.value = e.message
  }
}

// applyFilters resets to the first page — otherwise a narrower filter can leave
// you stranded on an offset past the end of the new result set.
function applyFilters(name, load) {
  filters[name].offset = 0
  load()
}
function pageBy(name, load, delta) {
  const next = filters[name].offset + delta * PAGE_SIZE
  if (next < 0 || next >= filters[name].total) return
  filters[name].offset = next
  load()
}
// Support drill-in (#119): one call returns the whole context for a row.
const detail = ref(null)
const detailKind = ref('')
const detailLoading = ref(false)

async function openDetail(kind, id) {
  detailKind.value = kind
  detail.value = null
  detailLoading.value = true
  error.value = ''
  try {
    detail.value = await api.get(`/admin/${kind}/${id}`)
  } catch (e) {
    error.value = e.message
    detailKind.value = ''
  } finally {
    detailLoading.value = false
  }
}
const closeDetail = () => {
  detail.value = null
  detailKind.value = ''
}

const pageInfo = (name) => {
  const f = filters[name]
  if (!f.total) return '0'
  return `${f.offset + 1}–${Math.min(f.offset + PAGE_SIZE, f.total)} / ${f.total}`
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

async function deletePromo(p) {
  if (!window.confirm(i18n.global.t('admin.confirmDeletePromo', { code: p.code }))) return
  error.value = ''
  try {
    await api.del(`/admin/promos/${p.id}`)
    await loadPromos()
  } catch (e) {
    error.value = e.message
  }
}

// Audit log (#121).
const audits = ref([])
async function loadAudit() {
  error.value = ''
  try {
    audits.value = page(await api.get('/admin/audit?limit=100')).items
  } catch (e) {
    error.value = e.message
  }
}

function select(t) {
  tab.value = t
  error.value = ''
  // Always refetch: a filter that legitimately matches nothing must not look
  // like "not loaded yet", and an admin console should show current data.
  if (t === 'users') loadUsers()
  if (t === 'chefs') loadChefs()
  if (t === 'orders') loadOrders()
  if (t === 'promos' && !promos.value.length) loadPromos()
  if (t === 'audit') loadAudit()
}

async function toggleUser(u) {
  error.value = ''
  // Deactivation is destructive and requires a recorded reason (#121).
  let reason = ''
  if (u.is_active) {
    reason = window.prompt(i18n.global.t('admin.reasonPrompt'))
    if (reason === null) return
  }
  try {
    const r = await api.patch(`/admin/users/${u.id}/active`, { active: !u.is_active, reason })
    u.is_active = r.active
  } catch (e) {
    error.value = e.message
  }
}
async function toggleChef(c) {
  error.value = ''
  let reason = ''
  if (c.is_active) {
    reason = window.prompt(i18n.global.t('admin.reasonPrompt'))
    if (reason === null) return
  }
  try {
    const r = await api.patch(`/admin/chefs/${c.id}/active`, { active: !c.is_active, reason })
    c.is_active = r.active
  } catch (e) {
    error.value = e.message
  }
}
// Drive a chef's presence / availability on their behalf (support).
async function toggleChefOnline(c) {
  error.value = ''
  try {
    const r = await api.patch(`/admin/chefs/${c.id}/status`, { online: !c.is_online })
    c.is_online = r.online
  } catch (e) {
    error.value = e.message
  }
}
async function toggleChefAvailability(c) {
  error.value = ''
  try {
    const r = await api.patch(`/admin/chefs/${c.id}/availability`, { accepting_orders: !c.is_accepting_orders })
    c.is_accepting_orders = r.accepting_orders
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

    <!-- Support drill-in (#119): the whole context for one row, read-only. -->
    <div v-if="detailKind" class="card space-y-3 border-brand-200 bg-brand-50/40">
      <div class="flex items-center justify-between">
        <h2 class="font-semibold">{{ $t(`admin.detail_${detailKind}`) }}</h2>
        <button class="text-gray-400 hover:text-gray-600" @click="closeDetail">✕</button>
      </div>
      <p v-if="detailLoading" class="text-sm text-gray-500">…</p>

      <!-- User -->
      <div v-else-if="detail && detailKind === 'users'" class="space-y-2 text-sm">
        <p>
          <span class="font-medium">{{ detail.user.username }}</span>
          <span class="text-gray-500"> · {{ detail.user.email }}</span>
          <span class="badge ml-2 bg-gray-100 text-gray-600">{{ detail.user.role }}</span>
          <span class="badge ml-1" :class="detail.user.is_active ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'">
            {{ detail.user.is_active ? $t('admin.active') : $t('admin.inactive') }}
          </span>
        </p>
        <p v-if="detail.user.phone_number" class="text-gray-500">📞 {{ detail.user.phone_number }}</p>
        <p v-if="detail.chef" class="text-gray-600">🍲 {{ detail.chef.business_name }}</p>
        <p class="font-medium">{{ $t('admin.recentOrders') }} ({{ detail.orders.length }})</p>
        <ul class="space-y-1">
          <li v-for="o in detail.orders" :key="o.id" class="flex justify-between">
            <span class="font-mono text-xs text-gray-500">{{ o.order_code }}</span>
            <span class="badge" :class="statusClass(o.status)">{{ $t(`status.${o.status}`) }}</span>
          </li>
        </ul>
        <p class="font-medium">{{ $t('admin.reviewsWritten') }} ({{ detail.reviews.length }})</p>
      </div>

      <!-- Order -->
      <div v-else-if="detail && detailKind === 'orders'" class="space-y-2 text-sm">
        <p>
          <span class="font-mono text-gray-500">{{ detail.order.order_code }}</span>
          <span class="badge ml-2" :class="statusClass(detail.order.status)">{{ $t(`status.${detail.order.status}`) }}</span>
          <span class="badge ml-1 bg-gray-100 text-gray-600">{{ $t(`payment.${detail.order.payment_status}`) }}</span>
        </p>
        <p v-if="detail.customer" class="text-gray-600">
          👤 {{ detail.customer.username }} · {{ detail.customer.email }}
        </p>
        <p class="text-gray-500">📍 {{ detail.order.delivery_address }}</p>
        <ul class="space-y-1">
          <li v-for="it in detail.order.items" :key="it.id">{{ it.quantity }}× {{ it.item_name }}</li>
        </ul>
        <div v-if="detail.order.sub_orders?.length" class="flex flex-wrap gap-2">
          <span v-for="s in detail.order.sub_orders" :key="s.id" class="badge" :class="statusClass(s.status)">
            {{ s.chef_name }}: {{ $t(`status.${s.status}`) }}
          </span>
        </div>
        <p class="font-medium">{{ $t('admin.paymentAttempts') }} ({{ detail.payments.length }})</p>
        <ul class="space-y-1">
          <li v-for="p in detail.payments" :key="p.id" class="text-gray-600">
            {{ p.status }} · {{ new Date(p.created_at).toLocaleString() }}
          </li>
        </ul>
      </div>

      <!-- Chef -->
      <div v-else-if="detail && detailKind === 'chefs'" class="space-y-2 text-sm">
        <p>
          <span class="font-medium">{{ detail.chef.business_name }}</span>
          <span class="badge ml-2" :class="detail.chef.is_active ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'">
            {{ detail.chef.is_active ? $t('admin.active') : $t('admin.inactive') }}
          </span>
        </p>
        <p v-if="detail.owner" class="text-gray-600">👤 {{ detail.owner.username }} · {{ detail.owner.email }}</p>
        <p class="font-medium">{{ $t('admin.dishes') }} ({{ detail.items.length }})</p>
        <ul class="space-y-1">
          <li v-for="it in detail.items" :key="it.id" class="text-gray-600">{{ it.name }}</li>
        </ul>
        <p class="font-medium">{{ $t('admin.recentOrders') }} ({{ detail.orders.length }})</p>
      </div>
    </div>

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
    <div v-else-if="tab === 'users'" class="card space-y-3 overflow-x-auto">
      <div class="flex flex-wrap items-end gap-2">
        <input
          v-model="filters.users.q"
          class="input max-w-52"
          :placeholder="$t('admin.searchUsers')"
          @keyup.enter="applyFilters('users', loadUsers)"
        />
        <select v-model="filters.users.role" class="input max-w-36" @change="applyFilters('users', loadUsers)">
          <option value="">{{ $t('admin.allRoles') }}</option>
          <option value="customer">customer</option>
          <option value="chef">chef</option>
          <option value="admin">admin</option>
        </select>
        <select v-model="filters.users.active" class="input max-w-36" @change="applyFilters('users', loadUsers)">
          <option value="">{{ $t('admin.anyStatus') }}</option>
          <option value="true">{{ $t('admin.active') }}</option>
          <option value="false">{{ $t('admin.inactive') }}</option>
        </select>
        <button class="btn-ghost" @click="applyFilters('users', loadUsers)">{{ $t('admin.search') }}</button>
        <span class="ml-auto text-sm text-gray-500">{{ pageInfo('users') }}</span>
        <button class="btn-ghost" :disabled="filters.users.offset === 0" @click="pageBy('users', loadUsers, -1)">‹</button>
        <button
          class="btn-ghost"
          :disabled="filters.users.offset + 20 >= filters.users.total"
          @click="pageBy('users', loadUsers, 1)"
        >›</button>
      </div>
      <p v-if="!users.length" class="text-sm text-gray-500">{{ $t('admin.noResults') }}</p>
      <table v-else class="w-full text-sm">
        <thead class="text-left text-gray-500"><tr><th class="py-1">{{ $t('admin.user') }}</th><th>{{ $t('admin.role') }}</th><th>{{ $t('admin.status') }}</th><th></th></tr></thead>
        <tbody>
          <tr v-for="u in users" :key="u.id" class="cursor-pointer border-t border-gray-100 hover:bg-gray-50" @click="openDetail('users', u.id)">
            <td class="py-1.5"><span class="font-medium">{{ u.username }}</span> <span class="text-gray-400">{{ u.email }}</span></td>
            <td><span class="badge bg-gray-100 text-gray-600">{{ u.role }}</span></td>
            <td><span class="badge" :class="u.is_active ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'">{{ u.is_active ? $t('admin.active') : $t('admin.inactive') }}</span></td>
            <td class="text-right"><button class="text-sm hover:underline" :class="u.is_active ? 'text-red-600' : 'text-green-600'" @click.stop="toggleUser(u)">{{ u.is_active ? $t('admin.deactivate') : $t('admin.activate') }}</button></td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Chefs -->
    <div v-else-if="tab === 'chefs'" class="card space-y-3 overflow-x-auto">
      <div class="flex flex-wrap items-end gap-2">
        <input
          v-model="filters.chefs.q"
          class="input max-w-52"
          :placeholder="$t('admin.searchChefs')"
          @keyup.enter="applyFilters('chefs', loadChefs)"
        />
        <select v-model="filters.chefs.active" class="input max-w-36" @change="applyFilters('chefs', loadChefs)">
          <option value="">{{ $t('admin.anyStatus') }}</option>
          <option value="true">{{ $t('admin.active') }}</option>
          <option value="false">{{ $t('admin.inactive') }}</option>
        </select>
        <button class="btn-ghost" @click="applyFilters('chefs', loadChefs)">{{ $t('admin.search') }}</button>
        <span class="ml-auto text-sm text-gray-500">{{ pageInfo('chefs') }}</span>
        <button class="btn-ghost" :disabled="filters.chefs.offset === 0" @click="pageBy('chefs', loadChefs, -1)">‹</button>
        <button
          class="btn-ghost"
          :disabled="filters.chefs.offset + 20 >= filters.chefs.total"
          @click="pageBy('chefs', loadChefs, 1)"
        >›</button>
      </div>
      <p v-if="!chefs.length" class="text-sm text-gray-500">{{ $t('admin.noResults') }}</p>
      <table v-else class="w-full text-sm">
        <thead class="text-left text-gray-500"><tr><th class="py-1">{{ $t('admin.kitchen') }}</th><th>★</th><th>{{ $t('admin.status') }}</th><th></th></tr></thead>
        <tbody>
          <tr v-for="c in chefs" :key="c.id" class="cursor-pointer border-t border-gray-100 hover:bg-gray-50" @click="openDetail('chefs', c.id)">
            <td class="py-1.5 font-medium">{{ c.business_name }}</td>
            <td>{{ c.rating?.toFixed(1) ?? '—' }} ({{ c.total_reviews }})</td>
            <td>
              <span class="badge" :class="c.is_active ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'">{{ c.is_active ? $t('admin.active') : $t('admin.inactive') }}</span>
              <span v-if="c.is_online" class="badge ml-1 bg-emerald-100 text-emerald-700">●</span>
              <span v-if="!c.is_accepting_orders" class="badge ml-1 bg-amber-100 text-amber-700">🌴</span>
            </td>
            <td class="space-x-2 text-right">
              <button class="text-xs text-gray-500 hover:underline" @click.stop="toggleChefOnline(c)">{{ c.is_online ? $t('admin.goOffline') : $t('admin.goOnline') }}</button>
              <button class="text-xs text-gray-500 hover:underline" @click.stop="toggleChefAvailability(c)">{{ c.is_accepting_orders ? $t('admin.pause') : $t('admin.reopen') }}</button>
              <button class="text-sm hover:underline" :class="c.is_active ? 'text-red-600' : 'text-green-600'" @click.stop="toggleChef(c)">{{ c.is_active ? $t('admin.deactivate') : $t('admin.activate') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Orders overview -->
    <div v-else-if="tab === 'orders'" class="card space-y-3 overflow-x-auto">
      <div class="flex flex-wrap items-end gap-2">
        <select v-model="filters.orders.status" class="input max-w-40" @change="applyFilters('orders', loadOrders)">
          <option value="">{{ $t('admin.anyOrderStatus') }}</option>
          <option v-for="s in ['pending','confirmed','preparing','ready','delivering','delivered','cancelled']" :key="s" :value="s">
            {{ $t(`status.${s}`) }}
          </option>
        </select>
        <select v-model="filters.orders.payment_status" class="input max-w-40" @change="applyFilters('orders', loadOrders)">
          <option value="">{{ $t('admin.anyPayment') }}</option>
          <option v-for="s in ['pending','paid','failed','refunded']" :key="s" :value="s">{{ $t(`payment.${s}`) }}</option>
        </select>
        <span class="ml-auto text-sm text-gray-500">{{ pageInfo('orders') }}</span>
        <button class="btn-ghost" :disabled="filters.orders.offset === 0" @click="pageBy('orders', loadOrders, -1)">‹</button>
        <button
          class="btn-ghost"
          :disabled="filters.orders.offset + 20 >= filters.orders.total"
          @click="pageBy('orders', loadOrders, 1)"
        >›</button>
      </div>
      <p v-if="!orders.length" class="text-sm text-gray-500">{{ $t('admin.noResults') }}</p>
      <table v-else class="w-full text-sm">
        <thead class="text-left text-gray-500"><tr><th class="py-1">{{ $t('admin.order') }}</th><th>{{ $t('admin.status') }}</th><th>{{ $t('admin.payment') }}</th><th class="text-right">{{ $t('admin.total') }}</th></tr></thead>
        <tbody>
          <tr v-for="o in orders" :key="o.id" class="cursor-pointer border-t border-gray-100 hover:bg-gray-50" @click="openDetail('orders', o.id)">
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
              <td class="space-x-2 text-right">
                <button class="text-sm hover:underline" :class="p.is_active ? 'text-red-600' : 'text-green-600'" @click="togglePromo(p)">{{ p.is_active ? $t('admin.deactivate') : $t('admin.activate') }}</button>
                <button class="text-sm text-red-600 hover:underline" @click="deletePromo(p)">{{ $t('admin.delete') }}</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Audit log (#121): read-only trail of every admin mutation -->
    <div v-else-if="tab === 'audit'" class="card overflow-x-auto">
      <p v-if="!audits.length" class="text-sm text-gray-500">{{ $t('admin.noAudit') }}</p>
      <table v-else class="w-full text-sm">
        <thead class="text-left text-gray-500">
          <tr><th class="py-1">{{ $t('admin.when') }}</th><th>{{ $t('admin.action') }}</th><th>{{ $t('admin.target') }}</th><th>{{ $t('admin.actor') }}</th><th>{{ $t('admin.reason') }}</th></tr>
        </thead>
        <tbody>
          <tr v-for="a in audits" :key="a.id" class="border-t border-gray-100">
            <td class="py-1.5 whitespace-nowrap text-gray-500">{{ new Date(a.created_at).toLocaleString() }}</td>
            <td class="font-mono text-xs">{{ a.action }}</td>
            <td>{{ a.target_type }} #{{ a.target_id }}</td>
            <td>#{{ a.actor_user_id }}</td>
            <td class="text-gray-500">{{ a.reason || '—' }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
