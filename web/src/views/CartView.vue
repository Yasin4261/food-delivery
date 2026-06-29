<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import { useCartStore } from '@/stores/cart'

const cart = useCartStore()
const router = useRouter()

const deliveryAddress = ref('')
const paymentMethod = ref('cash')
const notes = ref('')
const error = ref('')
const placing = ref(false)

async function placeOrder() {
  error.value = ''
  placing.value = true
  try {
    const order = await api.post('/orders', {
      delivery_address: deliveryAddress.value,
      payment_method: paymentMethod.value,
      customer_notes: notes.value,
      items: cart.lines.map((l) => ({ menu_item_id: l.menuItemId, quantity: l.quantity })),
    })
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
    <h1 class="text-2xl font-bold">Your cart</h1>

    <p v-if="!cart.lines.length" class="text-gray-500">Your cart is empty.</p>

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
            <button class="text-sm text-red-600 hover:underline" @click="cart.remove(line.menuItemId)">remove</button>
          </div>
        </div>
      </div>

      <div class="flex justify-between text-lg font-semibold">
        <span>Total</span>
        <span>${{ cart.total.toFixed(2) }}</span>
      </div>

      <form class="card space-y-4" @submit.prevent="placeOrder">
        <div>
          <label class="label">Delivery address</label>
          <input v-model="deliveryAddress" class="input" required />
        </div>
        <div>
          <label class="label">Payment</label>
          <select v-model="paymentMethod" class="input">
            <option value="cash">Cash</option>
            <option value="card">Card</option>
          </select>
        </div>
        <div>
          <label class="label">Notes (optional)</label>
          <input v-model="notes" class="input" />
        </div>
        <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
        <button class="btn-primary w-full" :disabled="placing">
          {{ placing ? 'Placing…' : `Place order — $${cart.total.toFixed(2)}` }}
        </button>
      </form>
    </template>
  </div>
</template>
