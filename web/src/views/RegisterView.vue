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
  <div class="mx-auto mt-6 max-w-sm">
    <div class="mb-6 text-center">
      <div class="text-4xl">👋</div>
      <h1 class="mt-2 text-2xl font-bold tracking-tight">Join Home Chef</h1>
      <p class="page-subtitle">Order homemade food — or start selling yours.</p>
    </div>
    <form class="card space-y-4 shadow-md" @submit.prevent="submit">
      <div>
        <label class="label">Username</label>
        <input v-model="form.username" class="input" required minlength="3" placeholder="yasin" />
      </div>
      <div>
        <label class="label">Email</label>
        <input v-model="form.email" type="email" class="input" required placeholder="you@example.com" />
      </div>
      <div>
        <label class="label">Password</label>
        <input v-model="form.password" type="password" class="input" required minlength="6" placeholder="min. 6 characters" />
      </div>
      <div>
        <label class="label">I want to…</label>
        <div class="grid grid-cols-2 gap-2">
          <button
            type="button"
            class="rounded-lg border px-3 py-2 text-sm font-medium transition"
            :class="form.role === 'customer' ? 'border-brand-500 bg-brand-50 text-brand-700' : 'border-gray-300 text-gray-600 hover:bg-gray-50'"
            @click="form.role = 'customer'"
          >
            🛍️ Order food
          </button>
          <button
            type="button"
            class="rounded-lg border px-3 py-2 text-sm font-medium transition"
            :class="form.role === 'chef' ? 'border-brand-500 bg-brand-50 text-brand-700' : 'border-gray-300 text-gray-600 hover:bg-gray-50'"
            @click="form.role = 'chef'"
          >
            👨‍🍳 Cook &amp; sell
          </button>
        </div>
      </div>
      <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>
      <button class="btn-primary w-full" :disabled="loading">{{ loading ? 'Creating…' : 'Create account' }}</button>
      <p class="text-center text-sm text-gray-500">
        Have an account?
        <RouterLink to="/login" class="font-medium text-brand-600 hover:underline">Log in</RouterLink>
      </p>
    </form>
  </div>
</template>
