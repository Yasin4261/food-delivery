<script setup>
import { onMounted, ref } from 'vue'
import { api } from '@/api/client'
import { i18n } from '@/i18n'
import StarRating from '@/components/StarRating.vue'

// Review targets for an order: each chef whose slice was DELIVERED (multi-chef
// carts advance per chef — food you never received isn't reviewable) and each
// of their dishes. Existing reviews load first, so already-rated targets show
// the given rating instead of an empty form — the rating history.
const props = defineProps({ order: { type: Object, required: true } })

const targets = ref([])
const loading = ref(true)

// Chefs whose slice arrived. Orders predating sub-orders fall back to the
// order-level status (their backfilled sub-orders carry it anyway).
function deliveredChefIds() {
  const subs = props.order.sub_orders
  if (subs?.length) return new Set(subs.filter((s) => s.status === 'delivered').map((s) => s.chef_id))
  return props.order.status === 'delivered' ? new Set((props.order.items ?? []).map((i) => i.chef_id)) : new Set()
}

function chefName(id) {
  return props.order.sub_orders?.find((s) => s.chef_id === id)?.chef_name || `Chef #${id}`
}

onMounted(async () => {
  let existing = []
  try {
    existing = await api.get(`/orders/${props.order.id}/reviews`)
  } catch {
    /* history unavailable -> plain forms */
  }
  const byChef = new Map(existing.filter((r) => r.chef_id).map((r) => [r.chef_id, r]))
  const byItem = new Map(existing.filter((r) => r.menu_item_id).map((r) => [r.menu_item_id, r]))

  const delivered = deliveredChefIds()
  const items = (props.order.items ?? []).filter((i) => delivered.has(i.chef_id))

  const out = []
  for (const id of new Set(items.map((i) => i.chef_id))) {
    const done = byChef.get(id)
    out.push({
      key: `chef-${id}`,
      label: `👨‍🍳 ${chefName(id)}`,
      field: 'chef_id',
      id,
      rating: done?.rating ?? 0,
      comment: done?.comment ?? '',
      state: done ? 'done' : 'idle',
      error: '',
    })
  }
  const seen = new Set()
  for (const it of items) {
    if (seen.has(it.menu_item_id)) continue
    seen.add(it.menu_item_id)
    const done = byItem.get(it.menu_item_id)
    out.push({
      key: `item-${it.menu_item_id}`,
      label: `🍽️ ${it.item_name}`,
      field: 'menu_item_id',
      id: it.menu_item_id,
      rating: done?.rating ?? 0,
      comment: done?.comment ?? '',
      state: done ? 'done' : 'idle',
      error: '',
    })
  }
  targets.value = out
  loading.value = false
})

async function submit(t) {
  if (!t.rating) {
    t.error = i18n.global.t('review.pickRating')
    return
  }
  t.state = 'saving'
  t.error = ''
  try {
    await api.post('/reviews', { order_id: props.order.id, [t.field]: t.id, rating: t.rating, comment: t.comment })
    t.state = 'done'
  } catch (e) {
    t.state = 'idle'
    t.error = e.status === 409 ? i18n.global.t('review.alreadyReviewed') : e.message
  }
}
</script>

<template>
  <div class="space-y-3 rounded-lg bg-gray-50 p-4">
    <p class="text-sm font-medium text-gray-700">{{ $t('review.prompt') }}</p>
    <p v-if="loading" class="text-sm text-gray-400">…</p>
    <p v-else-if="!targets.length" class="text-sm text-gray-500">{{ $t('review.nothingDeliveredYet') }}</p>
    <div
      v-for="t in targets"
      :key="t.key"
      class="flex flex-wrap items-center gap-3 rounded-lg border border-gray-200 bg-white px-3 py-2"
    >
      <span class="min-w-36 text-sm font-medium">{{ t.label }}</span>
      <template v-if="t.state === 'done'">
        <StarRating :model-value="t.rating" readonly />
        <span v-if="t.comment" class="max-w-60 truncate text-sm italic text-gray-500">“{{ t.comment }}”</span>
        <span class="text-sm text-green-600">{{ $t('review.yourRating') }}</span>
      </template>
      <template v-else>
        <StarRating v-model="t.rating" />
        <input v-model="t.comment" class="input max-w-52 grow" :placeholder="$t('review.commentPlaceholder')" />
        <button class="btn-ghost" :disabled="t.state === 'saving'" @click="submit(t)">
          {{ t.state === 'saving' ? $t('review.saving') : $t('review.submit') }}
        </button>
        <span v-if="t.error" class="text-sm text-red-600">{{ t.error }}</span>
      </template>
    </div>
  </div>
</template>
