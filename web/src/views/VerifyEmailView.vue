<script setup>
import { onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const auth = useAuthStore()

// state: 'verifying' | 'success' | 'failed' | 'missing'
const state = ref('verifying')

onMounted(async () => {
  const token = route.query.token
  if (!token) {
    state.value = 'missing'
    return
  }
  try {
    await api.post('/auth/verify-email', { token })
    state.value = 'success'
    // If this browser is logged in, clear the unverified banner immediately.
    try {
      await auth.refresh()
    } catch {
      /* not logged in here — the banner will clear on next login */
    }
  } catch {
    state.value = 'failed'
  }
})
</script>

<template>
  <div class="mx-auto max-w-md">
    <div class="card space-y-4 text-center">
      <div v-if="state === 'verifying'" class="space-y-3">
        <div class="skeleton mx-auto h-6 w-40"></div>
        <p class="text-gray-500">{{ $t('verify.verifying') }}</p>
      </div>

      <template v-else-if="state === 'success'">
        <h1 class="page-title">{{ $t('verify.success') }}</h1>
        <p class="text-gray-500">{{ $t('verify.successText') }}</p>
        <RouterLink to="/" class="btn-primary inline-block">{{ $t('verify.continue') }}</RouterLink>
      </template>

      <template v-else-if="state === 'missing'">
        <h1 class="page-title">{{ $t('verify.failed') }}</h1>
        <p class="text-gray-500">{{ $t('verify.missing') }}</p>
        <RouterLink to="/login" class="btn-ghost inline-block">{{ $t('auth.backToLogin') }}</RouterLink>
      </template>

      <template v-else>
        <h1 class="page-title">{{ $t('verify.failed') }}</h1>
        <p class="text-gray-500">{{ $t('verify.failedHint') }}</p>
        <RouterLink to="/login" class="btn-ghost inline-block">{{ $t('auth.backToLogin') }}</RouterLink>
      </template>
    </div>
  </div>
</template>
