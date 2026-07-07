<script setup>
import { ref } from 'vue'
import { api } from '@/api/client'

const emit = defineEmits(['created'])

const form = ref({
  business_name: '',
  kitchen_address: '',
  kitchen_city: '',
  specialty: '',
  bio: '',
  latitude: '',
  longitude: '',
  delivery_radius: 5,
})
const error = ref('')
const saving = ref(false)

async function submit() {
  error.value = ''
  saving.value = true
  try {
    const f = form.value
    const payload = {
      business_name: f.business_name,
      kitchen_address: f.kitchen_address,
      kitchen_city: f.kitchen_city,
      specialty: f.specialty,
      bio: f.bio,
      delivery_radius: Number(f.delivery_radius) || 0,
    }
    // Coordinates must be sent together (the API rejects half a pair).
    if (f.latitude !== '' && f.longitude !== '') {
      payload.latitude = Number(f.latitude)
      payload.longitude = Number(f.longitude)
    }
    await api.post('/chefs', payload)
    emit('created')
  } catch (e) {
    error.value = e.message
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="mx-auto max-w-lg">
    <h1 class="mb-1 text-2xl font-bold">Set up your kitchen</h1>
    <p class="mb-4 text-gray-600">
      Create your chef profile to start receiving orders. You can add menus and dishes right after.
    </p>
    <form class="card space-y-4" @submit.prevent="submit">
      <div>
        <label class="label">Kitchen / business name</label>
        <input v-model="form.business_name" class="input" required placeholder="Yasin's Kitchen" />
      </div>
      <div>
        <label class="label">Kitchen address</label>
        <input v-model="form.kitchen_address" class="input" required placeholder="123 Main St" />
      </div>
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="label">City</label>
          <input v-model="form.kitchen_city" class="input" placeholder="Istanbul" />
        </div>
        <div>
          <label class="label">Specialty</label>
          <input v-model="form.specialty" class="input" placeholder="Homestyle Turkish" />
        </div>
      </div>
      <div>
        <label class="label">Bio</label>
        <textarea v-model="form.bio" class="input" rows="2" placeholder="Tell customers about your cooking…" />
      </div>
      <div class="grid grid-cols-3 gap-3">
        <div>
          <label class="label">Latitude</label>
          <input v-model="form.latitude" class="input" placeholder="41.0082" />
        </div>
        <div>
          <label class="label">Longitude</label>
          <input v-model="form.longitude" class="input" placeholder="28.9784" />
        </div>
        <div>
          <label class="label">Radius (km)</label>
          <input v-model="form.delivery_radius" type="number" min="1" class="input" />
        </div>
      </div>
      <p class="text-xs text-gray-500">
        Coordinates let customers find you with "nearby" search — you only appear there if they're set.
      </p>
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
      <button class="btn-primary w-full" :disabled="saving">
        {{ saving ? 'Creating…' : 'Create my kitchen' }}
      </button>
    </form>
  </div>
</template>
