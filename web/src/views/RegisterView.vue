<script setup>
import { ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

const form = ref({ username: '', email: '', password: '', role: 'customer' })
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.register({ ...form.value })
    router.push({ name: auth.isChef ? 'chef-dashboard' : 'chefs' })
  } catch (e) {
    error.value = e.message || 'registration failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="mx-auto max-w-sm">
    <h1 class="mb-4 text-2xl font-bold">Create an account</h1>
    <form class="card space-y-4" @submit.prevent="submit">
      <div>
        <label class="label">Username</label>
        <input v-model="form.username" class="input" required minlength="3" />
      </div>
      <div>
        <label class="label">Email</label>
        <input v-model="form.email" type="email" class="input" required />
      </div>
      <div>
        <label class="label">Password</label>
        <input v-model="form.password" type="password" class="input" required minlength="6" />
      </div>
      <div>
        <label class="label">I am a…</label>
        <select v-model="form.role" class="input">
          <option value="customer">Customer</option>
          <option value="chef">Chef</option>
        </select>
      </div>
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
      <button class="btn-primary w-full" :disabled="loading">{{ loading ? 'Creating…' : 'Sign up' }}</button>
      <p class="text-center text-sm text-gray-500">
        Have an account?
        <RouterLink to="/login" class="text-brand-600 hover:underline">Log in</RouterLink>
      </p>
    </form>
  </div>
</template>
