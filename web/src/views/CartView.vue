<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { api } from '@/api/client'
import { useCartStore } from '@/stores/cart'

const cart = useCartStore()
const router = useRouter()

const deliveryAddress = ref('')
const paymentMethod = ref('cash')
const notes = ref('')
const error = ref('')
const placing = ref(false)

// Saved address book: preselect the default; "other" reveals the free-text
// input. selectedAddressId === 0 means "type a one-off address".
const savedAddresses = ref([])
const selectedAddressId = ref(0)
const usingSaved = computed(() => selectedAddressId.value !== 0)

onMounted(async () => {
  try {
    savedAddresses.value = await api.get('/addresses')
    const def = savedAddresses.value.find((a) => a.is_default) || savedAddresses.value[0]
    if (def) selectedAddressId.value = def.id
  } catch {
    /* no book -> free-text input only */
  }
})

async function placeOrder() {
  error.value = ''
  placing.value = true
  try {
    const payload = {
      payment_method: paymentMethod.value,
      customer_notes: notes.value,
      items: cart.lines.map((l) => ({ menu_item_id: l.menuItemId, quantity: l.quantity })),
    }
    if (usingSaved.value) payload.address_id = selectedAddressId.value
    else payload.delivery_address = deliveryAddress.value
    const order = await api.post('/orders', payload)
    cart.clear()
    router.push({ name: 'orders', query: { placed: order.id } })
  } catch (e) {
    error.value = e.message
  } finally {
    placing.value = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="page-title">{{ $t('cart.title') }}</h1>
      <p class="page-subtitle">{{ $t('cart.subtitle') }}</p>
    </div>

    <div v-if="!cart.lines.length" class="empty-state">
      <span class="empty-state-emoji">🛒</span>
      <p class="font-medium text-gray-600">{{ $t('cart.empty') }}</p>
      <p class="text-sm">
        <i18n-t keypath="cart.emptyHint" tag="span"><template #browse><RouterLink to="/" class="text-brand-600 hover:underline">{{ $t('cart.browseChefs') }}</RouterLink></template></i18n-t>
      </p>
    </div>

    <template v-else>
      <!-- One section per chef (a single order can span multiple chefs). -->
      <div v-for="group in cart.byChef" :key="group.chefId" class="card space-y-2">
        <h2 class="font-semibold">{{ group.chefName }}</h2>
        <div v-for="line in group.lines" :key="line.menuItemId" class="flex items-center justify-between gap-3">
          <span>{{ line.name }}</span>
          <div class="flex items-center gap-3">
            <input
              type="number"
              min="1"
              class="input w-16"
              :value="line.quantity"
              @input="cart.setQuantity(line.menuItemId, Number($event.target.value))"
            />
            <span class="w-16 text-right text-sm">${{ (line.price * line.quantity).toFixed(2) }}</span>
            <button class="text-sm text-red-600 hover:underline" @click="cart.remove(line.menuItemId)">{{ $t('cart.remove') }}</button>
          </div>
        </div>
      </div>

      <div class="flex justify-between text-lg font-semibold">
        <span>{{ $t('cart.total') }}</span>
        <span>${{ cart.total.toFixed(2) }}</span>
      </div>

      <form class="card space-y-4" @submit.prevent="placeOrder">
        <div>
          <label class="label">{{ $t('cart.deliveryAddress') }}</label>
          <select v-if="savedAddresses.length" v-model="selectedAddressId" class="input mb-2">
            <option v-for="a in savedAddresses" :key="a.id" :value="a.id">
              {{ a.label }} — {{ a.address }}<template v-if="a.city">, {{ a.city }}</template>
            </option>
            <option :value="0">{{ $t('cart.otherAddress') }}</option>
          </select>
          <input v-if="!usingSaved" v-model="deliveryAddress" class="input" required />
        </div>
        <div>
          <label class="label">{{ $t('cart.paymentLabel') }}</label>
          <select v-model="paymentMethod" class="input">
            <option value="cash">{{ $t('cart.cash') }}</option>
            <option value="card">{{ $t('cart.card') }}</option>
          </select>
        </div>
        <div>
          <label class="label">{{ $t('cart.notes') }}</label>
          <input v-model="notes" class="input" />
        </div>
        <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
        <button class="btn-primary w-full" :disabled="placing">
          {{ placing ? $t('cart.placing') : $t('cart.placeOrder', { total: `$${cart.total.toFixed(2)}` }) }}
        </button>
      </form>
    </template>
  </div>
</template>
