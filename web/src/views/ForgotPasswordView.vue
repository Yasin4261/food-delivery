<script setup>
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import { api } from '@/api/client'

const email = ref('')
const sent = ref(false)
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    // Always 202 — the API never reveals whether the email is registered.
    await api.post('/auth/forgot-password', { email: email.value })
    sent.value = true
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="mx-auto mt-6 max-w-sm">
    <div class="mb-6 text-center">
      <div class="text-4xl">🔑</div>
      <h1 class="mt-2 text-2xl font-bold tracking-tight">{{ $t('auth.forgotTitle') }}</h1>
      <p class="page-subtitle">{{ $t('auth.forgotSubtitle') }}</p>
    </div>

    <div v-if="sent" class="card space-y-3 text-center shadow-md">
      <div class="text-3xl">📬</div>
      <p class="font-medium text-gray-700">{{ $t('auth.checkInbox') }}</p>
      <p class="text-sm text-gray-500">{{ $t('auth.resetSent', { email }) }}</p>
      <RouterLink to="/login" class="btn-ghost w-full">{{ $t('auth.backToLogin') }}</RouterLink>
    </div>

    <form v-else class="card space-y-4 shadow-md" @submit.prevent="submit">
      <div>
        <label class="label">{{ $t('auth.email') }}</label>
        <input v-model="email" type="email" class="input" required autocomplete="email" :placeholder="$t('auth.emailPlaceholder')" />
      </div>
      <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>
      <button class="btn-primary w-full" :disabled="loading">{{ loading ? $t('auth.sending') : $t('auth.sendReset') }}</button>
      <p class="text-center text-sm text-gray-500">
        {{ $t('auth.remembered') }}
        <RouterLink to="/login" class="font-medium text-brand-600 hover:underline">{{ $t('auth.login') }}</RouterLink>
      </p>
    </form>
  </div>
</template>
