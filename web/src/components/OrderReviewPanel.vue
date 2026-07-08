<script setup>
import { onMounted, ref } from 'vue'
import { api } from '@/api/client'
import StarRating from '@/components/StarRating.vue'

// Review targets for a delivered order: each chef involved (multi-chef carts
// have several) and each dish. Every target submits its own POST /reviews.
const props = defineProps({ order: { type: Object, required: true } })

const targets = ref([])

onMounted(async () => {
  const items = props.order.items ?? []
  const chefIds = [...new Set(items.map((i) => i.chef_id))]
  const chefTargets = await Promise.all(
    chefIds.map(async (id) => {
      let label = `Chef #${id}`
      try {
        label = (await api.get(`/chefs/${id}`)).business_name
      } catch {
        // keep the fallback label
      }
      return { key: `chef-${id}`, label: `👨‍🍳 ${label}`, field: 'chef_id', id, rating: 0, comment: '', state: 'idle', error: '' }
    }),
  )
  const seen = new Set()
  const itemTargets = []
  for (const it of items) {
    if (seen.has(it.menu_item_id)) continue
    seen.add(it.menu_item_id)
    itemTargets.push({
      key: `item-${it.menu_item_id}`,
      label: `🍽️ ${it.item_name}`,
      field: 'menu_item_id',
      id: it.menu_item_id,
      rating: 0,
      comment: '',
      state: 'idle',
      error: '',
    })
  }
  targets.value = [...chefTargets, ...itemTargets]
})

async function submit(t) {
  if (!t.rating) {
    t.error = 'pick a star rating first'
    return
  }
  t.state = 'saving'
  t.error = ''
  try {
    await api.post('/reviews', { order_id: props.order.id, [t.field]: t.id, rating: t.rating, comment: t.comment })
    t.state = 'done'
  } catch (e) {
    t.state = 'idle'
    t.error = e.status === 409 ? 'you already reviewed this' : e.message
  }
}
</script>

<template>
  <div class="space-y-3 rounded-lg bg-gray-50 p-4">
    <p class="text-sm font-medium text-gray-700">How was it? Rate the chef and dishes ⭐</p>
    <div
      v-for="t in targets"
      :key="t.key"
      class="flex flex-wrap items-center gap-3 rounded-lg border border-gray-200 bg-white px-3 py-2"
    >
      <span class="min-w-36 text-sm font-medium">{{ t.label }}</span>
      <template v-if="t.state === 'done'">
        <span class="text-sm text-green-600">✓ Thanks for your review!</span>
      </template>
      <template v-else>
        <StarRating v-model="t.rating" />
        <input v-model="t.comment" class="input max-w-52 grow" placeholder="Say a few words (optional)" />
        <button class="btn-ghost" :disabled="t.state === 'saving'" @click="submit(t)">
          {{ t.state === 'saving' ? 'Saving…' : 'Submit' }}
        </button>
        <span v-if="t.error" class="text-sm text-red-600">{{ t.error }}</span>
      </template>
    </div>
  </div>
</template>
