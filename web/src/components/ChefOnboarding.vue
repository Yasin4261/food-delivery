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
    <h1 class="mb-1 text-2xl font-bold">{{ $t('onboarding.title') }}</h1>
    <p class="mb-4 text-gray-600">
      {{ $t('onboarding.subtitle') }}
    </p>
    <form class="card space-y-4" @submit.prevent="submit">
      <div>
        <label class="label">{{ $t('onboarding.businessName') }}</label>
        <input v-model="form.business_name" class="input" required :placeholder="$t('onboarding.businessPlaceholder')" />
      </div>
      <div>
        <label class="label">{{ $t('onboarding.address') }}</label>
        <input v-model="form.kitchen_address" class="input" required :placeholder="$t('onboarding.addressPlaceholder')" />
      </div>
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="label">{{ $t('onboarding.city') }}</label>
          <input v-model="form.kitchen_city" class="input" :placeholder="$t('onboarding.cityPlaceholder')" />
        </div>
        <div>
          <label class="label">{{ $t('onboarding.specialty') }}</label>
          <input v-model="form.specialty" class="input" :placeholder="$t('onboarding.specialtyPlaceholder')" />
        </div>
      </div>
      <div>
        <label class="label">{{ $t('onboarding.bio') }}</label>
        <textarea v-model="form.bio" class="input" rows="2" :placeholder="$t('onboarding.bioPlaceholder')" />
      </div>
      <div class="grid grid-cols-3 gap-3">
        <div>
          <label class="label">{{ $t('onboarding.lat') }}</label>
          <input v-model="form.latitude" class="input" placeholder="41.0082" />
        </div>
        <div>
          <label class="label">{{ $t('onboarding.lng') }}</label>
          <input v-model="form.longitude" class="input" placeholder="28.9784" />
        </div>
        <div>
          <label class="label">{{ $t('onboarding.radius') }}</label>
          <input v-model="form.delivery_radius" type="number" min="1" class="input" />
        </div>
      </div>
      <p class="text-xs text-gray-500">
        {{ $t('onboarding.coordsHint') }}
      </p>
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
      <button class="btn-primary w-full" :disabled="saving">
        {{ saving ? $t('onboarding.creating') : $t('onboarding.create') }}
      </button>
    </form>
  </div>
</template>
