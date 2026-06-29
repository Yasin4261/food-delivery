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
  <div class="mx-auto max-w-sm">
    <h1 class="mb-4 text-2xl font-bold">Log in</h1>
    <form class="card space-y-4" @submit.prevent="submit">
      <div>
        <label class="label">Email</label>
        <input v-model="email" type="email" class="input" required autocomplete="email" />
      </div>
      <div>
        <label class="label">Password</label>
        <input v-model="password" type="password" class="input" required autocomplete="current-password" />
      </div>
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
      <button class="btn-primary w-full" :disabled="loading">{{ loading ? 'Logging in…' : 'Log in' }}</button>
      <p class="text-center text-sm text-gray-500">
        No account?
        <RouterLink to="/register" class="text-brand-600 hover:underline">Sign up</RouterLink>
      </p>
    </form>
  </div>
</template>
