<script setup>
import { ref } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(email.value, password.value)
    const redirect = route.query.redirect
    if (redirect) router.push(redirect)
    else router.push({ name: auth.isChef ? 'chef-dashboard' : 'chefs' })
  } catch (e) {
    error.value = e.message || 'login failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="mx-auto mt-6 max-w-sm">
    <div class="mb-6 text-center">
      <div class="text-4xl">🍲</div>
      <h1 class="mt-2 text-2xl font-bold tracking-tight">{{ $t('auth.welcomeTitle') }}</h1>
      <p class="page-subtitle">{{ $t('auth.welcomeSubtitle') }}</p>
    </div>
    <form class="card space-y-4 shadow-md" @submit.prevent="submit">
      <div>
        <label class="label">{{ $t('auth.email') }}</label>
        <input v-model="email" type="email" class="input" required autocomplete="email" :placeholder="$t('auth.emailPlaceholder')" />
      </div>
      <div>
        <div class="mb-1 flex items-center justify-between">
          <label class="label !mb-0">{{ $t('auth.password') }}</label>
          <RouterLink to="/forgot-password" class="text-xs font-medium text-brand-600 hover:underline">
            {{ $t('auth.forgotPassword') }}
          </RouterLink>
        </div>
        <input
          v-model="password"
          type="password"
          class="input"
          required
          autocomplete="current-password"
          :placeholder="$t('auth.passwordDots')"
        />
      </div>
      <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>
      <button class="btn-primary w-full" :disabled="loading">{{ loading ? $t('auth.loggingIn') : $t('auth.login') }}</button>
      <p class="text-center text-sm text-gray-500">
        {{ $t('auth.noAccount') }}
        <RouterLink to="/register" class="font-medium text-brand-600 hover:underline">{{ $t('auth.signup') }}</RouterLink>
      </p>
    </form>
  </div>
</template>
