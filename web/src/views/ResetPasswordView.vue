<script setup>
import { ref } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { api } from '@/api/client'

const route = useRoute()

// The emailed link is APP_BASE_URL/reset-password?token=…; allow manual entry
// as a fallback when the token is missing from the URL.
const token = ref(typeof route.query.token === 'string' ? route.query.token : '')
const password = ref('')
const confirm = ref('')
const done = ref(false)
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  if (password.value !== confirm.value) {
    error.value = 'passwords do not match'
    return
  }
  loading.value = true
  try {
    await api.post('/auth/reset-password', { token: token.value, password: password.value })
    done.value = true
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
      <div class="text-4xl">🔒</div>
      <h1 class="mt-2 text-2xl font-bold tracking-tight">Choose a new password</h1>
      <p class="page-subtitle">Reset links are single-use and expire after 1 hour.</p>
    </div>

    <div v-if="done" class="card space-y-3 text-center shadow-md">
      <div class="text-3xl">✅</div>
      <p class="font-medium text-gray-700">Password updated</p>
      <p class="text-sm text-gray-500">You can now log in with your new password.</p>
      <RouterLink to="/login" class="btn-primary w-full">Log in</RouterLink>
    </div>

    <form v-else class="card space-y-4 shadow-md" @submit.prevent="submit">
      <div v-if="!route.query.token">
        <label class="label">Reset token</label>
        <input v-model="token" class="input" required placeholder="paste the token from your email" />
      </div>
      <div>
        <label class="label">New password</label>
        <input
          v-model="password"
          type="password"
          class="input"
          required
          minlength="6"
          autocomplete="new-password"
          placeholder="min. 6 characters"
        />
      </div>
      <div>
        <label class="label">Confirm password</label>
        <input v-model="confirm" type="password" class="input" required autocomplete="new-password" placeholder="repeat it" />
      </div>
      <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>
      <button class="btn-primary w-full" :disabled="loading || !token">
        {{ loading ? 'Saving…' : 'Reset password' }}
      </button>
      <p class="text-center text-sm text-gray-500">
        Link expired?
        <RouterLink to="/forgot-password" class="font-medium text-brand-600 hover:underline">Request a new one</RouterLink>
      </p>
    </form>
  </div>
</template>
